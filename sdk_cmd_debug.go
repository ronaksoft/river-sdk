package riversdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	fileCtrl "git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_file"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	mon "git.ronaksoftware.com/ronak/riversdk/pkg/monitoring"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"git.ronaksoftware.com/ronak/riversdk/pkg/salt"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/dustin/go-humanize"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
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

type dummyDelegate struct{}

func (d dummyDelegate) OnComplete(b []byte) {}

func (d dummyDelegate) OnTimeout(err error) {}

func sendToSavedMessage(r *River, body string) {
	req := &msg.MessagesSend{
		RandomID: 0,
		Peer: &msg.InputPeer{
			ID:         r.ConnInfo.UserID,
			Type:       msg.PeerUser,
			AccessHash: 0,
		},
		Body:       body,
		ReplyTo:    0,
		ClearDraft: true,
		Entities:   nil,
	}
	reqBytes, _ := req.Marshal()
	_, _ = r.ExecuteCommand(msg.C_MessagesSend, reqBytes, &dummyDelegate{}, false, false)
}

func sendMediaToSaveMessage(r *River, filePath string) {
	req := &msg.ClientSendMessageMedia{
		Peer: &msg.InputPeer{
			ID:         r.ConnInfo.UserID,
			Type:       msg.PeerUser,
			AccessHash: 0,
		},
		MediaType:      msg.InputMediaTypeUploadedDocument,
		Caption:        "",
		FileName:       "",
		FilePath:       filePath,
		ThumbFilePath:  "",
		FileMIME:       "",
		ThumbMIME:      "",
		ReplyTo:        0,
		ClearDraft:     false,
		Attributes:     nil,
		FileUploadID:   "",
		ThumbUploadID:  "",
		FileID:         0,
		ThumbID:        0,
		FileTotalParts: 0,
	}
	reqBytes, _ := req.Marshal()
	_, _ = r.ExecuteCommand(msg.C_ClientSendMessageMedia, reqBytes, &dummyDelegate{}, false, false)
}

func (r *River) handleDebugActions(txt string) {
	parts := strings.Fields(strings.ToLower(txt))
	if len(parts) == 0 {
		return
	}
	switch parts[0] {
	case "//sdk_clear_salt":
		resetSalt(r)
	case "//sdk_memory_stats":
		getMemoryStats(r)
	case "//sdk_monitor":
		getMonitorStats(r)
	case "//sdk_live_logger":
		if len(parts) < 2 {
			sendToSavedMessage(r, "//sdk_live_logger <url>")
		}
		liveLogger(r, parts[1])
	case "//sdk_heap_profile":
		filePath := heapProfile()
		if filePath == "" {
			sendToSavedMessage(r, "something wrong, check sdk logs")
		}
		sendMediaToSaveMessage(r, filePath)
	case "//sdk_logs_clear":
		_ = filepath.Walk(logs.LogDir, func(path string, info os.FileInfo, err error) error {
			if strings.HasSuffix(info.Name(), ".log") {
				_ = os.Remove(path)
			}
			return nil
		})
	case "//sdk_logs_update":
		sendUpdateLogs(r)



	}
}

func resetSalt(r *River) {
	salt.Reset()
	sendToSavedMessage(r, "SDK salt is cleared")
}

func getMemoryStats(r *River) []byte {
	ms := new(runtime.MemStats)
	runtime.ReadMemStats(ms)
	m := ronak.M{
		"HeapAlloc":   humanize.Bytes(ms.HeapAlloc),
		"HeapInuse":   humanize.Bytes(ms.HeapInuse),
		"HeapIdle":    humanize.Bytes(ms.HeapIdle),
		"HeapObjects": ms.HeapObjects,
	}
	b, _ := json.MarshalIndent(m, "", "    ")
	sendToSavedMessage(r, ronak.ByteToStr(b))
	return b
}

func getMonitorStats(r *River) []byte {
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

func liveLogger(r *River, url string) {
	logs.SetRemoteLog(url)
	sendToSavedMessage(r, "Live Logger is On")
}

func heapProfile() (filePath string) {
	buf := new(bytes.Buffer)
	err := pprof.WriteHeapProfile(buf)
	if err != nil {
		logs.Error("Error On HeapProfile", zap.Error(err))
		return ""
	}
	now := time.Now()
	filePath = path.Join(fileCtrl.DirCache, fmt.Sprintf("MemHeap-%04d-%02d-%02d.out", now.Year(), now.Month(), now.Day()))
	if err := ioutil.WriteFile(filePath, buf.Bytes(), os.ModePerm); err != nil {
		logs.Warn("River got error on creating memory heap file", zap.Error(err))
		return ""
	}
	return
}

func sendUpdateLogs(r *River) {
	_ = filepath.Walk(logs.LogDir, func(path string, info os.FileInfo, err error) error {
		if strings.HasPrefix(info.Name(), "UPDT") {
			sendMediaToSaveMessage(r, path)
		}
		return nil
	})
}

func (r *River) GetHole(peerID int64, peerType int32) []byte {
	return repo.MessagesExtra.GetHoles(peerID, peerType)
}
