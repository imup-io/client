package connectivity_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/imup-io/client/connectivity"
	"github.com/matryer/is"
)

func TestDial(t *testing.T) {
	cases := []struct {
		Name          string
		Connected     bool
		Downtime      bool
		ExternalAddrs []string
		Opts          connectivity.Options
	}{
		{
			Name:          "no-connectivity",
			ExternalAddrs: []string{"240.0.0.0"},
			Downtime:      true,
			Connected:     false,
			Opts: connectivity.Options{
				Count:    2,
				Debug:    false,
				Delay:    time.Duration(100) * time.Millisecond,
				Interval: time.Duration(1) * time.Second,
				Timeout:  time.Duration(1) * time.Second,
			},
		},
		{
			Name:          "connectivity",
			ExternalAddrs: []string{"8.8.8.8, 8.0.0.8"},
			Downtime:      true,
			Connected:     false,
			Opts: connectivity.Options{
				Count:    2,
				Debug:    false,
				Delay:    time.Duration(100) * time.Millisecond,
				Interval: time.Duration(1) * time.Second,
				Timeout:  time.Duration(1) * time.Second,
			},
		},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("testing testNoConnData for %s", c.Name), testNoConnData(c.Connected, c.Downtime, c.ExternalAddrs, c.Opts))
	}
}

func testNoConnData(connected, downtime bool, externalAddrs []string, opts connectivity.Options) func(t *testing.T) {
	return func(t *testing.T) {
		is := is.New(t)

		dialer := connectivity.NewDialerCollector(opts)
		data := dialer.Collect(context.Background(), externalAddrs)
		_, dt := dialer.DetectDowntime(data)
		if downtime {
			is.Equal(dt, 1)
		} else {
			is.Equal(dt, 0)
		}

		for _, v := range data {
			is.Equal(connected, v.Success)
		}
	}
}
