package riversdk

import (
	"bytes"
	"encoding/json"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	mon "git.ronaksoftware.com/ronak/riversdk/pkg/monitoring"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/dustin/go-humanize"
	"go.uber.org/zap"
	"runtime"
	"runtime/pprof"
	"time"
)

/*
   Creation Time: 2019 - Jul - 28
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

func (r *River) GetHole(peerID int64, peerType int32) []byte {
	return repo.MessagesExtra.GetHoles(peerID, peerType)
}

func (r *River) GetMonitorStats() []byte {
	lsmSize, logSize := repo.DbSize()
	s := mon.Stats
	m := ronak.M{
		"AvgServerTime":   s.AvgServerResponseTime.String(),
		"MaxServerTime":   s.MaxServerResponseTime.String(),
		"MinServerTime":   s.MinServerResponseTime.String(),
		"ServerRequests":  s.TotalServerRequests,
		"AvgFunctionTime": s.AvgFunctionResponseTime.String(),
		"MaxFunctionTime": s.MaxFunctionResponseTime.String(),
		"MinFunctionTime": s.MinFunctionResponseTime.String(),
		"FunctionCalls":   s.TotalFunctionCalls,
		"AvgQueueTime":    s.AvgQueueTime.String(),
		"MaxQueueTime":    s.MaxQueueTime.String(),
		"MinQueueTime":    s.MinQueueTime.String(),
		"QueueItems":      s.TotalQueueItems,
		"RecordTime":      time.Now().Sub(s.StartTime).String(),
		"TableInfos":      repo.TableInfo(),
		"LsmSize":         humanize.Bytes(uint64(lsmSize)),
		"LogSize":         humanize.Bytes(uint64(logSize)),
		"Version":         r.Version(),
	}

	b, _ := json.MarshalIndent(m, "", "    ")
	return b
}

func (r *River) GetMemoryStats() []byte {
	ms := new(runtime.MemStats)
	runtime.ReadMemStats(ms)
	m := ronak.M{
		"HeapAlloc":   humanize.Bytes(ms.HeapAlloc),
		"HeapInuse":   humanize.Bytes(ms.HeapInuse),
		"HeapIdle":    humanize.Bytes(ms.HeapIdle),
		"HeapObjects": ms.HeapObjects,
	}

	b, _ := json.MarshalIndent(m, "", "    ")
	return b
}

func (r *River) GetHeapProfile() []byte {
	buf := new(bytes.Buffer)
	err := pprof.WriteHeapProfile(buf)
	if err != nil {
		logs.Error("Error On HeapProfile", zap.Error(err))
		return nil
	}
	return buf.Bytes()
}

func (r *River) TurnOnLiveLogger(url string) {
	logs.SetRemoteLog(url)
}
