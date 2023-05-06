package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestPing(t *testing.T) {
	cases := []struct {
		Name   string
		ApiKey string
		Email  string
		HostID string
		Ping   string
	}{
		{Name: "org", ApiKey: "1234", Email: "org-test@example.com", HostID: "org-based-host", Ping: "true"},
		{Name: "user", ApiKey: "", Email: "test@example.com", HostID: "email-based-host", Ping: "true"},
	}

	for _, c := range cases {
		os.Clearenv()
		os.Setenv("API_KEY", c.ApiKey)
		os.Setenv("EMAIL", c.Email)
		os.Setenv("HOST_ID", c.HostID)
		os.Setenv("PING_ENABLED", c.Ping)
		os.Setenv("PING_INTERVAL", "1")
		os.Setenv("PING_REQUESTS", "2")
		os.Setenv("PING_DELAY", "100")
		os.Setenv("PING_ADDRESS_INTERNAL", "127.0.0.1")

		t.Run(fmt.Sprintf("test testCollectPingData for %s", c.Name), testCollectPingData())
		t.Run(fmt.Sprintf("test testSendPingData for %s", c.Name), testSendPingData())
		t.Run(fmt.Sprintf("test testPing for %s", c.Name), testPing())
		t.Run(fmt.Sprintf("test testNoPingExternal for %s", c.Name), testNoPingExternal())
		t.Run(fmt.Sprintf("test testNoPing for %s", c.Name), testNoPinger())
	}
}

func testCollectPingData() func(t *testing.T) {
	return func(t *testing.T) {
		is := is.New(t)

		testURL := "8.8.8.8"

		os.Setenv("PING_ADDRESS", testURL)
		imup := newApp()

		ping := imup.newPingStats()
		data := ping.Collect(context.Background(), []string{testURL, "8.4.4.8"})
		is.True(len(data) == 1)
		for _, v := range data {
			is.True(v.Success)
			is.Equal(v.PacketsSent, imup.cfg.PingRequestsCount())
			is.Equal(v.PacketsRecv, imup.cfg.PingRequestsCount())
		}
	}
}

func testSendPingData() func(t *testing.T) {
	return func(t *testing.T) {
		s := defaultApiServer()
		defer s.Close()
		testURL, _ := url.Parse(s.URL)

		os.Setenv("IMUP_ADDRESS", testURL.String())
		imup := newApp()

		data := imupData{
			Downtime:      0,
			StatusChanged: false,
			Email:         imup.cfg.EmailAddress(),
			ID:            imup.cfg.HostID(),
			Key:           imup.cfg.APIKey(),
			GroupID:       imup.cfg.GroupID(),
			IMUPData:      []pingStats{},
		}

		sendImupData(context.Background(), sendDataJob{imup.cfg.PostConnectionData(), data})
	}
}

func testPing() func(t *testing.T) {
	return func(t *testing.T) {
		is := is.New(t)

		imup := newApp()

		ps := imup.newPingStats()
		pt, ok := ps.(*pingTest)
		is.True(ok)
		pinger, err := pt.setupExternalPinger(context.Background(), []string{"127.0.0.1"}, nil)
		is.NoErr(err)

		pinger.Timeout = time.Duration(imup.cfg.PingIntervalSeconds()) * time.Second
		pinger.Interval = time.Duration(imup.cfg.PingDelayMilli()) * time.Millisecond
		pinger.Count = imup.cfg.PingRequestsCount()
		// successful run returns non nil error
		stats, _ := pt.run(context.Background(), pinger)

		is.True(stats.PacketsSent > 0)
		is.True(stats.PacketsRecv > 0)
	}
}

func testNoPingExternal() func(t *testing.T) {
	return func(t *testing.T) {
		is := is.New(t)
		imup := newApp()

		// https://superuser.com/questions/698244/ip-address-that-is-the-equivalent-of-dev-null
		wg := sync.WaitGroup{}
		wg.Add(1)
		data := []pingStats{}
		ps := imup.newPingStats()

		go func() {
			defer wg.Done()
			data = ps.Collect(context.Background(), []string{"240.0.0.0"})
		}()

		wg.Wait()

		for _, v := range data {
			fmt.Println(v)
			is.True(v.SuccessInternal)
		}

		changed, dt := ps.DetectDowntime(data)
		is.True(changed == false)
		is.Equal(dt, 0)

	}
}

func testNoPinger() func(t *testing.T) {
	return func(t *testing.T) {
		is := is.New(t)

		// https://superuser.com/questions/698244/ip-address-that-is-the-equivalent-of-dev-null
		os.Setenv("IMUP_ADDRESS", "240.0.0.0")
		imup := newApp()

		ps := imup.newPingStats()
		pt, ok := ps.(*pingTest)
		is.True(ok)

		pinger, err := pt.setupExternalPinger(context.Background(), []string{"240.0.0.0"}, nil)
		is.True(err != nil)
		is.True(pinger == nil)
	}
}
