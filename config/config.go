package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/imup-io/client/util"
	log "golang.org/x/exp/slog"
)

var (
	setupFlags sync.Once

	apiKey         *string
	allowlistedIPs *string
	blocklistedIPs *string
	configVersion  *string
	email          *string
	groupID        *string
	hostID         *string
	verbosity      *string

	insecureSpeedTest  *bool
	logToFile          *bool
	noGatewayDiscovery *bool
	nonvolatile        *bool
	noSpeedTest        *bool
	pingEnabled        *bool
	realtimeEnabled    *bool

	mu sync.RWMutex
)

// Reloadable is the interface to a remote configuration
// this interface exposes read and write, thread safe methods
// allowing it to be accessed and written to concurrently
type Reloadable interface {
	APIKey() string
	DiscoverGateway() string
	EmailAddress() string
	Group() string
	HostID() string
	PublicIP() string
	RefreshPublicIP() string
	Version() string

	Realtime() bool
	LogToFile() bool
	SpeedTests() bool
	StoreJobsOnDisk() bool
	InsecureSpeedTests() bool
	PingTests() bool

	EnableRealtime()
	DisableRealtime()

	Verbosity() log.Level

	AllowedIPs() []string
	BlockedIPs() []string
}

// cfg is intentionally declared in the global space, but un-exported
// its primary purpose is to provide synchronization of a read only
// reloadable configuration
var cfg *config

type config struct {
	apiKey   string
	email    string
	hostID   string
	publicIP string

	logLevel log.Level

	ConfigVersion string `json:"version"`
	GroupID       string `json:"groupID"`
	LogLevel      string `json:"verbosity"`

	InsecureSpeedTest bool `json:"insecureSpeedTest"`
	FileLogger        bool `json:"fileLogger"`
	NoDiscoverGateway bool `json:"noDiscoverGateway"`
	Nonvolatile       bool `json:"nonvolatile"`
	PingEnabled       bool `json:"pingEnabled"`
	RealtimeEnabled   bool `json:"realtimeEnabled"`
	SpeedTestEnabled  bool `json:"speedTestEnabled"`

	AllowlistedIPs []string `json:"allowlisted_ips"`
	BlocklistedIPs []string `json:"blocklisted_ips"`
}

// New returns a freshly setup Reloadable config.
func New() (Reloadable, error) {
	// do not instantiate a new copy of config, use the package level global
	cfg = &config{}

	setupFlags.Do(func() {
		allowlistedIPs = flag.String("allowlisted-ips", "", "Allowed IPs for speed tests")
		apiKey = flag.String("key", "", "api key")
		blocklistedIPs = flag.String("blocklisted-ips", "", "Blocked IPs for speed tests")
		configVersion = flag.String("config-version", "", "config version") //todo: placeholder for reloadable configs
		email = flag.String("email", "", "email address")
		groupID = flag.String("group-id", "", "org users group id")
		hostID = flag.String("host-id", "", "host id")
		verbosity = flag.String("verbosity", "", "How verbose log output should be (Default Info)")

		insecureSpeedTest = flag.Bool("insecure", false, "run insecure speed tests (ws:// and not wss://)")
		logToFile = flag.Bool("log-to-file", false, "if enabled, will log to the default root directory to use for user-specific cached data")
		noGatewayDiscovery = flag.Bool("no-gateway-discovery", false, "do not attempt to discover a default gateway")
		noSpeedTest = flag.Bool("no-speed-test", false, "don't run speed tests")
		nonvolatile = flag.Bool("nonvolatile", false, "use disk to store collected data between tests to ensure reliability")
		pingEnabled = flag.Bool("ping", false, "use ICMP ping for connectivity tests")
		realtimeEnabled = flag.Bool("realtime", true, "enable realtime features, enabled by default")

		flag.Parse()
	})

	hostname, _ := os.Hostname()

	cfg.AllowlistedIPs = strings.Split(util.ValueOr(allowlistedIPs, "ALLOWLISTED_IPS", ""), ",")
	cfg.apiKey = util.ValueOr(apiKey, "API_KEY", "")
	cfg.BlocklistedIPs = strings.Split(util.ValueOr(blocklistedIPs, "BLOCKLISTED_IPS", ""), ",")
	cfg.ConfigVersion = util.ValueOr(configVersion, "CONFIG_VERSION", "dev-preview") //todo: placeholder for reloadable configs
	cfg.email = util.ValueOr(email, "EMAIL", "unknown")
	cfg.hostID = util.ValueOr(hostID, "HOST_ID", hostname)
	cfg.GroupID = util.ValueOr(groupID, "GROUP_ID", "production")

	cfg.InsecureSpeedTest = util.BooleanValueOr(insecureSpeedTest, "INSECURE_SPEED_TEST", "false")
	cfg.FileLogger = util.BooleanValueOr(logToFile, "LOG_TO_FILE", "false")
	cfg.NoDiscoverGateway = util.BooleanValueOr(noGatewayDiscovery, "NO_GATEWAY_DISCOVERY", "false")
	cfg.SpeedTestEnabled = !util.BooleanValueOr(noSpeedTest, "NO_SPEED_TEST", "false")
	cfg.Nonvolatile = util.BooleanValueOr(nonvolatile, "NONVOLATILE", "false")
	cfg.PingEnabled = util.BooleanValueOr(pingEnabled, "PING_ENABLED", "false")
	cfg.RealtimeEnabled = util.BooleanValueOr(realtimeEnabled, "REALTIME", "true")

	cfg.logLevel = util.LevelMap(verbosity, "VERBOSITY", "INFO")

	var w io.Writer
	if cfg.FileLogger {
		w = logToUserCache()
	} else {
		w = os.Stderr
	}
	configureLogger(cfg.logLevel, w)

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("configuration of client is not valid: %s", err)
	}

	return cfg, nil
}

func configureLogger(verbosity log.Level, w io.Writer) {
	h := log.HandlerOptions{Level: verbosity}.NewJSONHandler(w)
	log.SetDefault(log.New(h))
}

// // Reload expects a payload that is compatible with a base reloadable config and
// // will update the underlying global configuration.
// func Reload(data []byte) (Reloadable, error) {
// 	c := &remoteConfigResp{}
// 	if err := json.Unmarshal(data, c); err != nil {
// 		return nil, fmt.Errorf("cannot unmarshal new configuration: %v", err)
// 	}

// 	if cfg.ConfigVersion == c.CFG.ConfigVersion {
// 		return nil, fmt.Errorf("configuration matches existing config")
// 	}

// 	c.CFG.email = cfg.email
// 	c.CFG.hostID = cfg.hostID
// 	c.CFG.apiKey = cfg.apiKey

// 	if err := c.CFG.validate(); err != nil {
// 		return nil, fmt.Errorf("invalid configuration: %v", err)
// 	}

// 	mu.Lock()
// 	log.Info("imup config reloaded", "config", fmt.Sprintf("config: %+v", c.CFG))
// 	cfg = c.CFG
// 	defer mu.Unlock()
// 	return cfg, nil
// }

func logToUserCache() *os.File {
	cache, err := os.UserCacheDir()
	if err != nil {
		log.Error("$HOME is likely undefined", "error", err)
	}

	targetDir := filepath.Join(cache, "imup", "logs")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		log.Error("cannot create directory in user cache", "error", err)
	}

	f, err := os.OpenFile(targetDir+"/imup.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Error("cannot open file", "error", err)
	}

	return f
}

func (cfg *config) validate() error {
	if (cfg.email == "unknown" || cfg.email == "") && (cfg.apiKey == "" || cfg.hostID == "") {
		return fmt.Errorf("please supply an email address (--email) or api key and host id (--key, --id)!: email: %s, key: %s, id: %s", cfg.email, cfg.apiKey, cfg.hostID)
	}

	return nil
}

// Public Read Only (non reloadable) Configuration
//

// APIKey is an organization API key used for imUp.io's org product
func (c *config) APIKey() string {
	mu.RLock()
	defer mu.RUnlock()
	return c.apiKey
}

// HostID is the configured or local host id to associate test data with
func (c *config) HostID() string {
	mu.RLock()
	defer mu.RUnlock()
	return c.hostID
}

// EmailAddress the email address to associate test data with
func (c *config) EmailAddress() string {
	mu.RLock()
	defer mu.RUnlock()
	return c.email
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

// PublicIP retrieves the clients public ip address
func (c *config) PublicIP() string {
	mu.RLock()
	defer mu.RUnlock()
	return c.publicIP
}

// Realtime boolean indicating wether or not realtime features should be used
func (c *config) Realtime() bool {
	mu.RLock()
	defer mu.RUnlock()
	return c.RealtimeEnabled
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

func ips(ips []string) []string {
	hosts := []string{}
	for _, ip := range ips {
		if ip == "" {
			continue
		}

		if ipAddr, ipNet, err := net.ParseCIDR(ip); err != nil {
			log.Warn("cannot parse as cidr, assuming individual ip address", ip, err)
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
