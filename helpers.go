package main

import (
	"time"

	// xrand "golang.org/x/exp/rand"
	log "golang.org/x/exp/slog"
	"gonum.org/v1/gonum/stat/distuv"
)

// timePeriodMinutes is the desired interval between tests
// the default (fixed) configuration is one tests every six hours
const timePeriodMinutes = 6 * 60

// speedTestInterval takes advantage of a poisson distribution
// to generate pseudo random speed tests
// this distribution helps guarantee a consistent number of speed tests
// with a smaller chance of frequent speed tests consuming large amounts of data
// or saturating a network
func speedTestInterval() time.Duration {
	t := distuv.Poisson{
		Lambda: timePeriodMinutes,
	}.Rand()

	return time.Duration(t * float64(time.Minute))
}

func drain(c chan sendDataJob) {
	toDrain := len(c)
	log.Debug("unprocessed jobs at shutdown", "channel depth", toDrain)
	i := 0
	for job := range c {
		i++
		log.Debug("draining unsent data", "jobs", i, "left to drain", toDrain)
		toUserCache(job)

		if i >= toDrain {
			break
		}
	}
}
