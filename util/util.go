package util

import (
	"math/rand"
	nethttp "net/http"
	"os"
	"strconv"
	"strings"
	"time"

	log "golang.org/x/exp/slog"
)

// ValueOr returns a de-referenced string pointer, an environment variable, or a fallback
func ValueOr(ptr *string, varName, defaultVal string) string {
	if ptr != nil && *ptr != "" {
		return *ptr
	}

	return GetEnv(varName, defaultVal)
}

// BooleanValueOr returns a de-referenced boolean pointer, an environment variable, or a fallback
func BooleanValueOr(ptr *bool, varName, defaultVal string) bool {
	valStr := GetEnv(varName, defaultVal)
	val, err := strconv.ParseBool(valStr)
	if err != nil {
		return false
	}

	// edge case if the pointer is unset
	if ptr == nil {
		return val
	}

	// pointer takes precedence if both are set
	if *ptr != val {
		return *ptr
	}

	return val
}

// GetEnv looks up an environment with a fallback default
func GetEnv(varName, defaultVal string) string {
	if value, isPresent := os.LookupEnv(varName); isPresent {
		return value
	}

	return defaultVal
}

var levelMap = map[string]log.Level{
	"debug": log.LevelDebug,
	"info":  log.LevelInfo,
	"warn":  log.LevelWarn,
	"error": log.LevelError,
}

// LevelMap returns a concrete log level from a string pointer, an environment variable, or a fallback
func LevelMap(ptr *string, varName, defaultVal string) log.Level {

	if ptr != nil && *ptr != "" {
		if val, ok := levelMap[strings.ToLower(*ptr)]; ok {
			return val
		}
	}

	return levelMap[strings.ToLower(GetEnv(varName, defaultVal))]
}

// IPMonitored considers configured allowed and blocked ip addresses and inspects a clients
// public ip address to determine if it should be used for speed and connectivity testing
func IPMonitored(publicIP string, allowed, blocked []string) bool {
	return ipAllowed(publicIP, allowed) && !ipBlocked(publicIP, blocked)
}

// iterate over list of allowed ips and ensure the public ip is a match
func ipAllowed(publicIP string, ips []string) bool {
	allowed := true
	for _, v := range ips {
		if v == "" {
			continue
		}

		if publicIP == v {
			allowed = true
			return true
		}

		allowed = false
	}

	return allowed
}

// iterate over list of blocked ips and ensure the public ip is not a match
func ipBlocked(publicIP string, ips []string) bool {
	blocked := false

	for _, v := range ips {
		if v == "" {
			continue
		}

		if publicIP == v {
			blocked = true
			break
		}

		blocked = false
	}

	return blocked
}

// exactJitterBackoff provides a callback for Client.Backoff which will
// perform exact backoff (not based on the attempt number) and with jitter to
// prevent a thundering herd.
//
// min and max here are *not* absolute values. The number will be chosen at
// random from between them, thus they are bounding the jitter.
//
// For instance:
// * To get strictly backoff of one second, set both to
// one second (1s, 1s, 1s, 1s, ...)
// * To get a small amount of jitter centered around one second, set to
// around one second, such as a min of 800ms and max of 1200ms
// (892ms, 1102ms, 945ms, 1012ms, ...)
// * To get extreme jitter, set to a very wide spread, such as a min of 100ms
// and a max of 20s (15382ms, 292ms, 11321ms, 15234ms, ...)
func ExactJitterBackoff(min, max time.Duration, attemptNum int, resp *nethttp.Response) time.Duration {
	if max <= min {
		// Unclear what to do here, or they are the same, so return min
		return min
	}

	// Seed rand; doing this every time is fine
	rand := rand.New(rand.NewSource(int64(time.Now().Nanosecond())))

	// Pick a random number that lies somewhere between the min and max and
	// multiply by the attemptNum. attemptNum starts at zero so we always
	// increment here. We first get a random percentage, then apply that to the
	// difference between min and max, and add to min.
	jitter := rand.Float64() * float64(max-min)
	jitterMin := int64(jitter) + int64(min)
	return time.Duration(jitterMin)
}
