package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	gw "github.com/jackpal/gateway"

	"github.com/imup-io/client/util"
	log "golang.org/x/exp/slog"
)

type remoteConfigResp struct {
	CFG *config `json:"config"`
}

// Reload expects a payload that is compatible with a base reloadable config and
// will update the underlying global configuration.
func Reload(data []byte) (Reloadable, error) {
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

	// reload logger using configuration from API
	if reloadLogger {
		configureLogger(c.CFG.logLevel, w)
	}

	// lock the configuration
	mu.Lock()

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
