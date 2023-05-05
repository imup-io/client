package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/matryer/is"
)

func TestConn(t *testing.T) {
	cases := []struct {
		Name   string
		ApiKey string
		Email  string
		HostID string
	}{
		{Name: "org", ApiKey: "1234", Email: "org-test@example.com", HostID: "org-based-host"},
		{Name: "user", ApiKey: "", Email: "test@example.com", HostID: "email-based-host"},
	}

	for _, c := range cases {
		os.Clearenv()
		os.Setenv("API_KEY", c.ApiKey)
		os.Setenv("EMAIL", c.Email)
		os.Setenv("HOST_ID", c.HostID)
		os.Setenv("CONN_INTERVAL", "1")
		os.Setenv("CONN_REQUESTS", "2")
		os.Setenv("CONN_DELAY", "100")

		t.Run(fmt.Sprintf("testing testCollectConnData for %s", c.Name), testCollectConnData())
		t.Run(fmt.Sprintf("testing testSendConnData for %s", c.Name), testSendConnData())
		t.Run(fmt.Sprintf("testing testNoConnData for %s", c.Name), testNoConnData())
	}
}

func testCollectConnData() func(t *testing.T) {
	return func(t *testing.T) {
		is := is.New(t)
		imup := newApp()

		testURL := "1.0.0.1"

		dialer := imup.newDialerStats()
		data := dialer.Collect(context.Background(), []string{testURL})

		for _, v := range data {
			is.True(v.Success)
		}
		_, dt := dialer.DetectDowntime(data)
		is.Equal(dt, 0)
	}
}

func testSendConnData() func(t *testing.T) {
	return func(t *testing.T) {
		s := defaultApiServer()
		defer s.Close()
		testURL, _ := url.Parse(s.URL)
		os.Setenv("IMUP_ADDRESS", testURL.String())

		imup := newApp()

		data := imupData{
			Downtime:      0,
			StatusChanged: false,
			Email:         "email@test.com",
			ID:            "",
			Key:           "",
			IMUPData:      []pingStats{},
		}

		sendImupData(context.Background(), sendDataJob{imup.cfg.PostConnectionData(), data})
	}
}

func testNoConnData() func(t *testing.T) {
	return func(t *testing.T) {
		is := is.New(t)

		imup := newApp()

		testURL := "240.0.0.0"

		conn := imup.newDialerStats()
		data := conn.Collect(context.Background(), []string{testURL})

		for _, v := range data {
			is.True(!v.Success)
		}

		sc, dt := conn.DetectDowntime(data)
		is.True(!sc)
		is.Equal(dt, 1)
	}
}
