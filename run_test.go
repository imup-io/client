package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/matryer/is"
	log "github.com/sirupsen/logrus"
)

func Test_Run(t *testing.T) {
	is := is.New(t)
	// unset test environment
	defer os.Clearenv()
	// clear pending workers
	defer clearCache()

	wg := sync.WaitGroup{}
	wg.Add(1)

	ts := defaultApiServer()

	// imup setup
	os.Setenv("IMUP_ADDRESS", ts.URL)
	os.Setenv("IMUP_ADDRESS_SPEEDTEST", ts.URL)
	os.Setenv("IMUP_LIVENESS_CHECKIN_ADDRESS", ts.URL)
	os.Setenv("IMUP_SHOULD_RUN_SPEEDTEST_ADDRESS", ts.URL)
	os.Setenv("IMUP_SPEED_TEST_RESULTS_ADDRESS", ts.URL)
	os.Setenv("IMUP_SPEED_TEST_STATUS_ADDRESS", ts.URL)
	os.Setenv("IMUP_REALTIME_AUTHORIZED", ts.URL)
	os.Setenv("IMUP_REALTIME_CONFIG", ts.URL)
	os.Setenv("PING_INTERVAL", "1")
	os.Setenv("PING_DELAY", "100")
	os.Setenv("PING_REQUESTS", "2")

	os.Setenv("PING_ADDRESS", "127.0.0.1")
	os.Setenv("PING_ADDRESS_INTERNAL", "127.0.0.1")

	os.Setenv("IMUP_REALTIME_CONFIG", ts.URL)

	// application configuration
	os.Setenv("EMAIL", "test@example.com")

	shutdown := make(chan os.Signal, 1)
	// cctx, cancel := context.WithCancel(context.Background())
	start := time.Now().Unix()
	var err error
	go func() {
		defer wg.Done()
		err = run(context.Background(), shutdown)
		log.Info(err)
	}()

	time.Sleep(2 * time.Second)

	// cancel()
	shutdown <- syscall.SIGINT
	wg.Wait()
	elapsed := time.Now().Unix() - start
	is.True(elapsed < 4)
	is.NoErr(err)
}

func defaultApiServer() *httptest.Server {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.WriteHeader(http.StatusOK)
	}))

	return s
}

func defaultConfigurableApiServer(retcode int) *httptest.Server {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.WriteHeader(http.StatusNoContent)
	}))

	return s
}
