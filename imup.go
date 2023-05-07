package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/imup-io/client/config"

	log "golang.org/x/exp/slog"
)

type sendDataJob struct {
	IMUPAddress string
	IMUPData    any
}

type imupData struct {
	Downtime      int    `json:"downtime,omitempty"`
	StatusChanged bool   `json:"statusChanged"`
	Email         string `json:"email,omitempty"`
	ID            string `json:"hostId,omitempty"`
	Key           string `json:"apiKey,omitempty"`
	GroupID       string `json:"group_id,omitempty"`
	IMUPData      any    `json:"data,omitempty"`
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
	cfg                config.Reloadable
	SpeedTestLock      sync.Mutex
	ChannelImupData    chan sendDataJob
	PingAddressesAvoid map[string]bool
	Errors             *ErrMap
}

func newApp() *imup {
	cfg, err := config.New()
	if err != nil {
		log.Error("error", err)
		os.Exit(1)
	}

	imup := &imup{
		PingAddressesAvoid: map[string]bool{},
		cfg:                cfg,
	}

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
	client.Backoff = exactJitterBackoff
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

func (i *imup) reloadConfig(data []byte) {
	if cfg, err := config.Reload(data); err != nil {
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
