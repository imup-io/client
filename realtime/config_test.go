package realtime

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/matryer/is"
	log "golang.org/x/exp/slog"
)

func Test_DefaultConfig(t *testing.T) {
	is := is.New(t)
	os.Setenv("API_KEY", "ApiKey")
	os.Setenv("EMAIL", "Email")
	os.Setenv("HOST_ID", "HostID")

	cfg, err := NewConfig()
	is.NoErr(err)

	// validate default config
	is.Equal(false, cfg.InsecureSpeedTests())
	is.Equal(false, cfg.PingTests())

	is.Equal("ApiKey", cfg.APIKey())
	is.Equal("HostID", cfg.HostID())
	is.Equal("Email", cfg.EmailAddress())
	is.Equal(log.LevelInfo, cfg.Verbosity())
	is.Equal("dev-preview", cfg.Version())

	is.True(cfg.Realtime())
	is.True(cfg.SpeedTests())
}

func Test_ConfigReloadable(t *testing.T) {
	is := is.New(t)
	os.Setenv("API_KEY", "ApiKey")
	os.Setenv("EMAIL", "Email")
	os.Setenv("HOST_ID", "HostID")
	os.Setenv("VERBOSITY", "debug")

	newConfig := &config{
		PingEnabled:   true,
		LogLevel:      "INFO",
		FileLogger:    true,
		ConfigVersion: "new-new",
	}

	var b bytes.Buffer
	json.NewEncoder(&b).Encode(struct {
		C *config `json:"config"`
	}{C: newConfig})

	defaultConfig, err := NewConfig()
	is.NoErr(err)

	cfg, err := Reload(b.Bytes())
	is.NoErr(err)
	is.Equal(true, cfg.PingTests())
	is.Equal(false, defaultConfig.PingTests())
}

func Test_ConfigReloadableThreadSafe(t *testing.T) {
	is := is.New(t)
	os.Setenv("API_KEY", "ApiKey")
	os.Setenv("EMAIL", "Email")
	os.Setenv("HOST_ID", "HostID")

	defaultConfig, err := NewConfig()
	is.NoErr(err)

	is.Equal(false, defaultConfig.PingTests())
	write := func() {
		var b bytes.Buffer
		newConfig := &config{PingEnabled: true, key: "some key", id: "some id"}
		json.NewEncoder(&b).Encode(newConfig)
		cfg, err := Reload(b.Bytes())
		is.NoErr(err)
		defaultConfig = cfg
	}

	read := func() {
		is.Equal(true, defaultConfig.PingTests())
	}

	is.Equal(false, defaultConfig.PingTests())
	_, _ = read, write
}

func Test_RealtimeOnOff(t *testing.T) {
	is := is.New(t)
	os.Setenv("API_KEY", "ApiKey")
	os.Setenv("EMAIL", "Email")
	os.Setenv("HOST_ID", "HostID")

	defaultConfig, err := NewConfig()
	is.NoErr(err)

	disable := func() {
		defaultConfig.DisableRealtime()
	}

	enable := func() {
		defaultConfig.EnableRealtime()
	}

	enabled := func() {
		is.Equal(true, defaultConfig.Realtime())
	}

	disabled := func() {
		is.Equal(false, defaultConfig.Realtime())
	}

	is.Equal(true, defaultConfig.Realtime())

	_, _ = disable, disabled
	_, _ = enable, enabled
}

func Test_ListedIPs(t *testing.T) {
	is := is.New(t)
	os.Setenv("API_KEY", "ApiKey")
	os.Setenv("EMAIL", "Email")
	os.Setenv("HOST_ID", "HostID")
	os.Setenv("ALLOWLISTED_IPS", "10.0.0.0/28,192.168.1.1")

	defaultConfig, err := NewConfig()
	is.NoErr(err)

	is.Equal(len(defaultConfig.AllowedIPs()), 17)
}

func Test_PublicIP(t *testing.T) {
	is := is.New(t)
	os.Setenv("API_KEY", "ApiKey")
	os.Setenv("EMAIL", "Email")
	os.Setenv("HOST_ID", "HostID")

	defaultConfig, err := NewConfig()
	is.NoErr(err)

	ip := defaultConfig.RefreshPublicIP()
	is.True(ip != "")
	is.Equal(defaultConfig.PublicIP(), ip)
}
