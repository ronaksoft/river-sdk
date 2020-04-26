package uiexec

import (
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
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
				logs.Error("Too Long UIExec", zap.Duration("D", d))
			}
		}
	}()

}

// Exec pass given function to UIExecutor buffered channel
func Exec(fn func()) {
	select {
	case funcChan <- fn:
	default:
		logs.Error("Error On Pushing To UIExec")
	}

}
