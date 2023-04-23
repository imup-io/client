package realtime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/imup-io/client/util"

	log "golang.org/x/exp/slog"
)

// ApiPayload is the imup realtime structure expected by the api
type apiPayload struct {
	ID      string      `json:"hostId,omitempty"`
	Email   string      `json:"email,omitempty"`
	GroupID string      `json:"groupID,omitempty"`
	Key     string      `json:"apiKey,omitempty"`
	Version string      `json:"version,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func SendClientHealthy(ctx context.Context, url string) error {
	data := &apiPayload{
		ID: cfg.HostID(), Key: cfg.APIKey(), Email: cfg.EmailAddress(),
	}

	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("json.Marshal: %v", err)
	}

	return send(ctx, bytes.NewBuffer(b), url)
}

func ShouldRunSpeedtest(ctx context.Context, url string) (bool, error) {
	data := &apiPayload{
		ID: cfg.HostID(), Key: cfg.APIKey(), Email: cfg.EmailAddress(),
	}

	b, err := json.Marshal(data)
	if err != nil {
		return false, fmt.Errorf("json.Marshal: %v", err)
	}

	req, err := retryablehttp.NewRequest("POST", url, bytes.NewBuffer(b))
	req = req.WithContext(ctx)
	if err != nil {
		return false, fmt.Errorf("NewRequest: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := retryablehttp.NewClient()
	client.Backoff = util.ExactJitterBackoff

	client.RetryMax = 2
	client.RetryWaitMin = time.Duration(30) * time.Second
	client.RetryWaitMax = time.Duration(60) * time.Second
	client.Logger = log.New(log.Default().Handler())

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("addr: %s, client.Do: %v", url, err)
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

func PostSpeedTestRealtimeStatus(ctx context.Context, status, url string) error {
	data := &apiPayload{
		ID: cfg.HostID(), Key: cfg.APIKey(), Email: cfg.EmailAddress(), Data: status,
	}

	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("json.Marshal: %v", err)
	}

	return send(ctx, bytes.NewBuffer(b), url)
}

func PostSpeedTestRealtimeResults(ctx context.Context, status, url string, up, down float64) error {
	res := struct {
		Data     string  `json:"data,omitempty"`
		Download float64 `json:"download,omitempty"`
		Upload   float64 `json:"upload,omitempty"`
	}{
		Data:     status,
		Download: down,
		Upload:   up,
	}

	data := &apiPayload{
		ID: cfg.HostID(), Key: cfg.APIKey(), Email: cfg.EmailAddress(), Data: res,
	}

	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal: %v", err)
	}

	return send(ctx, bytes.NewBuffer(b), url)
}

// RemoteConfigReload shares its config version with imup and determines if
// a new remote configuration is available
// this feature is WIP and not yet released
func RemoteConfigReload(ctx context.Context, url string) (*config, error) {
	// NOTE: this feature is only intended for (org) clients running with an API key
	if cfg.APIKey() == "" {
		return nil, errors.New("realtime remote config requires an api key")
	}

	data := &apiPayload{
		ID:      cfg.HostID(),
		Email:   cfg.EmailAddress(),
		GroupID: cfg.GroupID(),
		Key:     cfg.APIKey(),
		Version: cfg.Version(),
	}

	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := retryablehttp.NewRequest("POST", url, bytes.NewBuffer(b))
	req = req.WithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating request to check for a remote config %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := retryablehttp.NewClient()
	client.Backoff = util.ExactJitterBackoff

	client.RetryMax = 50_000
	client.RetryWaitMin = time.Duration(30) * time.Second
	client.RetryWaitMax = time.Duration(60) * time.Second
	client.Logger = log.New(log.Default().Handler())

	resp, err := client.Do(req)
	if err != nil {
		if err == context.Canceled {
			return nil, err
		}

		return nil, fmt.Errorf("error posting to %s :%s", i.RealtimeConfig, err)
	}
	defer resp.Body.Close()

	if retcode := resp.StatusCode; retcode == http.StatusOK {
		if data, err := io.ReadAll(resp.Body); err != nil {
			return nil, fmt.Errorf("cannot read raw remote config from api %s", err)
		} else {
			i.reloadConfig(data)
		}
	} else if retcode == http.StatusNoContent {
		log.Debug("config has not changed")
	} else if cfg.Verbosity() == log.LevelDebug {
		log.Debug("unexpected response returned from api", "retcode", retcode)
	}

	return nil, nil
}

func send(ctx context.Context, b *bytes.Buffer, addr string) error {
	req, err := retryablehttp.NewRequest("POST", addr, b)
	req = req.WithContext(ctx)
	if err != nil {
		return fmt.Errorf("NewRequest: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := retryablehttp.NewClient()
	client.Backoff = util.ExactJitterBackoff
	client.RetryMax = 3
	client.RetryWaitMin = time.Duration(200) * time.Millisecond
	client.RetryWaitMax = time.Duration(3) * time.Second
	client.Logger = log.New(log.Default().Handler())

	if _, err := client.Do(req); err != nil {
		return fmt.Errorf("addr: %s, client.Do: %s", addr, err)
	}

	return nil
}
