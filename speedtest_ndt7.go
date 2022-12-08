package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	ndt7 "github.com/m-lab/ndt7-client-go"
	"github.com/m-lab/ndt7-client-go/spec"
)

// NOTE: ClientName is set via build flags
// https://github.com/m-lab/ndt7-client-go#building-with-a-custom-client-name
var ClientName = "client"

const (
	clientVersion  = "0.7.0"
	defaultTimeout = 55 * time.Second
)

type startFunc func(context.Context) (<-chan spec.Measurement, error)

var lock sync.Mutex

func RunSpeedTest(ctx context.Context, insecure, quiet bool) (*speedTestData, error) {
	lock.Lock()
	defer lock.Unlock()

	client := ndt7.NewClient(ClientName, clientVersion)

	if insecure {
		client.Scheme = "ws"
	}

	tests := map[spec.TestKind]startFunc{
		spec.TestDownload: client.StartDownload,
		spec.TestUpload:   client.StartUpload,
	}

	e := NewEmitterOutput(os.Stdout)
	for spec, f := range tests {
		if quiet {
			testRunner(ctx, client.FQDN, spec, f)
		} else {
			e.testRunner(ctx, client.FQDN, spec, f)
		}
	}

	result := summary(client)
	if !quiet {
		e.Summary(result)
	}

	return result, nil
}

func testRunner(ctx context.Context, fqdn string, kind spec.TestKind, start startFunc) error {
	ch, err := start(ctx)
	if err != nil {
		return err
	}

	errChan := make(chan error)
	for event := range ch {
		func(m *spec.Measurement) {
			// switch on tcp info or app info depending on test type
			switch m.Test {
			case spec.TestDownload:
				if m.Origin == spec.OriginClient {
					if m.AppInfo == nil || m.AppInfo.ElapsedTime <= 0 {
						errChan <- fmt.Errorf("missing m.AppInfo or invalid m.AppInfo.ElapsedTime")
					}
				}
			case spec.TestUpload:
				if m.Origin == spec.OriginServer {
					if m.TCPInfo == nil || m.TCPInfo.ElapsedTime <= 0 {
						errChan <- fmt.Errorf("missing m.TCPInfo or invalid m.TCPInfo.ElapsedTime")
					}
				}
			}
		}(&event)
	}

	close(errChan)

	var errs error
	for err = range errChan {
		if err != nil {
			e := fmt.Errorf("%v", errs)
			errs = fmt.Errorf("%v", e)
		}
	}

	return errs
}

func summary(client *ndt7.Client) *speedTestData {
	data := &speedTestData{}

	data.Metadata = map[string]string{}
	data.Metadata["Server"] = client.FQDN

	if dl, ok := client.Results()[spec.TestDownload]; ok {
		if dl.Client.AppInfo != nil && dl.Client.AppInfo.ElapsedTime > 0 {
			data.Metadata["Client IP"] = dl.ConnectionInfo.Client

			elapsed := float64(dl.Client.AppInfo.ElapsedTime) / 1e06
			data.DownloadedBytes = float64(dl.Client.AppInfo.NumBytes)
			data.DownloadMbps = (8.0 * data.DownloadedBytes) / elapsed / (1000.0 * 1000.0)
		}

		if dl.Server.TCPInfo != nil {
			if dl.Server.TCPInfo.BytesSent > 0 {
				data.DownloadMinRtt = float64(dl.Server.TCPInfo.MinRTT) / 1000.0
				data.DownloadRetrans = float64(dl.Server.TCPInfo.BytesRetrans) / float64(dl.Server.TCPInfo.BytesSent) * 100.0
				data.DownloadRTTVar = float64(dl.Server.TCPInfo.RTTVar) / 1000.0
			}
		}
	}

	if ul, ok := client.Results()[spec.TestUpload]; ok {
		if ul.Server.TCPInfo != nil && ul.Server.TCPInfo.BytesReceived > 0 {
			data.Metadata["Server IP"] = ul.ConnectionInfo.Server
			data.Metadata["Server UUID"] = ul.ConnectionInfo.UUID

			elapsed := float64(ul.Server.TCPInfo.ElapsedTime) / 1e06
			data.UploadedBytes = float64(ul.Server.TCPInfo.BytesReceived)
			data.UploadMbps = (8.0 * data.UploadedBytes) / elapsed / (1000.0 * 1000.0)
			data.UploadMinRTT = float64(ul.Server.TCPInfo.MinRTT) / 1000.0
			data.UploadRetrans = float64(ul.Server.TCPInfo.BytesRetrans) / float64(ul.Server.TCPInfo.BytesSent) * 100.0
			data.UploadRTTVar = float64(ul.Server.TCPInfo.RTTVar) / 1000.0
		}
	}

	return data
}
