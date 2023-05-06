package main

import (
	"context"
	"os"
	"runtime"
	"syscall"

	"net"
	"time"

	log "golang.org/x/exp/slog"
)

type dialer struct {
	address    string
	avoidAddrs map[string]bool
	count      int
	debug      bool
	port       string
	connected  int

	delay    time.Duration
	interval time.Duration
	timeout  time.Duration
}

func (i *imup) newDialerStats() imupStatCollector {
	return &dialer{
		avoidAddrs: i.PingAddressesAvoid,
		count:      i.cfg.ConnRequestsCount(),
		debug:      i.cfg.Verbosity() == log.LevelDebug,
		port:       "53",
		connected:  0,
		delay:      time.Duration(i.cfg.ConnDelayMilli()) * time.Millisecond,
		interval:   time.Duration(i.cfg.ConnIntervalSeconds()) * time.Second,
		timeout:    time.Duration(i.cfg.ConnIntervalSeconds()) * time.Second,
	}
}

// Interval is the time to wait between dialer tests
func (d *dialer) Interval() time.Duration {
	return d.interval
}

// Collect takes a list of address' to test against and collects connectivity statistics once per Interval.
func (d *dialer) Collect(ctx context.Context, pingAddrs []string) []pingStats {
	d.address = pingAddress(pingAddrs, d.avoidAddrs)

	d.checkConnectivity(ctx)
	log.Debug("check connectivity", "result", d.connected)
	if d.connected < 0 {
		log.Info("unable to verify connectivity, avoid ip next check", "address", d.address)
		// avoid current ping addr for next attempt
		d.avoidAddrs[d.address] = true
	}

	return []pingStats{
		{
			PingAddress:   d.address,
			Success:       d.connected > 0,
			TimeStamp:     time.Now().UnixNano(),
			ClientVersion: ClientVersion,
			OS:            runtime.GOOS,
		},
	}
}

// DetectDowntime only increments downtime if Success is false but Internal Success is true
func (d *dialer) DetectDowntime(data []pingStats) (bool, int) {
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
func (d *dialer) checkConnectivity(ctx context.Context) {
	ticker := time.NewTicker(d.delay)
	defer ticker.Stop()

	// blocks until finished unless canceled
	ticks := 0
	for ticks <= d.count || ctx.Err() != nil {
		select {
		case <-ticker.C:
			ticks++

			connected, err := d.run()
			if err != nil {
				log.Error("error", err)
			}

			d.connected += connected

		case <-ctx.Done():
			log.Debug("shutdown detected, canceling connectivity check")
			return
		}
	}
}

// run returns connection status, if a conn cannot be established it will return an error
func (d *dialer) run() (int, error) {
	timeout := d.timeout
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(d.address, d.port), timeout)

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
