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
	"github.com/dustin/go-humanize"
	"github.com/olekukonko/tablewriter"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/errors"
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

func (r *message) messagesGetDialogs(da request.Callback) {
	req := &msg.MessagesGetDialogs{}
	if err := da.RequestData(req); err != nil {
		return
	}
	res := &msg.MessagesDialogs{}
	res.Dialogs, _ = repo.Dialogs.List(da.TeamID(), req.Offset, req.Limit)
	res.Count = repo.Dialogs.CountDialogs(da.TeamID())

	// If the localDB had no data send the request to server
	if len(res.Dialogs) == 0 {
		res.UpdateID = r.SDK().SyncCtrl().GetUpdateID()
		da.Response(msg.C_MessagesDialogs, res)
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

	da.Response(msg.C_MessagesDialogs, res)
}

func (r *message) messagesGetDialog(da request.Callback) {
	req := &msg.MessagesGetDialog{}
	if err := da.RequestData(req); err != nil {
		return
	}

	res, err := repo.Dialogs.Get(da.TeamID(), req.Peer.ID, int32(req.Peer.Type))

	// if the localDB had no data send the request to server
	if err != nil {
		r.Log().Warn("got error on repo GetDialog", zap.Error(err), zap.Int64("PeerID", req.Peer.ID))
		r.SDK().QueueCtrl().EnqueueCommand(da)
		return
	}

	da.Response(msg.C_Dialog, res)
}

func (r *message) messagesSend(da request.Callback) {
	req := &msg.MessagesSend{}
	if err := da.RequestData(req); err != nil {
		return
	}

	// do not allow empty message
	if strings.TrimSpace(req.Body) == "" {
		e := &rony.Error{
			Code:  "n/a",
			Items: "empty message is not allowed",
		}
		da.Response(rony.C_Error, e)
		return
	}

	// for saved messages we have special cases for debugging purpose
	if req.Peer.ID == r.SDK().GetConnInfo().PickupUserID() {
		r.handleDebugActions(req.Body)
	}

	// this will be used as next requestID
	req.RandomID = domain.SequentialUniqueID()
	msgID := -req.RandomID
	res, err := repo.PendingMessages.Save(da.TeamID(), da.TeamAccess(), msgID, r.SDK().GetConnInfo().PickupUserID(), req)
	if err != nil {
		e := &rony.Error{
			Code:  "n/a",
			Items: "Failed to save to pendingMessages : " + err.Error(),
		}
		da.Response(rony.C_Error, e)
		return
	}

	// using req randomID as requestID later in queue processing and network controller messageHandler
	da.Discard()
	r.SDK().QueueCtrl().EnqueueCommand(
		request.NewCallback(
			da.TeamID(), da.TeamAccess(), uint64(req.RandomID), msg.C_MessagesSend, req,
			nil, nil, nil, da.UI(), da.Flags(), da.Timeout(),
		),
	)

	// return to CallBack with pending message data : Done
	// later when queue got processed and server returned response we should check if the requestID
	// exist in pendingTable we remove it and insert new message with new id to message table
	// invoke new OnUpdate with new proto buffer to inform ui that pending message got delivered
	da.Response(msg.C_ClientPendingMessage, res)
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

	r.messagesSend(
		request.NewCallback(
			0, 0, domain.NextRequestID(), msg.C_MessagesSend, req,
			nil, nil, nil, false, 0, 0,
		),
	)
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
	r.clientSendMessageMedia(
		request.NewCallback(
			0, 0, domain.NextRequestID(), msg.C_MessagesSend, req,
			nil, nil, nil, false, 0, 0,
		),
	)
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
		if err != nil {
			return nil
		}
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

func (r *message) messagesSendMedia(da request.Callback) {
	req := &msg.MessagesSendMedia{}
	if err := da.RequestData(req); err != nil {
		return
	}

	switch req.MediaType {
	case msg.InputMediaType_InputMediaTypeEmpty:
		// sending text messages MUST be handled by MessagesSendMedia
		da.Response(rony.C_Error, errors.New("00", "USE_MessagesSendMedia"))
		return
	case msg.InputMediaType_InputMediaTypeUploadedDocument:
		// sending uploaded document types MUST be handled by ClientSendMessageMedia
		da.Response(rony.C_Error, errors.New("00", "USE_ClientSendMessageMedia"))
		return
	case msg.InputMediaType_InputMediaTypeContact, msg.InputMediaType_InputMediaTypeGeoLocation,
		msg.InputMediaType_InputMediaTypeDocument, msg.InputMediaType_InputMediaTypeMessageDocument:
		// This will be used as next requestID
		// Insert into pending messages, id is negative nano timestamp and save RandomID too
		req.RandomID = domain.SequentialUniqueID()
		dbID := -req.RandomID

		res, err := repo.PendingMessages.SaveMessageMedia(da.TeamID(), da.TeamAccess(), dbID, r.SDK().GetConnInfo().PickupUserID(), req)
		if err != nil {
			e := &rony.Error{
				Code:  "n/a",
				Items: "Failed to save to pendingMessages : " + err.Error(),
			}
			da.Response(rony.C_Error, e)
			return
		}
		// return temporary response to the UI until UpdateMessageID/UpdateNewMessage arrived.
		da.Response(msg.C_ClientPendingMessage, res)
	}

	da.Discard()
	r.SDK().QueueCtrl().EnqueueCommand(
		request.NewCallback(
			da.TeamID(), da.TeamAccess(), uint64(req.RandomID), msg.C_MessagesSendMedia, req,
			nil, nil, nil, da.UI(), da.Flags(), da.Timeout(),
		),
	)

}

func (r *message) messagesReadHistory(da request.Callback) {
	req := &msg.MessagesReadHistory{}
	if err := da.RequestData(req); err != nil {
		return
	}

	dialog, _ := repo.Dialogs.Get(da.TeamID(), req.Peer.ID, int32(req.Peer.Type))
	if dialog == nil {
		return
	}
	if dialog.ReadInboxMaxID > req.MaxID {
		return
	}

	// update read inbox max id
	err := repo.Dialogs.UpdateReadInboxMaxID(r.SDK().GetConnInfo().PickupUserID(), da.TeamID(), req.Peer.ID, int32(req.Peer.Type), req.MaxID)
	r.Log().WarnOnErr("could not update read inbox max id", err)

	// send the request to server
	r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *message) messagesGetHistory(da request.Callback) {
	req := &msg.MessagesGetHistory{}
	if err := da.RequestData(req); err != nil {
		return
	}

	// Load the dialog
	dialog, _ := repo.Dialogs.Get(da.TeamID(), req.Peer.ID, int32(req.Peer.Type))
	if dialog == nil {
		fillMessagesMany(da, []*msg.UserMessage{}, []*msg.User{}, []*msg.Group{})
		return
	}

	// Prepare the the result before sending back to the client
	da.ReplaceCompleteCB(r.genGetHistoryCB(da.OnComplete, da.TeamID(), req.Peer.ID, int32(req.Peer.Type), req.MinID, req.MaxID, dialog.TopMessageID))
	// We are Offline/Disconnected
	if !r.SDK().NetCtrl().Connected() {
		messages, users, groups := repo.Messages.GetMessageHistory(da.TeamID(), req.Peer.ID, int32(req.Peer.Type), req.MinID, req.MaxID, req.Limit)
		if len(messages) > 0 {
			pendingMessages := repo.PendingMessages.GetByPeer(da.TeamID(), req.Peer.ID, int32(req.Peer.Type))
			if len(pendingMessages) > 0 {
				messages = append(pendingMessages, messages...)
			}
			fillMessagesMany(da, messages, users, groups)
			return
		}
	}

	// We are Online
	switch {
	case req.MinID == 0 && req.MaxID == 0:
		req.MaxID = dialog.TopMessageID
		fallthrough
	case req.MinID == 0 && req.MaxID != 0:
		b, bar := messageHole.GetLowerFilled(da.TeamID(), req.Peer.ID, int32(req.Peer.Type), 0, req.MaxID)
		if !b {
			r.Log().Info("detected hole (With MaxID Only)",
				zap.Int64("MaxID", req.MaxID),
				zap.Int64("PeerID", req.Peer.ID),
				zap.Int64("TopMsgID", dialog.TopMessageID),
				zap.String("Holes", messageHole.PrintHole(da.TeamID(), req.Peer.ID, int32(req.Peer.Type), 0)),
			)
			r.SDK().QueueCtrl().EnqueueCommand(da)
			return
		}
		messages, users, groups := repo.Messages.GetMessageHistory(da.TeamID(), req.Peer.ID, int32(req.Peer.Type), bar.Min, bar.Max, req.Limit)
		fillMessagesMany(da, messages, users, groups)
	case req.MinID != 0 && req.MaxID == 0:
		b, bar := messageHole.GetUpperFilled(da.TeamID(), req.Peer.ID, int32(req.Peer.Type), 0, req.MinID)
		if !b {
			r.Log().Info("detected hole (With MinID Only)",
				zap.Int64("MinID", req.MinID),
				zap.Int64("PeerID", req.Peer.ID),
				zap.Int64("TopMsgID", dialog.TopMessageID),
				zap.String("Holes", messageHole.PrintHole(da.TeamID(), req.Peer.ID, int32(req.Peer.Type), 0)),
			)
			r.SDK().QueueCtrl().EnqueueCommand(da)
			return
		}
		messages, users, groups := repo.Messages.GetMessageHistory(da.TeamID(), req.Peer.ID, int32(req.Peer.Type), bar.Min, 0, req.Limit)
		fillMessagesMany(da, messages, users, groups)
	default:
		b := messageHole.IsHole(da.TeamID(), req.Peer.ID, int32(req.Peer.Type), 0, req.MinID, req.MaxID)
		if b {
			r.Log().Info("detected hole (With Min & Max)",
				zap.Int64("MinID", req.MinID),
				zap.Int64("MaxID", req.MaxID),
				zap.Int64("PeerID", req.Peer.ID),
				zap.Int64("TopMsgID", dialog.TopMessageID),
			)
			r.SDK().QueueCtrl().EnqueueCommand(da)
			return
		}
		messages, users, groups := repo.Messages.GetMessageHistory(da.TeamID(), req.Peer.ID, int32(req.Peer.Type), req.MinID, req.MaxID, req.Limit)
		fillMessagesMany(da, messages, users, groups)
	}
}
func fillMessagesMany(
	da request.Callback, messages []*msg.UserMessage, users []*msg.User, groups []*msg.Group,
) {
	da.Response(msg.C_MessagesMany,
		&msg.MessagesMany{
			Messages: messages,
			Users:    users,
			Groups:   groups,
		},
	)
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

			// sort the received messages by id
			sort.Slice(x.Messages, func(i, j int) bool {
				return x.Messages[i].ID > x.Messages[j].ID
			})

			// fill messages hole based on the server response
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
			r.Log().Warn("received error on GetHistory", zap.Error(domain.ParseServerError(m.Message)))
		default:
		}

		// Call the actual success callback function
		cb(m)
	}
}

func (r *message) messagesGetMediaHistory(da request.Callback) {
	req := &msg.MessagesGetMediaHistory{}
	if err := da.RequestData(req); err != nil {
		return
	}

	// Load the dialog
	dialog, _ := repo.Dialogs.Get(da.TeamID(), req.Peer.ID, int32(req.Peer.Type))
	if dialog == nil {
		fillMessagesMany(da, []*msg.UserMessage{}, []*msg.User{}, []*msg.Group{})
		return
	}

	// We are Online
	if req.MaxID == 0 {
		req.MaxID = dialog.TopMessageID
	}

	// We are Offline/Disconnected
	if !r.SDK().NetCtrl().Connected() {
		messages, users, groups := repo.Messages.GetMediaMessageHistory(da.TeamID(), req.Peer.ID, int32(req.Peer.Type), 0, req.MaxID, req.Limit, req.Cat)
		if len(messages) > 0 {
			fillMessagesMany(da, messages, users, groups)
			return
		}
	}

	// Prepare the the result before sending back to the client
	da.ReplaceCompleteCB(r.genGetMediaHistoryCB(da.OnComplete, da.TeamID(), req.Peer.ID, int32(req.Peer.Type), req.MaxID, req.Cat))
	b, bar := messageHole.GetLowerFilled(da.TeamID(), req.Peer.ID, int32(req.Peer.Type), req.Cat, req.MaxID)
	if !b {
		r.Log().Info("detected hole (With MaxID Only)",
			zap.Int64("MaxID", req.MaxID),
			zap.Int64("PeerID", req.Peer.ID),
			zap.String("PeerType", req.Peer.Type.String()),
			zap.String("Cat", req.Cat.String()),
			zap.Int64("TopMsgID", dialog.TopMessageID),
			zap.String("Holes", messageHole.PrintHole(da.TeamID(), req.Peer.ID, int32(req.Peer.Type), req.Cat)),
		)
		r.SDK().QueueCtrl().EnqueueCommand(da)
		return
	}

	messages, users, groups := repo.Messages.GetMediaMessageHistory(da.TeamID(), req.Peer.ID, int32(req.Peer.Type), 0, bar.Max, req.Limit, req.Cat)
	fillMessagesMany(da, messages, users, groups)
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

			// sort the received messages by id
			sort.Slice(x.Messages, func(i, j int) bool {
				return x.Messages[i].ID > x.Messages[j].ID
			})

			// fill messages hole based on the server response
			if msgCount := len(x.Messages); msgCount > 0 {
				if maxID == 0 {
					messageHole.InsertFill(teamID, peerID, peerType, cat, x.Messages[msgCount-1].ID, x.Messages[0].ID)
				} else {
					messageHole.InsertFill(teamID, peerID, peerType, cat, x.Messages[msgCount-1].ID, maxID)
				}
			}

			m.Message, _ = x.Marshal()
		case rony.C_Error:
			r.Log().Warn("received error on GetMediaHistory", zap.Error(domain.ParseServerError(m.Message)))
		default:
		}

		// Call the actual success callback function
		cb(m)
	}
}

func (r *message) messagesDelete(da request.Callback) {
	req := &msg.MessagesDelete{}
	if err := da.RequestData(req); err != nil {
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
			pm, _ := repo.PendingMessages.GetByID(id)
			if pm == nil {
				continue
			}
			if pm.FileID != 0 {
				r.SDK().FileCtrl().CancelUploadRequest(pm.FileID)
			}

			_ = repo.PendingMessages.Delete(id)
		}
	}

	// send the request to server
	r.SDK().QueueCtrl().EnqueueCommand(da)

}

func (r *message) messagesGet(da request.Callback) {
	req := &msg.MessagesGet{}
	if err := da.RequestData(req); err != nil {
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
		da.Response(msg.C_MessagesDialogs, res)
		return
	}

	// WebsocketSend the request to the server
	r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *message) messagesClearHistory(da request.Callback) {
	req := &msg.MessagesClearHistory{}
	if err := da.RequestData(req); err != nil {
		return
	}

	if req.MaxID == 0 {
		d, err := repo.Dialogs.Get(da.TeamID(), req.Peer.ID, int32(req.Peer.Type))
		if err != nil {
			da.Response(rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
			return
		}
		req.MaxID = d.TopMessageID
	}

	err := repo.Messages.ClearHistory(r.SDK().GetConnInfo().PickupUserID(), da.TeamID(), req.Peer.ID, int32(req.Peer.Type), req.MaxID)
	r.Log().WarnOnErr("got error on clear history", err,
		zap.Int64("PeerID", req.Peer.ID),
		zap.Int64("TeamID", da.TeamID()),
	)

	if req.Delete {
		err = repo.Dialogs.Delete(da.TeamID(), req.Peer.ID, int32(req.Peer.Type))
		r.Log().WarnOnErr("got error on deleting dialogs", err,
			zap.Int64("PeerID", req.Peer.ID),
			zap.Int64("TeamID", da.TeamID()),
		)
	}

	// send the request to server
	r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *message) messagesReadContents(da request.Callback) {
	req := &msg.MessagesReadContents{}
	if err := da.RequestData(req); err != nil {
		return
	}

	_ = repo.Messages.SetContentRead(req.Peer.ID, int32(req.Peer.Type), req.MessageIDs)

	// send the request to server
	r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *message) messagesSaveDraft(da request.Callback) {
	req := &msg.MessagesSaveDraft{}
	if err := da.RequestData(req); err != nil {
		return
	}

	dialog, _ := repo.Dialogs.Get(da.TeamID(), req.Peer.ID, int32(req.Peer.Type))
	if dialog != nil {
		dialog.Draft = &msg.DraftMessage{
			Body:     req.Body,
			Entities: req.Entities,
			PeerID:   req.Peer.ID,
			PeerType: int32(req.Peer.Type),
			Date:     tools.TimeUnix(),
			ReplyTo:  req.ReplyTo,
		}
		_ = repo.Dialogs.Save(dialog)
	}

	// send the request to server
	r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *message) messagesClearDraft(da request.Callback) {
	req := &msg.MessagesClearDraft{}
	if err := da.RequestData(req); err != nil {
		return
	}

	dialog, _ := repo.Dialogs.Get(da.TeamID(), req.Peer.ID, int32(req.Peer.Type))
	if dialog != nil {
		dialog.Draft = nil
		_ = repo.Dialogs.Save(dialog)
	}

	// send the request to server
	r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *message) messagesTogglePin(da request.Callback) {
	req := &msg.MessagesTogglePin{}
	if err := da.RequestData(req); err != nil {
		return
	}

	err := repo.Dialogs.UpdatePinMessageID(da.TeamID(), req.Peer.ID, int32(req.Peer.Type), req.MessageID)
	r.Log().ErrorOnErr("go error on toggle pin dialog", err)

	r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *message) messagesSendReaction(da request.Callback) {
	req := &msg.MessagesSendReaction{}
	if err := da.RequestData(req); err != nil {
		return
	}

	err := repo.Reactions.IncrementReactionUseCount(req.Reaction, 1)
	r.Log().ErrorOnErr("got error on send message reaction", err)

	r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *message) messagesDeleteReaction(da request.Callback) {
	req := &msg.MessagesDeleteReaction{}
	if err := da.RequestData(req); err != nil {
		return
	}

	for _, react := range req.Reactions {
		err := repo.Reactions.IncrementReactionUseCount(react, -1)
		r.Log().ErrorOnErr("got error on deleting message reaction", err)
	}

	r.SDK().QueueCtrl().EnqueueCommand(da)
}

func (r *message) messagesToggleDialogPin(da request.Callback) {
	req := &msg.MessagesToggleDialogPin{}
	if err := da.RequestData(req); err != nil {
		return
	}

	dialog, _ := repo.Dialogs.Get(da.TeamID(), req.Peer.ID, int32(req.Peer.Type))
	if dialog != nil {
		dialog.Pinned = req.Pin
		_ = repo.Dialogs.Save(dialog)
	}

	// send the request to server
	r.SDK().QueueCtrl().EnqueueCommand(da)

}

func (r *message) clientGetMediaHistory(da request.Callback) {
	req := &msg.ClientGetMediaHistory{}
	if err := da.RequestData(req); err != nil {
		return
	}

	messages, users, groups := repo.Messages.GetMediaMessageHistory(da.TeamID(), req.Peer.ID, int32(req.Peer.Type), req.MinID, req.MaxID, req.Limit, req.Cat)
	if len(messages) > 0 {
		da.Response(msg.C_MessagesMany,
			&msg.MessagesMany{
				Messages: messages,
				Users:    users,
				Groups:   groups,
			},
		)
		return
	}
}

func (r *message) clientSendMessageMedia(da request.Callback) {
	reqMedia := &msg.ClientSendMessageMedia{}
	if err := da.RequestData(reqMedia); err != nil {
		return
	}

	// support IOS file path
	reqMedia.FilePath = strings.TrimPrefix(reqMedia.FilePath, "file://")
	reqMedia.ThumbFilePath = strings.TrimPrefix(reqMedia.ThumbFilePath, "file://")

	// insert into pending messages, id is negative nano timestamp and save RandomID too : Done
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
		da.TeamID(), da.TeamAccess(), msgID, r.SDK().GetConnInfo().PickupUserID(), fileID, fileID, thumbID, reqMedia, h,
	)
	if err != nil {
		e := &rony.Error{
			Code:  "n/a",
			Items: "Failed to save to pendingMessages : " + err.Error(),
		}
		da.Response(rony.C_Error, e)
		return
	}

	// start the upload process
	r.SDK().FileCtrl().UploadMessageDocument(pendingMessage.ID, reqMedia.FilePath, reqMedia.ThumbFilePath, fileID, thumbID, h, pendingMessage.PeerID, checkSha256)

	da.Response(msg.C_ClientPendingMessage, pendingMessage)
}

func (r *message) clientGetFrequentReactions(da request.Callback) {
	reactions := domain.SysConfig.Reactions

	useCountsMap := make(map[string]uint32, len(reactions))

	for _, r := range reactions {
		useCount, _ := repo.Reactions.GetReactionUseCount(r)
		useCountsMap[r] = useCount
	}

	sort.Slice(reactions, func(i, j int) bool {
		return useCountsMap[reactions[i]] > useCountsMap[reactions[j]]
	})

	da.Response(msg.C_ClientFrequentReactions,
		&msg.ClientFrequentReactions{
			Reactions: reactions,
		},
	)
}

func (r *message) clientGetCachedMedia(da request.Callback) {
	req := &msg.ClientGetCachedMedia{}
	if err := da.RequestData(req); err != nil {
		return
	}

	res := repo.Files.GetCachedMedia(da.TeamID())
	da.Response(msg.C_ClientCachedMediaInfo, res)
}

func (r *message) clientClearCachedMedia(da request.Callback) {
	req := &msg.ClientClearCachedMedia{}
	if err := da.RequestData(req); err != nil {
		return
	}

	if req.Peer != nil {
		repo.Files.DeleteCachedMediaByPeer(da.TeamID(), req.Peer.ID, int32(req.Peer.Type), req.MediaTypes)
	} else if len(req.MediaTypes) > 0 {
		repo.Files.DeleteCachedMediaByMediaType(da.TeamID(), req.MediaTypes)
	} else {
		repo.Files.ClearCache()
	}

	da.Response(msg.C_Bool, &msg.Bool{Result: true})
}

func (r *message) clientGetLastBotKeyboard(da request.Callback) {
	req := &msg.ClientGetLastBotKeyboard{}
	if err := da.RequestData(req); err != nil {
		return
	}

	lastKeyboardMsg, _ := repo.Messages.GetLastBotKeyboard(da.TeamID(), req.Peer.ID, int32(req.Peer.Type))
	if lastKeyboardMsg == nil {
		da.Response(rony.C_Error, &rony.Error{Code: "00", Items: "message not found"})
		return
	}

	da.Response(msg.C_UserMessage, lastKeyboardMsg)
}
