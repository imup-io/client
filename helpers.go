package main

import (
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"
)

// https://github.com/m-lab/ndt-server/blob/master/spec/ndt7-protocol.md#requirements-for-non-interactive-clients
func sleepTime() time.Duration {
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
	log.Debugf("channel depth at shutdown: %v", toDrain)
	i := 0
	for job := range c {
		i++
		log.Debugf("draining: %v of %v jobs", i, toDrain)
		toUserCache(job)

		if i >= toDrain {
			break
		}
	}
}
