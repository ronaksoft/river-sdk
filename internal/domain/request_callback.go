package domain

import (
	"github.com/ronaksoft/rony"
	"sync"
	"time"
)

var (
	mxCB             sync.Mutex
	requestCallbacks map[uint64]*RequestCallback
)

// RequestCallback ...
// This will be stored in memory until the request sent to server, if server responds in the acceptable time frame,
// SuccessCallback will be called otherwise TimeoutCallback will be called. If for any reason the memory was reset
// then these callbacks will be forgot and the response will be saved on disk to be fetched later.
type RequestCallback struct {
	RequestID       uint64
	SuccessCallback MessageHandler
	TimeoutCallback TimeoutCallback
	ResponseChannel chan *rony.MessageEnvelope
	CreatedOn       time.Time
	Timeout         time.Duration
	IsUICallback    bool
	DepartureTime   time.Time
	Constructor     int64
}

func init() {
	requestCallbacks = make(map[uint64]*RequestCallback, 0)
}

// AddRequestCallback in memory cache to save requests
func AddRequestCallback(reqID uint64, constructor int64, successCB MessageHandler, timeOut time.Duration, timeoutCB TimeoutCallback, isUICallback bool) *RequestCallback {
	cb := &RequestCallback{
		RequestID:       reqID,
		SuccessCallback: successCB,
		TimeoutCallback: timeoutCB,
		ResponseChannel: make(chan *rony.MessageEnvelope),
		CreatedOn:       time.Now(),
		DepartureTime:   time.Now(),
		Timeout:         timeOut,
		IsUICallback:    isUICallback,
		Constructor:     constructor,
	}
	mxCB.Lock()
	requestCallbacks[reqID] = cb
	mxCB.Unlock()
	return cb
}

// RemoveRequestCallback remove from in memory cache
func RemoveRequestCallback(reqID uint64) {
	mxCB.Lock()
	delete(requestCallbacks, reqID)
	mxCB.Unlock()
}

// GetRequestCallback fetch request
func GetRequestCallback(reqID uint64) (cb *RequestCallback) {
	mxCB.Lock()
	val, ok := requestCallbacks[reqID]
	mxCB.Unlock()
	if ok {
		cb = val
	}
	return
}
