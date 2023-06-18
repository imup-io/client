package speedtesting_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/imup-io/client/speedtesting"
	"github.com/m-lab/ndt-server/ndt7/handler"
	"github.com/m-lab/ndt-server/ndt7/spec"
	"github.com/m-lab/ndt-server/netx"
	"github.com/matryer/is"
)

func TestSpeedTest(t *testing.T) {
	h, srv := NewNDT7Server(t)
	defer os.RemoveAll(h.DataDir)
	defer srv.Close()

	URL, _ := url.Parse(srv.URL)
	URL.Scheme = "ws"
	downloadURL, uploadURL := URL, URL
	uploadURL.Path = spec.UploadURLPath
	downloadURL.Path = spec.DownloadURLPath

	cases := []struct {
		name string
		ci   bool

		opts    speedtesting.Options
		timeout time.Duration
	}{
		{name: "mocked-upload", ci: true, timeout: time.Second * 1, opts: speedtesting.Options{
			Insecure:      true,
			OnDemand:      false,
			ClientVersion: "test-client",
			Server:        "",
			ServiceURL:    uploadURL,
		}},
		{name: "mocked-download", ci: true, timeout: time.Second * 1, opts: speedtesting.Options{
			Insecure:      true,
			OnDemand:      false,
			ClientVersion: "test-client",
			Server:        "",
			ServiceURL:    downloadURL,
		}},
		{name: "mocked-on-demand", ci: true, timeout: time.Second * 1, opts: speedtesting.Options{
			Insecure:      true,
			OnDemand:      true,
			ClientVersion: "test-client",
			Server:        "",
			ServiceURL:    URL,
		}},
		// integration-test runs a real speed test with actual ndt7 servers
		{name: "integration-test", ci: false, timeout: time.Minute * 5, opts: speedtesting.Options{
			Insecure:      false,
			OnDemand:      false,
			ClientVersion: "test-client",
			Server:        "",
			ServiceURL:    nil,
		}},
	}

	for _, c := range cases {
		os.Clearenv()

		// do not run integration test in ci
		if _, ok := os.LookupEnv("CI"); !ok {
			t.Run(c.name, testRunSpeedTest(c.opts, c.timeout))
		} else if c.ci {
			t.Run(c.name, testRunSpeedTest(c.opts, c.timeout))
		}
	}
}

func testRunSpeedTest(opts speedtesting.Options, timeout time.Duration) func(t *testing.T) {
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
			time.Sleep(timeout)
			cancel()
		}()

		result, err := speedtesting.Run(cctx, opts)

		is.NoErr(err)
		is.True(result != nil)
	}
}

// NewNDT7Server creates a local httptest server capable of running an ndt7
// measurement in unittests.
// https://raw.githubusercontent.com/m-lab/ndt-server/main/ndt7/ndt7test/ndt7test.go
func NewNDT7Server(t *testing.T) (*handler.Handler, *httptest.Server) {
	dir, err := ioutil.TempDir("", "ndt7test-*")
	if err != nil {
		t.Fatal(fmt.Errorf("failed to create temp dir: %v error: %s", dir, err))
	}

	ndt7Handler := &handler.Handler{DataDir: dir}

	ndt7Mux := http.NewServeMux()
	ndt7Mux.Handle(spec.DownloadURLPath, http.HandlerFunc(ndt7Handler.Download))
	ndt7Mux.Handle(spec.UploadURLPath, http.HandlerFunc(ndt7Handler.Upload))

	// Create un-started so we can setup a custom netx.Listener.
	ts := httptest.NewUnstartedServer(ndt7Mux)
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}

	addr := (listener.(*net.TCPListener)).Addr().(*net.TCPAddr)

	// Populate insecure port value with dynamic port.
	ndt7Handler.InsecurePort = fmt.Sprintf(":%d", addr.Port)
	ts.Listener = netx.NewListener(listener.(*net.TCPListener))

	// Now that the test server has our custom listener, start it.
	ts.Start()
	return ndt7Handler, ts
}
