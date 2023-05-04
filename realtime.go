package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	log "golang.org/x/exp/slog"
)

type realtimeApiPayload struct {
	ID        string      `json:"hostId,omitempty"`
	Email     string      `json:"email,omitempty"`
	GroupID   string      `json:"groupID,omitempty"`
	GroupName string      `json:"groupName,omitempty"`
	Key       string      `json:"apiKey,omitempty"`
	Version   string      `json:"version,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

func (i *imup) sendClientHealthy(ctx context.Context) error {
	data := &realtimeApiPayload{
		ID: i.cfg.HostID(), Key: i.cfg.APIKey(), Email: i.cfg.EmailAddress(),
	}

	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("json.Marshal: %v", err)
	}

	return sendRealtimeData(ctx, bytes.NewBuffer(b), i.LivenessCheckInAddress)
}

func sendRealtimeData(ctx context.Context, b *bytes.Buffer, addr string) error {
	req, err := retryablehttp.NewRequest("POST", addr, b)
	req = req.WithContext(ctx)
	if err != nil {
		return fmt.Errorf("NewRequest: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := retryablehttp.NewClient()
	client.Backoff = exactJitterBackoff
	client.RetryMax = 3
	client.RetryWaitMin = time.Duration(200) * time.Millisecond
	client.RetryWaitMax = time.Duration(3) * time.Second
	client.Logger = log.New(log.Default().Handler())

	if _, err := client.Do(req); err != nil {
		return fmt.Errorf("addr: %s, client.Do: %s", addr, err)
	}

	return nil
}

func (i *imup) shouldRunSpeedtest(ctx context.Context) (bool, error) {
	data := &realtimeApiPayload{
		ID: i.cfg.HostID(), Key: i.cfg.APIKey(), Email: i.cfg.EmailAddress(),
	}

	b, err := json.Marshal(data)
	if err != nil {
		return false, fmt.Errorf("json.Marshal: %v", err)
	}

	req, err := retryablehttp.NewRequest("POST", i.ShouldRunSpeedTestAddress, bytes.NewBuffer(b))
	req = req.WithContext(ctx)
	if err != nil {
		return false, fmt.Errorf("NewRequest: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := retryablehttp.NewClient()
	client.Backoff = exactJitterBackoff

	client.RetryMax = 2
	client.RetryWaitMin = time.Duration(30) * time.Second
	client.RetryWaitMax = time.Duration(60) * time.Second
	client.Logger = log.New(log.Default().Handler())

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("addr: %s, client.Do: %v", i.ShouldRunSpeedTestAddress, err)
	}
	defer resp.Body.Close()

	sr := struct {
		Success bool `json:"success,omitempty"`
		Data    bool `json:"data,omitempty"`
	}{}

	if err = json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return false, fmt.Errorf("error parsing server response: %v", err)
	}

	return sr.Data, nil
}

func (i *imup) postSpeedTestRealtimeStatus(ctx context.Context, status string) error {
	data := &realtimeApiPayload{
		ID: i.cfg.HostID(), Key: i.cfg.APIKey(), Email: i.cfg.EmailAddress(), Data: status,
	}

	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("json.Marshal: %v", err)
	}

	return sendRealtimeData(ctx, bytes.NewBuffer(b), i.SpeedTestStatusUpdateAddress)
}

func (i *imup) postSpeedTestRealtimeResults(ctx context.Context, status string, testResult *speedTestData) error {
	res := struct {
		Data     string  `json:"data,omitempty"`
		Download float64 `json:"download,omitempty"`
		Upload   float64 `json:"upload,omitempty"`
	}{
		Data:     status,
		Download: testResult.DownloadMbps,
		Upload:   testResult.UploadMbps,
	}

	data := &realtimeApiPayload{
		ID: i.cfg.HostID(), Key: i.cfg.APIKey(), Email: i.cfg.EmailAddress(), Data: res,
	}

	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal: %v", err)
	}

	return sendRealtimeData(ctx, bytes.NewBuffer(b), i.SpeedTestResultsAddress)
}

// remoteConfigReload shares its config version with imup and determines if
// a new remote configuration is available
// this feature is WIP and not yet released
func (i *imup) remoteConfigReload(ctx context.Context) error {
	// NOTE: this feature is only intended for (org) clients running with an API key
	if i.cfg.APIKey() == "" {
		return nil
	}

	data := &realtimeApiPayload{
		ID:        i.cfg.HostID(),
		Email:     i.cfg.EmailAddress(),
		GroupID:   i.cfg.GroupID(),
		GroupName: i.cfg.Group(),
		Key:       i.cfg.APIKey(),
		Version:   i.cfg.Version(),
	}

	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := retryablehttp.NewRequest("POST", i.RealtimeConfig, bytes.NewBuffer(b))
	req = req.WithContext(ctx)
	if err != nil {
		return fmt.Errorf("creating request to check for a remote config %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := retryablehttp.NewClient()
	client.Backoff = exactJitterBackoff

	client.RetryMax = 50_000
	client.RetryWaitMin = time.Duration(30) * time.Second
	client.RetryWaitMax = time.Duration(60) * time.Second
	client.Logger = log.New(log.Default().Handler())

	resp, err := client.Do(req)
	if err != nil {
		if err == context.Canceled {
			return nil
		}

		return fmt.Errorf("error posting to %s :%s", i.RealtimeConfig, err)
	}
	defer resp.Body.Close()

	if retcode := resp.StatusCode; retcode == http.StatusOK {
		if data, err := io.ReadAll(resp.Body); err != nil {
			return fmt.Errorf("cannot read raw remote config from api %s", err)
		} else {
			i.reloadConfig(data)
		}
	} else if retcode == http.StatusNoContent {
		log.Debug("config has not changed")
	} else if i.cfg.Verbosity() == log.LevelDebug {
		log.Debug("unexpected response returned from api", "retcode", retcode)
	}

	return nil
}
