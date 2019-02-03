package shared

import (
	"sync/atomic"
	"time"
)

var (
	// DefaultServerURL websocket url
	DefaultServerURL = "ws://test.river.im"

	// DefaultTimeout request timeout
	DefaultTimeout = 10 * time.Second

	// DefaultSendTimeout write to ws timeout
	DefaultSendTimeout = 10 * time.Second

	// MaxWorker concurrent go routins
	MaxWorker = 96

	// MaxQueueBuffer queue channel buffer size
	MaxQueueBuffer = 10000
)

var (
	counter = time.Now().UnixNano() //int64(10000)
)

func GetSeqID() int64 {
	atomic.AddInt64(&counter, 1)
	return counter
}
