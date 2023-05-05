package connectivity

import (
	"context"
	"math/rand"
	"time"
)

type StatCollector interface {
	Interval() time.Duration
	Collect(context.Context, []string) []Statistics
	DetectDowntime([]Statistics) (bool, int)
}

type Statistics struct {
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
