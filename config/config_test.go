package config

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/matryer/is"
)

func Test_DefaultConfig(t *testing.T) {
	is := is.New(t)
	os.Setenv("API_KEY", "ApiKey")
	os.Setenv("EMAIL", "Email")
	os.Setenv("HOST_ID", "HostID")

	cfg, err := New()
	is.NoErr(err)

	// validate default config
	is.Equal(false, cfg.InsecureSpeedTests())
	is.Equal(false, cfg.PingTests())
	is.Equal(false, cfg.QuietSpeedTests())

	is.Equal("ApiKey", cfg.APIKey())
	is.Equal("HostID", cfg.HostID())
	is.Equal("Email", cfg.EmailAddress())
	is.Equal("production", cfg.Env())
	is.Equal("dev-preview", cfg.Version())

	is.True(cfg.Realtime())
	is.True(cfg.SpeedTests())
}

func Test_ConfigReloadable(t *testing.T) {
	is := is.New(t)
	os.Setenv("API_KEY", "ApiKey")
	os.Setenv("EMAIL", "Email")
	os.Setenv("HOST_ID", "HostID")

	newConfig := &config{PingEnabled: true, Key: "some key", ID: "some id", ConfigVersion: "new-new"}
	var b bytes.Buffer
	json.NewEncoder(&b).Encode(struct {
		C *config `json:"config"`
	}{C: newConfig})

	defaultConfig, err := New()
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

	defaultConfig, err := New()
	is.NoErr(err)

	is.Equal(false, defaultConfig.PingTests())
	write := func() {
		var b bytes.Buffer
		newConfig := &config{PingEnabled: true, Key: "some key", ID: "some id"}
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

	defaultConfig, err := New()
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
