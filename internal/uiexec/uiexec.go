package uiexec

import (
	"context"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/pools"
	"github.com/ronaksoft/rony/registry"
	"github.com/ronaksoft/rony/tools"
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
	funcChan     = make(chan execItem, 128)
	logger       *logs.Logger
)

type execItem struct {
	fn          func()
	insertTime  int64
	constructor int64
	kind        string
}

func init() {
	logger = logs.With("UIExec")
	updateCB = func(constructor int64, msg []byte) {}
	dataSyncedCB = func(dialogs, contacts, gifs bool) {}
	go executor()
}

func executor() {
	for it := range funcChan {
		startTime := tools.NanoTime()
		ctx, cf := context.WithTimeout(context.Background(), time.Second)
		doneChan := make(chan struct{})
		go func() {
			it.fn()
			doneChan <- struct{}{}
		}()
		select {
		case <-doneChan:
		case <-ctx.Done():
			logger.Error("timeout waiting for UI-Exec to return",
				zap.String("C", registry.ConstructorName(it.constructor)),
				zap.String("Kind", it.kind),
			)
		}
		cf() // Cancel func
		endTime := tools.NanoTime()
		if d := time.Duration(endTime - it.insertTime); d > maxDelay {
			logger.Error("Too Long UIExec",
				zap.String("C", registry.ConstructorName(it.constructor)),
				zap.String("Kind", it.kind),
				zap.Duration("ExecT", time.Duration(endTime-startTime)),
				zap.Duration("WaitT", time.Duration(endTime-it.insertTime)),
			)
		}
	}
}

func Init(updateReceived domain.UpdateReceivedCallback, dataSynced domain.DataSyncedCallback) {
	updateCB = updateReceived
	dataSyncedCB = dataSynced
}

func ExecCompleteCB(handler domain.MessageHandler, out *rony.MessageEnvelope) {
	if handler == nil {
		return
	}
	exec("completeCB", out.Constructor, func() {
		handler(out)
	})
}

func ExecUpdate(constructor int64, m proto.Message) {
	buf := pools.Buffer.FromProto(m)
	exec("update", constructor, func() {
		updateCB(constructor, *buf.Bytes())
		pools.Buffer.Put(buf)
	})
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
		insertTime:  tools.NanoTime(),
		constructor: constructor,
	}:
	default:
		logger.Error("Error On Pushing To UIExec",
			zap.String("Kind", kind),
			zap.String("C", registry.ConstructorName(constructor)),
		)
	}

}
