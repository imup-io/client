package speedtesting

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"runtime"
	"sync"

	"time"

	log "golang.org/x/exp/slog"
)

var mu sync.Mutex

type SpeedTestResult struct {
	DownloadedBytes float64 `json:"downloadedBytes,omitempty"`
	DownloadMbps    float64 `json:"downloadMbps,omitempty"`
	DownloadRetrans float64 `json:"downloadRetrans,omitempty"`
	DownloadMinRtt  float64 `json:"downloadMinRTT,omitempty"`
	DownloadRTTVar  float64 `json:"downloadRTTVar,omitempty"`

	UploadedBytes float64 `json:"uploadedBytes,omitempty"`
	UploadMbps    float64 `json:"uploadMbps,omitempty"`
	UploadRetrans float64 `json:"uploadRetrans,omitempty"`
	UploadMinRTT  float64 `json:"uploadMinRTT,omitempty"`
	UploadRTTVar  float64 `json:"uploadRTTVar,omitempty"`

	TimeStampStart  int64 `json:"timestampStart,omitempty"`
	TimeStampFinish int64 `json:"timestampFinish,omitempty"`

	Metadata      map[string]string `json:"metadata,omitempty"`
	ClientVersion string            `json:"clientVersion,omitempty"`
	OS            string            `json:"operatingSystem,omitempty"`
	TestServer    string            `json:"testServer,omitempty"`
}

type Options struct {
	Insecure bool
	OnDemand bool

	ClientVersion string

	// if set takes precedence over ndt7 locate API as well as the serviceURL
	Server string

	// if set takes precedence over ndt7 locate API
	ServiceURL *url.URL
}

func Run(ctx context.Context, opts Options) (*SpeedTestResult, error) {
	// lock speed test from running again while this is executing
	mu.Lock()
	defer mu.Unlock()

	startTime := time.Now().UnixNano()
	// TODO: fix error handling for windows speed testing
	result, err := RunSpeedTest(ctx, &opts)
	endTime := time.Now().UnixNano()

	if err != nil && !errors.Is(err, context.Canceled) {
		log.Error("error running speed test", err)
		return nil, fmt.Errorf("error running speed test: %v", err)
	}

	result.TestServer = result.Metadata["Server"]
	result.TimeStampStart = startTime
	result.TimeStampFinish = endTime
	result.ClientVersion = opts.ClientVersion
	result.OS = runtime.GOOS

	return result, nil
}
