package realtime_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/imup-io/client/realtime"
	"github.com/imup-io/client/speedtesting"
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
		// {Name: "org", ApiKey: "1234", Email: "org-test@example.com", HostID: "org-based-host", Realtime: "false", Status: "running", Version: "1.0.0", RetCode: http.StatusOK},
		{Name: "user", ApiKey: "", Email: "test@example.com", HostID: "email-based-host", Realtime: "true", Version: "unknown", RetCode: http.StatusNoContent},
		// {Name: "user", ApiKey: "", Email: "test@example.com", HostID: "email-based-host", Realtime: "false", Version: "unknown", RetCode: http.StatusNoContent},
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
		t.Run(c.Name, testRealTimeAuthorized(c.ApiKey, c.Email, c.Status))
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
		// TODO: include a test from run_test that sets this env var
		// os.Setenv("IMUP_LIVENESS_CHECKIN_ADDRESS", testURL.String())

		err := realtime.SendClientHealthy(context.Background(), testURL.String())
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

		// TODO: include a test from run_test that sets this env var
		// os.Setenv("IMUP_SHOULD_RUN_SPEEDTEST_ADDRESS", testURL.String())
		// os.Setenv("IMUP_SPEED_TEST_STATUS_ADDRESS", testURL.String())
		// os.Setenv("IMUP_SPEED_TEST_RESULTS_ADDRESS", testURL.String())

		ok, err := realtime.ShouldRunSpeedtest(context.Background(), testURL.String())

		is.NoErr(err)
		is.True(ok)

		err = realtime.PostSpeedTestRealtimeStatus(context.Background(), "running", testURL.String())
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

		// TODO: include a test from run_test that sets this env var
		// os.Setenv("IMUP_SPEED_TEST_STATUS_ADDRESS", testURL.String())
		// os.Setenv("IMUP_SHOULD_RUN_SPEEDTEST_ADDRESS", testURL.String())
		// os.Setenv("IMUP_SPEED_TEST_RESULTS_ADDRESS", testURL.String())

		ok, err := realtime.ShouldRunSpeedtest(context.Background(), testURL.String())
		is.NoErr(err)
		is.Equal(ok, false)

		err = realtime.PostSpeedTestRealtimeStatus(context.Background(), "error", testURL.String())
		is.NoErr(err)
	}
}

func testPostSpeedTestResult() func(t *testing.T) {
	return func(t *testing.T) {
		is := is.New(t)

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer s.Close()
		testURL, _ := url.Parse(s.URL)

		// TODO: include a test from run_test that sets this env var
		// os.Setenv("IMUP_ADDRESS_SPEEDTEST", testURL.String())
		// os.Setenv("IMUP_SPEED_TEST_RESULTS_ADDRESS", testURL.String())

		tr := &speedtesting.SpeedTestData{UploadMbps: 1.0, DownloadMbps: 1.0}
		err := realtime.PostSpeedTestRealtimeResults(context.Background(), "complete", testURL.String(), tr.UploadMbps, tr.DownloadMbps)
		is.NoErr(err)
	}
}

// RealtimeAuthorized
// v1/auth/realtimeAuthorized
// IMUP_REALTIME_AUTHORIZED
func testRealTimeAuthorized(apiKey, email, status string) func(t *testing.T) {
	return func(t *testing.T) {
		is := is.New(t)

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer s.Close()
		testURL, _ := url.Parse(s.URL)

		// TODO: include a test from run_test that sets this env var
		// os.Setenv("IMUP_REALTIME_AUTHORIZED", testURL.String())

		ok, err := realtime.Authorized(context.Background(), apiKey, email, testURL.String())

		is.Equal(ok, true)
		is.NoErr(err)
	}
}

// TODO: right now this test is not configured with a proper api server response
// func testRemoteConfig(retcode int) func(t *testing.T) {
// 	return func(t *testing.T) {
// 		is := is.New(t)

// 		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			w.WriteHeader(http.StatusNoContent)
// 		}))
// 		defer s.Close()
// 		testURL, _ := url.Parse(s.URL)

// 		// TODO: include a test from run_test that sets this env var
// 		// os.Setenv("IMUP_REALTIME_CONFIG", testURL.String())
// 		cfg, err := realtime.NewConfig()
// 		is.NoErr(err)

// 		cfg.DisableRealtime()

// 		c, err := realtime.RemoteConfigReload(context.Background(), testURL.String())
// 		is.True(c.Realtime())
// 		is.NoErr(err)
// 	}
// }

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

		// TODO: include a test from run_test that sets this env var
		// os.Setenv("IMUP_REALTIME_CONFIG", testURL.String())

		config, err := realtime.NewConfig()
		is.NoErr(err)

		config.DisableRealtime()

		cfg, err := realtime.RemoteConfigReload(context.Background(), testURL.String())

		if name == "org" {
			is.NoErr(err)
			is.True(cfg.Realtime())
		} else {
			is.NoErr(err)
			is.Equal(cfg.Realtime(), false)
		}
	}
}
