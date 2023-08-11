package main

import (
	"time"

	"golang.org/x/exp/constraints"
	log "golang.org/x/exp/slog"
	"gonum.org/v1/gonum/stat/distuv"
)

// speedTestInterval takes advantage of a poisson distribution
// to generate pseudo random speed tests
// this distribution helps guarantee a consistent number of speed tests
// with a smaller chance of frequent speed tests consuming large amounts of data
// or saturating a network
//
// numTests is the desired number of speed tests to run, within a 24 hour period
func speedTestInterval(numTests int) time.Duration {
	if numTests < 1 {
		numTests = 1
	}

	if numTests > 32 {
		numTests = 32
	}

	t := distuv.Poisson{
		Lambda: float64(1440 / numTests),
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

func max[T constraints.Ordered](s []T) T {
	if len(s) == 0 {
		var zero T
		return zero
	}
	m := s[0]
	for _, v := range s {
		if m < v {
			m = v
		}
	}
	return m
}

func min[T constraints.Ordered](s []T) T {
	if len(s) == 0 {
		var zero T
		return zero
	}
	m := s[0]
	for _, v := range s {
		if m > v {
			m = v
		}
	}
	return m
}
