package riversdk

import (
	"fmt"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/cmd"
	"git.ronaksoftware.com/ronak/riversdk/configs"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/log"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/repo"
	"go.uber.org/zap"
)

func (r *River) messagesGetDialogs(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesGetDialogs)
	if err := req.Unmarshal(in.Message); err != nil {
		log.LOG.Debug(err.Error())
		return
	}
	res := new(msg.MessagesDialogs)
	res.Dialogs = repo.Ctx().Dialogs.GetDialogs(req.Offset, req.Limit)

	// if the localDB had no data send the request to server
	if len(res.Dialogs) == 0 {
		r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB)
		return
	}

	mUsers := domain.MInt64B{}
	mMessages := domain.MInt64B{}
	mPendingMesssage := domain.MInt64B{}
	for _, d := range res.Dialogs {
		if d.PeerType == int32(msg.PeerType_PeerUser) {
			mUsers[d.PeerID] = true
		}
		mMessages[d.TopMessageID] = true
		if d.TopMessageID < 0 {
			mPendingMesssage[d.TopMessageID] = true
		}
	}
	log.LOG.Debug("Messages",
		zap.Int64s("MessageIDs", mMessages.ToArray()),
		zap.Int64s("Users", mUsers.ToArray()),
	)
	res.Messages = repo.Ctx().Messages.GetManyMessages(mMessages.ToArray())

	//Load Pending messages
	pendingMessages := repo.Ctx().PendingMessages.GetManyPendingMessages(mPendingMesssage.ToArray())
	res.Messages = append(res.Messages, pendingMessages...)

	for _, m := range res.Messages {
		mUsers[m.SenderID] = true
	}
	res.Users = repo.Ctx().Users.GetManyUsers(mUsers.ToArray())
	out.Constructor = msg.C_MessagesDialogs
	out.Message, _ = res.Marshal()

	cmd.GetUIExecuter().Exec(func() { successCB(out) }) //successCB(out)
}

func (r *River) messagesSend(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesSend)
	if err := req.Unmarshal(in.Message); err != nil {
		log.LOG.Debug(err.Error())
		return
	}
	// TODO :
	// 0. fix database and add PendingMessages table : Done

	// 1. insert into pending messages, id is negative nano timestamp and save RandomID too : Done
	dbID := -time.Now().UnixNano()
	res, err := repo.Ctx().PendingMessages.Save(dbID, configs.Get().UserID, req)

	if err != nil {
		e := new(msg.Error)
		e.Code = "n/a"
		e.Items = "Failed to save to pendingMessages : " + err.Error()
		msg.ResultError(out, e)
		return
	}
	// 2. add to queue [ looks like there is general queue to send messages ] : Done
	requestBytes, _ := req.Marshal()
	// r.queueCtrl.addToWaitingList(req) // this needs to be wrapped by MessageEnvelope
	// using req randomID as requestID later in queue processing and network controller messageHandler
	r.queueCtrl.ExecuteCommand(uint64(req.RandomID), msg.C_MessagesSend, requestBytes, timeoutCB, successCB)

	// 3. return to CallBack with pending message data : Done
	out.Constructor = msg.C_ClientPendingMessage
	out.Message, _ = res.Marshal()

	// 4. later when queue got processed and server returned response we should check if the requestID
	//   exist in pendindTable we remove it and insert new message with new id to message table
	//   invoke new OnUpdate with new protobuff to inform ui that pending message got delivered
	cmd.GetUIExecuter().Exec(func() { successCB(out) }) //successCB(out)
}

func (r *River) messageGetHistory(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {

	req := new(msg.MessagesGetHistory)
	if err := req.Unmarshal(in.Message); err != nil {
		log.LOG.Debug(err.Error())
		return
	}

	// log not required for now

	res := new(msg.MessagesMany)

	// fetch messages
	messages, users := repo.Ctx().Messages.GetMessageHistory(req.Peer.ID, int32(req.Peer.Type), req.MinID, req.MaxID, req.Limit)
	res.Messages = messages
	res.Users = users

	var maxMessageID int64
	for _, v := range messages {
		if maxMessageID < v.ID {
			maxMessageID = v.ID
		}
	}

	// if the localDB had no or outdated data send the request to server
	if (req.MaxID-maxMessageID) > int64(req.Limit) || len(res.Messages) <= 1 {
		r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB)
		return
	}

	// result
	out.RequestID = in.RequestID
	out.Constructor = msg.C_MessagesMany
	out.Message, _ = res.Marshal()

	cmd.GetUIExecuter().Exec(func() { successCB(out) }) //successCB(out)
}

func (r *River) contactGet(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.ContactsGet)
	if err := req.Unmarshal(in.Message); err != nil {
		log.LOG.Debug(err.Error())
		return
	}

	res := new(msg.ContactsMany)
	res.Users, res.Contacts = repo.Ctx().Users.GetContacts()

	// if didn't find anything send request to server
	if len(res.Users) == 0 || len(res.Contacts) == 0 {
		r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB)
		return
	}

	out.Constructor = msg.C_ContactsMany
	out.Message, _ = res.Marshal()

	cmd.GetUIExecuter().Exec(func() { successCB(out) }) // successCB(out)
}

func (r *River) messageReadHistory(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesReadHistory)
	req.Unmarshal(in.Message)

	dialog := repo.Ctx().Dialogs.GetDialog(req.Peer.ID, int32(req.Peer.Type))
	if dialog == nil {
		return
	}
	if dialog.ReadInboxMaxID > req.MaxID {
		return
	}

	repo.Ctx().Dialogs.UpdateReadInboxMaxID(configs.Get().UserID, req.Peer.ID, int32(req.Peer.Type), req.MaxID)

	// send the request to server
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB)

}

func (r *River) usersGet(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.UsersGet)
	err := req.Unmarshal(in.Message)
	if err != nil {
		log.LOG.Debug(err.Error())
		return
	}
	userIDs := domain.MInt64B{}
	for _, v := range req.Users {
		userIDs[v.UserID] = true
	}

	users := repo.Ctx().Users.GetAnyUsers(userIDs.ToArray())
	// if db allready had all users

	log.LOG.Debug("SDk_Commands::usersGet()",
		zap.Int("userID Count", len(userIDs)),
		zap.Int("users Count", len(users)),
		zap.String("Users", fmt.Sprint(users)),
	)

	if len(users) == len(userIDs) {
		res := new(msg.UsersMany)
		res.Users = users

		out.Constructor = msg.C_UsersMany
		out.Message, _ = res.Marshal()
		cmd.GetUIExecuter().Exec(func() { successCB(out) }) //successCB(out)
	} else {
		log.LOG.Debug("SDk_Commands::usersGet() send request to server")
		// send the request to server
		r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB)
	}
}

func (r *River) messagesGet(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesGet)
	err := req.Unmarshal(in.Message)
	if err != nil {
		log.LOG.Debug(err.Error())
		return
	}
	msgIDs := domain.MInt64B{}
	for _, v := range req.MessagesIDs {
		msgIDs[v] = true
	}

	user := repo.Ctx().Users.GetUser(req.Peer.ID)
	messages := repo.Ctx().Messages.GetManyMessages(msgIDs.ToArray())

	// if db allready had all users
	if len(messages) == len(msgIDs) && user != nil {
		res := new(msg.MessagesMany)
		res.Users = make([]*msg.User, 0)
		res.Messages = messages
		res.Users = append(res.Users, user)

		out.Constructor = msg.C_MessagesMany
		out.Message, _ = res.Marshal()
		cmd.GetUIExecuter().Exec(func() { successCB(out) }) //successCB(out)
	} else {
		// send the request to server
		r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB)
	}
}
