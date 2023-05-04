package main

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"

	"time"

	log "golang.org/x/exp/slog"
)

type speedtestD struct {
	Email     string `json:"email,omitempty"`
	ID        string `json:"hostId,omitempty"`
	Key       string `json:"apiKey,omitempty"`
	GroupID   string `json:"group_id,omitempty"`
	GroupName string `json:"group_name,omitempty"`

	IMUPData *speedTestData `json:"data,omitempty"`
}

type speedTestData struct {
	DownloadMbps    float64 `json:"downloadMbps,omitempty"`
	DownloadedBytes float64 `json:"downloadedBytes,omitempty"`
	DownloadRetrans float64 `json:"downloadRetrans,omitempty"`
	DownloadMinRtt  float64 `json:"downloadMinRTT,omitempty"`
	DownloadRTTVar  float64 `json:"downloadRTTVar,omitempty"`

	UploadMbps    float64 `json:"uploadMbps,omitempty"`
	UploadedBytes float64 `json:"uploadedBytes,omitempty"`
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

// speed test recursively attempts to get a speed test result utilizing the ndt7 back-off spec and a max retry
func (i *imup) speedTest(ctx context.Context, retries int) (*speedTestData, error) {
	s, err := RunSpeedTest(ctx, i.cfg.InsecureSpeedTests())
	if err != nil && retries < i.SpeedTestRetries {
		retries += 1
		// https://github.com/m-lab/ndt-server/blob/master/spec/ndt7-protocol.md#requirements-for-non-interactive-clients
		for mean := 60.0; mean <= 960.0; mean *= 2.0 {
			stdev := 0.05 * mean
			seconds := rand.NormFloat64()*stdev + mean
			time.Sleep(time.Duration(seconds * float64(time.Second)))

			return i.speedTest(ctx, retries)
		}
	}

	return s, err
}

func (i *imup) runSpeedTest(ctx context.Context) error {
	// lock speed test from running again while this is executing
	i.SpeedTestLock.Lock()
	defer i.SpeedTestLock.Unlock()

	i.SpeedTestRunning = true

	startTime := time.Now().UnixNano()
	tr, err := i.speedTest(ctx, 0)
	endTime := time.Now().UnixNano()

	if err != nil {
		if i.OnDemandSpeedTest {
			// ensure that if this request fails we will release the lock
			go i.postSpeedTestRealtimeStatus(ctx, "error")
		}

		log.Error("error running speed test", err)
		return fmt.Errorf("error running speed test: %v", err)
	}

	tr.TestServer = tr.Metadata["Server"]
	tr.TimeStampStart = startTime
	tr.TimeStampFinish = endTime
	tr.ClientVersion = ClientVersion
	tr.OS = runtime.GOOS

	d := sendDataJob{
		IMUPAddress: i.APIPostSpeedTestData,
		IMUPData: &speedtestD{
			Email:    i.cfg.EmailAddress(),
			ID:       i.cfg.HostID(),
			Key:      i.cfg.APIKey(),
			IMUPData: tr,
		},
	}

	// enqueue a job
	i.ChannelImupData <- d

	// update on-demand test
	if i.OnDemandSpeedTest {
		// ensure that if this request fails we will release the lock
		go i.postSpeedTestRealtimeResults(ctx, "complete", tr)
	}

	i.SpeedTestRunning = false

	return nil
}
