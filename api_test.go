package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/matryer/is"
)

// 	imup.APIPostConnectionData = getEnv("IMUP_ADDRESS", "https://api.imup.io/v1/data/connectivity")

// imup.SpeedTestStatusUpdateAddress = getEnv("IMUP_SPEED_TEST_STATUS_ADDRESS", "https://api.imup.io/v1/realtime/speedTestStatusUpdate")
func TestApi_PostSpeedTestStatus(t *testing.T) {
	cases := []struct {
		ApiKey   string
		Email    string
		EndPoint string
		HostID   string
		Payload  interface{}
		RetCode  int
		Type     string
	}{
		{ApiKey: "1234", HostID: "homer", Email: "org-test@example.com", EndPoint: "realtime/speedTestStatusUpdate", Type: "org", RetCode: http.StatusOK, Payload: "running"},
		{ApiKey: "1234", HostID: "homer", Email: "org-test@example.com", EndPoint: "realtime/speedTestStatusUpdate", Type: "org", RetCode: http.StatusInternalServerError, Payload: "error"},
		{Email: "test@example.com", EndPoint: "realtime/speedTestStatusUpdate", Type: "user", RetCode: http.StatusOK, Payload: "running"},
		{Email: "test@example.com", EndPoint: "realtime/speedTestStatusUpdate", Type: "user", RetCode: http.StatusInternalServerError, Payload: "error"},
	}

	for _, c := range cases {
		data := &realtimeApiPayload{
			ID: c.HostID, Key: c.ApiKey, Email: c.Email, Data: c.Payload,
		}
		s := apiTestServer(c.EndPoint, data, c.RetCode, t)
		defer s.Close()
		testURL, _ := url.Parse(s.URL)
		os.Setenv("IMUP_SPEED_TEST_STATUS_ADDRESS", testURL.String())
		os.Setenv("API_KEY", c.ApiKey)
		os.Setenv("EMAIL", c.Email)
		os.Setenv("HOST_ID", c.HostID)

		is := is.New(t)

		imup := newApp()

		wg := sync.WaitGroup{}
		wg.Add(1)
		var err error

		go func() {
			defer wg.Done()
			err = imup.postSpeedTestRealtimeStatus(context.Background(), data.Data.(string))
		}()

		wg.Wait()

		if c.RetCode >= 500 {
			is.True(err != nil)
		} else {
			is.NoErr(err)
		}
	}
}

// imup.SpeedTestResultsAddress = getEnv("IMUP_SPEED_TEST_RESULTS_ADDRESS", "https://api.imup.io/v1/realtime/speedTestResults")
func TestApi_PostSpeedTestResults(t *testing.T) {
	payload := struct {
		u    float64
		d    float64
		data string
	}{1.001, 1.001, "complete"}

	cases := []struct {
		ApiKey   string
		Email    string
		EndPoint string
		HostID   string
		Payload  interface{}
		RetCode  int
		Type     string
	}{
		{ApiKey: "1234", HostID: "homer", Email: "org-test@example.com", EndPoint: "realtime/speedTestResults", Type: "org", RetCode: http.StatusOK, Payload: payload},
		{Email: "test@example.com", EndPoint: "realtime/speedTestResults", Type: "user", RetCode: http.StatusOK, Payload: payload},
	}

	for _, c := range cases {
		data := &realtimeApiPayload{
			ID: c.HostID, Key: c.ApiKey, Email: c.Email, Data: c.Payload,
		}

		s := apiTestServer(c.EndPoint, data, c.RetCode, t)
		defer s.Close()
		testURL, _ := url.Parse(s.URL)
		os.Setenv("IMUP_SPEED_TEST_RESULTS_ADDRESS", testURL.String())
		os.Setenv("API_KEY", c.ApiKey)
		os.Setenv("EMAIL", c.Email)
		os.Setenv("HOST_ID", c.HostID)

		is := is.New(t)

		imup := newApp()

		wg := sync.WaitGroup{}
		wg.Add(1)

		var err error
		go func() {
			defer wg.Done()

			tr := &speedTestData{UploadMbps: payload.u, DownloadMbps: payload.d}
			err = imup.postSpeedTestRealtimeResults(context.Background(), payload.data, tr)
		}()

		wg.Wait()
		is.NoErr(err)
	}
}

// imup.ShouldRunSpeedTestAddress = getEnv("IMUP_SHOULD_RUN_SPEEDTEST_ADDRESS", "https://api.imup.io/v1/realtime/shouldClientRunSpeedTest")
func TestApi_PostShouldClientRunSpeedTest(t *testing.T) {
	cases := []struct {
		ApiKey   string
		Email    string
		EndPoint string
		HostID   string
		Payload  interface{}
		RetCode  int
		Type     string
	}{
		// TODO: need to pass in an extra param to test failure cases here
		{ApiKey: "1234", HostID: "homer", Email: "org-test@example.com", EndPoint: "realtime/shouldClientRunSpeedTest", Type: "org", RetCode: http.StatusOK},
		{Email: "test@example.com", EndPoint: "realtime/shouldClientRunSpeedTest", Type: "user", RetCode: http.StatusOK},
	}

	for _, c := range cases {
		senddata := &realtimeApiPayload{ID: c.HostID, Key: c.ApiKey, Email: c.Email}
		s := apiTestServer(c.EndPoint, senddata, c.RetCode, t)
		defer s.Close()
		testURL, _ := url.Parse(s.URL)
		os.Setenv("IMUP_SHOULD_RUN_SPEEDTEST_ADDRESS", testURL.String())
		os.Setenv("API_KEY", c.ApiKey)
		os.Setenv("EMAIL", c.Email)
		os.Setenv("HOST_ID", c.HostID)

		is := is.New(t)

		imup := newApp()

		wg := sync.WaitGroup{}
		wg.Add(1)
		var ok bool
		var err error
		go func() {
			defer wg.Done()
			ok, err = imup.shouldRunSpeedtest(context.Background())
		}()

		wg.Wait()

		is.NoErr(err)
		is.True(ok)
	}
}

// imup.LivenessCheckInAddress = getEnv("IMUP_LIVENESS_CHECKIN_ADDRESS", "https://api.imup.io/v1/realtime/livenesscheckin")
func TestApi_PostLivenesscheckin(t *testing.T) {
	cases := []struct {
		ApiKey   string
		Email    string
		EndPoint string
		HostID   string
		Payload  interface{}
		RetCode  int
		Type     string
	}{
		{ApiKey: "1234", HostID: "homer", Email: "org-test@example.com", EndPoint: "realtime/livenesscheckin", Type: "org", RetCode: http.StatusOK, Payload: ""},
		{ApiKey: "1234", HostID: "homer", Email: "org-test@example.com", EndPoint: "realtime/livenesscheckin", Type: "org", RetCode: http.StatusInternalServerError, Payload: ""},
		{ApiKey: "1234", HostID: "homer", EndPoint: "realtime/livenesscheckin", Type: "org", RetCode: http.StatusOK, Payload: "running"},
		{Email: "test@example.com", EndPoint: "realtime/livenesscheckin", Type: "user", RetCode: http.StatusOK, Payload: ""},
		{Email: "test@example.com", EndPoint: "realtime/livenesscheckin", Type: "user", RetCode: http.StatusInternalServerError, Payload: ""},
	}

	for _, c := range cases {
		c.Payload = &realtimeApiPayload{ID: c.HostID, Key: c.ApiKey, Email: c.Email}
		s := apiTestServer(c.EndPoint, c.Payload, c.RetCode, t)
		defer s.Close()
		testURL, _ := url.Parse(s.URL)
		os.Setenv("IMUP_LIVENESS_CHECKIN_ADDRESS", testURL.String())
		os.Setenv("API_KEY", c.ApiKey)
		os.Setenv("EMAIL", c.Email)
		os.Setenv("HOST_ID", c.HostID)

		is := is.New(t)

		imup := newApp()

		err := imup.sendClientHealthy(context.Background())
		if c.RetCode > 499 {
			is.True(err != nil)
		} else {
			is.NoErr(err)
		}

	}
}

// imup.APIPostSpeedTestData = getEnv("IMUP_ADDRESS_SPEEDTEST", "https://api.imup.io/v1/data/speedtest")
func TestApi_PostSpeedTestData(t *testing.T) {
	cases := []struct {
		ApiKey   string
		Email    string
		EndPoint string
		HostID   string
		Payload  interface{}
		RetCode  int
		Type     string
	}{
		// Testing failure cases are tricky since we retry this endpoint forever.
		// TODO: Make retries configurable for tests.
		{ApiKey: "1234", HostID: "homer", Email: "org-test@example.com", EndPoint: "data/speedtest", Type: "org", RetCode: http.StatusOK},
		{Email: "test@example.com", EndPoint: "data/speedtest", Type: "user", RetCode: http.StatusOK},
	}

	for _, c := range cases {
		c.Payload = &speedtestD{
			Email: c.Email,
			ID:    c.HostID,
			Key:   c.ApiKey,
			IMUPData: &speedTestData{
				DownloadMbps:    1.001,
				UploadMbps:      0.1001,
				DownloadMinRtt:  0.9,
				TestServer:      "somewhere",
				UploadedBytes:   10000000.0,
				DownloadedBytes: 9999999.0,
				TimeStampStart:  time.Now().Unix(),
				TimeStampFinish: time.Now().Unix(),
				ClientVersion:   ClientVersion,
				OS:              runtime.GOOS,
			},
		}

		s := apiTestServer(c.EndPoint, c.Payload, c.RetCode, t)
		defer s.Close()
		testURL, _ := url.Parse(s.URL)
		os.Setenv("IMUP_ADDRESS_SPEEDTEST", testURL.String())
		os.Setenv("API_KEY", c.ApiKey)
		os.Setenv("EMAIL", c.Email)
		os.Setenv("HOST_ID", c.HostID)

		imup := newApp()

		wg := sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer wg.Done()
			sendImupData(context.Background(), sendDataJob{imup.APIPostSpeedTestData, c.Payload})
		}()
		wg.Wait()
	}
}

// imup.APIPostConnectionData = getEnv("IMUP_ADDRESS", "https://api.imup.io/v1/data/connectivity")
func TestApi_PostPingConnectionData(t *testing.T) {
	cases := []struct {
		ApiKey   string
		Email    string
		EndPoint string
		HostID   string
		Payload  interface{}
		RetCode  int
		Type     string
	}{
		{ApiKey: "1234", HostID: "homer", Email: "org-test@example.com", EndPoint: "data/connectivity/ping", Type: "org", RetCode: http.StatusOK},
		{Email: "test@example.com", EndPoint: "data/connectivity/ping", Type: "user", RetCode: http.StatusOK},
	}

	for _, c := range cases {
		c.Payload = imupData{
			Downtime:      0,
			StatusChanged: false,
			Email:         c.Email,
			ID:            c.HostID,
			Key:           c.ApiKey,
			IMUPData:      []pingStats{},
		}

		s := apiTestServer(c.EndPoint, c.Payload, c.RetCode, t)
		defer s.Close()
		testURL, _ := url.Parse(s.URL)
		os.Setenv("IMUP_ADDRESS", testURL.String())
		os.Setenv("API_KEY", c.ApiKey)
		os.Setenv("EMAIL", c.Email)
		os.Setenv("HOST_ID", c.HostID)

		imup := newApp()

		wg := sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer wg.Done()
			sendImupData(context.Background(), sendDataJob{imup.APIPostConnectionData, c.Payload})
		}()
		wg.Wait()
	}
}

// imup.APIPostConnectionData = getEnv("IMUP_ADDRESS", "https://api.imup.io/v1/data/connectivity")
func TestApi_PostConnectionData(t *testing.T) {
	cases := []struct {
		ApiKey   string
		Email    string
		EndPoint string
		HostID   string
		Payload  interface{}
		RetCode  int
		Type     string
	}{
		{ApiKey: "1234", HostID: "homer", Email: "org-test@example.com", EndPoint: "data/connectivity", Type: "org", RetCode: http.StatusOK},
		{Email: "test@example.com", EndPoint: "data/connectivity", Type: "user", RetCode: http.StatusOK},
	}

	for _, c := range cases {
		c.Payload = imupData{
			Downtime:      0,
			StatusChanged: false,
			Email:         c.Email,
			ID:            c.HostID,
			Key:           c.ApiKey,
			IMUPData:      []pingStats{},
		}

		s := apiTestServer(c.EndPoint, c.Payload, c.RetCode, t)
		defer s.Close()
		testURL, _ := url.Parse(s.URL)
		os.Setenv("IMUP_ADDRESS", testURL.String())
		os.Setenv("API_KEY", c.ApiKey)
		os.Setenv("EMAIL", c.Email)
		os.Setenv("HOST_ID", c.HostID)

		imup := newApp()

		wg := sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer wg.Done()
			sendImupData(context.Background(), sendDataJob{imup.APIPostConnectionData, c.Payload})
		}()
		wg.Wait()
	}
}

// imup.RealtimeConfig = getEnv("IMUP_REALTIME_CONFIG", "https://api.imup.io/v1/realtime/config")
func TestApi_PostConfigReload(t *testing.T) {
	cases := []struct {
		ApiKey   string
		Email    string
		EndPoint string
		HostID   string
		GroupID  string
		Version  string
		RetCode  int
		Type     string
	}{
		{ApiKey: "1234", HostID: "homer", GroupID: "", Version: "dev preview", Email: "org-test@example.com", EndPoint: "realtime/remoteConfigReload", Type: "org", RetCode: http.StatusNoContent},
		{ApiKey: "1234", HostID: "homer", GroupID: "uuid", Version: "2023.04.02v1", Email: "org-test1@example.com", EndPoint: "realtime/remoteConfigReload", Type: "org", RetCode: http.StatusOK},
	}

	for _, c := range cases {
		sendData := &realtimeApiPayload{ID: c.HostID, Key: c.ApiKey, Email: c.Email, GroupID: c.GroupID, Version: c.Version}
		s := apiTestServer(c.EndPoint, sendData, c.RetCode, t)
		defer s.Close()
		testURL, _ := url.Parse(s.URL)
		os.Setenv("IMUP_REALTIME_CONFIG", testURL.String())
		os.Setenv("API_KEY", c.ApiKey)
		os.Setenv("EMAIL", c.Email)
		os.Setenv("HOST_ID", c.HostID)
		os.Setenv("GROUP_ID", c.GroupID)

		is := is.New(t)

		imup := newApp()

		wg := sync.WaitGroup{}
		wg.Add(1)
		var err error
		go func() {
			defer wg.Done()
			err = imup.remoteConfigReload(context.Background())
		}()

		wg.Wait()
		is.NoErr(err)

		if c.RetCode == http.StatusOK {
			is.Equal(imup.cfg.Version(), "2023.04.02v2")
		} else {
			is.Equal(imup.cfg.Version(), "dev-preview")
		}
	}
}

// loose reflection of imup api endpoints and their expected payloads
func apiTestServer(endpoint string, payload interface{}, retcode int, t *testing.T) *httptest.Server {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(retcode)
		if retcode > 499 {
			return
		}

		switch endpoint {
		case "data/connectivity":
			recvdata := &imupData{}
			if err := json.NewDecoder(r.Body).Decode(recvdata); err != nil {
				t.Error(err)
			}
			expected, ok := payload.(imupData)
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				t.Error("payload is not the expected type")
			}
			if recvdata.Key != expected.Key {
				t.Errorf("Expected: %v, Got: %v", expected.Key, recvdata.Key)
			}
		case "data/connectivity/ping":
			recvdata := &imupData{}
			if err := json.NewDecoder(r.Body).Decode(recvdata); err != nil {
				t.Error(err)
			}
			expected, ok := payload.(imupData)
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				t.Error("payload is not the expected type")
			}
			if expected.Key != recvdata.Key {
				t.Errorf("Expected: %v, Got: %v", expected.Key, recvdata.Key)
			}

		case "data/speedtest":
			recvdata := &speedtestD{}
			if err := json.NewDecoder(r.Body).Decode(recvdata); err != nil {
				t.Error(err)
			}

			expected := payload.(*speedtestD)
			if recvdata.IMUPData.DownloadMbps != expected.IMUPData.DownloadMbps {
				t.Errorf("Expected: %v, Got: %v", expected.IMUPData.DownloadMbps, recvdata.IMUPData.DownloadMbps)
			}
			if recvdata.IMUPData.DownloadedBytes != expected.IMUPData.DownloadedBytes {
				t.Errorf("Expected: %v, Got: %v", expected.IMUPData.DownloadedBytes, recvdata.IMUPData.DownloadedBytes)
			}
			if recvdata.IMUPData.UploadMbps != expected.IMUPData.UploadMbps {
				t.Errorf("Expected: %v, Got: %v", expected.IMUPData.UploadMbps, recvdata.IMUPData.UploadMbps)
			}
			if recvdata.IMUPData.UploadedBytes != expected.IMUPData.UploadedBytes {
				t.Errorf("Expected: %v, Got: %v", expected.IMUPData.UploadedBytes, recvdata.IMUPData.UploadedBytes)
			}

		case "realtime/livenesscheckin":
			recvdata := &realtimeApiPayload{}

			if err := json.NewDecoder(r.Body).Decode(recvdata); err != nil {
				t.Error(err)
			}

			m := payload.(*realtimeApiPayload)

			if m.Email != recvdata.Email {
				t.Errorf("Expected: %v Got: %v", m.Email, recvdata.Email)
			}
			if m.ID != recvdata.ID {
				t.Errorf("Expected: %v Got: %v", m.ID, recvdata.ID)
			}
			if m.Key != recvdata.Key {
				t.Errorf("Expected: %v Got: %v", m.Key, recvdata.Key)
			}

			// is.Equal(data, payload)

		case "realtime/shouldClientRunSpeedTest":
			recvdata := &realtimeApiPayload{}

			if err := json.NewDecoder(r.Body).Decode(recvdata); err != nil {
				t.Error(err)
			}

			expected, ok := payload.(*realtimeApiPayload)
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				t.Error("payload is not the expected type")
			}
			if expected.ID != recvdata.ID {
				t.Errorf("Expected: %v Got: %v", expected.ID, recvdata.ID)
			}
			if expected.Email != recvdata.Email {
				t.Errorf("Expected: %v Got: %v", expected.Email, recvdata.Email)
			}
			if expected.Key != recvdata.Key {
				t.Errorf("Expected: %v Got: %v", expected.Key, recvdata.Key)
			}

			// Api Server Response
			w.Header().Set("Content-Type", "application/json")
			data := struct {
				Success bool `json:"success"`
				Data    bool `json:"data"`
			}{true, true}

			if err := json.NewEncoder(w).Encode(data); err != nil {
				t.Error(err)
			}

		case "realtime/speedTestResults":
			recvdata := &realtimeApiPayload{}

			if err := json.NewDecoder(r.Body).Decode(recvdata); err != nil {
				t.Error(err)
			}

			expected, ok := payload.(*realtimeApiPayload)
			if !ok {
				w.WriteHeader(http.StatusBadRequest)
				t.Error("payload is not the expected type")
			}
			if expected.ID != recvdata.ID {
				t.Errorf("Expected: %v Got: %v", expected.ID, recvdata.ID)
			}
			if expected.Email != recvdata.Email {
				t.Errorf("Expected: %v Got: %v", expected.Email, recvdata.Email)
			}
			if expected.Key != recvdata.Key {
				t.Errorf("Expected: %v Got: %v", expected.Key, recvdata.Key)
			}

		case "realtime/speedTestStatusUpdate":
			recvdata := &realtimeApiPayload{}

			if err := json.NewDecoder(r.Body).Decode(recvdata); err != nil {
				t.Error(err)
			}

			apiPayload, ok := payload.(*realtimeApiPayload)
			if !ok {
				t.Error("payload was not the expected type")
				fmt.Fprintf(w, "payload was not the expected type")
			}
			got := recvdata.Data.(string)
			expected := apiPayload.Data.(string)
			if got != expected {
				t.Errorf("Expected: %v Got: %v", expected, got)
			}

		case "realtime/remoteConfigReload":
			if retcode == http.StatusNoContent {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			recvdata := &realtimeApiPayload{}
			if err := json.NewDecoder(r.Body).Decode(recvdata); err != nil {
				t.Error(err)
			}

			// Api Server Response
			w.Header().Set("Content-Type", "application/json")
			data := `{"config": {"version": "2023.04.02v2","environment": "","groupID": "new-group-id","groupName": "","insecureSpeedTest": false,"noDiscoverGateway": false,"nonvolatile": false,"pingEnabled": false,"realtimeEnabled": false,"speedTestEnabled": false,"allowlisted_ips": null,"blocklisted_ips": null}}`

			if _, err := fmt.Fprint(w, data); err != nil {
				t.Error(err)
			}
		default:
			t.Errorf("unknown endpoint tested against %s", endpoint)
			t.Fail()
		}
	}))

	return s
}
