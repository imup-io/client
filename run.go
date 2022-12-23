package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

func run(ctx context.Context, shutdown chan os.Signal) error {
	log.Debugf("Version: %s", ClientVersion)
	imup := newApp()
	imup.Errors = NewErrMap(imup.cfg.HostID(), imup.cfg.Env())

	log.Infof("imup: %+v \n", imup)
	log.Infof("config: %+v \n", imup.cfg)

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
				log.Error(err)
			} else if err := imup.authorized(cctx, bytes.NewBuffer(b), imup.RealtimeAuthorized); err != nil {
				log.Errorf("failed to check client authorization %s", err)
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
	// Random Speed Testing
	//
	// collects speed test data using the ndt7 protocol
	// data is collected at least once every 6 hours
	go func() {
		ticker := time.NewTicker(sleepTime())
		defer ticker.Stop()
		for {
			if imup.cfg.SpeedTests() {
				if err := imup.runSpeedTest(cctx); err != nil {
					log.Error(err)
					imup.Errors.write("CollectSpeedTestData", err)
				} else {
					imup.Errors.reportErrors("CollectSpeedTestData")
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
	wg := sync.WaitGroup{}
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
			collected := tester.Collect(cctx, strings.Split(imup.PingAddressesExternal, ","))
			data = append(data, collected...)
			log.Debugf("data points collected: %v", len(data))

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
				log.Debugf("data points to persist? %v", len(data) > 0)
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

	// ======================================================================
	// Realtime
	//
	// These functions should run on their own goroutines so
	// as not to block each other

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
						log.Error(err)
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
						log.Error(err)
						imup.Errors.write("ShouldRunSpeedtest", err)
					} else if ok {
						// let the api know we're ready to run the speed test
						imup.OnDemandSpeedTest = true
						if err := imup.postSpeedTestRealtimeStatus(cctx, "running"); err != nil {
							log.Error(err)
							imup.Errors.write("PostSpeedTestStatus", err)
						}

						// run speed test
						if err := imup.runSpeedTest(cctx); err != nil {
							log.Error(err)
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

	// remote configuration reload
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
						log.Error(err)
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

	sig := <-shutdown

	log.Infof("shutdown started signal: %v", sig)
	cancel()
	wg.Wait()
	defer log.Infof("shutdown completed. signal: %v", sig)

	return nil
}
