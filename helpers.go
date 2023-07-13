package main

import (
	"math/rand"
	"time"

	log "golang.org/x/exp/slog"
)

// https://github.com/m-lab/ndt-server/blob/master/spec/ndt7-protocol.md#requirements-for-non-interactive-clients
func sleepTime() time.Duration {
	rand := rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
	t := rand.ExpFloat64() * 21600
	if t < 2160 {
		t = 2160
	} else if t > 54000 {
		t = 54000
	}

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
