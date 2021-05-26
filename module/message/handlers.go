package message

import (
	"bytes"
	"encoding/json"
	"fmt"
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	messageHole "git.ronaksoft.com/river/sdk/internal/message_hole"
	mon "git.ronaksoft.com/river/sdk/internal/monitoring"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"git.ronaksoft.com/river/sdk/internal/request"
	"git.ronaksoft.com/river/sdk/internal/salt"
	"git.ronaksoft.com/river/sdk/internal/uiexec"
	"github.com/dustin/go-humanize"
	"github.com/olekukonko/tablewriter"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/tools"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"sort"
	"strings"
	"time"
)

/*
   Creation Time: 2021 - May - 10
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func (r *message) messagesGetDialogs(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.MessagesGetDialogs{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}
	res := &msg.MessagesDialogs{}
	res.Dialogs, _ = repo.Dialogs.List(domain.GetTeamID(in), req.Offset, req.Limit)
	res.Count = repo.Dialogs.CountDialogs(domain.GetTeamID(in))

	// If the localDB had no data send the request to server
	if len(res.Dialogs) == 0 {
		res.UpdateID = r.SDK().SyncCtrl().GetUpdateID()
		out.Constructor = msg.C_MessagesDialogs
		buff, err := res.Marshal()
		if err != nil {
			r.Log().Error("got error on marshal MessagesDialogs", zap.Error(err))
		}

		out.Message = buff
		uiexec.ExecSuccessCB(da.OnComplete, out)
		return
	}

	pendingMessages := repo.PendingMessages.GetAndConvertAll()
	dialogPMs := make(map[string]*msg.UserMessage, len(pendingMessages))
	for _, pm := range pendingMessages {
		keyID := fmt.Sprintf("%d.%d", pm.PeerID, pm.PeerType)
		v, ok := dialogPMs[keyID]
		if !ok {
			dialogPMs[keyID] = pm
		} else if pm.ID < v.ID {
			dialogPMs[keyID] = pm
		}
	}
	mUsers := domain.MInt64B{}
	mGroups := domain.MInt64B{}
	mMessages := domain.MInt64B{}
	for _, dialog := range res.Dialogs {
		if dialog.PeerType == int32(msg.PeerType_PeerUser) {
			if dialog.PeerID != 0 {
				mUsers[dialog.PeerID] = true
			}
		}
		mMessages[dialog.TopMessageID] = true
		keyID := fmt.Sprintf("%d.%d", dialog.PeerID, dialog.PeerType)
		if pm, ok := dialogPMs[keyID]; ok {
			dialog.TopMessageID = pm.ID
		}
	}

	// Load Messages
	res.Messages, _ = repo.Messages.GetMany(mMessages.ToArray())

	// Load Pending messages
	res.Messages = append(res.Messages, pendingMessages...)
	for _, m := range res.Messages {
		switch msg.PeerType(m.PeerType) {
		case msg.PeerType_PeerUser:
			mUsers[m.PeerID] = true
		case msg.PeerType_PeerGroup:
			mGroups[m.PeerID] = true
		}
		if m.SenderID != 0 {
			mUsers[m.SenderID] = true
		}
		if m.FwdSenderID != 0 {
			mUsers[m.FwdSenderID] = true
		}

		// load MessageActionData users
		actUserIDs := domain.ExtractActionUserIDs(m.MessageAction, m.MessageActionData)
		for _, id := range actUserIDs {
			if id != 0 {
				mUsers[id] = true
			}
		}
	}
	res.Groups, _ = repo.Groups.GetMany(mGroups.ToArray())
	if len(res.Groups) != len(mGroups) {
		r.Log().Warn("found unmatched dialog groups", zap.Int("Got", len(res.Groups)), zap.Int("Need", len(mGroups)))
		for groupID := range mGroups {
			found := false
			for _, g := range res.Groups {
				if g.ID == groupID {
					found = true
					break
				}
			}
			if !found {
				r.Log().Warn("missed group", zap.Int64("GroupID", groupID))
			}
		}
	}
	res.Users, _ = repo.Users.GetMany(mUsers.ToArray())
	if len(res.Users) != len(mUsers) {
		r.Log().Warn("found unmatched dialog users", zap.Int("Got", len(res.Users)), zap.Int("Need", len(mUsers)))
		for userID := range mUsers {
			found := false
			for _, g := range res.Users {
				if g.ID == userID {
					found = true
					break
				}
			}
			if !found {
				r.Log().Warn("missed user", zap.Int64("UserID", userID))
			}
		}
	}

	out.Constructor = msg.C_MessagesDialogs
	buff, err := res.Marshal()
	if err != nil {
		r.Log().Error("got error on marshal MessagesDialogs", zap.Error(err))
	}

	out.Message = buff
	uiexec.ExecSuccessCB(da.OnComplete, out)
}

func (r *message) messagesGetDialog(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.MessagesGetDialog{}
	err := req.Unmarshal(in.Message)
	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}
	res := &msg.Dialog{}
	res, err = repo.Dialogs.Get(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type))

	// if the localDB had no data send the request to server
	if err != nil {
		r.Log().Warn("got error on repo GetDialog", zap.Error(err), zap.Int64("PeerID", req.Peer.ID))
		r.SDK().QueueCtrl().EnqueueCommand(da)
		return
	}

	out.Constructor = msg.C_Dialog
	out.Message, _ = res.Marshal()

	uiexec.ExecSuccessCB(da.OnComplete, out)
}

func (r *message) messagesSend(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.MessagesSend{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	// do not allow empty message
	if strings.TrimSpace(req.Body) == "" {
		e := &rony.Error{
			Code:  "n/a",
			Items: "empty message is not allowed",
		}
		out.Fill(out.RequestID, rony.C_Error, e)
		uiexec.ExecSuccessCB(da.OnComplete, out)
		return
	}

	if req.Peer.ID == r.SDK().GetConnInfo().PickupUserID() {
		r.handleDebugActions(req.Body)
	}

	// this will be used as next requestID
	req.RandomID = domain.SequentialUniqueID()
	msgID := -req.RandomID
	res, err := repo.PendingMessages.Save(domain.GetTeamID(in), domain.GetTeamAccess(in), msgID, r.SDK().GetConnInfo().PickupUserID(), req)
	if err != nil {
		e := &rony.Error{
			Code:  "n/a",
			Items: "Failed to save to pendingMessages : " + err.Error(),
		}
		out.Fill(out.RequestID, rony.C_Error, e)
		uiexec.ExecSuccessCB(da.OnComplete, out)
		return
	}

	// using req randomID as requestID later in queue processing and network controller messageHandler
	da.Discard()
	r.SDK().QueueCtrl().EnqueueCommand(
		request.NewCallback(
			da.TeamID(), da.TeamAccess(), uint64(req.RandomID), msg.C_MessagesSend, req,
			da.OnTimeout, da.OnComplete, da.OnProgress, da.UI(), da.Flags(), da.Timeout(),
		),
	)

	// 3. return to CallBack with pending message data : Done
	out.Constructor = msg.C_ClientPendingMessage
	out.Message, _ = res.Marshal()

	// 4. later when queue got processed and server returned response we should check if the requestID
	//   exist in pendingTable we remove it and insert new message with new id to message table
	//   invoke new OnUpdate with new proto buffer to inform ui that pending message got delivered
	uiexec.ExecSuccessCB(da.OnComplete, out)
}
func (r *message) handleDebugActions(txt string) {
	parts := strings.Fields(strings.ToLower(txt))
	if len(parts) == 0 {
		return
	}
	cmd := parts[0]
	args := parts[1:]
	switch cmd {
	case "//sdk_clear_salt":
		r.resetSalt()
	case "//sdk_memory_stats":
		r.sendToSavedMessage(tools.ByteToStr(r.getMemoryStats()))
	case "//sdk_monitor":
		txt := tools.ByteToStr(r.getMonitorStats())
		r.sendToSavedMessage(
			txt,
			&msg.MessageEntity{
				Type:   msg.MessageEntityType_MessageEntityTypeCode,
				Offset: 0,
				Length: int32(len(txt)),
				UserID: 0,
			},
		)
	case "//sdk_monitor_reset":
		mon.ResetUsage()
	case "//sdk_live_logger":
		username := r.SDK().GetConnInfo().PickupUsername()
		if len(args) < 1 {
			if username == "" {
				r.sendToSavedMessage("//sdk_live_logger <url>")
			} else {
				r.liveLogger(fmt.Sprintf("https://livelog.ronaksoftware.com/%s", username))
			}
			return
		}

		r.liveLogger(args[0])
	case "//sdk_heap_profile":
		filePath := r.heapProfile()
		if filePath == "" {
			r.sendToSavedMessage("something wrong, check sdk logs")
		}
		r.sendMediaToSaveMessage(filePath, "SdkHeapProfile.out")
	case "//sdk_logs_clear":
		_ = filepath.Walk(logs.Directory(), func(path string, info os.FileInfo, err error) error {
			if strings.HasSuffix(info.Name(), ".log") {
				_ = os.Remove(path)
			}
			return nil
		})
	case "//sdk_logs":
		r.sendLogs()
	case "//sdk_logs_update":
		r.sendUpdateLogs()
	case "//sdk_export_messages":
		if len(args) < 2 {
			r.Log().Warn("invalid args: //sdk_export_messages [peerType] [peerID]")
			return
		}
		peerType := tools.StrToInt32(args[0])
		peerID := tools.StrToInt64(args[1])
		r.sendMediaToSaveMessage(r.exportMessages(peerType, peerID), fmt.Sprintf("Messages-%s-%d.txt", msg.PeerType(peerType).String(), peerID))
	case "//sdk_update_state_get":
		r.getUpdateState()
	case "//sdk_update_state_set":
		r.setUpdateState(tools.StrToInt64(args[0]))
	}
}
func (r *message) sendToSavedMessage(body string, entities ...*msg.MessageEntity) {
	req := &msg.MessagesSend{
		RandomID: 0,
		Peer: &msg.InputPeer{
			ID:         r.SDK().GetConnInfo().PickupUserID(),
			Type:       msg.PeerType_PeerUser,
			AccessHash: 0,
		},
		Body:       body,
		ReplyTo:    0,
		ClearDraft: true,
		Entities:   entities,
	}

	in := &rony.MessageEnvelope{}
	out := &rony.MessageEnvelope{}
	in.Fill(domain.NextRequestID(), msg.C_MessagesSend, req)
	r.messagesSend(in, out, request.EmptyCallback())
}
func (r *message) sendMediaToSaveMessage(filePath string, filename string) {
	attrFile := msg.DocumentAttributeFile{Filename: filename}
	attBytes, _ := attrFile.Marshal()
	req := &msg.ClientSendMessageMedia{
		Peer: &msg.InputPeer{
			ID:         r.SDK().GetConnInfo().PickupUserID(),
			Type:       msg.PeerType_PeerUser,
			AccessHash: 0,
		},
		MediaType:     msg.InputMediaType_InputMediaTypeUploadedDocument,
		Caption:       "",
		FileName:      filename,
		FilePath:      filePath,
		ThumbFilePath: "",
		FileMIME:      "",
		ThumbMIME:     "",
		ReplyTo:       0,
		ClearDraft:    false,
		Attributes: []*msg.DocumentAttribute{
			{Type: msg.DocumentAttributeType_AttributeTypeFile, Data: attBytes},
		},
		FileUploadID:   "",
		ThumbUploadID:  "",
		FileID:         0,
		ThumbID:        0,
		FileTotalParts: 0,
	}

	in := &rony.MessageEnvelope{}
	out := &rony.MessageEnvelope{}
	in.Fill(domain.NextRequestID(), msg.C_MessagesSend, req)
	r.clientSendMessageMedia(in, out, request.EmptyCallback())
}
func (r *message) exportMessages(peerType int32, peerID int64) (filePath string) {
	filePath = path.Join(repo.DirCache, fmt.Sprintf("Messages-%s-%d.txt", msg.PeerType(peerType).String(), peerID))
	file, err := os.Create(filePath)
	if err != nil {
		r.Log().Error("Error On Create file", zap.Error(err))
	}

	t := tablewriter.NewWriter(file)
	t.SetHeader([]string{"ID", "Date", "Sender", "Body", "Media"})
	maxID, _ := repo.Messages.GetTopMessageID(domain.GetCurrTeamID(), peerID, peerType)
	limit := int32(100)
	cnt := 0
	for {
		ms, us, _ := repo.Messages.GetMessageHistory(domain.GetCurrTeamID(), peerID, peerType, 0, maxID, limit)
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
	t.Render()
	_, _ = io.WriteString(file, "\n\n")
	return
}
func (r *message) resetSalt() {
	salt.Reset()
	r.sendToSavedMessage("SDK salt is cleared")
}
func (r *message) getMemoryStats() []byte {
	ms := new(runtime.MemStats)
	runtime.ReadMemStats(ms)
	m := domain.M{
		"HeapAlloc":   humanize.Bytes(ms.HeapAlloc),
		"HeapInuse":   humanize.Bytes(ms.HeapInuse),
		"HeapIdle":    humanize.Bytes(ms.HeapIdle),
		"HeapObjects": ms.HeapObjects,
	}
	b, _ := json.MarshalIndent(m, "", "    ")
	r.sendToSavedMessage(tools.ByteToStr(b))
	return b
}
func (r *message) getMonitorStats() []byte {
	lsmSize, logSize := repo.DbSize()
	s := mon.Stats
	m := domain.M{
		"ServerAvgTime":    (time.Duration(s.AvgResponseTime) * time.Millisecond).String(),
		"ServerRequests":   s.TotalServerRequests,
		"RecordTime":       time.Since(s.StartTime).String(),
		"ForegroundTime":   (time.Duration(s.ForegroundTime) * time.Second).String(),
		"SentMessages":     s.SentMessages,
		"SentMedia":        s.SentMedia,
		"ReceivedMessages": s.ReceivedMessages,
		"ReceivedMedia":    s.ReceivedMedia,
		"Upload":           humanize.Bytes(uint64(s.TotalUploadBytes)),
		"Download":         humanize.Bytes(uint64(s.TotalDownloadBytes)),
		"LsmSize":          humanize.Bytes(uint64(lsmSize)),
		"LogSize":          humanize.Bytes(uint64(logSize)),
		"Version":          r.SDK().Version(),
	}

	b, _ := json.MarshalIndent(m, "", "  ")
	return b
}
func (r *message) liveLogger(url string) {
	logs.SetRemoteLog(url)
	r.sendToSavedMessage("Live Logger is On")
}
func (r *message) heapProfile() (filePath string) {
	buf := new(bytes.Buffer)
	err := pprof.WriteHeapProfile(buf)
	if err != nil {
		r.Log().Error("got error on getting heap profile", zap.Error(err))
		return ""
	}
	now := time.Now()
	filePath = path.Join(repo.DirCache, fmt.Sprintf("MemHeap-%04d-%02d-%02d.out", now.Year(), now.Month(), now.Day()))
	if err := ioutil.WriteFile(filePath, buf.Bytes(), os.ModePerm); err != nil {
		r.Log().Warn("got error on creating memory heap file", zap.Error(err))
		return ""
	}
	return
}
func (r *message) sendUpdateLogs() {
	_ = filepath.Walk(logs.Directory(), func(path string, info os.FileInfo, err error) error {
		if strings.HasPrefix(info.Name(), "UPDT") {
			r.sendMediaToSaveMessage(path, info.Name())
		}
		return nil
	})
}
func (r *message) sendLogs() {
	_ = filepath.Walk(logs.Directory(), func(filePath string, info os.FileInfo, err error) error {
		if strings.HasPrefix(info.Name(), "LOG") {
			outPath := path.Join(repo.DirCache, info.Name())
			err = domain.CopyFile(filePath, outPath)
			if err != nil {
				return err
			}
			r.sendMediaToSaveMessage(outPath, info.Name())
		}
		return nil
	})
}
func (r *message) getUpdateState() {
	r.sendToSavedMessage(fmt.Sprintf("UpdateState is %d", r.SDK().SyncCtrl().GetUpdateID()))
}
func (r *message) setUpdateState(updateID int64) {
	r.sendToSavedMessage(fmt.Sprintf("UpdateState set to: %d", updateID))
	_ = r.SDK().SyncCtrl().SetUpdateID(updateID)
	go r.SDK().SyncCtrl().Sync()
}

func (r *message) messagesSendMedia(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.MessagesSendMedia{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	switch req.MediaType {
	case msg.InputMediaType_InputMediaTypeContact, msg.InputMediaType_InputMediaTypeGeoLocation,
		msg.InputMediaType_InputMediaTypeDocument, msg.InputMediaType_InputMediaTypeMessageDocument:
		// This will be used as next requestID
		req.RandomID = domain.SequentialUniqueID()

		// Insert into pending messages, id is negative nano timestamp and save RandomID too : Done
		dbID := -req.RandomID

		res, err := repo.PendingMessages.SaveMessageMedia(domain.GetTeamID(in), domain.GetTeamAccess(in), dbID, r.SDK().GetConnInfo().PickupUserID(), req)
		if err != nil {
			e := &rony.Error{
				Code:  "n/a",
				Items: "Failed to save to pendingMessages : " + err.Error(),
			}
			out.Fill(out.RequestID, rony.C_Error, e)
			uiexec.ExecSuccessCB(da.OnComplete, out)
			return
		}
		// Return to CallBack with pending message data : Done
		out.Constructor = msg.C_ClientPendingMessage

		out.Message, _ = res.Marshal()
		uiexec.ExecSuccessCB(da.OnComplete, out)

	case msg.InputMediaType_InputMediaTypeUploadedDocument:
		// no need to insert pending message cuz we already insert one b4 start uploading
	}

	da.Discard()
	r.SDK().QueueCtrl().EnqueueCommand(
		request.NewCallback(
			da.TeamID(), da.TeamAccess(), uint64(req.RandomID), msg.C_MessagesSendMedia, req,
			da.OnTimeout, da.OnComplete, da.OnProgress, da.UI(), da.Flags(), da.Timeout(),
		),
	)

}

func (r *message) messagesReadHistory(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.MessagesReadHistory{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	dialog, _ := repo.Dialogs.Get(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type))
	if dialog == nil {
		return
	}
	if dialog.ReadInboxMaxID > req.MaxID {
		return
	}

	// update read inbox max id
	_ = repo.Dialogs.UpdateReadInboxMaxID(r.SDK().GetConnInfo().PickupUserID(), domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), req.MaxID)

	// send the request to server
	r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *message) messagesGetHistory(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.MessagesGetHistory{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	// Load the dialog
	dialog, _ := repo.Dialogs.Get(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type))
	if dialog == nil {
		fillMessagesMany(out, []*msg.UserMessage{}, []*msg.User{}, []*msg.Group{}, in.RequestID, da.OnComplete)
		return
	}

	// Prepare the the result before sending back to the client
	preSuccessCB := r.genGetHistoryCB(da.OnComplete, domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), req.MinID, req.MaxID, dialog.TopMessageID)

	// We are Offline/Disconnected
	if !r.SDK().NetCtrl().Connected() {
		messages, users, groups := repo.Messages.GetMessageHistory(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), req.MinID, req.MaxID, req.Limit)
		if len(messages) > 0 {
			pendingMessages := repo.PendingMessages.GetByPeer(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type))
			if len(pendingMessages) > 0 {
				messages = append(pendingMessages, messages...)
			}
			fillMessagesMany(out, messages, users, groups, in.RequestID, da.OnComplete)
			return
		}
	}

	// We are Online
	switch {
	case req.MinID == 0 && req.MaxID == 0:
		req.MaxID = dialog.TopMessageID
		fallthrough
	case req.MinID == 0 && req.MaxID != 0:
		b, bar := messageHole.GetLowerFilled(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), 0, req.MaxID)
		if !b {
			r.Log().Info("detected hole (With MaxID Only)",
				zap.Int64("MaxID", req.MaxID),
				zap.Int64("PeerID", req.Peer.ID),
				zap.Int64("TopMsgID", dialog.TopMessageID),
				zap.String("Holes", messageHole.PrintHole(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), 0)),
			)
			r.SDK().QueueCtrl().EnqueueCommand(da.ReplaceCompleteCB(preSuccessCB))
			return
		}
		messages, users, groups := repo.Messages.GetMessageHistory(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), bar.Min, bar.Max, req.Limit)
		fillMessagesMany(out, messages, users, groups, in.RequestID, preSuccessCB)
	case req.MinID != 0 && req.MaxID == 0:
		b, bar := messageHole.GetUpperFilled(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), 0, req.MinID)
		if !b {
			r.Log().Info("detected hole (With MinID Only)",
				zap.Int64("MinID", req.MinID),
				zap.Int64("PeerID", req.Peer.ID),
				zap.Int64("TopMsgID", dialog.TopMessageID),
				zap.String("Holes", messageHole.PrintHole(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), 0)),
			)
			r.SDK().QueueCtrl().EnqueueCommand(da.ReplaceCompleteCB(preSuccessCB))
			return
		}
		messages, users, groups := repo.Messages.GetMessageHistory(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), bar.Min, 0, req.Limit)
		fillMessagesMany(out, messages, users, groups, in.RequestID, preSuccessCB)
	default:
		b := messageHole.IsHole(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), 0, req.MinID, req.MaxID)
		if b {
			r.Log().Info("detected hole (With Min & Max)",
				zap.Int64("MinID", req.MinID),
				zap.Int64("MaxID", req.MaxID),
				zap.Int64("PeerID", req.Peer.ID),
				zap.Int64("TopMsgID", dialog.TopMessageID),
			)
			r.SDK().QueueCtrl().EnqueueCommand(da.ReplaceCompleteCB(preSuccessCB))
			return
		}
		messages, users, groups := repo.Messages.GetMessageHistory(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), req.MinID, req.MaxID, req.Limit)
		fillMessagesMany(out, messages, users, groups, in.RequestID, preSuccessCB)
	}
}
func fillMessagesMany(
	out *rony.MessageEnvelope, messages []*msg.UserMessage, users []*msg.User, groups []*msg.Group, requestID uint64, successCB domain.MessageHandler,
) {
	res := &msg.MessagesMany{
		Messages: messages,
		Users:    users,
		Groups:   groups,
	}

	out.RequestID = requestID
	out.Constructor = msg.C_MessagesMany
	out.Message, _ = res.Marshal()
	uiexec.ExecSuccessCB(successCB, out)
}
func (r *message) genGetHistoryCB(
	cb domain.MessageHandler, teamID, peerID int64, peerType int32, minID, maxID int64, topMessageID int64,
) domain.MessageHandler {
	return func(m *rony.MessageEnvelope) {
		pendingMessages := repo.PendingMessages.GetByPeer(teamID, peerID, peerType)
		switch m.Constructor {
		case msg.C_MessagesMany:
			x := &msg.MessagesMany{}
			err := x.Unmarshal(m.Message)
			r.Log().WarnOnErr("Error On Unmarshal MessagesMany", err)

			// 1st sort the received messages by id
			sort.Slice(x.Messages, func(i, j int) bool {
				return x.Messages[i].ID > x.Messages[j].ID
			})

			// Fill Messages Hole
			if msgCount := len(x.Messages); msgCount > 0 {
				switch {
				case minID == 0 && maxID != 0:
					messageHole.InsertFill(teamID, peerID, peerType, 0, x.Messages[msgCount-1].ID, maxID)
				case minID != 0 && maxID == 0:
					messageHole.InsertFill(teamID, peerID, peerType, 0, minID, x.Messages[0].ID)
				case minID == 0 && maxID == 0:
					messageHole.InsertFill(teamID, peerID, peerType, 0, x.Messages[msgCount-1].ID, x.Messages[0].ID)
				}
			}

			if len(pendingMessages) > 0 {
				if maxID == 0 || (len(x.Messages) > 0 && x.Messages[len(x.Messages)-1].ID == topMessageID) {
					x.Messages = append(pendingMessages, x.Messages...)
				}
			}

			m.Message, _ = x.Marshal()
		case rony.C_Error:
			r.Log().Warn("MessageModule received error on GetHistory", zap.Error(domain.ParseServerError(m.Message)))
		default:
		}

		// Call the actual success callback function
		cb(m)
	}
}

func (r *message) messagesGetMediaHistory(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.MessagesGetMediaHistory{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	// Load the dialog
	dialog, _ := repo.Dialogs.Get(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type))
	if dialog == nil {
		fillMessagesMany(out, []*msg.UserMessage{}, []*msg.User{}, []*msg.Group{}, in.RequestID, da.OnComplete)
		return
	}

	// We are Online
	if req.MaxID == 0 {
		req.MaxID = dialog.TopMessageID
	}

	// We are Offline/Disconnected
	if !r.SDK().NetCtrl().Connected() {
		messages, users, groups := repo.Messages.GetMediaMessageHistory(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), 0, req.MaxID, req.Limit, req.Cat)
		if len(messages) > 0 {
			fillMessagesMany(out, messages, users, groups, in.RequestID, da.OnComplete)
			return
		}
	}

	// Prepare the the result before sending back to the client
	preSuccessCB := r.genGetMediaHistoryCB(da.OnComplete, domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), req.MaxID, req.Cat)

	b, bar := messageHole.GetLowerFilled(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), req.Cat, req.MaxID)
	if !b {
		r.Log().Info("detected hole (With MaxID Only)",
			zap.Int64("MaxID", req.MaxID),
			zap.Int64("PeerID", req.Peer.ID),
			zap.String("PeerType", req.Peer.Type.String()),
			zap.String("Cat", req.Cat.String()),
			zap.Int64("TopMsgID", dialog.TopMessageID),
			zap.String("Holes", messageHole.PrintHole(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), req.Cat)),
		)
		r.SDK().QueueCtrl().EnqueueCommand(da.ReplaceCompleteCB(preSuccessCB))
		return
	}

	messages, users, groups := repo.Messages.GetMediaMessageHistory(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), 0, bar.Max, req.Limit, req.Cat)
	fillMessagesMany(out, messages, users, groups, in.RequestID, preSuccessCB)
}
func (r *message) genGetMediaHistoryCB(
	cb domain.MessageHandler, teamID, peerID int64, peerType int32, maxID int64, cat msg.MediaCategory,
) domain.MessageHandler {
	return func(m *rony.MessageEnvelope) {
		switch m.Constructor {
		case msg.C_MessagesMany:
			x := &msg.MessagesMany{}
			err := x.Unmarshal(m.Message)
			r.Log().WarnOnErr("Error On Unmarshal MessagesMany", err)

			// 1st sort the received messages by id
			sort.Slice(x.Messages, func(i, j int) bool {
				return x.Messages[i].ID > x.Messages[j].ID
			})

			// Fill Messages Hole
			if msgCount := len(x.Messages); msgCount > 0 {
				if maxID == 0 {
					messageHole.InsertFill(teamID, peerID, peerType, cat, x.Messages[msgCount-1].ID, x.Messages[0].ID)
				} else {
					messageHole.InsertFill(teamID, peerID, peerType, cat, x.Messages[msgCount-1].ID, maxID)
				}
			}

			m.Message, _ = x.Marshal()
		case rony.C_Error:
			r.Log().Warn("MessageModule received error on GetHistory", zap.Error(domain.ParseServerError(m.Message)))
		default:
		}

		// Call the actual success callback function
		cb(m)
	}
}

func (r *message) messagesDelete(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.MessagesDelete{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	// Get PendingMessage and cancel its requests
	pendingMessageIDs := make([]int64, 0, 4)
	for _, id := range req.MessageIDs {
		if id < 0 {
			pendingMessageIDs = append(pendingMessageIDs, id)
		}
	}
	if len(pendingMessageIDs) > 0 {
		for _, id := range pendingMessageIDs {
			pmsg, _ := repo.PendingMessages.GetByID(id)
			if pmsg == nil {
				return
			}
			if pmsg.FileID != 0 {
				r.SDK().FileCtrl().CancelUploadRequest(pmsg.FileID)
			}

			_ = repo.PendingMessages.Delete(id)

		}
	}

	// send the request to server
	r.SDK().QueueCtrl().EnqueueCommand(da)

}

func (r *message) messagesGet(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.MessagesGet{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}
	msgIDs := domain.MInt64B{}
	pMsgIDs := domain.MInt64B{}

	for _, v := range req.MessagesIDs {
		if v > 0 {
			msgIDs[v] = true
		} else {
			pMsgIDs[v] = true
		}
	}

	messages, _ := repo.Messages.GetMany(msgIDs.ToArray())
	messages = append(messages, repo.PendingMessages.GetMany(pMsgIDs.ToArray())...)

	mUsers := domain.MInt64B{}
	mUsers[req.Peer.ID] = true
	for _, m := range messages {
		mUsers[m.SenderID] = true
		mUsers[m.FwdSenderID] = true
		actUserIDs := domain.ExtractActionUserIDs(m.MessageAction, m.MessageActionData)
		for _, id := range actUserIDs {
			mUsers[id] = true
		}
	}
	users, _ := repo.Users.GetMany(mUsers.ToArray())

	// if db already had all users
	if len(messages) == (len(msgIDs)+len(pMsgIDs)) && len(users) > 0 {
		res := &msg.MessagesMany{
			Messages: messages,
			Users:    users,
		}

		out.Constructor = msg.C_MessagesMany
		out.Message, _ = res.Marshal()
		uiexec.ExecSuccessCB(da.OnComplete, out)
		return
	}

	// WebsocketSend the request to the server
	r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *message) messagesClearHistory(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.MessagesClearHistory{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	if req.MaxID == 0 {
		d, err := repo.Dialogs.Get(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type))
		if err != nil {
			out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
			da.OnComplete(out)
			return
		}
		req.MaxID = d.TopMessageID
	}

	err := repo.Messages.ClearHistory(r.SDK().GetConnInfo().PickupUserID(), domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), req.MaxID)
	r.Log().WarnOnErr("got error on clear history", err,
		zap.Int64("PeerID", req.Peer.ID),
		zap.Int64("TeamID", domain.GetTeamID(in)),
	)

	if req.Delete {
		err = repo.Dialogs.Delete(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type))
		r.Log().WarnOnErr("got error on deleting dialogs", err,
			zap.Int64("PeerID", req.Peer.ID),
			zap.Int64("TeamID", domain.GetTeamID(in)),
		)
	}

	// send the request to server
	r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *message) messagesReadContents(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.MessagesReadContents{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	repo.Messages.SetContentRead(req.Peer.ID, int32(req.Peer.Type), req.MessageIDs)

	// send the request to server
	r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *message) messagesSaveDraft(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.MessagesSaveDraft{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	dialog, _ := repo.Dialogs.Get(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type))
	if dialog != nil {
		draftMessage := msg.DraftMessage{
			Body:     req.Body,
			Entities: req.Entities,
			PeerID:   req.Peer.ID,
			PeerType: int32(req.Peer.Type),
			Date:     time.Now().Unix(),
			ReplyTo:  req.ReplyTo,
		}

		dialog.Draft = &draftMessage

		repo.Dialogs.Save(dialog)
	}

	// send the request to server
	r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *message) messagesClearDraft(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.MessagesClearDraft{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	dialog, _ := repo.Dialogs.Get(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type))
	if dialog != nil {
		dialog.Draft = nil
		repo.Dialogs.Save(dialog)
	}

	// send the request to server
	r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *message) messagesTogglePin(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.MessagesTogglePin{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	err := repo.Dialogs.UpdatePinMessageID(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), req.MessageID)
	r.Log().ErrorOnErr("MessagesTogglePin", err)

	r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *message) messagesSendReaction(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.MessagesSendReaction{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	err := repo.Reactions.IncrementReactionUseCount(req.Reaction, 1)
	r.Log().ErrorOnErr("messagesSendReaction", err)

	r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *message) messagesDeleteReaction(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.MessagesDeleteReaction{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	for _, react := range req.Reactions {
		err := repo.Reactions.IncrementReactionUseCount(react, -1)
		r.Log().ErrorOnErr("messagesDeleteReaction", err)
	}

	r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *message) messagesToggleDialogPin(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.MessagesToggleDialogPin{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	dialog, _ := repo.Dialogs.Get(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type))
	if dialog != nil {
		dialog.Pinned = req.Pin
		_ = repo.Dialogs.Save(dialog)
	}

	// send the request to server
	r.SDK().QueueCtrl().EnqueueCommand(da)

}

func (r *message) clientGetMediaHistory(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.ClientGetMediaHistory{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	messages, users, groups := repo.Messages.GetMediaMessageHistory(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), req.MinID, req.MaxID, req.Limit, req.Cat)
	if len(messages) > 0 {
		res := &msg.MessagesMany{
			Messages: messages,
			Users:    users,
			Groups:   groups,
		}

		out.RequestID = in.RequestID
		out.Constructor = msg.C_MessagesMany
		out.Message, _ = res.Marshal()
		uiexec.ExecSuccessCB(da.OnComplete, out)
		return
	}
}

func (r *message) clientSendMessageMedia(in, out *rony.MessageEnvelope, da request.Callback) {
	reqMedia := &msg.ClientSendMessageMedia{}
	if err := reqMedia.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		uiexec.ExecSuccessCB(da.OnComplete, out)
		return
	}

	// support IOS file path
	if strings.HasPrefix(reqMedia.FilePath, "file://") {
		reqMedia.FilePath = reqMedia.FilePath[7:]
	}
	if strings.HasPrefix(reqMedia.ThumbFilePath, "file://") {
		reqMedia.ThumbFilePath = reqMedia.ThumbFilePath[7:]
	}

	// 1. insert into pending messages, id is negative nano timestamp and save RandomID too : Done
	fileID := domain.SequentialUniqueID()
	msgID := -fileID
	thumbID := int64(0)
	reqMedia.FileUploadID = fmt.Sprintf("%d", fileID)
	reqMedia.FileID = fileID
	if reqMedia.ThumbFilePath != "" {
		thumbID = domain.RandomInt63()
		reqMedia.ThumbID = thumbID
		reqMedia.ThumbUploadID = fmt.Sprintf("%d", thumbID)
	}

	checkSha256 := true
	switch reqMedia.MediaType {
	case msg.InputMediaType_InputMediaTypeUploadedDocument:
		for _, attr := range reqMedia.Attributes {
			if attr.Type == msg.DocumentAttributeType_AttributeTypeAudio {
				x := &msg.DocumentAttributeAudio{}
				_ = x.Unmarshal(attr.Data)
				if x.Voice {
					checkSha256 = false
				}
			}
		}
	default:
		panic("Invalid MediaInputType")
	}

	h, _ := domain.CalculateSha256(reqMedia.FilePath)
	pendingMessage, err := repo.PendingMessages.SaveClientMessageMedia(
		domain.GetTeamID(in), domain.GetTeamAccess(in), msgID, r.SDK().GetConnInfo().PickupUserID(), fileID, fileID, thumbID, reqMedia, h,
	)
	if err != nil {
		e := &rony.Error{
			Code:  "n/a",
			Items: "Failed to save to pendingMessages : " + err.Error(),
		}
		out.Fill(out.RequestID, rony.C_Error, e)
		uiexec.ExecSuccessCB(da.OnComplete, out)
		return
	}

	// 3. return to CallBack with pending message data : Done
	out.Fill(out.RequestID, msg.C_ClientPendingMessage, pendingMessage)

	// 4. Start the upload process
	r.SDK().FileCtrl().UploadMessageDocument(pendingMessage.ID, reqMedia.FilePath, reqMedia.ThumbFilePath, fileID, thumbID, h, pendingMessage.PeerID, checkSha256)

	uiexec.ExecSuccessCB(da.OnComplete, out)
}

func (r *message) clientGetFrequentReactions(in, out *rony.MessageEnvelope, da request.Callback) {
	reactions := domain.SysConfig.Reactions
	r.Log().Info("Reactions", zap.Int("ReactionsCount", len(reactions)))

	useCountsMap := make(map[string]uint32, len(reactions))

	for _, r := range reactions {
		useCount, _ := repo.Reactions.GetReactionUseCount(r)
		useCountsMap[r] = useCount
	}

	sort.Slice(reactions, func(i, j int) bool {
		return useCountsMap[reactions[i]] > useCountsMap[reactions[j]]
	})

	res := &msg.ClientFrequentReactions{
		Reactions: reactions,
	}
	out.Fill(out.RequestID, msg.C_ClientFrequentReactions, res)
	da.OnComplete(out)
}

func (r *message) clientGetCachedMedia(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.ClientGetCachedMedia{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	res := repo.Files.GetCachedMedia(domain.GetTeamID(in))

	out.Fill(in.RequestID, msg.C_ClientCachedMediaInfo, res)
	uiexec.ExecSuccessCB(da.OnComplete, out)
}

func (r *message) clientClearCachedMedia(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.ClientClearCachedMedia{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	if req.Peer != nil {
		repo.Files.DeleteCachedMediaByPeer(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), req.MediaTypes)
	} else if len(req.MediaTypes) > 0 {
		repo.Files.DeleteCachedMediaByMediaType(domain.GetTeamID(in), req.MediaTypes)
	} else {
		repo.Files.ClearCache()
	}

	res := &msg.Bool{
		Result: true,
	}
	out.Fill(in.RequestID, msg.C_Bool, res)
	uiexec.ExecSuccessCB(da.OnComplete, out)
}

func (r *message) clientGetLastBotKeyboard(in, out *rony.MessageEnvelope, da request.Callback) {
	req := &msg.ClientGetLastBotKeyboard{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		da.OnComplete(out)
		return
	}

	lastKeyboardMsg, _ := repo.Messages.GetLastBotKeyboard(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type))

	if lastKeyboardMsg == nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: "message not found"})
		da.OnComplete(out)
		return
	}

	out.Fill(in.RequestID, msg.C_UserMessage, lastKeyboardMsg)
	uiexec.ExecSuccessCB(da.OnComplete, out)
}
