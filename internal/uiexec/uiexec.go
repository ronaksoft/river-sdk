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
	callbackChan = make(chan execItem, 128)
	rateLimit    = make(chan struct{}, 5)
	logger       *logs.Logger
)

type kind int64

const (
	update kind = iota
	completeCB
	timeoutCB
)

func (k kind) String() string {
	switch k {
	case update:
		return "Update"
	case completeCB:
		return "CompleteCB"
	case timeoutCB:
		return "TimeoutCB"
	}
	panic("invalid ui-exec kind")
}

type execItem struct {
	insertTime  int64
	constructor int64
	kind        kind
	fn          func()
}

func init() {
	logger = logs.With("UIExec")
	updateCB = func(constructor int64, msg []byte) {}
	dataSyncedCB = func(dialogs, contacts, gifs bool) {}
	go executor()
}

func executor() {
	for it := range callbackChan {
		rateLimit <- struct{}{}
		go func(it execItem) {
			defer func() {
				<-rateLimit
			}()
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
					zap.String("Kind", it.kind.String()),
				)
			}
			cf() // Cancel func
			endTime := tools.NanoTime()
			if d := time.Duration(endTime - it.insertTime); d > maxDelay {
				logger.Error("Too Long UIExec",
					zap.String("C", registry.ConstructorName(it.constructor)),
					zap.String("Kind", it.kind.String()),
					zap.Duration("ExecT", time.Duration(endTime-startTime)),
					zap.Duration("WaitT", time.Duration(endTime-it.insertTime)),
				)
			}
		}(it)

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
	exec(completeCB, out.Constructor, func() {
		handler(out)
	})
}

func ExecTimeoutCB(h domain.TimeoutCallback) {
	if h == nil {
		return
	}
	exec(timeoutCB, 0, func() {
		h()
	})
}

func ExecUpdate(constructor int64, m proto.Message) {
	buf := pools.Buffer.FromProto(m)
	exec(update, constructor, func() {
		updateCB(constructor, *buf.Bytes())
		pools.Buffer.Put(buf)
	})

}

func ExecDataSynced(dialogs, contacts, gifs bool) {
	dataSyncedCB(dialogs, contacts, gifs)
}

// Exec pass given function to UIExecutor buffered channel
func exec(kind kind, constructor int64, fn func()) {
	select {
	case callbackChan <- execItem{
		kind:        kind,
		fn:          fn,
		insertTime:  tools.NanoTime(),
		constructor: constructor,
	}:
	default:
		logger.Error("func channel is full we discard item",
			zap.String("Kind", kind.String()),
			zap.String("C", registry.ConstructorName(constructor)),
		)
	}

}
