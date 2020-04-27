package uiexec

import (
	msg "git.ronaksoftware.com/river/msg/chat"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
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
}

func init() {
	go func() {
		for it := range funcChan {
			startTime := time.Now()
			it.fn()
			endTime := time.Now()
			if d := endTime.Sub(it.insertTime); d > maxDelay {
				logs.Error("Too Long UIExec",
					zap.String("C", msg.ConstructorNames[it.constructor]),
					zap.Duration("ExecT", endTime.Sub(startTime)),
					zap.Duration("WaitT", endTime.Sub(it.insertTime)),
				)
			}
		}
	}()

}

func ExecSuccessCB(handler domain.MessageHandler, out *msg.MessageEnvelope) {
	if handler != nil {
		exec(0, func() {
			handler(out)
		})
	}
}

func ExecTimeoutCB(h domain.TimeoutCallback) {
	if h != nil {
		exec(0, func() {
			h()
		})
	}
}

func ExecUpdate(cb domain.UpdateReceivedCallback, constructor int64, proto domain.Proto) {
	b := pbytes.GetLen(proto.Size())
	n, err := proto.MarshalToSizedBuffer(b)
	if err == nil {
		exec(constructor, func() {
			cb(constructor, b[:n])
			pbytes.Put(b)
		})
	}
}

// Exec pass given function to UIExecutor buffered channel
func exec(constructor int64, fn func()) {
	select {
	case funcChan <- execItem{
		fn:          fn,
		insertTime:  time.Now(),
		constructor: constructor,
	}:
	default:
		logs.Error("Error On Pushing To UIExec")
	}

}
