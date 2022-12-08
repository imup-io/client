package main

import (
	"context"
	"net/url"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestSpeedTest(t *testing.T) {
	os.Clearenv()
	cases := []struct {
		Name     string
		ApiKey   string
		Email    string
		HostID   string
		Insecure string
		Quiet    string
		Realtime string
	}{
		{Name: "org", ApiKey: "1234", Email: "org-test@example.com", HostID: "org-based-host", Realtime: "true", Quiet: "true"},
		{Name: "user", ApiKey: "", Email: "test@example.com", HostID: "email-based-host", Realtime: "false", Insecure: "true"},
	}

	for _, c := range cases {
		os.Clearenv()
		os.Setenv("API_KEY", c.ApiKey)
		os.Setenv("EMAIL", c.Email)
		os.Setenv("HOST_ID", c.HostID)
		os.Setenv("INSECURE_SPEED_TEST", c.Insecure)
		os.Setenv("QUIET_SPEED_TEST", c.Quiet)
		os.Setenv("REALTIME", c.Realtime)

		t.Run(c.Name, testSendSpeedTestData())
		t.Run(c.Name, testRunSpeedTest())
	}
}

func testSendSpeedTestData() func(t *testing.T) {
	return func(t *testing.T) {
		s := defaultApiServer()
		defer s.Close()
		testURL, _ := url.Parse(s.URL)

		os.Setenv("IMUP_ADDRESS_SPEEDTEST", testURL.String())
		imup := newApp()

		data := &speedTestData{
			DownloadMbps:    1.0,
			UploadMbps:      1.0,
			DownloadMinRtt:  0.0,
			TimeStampStart:  time.Now().Unix(),
			TimeStampFinish: time.Now().Unix(),
			ClientVersion:   ClientVersion,
			OS:              runtime.GOOS,
		}

		da := speedtestD{
			Email:    imup.cfg.EmailAddress(),
			ID:       imup.cfg.HostID(),
			Key:      imup.cfg.APIKey(),
			IMUPData: data,
		}

		sendImupData(context.Background(), sendDataJob{imup.APIPostSpeedTestData, da})
	}
}

func testRunSpeedTest() func(t *testing.T) {
	return func(t *testing.T) {
		is := is.New(t)

		s := defaultApiServer()
		defer s.Close()
		testURL, _ := url.Parse(s.URL)

		os.Setenv("IMUP_SPEED_TEST_STATUS_ADDRESS", testURL.String())
		os.Setenv("IMUP_SPEED_TEST_RESULTS_ADDRESS", testURL.String())
		os.Setenv("IMUP_SHOULD_RUN_SPEEDTEST_ADDRESS", testURL.String())
		os.Setenv("IMUP_ADDRESS_SPEEDTEST", testURL.String())

		imup := newApp()

		wg := sync.WaitGroup{}
		cctx, cancel := context.WithCancel(context.Background())
		var err error
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(5 * time.Second)
			cancel()
		}()

		err = imup.runSpeedTest(cctx)
		wg.Wait()
		is.NoErr(err)
	}
}
