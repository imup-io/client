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
		Connected         bool
		ConnectedInternal bool
		Downtime          bool
		ExternalPingAddrs []string
		Opts              connectivity.PingOptions
	}{
		{
			Name:              "connectivity-external-ping",
			ExternalPingAddrs: []string{"8.8.8.8", "8.0.0.8"},
			Connected:         true,
			ConnectedInternal: true,
			Downtime:          false,
			Opts: connectivity.PingOptions{
				AddressInternal: "",
				AvoidAddrs:      map[string]bool{},
				Count:           2,
				Debug:           false,
				Delay:           time.Duration(100) * time.Millisecond,
				PingInterval:    time.Duration(1) * time.Second,
				Timeout:         time.Duration(1) * time.Second,
			},
		},
		{
			Name:              "connectivity-no-internal-ping",
			ExternalPingAddrs: []string{"240.0.0.0"},
			Connected:         false,
			ConnectedInternal: false,
			Downtime:          false,
			Opts: connectivity.PingOptions{
				AddressInternal: "240.0.0.0",
				AvoidAddrs:      map[string]bool{},
				Count:           2,
				Debug:           false,
				Delay:           time.Duration(100) * time.Millisecond,
				PingInterval:    time.Duration(1) * time.Second,
				Timeout:         time.Duration(1) * time.Second,
			},
		},
		{
			Name:              "no-connectivity-internal-ping",
			ExternalPingAddrs: []string{"240.0.0.0"},
			Connected:         false,
			ConnectedInternal: true,
			Downtime:          true,
			Opts: connectivity.PingOptions{
				AddressInternal: "127.0.0.1",
				AvoidAddrs:      map[string]bool{},
				Count:           2,
				Debug:           false,
				Delay:           time.Duration(100) * time.Millisecond,
				PingInterval:    time.Duration(1) * time.Second,
				Timeout:         time.Duration(1) * time.Second,
			},
		},
		{
			Name:              "no-connectivity-no-ping",
			ExternalPingAddrs: []string{"240.0.0.0"},
			Connected:         false,
			ConnectedInternal: false,
			Downtime:          false,
			Opts: connectivity.PingOptions{
				AddressInternal: "",
				AvoidAddrs:      map[string]bool{},
				Count:           2,
				Debug:           false,
				Delay:           time.Duration(100) * time.Millisecond,
				PingInterval:    time.Duration(1) * time.Second,
				Timeout:         time.Duration(1) * time.Second,
			},
		},
	}

	for _, c := range cases {
		os.Clearenv()
		// TODO: include a test from run_test that sets this env var
		// os.Setenv("API_KEY", c.ApiKey)
		// os.Setenv("EMAIL", c.Email)
		// os.Setenv("HOST_ID", c.HostID)
		// os.Setenv("PING_ENABLED", c.Ping)
		// os.Setenv("PING_INTERVAL", "1")
		// os.Setenv("PING_REQUESTS", "2")
		// os.Setenv("PING_DELAY", "100")
		// os.Setenv("PING_ADDRESS_INTERNAL", "127.0.0.1")
		// os.Setenv("PING_ADDRESS", testURL)
		t.Run(fmt.Sprintf("test testCollectPingData for %s", c.Name), testCollectPingData(c.ExternalPingAddrs, c.Connected, c.ConnectedInternal, c.Downtime, c.Opts))
	}
}

func testCollectPingData(externalPingAddrs []string, connected, connectedInternal, downtime bool, opts connectivity.PingOptions) func(t *testing.T) {
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

// TODO: move to imup_test
// func testSendPingData() func(t *testing.T) {
// 	return func(t *testing.T) {
// 		s := defaultApiServer()
// 		defer s.Close()
// 		testURL, _ := url.Parse(s.URL)

// 		os.Setenv("IMUP_ADDRESS", testURL.String())
// 		imup := newApp()

// 		data := imupData{
// 			Downtime:      0,
// 			StatusChanged: false,
// 			Email:         imup.cfg.EmailAddress(),
// 			ID:            imup.cfg.HostID(),
// 			Key:           imup.cfg.APIKey(),
// 			GroupID:       imup.cfg.GroupID(),
// 			IMUPData:      []pingStats{},
// 		}

// 		sendImupData(context.Background(), sendDataJob{imup.APIPostConnectionData, data})
// 	}
// }
