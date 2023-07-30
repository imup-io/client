package connectivity

import (
	"context"
	"os"
	"runtime"
	"syscall"

	"net"
	"time"

	log "golang.org/x/exp/slog"
)

type dialCollector struct {
	avoidAddrs map[string]bool

	clientVersion string
	count         int
	debug         bool
	port          string
	connected     int

	delay    time.Duration
	interval time.Duration
	timeout  time.Duration
}

func NewDialerCollector(opts Options) StatCollector {
	return &dialCollector{
		avoidAddrs:    map[string]bool{},
		clientVersion: opts.ClientVersion,
		count:         opts.Count,
		connected:     0,
		debug:         opts.Debug,
		port:          "53",
		delay:         opts.Delay,
		interval:      opts.Interval,
		timeout:       opts.Timeout,
	}
}

// Interval is the time to wait between dialer tests
func (d *dialCollector) Interval() time.Duration {
	return d.interval
}

// Collect takes a list of address' to test against and collects connectivity statistics once per Interval.
func (d *dialCollector) Collect(ctx context.Context, pingAddrs []string) []Statistics {
	address := pingAddress(pingAddrs, d.avoidAddrs)

	d.checkConnectivity(ctx, address)
	log.Debug("check connectivity", "result", d.connected)
	if d.connected < 0 {
		log.Info("unable to verify connectivity, avoid ip next check", "address", address)
		// avoid current ping addr for next attempt
		d.avoidAddrs[address] = true
	}

	return []Statistics{
		{
			PingAddress:   address,
			Success:       d.connected > 0,
			TimeStamp:     time.Now().UnixNano(),
			ClientVersion: d.clientVersion,
			OS:            runtime.GOOS,
		},
	}
}

// DetectDowntime only increments downtime if Success is false but Internal Success is true
func (d *dialCollector) DetectDowntime(data []Statistics) (bool, int) {
	if len(data) == 0 {
		return false, 0
	}

	changed := false
	downtime := 0

	lastStatus := (data)[0].Success
	for _, p := range data {
		if !p.Success {
			downtime++
		}

		if p.Success != lastStatus {
			changed = true
		}
		lastStatus = p.Success
	}

	return changed, downtime
}

// checkConnectivity tests TCP connectivity for a given address
func (d *dialCollector) checkConnectivity(ctx context.Context, addr string) {
	ticker := time.NewTicker(d.delay)
	defer ticker.Stop()

	// blocks until finished unless canceled
	ticks := 0
	for ticks <= d.count || ctx.Err() != nil {
		select {
		case <-ticker.C:
			ticks++

			connected, err := d.run(addr)
			if err != nil {
				log.Error("cannot check connectivity", "error", err)
			}

			d.connected += connected

		case <-ctx.Done():
			log.Debug("shutdown detected, canceling connectivity check")
			return
		}
	}
}

// run returns connection status, if a conn cannot be established it will return an error
func (d *dialCollector) run(addr string) (int, error) {
	timeout := d.timeout
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(addr, d.port), timeout)

	if d.debug && err != nil {
		// additional detail around dialer errors
		if netErr, ok := err.(*net.OpError); ok {
			switch nestErr := netErr.Err.(type) {
			case *net.DNSError:
				log.Warn("dialer failed with net.DNSError")
			case *os.SyscallError:
				if nestErr.Err == syscall.ECONNREFUSED {
					log.Warn("dialer failed with syscall.ECONNREFUSED")
				}
				log.Warn("dialer failed with syscall.ECONNREFUSED")
			}
			if netErr.Timeout() {
				log.Warn("connection failed with timeout")
			}
		} else if err == context.Canceled || err == context.DeadlineExceeded {
			log.Warn("connection failed with timeout")
		}
	}

	if conn != nil {
		defer conn.Close()
		return 1, nil
	}

	return -1, err
}
