package main

import (
	"time"

	xrand "golang.org/x/exp/rand"
	log "golang.org/x/exp/slog"
	"gonum.org/v1/gonum/stat/distuv"
)

// timePeriodSeconds at 21600 is six hours
const timePeriodSeconds = 21600

// speedTestInterval takes advantage of a poisson distribution
// to generate pseudo random speed tests
// this distribution helps guarantee a consistent number of speed tests
// with a smaller chance of frequent speed tests consuming large amounts of data
// or saturating a network
func speedTestInterval() time.Duration {
	poisson := distuv.Poisson{
		Lambda: timePeriodSeconds,
		Src:    xrand.NewSource(uint64(time.Now().UnixNano())),
	}

	t := poisson.Rand()

	return time.Duration(t * float64(time.Second))
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
