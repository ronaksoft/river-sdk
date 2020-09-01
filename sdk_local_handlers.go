package riversdk

import (
	"encoding/json"
	"fmt"
	messageHole "git.ronaksoft.com/ronak/riversdk/pkg/message_hole"
	"git.ronaksoft.com/ronak/riversdk/pkg/uiexec"
	"sort"
	"strings"
	"sync"
	"time"

	"git.ronaksoft.com/river/msg/msg"
	"git.ronaksoft.com/ronak/riversdk/internal/logs"
	"git.ronaksoft.com/ronak/riversdk/pkg/domain"
	"git.ronaksoft.com/ronak/riversdk/pkg/repo"
	"go.uber.org/zap"
)

func (r *River) messagesGetDialogs(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.MessagesGetDialogs{}
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}
	res := &msg.MessagesDialogs{}
	res.Dialogs = repo.Dialogs.List(in.Team.ID, req.Offset, req.Limit)

	// If the localDB had no data send the request to server
	if len(res.Dialogs) == 0 {
		res.UpdateID = r.syncCtrl.UpdateID()
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
		if dialog.PeerType == int32(msg.PeerUser) {
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
		r.syncCtrl.GetAllDialogs(waitGroup, in.Team, 0, 100)
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
		case msg.PeerUser:
			mUsers[m.PeerID] = true
		case msg.PeerGroup:
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

func (r *River) messagesGetDialog(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.MessagesGetDialog{}
	err := req.Unmarshal(in.Message)
	if err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}
	res := &msg.Dialog{}
	res, err = repo.Dialogs.Get(in.Team.ID, req.Peer.ID, int32(req.Peer.Type))

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

func (r *River) messagesSend(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.MessagesSend{}
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
		uiexec.ExecSuccessCB(successCB, out)
		return
	}

	if req.Peer.ID == r.ConnInfo.UserID {
		r.HandleDebugActions(req.Body)
	}

	// this will be used as next requestID
	req.RandomID = domain.SequentialUniqueID()
	msgID := -req.RandomID
	res, err := repo.PendingMessages.Save(in.Team, msgID, r.ConnInfo.UserID, req)
	if err != nil {
		e := &msg.Error{
			Code:  "n/a",
			Items: "Failed to save to pendingMessages : " + err.Error(),
		}
		msg.ResultError(out, e)
		uiexec.ExecSuccessCB(successCB, out)
		return
	}
	// 2. add to queue [ looks like there is general queue to send messages ] : Done
	requestBytes, _ := req.Marshal()
	// r.queueCtrl.addToWaitingList(req) // this needs to be wrapped by MessageEnvelope
	// using req randomID as requestID later in queue processing and network controller messageHandler
	r.queueCtrl.EnqueueCommand(
		&msg.MessageEnvelope{
			Team:        in.Team,
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

func (r *River) messagesReadHistory(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesReadHistory)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	dialog, _ := repo.Dialogs.Get(in.Team.ID, req.Peer.ID, int32(req.Peer.Type))
	if dialog == nil {
		return
	}
	if dialog.ReadInboxMaxID > req.MaxID {
		return
	}

	repo.Dialogs.UpdateReadInboxMaxID(r.ConnInfo.UserID, in.Team.ID, req.Peer.ID, int32(req.Peer.Type), req.MaxID)

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) messagesGetHistory(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesGetHistory)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	// Load the dialog
	dialog, _ := repo.Dialogs.Get(in.Team.ID, req.Peer.ID, int32(req.Peer.Type))
	if dialog == nil {
		logs.Debug("asking for a nil dialog")
		fillMessagesMany(out, []*msg.UserMessage{}, []*msg.User{}, in.RequestID, successCB)
		return
	}

	// Prepare the request and update its parameters if necessary
	if req.MaxID < 0 {
		req.MaxID = dialog.TopMessageID
	}

	// Update the request before sending to server
	in.Message, _ = req.Marshal()

	// Prepare the the result before sending back to the client
	preSuccessCB := genSuccessCallback(successCB, in.Team.ID, req.Peer.ID, int32(req.Peer.Type), req.MinID, req.MaxID, dialog.TopMessageID)

	// Offline mode
	if !r.networkCtrl.Connected() {
		messages, users := repo.Messages.GetMessageHistory(in.Team.ID, req.Peer.ID, int32(req.Peer.Type), req.MinID, req.MaxID, req.Limit)
		if len(messages) > 0 {
			pendingMessages := repo.PendingMessages.GetByPeer(req.Peer.ID, int32(req.Peer.Type))
			if len(pendingMessages) > 0 {
				messages = append(pendingMessages, messages...)
			}
			fillMessagesMany(out, messages, users, in.RequestID, successCB)
			return
		}
	}

	switch {
	case req.MinID == 0 && req.MaxID == 0:
		req.MaxID = dialog.TopMessageID
		fallthrough
	case req.MinID == 0 && req.MaxID != 0:
		b, bar := messageHole.GetLowerFilled(in.Team.ID, req.Peer.ID, int32(req.Peer.Type), req.MaxID)
		if !b {
			logs.Info("River detected hole (With MaxID Only)",
				zap.Int64("MaxID", req.MaxID),
				zap.Int64("PeerID", req.Peer.ID),
				zap.Int64("TopMsgID", dialog.TopMessageID),
				zap.String("Holes", messageHole.PrintHole(in.Team.ID, req.Peer.ID, int32(req.Peer.Type))),
			)
			r.queueCtrl.EnqueueCommand(in, timeoutCB, preSuccessCB, true)
			return
		}
		messages, users := repo.Messages.GetMessageHistory(in.Team.ID, req.Peer.ID, int32(req.Peer.Type), bar.Min, bar.Max, req.Limit)
		fillMessagesMany(out, messages, users, in.RequestID, preSuccessCB)
	case req.MinID != 0 && req.MaxID == 0:
		b, bar := messageHole.GetUpperFilled(in.Team.ID, req.Peer.ID, int32(req.Peer.Type), req.MinID)
		if !b {
			logs.Info("River detected hole (With MinID Only)",
				zap.Int64("MinID", req.MinID),
				zap.Int64("PeerID", req.Peer.ID),
				zap.Int64("TopMsgID", dialog.TopMessageID),
				zap.String("Holes", messageHole.PrintHole(in.Team.ID, req.Peer.ID, int32(req.Peer.Type))),
			)
			r.queueCtrl.EnqueueCommand(in, timeoutCB, preSuccessCB, true)
			return
		}
		messages, users := repo.Messages.GetMessageHistory(in.Team.ID, req.Peer.ID, int32(req.Peer.Type), bar.Min, 0, req.Limit)
		fillMessagesMany(out, messages, users, in.RequestID, preSuccessCB)
	default:
		b := messageHole.IsHole(in.Team.ID, req.Peer.ID, int32(req.Peer.Type), req.MinID, req.MaxID)
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
		messages, users := repo.Messages.GetMessageHistory(in.Team.ID, req.Peer.ID, int32(req.Peer.Type), req.MinID, req.MaxID, req.Limit)
		fillMessagesMany(out, messages, users, in.RequestID, preSuccessCB)
	}
}
func fillMessagesMany(out *msg.MessageEnvelope, messages []*msg.UserMessage, users []*msg.User, requestID uint64, successCB domain.MessageHandler) {
	res := new(msg.MessagesMany)
	res.Messages = messages
	res.Users = users

	out.RequestID = requestID
	out.Constructor = msg.C_MessagesMany
	out.Message, _ = res.Marshal()
	uiexec.ExecSuccessCB(successCB, out)
}
func genSuccessCallback(cb domain.MessageHandler, teamID, peerID int64, peerType int32, minID, maxID int64, topMessageID int64) domain.MessageHandler {
	return func(m *msg.MessageEnvelope) {
		pendingMessages := repo.PendingMessages.GetByPeer(peerID, peerType)
		switch m.Constructor {
		case msg.C_MessagesMany:
			x := new(msg.MessagesMany)
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
					messageHole.InsertFill(teamID, peerID, peerType, x.Messages[msgCount-1].ID, maxID)
				case minID != 0 && maxID == 0:
					messageHole.InsertFill(teamID, peerID, peerType, minID, x.Messages[0].ID)
				case minID == 0 && maxID == 0:
					messageHole.InsertFill(teamID, peerID, peerType, x.Messages[msgCount-1].ID, x.Messages[0].ID)
				}
			}

			if len(pendingMessages) > 0 {
				if maxID == 0 || (len(x.Messages) > 0 && x.Messages[len(x.Messages)-1].ID == topMessageID) {
					x.Messages = append(pendingMessages, x.Messages...)
				}
			}

			m.Message, _ = x.Marshal()
		case msg.C_Error:
			logs.Warn("We received error on GetHistory", zap.Error(domain.ParseServerError(m.Message)))
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

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)

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
	users, _ := repo.Users.GetMany(mUsers.ToArray())

	// if db already had all users
	if len(messages) == len(msgIDs) && len(users) > 0 {
		res := new(msg.MessagesMany)
		res.Users = make([]*msg.User, 0)
		res.Messages = messages
		res.Users = append(res.Users, users...)

		out.Constructor = msg.C_MessagesMany
		out.Message, _ = res.Marshal()
		uiexec.ExecSuccessCB(successCB, out)
		return
	}

	// SendWebsocket the request to the server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) messagesClearHistory(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesClearHistory)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
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
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) messagesSendMedia(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.MessagesSendMedia{}
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	switch req.MediaType {
	case msg.InputMediaTypeContact, msg.InputMediaTypeGeoLocation,
		msg.InputMediaTypeDocument, msg.InputMediaTypeMessageDocument:
		// This will be used as next requestID
		req.RandomID = domain.SequentialUniqueID()

		// Insert into pending messages, id is negative nano timestamp and save RandomID too : Done
		dbID := -req.RandomID

		res, err := repo.PendingMessages.SaveMessageMedia(in.Team, dbID, r.ConnInfo.UserID, req)
		if err != nil {
			e := &msg.Error{
				Code:  "n/a",
				Items: "Failed to save to pendingMessages : " + err.Error(),
			}
			msg.ResultError(out, e)
			uiexec.ExecSuccessCB(successCB, out)
			return
		}
		// Return to CallBack with pending message data : Done
		out.Constructor = msg.C_ClientPendingMessage
		out.Message, _ = res.Marshal()
		uiexec.ExecSuccessCB(successCB, out)

	case msg.InputMediaTypeUploadedDocument:
		// no need to insert pending message cuz we already insert one b4 start uploading
	}

	requestBytes, _ := req.Marshal()
	r.queueCtrl.EnqueueCommand(
		&msg.MessageEnvelope{
			Constructor: msg.C_MessagesSendMedia,
			RequestID:   uint64(req.RandomID),
			Message:     requestBytes,
			Team:        in.Team,
		},
		timeoutCB, successCB, true,
	)
}

func (r *River) contactsGet(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.ContactsGet{}
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	res := &msg.ContactsMany{}
	res.ContactUsers, res.Contacts = repo.Users.GetContacts(in.Team.ID)

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

func (r *River) contactsAdd(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	if in.Team.ID != 0 {
		msg.ResultError(out, &msg.Error{Code: "00", Items: "teams cannot add contact"})
		successCB(out)
		return
	}

	req := &msg.ContactsAdd{}
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	user, _ := repo.Users.Get(req.User.UserID)
	if user != nil {
		user.FirstName = req.FirstName
		user.LastName = req.LastName
		user.Phone = req.Phone
		_ = repo.Users.SaveContact(in.Team.ID,&msg.ContactUser{
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
	_ = repo.System.SaveInt(domain.GetContactsGetHashKey(in.Team.ID), 0)
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) contactsImport(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	if in.Team.ID != 0 {
		msg.ResultError(out, &msg.Error{Code: "00", Items: "teams cannot import contact"})
		successCB(out)
		return
	}

	req := new(msg.ContactsImport)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
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
		msg.ResultContactsImported(out, res)
		successCB(out)
		return
	}

	// not equal save it to DB
	err = repo.System.SaveInt(domain.SkContactsImportHash, newHash)
	if err != nil {
		logs.Error("We got error on saving ContactsImportHash", zap.Error(err))
	}

	// extract differences between existing contacts and new contacts
	_, contacts := repo.Users.GetContacts(in.Team.ID)
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

func (r *River) contactsDelete(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	if in.Team.ID != 0 {
		msg.ResultError(out, &msg.Error{Code: "00", Items: "teams cannot delete contact"})
		successCB(out)
		return
	}

	req := &msg.ContactsDelete{}
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	_ = repo.Users.DeleteContact(in.Team.ID,req.UserIDs...)
	_ = repo.System.SaveInt(domain.GetContactsGetHashKey(in.Team.ID), 0)

	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
	return
}

func (r *River) contactsDeleteAll(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.ContactsDeleteAll{}
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	_ = repo.Users.DeleteAllContacts(in.Team.ID)
	_ = repo.System.SaveInt(domain.GetContactsGetHashKey(in.Team.ID), 0)
	_ = repo.System.SaveInt(domain.SkContactsImportHash, 0)
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
	return
}

func (r *River) contactsGetTopPeers(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.ContactsGetTopPeers{}
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}
	res := &msg.ContactsTopPeers{}
	topPeers, _ := repo.TopPeers.List(in.Team.ID, req.Category, req.Offset, req.Limit)
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
		case msg.PeerUser:
			mUsers[topPeer.Peer.ID] = true
		case msg.PeerGroup:
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

func (r *River) contactsResetTopPeer(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.ContactsResetTopPeer{}
	err := req.Unmarshal(in.Message)
	if err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	err = repo.TopPeers.Delete(req.Category, in.Team.ID, req.Peer.ID, int32(req.Peer.Type))
	if err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
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
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
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
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) accountUnregisterDevice(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.AccountUnregisterDevice)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "E00", Items: err.Error()})
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
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) accountSetNotifySettings(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.AccountSetNotifySettings)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	dialog, _ := repo.Dialogs.Get(in.Team.ID, req.Peer.ID, int32(req.Peer.Type))
	if dialog == nil {
		return
	}

	dialog.NotifySettings = req.Settings
	repo.Dialogs.Save(dialog)

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)

}

func (r *River) gifSave(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.GifSave{}
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	cf, err := repo.Files.Get(req.Doc.ClusterID, req.Doc.ID, req.Doc.AccessHash)
	if err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
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
			msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
			successCB(out)
			return
		}
	}
	_ = repo.Gifs.UpdateLastAccess(cf.ClusterID, cf.FileID, domain.Now().Unix())

	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) gifDelete(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.GifDelete{}
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	err := repo.Gifs.Delete(req.Doc.ClusterID, req.Doc.ID)
	if err != nil {
		logs.Warn("We got error on deleting GIF document", zap.Error(err))
	}

	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) dialogTogglePin(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesToggleDialogPin)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	dialog, _ := repo.Dialogs.Get(in.Team.ID, req.Peer.ID, int32(req.Peer.Type))
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

func (r *River) accountRemovePhoto(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	x := new(msg.AccountRemovePhoto)
	_ = x.Unmarshal(in.Message)

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)

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
		req.FirstName, req.LastName, r.ConnInfo.Username, req.Bio, r.ConnInfo.Phone,
	)

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
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
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)

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
		_ = repo.Groups.AddParticipant(req.GroupID, gp)
	}

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)

}

func (r *River) groupDeleteUser(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.GroupsDeleteUser)
	err := req.Unmarshal(in.Message)
	if err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
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

func (r *River) groupsGetFull(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.GroupsGetFull)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	res, err := repo.Groups.GetFull(req.GroupID)
	if err != nil {
		r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
		return
	}

	// NotifySettings
	dlg, _ := repo.Dialogs.Get(in.Team.ID, req.GroupID, int32(msg.PeerGroup))
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

func (r *River) groupUpdateAdmin(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.GroupsUpdateAdmin)
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	repo.Groups.UpdateMemberType(req.GroupID, req.User.UserID, req.Admin)

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) groupToggleAdmin(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.GroupsToggleAdmins)
	err := req.Unmarshal(in.Message)
	if err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
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

func (r *River) groupRemovePhoto(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
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

func (r *River) usersGetFull(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.UsersGetFull{}

	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::usersGetFull()-> Unmarshal()", zap.Error(err))
		return
	}
	userIDs := domain.MInt64B{}
	for _, v := range req.Users {
		userIDs[v.UserID] = true
	}

	users, _ := repo.Users.GetMany(userIDs.ToArray())

	if len(users) == len(userIDs) {
		res := new(msg.UsersMany)
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
		return
	}

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) usersGet(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.UsersGet{}
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::usersGet()-> Unmarshal()", zap.Error(err))
		return
	}
	userIDs := domain.MInt64B{}
	for _, v := range req.Users {
		userIDs[v.UserID] = true
	}

	users, _ := repo.Users.GetMany(userIDs.ToArray())
	if len(users) == len(userIDs) {
		res := new(msg.UsersMany)
		res.Users = users

		out.Constructor = msg.C_UsersMany
		out.Message, _ = res.Marshal()
		uiexec.ExecSuccessCB(successCB, out)
		return
	}

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) messagesSaveDraft(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.MessagesSaveDraft{}
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::messagesSaveDraft()-> Unmarshal()", zap.Error(err))
		return
	}

	dialog, _ := repo.Dialogs.Get(in.Team.ID, req.Peer.ID, int32(req.Peer.Type))
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

func (r *River) messagesClearDraft(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.MessagesClearDraft{}
	if err := req.Unmarshal(in.Message); err != nil {
		logs.Error("River::messagesClearDraft()-> Unmarshal()", zap.Error(err))
		return
	}

	dialog, _ := repo.Dialogs.Get(in.Team.ID, req.Peer.ID, int32(req.Peer.Type))
	if dialog != nil {
		dialog.Draft = nil
		repo.Dialogs.Save(dialog)
	}

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)
}

func (r *River) labelsGet(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.LabelsGet{}
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	labels := repo.Labels.GetAll()
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

func (r *River) labelsListItems(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.LabelsListItems{}
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
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
		messages, users, groups := repo.Labels.ListMessages(req.LabelID, in.Team.ID, req.Limit, req.MinID, req.MaxID)
		fillLabelItems(out, messages, users, groups, in.RequestID, successCB)
		return
	}

	preSuccessCB := func(m *msg.MessageEnvelope) {
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
					_ = repo.Labels.Fill(req.LabelID, x.Messages[msgCount-1].ID, req.MaxID)
				case req.MinID != 0 && req.MaxID == 0:
					_ = repo.Labels.Fill(req.LabelID, req.MinID, x.Messages[0].ID)
				case req.MinID == 0 && req.MaxID == 0:
					_ = repo.Labels.Fill(req.LabelID, x.Messages[msgCount-1].ID, x.Messages[0].ID)
				}
			}
		default:
			logs.Warn("We received unexpected response", zap.String("C", msg.ConstructorNames[m.Constructor]))
		}

		successCB(m)
	}

	switch {
	case req.MinID == 0 && req.MaxID == 0:
		bar := repo.Labels.GetFilled(req.LabelID)
		logs.Debug("Label Filled", zap.Int32("LabelID", req.LabelID),
			zap.Int64("MinID", bar.MinID),
			zap.Int64("MaxID", bar.MaxID),
		)
		req.MaxID = bar.MaxID
		fallthrough
	case req.MinID == 0 && req.MaxID != 0:
		b, bar := repo.Labels.GetLowerFilled(req.LabelID, req.MaxID)
		if !b {
			logs.Info("River detected label hole (With MaxID Only)",
				zap.Int32("LabelID", req.LabelID),
				zap.Int64("MaxID", req.MaxID),
				zap.Int64("MinID", req.MinID),
			)
			r.queueCtrl.EnqueueCommand(in, timeoutCB, preSuccessCB, true)
			return
		}
		messages, users, groups := repo.Labels.ListMessages(req.LabelID, in.Team.ID, req.Limit, bar.MinID, bar.MaxID)
		logs.Debug("List Messages By Label",
			zap.Int32("LabelID", req.LabelID),
			zap.Int64("MinID", bar.MinID),
			zap.Int64("MaxID", bar.MaxID),
			zap.Int("Count", len(messages)),
		)
		fillLabelItems(out, messages, users, groups, in.RequestID, preSuccessCB)
	case req.MinID != 0 && req.MaxID == 0:
		b, bar := repo.Labels.GetUpperFilled(req.LabelID, req.MinID)
		if !b {
			logs.Info("River detected label hole (With MinID Only)",
				zap.Int32("LabelID", req.LabelID),
				zap.Int64("MinID", req.MinID),
				zap.Int64("MaxID", req.MaxID),
			)
			r.queueCtrl.EnqueueCommand(in, timeoutCB, preSuccessCB, true)
			return
		}
		messages, users, groups := repo.Labels.ListMessages(req.LabelID, in.Team.ID, req.Limit, bar.MinID, bar.MaxID)
		fillLabelItems(out, messages, users, groups, in.RequestID, preSuccessCB)
	default:
		r.queueCtrl.EnqueueCommand(in, timeoutCB, preSuccessCB, true)
		return
	}
}

func (r *River) labelAddToMessage(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.LabelsAddToMessage{}
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	logs.Debug("LabelsAddToMessage local handler called",
		zap.Int64s("MsgIDs", req.MessageIDs),
		zap.Int32s("LabelIDs", req.LabelIDs),
	)
	if len(req.MessageIDs) != 0 {
		_ = repo.Labels.AddLabelsToMessages(req.LabelIDs, in.Team.ID, req.Peer.ID, int32(req.Peer.Type), req.MessageIDs)
		for _, labelID := range req.LabelIDs {
			bar := repo.Labels.GetFilled(labelID)
			for _, msgID := range req.MessageIDs {
				if msgID > bar.MaxID {
					_ = repo.Labels.Fill(labelID, bar.MaxID, msgID)
				} else if msgID < bar.MinID {
					_ = repo.Labels.Fill(labelID, msgID, bar.MinID)
				}
			}
		}
	}

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)

}

func (r *River) labelRemoveFromMessage(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.LabelsRemoveFromMessage{}
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	logs.Debug("LabelsRemoveFromMessage local handler called",
		zap.Int64s("MsgIDs", req.MessageIDs),
		zap.Int32s("LabelIDs", req.LabelIDs),
	)

	if len(req.MessageIDs) != 0 {
		_ = repo.Labels.RemoveLabelsFromMessages(req.LabelIDs, in.Team.ID, req.Peer.ID, int32(req.Peer.Type), req.MessageIDs)
	}

	// send the request to server
	r.queueCtrl.EnqueueCommand(in, timeoutCB, successCB, true)

}

func fillLabelItems(out *msg.MessageEnvelope, messages []*msg.UserMessage, users []*msg.User, groups []*msg.Group, requestID uint64, successCB domain.MessageHandler) {
	res := new(msg.LabelItems)
	res.Messages = messages
	res.Users = users
	res.Groups = groups

	out.RequestID = requestID
	out.Constructor = msg.C_LabelItems
	out.Message, _ = res.Marshal()
	uiexec.ExecSuccessCB(successCB, out)
}

func (r *River) clientSendMessageMedia(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	reqMedia := new(msg.ClientSendMessageMedia)
	if err := reqMedia.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		uiexec.ExecSuccessCB(successCB, out)
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
	if reqMedia.ThumbFilePath != "" {
		thumbID = domain.RandomInt63()
	}
	reqMedia.FileUploadID = fmt.Sprintf("%d", fileID)
	reqMedia.FileID = fileID
	if thumbID > 0 {
		reqMedia.ThumbID = thumbID
		reqMedia.ThumbUploadID = fmt.Sprintf("%d", thumbID)
	}

	checkSha256 := true
	switch reqMedia.MediaType {
	case msg.InputMediaTypeUploadedDocument:
		for _, attr := range reqMedia.Attributes {
			if attr.Type == msg.AttributeTypeAudio {
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
	pendingMessage, err := repo.PendingMessages.SaveClientMessageMedia(in.Team, msgID, r.ConnInfo.UserID, fileID, fileID, thumbID, reqMedia, h)
	if err != nil {
		e := new(msg.Error)
		e.Code = "n/a"
		e.Items = "Failed to save to pendingMessages : " + err.Error()
		msg.ResultError(out, e)
		uiexec.ExecSuccessCB(successCB, out)
		return
	}

	// 3. return to CallBack with pending message data : Done
	msg.ResultClientPendingMessage(out, pendingMessage)

	// 4. Start the upload process
	r.fileCtrl.UploadMessageDocument(pendingMessage.ID, reqMedia.FilePath, reqMedia.ThumbFilePath, fileID, thumbID, h, pendingMessage.PeerID, checkSha256)

	uiexec.ExecSuccessCB(successCB, out)
}

func (r *River) clientGlobalSearch(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.ClientGlobalSearch{}
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	searchPhrase := strings.ToLower(req.Text)
	searchResults := &msg.ClientSearchResult{}
	var userContacts []*msg.ContactUser
	var nonContacts []*msg.ContactUser
	var msgs []*msg.UserMessage
	if len(req.LabelIDs) > 0 {
		if req.Peer != nil {
			msgs = repo.Messages.SearchByLabels(req.LabelIDs, req.Peer.ID, req.Limit)
		} else {
			msgs = repo.Messages.SearchByLabels(req.LabelIDs, 0, req.Limit)
		}

	} else if req.SenderID != 0 {
		msgs = repo.Messages.SearchBySender(searchPhrase, req.SenderID, req.Peer.ID, req.Limit)
	} else if req.Peer != nil {
		msgs = repo.Messages.SearchTextByPeerID(searchPhrase, req.Peer.ID, req.Limit)
	} else {
		msgs = repo.Messages.SearchText(searchPhrase, req.Limit)
	}

	// get users && group IDs
	userIDs := domain.MInt64B{}
	matchedUserIDs := domain.MInt64B{}
	groupIDs := domain.MInt64B{}
	for _, m := range msgs {
		if m.PeerType == int32(msg.PeerSelf) || m.PeerType == int32(msg.PeerUser) {
			userIDs[m.PeerID] = true
		}
		if m.PeerType == int32(msg.PeerGroup) {
			groupIDs[m.PeerID] = true
		}
		if m.SenderID > 0 {
			userIDs[m.SenderID] = true
		} else {
			groupIDs[m.PeerID] = true
		}
		if m.FwdSenderID > 0 {
			userIDs[m.FwdSenderID] = true
		} else {
			groupIDs[m.FwdSenderID] = true
		}
	}

	// if peerID == 0 then look for group and contact names too
	if req.Peer == nil {
		userContacts, _ = repo.Users.SearchContacts(searchPhrase)
		for _, userContact := range userContacts {
			matchedUserIDs[userContact.ID] = true
		}
		nonContacts = repo.Users.SearchNonContacts(searchPhrase)
		for _, userContact := range nonContacts {
			matchedUserIDs[userContact.ID] = true
		}
		searchResults.MatchedGroups = repo.Groups.Search(searchPhrase)
	}

	users, _ := repo.Users.GetMany(userIDs.ToArray())
	groups, _ := repo.Groups.GetMany(groupIDs.ToArray())
	matchedUsers, _ := repo.Users.GetMany(matchedUserIDs.ToArray())

	searchResults.Messages = msgs
	searchResults.Users = users
	searchResults.Groups = groups
	searchResults.MatchedUsers = matchedUsers

	out.RequestID = in.RequestID
	out.Constructor = msg.C_ClientSearchResult
	out.Message, _ = searchResults.Marshal()
	uiexec.ExecSuccessCB(successCB, out)
}

func (r *River) clientContactSearch(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.ClientContactSearch{}
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	searchPhrase := strings.ToLower(req.Text)
	logs.Info("SearchContacts", zap.String("Phrase", searchPhrase))

	users := &msg.UsersMany{}
	contactUsers, _ := repo.Users.SearchContacts(searchPhrase)
	userIDs := make([]int64, 0, len(contactUsers))
	for _, contactUser := range contactUsers {
		userIDs = append(userIDs, contactUser.ID)
	}
	users.Users, _ = repo.Users.GetMany(userIDs)

	out.Constructor = msg.C_UsersMany
	out.RequestID = in.RequestID
	out.Message, _ = users.Marshal()
	uiexec.ExecSuccessCB(successCB, out)
}

func (r *River) clientGetCachedMedia(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.ClientGetCachedMedia{}
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	res := repo.Files.GetCachedMedia()

	out.Constructor = msg.C_ClientCachedMediaInfo
	out.RequestID = in.RequestID
	out.Message, _ = res.Marshal()
	uiexec.ExecSuccessCB(successCB, out)
}

func (r *River) clientClearCachedMedia(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.ClientClearCachedMedia{}
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	if req.Peer != nil {
		repo.Files.DeleteCachedMediaByPeer(in.Team.ID, req.Peer.ID, int32(req.Peer.Type), req.MediaTypes)
	} else if len(req.MediaTypes) > 0 {
		repo.Files.DeleteCachedMediaByMediaType(req.MediaTypes)
	} else {
		repo.Files.ClearCache()
	}

	res := msg.Bool{Result: true}
	out.Constructor = msg.C_Bool
	out.RequestID = in.RequestID
	out.Message, _ = res.Marshal()
	uiexec.ExecSuccessCB(successCB, out)
}

func (r *River) clientGetMediaHistory(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.ClientGetMediaHistory{}
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	msgs, _ := repo.Messages.GetMediaHistory(req.MediaType)

	// get users && group IDs
	userIDs := domain.MInt64B{}
	groupIDs := domain.MInt64B{}
	for _, m := range msgs {
		if m.PeerType == int32(msg.PeerSelf) || m.PeerType == int32(msg.PeerUser) {
			userIDs[m.PeerID] = true
		}
		if m.PeerType == int32(msg.PeerGroup) {
			groupIDs[m.PeerID] = true
		}
		if m.SenderID > 0 {
			userIDs[m.SenderID] = true
		}
		if m.FwdSenderID > 0 {
			userIDs[m.FwdSenderID] = true
		}
	}

	users, _ := repo.Users.GetMany(userIDs.ToArray())
	groups, _ := repo.Groups.GetMany(groupIDs.ToArray())

	res := msg.MessagesMany{
		Messages:   msgs,
		Users:      users,
		Groups:     groups,
		Continuous: false,
	}

	out.Constructor = msg.C_MessagesMany
	out.RequestID = in.RequestID
	out.Message, _ = res.Marshal()
	uiexec.ExecSuccessCB(successCB, out)
}

func (r *River) clientGetLastBotKeyboard(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.ClientGetLastBotKeyboard{}
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	lastKeyboardMsg, _ := repo.Messages.GetLastBotKeyboard(in.Team.ID, req.Peer.ID, int32(req.Peer.Type))

	if lastKeyboardMsg == nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: "message not found"})
		successCB(out)
		return
	}

	out.Constructor = msg.C_UserMessage
	out.RequestID = in.RequestID
	out.Message, _ = lastKeyboardMsg.Marshal()
	uiexec.ExecSuccessCB(successCB, out)
}

func (r *River) clientGetRecentSearch(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.ClientGetRecentSearch{}
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	recentSearches := repo.RecentSearches.List(in.Team.ID, req.Limit)

	// get users && group IDs
	userIDs := domain.MInt64B{}
	groupIDs := domain.MInt64B{}
	for _, r := range recentSearches {
		if r.Peer.Type == int32(msg.PeerSelf) || r.Peer.Type == int32(msg.PeerUser) {
			userIDs[r.Peer.ID] = true
		}
		if r.Peer.Type == int32(msg.PeerGroup) {
			groupIDs[r.Peer.ID] = true
		}
	}

	users, _ := repo.Users.GetMany(userIDs.ToArray())
	groups, _ := repo.Groups.GetMany(groupIDs.ToArray())

	res := msg.RecentSearchMany{
		RecentSearches: recentSearches,
		Users:          users,
		Groups:         groups,
	}

	out.Constructor = msg.C_RecentSearchMany
	out.RequestID = in.RequestID
	out.Message, _ = res.Marshal()
	uiexec.ExecSuccessCB(successCB, out)
}

func (r *River) clientPutRecentSearch(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.ClientPutRecentSearch{}
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	peer := &msg.Peer{
		ID:         req.Peer.ID,
		Type:       int32(req.Peer.Type),
		AccessHash: req.Peer.AccessHash,
	}

	recentSearch := &msg.RecentSearch{
		Peer: peer,
		Date: int32(time.Now().Unix()),
	}

	err := repo.RecentSearches.Put(in.Team.ID, recentSearch)

	if err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	res := msg.Bool{
		Result: true,
	}

	out.Constructor = msg.C_Bool
	out.RequestID = in.RequestID
	out.Message, _ = res.Marshal()
	uiexec.ExecSuccessCB(successCB, out)
}

func (r *River) clientRemoveAllRecentSearches(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.ClientRemoveAllRecentSearches{}
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	err := repo.RecentSearches.Clear(in.Team.ID)

	if err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	res := msg.Bool{
		Result: true,
	}

	out.Constructor = msg.C_Bool
	out.RequestID = in.RequestID
	out.Message, _ = res.Marshal()
	uiexec.ExecSuccessCB(successCB, out)
}

func (r *River) clientRemoveRecentSearch(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.ClientRemoveRecentSearch{}
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	err := repo.RecentSearches.Delete(in.Team.ID, req.Peer)

	if err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	res := msg.Bool{
		Result: true,
	}

	out.Constructor = msg.C_Bool
	out.RequestID = in.RequestID
	out.Message, _ = res.Marshal()
	uiexec.ExecSuccessCB(successCB, out)
}

func (r *River) gifGetSaved(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.GifGetSaved{}
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}
	gifHash, _ := repo.System.LoadInt(domain.SkGifHash)

	var enqueSuccessCB domain.MessageHandler

	if gifHash != 0 {
		res, err := repo.Gifs.GetSaved()
		if err != nil {
			msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
			successCB(out)
			return
		}
		msg.ResultSavedGifs(out, res)
		successCB(out)

		// ignore success cb because we notify views on message hanlder
		enqueSuccessCB = func(m *msg.MessageEnvelope) {

		}
	} else {
		enqueSuccessCB = successCB
	}

	r.queueCtrl.EnqueueCommand(in, timeoutCB, enqueSuccessCB, true)
}

func (r *River) systemGetConfig(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	msg.ResultSystemConfig(out, domain.SysConfig)
	successCB(out)
}

func (r *River) getTeamCounters(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := &msg.GetTeamCounters{}
	if err := req.Unmarshal(in.Message); err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	unreadCount, mentionCount, err := repo.Dialogs.CountAllUnread(r.ConnInfo.UserID, req.Team.ID, req.WithMutes)

	if err != nil {
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	res := msg.TeamCounters{
		UnreadCount:  int64(unreadCount),
		MentionCount: int64(mentionCount),
	}

	out.Constructor = msg.C_TeamCounters
	out.RequestID = in.RequestID
	out.Message, _ = res.Marshal()
	uiexec.ExecSuccessCB(successCB, out)
}
