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

type DialerOptions struct {
	avoidAddrs map[string]bool

	ClientVersion string
	Count         int
	Debug         bool
	Port          string
	Connected     int

	Delay        time.Duration
	DialInterval time.Duration
	Timeout      time.Duration
}

func NewDialerCollector(opts DialerOptions) StatCollector {
	return &DialerOptions{
		avoidAddrs:   map[string]bool{},
		Count:        opts.Count,
		Debug:        opts.Debug,
		Port:         "53",
		Connected:    0,
		Delay:        opts.Delay,
		DialInterval: opts.DialInterval,
		Timeout:      opts.Timeout,
	}
}

// Interval is the time to wait between dialer tests
func (d *DialerOptions) Interval() time.Duration {
	return d.DialInterval
}

// Collect takes a list of address' to test against and collects connectivity statistics once per Interval.
func (d *DialerOptions) Collect(ctx context.Context, pingAddrs []string) []Statistics {
	address := pingAddress(pingAddrs, d.avoidAddrs)

	d.checkConnectivity(ctx, address)
	log.Debug("check connectivity", "result", d.Connected)
	if d.Connected < 0 {
		log.Info("unable to verify connectivity, avoid ip next check", "address", address)
		// avoid current ping addr for next attempt
		d.avoidAddrs[address] = true
	}

	return []Statistics{
		{
			PingAddress:   address,
			Success:       d.Connected > 0,
			TimeStamp:     time.Now().UnixNano(),
			ClientVersion: d.ClientVersion,
			OS:            runtime.GOOS,
		},
	}
}

// DetectDowntime only increments downtime if Success is false but Internal Success is true
func (d *DialerOptions) DetectDowntime(data []Statistics) (bool, int) {
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
func (d *DialerOptions) checkConnectivity(ctx context.Context, addr string) {
	ticker := time.NewTicker(d.Delay)
	defer ticker.Stop()

	// blocks until finished unless canceled
	ticks := 0
	for ticks <= d.Count || ctx.Err() != nil {
		select {
		case <-ticker.C:
			ticks++

			connected, err := d.run(addr)
			if err != nil {
				log.Error("cannot check connectivity", "error", err)
			}

			d.Connected += connected

		case <-ctx.Done():
			log.Debug("shutdown detected, canceling connectivity check")
			return
		}
	}
}

// run returns connection status, if a conn cannot be established it will return an error
func (d *DialerOptions) run(addr string) (int, error) {
	timeout := d.Timeout
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(addr, d.Port), timeout)

	if d.Debug && err != nil {
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
