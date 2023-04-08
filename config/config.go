package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/imup-io/client/util"
	gw "github.com/jackpal/gateway"
	"golang.org/x/exp/slog"
	log "golang.org/x/exp/slog"
)

var (
	setupFlags sync.Once

	apiKey         *string
	allowlistedIPs *string
	blocklistedIPs *string
	configVersion  *string
	email          *string
	environment    *string
	groupID        *string
	groupName      *string
	id             *string
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
	Group() string
	GroupID() string
	HostID() string
	PublicIP() string
	RefreshPublicIP() string
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

	AllowedIPs() []string
	BlockedIPs() []string
}

// cfg is intentionally declared in the global space, but un-exported
// its primary purpose is to provide synchronization of a read only
// reloadable configuration
var cfg *config

type config struct {
	id       string
	email    string
	key      string
	publicIP string

	ConfigVersion string `json:"version"`
	Environment   string `json:"environment"`
	GID           string `json:"groupID"`
	GroupName     string `json:"groupName"`
	// TODO: implement an app wide file logger
	// LogDirectory string `json:"LogDirectory"`

	InsecureSpeedTest bool `json:"insecureSpeedTest"`
	NoDiscoverGateway bool `json:"noDiscoverGateway"`
	Nonvolatile       bool `json:"nonvolatile"`
	PingEnabled       bool `json:"pingEnabled"`
	QuietSpeedTest    bool `json:"quietSpeedTest"`
	RealtimeEnabled   bool `json:"realtimeEnabled"`
	SpeedTestEnabled  bool `json:"speedTestEnabled"`

	AllowlistedIPs []string `json:"allowlisted_ips"`
	BlocklistedIPs []string `json:"blocklisted_ips"`
}

type remoteConfigResp struct {
	CFG *config `json:"config"`
}

// New returns a freshly setup Reloadable config.
func New() (Reloadable, error) {
	// do not instantiate a new copy of config, use the package level global
	cfg = &config{}

	setupFlags.Do(func() {
		apiKey = flag.String("key", "", "api key")
		allowlistedIPs = flag.String("allowlisted-ips", "", "Allowed IPs for speed tests")
		blocklistedIPs = flag.String("blocklisted-ips", "", "Blocked IPs for speed tests")
		configVersion = flag.String("config-version", "", "config version")
		email = flag.String("email", "", "email address")
		environment = flag.String("environment", "", "imUp environment (development, production)")
		groupID = flag.String("group-id", "", "org users group id")
		groupName = flag.String("group-name", "", "org users group name")
		id = flag.String("id", "", "host id")
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

	cfg.id = util.ValueOr(id, "HOST_ID", hostname)
	// TODO: implement an app wide file logger
	// cfg.LogDirectory = argOrEnvVar(logDirectory, "IMUP_LOG_DIRECTORY", "")
	cfg.AllowlistedIPs = strings.Split(util.ValueOr(allowlistedIPs, "ALLOWLISTED_IPS", ""), ",")
	cfg.BlocklistedIPs = strings.Split(util.ValueOr(blocklistedIPs, "BLOCKLISTED_IPS", ""), ",")
	cfg.ConfigVersion = util.ValueOr(configVersion, "CONFIG_VERSION", "dev-preview")
	cfg.email = util.ValueOr(email, "EMAIL", "unknown")
	cfg.Environment = util.ValueOr(environment, "ENVIRONMENT", "production")
	cfg.GID = util.ValueOr(groupID, "GROUP_ID", "production")
	cfg.GroupName = util.ValueOr(groupName, "GROUP_NAME", "production")
	cfg.key = util.ValueOr(apiKey, "API_KEY", "")

	cfg.SpeedTestEnabled = !util.BooleanValueOr(noSpeedTest, "NO_SPEED_TEST", "false")
	cfg.InsecureSpeedTest = util.BooleanValueOr(insecureSpeedTest, "INSECURE_SPEED_TEST", "false")
	cfg.NoDiscoverGateway = util.BooleanValueOr(noGatewayDiscovery, "NO_GATEWAY_DISCOVERY", "false")
	cfg.Nonvolatile = util.BooleanValueOr(nonvolatile, "NONVOLATILE", "false")
	cfg.QuietSpeedTest = util.BooleanValueOr(quietSpeedTest, "QUIET_SPEED_TEST", "false")
	cfg.PingEnabled = util.BooleanValueOr(pingEnabled, "PING_ENABLED", "false")
	cfg.RealtimeEnabled = util.BooleanValueOr(realtimeEnabled, "REALTIME", "true")

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

	c.CFG.email = cfg.email
	c.CFG.id = cfg.id
	c.CFG.key = cfg.key

	if err := c.CFG.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}

	mu.Lock()
	log.Info("imup config reloaded", "config", fmt.Sprintf("config: %+v", c.CFG))
	cfg = c.CFG
	defer mu.Unlock()
	return cfg, nil
}

// DiscoverGateway provides for automatic gateway discovery
func (c *config) DiscoverGateway() string {
	if g, err := gw.DiscoverGateway(); err != nil || c.NoDiscoverGateway {
		return ""
	} else {
		return g.String()
	}
}

func (cfg *config) validate() error {
	if (cfg.email == "unknown" || cfg.email == "") && (cfg.key == "" || cfg.id == "") {
		return fmt.Errorf("please supply an email address (--email) or api key and host id (--key, --id)!: email: %s, key: %s, id: %s", cfg.email, cfg.key, cfg.id)
	}

	return nil
}

// APIKey is an organization API key used for imUp.io's org product
func (c *config) APIKey() string {
	mu.RLock()
	defer mu.RUnlock()
	return c.key
}

// HostID is the configured or local host id to associate test data with
func (c *config) HostID() string {
	mu.RLock()
	defer mu.RUnlock()
	return c.id
}

// EmailAddress the email address to associate test data with
func (c *config) EmailAddress() string {
	mu.RLock()
	defer mu.RUnlock()
	return c.email
}

// Env production or development, used for realtime error tracking
func (c *config) Env() string {
	mu.RLock()
	defer mu.RUnlock()
	return c.Environment
}

// GroupID is the logical name for a group of org hosts
func (c *config) GroupID() string {
	mu.RLock()
	defer mu.RUnlock()
	return c.GID
}

// Group is the human readable name for a group of org hosts
func (c *config) Group() string {
	mu.RLock()
	defer mu.RUnlock()
	return c.GroupName
}

// Version returns the current version of package config
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

// Realtime boolean indicating wether or not realtime features should be used
func (c *config) Realtime() bool {
	mu.RLock()
	defer mu.RUnlock()
	return c.RealtimeEnabled
}

// DevelopmentEnvironment turns verbose logging on for some functions
func (c *config) DevelopmentEnvironment() bool {
	mu.RLock()
	defer mu.RUnlock()
	return c.Environment == "development"
}

// DisableRealtime turns off the imUp.io realtime feature set
func (c *config) DisableRealtime() {
	mu.Lock()
	defer mu.Unlock()
	c.RealtimeEnabled = false
}

// EnableRealtime enables the imUp.io realtime feature set
func (c *config) EnableRealtime() {
	mu.Lock()
	defer mu.Unlock()
	c.RealtimeEnabled = true
}

// StoreJobsOnDisk allows for extra redundancy between test by not caching test data in memory
func (c *config) StoreJobsOnDisk() bool {
	mu.RLock()
	defer mu.RUnlock()
	return c.Nonvolatile
}

// SpeedTests allow client to periodically run speed tests, per the NDT7 specification
func (c *config) SpeedTests() bool {
	mu.RLock()
	defer mu.RUnlock()
	return c.SpeedTestEnabled
}

// InsecureSpeedTests ndt7 configurable field
func (c *config) InsecureSpeedTests() bool {
	mu.RLock()
	defer mu.RUnlock()
	return c.InsecureSpeedTest
}

// QuietSpeedTests suppress speed test output
func (c *config) QuietSpeedTests() bool {
	mu.RLock()
	defer mu.RUnlock()
	return c.QuietSpeedTest
}

// PingTests determines if connectivity should use ICMP requests
func (c *config) PingTests() bool {
	mu.RLock()
	defer mu.RUnlock()
	return c.PingEnabled
}

// PublicIP retrieves the clients public ip address
func (c *config) PublicIP() string {
	mu.RLock()
	defer mu.RUnlock()
	return c.publicIP
}

// RefreshPublicIP uses an open api to retrieve the clients public ip address
func (c *config) RefreshPublicIP() string {
	ip, err := getIP()
	if err != nil {
		log.Warn("cannot get public ip", err)
		return c.publicIP
	}

	if ip != c.publicIP {
		mu.Lock()
		log.Debug("setting publicIP", "publicIP", ip)
		c.publicIP = ip
		defer mu.Unlock()
	}

	return c.publicIP
}

func getIP() (string, error) {
	req, err := http.Get("https://api64.ipify.org?format=json")
	if err != nil {
		return "", err
	}
	defer req.Body.Close()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return "", err
	}

	type IP struct {
		IP string `json:"ip"`
	}
	var ip IP
	if err := json.Unmarshal(body, &ip); err != nil {
		return "", err
	}

	return ip.IP, nil
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

func ips(ips []string) []string {
	hosts := []string{}
	for _, ip := range ips {
		if ip == "" {
			continue
		}

		if ipAddr, ipNet, err := net.ParseCIDR(ip); err != nil {
			slog.Warn("cannot parse as cidr, assuming individual ip address", ip, err)
			hosts = append(hosts, ip)
		} else {
			for ip := ipAddr.Mask(ipNet.Mask); ipNet.Contains(ip); incrementIPs(ip) {
				hosts = append(hosts, ip.String())
			}
		}
	}

	return hosts
}

func incrementIPs(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
