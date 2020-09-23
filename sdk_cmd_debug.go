package riversdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"git.ronaksoft.com/river/msg/msg"
	"git.ronaksoft.com/river/sdk/internal/logs"
	mon "git.ronaksoft.com/river/sdk/internal/monitoring"
	"git.ronaksoft.com/river/sdk/pkg/domain"
	"git.ronaksoft.com/river/sdk/pkg/repo"
	"git.ronaksoft.com/river/sdk/pkg/salt"
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

func (d dummyDelegate) Flags() int32 {
	return 0
}

func sendToSavedMessage(r *River, body string, entities ...*msg.MessageEntity) {
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
		Entities:   entities,
	}
	reqBytes, _ := req.Marshal()
	_, _ = r.ExecuteCommand(msg.C_MessagesSend, reqBytes, &dummyDelegate{})
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
	_, _ = r.ExecuteCommand(msg.C_ClientSendMessageMedia, reqBytes, &dummyDelegate{})
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
		sendToSavedMessage(r, domain.ByteToStr(getMemoryStats(r)))
	case "//sdk_monitor":
		txt := domain.ByteToStr(getMonitorStats(r))
		sendToSavedMessage(r, txt,
			&msg.MessageEntity{
				Type:   msg.MessageEntityTypeCode,
				Offset: 0,
				Length: int32(len(txt)),
				UserID: 0,
			},
		)
	case "//sdk_monitor_reset":
		mon.ResetUsage()
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
		_ = filepath.Walk(logs.Directory(), func(path string, info os.FileInfo, err error) error {
			if strings.HasSuffix(info.Name(), ".log") {
				_ = os.Remove(path)
			}
			return nil
		})
	case "//sdk_logs":
		sendLogs(r)
	case "//sdk_logs_update":
		sendUpdateLogs(r)
	case "//sdk_export_messages":
		if len(args) < 2 {
			logs.Warn("invalid args: //sdk_export_messages [peerType] [peerID]")
			return
		}
		peerType := domain.StrToInt32(args[0])
		peerID := domain.StrToInt64(args[1])
		sendMediaToSaveMessage(r, exportMessages(r, peerType, peerID), fmt.Sprintf("Messages-%s-%d.txt", msg.PeerType(peerType).String(), peerID))
	}
}
func exportMessages(r *River, peerType int32, peerID int64) (filePath string) {
	filePath = path.Join(repo.DirCache, fmt.Sprintf("Messages-%s-%d.txt", msg.PeerType(peerType).String(), peerID))
	file, err := os.Create(filePath)
	logs.ErrorOnErr("Error On Create file", err)

	t := tablewriter.NewWriter(file)
	t.SetHeader([]string{"ID", "Date", "Sender", "Body", "Media"})
	maxID, _ := repo.Messages.GetTopMessageID(domain.GetCurrTeamID(), peerID, peerType)
	limit := int32(100)
	cnt := 0
	for {
		ms, us := repo.Messages.GetMessageHistory(domain.GetCurrTeamID(), peerID, peerType, 0, maxID, limit)
		usMap := make(map[int64]*msg.User)
		for _, u := range us {
			usMap[u.ID] = u
		}
		for _, m := range ms {
			b := m.Body
			if idx := strings.Index(m.Body, "\n"); idx < 0 {
				if len(m.Body) > 100 {
					b = m.Body[:100]
				}
			} else if idx < 100 {
				b = m.Body[:idx]
			} else {
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

		if int32(len(ms)) < limit {
			break
		}
	}
	t.SetFooter([]string{"Total", fmt.Sprintf("%d", cnt), "", "", ""})
	t.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	t.SetCenterSeparator("|")
	// t.SetAutoMergeCells(true)
	// t.SetRowLine(true)
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
	m := domain.M{
		"HeapAlloc":   humanize.Bytes(ms.HeapAlloc),
		"HeapInuse":   humanize.Bytes(ms.HeapInuse),
		"HeapIdle":    humanize.Bytes(ms.HeapIdle),
		"HeapObjects": ms.HeapObjects,
	}
	b, _ := json.MarshalIndent(m, "", "    ")
	sendToSavedMessage(r, domain.ByteToStr(b))
	return b
}
func getMonitorStats(r *River) []byte {
	lsmSize, logSize := repo.DbSize()
	s := mon.Stats
	m := domain.M{
		"ServerAvgTime":    (time.Duration(s.AvgResponseTime) * time.Millisecond).String(),
		"ServerRequests":   s.TotalServerRequests,
		"RecordTime":       time.Now().Sub(s.StartTime).String(),
		"ForegroundTime":   (time.Duration(s.ForegroundTime) * time.Second).String(),
		"SentMessages":     s.SentMessages,
		"SentMedia":        s.SentMedia,
		"ReceivedMessages": s.ReceivedMessages,
		"ReceivedMedia":    s.ReceivedMedia,
		"Upload":           humanize.Bytes(uint64(s.TotalUploadBytes)),
		"Download":         humanize.Bytes(uint64(s.TotalDownloadBytes)),
		"LsmSize":          humanize.Bytes(uint64(lsmSize)),
		"LogSize":          humanize.Bytes(uint64(logSize)),
		"Version":          r.Version(),
	}

	b, _ := json.MarshalIndent(m, "", "  ")
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
		logs.Error("We got error on getting heap profile", zap.Error(err))
		return ""
	}
	now := time.Now()
	filePath = path.Join(repo.DirCache, fmt.Sprintf("MemHeap-%04d-%02d-%02d.out", now.Year(), now.Month(), now.Day()))
	if err := ioutil.WriteFile(filePath, buf.Bytes(), os.ModePerm); err != nil {
		logs.Warn("We got error on creating memory heap file", zap.Error(err))
		return ""
	}
	return
}
func sendUpdateLogs(r *River) {
	_ = filepath.Walk(logs.Directory(), func(path string, info os.FileInfo, err error) error {
		if strings.HasPrefix(info.Name(), "UPDT") {
			sendMediaToSaveMessage(r, path, info.Name())
		}
		return nil
	})
}
func sendLogs(r *River) {
	_ = filepath.Walk(logs.Directory(), func(filePath string, info os.FileInfo, err error) error {
		if strings.HasPrefix(info.Name(), "LOG") {
			outPath := path.Join(repo.DirCache, info.Name())
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
	return repo.MessagesExtra.GetHoles(domain.GetCurrTeamID(), peerID, peerType)
}

func (r *River) CancelFileRequest(reqID string) {
	err := repo.Files.DeleteFileRequest(reqID)
	if err != nil {
		logs.Warn("River got error on delete file request", zap.Error(err))
	}
}
