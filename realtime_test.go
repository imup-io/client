package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"
	"testing"

	"github.com/matryer/is"
)

func TestRealTime(t *testing.T) {
	cases := []struct {
		Name     string
		ApiKey   string
		Email    string
		HostID   string
		Realtime string
		Status   string
		Version  string
		RetCode  int
	}{
		{Name: "org", ApiKey: "1234", Email: "org-test@example.com", HostID: "org-based-host", Realtime: "true", Status: "running", Version: "unknown", RetCode: http.StatusNoContent},
		{Name: "user", ApiKey: "", Email: "test@example.com", HostID: "email-based-host", Realtime: "true", Version: "unknown", RetCode: http.StatusNoContent},
	}

	for _, c := range cases {
		os.Clearenv()
		os.Setenv("API_KEY", c.ApiKey)
		os.Setenv("EMAIL", c.Email)
		os.Setenv("HOST_ID", c.HostID)
		os.Setenv("REALTIME", c.Realtime)

		t.Run(c.Name, testClientHealthy())
		t.Run(c.Name, testShouldRunSpeedTest())
		t.Run(c.Name, testShouldRunSpeedTestErrors())
		t.Run(c.Name, testPostSpeedTestResult())
		t.Run(c.Name, testRealTimeAuthorized(c.Status))
		t.Run(c.Name, testRemoteConfig(c.RetCode))
		t.Run(c.Name, testRemoteUpdate(c.Name, c.Version))
	}
}

func testClientHealthy() func(t *testing.T) {
	return func(t *testing.T) {
		is := is.New(t)

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("ok"))
		}))
		defer s.Close()

		testURL, _ := url.Parse(s.URL)
		os.Setenv("IMUP_LIVENESS_CHECKIN_ADDRESS", testURL.String())

		imup := newApp()
		err := imup.sendClientHealthy(context.Background())
		is.NoErr(err)
	}
}

func testShouldRunSpeedTest() func(t *testing.T) {
	return func(t *testing.T) {
		is := is.New(t)

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"success":true,"data":true}`))
		}))
		defer s.Close()
		testURL, _ := url.Parse(s.URL)

		os.Setenv("IMUP_SHOULD_RUN_SPEEDTEST_ADDRESS", testURL.String())
		os.Setenv("IMUP_SPEED_TEST_STATUS_ADDRESS", testURL.String())
		os.Setenv("IMUP_SPEED_TEST_RESULTS_ADDRESS", testURL.String())

		imup := newApp()

		ok, err := imup.shouldRunSpeedtest(context.Background())

		is.NoErr(err)
		is.True(ok)

		err = imup.postSpeedTestRealtimeStatus(context.Background(), "running")
		is.NoErr(err)
	}
}

func testShouldRunSpeedTestErrors() func(t *testing.T) {
	return func(t *testing.T) {
		is := is.New(t)

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"success":false,"data":false}`)
		}))
		defer s.Close()

		testURL, _ := url.Parse(s.URL)
		os.Setenv("IMUP_SPEED_TEST_STATUS_ADDRESS", testURL.String())
		os.Setenv("IMUP_SHOULD_RUN_SPEEDTEST_ADDRESS", testURL.String())
		os.Setenv("IMUP_SPEED_TEST_RESULTS_ADDRESS", testURL.String())

		imup := newApp()
		ok, err := imup.shouldRunSpeedtest(context.Background())
		is.NoErr(err)
		is.Equal(ok, false)

		err = imup.postSpeedTestRealtimeStatus(context.Background(), "error")
		is.NoErr(err)
	}
}

func testPostSpeedTestResult() func(t *testing.T) {
	return func(t *testing.T) {
		is := is.New(t)

		s := defaultApiServer()
		defer s.Close()
		testURL, _ := url.Parse(s.URL)

		os.Setenv("IMUP_ADDRESS_SPEEDTEST", testURL.String())
		os.Setenv("IMUP_SPEED_TEST_RESULTS_ADDRESS", testURL.String())

		imup := newApp()
		tr := &speedTestData{UploadMbps: 1.0, DownloadMbps: 1.0}
		err := imup.postSpeedTestRealtimeResults(context.Background(), "complete", tr)
		is.NoErr(err)
	}
}

// RealtimeAuthorized
// v1/auth/realtimeAuthorized
// IMUP_REALTIME_AUTHORIZED
func testRealTimeAuthorized(status string) func(t *testing.T) {
	return func(t *testing.T) {
		is := is.New(t)

		s := defaultApiServer()
		defer s.Close()
		testURL, _ := url.Parse(s.URL)

		os.Setenv("IMUP_REALTIME_AUTHORIZED", testURL.String())

		imup := newApp()

		data := &authRequest{Email: imup.cfg.EmailAddress(), Key: imup.cfg.APIKey()}
		b, err := json.Marshal(data)
		is.NoErr(err)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			err = imup.authorized(context.Background(), bytes.NewBuffer(b), imup.RealtimeAuthorized)
		}()
		wg.Wait()

		is.NoErr(err)
	}
}

func testRemoteConfig(retcode int) func(t *testing.T) {
	return func(t *testing.T) {
		is := is.New(t)

		s := defaultConfigurableApiServer(retcode)
		defer s.Close()
		testURL, _ := url.Parse(s.URL)

		os.Setenv("IMUP_REALTIME_CONFIG", testURL.String())

		imup := newApp()
		imup.cfg.DisableRealtime()

		wg := sync.WaitGroup{}
		wg.Add(1)
		var err error
		go func() {
			defer wg.Done()
			err = imup.remoteConfigReload(context.Background())
		}()

		wg.Wait()
		is.NoErr(err)
	}
}

func testRemoteUpdate(name, version string) func(t *testing.T) {
	return func(t *testing.T) {
		is := is.New(t)

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			type cfg struct {
				ID            string `json:"ID"`
				Email         string `json:"Email"`
				Environment   string `json:"Environment"`
				Key           string `json:"Key"`
				ConfigVersion string `json:"Version"`
				// TODO: implement an app wide file logger
				// LogDirectory string `json:"LogDirectory"`

				InsecureSpeedTest bool `json:"InsecureSpeedTest"`
				NoDiscoverGateway bool `json:"NODiscoverGateway"`
				PingEnabled       bool `json:"PingEnabled"`
				RealtimeEnabled   bool `json:"RealtimeEnabled"`
				SpeedTestEnabled  bool `json:"SpeedTestEnabled"`
			}

			data := struct {
				C cfg `json:"config"`
			}{
				C: cfg{ConfigVersion: version, Email: "email", RealtimeEnabled: true},
			}

			err := json.NewEncoder(w).Encode(data)
			is.NoErr(err)
			w.WriteHeader(http.StatusOK)
		}))

		defer s.Close()
		testURL, _ := url.Parse(s.URL)

		os.Setenv("IMUP_REALTIME_CONFIG", testURL.String())

		imup := newApp()
		imup.cfg.DisableRealtime()

		wg := sync.WaitGroup{}
		wg.Add(1)
		var err error
		go func() {
			defer wg.Done()
			err = imup.remoteConfigReload(context.Background())
		}()

		wg.Wait()

		if name == "org" {
			is.NoErr(err)
			is.True(imup.cfg.Realtime())
		} else {
			is.NoErr(err)
			is.Equal(imup.cfg.Realtime(), false)
		}
	}
}
