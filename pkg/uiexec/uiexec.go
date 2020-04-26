package uiexec

import (
	"github.com/prometheus/common/log"
	"go.uber.org/zap"
	"time"
)

const (
	maxDelay = time.Millisecond * 100
)

var (
	funcChan = make(chan func(), 100)
)

func init() {
	go func() {
		for fn := range funcChan {
			startTime := time.Now()
			fn()
			if d := time.Now().Sub(startTime); d > maxDelay {
				log.Error("Too Long UIExec", zap.Duration("D", d))
			}
		}
	}()

}

// Exec pass given function to UIExecutor buffered channel
func Exec(fn func()) {
	select {
	case funcChan <- fn:
	default:
		log.Error("Error On Pushing To UIExec", len(funcChan))
	}

}
