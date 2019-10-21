package riversdk

import (
	"encoding/json"
	"fmt"
	messageHole "git.ronaksoftware.com/ronak/riversdk/pkg/message_hole"
	"git.ronaksoftware.com/ronak/riversdk/pkg/uiexec"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"sort"
	"strings"
	"sync"
	"time"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"go.uber.org/zap"
)

func (r *River) messagesGetDialogs(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesGetDialogs)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}
	res := new(msg.MessagesDialogs)
	res.Dialogs = repo.Dialogs.List(req.Offset, req.Limit)

	// If the localDB had no data send the request to server
	if len(res.Dialogs) == 0 {
		r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
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
		logs.Debug("Dialog", zap.Int64("MsgID", dialog.TopMessageID), zap.Int64("PeerID", dialog.PeerID))
		if dialog.PeerType == int32(msg.PeerUser) {
			mUsers[dialog.PeerID] = true
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
		logs.Warn("Rive found unmatched dialog messages", zap.Int("Got", len(res.Messages)), zap.Int("Need", len(mMessages)))
	}
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
	// Load Pending messages
	res.Messages = append(res.Messages, pendingMessages...)

	for _, m := range res.Messages {
		switch msg.PeerType(m.PeerType) {
		case msg.PeerUser:
			mUsers[m.PeerID] = true
		case msg.PeerGroup:
			mGroups[m.PeerID] = true
		}
		mUsers[m.SenderID] = true
		if m.FwdSenderID != 0 {
			mUsers[m.FwdSenderID] = true
		}

		// load MessageActionData users
		actUserIDs := domain.ExtractActionUserIDs(m.MessageAction, m.MessageActionData)
		for _, id := range actUserIDs {
			mUsers[id] = true
		}
	}
	res.Groups = repo.Groups.GetMany(mGroups.ToArray())
	if len(res.Groups) != len(mGroups) {
		logs.Warn("Rive found unmatched dialog groups", zap.Int("Got", len(res.Messages)), zap.Int("Need", len(mMessages)))
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
	res.Users = repo.Users.GetMany(mUsers.ToArray())
	if len(res.Users) != len(mUsers) {
		logs.Warn("Rive found unmatched dialog groups", zap.Int("Got", len(res.Messages)), zap.Int("Need", len(mMessages)))
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
	uiexec.Ctx().Exec(func() {
		if successCB != nil {
			successCB(out)
		}
	}) // successCB(out)
}

func (r *River) messagesGetDialog(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesGetDialog)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}
	res := new(msg.Dialog)
	res, _ = repo.Dialogs.Get(req.Peer.ID, int32(req.Peer.Type))

	// if the localDB had no data send the request to server
	if res == nil {
		r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
		return
	}

	out.Constructor = msg.C_Dialog
	out.Message, _ = res.Marshal()

	uiexec.Ctx().Exec(func() {
		if successCB != nil {
			successCB(out)
		}
	}) // successCB(out)
}

func (r *River) messagesSend(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesSend)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	// do not allow empty message
	if strings.TrimSpace(req.Body) == "" {
		e := new(msg.Error)
		e.Code = "n/a"
		e.Items = "empty message is not allowed"
		msg.ResultError(out, e)
		uiexec.Ctx().Exec(func() {
			if successCB != nil {
				successCB(out)
			}
		})
		return
	}

	// this will be used as next requestID
	req.RandomID = domain.SequentialUniqueID()
	msgID := -domain.SequentialUniqueID()
	res, err := repo.PendingMessages.Save(msgID, r.ConnInfo.UserID, req)
	if err != nil {
		e := new(msg.Error)
		e.Code = "n/a"
		e.Items = "Failed to save to pendingMessages : " + err.Error()
		msg.ResultError(out, e)
		uiexec.Ctx().Exec(func() {
			if successCB != nil {
				successCB(out)
			}
		})
		return
	}
	// 2. add to queue [ looks like there is general queue to send messages ] : Done
	requestBytes, _ := req.Marshal()
	// r.queueCtrl.addToWaitingList(req) // this needs to be wrapped by MessageEnvelope
	// using req randomID as requestID later in queue processing and network controller messageHandler
	r.queueCtrl.EnqueueCommand(uint64(req.RandomID), msg.C_MessagesSend, requestBytes, timeoutCB, successCB, true)

	// 3. return to CallBack with pending message data : Done
	out.Constructor = msg.C_ClientPendingMessage
	out.Message, _ = res.Marshal()

	// 4. later when queue got processed and server returned response we should check if the requestID
	//   exist in pendingTable we remove it and insert new message with new id to message table
	//   invoke new OnUpdate with new proto buffer to inform ui that pending message got delivered
	uiexec.Ctx().Exec(func() {
		if successCB != nil {
			successCB(out)
		}
	}) // successCB(out)
}

func (r *River) messagesReadHistory(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesReadHistory)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	dialog, _ := repo.Dialogs.Get(req.Peer.ID, int32(req.Peer.Type))
	if dialog == nil {
		return
	}
	if dialog.ReadInboxMaxID > req.MaxID {
		return
	}

	repo.Dialogs.UpdateReadInboxMaxID(r.ConnInfo.UserID, req.Peer.ID, int32(req.Peer.Type), req.MaxID)

	// send the request to server
	r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
}

func (r *River) messagesGetHistory(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesGetHistory)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	// Load the dialog
	dialog, _ := repo.Dialogs.Get(req.Peer.ID, int32(req.Peer.Type))
	if dialog == nil {
		logs.Debug("asking for a nil dialog")
		fillOutput(out, []*msg.UserMessage{}, []*msg.User{}, in.RequestID, successCB)
		return
	}

	// Prepare the request and update its parameters if necessary
	if req.MaxID < 0 {
		req.MaxID = dialog.TopMessageID
	}

	// Update the request before sending to server
	in.Message, _ = req.Marshal()

	// Offline mode
	if !r.networkCtrl.Connected() {
		messages, users := repo.Messages.GetMessageHistory(req.Peer.ID, int32(req.Peer.Type), req.MinID, req.MaxID, req.Limit)
		pendingMessages := repo.PendingMessages.GetByPeer(req.Peer.ID, int32(req.Peer.Type))
		if len(pendingMessages) > 0 {
			messages = append(messages, pendingMessages...)
		}
		fillOutput(out, messages, users, in.RequestID, successCB)
		return
	}

	// Prepare the the result before sending back to the client
	preSuccessCB := genSuccessCallback(successCB, req.Peer.ID, int32(req.Peer.Type), req.MinID, req.MaxID)

	switch {
	case req.MinID == 0 && req.MaxID == 0:
		req.MaxID = dialog.TopMessageID
		fallthrough
	case req.MinID == 0 && req.MaxID != 0:
		b, bar := messageHole.GetLowerFilled(req.Peer.ID, int32(req.Peer.Type), req.MaxID)
		if !b {
			logs.Info("Range in Hole",
				zap.Int64("PeerID", req.Peer.ID),
				zap.Int64("MaxID", req.MaxID),
				zap.Int64("MinID", req.MinID),
				zap.Int64("TopMsgID", dialog.TopMessageID),
				zap.String("Holes", messageHole.PrintHole(req.Peer.ID, int32(req.Peer.Type))),
			)
			r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, preSuccessCB, true)
			return
		}
		messages, users := repo.Messages.GetMessageHistory(req.Peer.ID, int32(req.Peer.Type), bar.Min, bar.Max, req.Limit)
		fillOutput(out, messages, users, in.RequestID, preSuccessCB)
	case req.MinID != 0 && req.MaxID == 0:
		b, bar := messageHole.GetUpperFilled(req.Peer.ID, int32(req.Peer.Type), req.MinID);
		if !b {
			logs.Info("Range in Hole",
				zap.Int64("PeerID", req.Peer.ID),
				zap.Int64("MaxID", req.MaxID),
				zap.Int64("MinID", req.MinID),
				zap.Int64("TopMsgID", dialog.TopMessageID),
				zap.String("Holes", messageHole.PrintHole(req.Peer.ID, int32(req.Peer.Type))),
			)
			r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, preSuccessCB, true)
			return
		}
		messages, users := repo.Messages.GetMessageHistory(req.Peer.ID, int32(req.Peer.Type), bar.Min, 0, req.Limit)
		fillOutput(out, messages, users, in.RequestID, preSuccessCB)
	default:
		// Min != 0 && MaxID != 0
		b := messageHole.IsHole(req.Peer.ID, int32(req.Peer.Type), req.MinID, req.MaxID)
		if b {
			logs.Info("Range in Hole",
				zap.Int64("PeerID", req.Peer.ID),
				zap.Int64("MaxID", req.MaxID),
				zap.Int64("MinID", req.MinID),
				zap.Int64("TopMsgID", dialog.TopMessageID),
			)
			r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, preSuccessCB, true)
			return
		}
		messages, users := repo.Messages.GetMessageHistory(req.Peer.ID, int32(req.Peer.Type), req.MinID, req.MaxID, req.Limit)
		fillOutput(out, messages, users, in.RequestID, preSuccessCB)
	}
}
func fillOutput(out *msg.MessageEnvelope, messages []*msg.UserMessage, users []*msg.User, requestID uint64, successCB domain.MessageHandler) {
	res := new(msg.MessagesMany)
	res.Messages = messages
	res.Users = users

	out.RequestID = requestID
	out.Constructor = msg.C_MessagesMany
	out.Message, _ = res.Marshal()
	if successCB != nil {
		uiexec.Ctx().Exec(func() {
			successCB(out)
		})
	}
}
func genSuccessCallback(cb domain.MessageHandler, peerID int64, peerType int32, minID, maxID int64) domain.MessageHandler {
	return func(m *msg.MessageEnvelope) {
		pendingMessages := repo.PendingMessages.GetByPeer(peerID, peerType)
		switch m.Constructor {
		case msg.C_MessagesMany:
			x := new(msg.MessagesMany)
			_ = x.Unmarshal(m.Message)
			// 1st sort the received messages by id
			sort.Slice(x.Messages, func(i, j int) bool {
				return x.Messages[i].ID > x.Messages[j].ID
			})

			// Fill Messages Hole
			if msgCount := len(x.Messages); msgCount > 0 {
				switch {
				case minID == 0 && maxID != 0:
					messageHole.InsertFill(peerID, peerType, x.Messages[msgCount-1].ID, maxID)
				case minID != 0 && maxID == 0:
					messageHole.InsertFill(peerID, peerType, minID, x.Messages[0].ID)
				case minID == 0 && maxID == 0:
					messageHole.InsertFill(peerID, peerType, x.Messages[msgCount-1].ID, x.Messages[0].ID)
				}
			}

			if len(pendingMessages) > 0 {
				// 2nd base on the reqMin values add the appropriate pending messages
				switch {
				case maxID == 0:
					x.Messages = append(x.Messages, pendingMessages...)
				default:
					// Min != 0
					for idx := range pendingMessages {
						if len(x.Messages) == 0 {
							x.Messages = append(x.Messages, pendingMessages[idx])
						} else if pendingMessages[idx].CreatedOn > x.Messages[len(x.Messages)-1].CreatedOn {
							x.Messages = append(x.Messages, pendingMessages[idx])
						}
					}
				}

				// 3rd sort again, to sort pending messages and actual messages
				sort.Slice(x.Messages, func(i, j int) bool {
					if x.Messages[i].ID < 0 || x.Messages[j].ID < 0 {
						return x.Messages[i].CreatedOn > x.Messages[j].CreatedOn
					} else {
						return x.Messages[i].ID > x.Messages[j].ID
					}
				})
			}

			m.Message, _ = x.Marshal()
		default:
		}

		// Call the actual success callback function
		cb(m)
	}
}

func (r *River) messagesDelete(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesDelete)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	// Get PendingMessage and cancel its requests
	pendingMessageIDs := make([]int64, 0)
	for _, id := range req.MessageIDs {
		if id < 0 {
			pendingMessageIDs = append(pendingMessageIDs, id)
		}
	}
	if len(pendingMessageIDs) > 0 {
		// remove from queue
		pendedRequestIDs := repo.PendingMessages.GetManyRequestIDs(pendingMessageIDs)
		for _, reqID := range pendedRequestIDs {
			r.queueCtrl.CancelRequest(reqID)
		}
		// remove from DB
		repo.PendingMessages.DeleteMany(pendingMessageIDs)
	}

	// remove message
	for _, msgID := range req.MessageIDs {
		repo.Messages.Delete(r.ConnInfo.UserID, req.Peer.ID, int32(req.Peer.Type), msgID)
	}

	// send the request to server
	r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)

}

func (r *River) messagesGet(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesGet)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}
	msgIDs := domain.MInt64B{}
	for _, v := range req.MessagesIDs {
		msgIDs[v] = true
	}

	messages := repo.Messages.GetMany(msgIDs.ToArray())
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
	users := repo.Users.GetMany(mUsers.ToArray())

	// if db already had all users
	if len(messages) == len(msgIDs) && len(users) > 0 {
		res := new(msg.MessagesMany)
		res.Users = make([]*msg.User, 0)
		res.Messages = messages
		res.Users = append(res.Users, users...)

		out.Constructor = msg.C_MessagesMany
		out.Message, _ = res.Marshal()
		uiexec.Ctx().Exec(func() {
			if successCB != nil {
				successCB(out)
			}
		}) // successCB(out)
		return
	}

	// SendWebsocket the request to the server
	r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
}

func (r *River) messagesClearHistory(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesClearHistory)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	repo.Dialogs.UpdateUnreadCount(req.Peer.ID, int32(req.Peer.Type), 0)

	// this will be handled on message update appliers too
	repo.Messages.Delete(r.ConnInfo.UserID, req.Peer.ID, int32(req.Peer.Type), req.MaxID)

	// send the request to server
	r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)

}

func (r *River) messagesReadContents(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesReadContents)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	repo.Messages.SetContentRead(req.Peer.ID, int32(req.Peer.Type), req.MessageIDs)

	// send the request to server
	r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
}

func (r *River) messagesSendMedia(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesSendMedia)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	switch req.MediaType {
	case msg.InputMediaTypeContact, msg.InputMediaTypeDocument, msg.InputMediaTypeGeoLocation:
		// this will be used as next requestID
		req.RandomID = domain.SequentialUniqueID()
		// 1. insert into pending messages, id is negative nano timestamp and save RandomID too : Done
		dbID := -domain.SequentialUniqueID()
		res, err := repo.PendingMessages.SaveMessageMedia(dbID, r.ConnInfo.UserID, req)
		if err != nil {
			e := new(msg.Error)
			e.Code = "n/a"
			e.Items = "Failed to save to pendingMessages : " + err.Error()
			msg.ResultError(out, e)
			uiexec.Ctx().Exec(func() {
				if successCB != nil {
					successCB(out)
				}
			})
			return
		}
		// 3. return to CallBack with pending message data : Done
		out.Constructor = msg.C_ClientPendingMessage
		out.Message, _ = res.Marshal()

		uiexec.Ctx().Exec(func() {
			if successCB != nil {
				successCB(out)
			}
		}) // successCB(out)

		requestBytes, _ := req.Marshal()
		r.queueCtrl.EnqueueCommand(uint64(req.RandomID), msg.C_MessagesSendMedia, requestBytes, timeoutCB, successCB, true)
	case msg.InputMediaTypeUploadedDocument:
		// no need to insert pending message cuz we already insert one b4 start uploading
		requestBytes, _ := req.Marshal()
		r.queueCtrl.EnqueueCommand(uint64(req.RandomID), msg.C_MessagesSendMedia, requestBytes, timeoutCB, successCB, true)

	}

}

func (r *River) clientSendMessageMedia(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	reqMedia := new(msg.ClientSendMessageMedia)
	if err := reqMedia.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	in.Message, _ = reqMedia.Marshal()
	// TODO : check if file has been uploaded b4

	// 1. insert into pending messages, id is negative nano timestamp and save RandomID too : Done
	msgID := -domain.SequentialUniqueID()
	fileID := ronak.RandomInt64(0)
	thumbID := int64(0)
	if reqMedia.ThumbFilePath != "" {
		thumbID = ronak.RandomInt64(0)
	}
	reqMedia.FileUploadID = fmt.Sprintf("%d", fileID)
	reqMedia.FileID = fileID
	if thumbID > 0 {
		reqMedia.ThumbID = thumbID
		reqMedia.ThumbUploadID = fmt.Sprintf("%d", thumbID)
	}
	pendingMessage, err := repo.PendingMessages.SaveClientMessageMedia(msgID, r.ConnInfo.UserID, fileID, fileID, thumbID, reqMedia)
	if err != nil {
		e := new(msg.Error)
		e.Code = "n/a"
		e.Items = "Failed to save to pendingMessages : " + err.Error()
		msg.ResultError(out, e)
		uiexec.Ctx().Exec(func() {
			if successCB != nil {
				successCB(out)
			}
		})
		return
	}

	// 3. return to CallBack with pending message data : Done
	out.Constructor = msg.C_ClientPendingMessage
	out.Message, _ = pendingMessage.Marshal()

	r.fileCtrl.UploadMessageDocument(pendingMessage.ID, reqMedia.FilePath, reqMedia.ThumbFilePath, fileID, thumbID)

	// 4. later when queue got processed and server returned response we should check if the requestID
	//   exist in pendingTable we remove it and insert new message with new id to message table
	//   invoke new OnUpdate with new protobuff to inform ui that pending message got delivered
	uiexec.Ctx().Exec(func() {
		if successCB != nil {
			successCB(out)
		}
	}) // successCB(out)

}

func (r *River) contactsGet(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.ContactsGet)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	res := new(msg.ContactsMany)
	res.ContactUsers, res.Contacts = repo.Users.GetContacts()

	// if didn't find anything send request to server
	if len(res.Users) == 0 || len(res.Contacts) == 0 {
		r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
		return
	}

	out.Constructor = msg.C_ContactsMany
	out.Message, _ = res.Marshal()

	uiexec.Ctx().Exec(func() {
		if successCB != nil {
			successCB(out)
		}
	}) // successCB(out)
}

func (r *River) contactsImport(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.ContactsImport)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	// TOF
	if len(req.Contacts) == 1 {
		// send request to server
		r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
		return
	}

	res := new(msg.ContactsMany)
	res.ContactUsers, res.Contacts = repo.Users.GetContacts()

	oldHash, err := repo.System.LoadInt(domain.SkContactsImportHash)
	if err != nil {
		logs.Warn("River::contactsImport()-> failed to get contactsImportHash ", zap.Error(err))
	}
	// calculate ContactsImportHash and compare with oldHash
	newHash := domain.CalculateContactsImportHash(req)

	// compare two hashes if equal return existing contacts
	if newHash == oldHash {
		msg.ResultContactsMany(out, res)
		successCB(out)
		return
	}

	// not equal save it to DB
	err = repo.System.SaveInt(domain.SkContactsImportHash, newHash)
	if err != nil {
		logs.Error("River::contactsImport() failed to save ContactsImportHash to DB", zap.Error(err))
	}

	// extract differences between existing contacts and new contacts
	diffContacts := domain.ExtractsContactsDifference(res.Contacts, req.Contacts)
	if len(diffContacts) <= 50 {
		// send the request to server
		r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
	} else {
		// chunk contacts by size of 50 and send them to server
		go r.sendChunkedImportContactRequest(req.Replace, diffContacts, out, successCB)
	}
}

func (r *River) sendChunkedImportContactRequest(replace bool, diffContacts []*msg.PhoneContact, out *msg.MessageEnvelope, successCB domain.MessageHandler) {
	result := new(msg.ContactsImported)
	result.Users = make([]*msg.User, 0)
	result.ContactUsers = make([]*msg.ContactUser, 0)

	mx := sync.Mutex{}
	wg := sync.WaitGroup{}
	cbTimeout := func() {
		wg.Done()
		logs.Error("sendChunkedImportContactRequest() -> cbTimeout() ")
	}

	cbSuccess := func(env *msg.MessageEnvelope) {
		mx.Lock()
		defer mx.Unlock()
		defer wg.Done()

		if env.Constructor == msg.C_ContactsImported {
			x := new(msg.ContactsImported)
			err := x.Unmarshal(env.Message)
			if err != nil {
				logs.Error("sendChunkedImportContactRequest() -> cbSuccess() failed to unmarshal", zap.Error(err))
				return
			}
			result.Users = append(result.Users, x.Users...)
			result.ContactUsers = append(result.ContactUsers, x.ContactUsers...)
		} else {
			logs.Error("sendChunkedImportContactRequest() -> cbSuccess() received unexpected response", zap.String("Constructor", msg.ConstructorNames[env.Constructor]))
		}
	}

	count := len(diffContacts)
	idx := 0
	for idx < count {
		req := new(msg.ContactsImport)
		req.Replace = replace
		req.Contacts = make([]*msg.PhoneContact, 0)

		for i := 0; i < 50 && idx < count; i++ {
			req.Contacts = append(req.Contacts, diffContacts[idx])
			idx++
		}

		reqID := uint64(domain.SequentialUniqueID())
		reqBytes, _ := req.Marshal()

		wg.Add(1)
		r.queueCtrl.EnqueueCommand(reqID, msg.C_ContactsImport, reqBytes, cbTimeout, cbSuccess, false)
	}

	wg.Wait()
	msg.ResultContactsImported(out, result)
	successCB(out)
}

func (r *River) accountUpdateUsername(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.AccountUpdateUsername)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	r.ConnInfo.Username = req.Username
	r.ConnInfo.Save()

	// send the request to server
	r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
}

func (r *River) accountRegisterDevice(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.AccountRegisterDevice)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
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
	r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
}

func (r *River) accountUnregisterDevice(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.AccountUnregisterDevice)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}
	r.DeviceToken = new(msg.AccountRegisterDevice)

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
	r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
}

func (r *River) accountSetNotifySettings(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.AccountSetNotifySettings)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	dialog, _ := repo.Dialogs.Get(req.Peer.ID, int32(req.Peer.Type))
	if dialog == nil {
		return
	}

	dialog.NotifySettings = req.Settings
	repo.Dialogs.Save(dialog)

	// send the request to server
	r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)

}

func (r *River) dialogTogglePin(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesToggleDialogPin)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	dialog, _ := repo.Dialogs.Get(req.Peer.ID, int32(req.Peer.Type))
	if dialog == nil {
		logs.Debug("River::dialogTogglePin()-> GetDialog()",
			zap.String("Error", "Dialog is null"),
		)
		return
	}

	dialog.Pinned = req.Pin
	repo.Dialogs.Save(dialog)

	// send the request to server
	r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)

}

func (r *River) accountRemovePhoto(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	x := new(msg.AccountRemovePhoto)
	_ = x.Unmarshal(in.Message)

	// send the request to server
	r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)

	user, err := repo.Users.Get(r.ConnInfo.UserID)
	if err != nil {
		return
	}

	if user.Photo != nil && user.Photo.PhotoID == x.PhotoID {
		_ = repo.Users.UpdatePhoto(r.ConnInfo.UserID, &msg.UserPhoto{
			PhotoBig:   &msg.FileLocation{},
			PhotoSmall: &msg.FileLocation{},
			PhotoID:    0,
		})
	}

	repo.Users.RemovePhotoGallery(r.ConnInfo.UserID, x.PhotoID)
}

func (r *River) accountUpdateProfile(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.AccountUpdateProfile)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	// TODO : add connInfo Bio and save it too
	r.ConnInfo.FirstName = req.FirstName
	r.ConnInfo.LastName = req.LastName
	r.ConnInfo.Bio = req.Bio
	r.ConnInfo.Save()

	_ = repo.Users.UpdateProfile(r.ConnInfo.UserID,
		req.FirstName, req.LastName, r.ConnInfo.Username, req.Bio,
	)

	// send the request to server
	r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
}

func (r *River) groupsEditTitle(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.GroupsEditTitle)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	repo.Groups.UpdateTitle(req.GroupID, req.Title)

	// send the request to server
	r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)

}

func (r *River) groupAddUser(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.GroupsAddUser)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
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
			Type:       msg.ParticipantTypeMember,
		}
		repo.Groups.SaveParticipant(req.GroupID, gp)
	}

	// send the request to server
	r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)

}

func (r *River) groupDeleteUser(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.GroupsDeleteUser)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	repo.Groups.DeleteMember(req.GroupID, req.User.UserID)

	// send the request to server
	r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
}

func (r *River) groupsGetFull(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.GroupsGetFull)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	res := new(msg.GroupFull)

	group, _ := repo.Groups.Get(req.GroupID)
	if group == nil {
		r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
		return
	}
	res.Group = group

	// Participants
	participants, err := repo.Groups.GetParticipants(req.GroupID)
	if err != nil {
		r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
		return
	}
	res.Participants = participants

	// NotifySettings
	dlg, _ := repo.Dialogs.Get(req.GroupID, int32(msg.PeerGroup))
	if dlg == nil {
		r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
		return
	}
	res.NotifySettings = dlg.NotifySettings

	// Get Group PhotoGallery
	res.PhotoGallery = repo.Groups.GetPhotoGallery(req.GroupID)

	// Users
	userIDs := domain.MInt64B{}
	for _, v := range participants {
		userIDs[v.UserID] = true
	}
	users := repo.Users.GetMany(userIDs.ToArray())
	if users == nil || len(participants) != len(users) || len(users) <= 0 {
		r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
		return
	}
	res.Users = users

	out.Constructor = msg.C_GroupFull
	out.Message, _ = res.Marshal()
	successCB(out)

	// send the request to server no need to pass server response to external handler (UI) just to re update catch DB
	r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, nil, nil, false)

}

func (r *River) groupUpdateAdmin(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.GroupsUpdateAdmin)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	repo.Groups.UpdateMemberType(req.GroupID, req.User.UserID, req.Admin)

	// send the request to server
	r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
}

func (r *River) groupRemovePhoto(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	// send the request to server
	r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)

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

func (r *River) usersGetFull(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.UsersGetFull)
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::usersGetFull()-> Unmarshal()", zap.Error(err))
		return
	}
	userIDs := domain.MInt64B{}
	for _, v := range req.Users {
		userIDs[v.UserID] = true
	}

	users := repo.Users.GetMany(userIDs.ToArray())

	if len(users) == len(userIDs) {
		res := new(msg.UsersMany)
		for _, user := range users {
			user.PhotoGallery = repo.Users.GetPhotoGallery(user.ID)
		}
		res.Users = users

		out.Constructor = msg.C_UsersMany
		out.Message, _ = res.Marshal()
		uiexec.Ctx().Exec(func() {
			if successCB != nil {
				successCB(out)
			}
		}) // successCB(out)
		return
	}

	// send the request to server
	r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
}

func (r *River) usersGet(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.UsersGet)
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::usersGet()-> Unmarshal()", zap.Error(err))
		return
	}
	userIDs := domain.MInt64B{}
	for _, v := range req.Users {
		userIDs[v.UserID] = true
	}

	users := repo.Users.GetMany(userIDs.ToArray())
	if len(users) == len(userIDs) {
		res := new(msg.UsersMany)
		res.Users = users

		out.Constructor = msg.C_UsersMany
		out.Message, _ = res.Marshal()
		uiexec.Ctx().Exec(func() {
			if successCB != nil {
				successCB(out)
			}
		}) // successCB(out)
		return
	}

	// send the request to server
	r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
}

func (r *River) messagesSaveDraft(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesSaveDraft)
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::messagesSaveDraft()-> Unmarshal()", zap.Error(err))
		return
	}

	dialog, _ := repo.Dialogs.Get(req.Peer.ID, int32(req.Peer.Type))
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
	r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
}

func (r *River) messagesClearDraft(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesClearDraft)
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::messagesClearDraft()-> Unmarshal()", zap.Error(err))
		return
	}

	dialog, _ := repo.Dialogs.Get(req.Peer.ID, int32(req.Peer.Type))
	if dialog != nil {
		dialog.Draft = nil
		repo.Dialogs.Save(dialog)
	}

	// send the request to server
	r.queueCtrl.EnqueueCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
}
