package main

import (
	"math/rand"
	nethttp "net/http"
	"time"
)

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
func exactJitterBackoff(min, max time.Duration, attemptNum int, resp *nethttp.Response) time.Duration {
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
