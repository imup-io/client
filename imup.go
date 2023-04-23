package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/imup-io/client/realtime"
	"github.com/imup-io/client/util"

	log "golang.org/x/exp/slog"
)

type sendDataJob struct {
	IMUPAddress string
	IMUPData    any
}

type imupStatCollector interface {
	Interval() time.Duration
	Collect(context.Context, []string) []pingStats
	DetectDowntime([]pingStats) (bool, int)
}

type imupData struct {
	Downtime      int         `json:"downtime,omitempty"`
	StatusChanged bool        `json:"statusChanged"`
	Email         string      `json:"email,omitempty"`
	ID            string      `json:"hostId,omitempty"`
	Key           string      `json:"apiKey,omitempty"`
	GroupID       string      `json:"group_id,omitempty"`
	IMUPData      []pingStats `json:"data,omitempty"`
}

type authRequest struct {
	Key   string `json:"apiKey,omitempty"`
	Email string `json:"email,omitempty"`
}

type pingStats struct {
	PingAddress     string        `json:"pingAddress,omitempty"`
	Success         bool          `json:"success,omitempty"`
	PacketsRecv     int           `json:"packetsRecv,omitempty"`
	PacketsSent     int           `json:"packetsSent,omitempty"`
	PacketLoss      float64       `json:"packetLoss,omitempty"`
	MinRtt          time.Duration `json:"minRtt,omitempty"`
	MaxRtt          time.Duration `json:"maxRtt,omitempty"`
	AvgRtt          time.Duration `json:"avgRtt,omitempty"`
	StdDevRtt       time.Duration `json:"stdDevRtt,omitempty"`
	TimeStamp       int64         `json:"timestamp,omitempty"`
	ClientVersion   string        `json:"clientVersion,omitempty"`
	OS              string        `json:"operatingSystem,omitempty"`
	EndpointType    string        `json:"endpointType,omitempty"`
	SuccessInternal bool          `json:"successInternal,omitempty"`
}

type imup struct {
	APIPostConnectionData        string
	APIPostSpeedTestData         string
	LivenessCheckInAddress       string
	PingAddressesExternal        string
	PingAddressInternal          string
	RealtimeAuthorized           string
	RealtimeConfig               string
	ShouldRunSpeedTestAddress    string
	SpeedTestResultsAddress      string
	SpeedTestStatusUpdateAddress string

	OnDemandSpeedTest bool
	SpeedTestRunning  bool

	ConnDelay        int
	ConnInterval     int
	ConnRequests     int
	IMUPDataLength   int
	PingDelay        int
	PingInterval     int
	PingRequests     int
	SpeedTestRetries int

	cfg                realtime.Reloadable
	SpeedTestLock      sync.Mutex
	ChannelImupData    chan sendDataJob
	PingAddressesAvoid map[string]bool
	Errors             *ErrMap
}

func newApp() *imup {
	cfg, err := realtime.NewConfig()
	if err != nil {
		log.Error("error", err)
		os.Exit(1)
	}

	imup := &imup{
		SpeedTestRetries:   10,
		SpeedTestRunning:   false,
		PingAddressesAvoid: map[string]bool{},
		cfg:                cfg,
	}

	imup.APIPostConnectionData = util.GetEnv("IMUP_ADDRESS", "https://api.imup.io/v1/data/connectivity")
	imup.APIPostSpeedTestData = util.GetEnv("IMUP_ADDRESS_SPEEDTEST", "https://api.imup.io/v1/data/speedtest")
	imup.LivenessCheckInAddress = util.GetEnv("IMUP_LIVENESS_CHECKIN_ADDRESS", "https://api.imup.io/v1/realtime/livenesscheckin")
	imup.ShouldRunSpeedTestAddress = util.GetEnv("IMUP_SHOULD_RUN_SPEEDTEST_ADDRESS", "https://api.imup.io/v1/realtime/shouldClientRunSpeedTest")
	imup.SpeedTestResultsAddress = util.GetEnv("IMUP_SPEED_TEST_RESULTS_ADDRESS", "https://api.imup.io/v1/realtime/speedTestResults")
	imup.SpeedTestStatusUpdateAddress = util.GetEnv("IMUP_SPEED_TEST_STATUS_ADDRESS", "https://api.imup.io/v1/realtime/speedTestStatusUpdate")
	imup.RealtimeAuthorized = util.GetEnv("IMUP_REALTIME_AUTHORIZED", "https://api.imup.io/v1/auth/realtimeAuthorized")
	imup.RealtimeConfig = util.GetEnv("IMUP_REALTIME_CONFIG", "https://api.imup.io/v1/realtime/config")
	// 1.1.1.1 and 1.0.0.1 are CloudFlare DNS servers
	imup.PingAddressesExternal = util.GetEnv("PING_ADDRESS", "1.1.1.1,1.0.0.1,8.8.8.8,8.8.4.4")
	imup.PingAddressInternal = util.GetEnv("PING_ADDRESS_INTERNAL", imup.cfg.DiscoverGateway())

	// run a ping test once every 60s
	pingIntervalStr := util.GetEnv("PING_INTERVAL", "60")
	imup.PingInterval, _ = strconv.Atoi(pingIntervalStr)

	// run a conn test once every 60s
	connIntervalStr := util.GetEnv("CONN_INTERVAL", "60")
	imup.ConnInterval, _ = strconv.Atoi(connIntervalStr)

	// wait 100ms between each ping
	pingDelayStr := util.GetEnv("PING_DELAY", "100")
	imup.PingDelay, _ = strconv.Atoi(pingDelayStr)

	// wait 200ms between each net conn
	connDelayStr := util.GetEnv("CONN_DELAY", "200")
	imup.ConnDelay, _ = strconv.Atoi(connDelayStr)

	// send 600 requests each test
	pingRequestsStr := util.GetEnv("PING_REQUESTS", "600")
	imup.PingRequests, _ = strconv.Atoi(pingRequestsStr)

	// send 300 requests each test
	connRequestsStr := util.GetEnv("CONN_REQUESTS", "300")
	imup.ConnRequests, _ = strconv.Atoi(connRequestsStr)

	// after we've collected 15 ping.Statistics() send the data to the api
	dataLengthStr := util.GetEnv("IMUP_DATA_LENGTH", "15")
	imup.IMUPDataLength, _ = strconv.Atoi(dataLengthStr)

	// make a channel with a capacity of 300.
	imup.ChannelImupData = make(chan sendDataJob, 300)

	// on startup get a clients public ip address
	imup.cfg.RefreshPublicIP()

	return imup
}

func sendImupData(ctx context.Context, job sendDataJob) {
	b, err := json.Marshal(job.IMUPData)
	if err != nil {
		log.Error("error", err)
		return
	}

	req, err := retryablehttp.NewRequest("POST", job.IMUPAddress, bytes.NewBuffer(b))
	if err != nil {
		log.Error("error", err)
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	client := retryablehttp.NewClient()
	client.Backoff = util.ExactJitterBackoff
	// 50000 should be 15-30 days with this config
	client.RetryMax = 50000
	client.RetryWaitMin = time.Duration(30) * time.Second
	client.RetryWaitMax = time.Duration(60) * time.Second
	client.Logger = log.New(log.Default().Handler())

	if _, err := client.Do(req); err != nil {
		if err == context.Canceled {
			// shutdown in progress
			toUserCache(job)
		}

		log.Error("error", err)
	}
}

func sendDataWorker(ctx context.Context, c chan sendDataJob) {
	for {
		select {
		case <-ctx.Done():
			if len(c) > 0 {
				log.Info("shutdown detected, persisting queued data")
				drain(c)
			}
			return
		case job := <-c:
			sendImupData(ctx, job)
		}
	}
}

// pingAddress chooses a semi random ping address from a list of ips
// it respects a dynamic ignore list, but if there are not enough ping addresses
// to test against, recreates the list from the base configuration
func pingAddress(addresses []string, avoidAddrs map[string]bool) string {
	pingAddrCount := len(addresses) - len(avoidAddrs)
	if pingAddrCount < 1 {
		// clear avoid list
		avoidAddrs = map[string]bool{}
		pingAddrCount = len(addresses)
	}

	allowedPingAddrs := []string{}
	for _, v := range addresses {
		if _, ok := avoidAddrs[v]; !ok {
			allowedPingAddrs = append(allowedPingAddrs, v)
		}
	}

	return allowedPingAddrs[rand.Intn(pingAddrCount)]
}

func (i *imup) reloadConfig(data []byte) {
	if cfg, err := realtime.Reload(data); err != nil {
		log.Info("cannot reload config", "error", err)
	} else {
		i.cfg = cfg
	}
}

func (i *imup) authorized(ctx context.Context, b *bytes.Buffer, addr string) error {
	req, err := retryablehttp.NewRequest("POST", addr, b)
	req = req.WithContext(ctx)
	if err != nil {
		return fmt.Errorf("NewRequest: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := retryablehttp.NewClient()
	client.Backoff = exactJitterBackoff

	client.RetryMax = 50_000
	client.RetryWaitMin = time.Duration(30) * time.Second
	client.RetryWaitMax = time.Duration(60) * time.Second
	client.Logger = log.New(log.Default().Handler())

	if resp, err := client.Do(req); err != nil {
		if err == context.Canceled {
			return nil
		}

		return fmt.Errorf("addr: %s, client.Do: %s", addr, err)
	} else {
		if resp.StatusCode == http.StatusOK {
			i.cfg.EnableRealtime()
		} else {
			i.cfg.DisableRealtime()
		}
	}

	return nil
}
