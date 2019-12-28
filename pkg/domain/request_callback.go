package domain

import (
	"sync"
	"time"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/chat"
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
	ResponseChannel chan *msg.MessageEnvelope
	CreatedOn       time.Time
	Timeout         time.Duration
	IsUICallback    bool
	FlightTime      time.Time
	Constructor     int64
}

func init() {
	requestCallbacks = make(map[uint64]*RequestCallback, 0)
}

// AddRequestCallback in memory cache to save requests
func AddRequestCallback(reqID uint64, constructor int64, success MessageHandler, timeOut time.Duration, timeout TimeoutCallback, isUICallback bool) *RequestCallback {
	cb := &RequestCallback{
		RequestID:       reqID,
		SuccessCallback: success,
		TimeoutCallback: timeout,
		ResponseChannel: make(chan *msg.MessageEnvelope),
		CreatedOn:       time.Now(),
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
