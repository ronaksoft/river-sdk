package shared

import (
	"sync/atomic"
	"time"
)

const (
	// DefaultServerURL websocket url
	DefaultServerURL = "ws://test.river.im"

	// DefaultTimeout request timeout
	DefaultTimeout = 15 * time.Second

	// DefaultSendTimeout write to ws timeout
	DefaultSendTimeout = 15 * time.Second

	// MaxWorker concurrent go routins
	MaxWorker = 96

	// MaxQueueBuffer queue channel buffer size
	MaxQueueBuffer = 2000
)

var (
	counter = int64(1)
)

func GetSeqID() int64 {
	atomic.AddInt64(&counter, 1)
	return counter
}
