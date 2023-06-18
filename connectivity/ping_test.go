package connectivity_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/imup-io/client/connectivity"
	"github.com/matryer/is"
)

func TestPing(t *testing.T) {
	cases := []struct {
		Name              string
		CI                bool
		Connected         bool
		ConnectedInternal bool
		Downtime          bool
		ExternalPingAddrs []string
		Opts              connectivity.Options
	}{
		{
			Name:              "connectivity-external-ping",
			ExternalPingAddrs: []string{"8.8.8.8", "8.0.0.8"},
			CI:                false,
			Connected:         true,
			ConnectedInternal: true,
			Downtime:          false,
			Opts: connectivity.Options{
				AddressInternal: "",
				Count:           2,
				Debug:           false,
				Delay:           time.Duration(100) * time.Millisecond,
				Interval:        time.Duration(1) * time.Second,
				Timeout:         time.Duration(1) * time.Second,
			},
		},
		{
			Name:              "connectivity-no-internal-ping",
			ExternalPingAddrs: []string{"240.0.0.0"},
			CI:                true,
			Connected:         false,
			ConnectedInternal: false,
			Downtime:          false,
			Opts: connectivity.Options{
				AddressInternal: "240.0.0.0",
				Count:           2,
				Debug:           false,
				Delay:           time.Duration(100) * time.Millisecond,
				Interval:        time.Duration(1) * time.Second,
				Timeout:         time.Duration(1) * time.Second,
			},
		},
		{
			Name:              "no-connectivity-internal-ping",
			CI:                false,
			ExternalPingAddrs: []string{"240.0.0.0"},
			Connected:         false,
			ConnectedInternal: true,
			Downtime:          true,
			Opts: connectivity.Options{
				AddressInternal: "127.0.0.1",
				Count:           2,
				Debug:           false,
				Delay:           time.Duration(100) * time.Millisecond,
				Interval:        time.Duration(1) * time.Second,
				Timeout:         time.Duration(1) * time.Second,
			},
		},
		{
			Name:              "no-connectivity-no-ping",
			ExternalPingAddrs: []string{"240.0.0.0"},
			CI:                true,
			Connected:         false,
			ConnectedInternal: false,
			Downtime:          false,
			Opts: connectivity.Options{
				AddressInternal: "",
				Count:           2,
				Debug:           false,
				Delay:           time.Duration(100) * time.Millisecond,
				Interval:        time.Duration(1) * time.Second,
				Timeout:         time.Duration(1) * time.Second,
			},
		},
	}

	for _, c := range cases {
		// do not run integration test in ci
		if _, ok := os.LookupEnv("CI"); !ok {
			t.Run(fmt.Sprintf("test testCollectPingData for %s", c.Name), testCollectPingData(c.ExternalPingAddrs, c.Connected, c.ConnectedInternal, c.Downtime, c.Opts))
		} else if c.CI {
			t.Run(fmt.Sprintf("test testCollectPingData for %s", c.Name), testCollectPingData(c.ExternalPingAddrs, c.Connected, c.ConnectedInternal, c.Downtime, c.Opts))
		}
	}
}

func testCollectPingData(externalPingAddrs []string, connected, connectedInternal, downtime bool, opts connectivity.Options) func(t *testing.T) {
	return func(t *testing.T) {
		is := is.New(t)

		ping := connectivity.NewPingCollector(opts)
		data := ping.Collect(context.Background(), externalPingAddrs)
		is.True(len(data) >= 1)
		_, dt := ping.DetectDowntime(data)
		if downtime {
			is.Equal(dt, 1)
		} else {
			is.Equal(dt, 0)
		}

		for _, stats := range data {
			if stats.EndpointType == "external" {
				is.Equal(connected, stats.Success)
			}
			if stats.EndpointType == "internal" {
				is.Equal(connectedInternal, stats.SuccessInternal)
			}

			if stats.Success {
				is.Equal(stats.PacketsSent, opts.Count)
				is.Equal(stats.PacketsRecv, opts.Count)
				is.Equal(stats.SuccessInternal, true)
			}
		}
	}
}
