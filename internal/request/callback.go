package request

import (
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/uiexec"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/tools"
	"sync"
	"time"
)

/*
   Creation Time: 2021 - May - 25
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

type Callback interface {
	RequestID() uint64
	Constructor() int64
	OnComplete(m *rony.MessageEnvelope)
	OnTimeout()
	OnProgress(percent int64)
	UI() bool
}

type callback struct {
	reqID       uint64
	constructor int64
	onComplete  func(m *rony.MessageEnvelope)
	onTimeout   func()
	onProgress  func(percent int64)
	ui          bool

	ResponseChannel chan *rony.MessageEnvelope
	CreatedOn       int64
	DepartureTime   int64
	Timeout         time.Duration
}

func (c *callback) RequestID() uint64 {
	return c.reqID
}

func (c *callback) Constructor() int64 {
	return c.constructor
}

func (c *callback) OnComplete(m *rony.MessageEnvelope) {
	if c.onComplete == nil {
		return
	}
	if c.ui {
		uiexec.ExecSuccessCB(c.onComplete, m)
	} else {
		c.onComplete(m)
	}
}

func (c *callback) OnTimeout() {
	if c.onTimeout == nil {
		return
	}
	if c.ui {
		uiexec.ExecTimeoutCB(c.onTimeout)
	} else {
		c.onTimeout()
	}

}

func (c *callback) OnProgress(percent int64) {
	if c.onProgress == nil {
		return
	}
	c.onProgress(percent)
}

func (c *callback) UI() bool {
	return c.ui
}

func NewCallback(
	reqID uint64, constructor int64,
	onTimeout domain.TimeoutCallback, onComplete domain.MessageHandler, onProgress func(int64),
	ui bool,
) *callback {
	return &callback{
		reqID:           reqID,
		constructor:     constructor,
		onComplete:      onComplete,
		onTimeout:       onTimeout,
		onProgress:      onProgress,
		ui:              ui,
		CreatedOn:       tools.NanoTime(),
		DepartureTime:   tools.NanoTime(),
		ResponseChannel: make(chan *rony.MessageEnvelope),
	}
}

func EmptyCallback() *callback {
	return &callback{}
}

var (
	callbacksMtx     sync.Mutex
	requestCallbacks map[uint64]*callback
)

func init() {
	requestCallbacks = make(map[uint64]*callback, 100)
}

func RegisterCallback(
	reqID uint64, constructor int64, completeCB domain.MessageHandler, timeOut time.Duration, timeoutCB domain.TimeoutCallback, isUICallback bool,
) *callback {
	cb := &callback{
		onComplete:      completeCB,
		onTimeout:       timeoutCB,
		ui:              isUICallback,
		reqID:           reqID,
		constructor:     constructor,
		ResponseChannel: make(chan *rony.MessageEnvelope),
		CreatedOn:       tools.NanoTime(),
		DepartureTime:   tools.NanoTime(),
		Timeout:         timeOut,
	}
	callbacksMtx.Lock()
	requestCallbacks[reqID] = cb
	callbacksMtx.Unlock()
	return cb
}

func UnregisterCallback(reqID uint64) {
	callbacksMtx.Lock()
	delete(requestCallbacks, reqID)
	callbacksMtx.Unlock()
}

func GetCallback(reqID uint64) (cb *callback) {
	callbacksMtx.Lock()
	cb = requestCallbacks[reqID]
	callbacksMtx.Unlock()
	return
}
