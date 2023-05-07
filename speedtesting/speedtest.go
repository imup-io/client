package speedtesting

import (
	"context"
	"fmt"
	"math/rand"
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

func Run(ctx context.Context, insecure, onDemand bool, version string) (*SpeedTestResult, error) {
	// lock speed test from running again while this is executing
	mu.Lock()
	defer mu.Unlock()

	startTime := time.Now().UnixNano()
	result, err := speedTest(ctx, 0, insecure)
	endTime := time.Now().UnixNano()

	if err != nil {
		log.Error("error running speed test", err)
		return nil, fmt.Errorf("error running speed test: %v", err)
	}

	result.TestServer = result.Metadata["Server"]
	result.TimeStampStart = startTime
	result.TimeStampFinish = endTime
	result.ClientVersion = version
	result.OS = runtime.GOOS

	return result, nil
}

// speed test recursively attempts to get a speed test result utilizing the ndt7 back-off spec and a max retry
func speedTest(ctx context.Context, retries int, insecure bool) (*SpeedTestResult, error) {
	s, err := RunSpeedTest(ctx, insecure)
	if err != nil && retries < 10 {
		retries += 1
		// https://github.com/m-lab/ndt-server/blob/master/spec/ndt7-protocol.md#requirements-for-non-interactive-clients
		for mean := 60.0; mean <= 960.0; mean *= 2.0 {
			stdev := 0.05 * mean
			seconds := rand.NormFloat64()*stdev + mean
			time.Sleep(time.Duration(seconds * float64(time.Second)))

			return speedTest(ctx, retries, insecure)
		}
	}

	return s, err
}
