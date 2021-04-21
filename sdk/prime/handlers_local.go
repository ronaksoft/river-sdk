package riversdk

import (
	"encoding/json"
	"fmt"
	messageHole "git.ronaksoft.com/river/sdk/internal/message_hole"
	"git.ronaksoft.com/river/sdk/internal/uiexec"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/registry"
	"sort"
	"strings"
	"sync"
	"time"

	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"go.uber.org/zap"
)

func (r *River) messagesGetDialogs(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.MessagesGetDialogs{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}
	res := &msg.MessagesDialogs{}
	res.Dialogs = repo.Dialogs.List(domain.GetTeamID(in), req.Offset, req.Limit)
	res.Count = repo.Dialogs.CountDialogs(domain.GetTeamID(in))

	// If the localDB had no data send the request to server
	if len(res.Dialogs) == 0 {
		res.UpdateID = r.syncCtrl.GetUpdateID()
		out.Constructor = msg.C_MessagesDialogs
		buff, err := res.Marshal()
		logs.ErrorOnErr("River got error on marshal MessagesDialogs", err)
		out.Message = buff
		uiexec.ExecSuccessCB(successCB, out)
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
	res.Messages = repo.Messages.GetMany(mMessages.ToArray())
	if len(res.Messages) != len(mMessages) {
		logs.Warn("River found unmatched dialog messages", zap.Int("Got", len(res.Messages)), zap.Int("Need", len(mMessages)))
		waitGroup := &sync.WaitGroup{}
		waitGroup.Add(1)
		r.syncCtrl.GetAllDialogs(waitGroup, domain.GetTeamID(in), domain.GetTeamAccess(in), 0, 100)
		for msgID := range mMessages {
			found := false
			for _, m := range res.Messages {
				if m.ID == msgID {
					found = true
					break
				}
			}
			if !found {
				logs.Warn("missed message", zap.Int64("MsgID", msgID))
			}
		}
		waitGroup.Wait()
		logs.Error("River re-synced dialogs")
	}

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
		logs.Warn("River found unmatched dialog groups", zap.Int("Got", len(res.Groups)), zap.Int("Need", len(mGroups)))
		for groupID := range mGroups {
			found := false
			for _, g := range res.Groups {
				if g.ID == groupID {
					found = true
					break
				}
			}
			if !found {
				logs.Warn("missed group", zap.Int64("GroupID", groupID))
			}
		}
	}
	res.Users, _ = repo.Users.GetMany(mUsers.ToArray())
	if len(res.Users) != len(mUsers) {
		logs.Warn("River found unmatched dialog users", zap.Int("Got", len(res.Users)), zap.Int("Need", len(mUsers)))
		for userID := range mUsers {
			found := false
			for _, g := range res.Users {
				if g.ID == userID {
					found = true
					break
				}
			}
			if !found {
				logs.Warn("missed user", zap.Int64("UserID", userID))
			}
		}
	}

	out.Constructor = msg.C_MessagesDialogs
	buff, err := res.Marshal()
	logs.ErrorOnErr("River got error on marshal MessagesDialogs", err)
	out.Message = buff
	uiexec.ExecSuccessCB(successCB, out)
}

func (r *River) messagesGetDialog(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.MessagesGetDialog{}
	err := req.Unmarshal(in.Message)
	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}
	res := &msg.Dialog{}
	res, err = repo.Dialogs.Get(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type))

	// if the localDB had no data send the request to server
	if err != nil {
		logs.Warn("We got error on repo GetDialog", zap.Error(err), zap.Int64("PeerID", req.Peer.ID))
		r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
		return
	}

	out.Constructor = msg.C_Dialog
	out.Message, _ = res.Marshal()

	uiexec.ExecSuccessCB(successCB, out)
}

func (r *River) messagesSend(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.MessagesSend{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	// do not allow empty message
	if strings.TrimSpace(req.Body) == "" {
		e := &rony.Error{
			Code:  "n/a",
			Items: "empty message is not allowed",
		}
		out.Fill(out.RequestID, rony.C_Error, e)
		uiexec.ExecSuccessCB(successCB, out)
		return
	}

	if req.Peer.ID == r.ConnInfo.UserID {
		r.HandleDebugActions(req.Body)
	}

	// this will be used as next requestID
	req.RandomID = domain.SequentialUniqueID()
	msgID := -req.RandomID
	res, err := repo.PendingMessages.Save(domain.GetTeamID(in), domain.GetTeamAccess(in), msgID, r.ConnInfo.UserID, req)
	if err != nil {
		e := &rony.Error{
			Code:  "n/a",
			Items: "Failed to save to pendingMessages : " + err.Error(),
		}
		out.Fill(out.RequestID, rony.C_Error, e)
		uiexec.ExecSuccessCB(successCB, out)
		return
	}
	// 2. add to queue [ looks like there is general queue to send messages ] : Done
	requestBytes, _ := req.Marshal()

	// using req randomID as requestID later in queue processing and network controller messageHandler
	r.queueCtrl.EnqueueCommand(
		&rony.MessageEnvelope{
			Header:      in.Header,
			Constructor: msg.C_MessagesSend,
			RequestID:   uint64(req.RandomID),
			Message:     requestBytes,
		},
		timeoutCB, successCB, true,
	)

	// 3. return to CallBack with pending message data : Done
	out.Constructor = msg.C_ClientPendingMessage
	out.Message, _ = res.Marshal()

	// 4. later when queue got processed and server returned response we should check if the requestID
	//   exist in pendingTable we remove it and insert new message with new id to message table
	//   invoke new OnUpdate with new proto buffer to inform ui that pending message got delivered
	uiexec.ExecSuccessCB(successCB, out)
}

func (r *River) messagesReadHistory(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.MessagesReadHistory{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
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
	_ = repo.Dialogs.UpdateReadInboxMaxID(r.ConnInfo.UserID, domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), req.MaxID)

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) messagesGetHistory(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.MessagesGetHistory{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	// Load the dialog
	dialog, _ := repo.Dialogs.Get(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type))
	if dialog == nil {
		fillMessagesMany(out, []*msg.UserMessage{}, []*msg.User{}, []*msg.Group{}, in.RequestID, successCB)
		return
	}

	// Prepare the the result before sending back to the client
	preSuccessCB := genGetHistoryCB(successCB, domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), req.MinID, req.MaxID, dialog.TopMessageID)

	// We are Offline/Disconnected
	if !r.networkCtrl.Connected() {
		messages, users, groups := repo.Messages.GetMessageHistory(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), req.MinID, req.MaxID, req.Limit)
		if len(messages) > 0 {
			pendingMessages := repo.PendingMessages.GetByPeer(req.Peer.ID, int32(req.Peer.Type))
			if len(pendingMessages) > 0 {
				messages = append(pendingMessages, messages...)
			}
			fillMessagesMany(out, messages, users, groups, in.RequestID, successCB)
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
			logs.Info("River detected hole (With MaxID Only)",
				zap.Int64("MaxID", req.MaxID),
				zap.Int64("PeerID", req.Peer.ID),
				zap.Int64("TopMsgID", dialog.TopMessageID),
				zap.String("Holes", messageHole.PrintHole(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), 0)),
			)
			r.queueCtrl.EnqueueCommand(in, timeoutCB, preSuccessCB, true)
			return
		}
		messages, users, groups := repo.Messages.GetMessageHistory(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), bar.Min, bar.Max, req.Limit)
		fillMessagesMany(out, messages, users, groups, in.RequestID, preSuccessCB)
	case req.MinID != 0 && req.MaxID == 0:
		b, bar := messageHole.GetUpperFilled(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), 0, req.MinID)
		if !b {
			logs.Info("River detected hole (With MinID Only)",
				zap.Int64("MinID", req.MinID),
				zap.Int64("PeerID", req.Peer.ID),
				zap.Int64("TopMsgID", dialog.TopMessageID),
				zap.String("Holes", messageHole.PrintHole(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), 0)),
			)
			r.queueCtrl.EnqueueCommand(in, timeoutCB, preSuccessCB, true)
			return
		}
		messages, users, groups := repo.Messages.GetMessageHistory(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), bar.Min, 0, req.Limit)
		fillMessagesMany(out, messages, users, groups, in.RequestID, preSuccessCB)
	default:
		b := messageHole.IsHole(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), 0, req.MinID, req.MaxID)
		if b {
			logs.Info("River detected hole (With Min & Max)",
				zap.Int64("MinID", req.MinID),
				zap.Int64("MaxID", req.MaxID),
				zap.Int64("PeerID", req.Peer.ID),
				zap.Int64("TopMsgID", dialog.TopMessageID),
			)
			r.queueCtrl.EnqueueCommand(in, timeoutCB, preSuccessCB, true)
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
func genGetHistoryCB(
	cb domain.MessageHandler, teamID, peerID int64, peerType int32, minID, maxID int64, topMessageID int64,
) domain.MessageHandler {
	return func(m *rony.MessageEnvelope) {
		pendingMessages := repo.PendingMessages.GetByPeer(peerID, peerType)
		switch m.Constructor {
		case msg.C_MessagesMany:
			x := &msg.MessagesMany{}
			err := x.Unmarshal(m.Message)
			logs.WarnOnErr("Error On Unmarshal MessagesMany", err)

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
			logs.Warn("We received error on GetHistory", zap.Error(domain.ParseServerError(m.Message)))
		default:
		}

		// Call the actual success callback function
		cb(m)
	}
}

func (r *River) messagesGetMediaHistory(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.MessagesGetMediaHistory{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	// Load the dialog
	dialog, _ := repo.Dialogs.Get(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type))
	if dialog == nil {
		fillMessagesMany(out, []*msg.UserMessage{}, []*msg.User{}, []*msg.Group{}, in.RequestID, successCB)
		return
	}

	// We are Offline/Disconnected
	if !r.networkCtrl.Connected() {
		messages, users, groups := repo.Messages.GetMediaMessageHistory(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), 0, req.MaxID, req.Limit, req.Cat)
		if len(messages) > 0 {
			fillMessagesMany(out, messages, users, groups, in.RequestID, successCB)
			return
		}
	}

	// Prepare the the result before sending back to the client
	preSuccessCB := genGetMediaHistoryCB(successCB, domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), req.MaxID, req.Cat)

	// We are Online
	if req.MaxID == 0 {
		req.MaxID = dialog.TopMessageID
	}
	b, bar := messageHole.GetLowerFilled(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), req.Cat, req.MaxID)
	if !b {
		logs.Info("River detected hole (With MaxID Only)",
			zap.Int64("MaxID", req.MaxID),
			zap.Int64("PeerID", req.Peer.ID),
			zap.Int64("TopMsgID", dialog.TopMessageID),
			zap.String("Holes", messageHole.PrintHole(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), 0)),
		)
		r.queueCtrl.EnqueueCommand(in, timeoutCB, preSuccessCB, true)
		return
	}
	messages, users, groups := repo.Messages.GetMediaMessageHistory(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), 0, bar.Max, req.Limit, req.Cat)
	fillMessagesMany(out, messages, users, groups, in.RequestID, preSuccessCB)
}
func genGetMediaHistoryCB(
	cb domain.MessageHandler, teamID, peerID int64, peerType int32, maxID int64, cat msg.MediaCategory,
) domain.MessageHandler {
	return func(m *rony.MessageEnvelope) {
		switch m.Constructor {
		case msg.C_MessagesMany:
			x := &msg.MessagesMany{}
			err := x.Unmarshal(m.Message)
			logs.WarnOnErr("Error On Unmarshal MessagesMany", err)

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
			logs.Warn("We received error on GetHistory", zap.Error(domain.ParseServerError(m.Message)))
		default:
		}

		// Call the actual success callback function
		cb(m)
	}
}

func (r *River) messagesDelete(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.MessagesDelete{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
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
			r.DeletePendingMessage(id)
		}
	}

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)

}

func (r *River) messagesGet(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.MessagesGet{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
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

	messages := repo.Messages.GetMany(msgIDs.ToArray())
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
		uiexec.ExecSuccessCB(successCB, out)
		return
	}

	// SendWebsocket the request to the server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) messagesClearHistory(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.MessagesClearHistory{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	if req.MaxID == 0 {
		d, err := repo.Dialogs.Get(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type))
		if err != nil {
			out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
			successCB(out)
			return
		}
		req.MaxID = d.TopMessageID
	}

	err := repo.Messages.ClearHistory(r.ConnInfo.UserID, domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), req.MaxID)
	logs.WarnOnErr("We got error on clear history", err,
		zap.Int64("PeerID", req.Peer.ID),
		zap.Int64("TeamID", domain.GetTeamID(in)),
	)

	if req.Delete {
		err = repo.Dialogs.Delete(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type))
		logs.WarnOnErr("We got error on deleting dialogs", err,
			zap.Int64("PeerID", req.Peer.ID),
			zap.Int64("TeamID", domain.GetTeamID(in)),
		)
	}

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) messagesReadContents(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.MessagesReadContents{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	repo.Messages.SetContentRead(req.Peer.ID, int32(req.Peer.Type), req.MessageIDs)

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) messagesSendMedia(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.MessagesSendMedia{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	switch req.MediaType {
	case msg.InputMediaType_InputMediaTypeContact, msg.InputMediaType_InputMediaTypeGeoLocation,
		msg.InputMediaType_InputMediaTypeDocument, msg.InputMediaType_InputMediaTypeMessageDocument:
		// This will be used as next requestID
		req.RandomID = domain.SequentialUniqueID()

		// Insert into pending messages, id is negative nano timestamp and save RandomID too : Done
		dbID := -req.RandomID

		res, err := repo.PendingMessages.SaveMessageMedia(domain.GetTeamID(in), domain.GetTeamAccess(in), dbID, r.ConnInfo.UserID, req)
		if err != nil {
			e := &rony.Error{
				Code:  "n/a",
				Items: "Failed to save to pendingMessages : " + err.Error(),
			}
			out.Fill(out.RequestID, rony.C_Error, e)
			uiexec.ExecSuccessCB(successCB, out)
			return
		}
		// Return to CallBack with pending message data : Done
		out.Constructor = msg.C_ClientPendingMessage

		out.Message, _ = res.Marshal()
		uiexec.ExecSuccessCB(successCB, out)

	case msg.InputMediaType_InputMediaTypeUploadedDocument:
		// no need to insert pending message cuz we already insert one b4 start uploading
	}

	requestBytes, _ := req.Marshal()
	r.queueCtrl.EnqueueCommand(
		&rony.MessageEnvelope{
			Constructor: msg.C_MessagesSendMedia,
			RequestID:   uint64(req.RandomID),
			Message:     requestBytes,
			Header:      in.Header,
		},
		timeoutCB, successCB, true,
	)
}

func (r *River) contactsGet(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.ContactsGet{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	res := &msg.ContactsMany{}
	res.ContactUsers, res.Contacts = repo.Users.GetContacts(domain.GetTeamID(in))

	userIDs := make([]int64, 0, len(res.ContactUsers))
	for idx := range res.ContactUsers {
		userIDs = append(userIDs, res.ContactUsers[idx].ID)
	}
	res.Users, _ = repo.Users.GetMany(userIDs)
	out.Constructor = msg.C_ContactsMany
	out.Message, _ = res.Marshal()

	logs.Info("We returned data locally, ContactsGet",
		zap.Int("Users", len(res.Users)),
		zap.Int("Contacts", len(res.Contacts)),
	)
	uiexec.ExecSuccessCB(successCB, out)
}

func (r *River) contactsAdd(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	if domain.GetTeamID(in) != 0 {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: "teams cannot add contact"})
		successCB(out)
		return
	}

	req := &msg.ContactsAdd{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	user, _ := repo.Users.Get(req.User.UserID)
	if user != nil {
		user.FirstName = req.FirstName
		user.LastName = req.LastName
		user.Phone = req.Phone
		_ = repo.Users.SaveContact(domain.GetTeamID(in), &msg.ContactUser{
			ID:         user.ID,
			FirstName:  user.FirstName,
			LastName:   user.LastName,
			AccessHash: user.AccessHash,
			Phone:      user.Phone,
			Username:   user.Username,
			ClientID:   0,
			Photo:      user.Photo,
		})
		_ = repo.Users.Save(user)
	}

	// reset contacts hash to update the contacts
	_ = repo.System.SaveInt(domain.GetContactsGetHashKey(domain.GetTeamID(in)), 0)
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) contactsImport(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	if domain.GetTeamID(in) != 0 {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: "teams cannot import contact"})
		successCB(out)
		return
	}

	req := &msg.ContactsImport{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	// If only importing one contact then we don't need to calculate contacts hash
	if len(req.Contacts) == 1 {
		// send request to server
		r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
		return
	}

	oldHash, err := repo.System.LoadInt(domain.SkContactsImportHash)
	if err != nil {
		logs.Warn("We got error on loading ContactsImportHash", zap.Error(err))
	}
	// calculate ContactsImportHash and compare with oldHash
	newHash := domain.CalculateContactsImportHash(req)
	logs.Info("We returned data locally, ContactsImport",
		zap.Uint64("Old", oldHash),
		zap.Uint64("New", newHash),
	)
	if newHash == oldHash {
		res := &msg.ContactsImported{
			ContactUsers: nil,
			Users:        nil,
			Empty:        true,
		}
		out.Fill(out.RequestID, msg.C_ContactsImported, res)
		successCB(out)
		return
	}

	// not equal save it to DB
	err = repo.System.SaveInt(domain.SkContactsImportHash, newHash)
	if err != nil {
		logs.Error("We got error on saving ContactsImportHash", zap.Error(err))
	}

	// extract differences between existing contacts and new contacts
	_, contacts := repo.Users.GetContacts(domain.GetTeamID(in))
	diffContacts := domain.ExtractsContactsDifference(contacts, req.Contacts)

	err = repo.Users.SavePhoneContact(diffContacts...)
	if err != nil {
		logs.Error("We got error on saving phone contacts in to the db", zap.Error(err))
	}

	if len(diffContacts) <= 250 {
		// send the request to server
		r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
		return
	}

	// chunk contacts by size of 50 and send them to server
	r.syncCtrl.ContactsImport(req.Replace, successCB, out)
}

func (r *River) contactsDelete(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	if domain.GetTeamID(in) != 0 {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: "teams cannot delete contact"})
		successCB(out)
		return
	}

	req := &msg.ContactsDelete{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	_ = repo.Users.DeleteContact(domain.GetTeamID(in), req.UserIDs...)
	_ = repo.System.SaveInt(domain.GetContactsGetHashKey(domain.GetTeamID(in)), 0)

	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
	return
}

func (r *River) contactsDeleteAll(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.ContactsDeleteAll{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	_ = repo.Users.DeleteAllContacts(domain.GetTeamID(in))
	_ = repo.System.SaveInt(domain.GetContactsGetHashKey(domain.GetTeamID(in)), 0)
	_ = repo.System.SaveInt(domain.SkContactsImportHash, 0)
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
	return
}

func (r *River) contactsGetTopPeers(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.ContactsGetTopPeers{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}
	res := &msg.ContactsTopPeers{}
	topPeers, _ := repo.TopPeers.List(domain.GetTeamID(in), req.Category, req.Offset, req.Limit)
	if len(topPeers) == 0 {
		r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
		return
	}

	res.Category = req.Category
	res.Peers = topPeers
	res.Count = int32(len(topPeers))

	mUsers := domain.MInt64B{}
	mGroups := domain.MInt64B{}
	for _, topPeer := range res.Peers {
		switch msg.PeerType(topPeer.Peer.Type) {
		case msg.PeerType_PeerUser, msg.PeerType_PeerExternalUser:
			mUsers[topPeer.Peer.ID] = true
		case msg.PeerType_PeerGroup:
			mGroups[topPeer.Peer.ID] = true
		}
	}
	res.Groups, _ = repo.Groups.GetMany(mGroups.ToArray())
	if len(res.Groups) != len(mGroups) {
		logs.Warn("River found unmatched top peers groups", zap.Int("Got", len(res.Groups)), zap.Int("Need", len(mGroups)))
		for groupID := range mGroups {
			found := false
			for _, g := range res.Groups {
				if g.ID == groupID {
					found = true
					break
				}
			}
			if !found {
				logs.Warn("missed group", zap.Int64("GroupID", groupID))
			}
		}
	}

	res.Users, _ = repo.Users.GetMany(mUsers.ToArray())
	if len(res.Users) != len(mUsers) {
		logs.Warn("River found unmatched top peers users", zap.Int("Got", len(res.Users)), zap.Int("Need", len(mUsers)))
		for userID := range mUsers {
			found := false
			for _, g := range res.Users {
				if g.ID == userID {
					found = true
					break
				}
			}
			if !found {
				logs.Warn("missed user", zap.Int64("UserID", userID))
			}
		}
	}

	out.Constructor = msg.C_ContactsTopPeers
	buff, err := res.Marshal()
	logs.ErrorOnErr("River got error on marshal ContactsTopPeers", err)
	out.Message = buff
	uiexec.ExecSuccessCB(successCB, out)
}

func (r *River) contactsResetTopPeer(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.ContactsResetTopPeer{}
	err := req.Unmarshal(in.Message)
	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	err = repo.TopPeers.Delete(req.Category, domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type))
	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) accountUpdateUsername(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.AccountUpdateUsername{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	r.ConnInfo.Username = req.Username
	r.ConnInfo.Save()

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) accountRegisterDevice(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.AccountRegisterDevice{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}
	r.DeviceToken = req

	val, err := json.Marshal(req)
	if err != nil {
		logs.Error("River::accountRegisterDevice()-> Json Marshal()", zap.Error(err))
		return
	}
	err = repo.System.SaveString(domain.SkDeviceToken, string(val))
	if err != nil {
		logs.Error("River::accountRegisterDevice()-> SaveString()", zap.Error(err))
		return
	}
	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) accountUnregisterDevice(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.AccountUnregisterDevice{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "E00", Items: err.Error()})
		successCB(out)
		return
	}
	r.DeviceToken = &msg.AccountRegisterDevice{}
	val, err := json.Marshal(r.DeviceToken)
	if err != nil {
		logs.Error("River::accountUnregisterDevice()-> Json Marshal()", zap.Error(err))
		return
	}
	err = repo.System.SaveString(domain.SkDeviceToken, string(val))
	if err != nil {
		logs.Error("River::accountUnregisterDevice()-> SaveString()", zap.Error(err))
		return
	}

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) accountSetNotifySettings(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.AccountSetNotifySettings{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	dialog, _ := repo.Dialogs.Get(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type))
	if dialog == nil {
		return
	}

	dialog.NotifySettings = req.Settings
	_ = repo.Dialogs.Save(dialog)

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)

}

func (r *River) gifSave(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.GifSave{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	cf, err := repo.Files.Get(req.Doc.ClusterID, req.Doc.ID, req.Doc.AccessHash)
	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	logs.Info("We are saving GIF",
		zap.Int64("FileID", cf.FileID),
		zap.Uint64("AccessHash", cf.AccessHash),
		zap.Int32("ClusterID", cf.ClusterID),
	)
	if !repo.Gifs.IsSaved(cf.ClusterID, cf.FileID) {
		md := &msg.MediaDocument{
			Doc: &msg.Document{
				ID:          cf.FileID,
				AccessHash:  cf.AccessHash,
				Date:        0,
				MimeType:    cf.MimeType,
				FileSize:    int32(cf.FileSize),
				Version:     cf.Version,
				ClusterID:   cf.ClusterID,
				Attributes:  req.Attributes,
				MD5Checksum: cf.MD5Checksum,
			},
		}
		err = repo.Gifs.Save(md)
		if err != nil {
			out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
			successCB(out)
			return
		}
	}
	_ = repo.Gifs.UpdateLastAccess(cf.ClusterID, cf.FileID, domain.Now().Unix())

	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) gifDelete(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.GifDelete{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	err := repo.Gifs.Delete(req.Doc.ClusterID, req.Doc.ID)
	if err != nil {
		logs.Warn("We got error on deleting GIF document", zap.Error(err))
	}

	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) gifGetSaved(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.GifGetSaved{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}
	gifHash, _ := repo.System.LoadInt(domain.SkGifHash)

	var enqueueSuccessCB domain.MessageHandler

	if gifHash != 0 {
		res, err := repo.Gifs.GetSaved()
		if err != nil {
			out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
			successCB(out)
			return
		}
		out.Fill(out.RequestID, msg.C_SavedGifs, res)
		successCB(out)

		// ignore success cb because we notify views on message hanlder
		enqueueSuccessCB = func(m *rony.MessageEnvelope) {

		}
	} else {
		enqueueSuccessCB = successCB
	}

	r.queueCtrl.EnqueueCommand(in, timeoutCB, enqueueSuccessCB, true)
}

func (r *River) dialogTogglePin(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.MessagesToggleDialogPin{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	dialog, _ := repo.Dialogs.Get(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type))
	if dialog == nil {
		logs.Debug("River::dialogTogglePin()-> GetDialog()",
			zap.String("Error", "Dialog is null"),
		)
		return
	}

	dialog.Pinned = req.Pin
	repo.Dialogs.Save(dialog)

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)

}

func (r *River) accountRemovePhoto(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	x := &msg.AccountRemovePhoto{}
	_ = x.Unmarshal(in.Message)

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)

	user, err := repo.Users.Get(r.ConnInfo.UserID)
	if err != nil {
		return
	}

	if user.Photo != nil && user.Photo.PhotoID == x.PhotoID {
		_ = repo.Users.UpdatePhoto(r.ConnInfo.UserID, &msg.UserPhoto{
			PhotoBig:      &msg.FileLocation{},
			PhotoSmall:    &msg.FileLocation{},
			PhotoBigWeb:   &msg.WebLocation{},
			PhotoSmallWeb: &msg.WebLocation{},
			PhotoID:       0,
		})
	}

	repo.Users.RemovePhotoGallery(r.ConnInfo.UserID, x.PhotoID)
}

func (r *River) accountUpdateProfile(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.AccountUpdateProfile{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	// TODO : add connInfo Bio and save it too
	r.ConnInfo.FirstName = req.FirstName
	r.ConnInfo.LastName = req.LastName
	r.ConnInfo.Bio = req.Bio
	r.ConnInfo.Save()

	_ = repo.Users.UpdateProfile(r.ConnInfo.UserID,
		req.FirstName, req.LastName, r.ConnInfo.Username, req.Bio, r.ConnInfo.Phone,
	)

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) groupsEditTitle(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.GroupsEditTitle)
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	repo.Groups.UpdateTitle(req.GroupID, req.Title)

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)

}

func (r *River) groupAddUser(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.GroupsAddUser)
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}
	user, _ := repo.Users.Get(req.User.UserID)
	if user != nil {
		gp := &msg.GroupParticipant{
			AccessHash: req.User.AccessHash,
			FirstName:  user.FirstName,
			LastName:   user.LastName,
			UserID:     req.User.UserID,
			Type:       msg.ParticipantType_ParticipantTypeMember,
		}
		_ = repo.Groups.AddParticipant(req.GroupID, gp)
	}

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)

}

func (r *River) groupDeleteUser(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.GroupsDeleteUser)
	err := req.Unmarshal(in.Message)
	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	err = repo.Groups.RemoveParticipant(req.GroupID, req.User.UserID)
	if err != nil {
		logs.Error("We got error on GroupDeleteUser local handler", zap.Error(err))
	}

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) groupsGetFull(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.GroupsGetFull)
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	res, err := repo.Groups.GetFull(req.GroupID)
	if err != nil {
		r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
		return
	}

	// NotifySettings
	dlg, _ := repo.Dialogs.Get(domain.GetTeamID(in), req.GroupID, int32(msg.PeerType_PeerGroup))
	if dlg == nil {
		r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
		return
	}
	res.NotifySettings = dlg.NotifySettings

	// Get Group PhotoGallery
	res.PhotoGallery, err = repo.Groups.GetPhotoGallery(req.GroupID)
	if err != nil {
		logs.Error("We got error on GetPhotoGallery in local handler", zap.Error(err))
	}

	// Users
	userIDs := domain.MInt64B{}
	for _, v := range res.Participants {
		userIDs[v.UserID] = true
	}
	users, _ := repo.Users.GetMany(userIDs.ToArray())
	if len(res.Participants) != len(users) {
		r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
		return
	}
	res.Users = users

	out.Constructor = msg.C_GroupFull
	out.Message, _ = res.Marshal()
	successCB(out)
}

func (r *River) groupUpdateAdmin(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.GroupsUpdateAdmin)
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	repo.Groups.UpdateMemberType(req.GroupID, req.User.UserID, req.Admin)

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) groupToggleAdmin(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.GroupsToggleAdmins)
	err := req.Unmarshal(in.Message)
	if err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	err = repo.Groups.ToggleAdmins(req.GroupID, req.AdminEnabled)
	if err != nil {
		logs.Warn("We got error on local handler for GroupToggleAdmin", zap.Error(err))
	}

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) groupRemovePhoto(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)

	req := new(msg.GroupsRemovePhoto)
	err := req.Unmarshal(in.Message)
	if err != nil {
		logs.Error("groupRemovePhoto() failed to unmarshal", zap.Error(err))
	}

	group, _ := repo.Groups.Get(req.GroupID)
	if group == nil {
		return
	}

	if group.Photo != nil && group.Photo.PhotoID == req.PhotoID {
		repo.Groups.UpdatePhoto(req.GroupID, &msg.GroupPhoto{
			PhotoBig:   &msg.FileLocation{},
			PhotoSmall: &msg.FileLocation{},
			PhotoID:    0,
		})
	}

	repo.Users.RemovePhotoGallery(r.ConnInfo.UserID, req.PhotoID)
}

func (r *River) usersGetFull(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.UsersGetFull{}
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::usersGetFull()-> Unmarshal()", zap.Error(err))
		return
	}
	userIDs := domain.MInt64B{}
	for _, v := range req.Users {
		userIDs[v.UserID] = true
	}

	users, outDated, _ := repo.Users.GetManyWithOutdated(userIDs.ToArray())
	allResolved := len(users) == len(userIDs)
	if allResolved {
		res := &msg.UsersMany{}
		for _, user := range users {
			user.PhotoGallery = repo.Users.GetPhotoGallery(user.ID)
			sort.Slice(user.PhotoGallery, func(i, j int) bool {
				return user.PhotoGallery[i].PhotoID > user.PhotoGallery[j].PhotoID
			})
		}
		res.Users = users

		out.Constructor = msg.C_UsersMany
		out.Message, _ = res.Marshal()
		uiexec.ExecSuccessCB(successCB, out)

		if len(outDated) > 0 {
			req.Users = req.Users[:0]
			for _, user := range outDated {
				req.Users = append(req.Users, &msg.InputUser{
					UserID:     user.ID,
					AccessHash: user.AccessHash,
				})
			}
			in.Fill(in.RequestID, in.Constructor, req, in.Header...)
			r.queueCtrl.EnqueueCommand(in, nil, nil, false)
		}
		return
	}

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) usersGet(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.UsersGet{}
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::usersGet()-> Unmarshal()", zap.Error(err))
		return
	}
	userIDs := domain.MInt64B{}
	for _, v := range req.Users {
		userIDs[v.UserID] = true
	}

	users, outDated, _ := repo.Users.GetManyWithOutdated(userIDs.ToArray())
	allResolved := len(users) == len(userIDs)
	if allResolved {
		res := new(msg.UsersMany)
		res.Users = users

		out.Constructor = msg.C_UsersMany
		out.Message, _ = res.Marshal()
		uiexec.ExecSuccessCB(successCB, out)

		if len(outDated) > 0 {
			req.Users = req.Users[:0]
			for _, user := range outDated {
				req.Users = append(req.Users, &msg.InputUser{
					UserID:     user.ID,
					AccessHash: user.AccessHash,
				})
			}
			in.Fill(in.RequestID, in.Constructor, req, in.Header...)
			r.queueCtrl.EnqueueCommand(in, nil, nil, false)
		}
		return
	}

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) messagesSaveDraft(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.MessagesSaveDraft{}
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::messagesSaveDraft()-> Unmarshal()", zap.Error(err))
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
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) messagesClearDraft(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.MessagesClearDraft{}
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::messagesClearDraft()-> Unmarshal()", zap.Error(err))
		return
	}

	dialog, _ := repo.Dialogs.Get(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type))
	if dialog != nil {
		dialog.Draft = nil
		repo.Dialogs.Save(dialog)
	}

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) labelsGet(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.LabelsGet{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	logs.Info("LabelGet", zap.Int64("TeamID", domain.GetTeamID(in)))
	labels := repo.Labels.GetAll(domain.GetTeamID(in))
	sort.Slice(labels, func(i, j int) bool {
		return labels[i].Count > labels[j].Count
	})
	if len(labels) != 0 {
		logs.Debug("We found labels locally", zap.Int("L", len(labels)))
		res := &msg.LabelsMany{}
		res.Labels = labels

		out.Constructor = msg.C_LabelsMany
		out.Message, _ = res.Marshal()
		uiexec.ExecSuccessCB(successCB, out)
		return
	}

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) labelsDelete(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.LabelsDelete{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	logs.Info("LabelsDelete", zap.Int64("TeamID", domain.GetTeamID(in)))
	err := repo.Labels.Delete(req.LabelIDs...)

	logs.ErrorOnErr("LabelsDelete", err)

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) labelsListItems(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.LabelsListItems{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	// Offline mode
	if !r.networkCtrl.Connected() {
		logs.Debug("We are offline then load from local db",
			zap.Int32("LabelID", req.LabelID),
			zap.Int64("MinID", req.MinID),
			zap.Int64("MaxID", req.MaxID),
		)
		messages, users, groups := repo.Labels.ListMessages(req.LabelID, domain.GetTeamID(in), req.Limit, req.MinID, req.MaxID)
		fillLabelItems(out, messages, users, groups, in.RequestID, successCB)
		return
	}

	preSuccessCB := func(m *rony.MessageEnvelope) {
		switch m.Constructor {
		case msg.C_LabelItems:
			x := &msg.LabelItems{}
			err := x.Unmarshal(m.Message)
			logs.WarnOnErr("Error On Unmarshal LabelItems", err)

			// 1st sort the received messages by id
			sort.Slice(x.Messages, func(i, j int) bool {
				return x.Messages[i].ID > x.Messages[j].ID
			})

			// Fill Messages Hole
			if msgCount := len(x.Messages); msgCount > 0 {
				logs.Debug("Update Label Range",
					zap.Int32("LabelID", x.LabelID),
					zap.Int64("MinID", x.Messages[msgCount-1].ID),
					zap.Int64("MaxID", x.Messages[0].ID),
				)

				switch {
				case req.MinID == 0 && req.MaxID != 0:
					_ = repo.Labels.Fill(domain.GetTeamID(in), req.LabelID, x.Messages[msgCount-1].ID, req.MaxID)
				case req.MinID != 0 && req.MaxID == 0:
					_ = repo.Labels.Fill(domain.GetTeamID(in), req.LabelID, req.MinID, x.Messages[0].ID)
				case req.MinID == 0 && req.MaxID == 0:
					_ = repo.Labels.Fill(domain.GetTeamID(in), req.LabelID, x.Messages[msgCount-1].ID, x.Messages[0].ID)
				}
			}
		default:
			logs.Warn("We received unexpected response", zap.String("C", registry.ConstructorName(m.Constructor)))
		}

		successCB(m)
	}

	switch {
	case req.MinID == 0 && req.MaxID == 0:
		r.queueCtrl.EnqueueCommand(in, timeoutCB, preSuccessCB, true)
	case req.MinID == 0 && req.MaxID != 0:
		b, _ := repo.Labels.GetLowerFilled(domain.GetTeamID(in), req.LabelID, req.MaxID)
		if !b {
			logs.Info("River detected label hole (With MaxID Only)",
				zap.Int32("LabelID", req.LabelID),
				zap.Int64("MaxID", req.MaxID),
				zap.Int64("MinID", req.MinID),
			)
			r.queueCtrl.EnqueueCommand(in, timeoutCB, preSuccessCB, true)
			return
		}
		messages, users, groups := repo.Labels.ListMessages(req.LabelID, domain.GetTeamID(in), req.Limit, 0, req.MaxID)
		fillLabelItems(out, messages, users, groups, in.RequestID, preSuccessCB)
	case req.MinID != 0 && req.MaxID == 0:
		b, _ := repo.Labels.GetUpperFilled(domain.GetTeamID(in), req.LabelID, req.MinID)
		if !b {
			logs.Info("River detected label hole (With MinID Only)",
				zap.Int32("LabelID", req.LabelID),
				zap.Int64("MinID", req.MinID),
				zap.Int64("MaxID", req.MaxID),
			)
			r.queueCtrl.EnqueueCommand(in, timeoutCB, preSuccessCB, true)
			return
		}
		messages, users, groups := repo.Labels.ListMessages(req.LabelID, domain.GetTeamID(in), req.Limit, req.MinID, 0)
		fillLabelItems(out, messages, users, groups, in.RequestID, preSuccessCB)
	default:
		r.queueCtrl.EnqueueCommand(in, timeoutCB, preSuccessCB, true)
		return
	}
}

func (r *River) labelAddToMessage(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.LabelsAddToMessage{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	logs.Debug("LabelsAddToMessage local handler called",
		zap.Int64s("MsgIDs", req.MessageIDs),
		zap.Int32s("LabelIDs", req.LabelIDs),
	)
	if len(req.MessageIDs) != 0 {
		_ = repo.Labels.AddLabelsToMessages(req.LabelIDs, domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), req.MessageIDs)
		for _, labelID := range req.LabelIDs {
			bar := repo.Labels.GetFilled(domain.GetTeamID(in), labelID)
			for _, msgID := range req.MessageIDs {
				if msgID > bar.MaxID {
					_ = repo.Labels.Fill(domain.GetTeamID(in), labelID, bar.MaxID, msgID)
				} else if msgID < bar.MinID {
					_ = repo.Labels.Fill(domain.GetTeamID(in), labelID, msgID, bar.MinID)
				}
			}
		}
	}

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)

}

func (r *River) labelRemoveFromMessage(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.LabelsRemoveFromMessage{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	logs.Debug("LabelsRemoveFromMessage local handler called",
		zap.Int64s("MsgIDs", req.MessageIDs),
		zap.Int32s("LabelIDs", req.LabelIDs),
	)

	if len(req.MessageIDs) != 0 {
		_ = repo.Labels.RemoveLabelsFromMessages(req.LabelIDs, domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), req.MessageIDs)
	}

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)

}

func fillLabelItems(out *rony.MessageEnvelope, messages []*msg.UserMessage, users []*msg.User, groups []*msg.Group, requestID uint64, successCB domain.MessageHandler) {
	res := new(msg.LabelItems)
	res.Messages = messages
	res.Users = users
	res.Groups = groups

	out.RequestID = requestID
	out.Constructor = msg.C_LabelItems
	out.Message, _ = res.Marshal()
	uiexec.ExecSuccessCB(successCB, out)
}

func (r *River) systemGetConfig(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	out.Fill(out.RequestID, msg.C_SystemConfig, domain.SysConfig)
	successCB(out)
}

func (r *River) accountsGetTeams(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.AccountGetTeams{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	teams := repo.Teams.List()

	if len(teams) > 0 {
		teamsMany := &msg.TeamsMany{
			Teams: teams,
		}
		out.Fill(out.RequestID, msg.C_TeamsMany, teamsMany)
		successCB(out)
		return
	}

	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) teamEdit(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.TeamEdit{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	team, _ := repo.Teams.Get(req.TeamID)

	if team != nil {
		team.Name = req.Name
		_ = repo.Teams.Save(team)
	}

	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) messagesTogglePin(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.MessagesTogglePin{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	err := repo.Dialogs.UpdatePinMessageID(domain.GetTeamID(in), req.Peer.ID, int32(req.Peer.Type), req.MessageID)
	logs.ErrorOnErr("MessagesTogglePin", err)

	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) messagesSendReaction(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.MessagesSendReaction{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	err := repo.Reactions.IncrementReactionUseCount(req.Reaction, 1)
	logs.ErrorOnErr("messagesSendReaction", err)

	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) messagesDeleteReaction(in, out *rony.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.MessagesDeleteReaction{}
	if err := req.Unmarshal(in.Message); err != nil {
		out.Fill(out.RequestID, rony.C_Error, &rony.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	for _, r := range req.Reactions {
		err := repo.Reactions.IncrementReactionUseCount(r, -1)
		logs.ErrorOnErr("messagesDeleteReaction", err)
	}

	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}
