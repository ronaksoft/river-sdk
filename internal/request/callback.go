package request

import (
	"git.ronaksoft.com/river/sdk/internal/domain"
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
	OnComplete(m *rony.MessageEnvelope)
	OnTimeout()
	OnProgress(percent int64)
	UI() bool
}

type callback struct {
	onComplete func(m *rony.MessageEnvelope)
	onTimeout  func()
	onProgress func(percent int64)
	ui         bool
}

func (c *callback) OnComplete(m *rony.MessageEnvelope) {
	if c.onComplete == nil {
		return
	}
	c.onComplete(m)
}

func (c *callback) OnTimeout() {
	if c.onTimeout == nil {
		return
	}
	c.onTimeout()
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

func NewCallback(onTimeout func(), onComplete func(envelope *rony.MessageEnvelope), onProgress func(int64), ui bool) *callback {
	return &callback{
		onComplete: onComplete,
		onTimeout:  onTimeout,
		onProgress: onProgress,
		ui:         ui,
	}
}

func EmptyCallback() *callback {
	return &callback{}
}


var (
	callbacksMtx     sync.Mutex
	requestCallbacks map[uint64]*RequestCallback
)

// RequestCallback ...
// This will be stored in memory until the request sent to server, if server responds in the acceptable time frame,
// SuccessCallback will be called otherwise TimeoutCallback will be called. If for any reason the memory was reset
// then these callbacks will be forgot and the response will be saved on disk to be fetched later.
type RequestCallback struct {
	RequestID       uint64
	SuccessCallback domain.MessageHandler
	TimeoutCallback domain.TimeoutCallback
	ResponseChannel chan *rony.MessageEnvelope
	CreatedOn       int64
	DepartureTime   int64
	Timeout         time.Duration
	IsUICallback    bool
	Constructor     int64
}

func (rcb *RequestCallback) OnSuccess(m *rony.MessageEnvelope) {
	// if rcb.SuccessCallback == nil {
	// 	return
	// }
	// if rcb.IsUICallback {
	// 	uiexec.ExecSuccessCB(rcb.SuccessCallback, m)
	// } else {
	// 	rcb.SuccessCallback(m)
	// }
}

func (rcb *RequestCallback) OnTimeout() {
	// if rcb.TimeoutCallback == nil {
	// 	return
	// }
	// if rcb.IsUICallback {
	// 	uiexec.ExecTimeoutCB(rcb.TimeoutCallback)
	// } else {
	// 	rcb.TimeoutCallback()
	// }
}

func init() {
	requestCallbacks = make(map[uint64]*RequestCallback, 100)
}

// AddRequestCallback in memory cache to save requests
func AddRequestCallback(
	reqID uint64, constructor int64, successCB domain.MessageHandler, timeOut time.Duration, timeoutCB domain.TimeoutCallback, isUICallback bool,
) *RequestCallback {
	cb := &RequestCallback{
		RequestID:       reqID,
		SuccessCallback: successCB,
		TimeoutCallback: timeoutCB,
		ResponseChannel: make(chan *rony.MessageEnvelope),
		CreatedOn:       tools.NanoTime(),
		DepartureTime:   tools.NanoTime(),
		Timeout:         timeOut,
		IsUICallback:    isUICallback,
		Constructor:     constructor,
	}
	callbacksMtx.Lock()
	requestCallbacks[reqID] = cb
	callbacksMtx.Unlock()
	return cb
}

// RemoveRequestCallback remove from in memory cache
func RemoveRequestCallback(reqID uint64) {
	callbacksMtx.Lock()
	delete(requestCallbacks, reqID)
	callbacksMtx.Unlock()
}

// GetRequestCallback fetch request
func GetRequestCallback(reqID uint64) (cb *RequestCallback) {
	callbacksMtx.Lock()
	val, ok := requestCallbacks[reqID]
	callbacksMtx.Unlock()
	if ok {
		cb = val
	}
	return
}
