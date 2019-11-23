package riversdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	fileCtrl "git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_file"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	mon "git.ronaksoftware.com/ronak/riversdk/pkg/monitoring"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"git.ronaksoftware.com/ronak/riversdk/pkg/salt"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/dustin/go-humanize"
	"github.com/olekukonko/tablewriter"
	"go.uber.org/zap"
	"io"
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

func sendMediaToSaveMessage(r *River, filePath string, filename string) {
	attrFile := msg.DocumentAttributeFile{Filename: filename}
	attBytes, _ := attrFile.Marshal()
	req := &msg.ClientSendMessageMedia{
		Peer: &msg.InputPeer{
			ID:         r.ConnInfo.UserID,
			Type:       msg.PeerUser,
			AccessHash: 0,
		},
		MediaType:     msg.InputMediaTypeUploadedDocument,
		Caption:       "",
		FileName:      filename,
		FilePath:      filePath,
		ThumbFilePath: "",
		FileMIME:      "",
		ThumbMIME:     "",
		ReplyTo:       0,
		ClearDraft:    false,
		Attributes: []*msg.DocumentAttribute{
			{Type: msg.AttributeTypeFile, Data: attBytes},
		},
		FileUploadID:   "",
		ThumbUploadID:  "",
		FileID:         0,
		ThumbID:        0,
		FileTotalParts: 0,
	}
	reqBytes, _ := req.Marshal()
	_, _ = r.ExecuteCommand(msg.C_ClientSendMessageMedia, reqBytes, &dummyDelegate{}, false, false)
}

func (r *River) HandleDebugActions(txt string) {
	parts := strings.Fields(strings.ToLower(txt))
	if len(parts) == 0 {
		return
	}
	cmd := parts[0]
	args := parts[1:]
	switch cmd {
	case "//sdk_clear_salt":
		resetSalt(r)
	case "//sdk_memory_stats":
		sendToSavedMessage(r, ronak.ByteToStr(getMemoryStats(r)))
	case "//sdk_monitor":
		sendToSavedMessage(r, ronak.ByteToStr(getMonitorStats(r)))
	case "//sdk_live_logger":

		if len(args) < 1 {
			sendToSavedMessage(r, "//sdk_live_logger <url>")
			return
		}
		liveLogger(r, args[0])
	case "//sdk_heap_profile":
		filePath := heapProfile()
		if filePath == "" {
			sendToSavedMessage(r, "something wrong, check sdk logs")
		}
		sendMediaToSaveMessage(r, filePath, "SdkHeapProfile.out")
	case "//sdk_logs_clear":
		_ = filepath.Walk(logs.LogDir, func(path string, info os.FileInfo, err error) error {
			if strings.HasSuffix(info.Name(), ".log") {
				_ = os.Remove(path)
			}
			return nil
		})
	case "//sdk_logs":
		sendLogs(r)
	case "//sdk_logs_update":
		sendUpdateLogs(r)
	case "//sdk_logs_window":
		r.mainDelegate.ShowLoggerAlert()
	case "//sdk_export_messages":
		if len(args) < 2 {
			logs.Warn("invalid args: //sdk_export_messages [peerType] [peerID]")
			return
		}
		peerType := ronak.StrToInt32(args[0])
		peerID := ronak.StrToInt64(args[1])
		sendMediaToSaveMessage(r, exportMessages(r, peerType, peerID), fmt.Sprintf("Messages-%s-%d.out", msg.PeerType(peerType).String(), peerID))
	}
}

func exportMessages(r *River, peerType int32, peerID int64) (filePath string) {
	filePath = path.Join(fileCtrl.DirCache, fmt.Sprintf("Messages-%s-%d.out", msg.PeerType(peerType).String(), peerID))
	file, err := os.Create(filePath)
	logs.ErrorOnErr("Error On Create file", err)

	t := tablewriter.NewWriter(file)
	t.SetHeader([]string{"ID", "Date", "Sender", "Body", "Media"})
	maxID, _ := repo.Messages.GetTopMessageID(peerID, peerType)
	limit := int32(100)
	cnt := 0
	for {
		ms, us := repo.Messages.GetMessageHistory(peerID, peerType, 0, maxID, limit)
		if int32(len(ms)) < limit {
			break
		}
		usMap := make(map[int64]*msg.User)
		for _, u := range us {
			usMap[u.ID] = u
		}
		for _, m := range ms {
			b := m.Body
			if len(m.Body) > 100 {
				b = m.Body[:100]
			}
			t.Append([]string{
				fmt.Sprintf("%d", m.ID),
				time.Unix(m.CreatedOn, 0).Format("02 Jan 06 3:04PM"),
				fmt.Sprintf("%s %s", usMap[m.SenderID].FirstName, usMap[m.SenderID].LastName),
				b,
				m.MediaType.String(),
			})
			cnt++
			if maxID > m.ID {
				maxID = m.ID
			}
		}

	}
	t.SetFooter([]string{"Total", fmt.Sprintf("%d", cnt), "", "", ""})
	t.Render()
	_, _ = io.WriteString(file, "\n\n")
	_, _ = io.WriteString(file, string(r.GetHole(peerID, peerType)))
	return
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
			sendMediaToSaveMessage(r, path, info.Name())
		}
		return nil
	})
}

func sendLogs(r *River) {
	_ = filepath.Walk(logs.LogDir, func(filePath string, info os.FileInfo, err error) error {
		if strings.HasPrefix(info.Name(), "LOG") {
			outPath := path.Join(fileCtrl.DirCache, info.Name())
			err = domain.CopyFile(filePath, outPath)
			if err != nil {
				return err
			}
			sendMediaToSaveMessage(r, outPath, info.Name())
		}
		return nil
	})
}

func (r *River) GetHole(peerID int64, peerType int32) []byte {
	return repo.MessagesExtra.GetHoles(peerID, peerType)
}
