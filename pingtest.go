package main

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"sync"
	"time"

	ping "github.com/prometheus-community/pro-bing"
	log "github.com/sirupsen/logrus"
)

type pingTest struct {
	addressInternal string
	avoidAddrs      map[string]bool
	count           int
	debug           bool

	delay    time.Duration
	interval time.Duration
	timeout  time.Duration
}

func (i *imup) newPingStats() imupStatCollector {
	return &pingTest{
		addressInternal: i.PingAddressInternal,
		avoidAddrs:      i.PingAddressesAvoid,
		count:           i.PingRequests,
		debug:           i.cfg.DevelopmentEnvironment(),
		delay:           time.Duration(i.PingDelay) * time.Millisecond,
		interval:        time.Duration(i.PingInterval) * time.Second,
		timeout:         time.Duration(i.PingInterval) * time.Second,
	}
}

// Interval is the time to wait between ping testing
func (p *pingTest) Interval() time.Duration {
	return p.interval
}

// Collect takes a list of address' to test against and collects ping statistics once per Interval.
func (p *pingTest) Collect(ctx context.Context, pingAddrs []string) []pingStats {
	externalPingResult := pingStats{}
	internalPingResult := pingStats{}
	var success bool
	var internalSuccess bool

	timestamp := time.Now().UnixNano()

	// run ping collectors in parallel
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		externalPingResult, success = p.checkConnectivity(ctx, "external", pingAddrs, timestamp)
	}()

	// do not test internal gateway if its disabled
	if p.addressInternal != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			internalPingResult, internalSuccess = p.checkConnectivity(ctx, "internal", []string{p.addressInternal}, timestamp)
		}()
	}
	wg.Wait()

	if success {
		return []pingStats{externalPingResult}
	}

	// internal testing disabled, return external result only
	if !success && p.addressInternal == "" {
		return []pingStats{externalPingResult}
	}

	if internalSuccess {
		log.Infof("No external endpoint could be reached but gateway at '%s' was reachable", p.addressInternal)
		// TODO: implement an app wide file logger
		// fileLogger.Info("Pinging external endpoint failed but gateway was reachable")
	} else {
		log.Infof("No external endpoint could be reached and gateway at '%s' was unreachable", p.addressInternal)
		// TODO: implement an app wide file logger
		// fileLogger.Info("Pinging external endpoint failed and gateway was unreachable")
	}

	// external data needs to send last for tests to pingDownTimeDetect to work properly
	externalPingResult.SuccessInternal = internalSuccess
	return []pingStats{internalPingResult, externalPingResult}
}

// DetectDowntime only increments downtime if Success is false but Internal Success is true
// demonstrating a connection to the gateway is not the problem
func (p *pingTest) DetectDowntime(data []pingStats) (bool, int) {
	if len((data)) == 0 {
		return false, 0
	}

	changed := false
	downtime := 0

	lastStatus := data[0].Success
	for _, pd := range data {
		if pd.EndpointType == "external" && !pd.Success && pd.SuccessInternal {
			downtime++
		}

		if pd.EndpointType == "external" && pd.Success != lastStatus {
			changed = true
		}
		lastStatus = pd.Success
	}

	return changed, downtime
}

// checkConnectivity gathers ICMP statistics for a given address
func (p *pingTest) checkConnectivity(ctx context.Context, testType string, pingAddrs []string, timestamp int64) (pingStats, bool) {
	var err error
	var pinger *ping.Pinger
	if testType == "external" {
		if pinger, err = p.setupExternalPinger(ctx, pingAddrs, nil); err != nil {
			log.Errorf("failed to setup external pinger: %s", err)
			return pingStats{PacketsSent: 0, PacketsRecv: 0, PacketLoss: 100.0, TimeStamp: timestamp}, false
		}
	} else {
		if pinger, err = p.setupInternalPinger(ctx, p.addressInternal, nil); err != nil {
			log.Errorf("failed to setup internal pinger: %s", err)
			return pingStats{PacketsSent: 0, PacketsRecv: 0, PacketLoss: 100.0, TimeStamp: timestamp}, false
		}
	}

	pinger.Timeout = p.timeout
	pinger.Interval = p.delay
	pinger.Count = p.count

	var info error
	var stats *ping.Statistics
	if stats, info = p.run(ctx, pinger); info != nil {
		log.Warnf("error sending ping to: %s: err: %s", stats.Addr, info)
	}

	loss := 100.0
	if !math.IsNaN(stats.PacketLoss) {
		loss = stats.PacketLoss
	}

	// default SuccessInternal to true for external tests
	success := stats.PacketsRecv > 0
	var successInternal bool
	if successInternal = success; testType == "external" {
		successInternal = true
	}

	data := pingStats{
		PingAddress:     stats.Addr,
		Success:         success,
		SuccessInternal: successInternal,
		TimeStamp:       timestamp,
		ClientVersion:   ClientVersion,
		OS:              runtime.GOOS,
		EndpointType:    testType,
		PacketsRecv:     stats.PacketsRecv,
		PacketsSent:     stats.PacketsSent,
		PacketLoss:      loss,
		// Rtts: 	stats.Rtts,
		MinRtt:    stats.MinRtt,
		MaxRtt:    stats.MaxRtt,
		AvgRtt:    stats.AvgRtt,
		StdDevRtt: stats.StdDevRtt,
	}

	return data, success
}

// setupInternalPinger is a helper function that verifies an internal gateway if available to ping before performing a longer running test
func (p *pingTest) setupInternalPinger(ctx context.Context, pingAddr string, errs error) (*ping.Pinger, error) {
	if ctx.Err() == context.Canceled {
		return nil, ctx.Err()
	}

	pinger, err := ping.NewPinger(pingAddr)
	if err != nil {
		return nil, fmt.Errorf("pinger could not be created: %v: %v", err, errs)
	}

	pinger.Timeout = time.Duration(time.Second * 5)
	pinger.Count = 1

	if stats, err := p.run(ctx, pinger); stats.PacketsRecv == 0 {
		return nil, fmt.Errorf("pinger could not be created: %s: %v", err, errs)
	}

	if pinger, err := ping.NewPinger(pingAddr); err != nil {
		return nil, fmt.Errorf("pinger could not be created: %v: %v", err, errs)
	} else {
		log.Debugf("successfully created pinger to test against %s", pinger.Addr())
		return pinger, nil
	}
}

// setupExternalPinger is a helper function that verifies an address is available to ping before performing a longer running test
// it will recursively exhaust the list of available addresses to test against before giving up
func (p *pingTest) setupExternalPinger(ctx context.Context, pingAddrs []string, errs error) (*ping.Pinger, error) {
	if ctx.Err() == context.Canceled {
		return nil, ctx.Err()
	}

	if len(pingAddrs) == 0 {
		return nil, fmt.Errorf("could not resolve any ping address to run pinger : %v", errs)
	}

	randomPingAddr := pingAddress(pingAddrs, p.avoidAddrs)
	pinger, err := ping.NewPinger(randomPingAddr)
	if err != nil {
		return nil, fmt.Errorf("pinger could not be created: %v: %v", err, errs)
	}

	pinger.Timeout = time.Duration(time.Second * 5)
	pinger.Count = 1

	if stats, err := p.run(ctx, pinger); stats.PacketsRecv == 0 {
		log.Debugf("avoiding pinging external endpoint at: %s next check", randomPingAddr)
		p.avoidAddrs[randomPingAddr] = true

		for i, addr := range pingAddrs {
			if addr == randomPingAddr {
				pingAddrs = append(pingAddrs[:i], pingAddrs[i+1:]...)
				return p.setupExternalPinger(ctx, pingAddrs, fmt.Errorf("test ping timed out for: %s : %v : %v", randomPingAddr, err, errs))
			}
		}
		return nil, fmt.Errorf("pinger could not be created: %s: %v", err, errs)
	}

	if pinger, err := ping.NewPinger(randomPingAddr); err != nil {
		return p.setupExternalPinger(ctx, pingAddrs, fmt.Errorf("ping succeeds but creating new pinger failed: %s : %v : %v", randomPingAddr, err, errs))
	} else {
		log.Debugf("successfully created pinger to test against %s", pinger.Addr())
		return pinger, nil
	}
}

// run always returns ping statistics and a non nil error
func (p *pingTest) run(ctx context.Context, pinger *ping.Pinger) (*ping.Statistics, error) {
	pinger.Debug = true
	pinger.OnRecv = func(pkt *ping.Packet) {
		log.Tracef("%d bytes sent %s: icmp_seq=%d time=%v\n", pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt)
	}

	pinger.OnFinish = func(stats *ping.Statistics) {
		log.Tracef("\n--- %s ping statistics ---\n", stats.Addr)
		log.Tracef("%d packets transmitted, %d packets received, %v%% packet loss\n", stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
	}

	// required for linux and windows: https://github.com/sparrc/go-ping/issues/4
	if runtime.GOOS == "linux" || runtime.GOOS == "windows" {
		pinger.SetPrivileged(true)
	}

	errCh := make(chan error)
	go func() {
		errCh <- pinger.Run()
	}()

	select {
	case <-ctx.Done():
		pinger.Stop()
		<-errCh
		return pinger.Statistics(), ctx.Err()
	case err := <-errCh:
		return pinger.Statistics(), fmt.Errorf("pinger.Run: %v", err)
	}
}
