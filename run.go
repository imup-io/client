package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/imup-io/client/util"
	log "golang.org/x/exp/slog"
)

func run(ctx context.Context, shutdown chan os.Signal) error {
	log.Debug("Starting Client", "Version", ClientVersion)
	imup := newApp()
	imup.Errors = NewErrMap(imup.cfg.HostID(), imup.cfg.Env())

	log.Info("imup setup", "client", fmt.Sprintf("imup: %+v", imup))
	log.Info("imup config", "config", fmt.Sprintf("config: %+v", imup.cfg))

	// define a context with cancel to coordinate shutdown behavior
	cctx, cancel := context.WithCancel(ctx)

	// sendDataWorker listens for imup data
	go sendDataWorker(cctx, imup.ChannelImupData)

	// check for and send data from local user cache
	if cachedJobs, ok := fromCacheDir(); ok {
		for _, job := range cachedJobs {
			imup.ChannelImupData <- job
		}
		clearCache()
	}

	// ======================================================================
	// Refresh Public IP Address
	//
	// refresh public ip address every 15 minutes if client has a defined allow or block list

	go func() {
		ticker := time.NewTicker((15 * time.Minute))
		defer ticker.Stop()

		for {
			// only refresh a clients public ip address if configured to allow/block specific ips
			if len(imup.cfg.AllowedIPs()) > 0 || len(imup.cfg.BlockedIPs()) > 0 {
				imup.cfg.RefreshPublicIP()
			}

			select {
			case <-ticker.C:
				continue
			case <-cctx.Done():
				return
			}
		}
	}()

	// ======================================================================
	// Authorization
	//
	// check to see if client is authorized for realtime features
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()

		for {
			ar := &authRequest{Key: imup.cfg.APIKey(), Email: imup.cfg.EmailAddress()}
			b, err := json.Marshal(ar)
			if err != nil {
				log.Error("failed to marshal auth request", "error", err)
			} else if err := imup.authorized(cctx, bytes.NewBuffer(b), imup.RealtimeAuthorized); err != nil {
				log.Error("failed to check client authorization", "error", err)
			}

			select {
			case <-ticker.C:
				continue
			case <-cctx.Done():
				return
			}
		}
	}()

	// ======================================================================
	// Realtime
	//
	// These functions should run on their own goroutines so
	// as not to block each other

	// remote configuration reload
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			ticker := time.NewTicker(time.Duration(1 * time.Hour))
			defer ticker.Stop()
			for {
				if imup.cfg.Realtime() {
					// when api sends a new config, reload it
					if err := imup.remoteConfigReload(cctx); err != nil {
						log.Error("failed to reload config", "error", err)
						imup.Errors.write("RemoteConfigReload", err)
					} else {
						imup.Errors.reportErrors("RemoteConfigReload")
					}
				}

				select {
				case <-ticker.C:
					continue
				case <-cctx.Done():
					return
				}
			}
		}
	}()

	// liveness checkin
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			ticker := time.NewTicker(time.Duration(10 * time.Second))
			defer ticker.Stop()
			for {

				if imup.cfg.Realtime() {
					// liveness checkin
					if err := imup.sendClientHealthy(cctx); err != nil {
						log.Error("failed liveness checkin", "error", err)
						imup.Errors.write("SendClientHealthy", err)
					} else {
						imup.Errors.reportErrors("SendClientHealthy")
					}
				}
				select {
				case <-ticker.C:
					continue
				case <-cctx.Done():
					return
				}
			}
		}
	}()

	// on demand speed tests
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			ticker := time.NewTicker(time.Duration(10 * time.Second))
			defer ticker.Stop()
			for {

				if imup.cfg.Realtime() {
					if ok, err := imup.shouldRunSpeedtest(cctx); err != nil {
						log.Error("failed on-demand speed test check", "error", err)
						imup.Errors.write("ShouldRunSpeedtest", err)
					} else if ok {
						// let the api know we're ready to run the speed test
						imup.OnDemandSpeedTest = true
						if err := imup.postSpeedTestRealtimeStatus(cctx, "running"); err != nil {
							log.Error("failed to post realtime speedtest", "error", err)
							imup.Errors.write("PostSpeedTestStatus", err)
						}

						// run speed test
						if err := imup.runSpeedTest(cctx); err != nil {
							log.Error("failed to run on-demand speed test", "error", err)
							imup.Errors.write("RunSpeedTestOnce", err)
						}

						imup.OnDemandSpeedTest = false
					} else {
						// nothing else to do so lets clear out any errors that we've collected
						imup.Errors.reportErrors("ShouldRunSpeedtest")
						imup.Errors.reportErrors("PostSpeedTestStatus")
						imup.Errors.reportErrors("RunSpeedTestOnce")
					}
				}
				select {
				case <-ticker.C:
					continue
				case <-cctx.Done():
					return
				}
			}
		}
	}()

	// ======================================================================
	// Random Speed Testing
	//
	// collects speed test data using the ndt7 protocol
	// data is collected at least once every 6 hours
	go func() {
		ticker := time.NewTicker(sleepTime())
		defer ticker.Stop()
		for {
			if imup.cfg.SpeedTests() {
				monitoring := util.IPMonitored(imup.cfg.PublicIP(), imup.cfg.AllowedIPs(), imup.cfg.BlockedIPs())

				// extra check if ip based speed testing is configured
				if monitoring {
					if err := imup.runSpeedTest(cctx); err != nil {
						log.Error("failed to run speed test", "error", err)
						imup.Errors.write("CollectSpeedTestData", err)
					} else {
						imup.Errors.reportErrors("CollectSpeedTestData")
					}
				}
			}

			select {
			case <-ticker.C:
				continue
			case <-cctx.Done():
				return
			}
		}
	}()

	// ======================================================================
	// Connectivity Testing
	//
	// using either ICMP or TCP setup run connectivity tests
	// on regular intervals, the default is continuous polling
	// with statistics calculated for each minute
	wg.Add(1)
	data := make([]pingStats, 0, 30)
	var tester imupStatCollector
	go func() {
		defer wg.Done()

		// initialize a tester
		if imup.cfg.PingTests() {
			tester = imup.newPingStats()
		} else {
			tester = imup.newDialerStats()
		}

		ticker := time.NewTicker(tester.Interval())
		defer ticker.Stop()
		for {
			monitoring := util.IPMonitored(imup.cfg.PublicIP(), imup.cfg.AllowedIPs(), imup.cfg.BlockedIPs())
			if monitoring {

				collected := tester.Collect(cctx, strings.Split(imup.PingAddressesExternal, ","))
				data = append(data, collected...)
				log.Debug("data points collected", "count", len(data))

				if imup.cfg.StoreJobsOnDisk() {
					sc, dt := tester.DetectDowntime(data)
					toUserCache(sendDataJob{
						IMUPAddress: imup.APIPostConnectionData,
						IMUPData: imupData{
							Downtime:      dt,
							StatusChanged: sc,
							Email:         imup.cfg.EmailAddress(),
							ID:            imup.cfg.HostID(),
							Key:           imup.cfg.APIKey(),
							IMUPData:      collected,
						}})
				}
			}

			if len(data) >= imup.IMUPDataLength {
				sc, dt := tester.DetectDowntime(data)
				// enqueue a job
				imup.ChannelImupData <- sendDataJob{
					IMUPAddress: imup.APIPostConnectionData,
					IMUPData: imupData{
						Downtime:      dt,
						StatusChanged: sc,
						Email:         imup.cfg.EmailAddress(),
						ID:            imup.cfg.HostID(),
						Key:           imup.cfg.APIKey(),
						IMUPData:      data,
					},
				}
				// reset connData slice
				data = nil
				if imup.cfg.StoreJobsOnDisk() {
					clearCache()
				}
			}

			select {
			case <-ticker.C:
				continue
			case <-cctx.Done():
				log.Debug("data points to persist?", "data > 0", len(data) > 0)
				if len(data) > 0 {
					sc, dt := tester.DetectDowntime(data)
					log.Debug("persisting pending conn data")
					toUserCache(sendDataJob{
						IMUPAddress: imup.APIPostConnectionData,
						IMUPData: imupData{
							Downtime:      dt,
							StatusChanged: sc,
							Email:         imup.cfg.EmailAddress(),
							ID:            imup.cfg.HostID(),
							Key:           imup.cfg.APIKey(),
							IMUPData:      data,
						}})
				}
			}
			return
		}
	}()

	sig := <-shutdown

	log.Info("shutdown started", "signal", sig)
	cancel()
	wg.Wait()
	defer log.Info("shutdown completed", "signal", sig)

	return nil
}
