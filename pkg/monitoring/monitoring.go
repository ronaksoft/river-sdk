package mon

import (
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"go.uber.org/zap"
	"sync"
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
	mtx                     sync.Mutex
	AvgServerResponseTime   time.Duration
	AvgFunctionResponseTime time.Duration
	AvgQueueTime            time.Duration
	MaxServerResponseTime   time.Duration
	MaxFunctionResponseTime time.Duration
	MaxQueueTime            time.Duration
	MinServerResponseTime   time.Duration
	MinFunctionResponseTime time.Duration
	MinQueueTime            time.Duration
	TotalQueueItems         int32
	TotalServerRequests     int32
	TotalFunctionCalls      int32
}

var Stats stats

func ServerResponseTime(constructor int64, t time.Duration) {
	if t > serverLongThreshold {
		logs.Warn("Too Long ServerResponse", zap.Duration("T", t), zap.String("Constructor", msg.ConstructorNames[constructor]))
	}
	Stats.mtx.Lock()
	Stats.AvgServerResponseTime = (Stats.AvgServerResponseTime*time.Duration(Stats.TotalServerRequests) + t) / time.Duration(Stats.TotalServerRequests)
	if t > Stats.MaxServerResponseTime {
		Stats.MaxServerResponseTime = t
	}
	if t < Stats.MinServerResponseTime || Stats.MinServerResponseTime == 0 {
		Stats.MinServerResponseTime = t
	}
	Stats.mtx.Unlock()
}

func QueueTime(constructor int64, t time.Duration) {
	if t > serverLongThreshold {
		logs.Warn("Too Long QueueTime", zap.Duration("T", t), zap.String("Constructor", msg.ConstructorNames[constructor]))
	}
	Stats.mtx.Lock()
	Stats.AvgQueueTime = (Stats.AvgQueueTime*time.Duration(Stats.TotalQueueItems) + t) / time.Duration(Stats.TotalServerRequests)
	if t > Stats.MaxQueueTime {
		Stats.MaxQueueTime = t
	}
	if t < Stats.MinQueueTime || Stats.MinQueueTime == 0 {
		Stats.MinQueueTime = t
	}
	Stats.mtx.Unlock()
}

func FunctionResponseTime(funcName string, t time.Duration, v ...interface{}) {
	if t > serverLongThreshold {
		logs.Warn("Too Long FunctionResponse", zap.Duration("T", t),
			zap.String("FN", funcName),
			zap.Any("Extra", v),
		)
	}
	Stats.mtx.Lock()
	Stats.AvgFunctionResponseTime = (Stats.AvgFunctionResponseTime*time.Duration(Stats.TotalFunctionCalls) + t) / time.Duration(Stats.TotalServerRequests)
	if t > Stats.MaxFunctionResponseTime {
		Stats.MaxFunctionResponseTime = t
	}
	if t < Stats.MinFunctionResponseTime || Stats.MinFunctionResponseTime == 0 {
		Stats.MinFunctionResponseTime = t
	}
	Stats.mtx.Unlock()
}
