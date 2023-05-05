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

func TestDial(t *testing.T) {
	cases := []struct {
		Name          string
		Connected     bool
		Downtime      bool
		ExternalAddrs []string
		Opts          connectivity.DialerOptions
	}{
		{
			Name:          "no-connectivity",
			ExternalAddrs: []string{"240.0.0.0"},
			Downtime:      true,
			Connected:     false,
			Opts: connectivity.DialerOptions{
				AvoidAddrs:   map[string]bool{},
				Count:        2,
				Debug:        false,
				Port:         "53",
				Connected:    0,
				Delay:        time.Duration(100) * time.Millisecond,
				DialInterval: time.Duration(1) * time.Second,
				Timeout:      time.Duration(1) * time.Second,
			},
		},
		{
			Name:          "connectivity",
			ExternalAddrs: []string{"8.8.8.8, 8.0.0.8"},
			Downtime:      true,
			Connected:     false,
			Opts: connectivity.DialerOptions{
				AvoidAddrs:   map[string]bool{},
				Count:        2,
				Debug:        false,
				Port:         "53",
				Connected:    0,
				Delay:        time.Duration(100) * time.Millisecond,
				DialInterval: time.Duration(1) * time.Second,
				Timeout:      time.Duration(1) * time.Second,
			},
		},
	}

	for _, c := range cases {
		os.Clearenv()
		// TODO: include a test from run_test that sets this env var
		// os.Setenv("API_KEY", c.ApiKey)
		// os.Setenv("EMAIL", c.Email)
		// os.Setenv("HOST_ID", c.HostID)
		// os.Setenv("CONN_INTERVAL", "1")
		// os.Setenv("CONN_REQUESTS", "2")
		// os.Setenv("CONN_DELAY", "100")

		t.Run(fmt.Sprintf("testing testNoConnData for %s", c.Name), testNoConnData(c.Connected, c.Downtime, c.ExternalAddrs, c.Opts))
	}
}

// // TODO: move to imup_test
// func testSendConnData() func(t *testing.T) {
// 	return func(t *testing.T) {
// 		s := defaultApiServer()
// 		defer s.Close()
// 		testURL, _ := url.Parse(s.URL)
// 		os.Setenv("IMUP_ADDRESS", testURL.String())

// 		imup := newApp()

// 		data := imupData{
// 			Downtime:      0,
// 			StatusChanged: false,
// 			Email:         "email@test.com",
// 			ID:            "",
// 			Key:           "",
// 			IMUPData:      []pingStats{},
// 		}

// 		sendImupData(context.Background(), sendDataJob{imup.APIPostConnectionData, data})
// 	}
// }

func testNoConnData(connected, downtime bool, externalAddrs []string, opts connectivity.DialerOptions) func(t *testing.T) {
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
