package uiexec

import (
	"context"
	"git.ronaksoft.com/river/msg/msg"
	"git.ronaksoft.com/ronak/riversdk/internal/logs"
	"git.ronaksoft.com/ronak/riversdk/pkg/domain"
	"github.com/gobwas/pool/pbytes"
	"go.uber.org/zap"
	"time"
)

const (
	maxDelay = time.Millisecond * 100
)

var (
	funcChan = make(chan execItem, 100)
)

type execItem struct {
	fn          func()
	insertTime  time.Time
	constructor int64
	kind        string
}

func init() {
	go func() {
		for it := range funcChan {
			startTime := time.Now()
			ctx, cf := context.WithTimeout(context.Background(), time.Second)
			doneChan := make(chan struct{})
			go func() {
				it.fn()
				doneChan <- struct{}{}
			}()
			select {
			case <-doneChan:
			case <-ctx.Done():
				logs.Error("We timeout waiting for UI-Exec to return",
					zap.String("C", msg.ConstructorNames[it.constructor]),
					zap.String("Kind", it.kind),
				)
			}
			cf() // Cancel func
			endTime := time.Now()
			if d := endTime.Sub(it.insertTime); d > maxDelay {
				logs.Error("Too Long UIExec",
					zap.String("C", msg.ConstructorNames[it.constructor]),
					zap.String("Kind", it.kind),
					zap.Duration("ExecT", endTime.Sub(startTime)),
					zap.Duration("WaitT", endTime.Sub(it.insertTime)),
				)
			}
		}
	}()

}

func ExecSuccessCB(handler domain.MessageHandler, out *msg.MessageEnvelope) {
	if handler != nil {
		exec("successCB", out.Constructor, func() {
			handler(out)
		})
	}
}

func ExecTimeoutCB(h domain.TimeoutCallback) {
	if h != nil {
		logs.Info("ExecTimeout")
		exec("timeoutCB", 0, func() {
			h()
		})
	}
}

func ExecUpdate(cb domain.UpdateReceivedCallback, constructor int64, proto domain.Proto) {
	b := pbytes.GetLen(proto.Size())
	n, err := proto.MarshalToSizedBuffer(b)
	if err == nil {
		exec("update", constructor, func() {
			cb(constructor, b[:n])
			pbytes.Put(b)
		})
	}
}

// Exec pass given function to UIExecutor buffered channel
func exec(kind string, constructor int64, fn func()) {
	select {
	case funcChan <- execItem{
		kind:        kind,
		fn:          fn,
		insertTime:  time.Now(),
		constructor: constructor,
	}:
	default:
		logs.Error("Error On Pushing To UIExec")
	}

}
