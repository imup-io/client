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
	"strconv"
	"strings"
	"sync"

	"github.com/imup-io/client/util"
	log "golang.org/x/exp/slog"
)

// NOTE: ImUpAPIHost is set via build flags
var ImUpAPIHost = "https://api.imup.io"

var (
	setupFlags sync.Once

	allowlistedIPs               *string
	apiKey                       *string
	apiPostConnectionData        *string
	apiPostSpeedTestData         *string
	blocklistedIPs               *string
	configVersion                *string
	connDelay                    *string
	connInterval                 *string
	connRequests                 *string
	email                        *string
	groupID                      *string
	hostID                       *string
	imupDataLength               *string
	logFile                      *string
	livenessCheckInAddress       *string
	pingAddressesExternal        *string
	pingAddressInternal          *string
	pingDelay                    *string
	pingInterval                 *string
	pingRequests                 *string
	realtimeAuthorized           *string
	realtimeConfig               *string
	shouldRunSpeedTestAddress    *string
	speedTestResultsAddress      *string
	speedTestStatusUpdateAddress *string
	verbosity                    *string

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
	EmailAddress() string
	GroupID() string
	HostID() string
	PublicIP() string
	RefreshPublicIP() string
	Version() string

	Realtime() bool
	SpeedTests() bool
	StoreJobsOnDisk() bool
	InsecureSpeedTests() bool
	PingTests() bool

	EnableRealtime()
	DisableRealtime()

	Verbosity() log.Level

	AllowedIPs() []string
	BlockedIPs() []string

	PostConnectionData() string
	PostSpeedTestData() string
	LivenessCheckInURL() string
	ShouldRunSpeedTestURL() string
	SpeedTestResultsURL() string
	SpeedTestStatusUpdateURL() string
	RealtimeAuth() string
	RealtimeConfigURL() string
	PingAddresses() []string
	InternalPingAddress() string
	PingIntervalSeconds() int
	ConnIntervalSeconds() int
	PingDelayMilli() int
	ConnDelayMilli() int
	PingRequestsCount() int
	ConnRequestsCount() int
	IMUPDataLen() int
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

	APIPostConnectionData        string
	APIPostSpeedTestData         string
	LivenessCheckInAddress       string
	PingAddressInternal          string
	RealtimeAuthorized           string
	RealtimeConfig               string
	ShouldRunSpeedTestAddress    string
	SpeedTestResultsAddress      string
	SpeedTestStatusUpdateAddress string

	ConnDelay      int
	ConnInterval   int
	ConnRequests   int
	IMUPDataLength int
	PingDelay      int
	PingInterval   int
	PingRequests   int

	PingAddressesExternal []string

	// reloadable elements
	ConfigVersion string `json:"version"`
	Group         string `json:"group_id"`
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
	mu.Lock()
	defer mu.Unlock()
	// do not instantiate a new copy of config, use the package level global
	cfg = &config{}

	setupFlags.Do(func() {
		allowlistedIPs = flag.String("allowlisted-ips", "", "comma separated list of CIDR strings to match against host IP that determines whether speed and connectivity testing will be run, default is allow all")
		apiKey = flag.String("key", "", "an api key associated with an imup organization")
		apiPostConnectionData = flag.String("api-post-connection-data", "", fmt.Sprintf("api endpoint for connectivity data ingestion, default is %s/v1/data/connectivity", ImUpAPIHost))
		apiPostSpeedTestData = flag.String("api-post-speed-test-data", "", fmt.Sprintf("api endpoint for speed data ingestion, default is %s/v1/data/speedtest", ImUpAPIHost))
		blocklistedIPs = flag.String("blocklisted-ips", "", "comma separated list of CIDR strings to match against host IP that determines whether speed and connectivity testing will be paused, default is block none")
		configVersion = flag.String("config-version", "", "config version for realtime reloadable configs") //todo: placeholder for reloadable configs
		connDelay = flag.String("conn-delay", "", "the delay between connectivity tests with a net dialer (milliseconds), default is 200")
		connInterval = flag.String("conn-interval", "", "how often a dial test is run (seconds), default is 60")
		connRequests = flag.String("conn-requests", "", "the number of dials executed during a connectivity test, default is 300")
		email = flag.String("email", "", "email address associated with the gathered connectivity and speed data")
		groupID = flag.String("group-id", "", "an imup org users group id")
		hostID = flag.String("host-id", "", "the host id associated with the gathered connectivity and speed data")
		imupDataLength = flag.String("imup-data-length", "", "the number of data points collected before sending data to the api, default is 15 data points")
		logFile = flag.String("log-file", "", "writes all logs to this file path, default is unset")
		livenessCheckInAddress = flag.String("liveness-check-in-address", "", fmt.Sprintf("api endpoint for liveness checkins default is %s/v1/realtime/livenesscheckin", ImUpAPIHost))
		pingAddressesExternal = flag.String("ping-addresses-external", "", "external IP addresses imup will use to validate connectivity, defaults are 1.1.1.1/32,1.0.0.1/32,8.8.8.8/32,8.8.4.4/32")
		pingAddressInternal = flag.String("ping-address-internal", "", "an internal gateway to differentiate between local networking issues and internet connectivity, by default imup attempts to discover your gateway")
		pingDelay = flag.String("ping-delay", "", "the delay between connectivity tests with ping (milliseconds), default is 100")
		pingInterval = flag.String("ping-interval", "", "how often a ping test is run (seconds), default is 60")
		pingRequests = flag.String("ping-requests", "", "the number of icmp echos executed during a ping test, default is 600")
		realtimeAuthorized = flag.String("realtime-authorized", "", fmt.Sprintf("api endpoint for imup real-time features, default is %s/v1/auth/realtimeAuthorized", ImUpAPIHost))
		realtimeConfig = flag.String("realtime-config", "", fmt.Sprintf("api endpoint for imup realtime reloadable configuration, default is %s/v1/realtime/config", ImUpAPIHost))
		shouldRunSpeedTestAddress = flag.String("should-run-speed-test-address", "", fmt.Sprintf("api endpoint for imup realtime speed tests, default is %s/v1/realtime/shouldClientRunSpeedTest", ImUpAPIHost))
		speedTestResultsAddress = flag.String("speed-test-results-address", "", fmt.Sprintf("api endpoint for imup realtime speed test results, default is %s/v1/realtime/speedTestResults", ImUpAPIHost))
		speedTestStatusUpdateAddress = flag.String("speed-test-status-update-address", "", fmt.Sprintf("api endpoint for imup real-time speed test status updates, default is %s/v1/realtime/speedTestStatusUpdate", ImUpAPIHost))
		verbosity = flag.String("verbosity", "", "verbosity for log output [debug, info, warn, error], default is info")

		insecureSpeedTest = flag.Bool("insecure", false, "run insecure speed tests (ws:// and not wss://), default is false")
		logToFile = flag.Bool("log-to-file", false, "if enabled, will log to the default root directory to use for user-specified cached data, default is false")
		noGatewayDiscovery = flag.Bool("no-gateway-discovery", false, "do not attempt to discover a default gateway, default is true")
		noSpeedTest = flag.Bool("no-speed-test", false, "do not run speed tests, default is false")
		nonvolatile = flag.Bool("nonvolatile", false, "use disk to store collected data between tests to ensure no lost data, default is false to be minimally invasive")
		pingEnabled = flag.Bool("ping", true, "use ICMP ping for connectivity tests, default is true")
		realtimeEnabled = flag.Bool("realtime", true, "enable realtime features, default is true")

		flag.Parse()
	})

	hostname, _ := os.Hostname()

	cfg.apiKey = util.ValueOr(apiKey, "API_KEY", "")
	cfg.email = util.ValueOr(email, "EMAIL", "unknown")
	cfg.hostID = util.ValueOr(hostID, "HOST_ID", hostname)

	cfg.AllowlistedIPs = strings.Split(util.ValueOr(allowlistedIPs, "ALLOWLISTED_IPS", ""), ",")
	cfg.APIPostConnectionData = util.ValueOr(apiPostConnectionData, "IMUP_ADDRESS", fmt.Sprintf("%s/v1/data/connectivity", ImUpAPIHost))
	cfg.APIPostSpeedTestData = util.ValueOr(apiPostSpeedTestData, "IMUP_ADDRESS_SPEEDTEST", fmt.Sprintf("%s/v1/data/speedtest", ImUpAPIHost))
	cfg.BlocklistedIPs = strings.Split(util.ValueOr(blocklistedIPs, "BLOCKLISTED_IPS", ""), ",")
	cfg.ConfigVersion = util.ValueOr(configVersion, "CONFIG_VERSION", "dev-preview") //todo: placeholder for reloadable configs
	cfg.Group = util.ValueOr(groupID, "GROUP_ID", "")

	cfg.PingAddressInternal = util.ValueOr(pingAddressInternal, "PING_ADDRESS_INTERNAL", cfg.discoverGateway())
	cfg.LivenessCheckInAddress = util.ValueOr(livenessCheckInAddress, "IMUP_LIVENESS_CHECKIN_ADDRESS", fmt.Sprintf("%s/v1/realtime/livenesscheckin", ImUpAPIHost))
	cfg.RealtimeAuthorized = util.ValueOr(realtimeAuthorized, "IMUP_REALTIME_AUTHORIZED", fmt.Sprintf("%s/v1/auth/realtimeAuthorized", ImUpAPIHost))
	cfg.RealtimeConfig = util.ValueOr(realtimeConfig, "IMUP_REALTIME_CONFIG", fmt.Sprintf("%s/v1/realtime/config", ImUpAPIHost))
	cfg.ShouldRunSpeedTestAddress = util.ValueOr(shouldRunSpeedTestAddress, "IMUP_SHOULD_RUN_SPEEDTEST_ADDRESS", fmt.Sprintf("%s/v1/realtime/shouldClientRunSpeedTest", ImUpAPIHost))
	cfg.SpeedTestResultsAddress = util.ValueOr(speedTestResultsAddress, "IMUP_SPEED_TEST_RESULTS_ADDRESS", fmt.Sprintf("%s/v1/realtime/speedTestResults", ImUpAPIHost))
	cfg.SpeedTestStatusUpdateAddress = util.ValueOr(speedTestStatusUpdateAddress, "IMUP_SPEED_TEST_STATUS_ADDRESS", fmt.Sprintf("%s/v1/realtime/speedTestStatusUpdate", ImUpAPIHost))

	cfg.PingAddressesExternal = strings.Split(util.ValueOr(pingAddressesExternal, "PING_ADDRESS", "1.1.1.1/32,1.0.0.1/32,8.8.8.8/32,8.8.4.4/32"), ",")

	var err error
	connDelayStr := util.ValueOr(connDelay, "CONN_DELAY", "200")
	cfg.ConnDelay, err = strconv.Atoi(connDelayStr)
	if err != nil {
		panic(err)
	}

	connIntervalStr := util.ValueOr(connInterval, "CONN_INTERVAL", "60")
	cfg.ConnInterval, err = strconv.Atoi(connIntervalStr)
	if err != nil {
		panic(err)
	}

	connRequestsStr := util.ValueOr(connRequests, "CONN_REQUESTS", "300")
	cfg.ConnRequests, err = strconv.Atoi(connRequestsStr)
	if err != nil {
		panic(err)
	}

	imupDataLengthStr := util.ValueOr(imupDataLength, "IMUP_DATA_LENGTH", "15")
	cfg.IMUPDataLength, err = strconv.Atoi(imupDataLengthStr)
	if err != nil {
		panic(err)
	}

	pingDelayStr := util.ValueOr(pingDelay, "PING_DELAY", "100")
	cfg.PingDelay, err = strconv.Atoi(pingDelayStr)
	if err != nil {
		panic(err)
	}

	pingIntervalStr := util.ValueOr(pingInterval, "PING_INTERVAL", "60")
	cfg.PingInterval, err = strconv.Atoi(pingIntervalStr)
	if err != nil {
		panic(err)
	}

	pingRequestsStr := util.ValueOr(pingRequests, "PING_REQUESTS", "600")
	cfg.PingRequests, err = strconv.Atoi(pingRequestsStr)
	if err != nil {
		panic(err)
	}

	logFilePathStr := util.ValueOr(logFile, "LOG_FILE", "")
	cfg.InsecureSpeedTest = util.BooleanValueOr(insecureSpeedTest, "INSECURE_SPEED_TEST", "false")
	cfg.FileLogger = util.BooleanValueOr(logToFile, "LOG_TO_FILE", "false")
	cfg.NoDiscoverGateway = util.BooleanValueOr(noGatewayDiscovery, "NO_GATEWAY_DISCOVERY", "false")
	cfg.SpeedTestEnabled = !util.BooleanValueOr(noSpeedTest, "NO_SPEED_TEST", "false")
	cfg.Nonvolatile = util.BooleanValueOr(nonvolatile, "NONVOLATILE", "false")
	cfg.PingEnabled = util.BooleanValueOr(pingEnabled, "PING_ENABLED", "true")
	cfg.RealtimeEnabled = util.BooleanValueOr(realtimeEnabled, "REALTIME", "true")

	cfg.logLevel = util.LevelMap(verbosity, "VERBOSITY", "info")

	var w io.Writer
	if logFilePathStr != "" {
		w = logToThisFile(logFilePathStr)
	} else if cfg.FileLogger {
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
	h := log.NewJSONHandler(w, &log.HandlerOptions{Level: verbosity})
	log.SetDefault(log.New(h))
}

func logToThisFile(file string) *os.File {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Error("cannot open file", "error", err)
	}

	log.Debug("log file at", "file", file)
	return f
}

func logToUserCache() *os.File {
	cache, err := os.UserCacheDir()
	if err != nil {
		log.Error("$HOME is likely undefined", "error", err)
	}

	targetDir := filepath.Join(cache, "imup", "logs")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		log.Error("cannot create directory in user cache", "error", err)
	}

	f, err := os.OpenFile(filepath.Join(targetDir, "imup.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Error("cannot open file", "error", err)
	}

	log.Debug("log file located at", "path", targetDir)
	return f
}

func (cfg *config) validate() error {
	if (cfg.email == "unknown" || cfg.email == "") && (cfg.apiKey == "" || cfg.hostID == "") {
		return fmt.Errorf("please supply an email address (--email) or api key and host id (--key, --host-id)!: email: %s, key: %s, host id: %s", cfg.email, cfg.apiKey, cfg.hostID)
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

// puke

func (c *config) PostConnectionData() string {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.APIPostConnectionData
}

func (c *config) PostSpeedTestData() string {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.APIPostSpeedTestData
}

func (c *config) LivenessCheckInURL() string {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.LivenessCheckInAddress
}

func (c *config) ShouldRunSpeedTestURL() string {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.ShouldRunSpeedTestAddress
}

func (c *config) SpeedTestResultsURL() string {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.SpeedTestResultsAddress
}

func (c *config) SpeedTestStatusUpdateURL() string {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.SpeedTestStatusUpdateAddress
}

func (c *config) RealtimeAuth() string {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.RealtimeAuthorized
}

func (c *config) RealtimeConfigURL() string {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.RealtimeConfig
}

func (c *config) PingAddresses() []string {
	mu.RLock()
	defer mu.RUnlock()
	return ips(cfg.PingAddressesExternal)
}

func (c *config) InternalPingAddress() string {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.PingAddressInternal
}

func (c *config) PingIntervalSeconds() int {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.PingInterval
}

func (c *config) ConnIntervalSeconds() int {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.ConnInterval
}

func (c *config) PingDelayMilli() int {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.PingDelay
}

func (c *config) ConnDelayMilli() int {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.ConnDelay
}

func (c *config) PingRequestsCount() int {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.PingRequests
}

func (c *config) ConnRequestsCount() int {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.ConnRequests
}

func (c *config) IMUPDataLen() int {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.IMUPDataLength
}
