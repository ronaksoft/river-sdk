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

	SeqIDCounter = time.Now().UnixNano()
)

func GetSeqID() int64 {
	return atomic.AddInt64(&SeqIDCounter, 1)
}
