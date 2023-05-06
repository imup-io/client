package speedtesting_test

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/imup-io/client/speedtesting"
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
		Realtime string
	}{
		{Name: "org", ApiKey: "1234", Email: "org-test@example.com", HostID: "org-based-host", Realtime: "true"},
		{Name: "user", ApiKey: "", Email: "test@example.com", HostID: "email-based-host", Realtime: "false", Insecure: "true"},
	}

	for _, c := range cases {
		os.Clearenv()
		os.Setenv("API_KEY", c.ApiKey)
		os.Setenv("EMAIL", c.Email)
		os.Setenv("HOST_ID", c.HostID)
		os.Setenv("INSECURE_SPEED_TEST", c.Insecure)
		os.Setenv("REALTIME", c.Realtime)

		t.Run(c.Name, testRunSpeedTest())
	}
}

func testRunSpeedTest() func(t *testing.T) {
	return func(t *testing.T) {
		is := is.New(t)

		// TODO: include a test from run_test that sets this env var
		// os.Setenv("IMUP_SPEED_TEST_STATUS_ADDRESS", testURL.String())
		// os.Setenv("IMUP_SPEED_TEST_RESULTS_ADDRESS", testURL.String())
		// os.Setenv("IMUP_SHOULD_RUN_SPEEDTEST_ADDRESS", testURL.String())
		// os.Setenv("IMUP_ADDRESS_SPEEDTEST", testURL.String())

		wg := sync.WaitGroup{}
		cctx, cancel := context.WithCancel(context.Background())
		var err error
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(5 * time.Second)
			cancel()
		}()

		result, err := speedtesting.Run(cctx, false, false, "client-version")
		wg.Wait()

		is.NoErr(err)
		is.True(result != nil)
	}
}
