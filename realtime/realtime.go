package realtime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

type authRequest struct {
	ApiKey string `json:"apiKey,omitempty"`
	Email  string `json:"email,omitempty"`
}

func Authorized(ctx context.Context, apiKey, email, url string) (bool, error) {
	b, err := json.Marshal(authRequest{Email: email, ApiKey: apiKey})
	if err != nil {
		return false, fmt.Errorf("cannot marshal request body")
	}

	req, err := retryablehttp.NewRequest("POST", url, b)
	req = req.WithContext(ctx)
	if err != nil {
		return false, fmt.Errorf("NewRequest: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := retryablehttp.NewClient()
	client.Backoff = util.ExactJitterBackoff

	client.RetryMax = 50_000
	client.RetryWaitMin = time.Duration(30) * time.Second
	client.RetryWaitMax = time.Duration(60) * time.Second
	client.Logger = log.New(log.Default().Handler())

	if resp, err := client.Do(req); err != nil {
		if err == context.Canceled {
			return false, nil
		}
		return false, fmt.Errorf("cannot authorize client for realtime use: addr: %s, client.Do: %s", url, err)
	} else {
		if resp.StatusCode == http.StatusOK {
			return true, nil
		} else {
			return false, nil
		}
	}
}

func SendClientHealthy(ctx context.Context, url string) error {
	data := &apiPayload{
		ID: cfg.HostID(), Key: cfg.APIKey(), Email: cfg.EmailAddress(),
	}

	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("json.Marshal: %v", err)
	}

	return util.Send(ctx, bytes.NewBuffer(b), url)
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

	return util.Send(ctx, bytes.NewBuffer(b), url)
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

	return util.Send(ctx, bytes.NewBuffer(b), url)
}
