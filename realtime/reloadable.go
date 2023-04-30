package realtime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	gw "github.com/jackpal/gateway"

	"github.com/imup-io/client/util"
	log "golang.org/x/exp/slog"
)

type remoteConfigResp struct {
	CFG *config `json:"config"`
}

// RemoteConfigReload shares its config version with imup and determines if
// a new remote configuration is available
// this feature is WIP and not yet released
func RemoteConfigReload(ctx context.Context, url string) (Reloadable, error) {
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

		return nil, fmt.Errorf("error posting to %s :%s", url, err)
	}
	defer resp.Body.Close()

	if retcode := resp.StatusCode; retcode == http.StatusOK {
		if data, err := io.ReadAll(resp.Body); err != nil {
			return nil, fmt.Errorf("cannot read raw remote config from api %s", err)
		} else {
			if cfg, err := reload(data); err != nil {
				log.Error("cannot reload config", "error", err)
			} else {
				return cfg, nil
			}
		}
	} else if retcode == http.StatusNoContent {
		log.Info("config has not changed")
	} else {
		log.Warn("unexpected response returned from api", "retcode", retcode)
	}

	return nil, nil
}

// Reload expects a payload that is compatible with a base reloadable config and
// will update the underlying global configuration.
func reload(data []byte) (Reloadable, error) {
	c := &remoteConfigResp{}
	if err := json.Unmarshal(data, c); err != nil {
		return nil, fmt.Errorf("cannot unmarshal new configuration: %v", err)
	}

	if cfg.ConfigVersion == c.CFG.ConfigVersion {
		return nil, fmt.Errorf("configuration matches existing config")
	}

	// keep existing configuration for non reloadable private fields
	c.CFG.email = cfg.email
	c.CFG.id = cfg.id
	c.CFG.key = cfg.key

	var reloadLogger bool
	if logLevel := util.LevelMap(&c.CFG.LogLevel, "VERBOSITY", "INFO"); logLevel != cfg.logLevel && c.CFG.LogLevel != "" {
		cfg.logLevel = logLevel
		reloadLogger = true
	}

	var w io.Writer
	if c.CFG.FileLogger != cfg.FileLogger {
		reloadLogger = true

		if c.CFG.FileLogger {
			w = logToUserCache()
		} else {
			w = os.Stderr
		}
	}

	if err := c.CFG.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}

	// lock the configuration
	mu.Lock()

	// reload logger using configuration from API
	if reloadLogger {
		configureLogger(c.CFG.logLevel, w)
	}

	// refresh a clients public IP after a config reload
	if ip, err := getIP(); err != nil {
		log.Info("cannot get public ip", "error", err)

	} else {
		cfg.publicIP = ip
	}

	log.Info("imup config reloaded", "config", fmt.Sprintf("config: %+v", c.CFG))

	cfg = c.CFG
	defer mu.Unlock()
	return cfg, nil
}

// AllowedIPs returns a reloadable list of allow-listed ips for running speed tests
func (c *config) AllowedIPs() []string {
	mu.RLock()
	defer mu.RUnlock()
	return ips(c.AllowlistedIPs)
}

// AllowedIPs returns a reloadable list of block-listed ips to avoid running speed tests for
func (c *config) BlockedIPs() []string {
	mu.RLock()
	defer mu.RUnlock()
	return ips(c.BlocklistedIPs)
}

// DiscoverGateway provides for automatic gateway discovery
func (c *config) DiscoverGateway() string {
	if g, err := gw.DiscoverGateway(); err != nil || c.NoDiscoverGateway {
		return ""
	} else {
		return g.String()
	}
}

// GroupID is the logical name for a group of org hosts
func (c *config) GroupID() string {
	mu.RLock()
	defer mu.RUnlock()
	return c.GID
}

// InsecureSpeedTests ndt7 configurable field
func (c *config) InsecureSpeedTests() bool {
	mu.RLock()
	defer mu.RUnlock()
	return c.InsecureSpeedTest
}

// LogToFile indicates if log output should be written to file
func (c *config) LogToFile() bool {
	mu.RLock()
	defer mu.RUnlock()
	return c.FileLogger
}

// PingTests determines if connectivity should use ICMP requests
func (c *config) PingTests() bool {
	mu.RLock()
	defer mu.RUnlock()
	return c.PingEnabled
}

// SpeedTests allow client to periodically run speed tests, per the NDT7 specification
func (c *config) SpeedTests() bool {
	mu.RLock()
	defer mu.RUnlock()
	return c.SpeedTestEnabled
}

// StoreJobsOnDisk allows for extra redundancy between test by not caching test data in memory
func (c *config) StoreJobsOnDisk() bool {
	mu.RLock()
	defer mu.RUnlock()
	return c.Nonvolatile
}

// Verbosity sets the level at which the client outputs logs
func (c *config) Verbosity() log.Level {
	mu.RLock()
	defer mu.RUnlock()
	return c.logLevel.Level()
}

// Version returns the current version of package config
func (c *config) Version() string {
	mu.RLock()
	defer mu.RUnlock()
	return c.ConfigVersion
}
