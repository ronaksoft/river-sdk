package mon

import (
	msg "git.ronaksoftware.com/river/msg/chat"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
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
	mtx                *sync.RWMutex
	StartTime          time.Time
	LastForegroundTime time.Time

	TotalServerRequests int64
	AvgResponseTime     int64
	TotalUploadBytes    int64
	TotalDownloadBytes  int64
	SentMessages        int64
	ReceivedMessages    int64
	SentMedia           int64
	ReceivedMedia       int64
	ForegroundTime      int64

	// File DataTransferRate
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
	Stats.TotalDownloadBytes += int64(totalDownloadBytes)
	Stats.TotalUploadBytes += int64(totalUploadBytes)
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
	ts := t.Milliseconds()
	Stats.mtx.Lock()
	Stats.TotalServerRequests += 1
	Stats.AvgResponseTime = (Stats.AvgResponseTime*(Stats.TotalServerRequests-1) + ts) / (Stats.TotalServerRequests)
	Stats.mtx.Unlock()
}

func IncMessageSent() {
	Stats.mtx.Lock()
	Stats.SentMessages += 1
	Stats.mtx.Unlock()
}

func IncMediaSent() {
	Stats.mtx.Lock()
	Stats.SentMedia += 1
	Stats.mtx.Unlock()
}

func IncMessageReceived() {
	Stats.mtx.Lock()
	Stats.ReceivedMessages += 1
	Stats.mtx.Unlock()
}

func IncMediaReceived() {
	Stats.mtx.Lock()
	Stats.ReceivedMedia += 1
	Stats.mtx.Unlock()
}

func SetForegroundTime() {
	Stats.mtx.Lock()
	Stats.LastForegroundTime = time.Now()
	Stats.mtx.Unlock()
}

func IncForegroundTime() {
	Stats.mtx.Lock()
	Stats.ForegroundTime += int64(time.Now().Sub(Stats.LastForegroundTime).Seconds())
	Stats.mtx.Unlock()
}

func LoadUsage() {
	now := time.Now()
	cu := &msg.ClientUsage{}
	b, err := repo.System.LoadBytes("ClientUsage")
	if err == nil {
		err = cu.Unmarshal(b)
	}
	if err != nil {
		cu.Year = int32(now.Year())
		cu.Month = int32(now.Month())
		cu.Day = int32(now.Day())
	}
	Stats.mtx.Lock()
	Stats.ForegroundTime = cu.ForegroundTime
	Stats.ReceivedMessages = cu.ReceivedMessages
	Stats.ReceivedMedia = cu.ReceivedMedia
	Stats.SentMedia = cu.SentMedia
	Stats.SentMessages = cu.SentMessages
	Stats.AvgResponseTime = cu.AvgResponseTime
	Stats.TotalServerRequests = cu.TotalRequests
	Stats.mtx.Unlock()
}

func SaveUsage() {
	cu := &msg.ClientUsage{}
	Stats.mtx.Lock()
	cu.ForegroundTime = Stats.ForegroundTime
	cu.ReceivedMedia = Stats.ReceivedMedia
	cu.ReceivedMessages = Stats.ReceivedMessages
	cu.SentMedia = Stats.SentMedia
	cu.SentMessages = Stats.SentMessages
	cu.AvgResponseTime = Stats.AvgResponseTime
	cu.TotalRequests = Stats.TotalServerRequests
	Stats.mtx.Unlock()
	b, err := cu.Marshal()
	if err == nil {
		err = repo.System.SaveBytes("ClientUsage", b)
		if err != nil {
			logs.Warn("We got error on saving ClientUsage into the db", zap.Error(err))
		}
	}
}

func GetUploadUsageTime() (int64, error) {
	return repo.System.LoadInt64("UsageUploadTS")
}

func SaveUploadUsageTime(ts int64) error {
	return repo.System.SaveInt("UsageUploadTS", uint64(ts))
}

func ResetUsage() {
	_ = repo.System.Delete("ClientUsage")
	Stats.mtx.Lock()
	Stats.ReceivedMessages = 0
	Stats.ReceivedMedia = 0
	Stats.SentMedia = 0
	Stats.SentMessages = 0
	Stats.AvgResponseTime = 0
	Stats.TotalServerRequests = 0
	Stats.mtx.Unlock()
	SaveUsage()
}
