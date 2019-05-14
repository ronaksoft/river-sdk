package shared

import (
	"fmt"
	"sync/atomic"
	"time"
)

var (
	// DefaultServerURL websocket url
	DefaultServerURL = "ws://test.river.im"
	// DefaultFileServerURL file server cluster
	DefaultFileServerURL = "http://test.river.im/file"

	// DefaultTimeout request timeout
	DefaultTimeout = 10 * time.Second

	// DefaultSendTimeout write to ws timeout
	DefaultSendTimeout = 10 * time.Second

	// MaxWorker concurrent go routines
	MaxWorker = 100

	// MaxQueueBuffer queue channel buffer size
	MaxQueueBuffer = 10000

	SeqIDCounter = time.Now().UnixNano()
)

var (
	// ServerIP = net.ParseIP("185.126.202.199")
	// ClientIP = net.ParseIP("192.168.17.99")
	ServerPort = uint16(80)
)

func GetSeqID() int64 {
	return atomic.AddInt64(&SeqIDCounter, 1)
}

func GetPhone(number int64) string {
	return fmt.Sprintf("237400%07d", number)
}
