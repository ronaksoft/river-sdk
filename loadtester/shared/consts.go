package shared

import (
	"sync/atomic"
	"time"
)

const (
	// DefaultServerURL websocket url
	DefaultServerURL = "ws://new.river.im"

	// DefaultTimeout request timeout
	DefaultTimeout = 10 * time.Second

	// DefaultSendTimeout write to ws timeout
	DefaultSendTimeout = 3 * time.Second

	// MaxWorker concurrent go routins
	MaxWorker = 96

	// MaxQueueBuffer queue channel buffer size
	MaxQueueBuffer = 512
)

var (
	counter = time.Now().UnixNano()
)

func GetSeqID() int64 {
	atomic.AddInt64(&counter, 1)
	return counter
}
