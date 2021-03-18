package node_checker

import (
	"time"
	log "github.com/sirupsen/logrus"
)

type jobTicker struct {
	timer *time.Timer
}

func (t *jobTicker) updateTimer(diff time.Duration) {
	log.Infof("Scheduling next update at around: %v", time.Now().Add(diff))
	if t.timer == nil {
		t.timer = time.NewTimer(diff)
	} else {
		t.timer.Reset(diff)
	}
}
