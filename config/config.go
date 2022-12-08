package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	gw "github.com/jackpal/gateway"
)

var (
	seedRandom sync.Once
	setupFlags sync.Once

	apiKey        *string
	email         *string
	environment   *string
	id            *string
	configVersion *string
	// TODO: implement an app wide file logger
	// logDirectory *string

	insecureSpeedTest  *bool
	noGatewayDiscovery *bool
	nonvolatile        *bool
	noSpeedTest        *bool
	pingEnabled        *bool
	quietSpeedTest     *bool
	realtimeEnabled    *bool

	mu sync.RWMutex
)

// TODO: implement an app wide file logger
// var fileLogger = log.New()

// Reloadable is the interface to a remote configuration
// this interface exposes read and write, thread safe methods
// allowing it to be accessed and written to concurrently
type Reloadable interface {
	APIKey() string
	DiscoverGateway() string
	EmailAddress() string
	Env() string
	HostID() string
	Version() string
	// TODO: implement an app wide file logger
	// LogDir() string

	Realtime() bool
	EnableRealtime()
	DevelopmentEnvironment() bool
	DisableRealtime()
	SpeedTests() bool
	StoreJobsOnDisk() bool
	InsecureSpeedTests() bool
	QuietSpeedTests() bool
	PingTests() bool
}

// cfg is intentionally declared in the global space, but un-exported
// its primary purpose is to provide synchronization of a read only
// reloadable configuration
var cfg *config

type config struct {
	ID            string `json:"ID"`
	Email         string `json:"Email"`
	Environment   string `json:"Environment"`
	Key           string `json:"Key"`
	ConfigVersion string `json:"Version"`
	// TODO: implement an app wide file logger
	// LogDirectory string `json:"LogDirectory"`

	InsecureSpeedTest bool `json:"InsecureSpeedTest"`
	NoDiscoverGateway bool `json:"NoDiscoverGateway"`
	Nonvolatile       bool `json:"Nonvolatile"`
	PingEnabled       bool `json:"PingEnabled"`
	QuietSpeedTest    bool `json:"QuietSpeedTest"`
	RealtimeEnabled   bool `json:"RealtimeEnabled"`
	SpeedTestEnabled  bool `json:"SpeedTestEnabled"`
}

type remoteConfigResp struct {
	CFG *config `json:"config"`
}

func New() (Reloadable, error) {
	// do not instantiate a new copy of config, use the package level global
	cfg = &config{}
	seedRandom.Do(func() {
		rand.Seed(time.Now().UTC().UnixNano())
	})

	setupFlags.Do(func() {
		apiKey = flag.String("key", "", "api key")
		email = flag.String("email", "", "email address")
		environment = flag.String("environment", "", "imUp environment (development, production)")
		id = flag.String("id", "", "host id")
		configVersion = flag.String("config-version", "", "config version")
		// TODO: implement an app wide file logger
		// logDirectory = flag.String("log-directory", "", "path to imUp log directory on filesystem")

		insecureSpeedTest = flag.Bool("insecure", false, "run insecure speed tests (ws:// and not wss://)")
		nonvolatile = flag.Bool("nonvolatile", false, "use disk to store collected data between tests to ensure reliability")
		noGatewayDiscovery = flag.Bool("no-gateway-discovery", false, "do not attempt to discover a default gateway")
		noSpeedTest = flag.Bool("no-speed-test", false, "don't run speed tests")
		pingEnabled = flag.Bool("ping", false, "use ICMP ping for connectivity tests")
		quietSpeedTest = flag.Bool("quiet-speed-test", false, "don't output speed test logs")
		realtimeEnabled = flag.Bool("realtime", true, "enable realtime features, enabled by default")

		flag.Parse()
	})

	hostname, _ := os.Hostname()

	cfg.ID = argOrEnvVar(id, "HOST_ID", hostname)
	// TODO: implement an app wide file logger
	// cfg.LogDirectory = argOrEnvVar(logDirectory, "IMUP_LOG_DIRECTORY", "")
	cfg.Email = argOrEnvVar(email, "EMAIL", "unknown")
	cfg.Environment = argOrEnvVar(environment, "ENVIRONMENT", "production")
	cfg.Key = argOrEnvVar(apiKey, "API_KEY", "")
	cfg.ConfigVersion = argOrEnvVar(configVersion, "CONFIG_VERSION", "dev-preview")

	cfg.SpeedTestEnabled = !argOrEnvVarBool(noSpeedTest, "NO_SPEED_TEST", false)
	cfg.InsecureSpeedTest = argOrEnvVarBool(insecureSpeedTest, "INSECURE_SPEED_TEST", false)
	cfg.NoDiscoverGateway = argOrEnvVarBool(noGatewayDiscovery, "NO_GATEWAY_DISCOVERY", false)
	cfg.Nonvolatile = argOrEnvVarBool(nonvolatile, "NONVOLATILE", false)
	cfg.QuietSpeedTest = argOrEnvVarBool(quietSpeedTest, "QUIET_SPEED_TEST", false)
	cfg.PingEnabled = argOrEnvVarBool(pingEnabled, "PING_ENABLED", false)
	cfg.RealtimeEnabled = argOrEnvVarBool(realtimeEnabled, "REALTIME", true)

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("configuration of client is not valid: %s", err)
	}

	return cfg, nil
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

	if err := c.CFG.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}

	mu.Lock()
	cfg = c.CFG
	defer mu.Unlock()
	return cfg, nil
}

func (c *config) DiscoverGateway() string {
	if g, err := gw.DiscoverGateway(); err != nil || c.NoDiscoverGateway {
		return ""
	} else {
		return g.String()
	}
}

func (cfg *config) validate() error {
	if (cfg.Email == "unknown" || cfg.Email == "") && (cfg.Key == "" || cfg.ID == "") {
		return fmt.Errorf("please supply an email address (--email) or api key and host id (--key, --id)!: email: %s, key: %s, id: %s", cfg.Email, cfg.Key, cfg.ID)
	}

	return nil
}

func argOrEnvVar(argVal *string, varName, defaultVal string) string {
	if *argVal != "" {
		return *argVal
	}

	return getEnv(varName, defaultVal)
}

// should use a pointer for everything so its possible to tell whats a default
func argOrEnvVarBool(argVal *bool, varName string, defaultVal bool) bool {
	if *argVal {
		return *argVal
	}

	valStr := getEnv(varName, "false")
	val, err := strconv.ParseBool(valStr)
	if err != nil {
		return false
	}
	return val
}

func getEnv(varName, defaultVal string) string {
	if value, isPresent := os.LookupEnv(varName); isPresent {
		return value
	}

	return defaultVal
}

func (c *config) APIKey() string {
	mu.RLock()
	defer mu.RUnlock()
	return c.Key
}

func (c *config) HostID() string {
	mu.RLock()
	defer mu.RUnlock()
	return c.ID
}

func (c *config) EmailAddress() string {
	mu.RLock()
	defer mu.RUnlock()
	return c.Email
}

func (c *config) Env() string {
	mu.RLock()
	defer mu.RUnlock()
	return c.Environment
}

func (c *config) Version() string {
	mu.RLock()
	defer mu.RUnlock()
	return c.ConfigVersion
}

// TODO: implement an app wide file logger
// func (c *config) LogDir() string {``
// 	mu.RLock()
// 	defer mu.RUnlock()
// 	return c.LogDirectory
// }

func (c *config) Realtime() bool {
	mu.RLock()
	defer mu.RUnlock()
	return c.RealtimeEnabled
}

func (c *config) DevelopmentEnvironment() bool {
	mu.RLock()
	defer mu.RUnlock()
	return c.Environment == "development"
}

func (c *config) DisableRealtime() {
	mu.Lock()
	defer mu.Unlock()
	c.RealtimeEnabled = false
}

func (c *config) EnableRealtime() {
	mu.Lock()
	defer mu.Unlock()
	c.RealtimeEnabled = true
}

func (c *config) StoreJobsOnDisk() bool {
	mu.RLock()
	defer mu.RUnlock()
	return c.Nonvolatile
}

func (c *config) SpeedTests() bool {
	mu.RLock()
	defer mu.RUnlock()
	return c.SpeedTestEnabled
}

func (c *config) InsecureSpeedTests() bool {
	mu.RLock()
	defer mu.RUnlock()
	return c.InsecureSpeedTest
}

func (c *config) QuietSpeedTests() bool {
	mu.RLock()
	defer mu.RUnlock()
	return c.QuietSpeedTest
}

func (c *config) PingTests() bool {
	mu.RLock()
	defer mu.RUnlock()
	return c.PingEnabled
}
