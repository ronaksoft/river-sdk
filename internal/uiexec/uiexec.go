package uiexec

import (
	"context"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"github.com/gobwas/pool/pbytes"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/registry"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"time"
)

const (
	maxDelay = time.Millisecond * 500
)

var (
	updateCB     domain.UpdateReceivedCallback
	dataSyncedCB domain.DataSyncedCallback
	funcChan     = make(chan execItem, 100)
)

type execItem struct {
	fn          func()
	insertTime  time.Time
	constructor int64
	kind        string
}

func init() {
	updateCB = func(constructor int64, msg []byte) {}
	dataSyncedCB = func(dialogs, contacts, gifs bool) {}
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
					zap.String("C", registry.ConstructorName(it.constructor)),
					zap.String("Kind", it.kind),
				)
			}
			cf() // Cancel func
			endTime := time.Now()
			if d := endTime.Sub(it.insertTime); d > maxDelay {
				logs.Error("Too Long UIExec",
					zap.String("C", registry.ConstructorName(it.constructor)),
					zap.String("Kind", it.kind),
					zap.Duration("ExecT", endTime.Sub(startTime)),
					zap.Duration("WaitT", endTime.Sub(it.insertTime)),
				)
			}
		}
	}()
}

func Init(updateReceived domain.UpdateReceivedCallback, dataSynced domain.DataSyncedCallback) {
	updateCB = updateReceived
	dataSyncedCB = dataSynced
}

func ExecSuccessCB(handler domain.MessageHandler, out *rony.MessageEnvelope) {
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

func ExecUpdate(constructor int64, m proto.Message) {
	mo := proto.MarshalOptions{UseCachedSize: true}
	b := pbytes.GetCap(mo.Size(m))
	b, err := mo.MarshalAppend(b, m)
	if err == nil {
		exec("update", constructor, func() {
			updateCB(constructor, b)
			pbytes.Put(b)
		})
	}
}

func ExecDataSynced(dialogs, contacts, gifs bool) {
	dataSyncedCB(dialogs, contacts, gifs)
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
