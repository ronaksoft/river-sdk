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

func ExecSuccessCB(handler domain.MessageHandler, out *msg.MessageEnvelope) {
	if handler != nil {
		exec(func() {
			handler(out)
		})
	}
}

func ExecTimeoutCB(h domain.TimeoutCallback) {
	if h != nil {
		exec(func() {
			h()
		})
	}
}

func ExecUpdate(cb domain.UpdateReceivedCallback, constructor int64, proto domain.Proto) {
	b := pbytes.GetLen(proto.Size())
	n, err := proto.MarshalToSizedBuffer(b)
	if err == nil {
		exec(func() {
			cb(constructor, b[:n])
			pbytes.Put(b)
		})
	}
}

// Exec pass given function to UIExecutor buffered channel
func exec(fn func()) {
	select {
	case funcChan <- fn:
	default:
		logs.Error("Error On Pushing To UIExec")
	}

}
