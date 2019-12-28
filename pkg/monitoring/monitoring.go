package mon

import (
	msg "git.ronaksoftware.com/ronak/riversdk/msg/chat"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"go.uber.org/zap"
	"sync"
	"sync/atomic"
	"time"
)

/*
   Creation Time: 2019 - Jul - 16
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

const (
	serverLongThreshold   = 2 * time.Second
	queueLongThreshold    = 100 * time.Millisecond
	functionLongThreshold = 50 * time.Millisecond
)

type stats struct {
	mtx *sync.RWMutex

	StartTime             time.Time
	AvgServerResponseTime time.Duration
	MaxServerResponseTime time.Duration
	MinServerResponseTime time.Duration
	TotalServerRequests   int32
	AvgQueueTime          time.Duration
	MaxQueueTime          time.Duration
	MinQueueTime          time.Duration
	TotalQueueItems       int32

	totalBytes         int
	dataTransferPeriod time.Duration
	lastDataTransfer   time.Time
}

var Stats stats

func init() {
	Stats.StartTime = time.Now()
	Stats.mtx = &sync.RWMutex{}
}

func DataTransfer(totalUploadBytes, totalDownloadBytes int, d time.Duration) {
	Stats.mtx.Lock()
	if time.Now().Sub(Stats.lastDataTransfer) > time.Second*30 {
		Stats.totalBytes = totalDownloadBytes + totalUploadBytes
		Stats.dataTransferPeriod = d
	} else {
		Stats.totalBytes += totalDownloadBytes + totalUploadBytes
		Stats.dataTransferPeriod += d
	}
	Stats.lastDataTransfer = time.Now()
	Stats.mtx.Unlock()
}

func GetDataTransferRate() int32 {
	Stats.mtx.RLock()
	rate := int32(Stats.totalBytes / int(Stats.dataTransferPeriod/time.Millisecond+1))
	Stats.mtx.RUnlock()
	return rate
}

func ServerResponseTime(reqConstructor, resConstructor int64, t time.Duration) {
	if t > serverLongThreshold {
		logs.Warn("Too Long ServerResponse",
			zap.Duration("T", t),
			zap.String("ResConstructor", msg.ConstructorNames[resConstructor]),
			zap.String("ReqConstructor", msg.ConstructorNames[reqConstructor]),
		)
	}
	total := atomic.AddInt32(&Stats.TotalServerRequests, 1)
	Stats.mtx.Lock()
	Stats.AvgServerResponseTime = (Stats.AvgServerResponseTime*time.Duration(total-1) + t) / time.Duration(total)
	if t > Stats.MaxServerResponseTime {
		Stats.MaxServerResponseTime = t
	}
	if t < Stats.MinServerResponseTime || Stats.MinServerResponseTime == 0 {
		Stats.MinServerResponseTime = t
	}
	Stats.mtx.Unlock()
}

func QueueTime(t time.Duration) {
	total := atomic.AddInt32(&Stats.TotalQueueItems, 1)
	Stats.mtx.Lock()
	Stats.AvgQueueTime = (Stats.AvgQueueTime*time.Duration(total-1) + t) / time.Duration(total)
	if t > Stats.MaxQueueTime {
		Stats.MaxQueueTime = t
	}
	if t < Stats.MinQueueTime || Stats.MinQueueTime == 0 {
		Stats.MinQueueTime = t
	}
	Stats.mtx.Unlock()
}
