package connectivity

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"sync"
	"time"

	ping "github.com/prometheus-community/pro-bing"
	log "golang.org/x/exp/slog"
)

type pingCollector struct {
	avoidAddrs map[string]bool

	addressInternal string
	clientVersion   string
	count           int
	debug           bool

	delay    time.Duration
	interval time.Duration
	timeout  time.Duration
}

func NewPingCollector(opts Options) StatCollector {
	return &pingCollector{
		avoidAddrs:      map[string]bool{},
		addressInternal: opts.AddressInternal,
		count:           opts.Count,
		clientVersion:   opts.ClientVersion,
		debug:           opts.Debug,
		delay:           opts.Delay,
		interval:        opts.Interval,
		timeout:         opts.Timeout,
	}
}

// Interval is the time to wait between ping testing
func (p *pingCollector) Interval() time.Duration {
	return p.interval
}

// Collect takes a list of address' to test against and collects ping statistics once per Interval.
func (p *pingCollector) Collect(ctx context.Context, pingAddrs []string) []Statistics {
	externalPingResult := Statistics{}
	internalPingResult := Statistics{}
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

	// internal testing disabled, return external result only regardless of result
	if p.addressInternal == "" {
		return []Statistics{externalPingResult}
	}

	// connectivity established, return result of both external and internal test result
	if success {
		return []Statistics{internalPingResult, externalPingResult}
	}

	log.Info("external endpoint unreachable", "internal-result", internalSuccess, "gateway-address", p.addressInternal)

	// external data needs to send last for tests to pingDownTimeDetect to work properly
	externalPingResult.SuccessInternal = internalSuccess
	return []Statistics{internalPingResult, externalPingResult}
}

// DetectDowntime only increments downtime if Success is false but Internal Success is true
// demonstrating a connection to the gateway is not the problem
func (p *pingCollector) DetectDowntime(data []Statistics) (bool, int) {
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
func (p *pingCollector) checkConnectivity(ctx context.Context, testType string, pingAddrs []string, timestamp int64) (Statistics, bool) {
	var err error
	var pinger *ping.Pinger
	if testType == "external" {
		if pinger, err = p.setupExternalPinger(ctx, pingAddrs, nil); err != nil {
			log.Error("failed to setup external pinger", "error", err)
			return Statistics{PacketsSent: 0, PacketsRecv: 0, PacketLoss: 100.0, TimeStamp: timestamp, EndpointType: testType}, false
		}
	} else {
		if pinger, err = p.setupInternalPinger(ctx, p.addressInternal, nil); err != nil {
			log.Error("failed to setup internal pinger", "error", err)
			return Statistics{PacketsSent: 0, PacketsRecv: 0, PacketLoss: 100.0, TimeStamp: timestamp, EndpointType: testType}, false
		}
	}

	pinger.Timeout = p.timeout
	pinger.Interval = p.delay
	pinger.Count = p.count

	var info error
	var stats *ping.Statistics
	if stats, info = p.run(ctx, pinger); info != nil {
		log.Warn("error sending ping", "address", stats.Addr, "error", info)
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

	data := Statistics{
		PingAddress:     stats.Addr,
		Success:         success,
		SuccessInternal: successInternal,
		TimeStamp:       timestamp,
		ClientVersion:   p.clientVersion,
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
func (p *pingCollector) setupInternalPinger(ctx context.Context, pingAddr string, errs error) (*ping.Pinger, error) {
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
		log.Debug("successfully created pinger", "address", pinger.Addr())
		return pinger, nil
	}
}

// setupExternalPinger is a helper function that verifies an address is available to ping before performing a longer running test
// it will recursively exhaust the list of available addresses to test against before giving up
func (p *pingCollector) setupExternalPinger(ctx context.Context, pingAddrs []string, errs error) (*ping.Pinger, error) {
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
		log.Debug("avoiding pinging external endpoint next check", "address", randomPingAddr)
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
		log.Debug("successfully created pinger to test", "address", pinger.Addr())
		return pinger, nil
	}
}

// run always returns ping statistics and a non nil error
func (p *pingCollector) run(ctx context.Context, pinger *ping.Pinger) (*ping.Statistics, error) {
	pinger.Debug = true
	pinger.OnRecv = func(pkt *ping.Packet) {
		log.Debug("pinger onRecv", "stats", fmt.Sprintf("%d bytes sent %s: icmp_seq=%d time=%v", pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt))
	}

	pinger.OnFinish = func(stats *ping.Statistics) {
		log.Debug("pinger on finish", "stats", fmt.Sprintf("%d packets transmitted, %d packets received, %v%% packet loss", stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss))
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
