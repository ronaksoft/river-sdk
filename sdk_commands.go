package riversdk

import (
	"encoding/json"
	"git.ronaksoftware.com/ronak/riversdk/pkg/filemanager"
	messageHole "git.ronaksoftware.com/ronak/riversdk/pkg/message_hole"
	"strings"
	"sync"

	"git.ronaksoftware.com/ronak/riversdk/pkg/uiexec"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"go.uber.org/zap"
)

func (r *River) messagesGetDialogs(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesGetDialogs)
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::messagesGetDialogs()-> Unmarshal()", zap.Error(err))
		return
	}
	res := new(msg.MessagesDialogs)
	res.Dialogs = repo.Dialogs.GetDialogs(req.Offset, req.Limit)

	// If the localDB had no data send the request to server
	if len(res.Dialogs) == 0 {
		r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
		return
	}

	mUsers := domain.MInt64B{}
	mGroups := domain.MInt64B{}
	mMessages := domain.MInt64B{}
	mPendingMessage := domain.MInt64B{}
	for _, d := range res.Dialogs {
		if d.PeerType == int32(msg.PeerUser) {
			mUsers[d.PeerID] = true
		}
		mMessages[d.TopMessageID] = true
		if d.TopMessageID < 0 {
			mPendingMessage[d.TopMessageID] = true
		}
	}

	// Load Messages
	res.Messages = repo.Messages.GetManyMessages(mMessages.ToArray())

	// Load Pending messages
	pendingMessages := repo.PendingMessages.GetManyPendingMessages(mPendingMessage.ToArray())
	res.Messages = append(res.Messages, pendingMessages...)

	for _, m := range res.Messages {
		switch msg.PeerType(m.PeerType) {
		case msg.PeerUser:
			mUsers[m.PeerID] = true
		case msg.PeerGroup:
			mGroups[m.PeerID] = true
		}

		mUsers[m.SenderID] = true
		mUsers[m.FwdSenderID] = true

		// load MessageActionData users
		actUserIDs := domain.ExtractActionUserIDs(m.MessageAction, m.MessageActionData)
		for _, id := range actUserIDs {
			mUsers[id] = true
		}
	}
	res.Groups = repo.Groups.GetManyGroups(mGroups.ToArray())
	res.Users = repo.Users.GetAnyUsers(mUsers.ToArray())
	out.Constructor = msg.C_MessagesDialogs
	buff, err := res.Marshal()
	if err != nil {
		logs.Error("River::messagesGetDialogs()-> res.Marshal()", zap.Error(err))
	}
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
		logs.Error("River::messagesGetDialog()-> Unmarshal()", zap.Error(err))
		return
	}
	res := new(msg.Dialog)
	res = repo.Dialogs.GetDialog(req.Peer.ID, int32(req.Peer.Type))

	// if the localDB had no data send the request to server
	if res == nil {
		r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
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
		logs.Error("River::messagesSend()-> Unmarshal()", zap.Error(err))
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

	// TODO :
	// 0. fix database and add MessagesPending table : Done

	// 1. insert into pending messages, id is negative nano timestamp and save RandomID too : Done
	dbID := -domain.SequentialUniqueID()
	res, err := repo.PendingMessages.Save(dbID, r.ConnInfo.UserID, req)
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
	r.queueCtrl.ExecuteCommand(uint64(req.RandomID), msg.C_MessagesSend, requestBytes, timeoutCB, successCB, true)

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
		logs.Error("River::messagesReadHistory()-> Unmarshal()", zap.Error(err))
	}

	dialog := repo.Dialogs.GetDialog(req.Peer.ID, int32(req.Peer.Type))
	if dialog == nil {
		return
	}
	if dialog.ReadInboxMaxID > req.MaxID {
		return
	}

	err := repo.Dialogs.UpdateReadInboxMaxID(r.ConnInfo.UserID, req.Peer.ID, int32(req.Peer.Type), req.MaxID)
	if err != nil {
		logs.Error("River::messagesReadHistory()-> UpdateReadInboxMaxID()", zap.Error(err))
	}

	// send the request to server
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
}

func (r *River) messagesGetHistory(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesGetHistory)
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::messagesGetHistory()-> Unmarshal()", zap.Error(err))
		return
	}

	dtoDialog := repo.Dialogs.GetDialog(req.Peer.ID, int32(req.Peer.Type))
	if dtoDialog == nil {
		out.Constructor = msg.C_Error
		out.RequestID = in.RequestID
		msg.ResultError(out, &msg.Error{Code: "-1", Items: "dialog does not exist"})
		uiexec.Ctx().Exec(func() {
			if successCB != nil {
				successCB(out)
			}
		})
		return
	}

	// Offline mode
	if !r.networkCtrl.Connected() {
		if dtoDialog.TopMessageID < 0 {
			messages, users := repo.Messages.GetMessageHistoryWithPendingMessages(req.Peer.ID, int32(req.Peer.Type), req.MinID, req.MaxID, req.Limit)
			messagesGetHistory(out, messages, users, in.RequestID, successCB)
		} else {
			messages, users := repo.Messages.GetMessageHistory(req.Peer.ID, int32(req.Peer.Type), req.MinID, req.MaxID, req.Limit)
			messagesGetHistory(out, messages, users, in.RequestID, successCB)
		}
		return
	}

	if req.MaxID < 0 {
		req.MaxID = 0
	}
	switch {
	case req.MinID == 0 && req.MaxID == 0:
		// Get the latest messages
		if dtoDialog.TopMessageID < 0 {
			// TODO:: WTF ?
			// fetch messages from localDB cuz there is a pending message it means we are not connected to server
			messages, users := repo.Messages.GetMessageHistoryWithPendingMessages(req.Peer.ID, int32(req.Peer.Type), 0, 0, req.Limit)
			messagesGetHistory(out, messages, users, in.RequestID, successCB)
			return
		}
		b, bar := messageHole.GetLowerFilled(req.Peer.ID, int32(req.Peer.Type), dtoDialog.TopMessageID)
		if !b {
			logs.Info("Range in Hole",
				zap.Int64("PeerID", req.Peer.ID),
				zap.Int64("MaxID", dtoDialog.TopMessageID),
				zap.Int64("MinID", req.MinID),
				zap.String("Holes", messageHole.PrintHole(req.Peer.ID, int32(req.Peer.Type))),
			)
			r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
			return
		}
		if bar.Min == 0 {
			messages, users := repo.Messages.GetMessageHistory(req.Peer.ID, int32(req.Peer.Type), bar.Min, bar.Max, req.Limit)
			messagesGetHistory(out, messages, users, in.RequestID, successCB)
			return
		}
		messages, users := repo.Messages.GetMessageHistory(req.Peer.ID, int32(req.Peer.Type), bar.Min, bar.Max, req.Limit)
		if len(messages) < int(req.Limit) {
			r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
			return
		}
		messagesGetHistory(out, messages, users, in.RequestID, successCB)
	case req.MinID == 0 && req.MaxID != 0:
		// Load more message, scroll up
		b, bar := messageHole.GetLowerFilled(req.Peer.ID, int32(req.Peer.Type), req.MaxID)
		if !b {
			logs.Info("Range in Hole",
				zap.Int64("PeerID", req.Peer.ID),
				zap.Int64("MaxID", req.MaxID),
				zap.Int64("MinID", req.MinID),
				zap.String("Holes", messageHole.PrintHole(req.Peer.ID, int32(req.Peer.Type))),
			)
			cb := func(m *msg.MessageEnvelope) {
				switch m.Constructor {
				case msg.C_MessagesMany:
					x := new(msg.MessagesMany)
					_ = x.Unmarshal(m.Message)
					if x.Continuous && len(x.Messages) < int(req.Limit) {
						_ = messageHole.SetLowerFilled(req.Peer.ID, int32(req.Peer.Type))
					}
				case msg.C_Error:
				default:
				}
				successCB(m)
			}
			r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, cb, true)
			return
		}
		if bar.Min == 0 {
			messages, users := repo.Messages.GetMessageHistory(req.Peer.ID, int32(req.Peer.Type), bar.Min, bar.Max, req.Limit)
			messagesGetHistory(out, messages, users, in.RequestID, successCB)
			return
		}
		messages, users := repo.Messages.GetMessageHistory(req.Peer.ID, int32(req.Peer.Type), bar.Min, bar.Max, req.Limit)
		if len(messages) < int(req.Limit) {
			r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
			return
		}
		messagesGetHistory(out, messages, users, in.RequestID, successCB)
	case req.MinID != 0 && req.MaxID == 0:
		// Load more message, scroll down
		b, bar := messageHole.GetUpperFilled(req.Peer.ID, int32(req.Peer.Type), req.MinID)
		if !b {
			logs.Info("Range in Hole",
				zap.Int64("PeerID", req.Peer.ID),
				zap.Int64("MaxID", req.MaxID),
				zap.Int64("MinID", req.MinID),
			)
			r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
			return
		}
		messages, users := repo.Messages.GetMessageHistory(req.Peer.ID, int32(req.Peer.Type), bar.Min, req.MaxID, req.Limit)
		messagesGetHistory(out, messages, users, in.RequestID, successCB)
	default:
		// Load a range
		b, _ := messageHole.IsHole(req.Peer.ID, int32(req.Peer.Type), req.MinID, req.MaxID)
		if b {
			logs.Info("Range in Hole",
				zap.Int64("PeerID", req.Peer.ID),
				zap.Int64("MaxID", req.MaxID),
				zap.Int64("MinID", req.MinID),
			)
			r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
			return
		}
		messages, users := repo.Messages.GetMessageHistory(req.Peer.ID, int32(req.Peer.Type), req.MinID, req.MaxID, req.Limit)
		messagesGetHistory(out, messages, users, in.RequestID, successCB)
	}
}

func messagesGetHistory(out *msg.MessageEnvelope, messages []*msg.UserMessage, users []*msg.User, requestID uint64, successCB domain.MessageHandler) {
	res := new(msg.MessagesMany)
	res.Messages = messages
	res.Users = users
	// result
	out.RequestID = requestID
	out.Constructor = msg.C_MessagesMany
	out.Message, _ = res.Marshal()
	uiexec.Ctx().Exec(func() {
		if successCB != nil {
			successCB(out)
		}
	})
}

func (r *River) messagesDelete(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesDelete)
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::messagesDelete()-> Unmarshal()", zap.Error(err))
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
		pendedRequestIDs := repo.PendingMessages.GetManyPendingMessagesRequestID(pendingMessageIDs)
		for _, reqID := range pendedRequestIDs {
			r.queueCtrl.CancelRequest(reqID)
		}
		// remove from DB
		err := repo.PendingMessages.DeleteManyPendingMessage(pendingMessageIDs)
		if err != nil {
			logs.Error("River::messagesDelete()-> DeletePendingMessage()", zap.Error(err))
		}
	}

	// remove message
	err := repo.Messages.DeleteMany(req.MessageIDs)
	if err != nil {
		logs.Error("River::messagesDelete()-> DeleteMany()", zap.Error(err))
	}

	// send the request to server
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)

}

func (r *River) messagesGet(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesGet)
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::messagesGet()-> Unmarshal()", zap.Error(err))
		return
	}
	msgIDs := domain.MInt64B{}
	for _, v := range req.MessagesIDs {
		msgIDs[v] = true
	}

	messages := repo.Messages.GetManyMessages(msgIDs.ToArray())
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
	users := repo.Users.GetAnyUsers(mUsers.ToArray())

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

	// Send the request to the server
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
}

func (r *River) messagesClearHistory(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesClearHistory)
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::messagesClearHistory()-> Unmarshal()", zap.Error(err))
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}
	// this will be handled on message update appliers too
	err := repo.Messages.DeleteDialogMessage(req.Peer.ID, int32(req.Peer.Type), req.MaxID)
	if err != nil {
		logs.Error("River::messagesClearHistory()-> DeleteDialogMessage()", zap.Error(err))
	}

	// send the request to server
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)

}

func (r *River) messagesReadContents(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesReadContents)
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::messagesReadContents()-> Unmarshal()", zap.Error(err))
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	err := repo.Messages.SetContentRead(req.MessageIDs)
	if err != nil {
		logs.Error("River::groupUpdateAdmin()-> UpdateGroupMemberType()", zap.Error(err))
	}

	// send the request to server
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
}

func (r *River) messagesSendMedia(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesSendMedia)
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::messagesSendMedia()-> Unmarshal()", zap.Error(err))
		return
	}

	switch req.MediaType {
	case msg.InputMediaTypeEmpty:
		// NOT IMPLEMENTED
	case msg.InputMediaTypeUploadedPhoto:
		// NOT IMPLEMENTED
	case msg.InputMediaTypePhoto:
		// NOT IMPLEMENTED
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
		r.queueCtrl.ExecuteCommand(uint64(req.RandomID), msg.C_MessagesSendMedia, requestBytes, timeoutCB, successCB, true)

	case msg.InputMediaTypeUploadedDocument:
		// no need to insert pending message cuz we already insert one b4 start uploading
		requestBytes, _ := req.Marshal()
		r.queueCtrl.ExecuteCommand(uint64(req.RandomID), msg.C_MessagesSendMedia, requestBytes, timeoutCB, successCB, true)

	case msg.Reserved1:
		// NOT IMPLEMENTED
	case msg.Reserved2:
		// NOT IMPLEMENTED
	case msg.Reserved3:
		// NOT IMPLEMENTED
	case msg.Reserved4:
		// NOT IMPLEMENTED
	case msg.Reserved5:
		// NOT IMPLEMENTED
	}

}

func (r *River) clientSendMessageMedia(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	reqMedia := new(msg.ClientSendMessageMedia)
	if err := reqMedia.Unmarshal(in.Message); err != nil {
		logs.Error("River::clientSendMessageMedia()-> Unmarshal()", zap.Error(err))
		return
	}

	// support IOS file path
	if strings.HasPrefix(reqMedia.FilePath, "file://") {
		reqMedia.FilePath = reqMedia.FilePath[7:]
		in.Message, _ = reqMedia.Marshal()
	}
	if strings.HasPrefix(reqMedia.ThumbFilePath, "file://") {
		reqMedia.ThumbFilePath = reqMedia.ThumbFilePath[7:]
		in.Message, _ = reqMedia.Marshal()
	}

	// TODO : check if file has been uploaded b4
	dtoFile := repo.Files.GetExistingFileDocument(reqMedia.FilePath)
	fileAlreadyUploaded := false
	if dtoFile != nil {
		fileAlreadyUploaded = dtoFile.ClusterID > 0 && dtoFile.DocumentID > 0 && dtoFile.AccessHash > 0
	}
	if fileAlreadyUploaded {
		msgID := -domain.SequentialUniqueID()
		fileID := uint64(domain.SequentialUniqueID())
		res, err := repo.PendingMessages.SaveClientMessageMedia(msgID, r.ConnInfo.UserID, int64(fileID), reqMedia)
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

		req := new(msg.MessagesSendMedia)
		req.ClearDraft = reqMedia.ClearDraft
		req.MediaType = msg.InputMediaTypeDocument
		doc := new(msg.InputMediaDocument)
		doc.Caption = reqMedia.Caption
		doc.Document = new(msg.InputDocument)
		doc.Document.AccessHash = uint64(dtoFile.AccessHash)
		doc.Document.ClusterID = dtoFile.ClusterID
		doc.Document.ID = dtoFile.DocumentID
		req.MediaData, _ = doc.Marshal()
		req.Peer = reqMedia.Peer
		req.RandomID = domain.SequentialUniqueID()
		req.ReplyTo = reqMedia.ReplyTo

		reqBytes, _ := req.Marshal()
		in.Constructor = msg.C_MessagesSendMedia
		in.Message = reqBytes

		// TODO :FIX THIS : required behaviour add to pending message and put this to progress pipe line and etc ...
		// 3. return to CallBack with pending message data : Done
		out.Constructor = msg.C_ClientPendingMessage
		out.Message, _ = res.Marshal()

		// 4. later when queue got processed and server returned response we should check if the requestID
		//   exist in pendingTable we remove it and insert new message with new id to message table
		//   invoke new OnUpdate with new protobuff to inform ui that pending message got delivered
		uiexec.Ctx().Exec(func() {
			if successCB != nil {
				successCB(out)
			}
		}) // successCB(out)

		r.onFileUploadCompleted(msgID, int64(fileID), 0, dtoFile.ClusterID, -1, domain.FileStateExistedUpload, dtoFile.FilePath, reqMedia, 0, 0)
		// send the request to server
		r.queueCtrl.ExecuteCommand(fileID, in.Constructor, in.Message, nil, nil, false)

	} else {
		// 1. insert into pending messages, id is negative nano timestamp and save RandomID too : Done
		msgID := -domain.SequentialUniqueID()
		fileID := int64(in.RequestID)
		res, err := repo.PendingMessages.SaveClientMessageMedia(msgID, r.ConnInfo.UserID, fileID, reqMedia)

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

		// 2. start file upload and send process
		err = filemanager.Ctx().Upload(fileID, res)
		if err != nil {
			e := new(msg.Error)
			e.Code = "n/a"
			e.Items = "Failed to start Upload : " + err.Error()
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

		// 4. later when queue got processed and server returned response we should check if the requestID
		//   exist in pendingTable we remove it and insert new message with new id to message table
		//   invoke new OnUpdate with new protobuff to inform ui that pending message got delivered
		uiexec.Ctx().Exec(func() {
			if successCB != nil {
				successCB(out)
			}
		}) // successCB(out)
	}

}

func (r *River) contactsGet(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.ContactsGet)
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::contactsGet()-> Unmarshal()", zap.Error(err))
		return
	}

	res := new(msg.ContactsMany)
	res.Users, res.Contacts = repo.Users.GetContacts()

	// if didn't find anything send request to server
	if len(res.Users) == 0 || len(res.Contacts) == 0 {
		r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
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
		logs.Error("River::contactsImport()-> Unmarshal()", zap.Error(err))
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	// TOF
	if len(req.Contacts) == 1 {
		// send request to server
		r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
		return
	}

	res := new(msg.ContactsMany)
	res.Users, res.Contacts = repo.Users.GetContacts()

	oldHash, err := repo.System.LoadInt(domain.ColumnContactsImportHash)
	if err != nil {
		logs.Error("River::contactsImport()-> failed to get contactsImportHash ", zap.Error(err))
	}
	// calculate ContactsImportHash and compare with oldHash
	newHash := domain.CalculateContactsImportHash(req)

	// compare two hashes if equal return existing contacts
	if int(newHash) == oldHash {
		msg.ResultContactsMany(out, res)
		successCB(out)
		return
	}

	// not equal save it to DB
	err = repo.System.SaveInt(domain.ColumnContactsImportHash, int32(newHash))
	if err != nil {
		logs.Error("River::contactsImport() failed to save ContactsImportHash to DB", zap.Error(err))
	}

	// update phone contacts just a double check
	if req.Replace {
		for _, c := range req.Contacts {
			err := repo.Users.UpdatePhoneContact(c)
			if err != nil {
				logs.Error("River::contactsImport()-> UpdatePhoneContact()", zap.Error(err))
			}
		}
	}
	// extract differences between existing contacts and new contacts
	diffContacts := domain.ExtractsContactsDifference(res.Contacts, req.Contacts)
	if len(diffContacts) <= 50 {
		// send the request to server
		r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
	} else {
		// chunk contacts by size of 50 and send them to server
		go r.sendChunkedImportContactRequest(req.Replace, diffContacts, out, successCB)
	}
}

func (r *River) sendChunkedImportContactRequest(replace bool, diffContacts []*msg.PhoneContact, out *msg.MessageEnvelope, successCB domain.MessageHandler) {
	result := new(msg.ContactsImported)
	result.Users = make([]*msg.ContactUser, 0)

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
		r.queueCtrl.ExecuteCommand(reqID, msg.C_ContactsImport, reqBytes, cbTimeout, cbSuccess, false)
	}

	wg.Wait()
	msg.ResultContactsImported(out, result)
	successCB(out)
}

func (r *River) accountUpdateUsername(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.AccountUpdateUsername)
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::accountUpdateUsername()-> Unmarshal()", zap.Error(err))
		return
	}

	r.ConnInfo.Username = req.Username
	r.ConnInfo.Save()

	// send the request to server
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
}

func (r *River) accountRegisterDevice(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.AccountRegisterDevice)
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::accountRegisterDevice()-> Unmarshal()", zap.Error(err))
		return
	}
	r.DeviceToken = req

	val, err := json.Marshal(req)
	if err != nil {
		logs.Error("River::accountRegisterDevice()-> Json Marshal()", zap.Error(err))
		return
	}
	err = repo.System.SaveString(domain.ColumnDeviceToken, string(val))
	if err != nil {
		logs.Error("River::accountRegisterDevice()-> SaveString()", zap.Error(err))
		return
	}
	// send the request to server
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
}

func (r *River) accountUnregisterDevice(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.AccountUnregisterDevice)
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::accountUnregisterDevice()-> Unmarshal()", zap.Error(err))
		return
	}
	r.DeviceToken = new(msg.AccountRegisterDevice)

	val, err := json.Marshal(r.DeviceToken)
	if err != nil {
		logs.Error("River::accountUnregisterDevice()-> Json Marshal()", zap.Error(err))
		return
	}
	err = repo.System.SaveString(domain.ColumnDeviceToken, string(val))
	if err != nil {
		logs.Error("River::accountUnregisterDevice()-> SaveString()", zap.Error(err))
		return
	}
	// send the request to server
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
}

func (r *River) accountSetNotifySettings(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.AccountSetNotifySettings)
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::accountSetNotifySettings()-> Unmarshal()", zap.Error(err))
		return
	}

	dialog := repo.Dialogs.GetDialog(req.Peer.ID, int32(req.Peer.Type))
	if dialog == nil {
		logs.Debug("River::accountSetNotifySettings()-> GetDialog()",
			zap.String("Error", "Dialog is null"),
		)
		return
	}

	dialog.NotifySettings = req.Settings
	err := repo.Dialogs.SaveDialog(dialog, 0)
	if err != nil {
		logs.Error("River::accountSetNotifySettings()-> SaveDialog()", zap.Error(err))
		return
	}

	// send the request to server
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)

}

func (r *River) dialogTogglePin(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesToggleDialogPin)
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::dialogTogglePin()-> Unmarshal()", zap.Error(err))
		return
	}

	dialog := repo.Dialogs.GetDialog(req.Peer.ID, int32(req.Peer.Type))
	if dialog == nil {
		logs.Debug("River::dialogTogglePin()-> GetDialog()",
			zap.String("Error", "Dialog is null"),
		)
		return
	}

	dialog.Pinned = req.Pin
	err := repo.Dialogs.SaveDialog(dialog, 0)
	if err != nil {
		logs.Error("River::dialogTogglePin()-> SaveDialog()", zap.Error(err))
		return
	}

	// send the request to server
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)

}

func (r *River) accountRemovePhoto(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {

	// send the request to server
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)

	err := repo.Users.RemoveUserPhoto(r.ConnInfo.UserID)
	if err != nil {
		logs.Error("accountRemovePhoto()", zap.Error(err))
	}
}

func (r *River) accountUpdateProfile(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.AccountUpdateProfile)
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::accountUpdateProfile()-> Unmarshal()", zap.Error(err))
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	// TODO : add connInfo Bio and save it too
	r.ConnInfo.FirstName = req.FirstName
	r.ConnInfo.LastName = req.LastName
	r.ConnInfo.Bio = req.Bio
	r.ConnInfo.Save()

	err := repo.Users.UpdateUserProfile(r.ConnInfo.UserID, req)
	if err != nil {
		logs.Error("River::accountUpdateProfile()-> UpdateUserProfile()", zap.Error(err))
	}

	// send the request to server
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
}

func (r *River) groupsEditTitle(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.GroupsEditTitle)
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::messagesEditGroupTitle()-> Unmarshal()", zap.Error(err))
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	err := repo.Groups.UpdateGroupTitle(req.GroupID, req.Title)
	if err != nil {
		logs.Error("River::messagesEditGroupTitle()-> UpdateGroupTitle()", zap.Error(err))
	}

	// send the request to server
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)

}

func (r *River) groupAddUser(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.GroupsAddUser)
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::groupAddUser()-> Unmarshal()", zap.Error(err))
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}
	user := repo.Users.GetUser(req.User.UserID)
	if user != nil {
		gp := &msg.GroupParticipant{
			AccessHash: req.User.AccessHash,
			FirstName:  user.FirstName,
			LastName:   user.LastName,
			UserID:     req.User.UserID,
			Type:       msg.ParticipantTypeMember,
		}
		err := repo.Groups.SaveParticipants(req.GroupID, gp)
		// TODO : Increase group ParticipantCount
		if err != nil {
			logs.Error("River::groupAddUser()-> SaveParticipants()", zap.Error(err))
		}
	}

	// send the request to server
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)

}

func (r *River) groupDeleteUser(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.GroupsDeleteUser)
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::groupDeleteUser()-> Unmarshal()", zap.Error(err))
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	err := repo.Groups.DeleteGroupMember(req.GroupID, req.User.UserID)
	// TODO : Decrease group ParticipantCount
	if err != nil {
		logs.Error("River::groupDeleteUser()-> SaveParticipants()", zap.Error(err))
	}

	// send the request to server
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
}

// FIXME : until implementation of GroupFull.Participants we dont get info of users that added to group when group created
func (r *River) groupsGetFull(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.GroupsGetFull)
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::groupsGetFull()-> Unmarshal()", zap.Error(err))
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	res := new(msg.GroupFull)
	// Group
	group, err := repo.Groups.GetGroup(req.GroupID)
	if err != nil {
		logs.Error("River::groupsGetFull()-> GetGroup() Sending Request To Server !!!", zap.Error(err))
		r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
		return
	}
	res.Group = group

	// Participants
	participents, err := repo.Groups.GetParticipants(req.GroupID)
	if err != nil {
		logs.Error("River::groupsGetFull()-> GetParticipants() Sending Request To Server !!!", zap.Error(err))
		r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
		return
	}
	res.Participants = participents

	// NotifySettings
	dlg := repo.Dialogs.GetDialog(req.GroupID, int32(msg.PeerGroup))
	if dlg == nil {
		logs.Warn("River::groupsGetFull()-> GetDialog() Sending Request To Server !!!")

		r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
		return
	}
	res.NotifySettings = dlg.NotifySettings

	// Users
	userIDs := domain.MInt64B{}
	for _, v := range participents {
		userIDs[v.UserID] = true
	}
	users := repo.Users.GetAnyUsers(userIDs.ToArray())
	if users == nil || len(participents) != len(users) || len(users) <= 0 {
		logs.Warn("River::groupsGetFull()-> GetAnyUsers() Sending Request To Server !!!",
			zap.Bool("Is user nil ? ", users == nil),
			zap.Int("Participanr Count", len(participents)),
			zap.Int("Users Count", len(users)),
		)

		r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
		return
	}
	res.Users = users

	out.Constructor = msg.C_GroupFull
	out.Message, _ = res.Marshal()
	successCB(out)

	// send the request to server no need to pass server response to external handler (UI) just to re update catch DB
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, nil, nil, false)

}

func (r *River) groupUpdateAdmin(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.GroupsUpdateAdmin)
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::groupUpdateAdmin()-> Unmarshal()", zap.Error(err))
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	err := repo.Groups.UpdateGroupMemberType(req.GroupID, req.User.UserID, req.Admin)
	// TODO : Decrease group ParticipantCount
	if err != nil {
		logs.Error("River::groupUpdateAdmin()-> UpdateGroupMemberType()", zap.Error(err))
	}

	// send the request to server
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
}

func (r *River) groupRemovePhoto(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	// send the request to server
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)

	req := new(msg.GroupsRemovePhoto)
	err := req.Unmarshal(in.Message)
	if err != nil {
		logs.Error("groupRemovePhoto() failed to unmarshal", zap.Error(err))
	}

	err = repo.Groups.RemoveGroupPhoto(req.GroupID)
	if err != nil {
		logs.Error("groupRemovePhoto()", zap.Error(err))
	}
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

	users := repo.Users.GetAnyUsers(userIDs.ToArray())

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
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
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

	users := repo.Users.GetAnyUsers(userIDs.ToArray())

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
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
}
