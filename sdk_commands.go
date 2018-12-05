package riversdk

import (
	"encoding/json"
	"strings"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/cmd"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/log"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/repo"
	"go.uber.org/zap"
)

func (r *River) messagesGetDialogs(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	log.LOG_Debug("River::messagesGetDialogs()")
	req := new(msg.MessagesGetDialogs)
	if err := req.Unmarshal(in.Message); err != nil {
		log.LOG_Debug("River::messagesGetDialogs()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
		return
	}
	res := new(msg.MessagesDialogs)
	res.Dialogs = repo.Ctx().Dialogs.GetDialogs(req.Offset, req.Limit)

	// if the localDB had no data send the request to server
	if len(res.Dialogs) == 0 {
		log.LOG_Debug("River::messagesGetDialogs()-> GetDialogs() nothing found in cacheDB pass request to server")
		r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
		return
	}

	mUsers := domain.MInt64B{}
	mGroups := domain.MInt64B{}
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

	res.Messages = repo.Ctx().Messages.GetManyMessages(mMessages.ToArray())

	//Load Pending messages
	pendingMessages := repo.Ctx().PendingMessages.GetManyPendingMessages(mPendingMesssage.ToArray())
	res.Messages = append(res.Messages, pendingMessages...)

	for _, m := range res.Messages {
		if m.PeerType == int32(msg.PeerType_PeerUser) {
			mUsers[m.SenderID] = true
		}
		if m.PeerType == int32(msg.PeerType_PeerGroup) {
			mGroups[m.PeerID] = true
		}
	}
	res.Groups = repo.Ctx().Groups.GetManyGroups(mGroups.ToArray())
	res.Users = repo.Ctx().Users.GetAnyUsers(mUsers.ToArray())
	out.Constructor = msg.C_MessagesDialogs
	out.Message, _ = res.Marshal()

	cmd.GetUIExecuter().Exec(func() {
		if successCB != nil {
			successCB(out)
		}
	}) //successCB(out)
}

func (r *River) messagesGetDialog(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	log.LOG_Debug("River::messagesGetDialog()")
	req := new(msg.MessagesGetDialog)
	if err := req.Unmarshal(in.Message); err != nil {
		log.LOG_Debug("River::messagesGetDialog()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
		return
	}
	res := new(msg.Dialog)
	res = repo.Ctx().Dialogs.GetDialog(req.Peer.ID, int32(req.Peer.Type))

	// if the localDB had no data send the request to server
	if res == nil {
		log.LOG_Debug("River::messagesGetDialog()-> GetDialog() nothing found in cacheDB pass request to server")
		r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
		return
	}

	out.Constructor = msg.C_Dialog
	out.Message, _ = res.Marshal()

	cmd.GetUIExecuter().Exec(func() {
		if successCB != nil {
			successCB(out)
		}
	}) //successCB(out)
}

func (r *River) messagesSend(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	log.LOG_Debug("River::messagesSend()")
	req := new(msg.MessagesSend)
	if err := req.Unmarshal(in.Message); err != nil {
		log.LOG_Debug("River::messagesSend()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
		return
	}

	// do not allow empty message
	if strings.TrimSpace(req.Body) == "" {
		e := new(msg.Error)
		e.Code = "n/a"
		e.Items = "empty message is not allowed"
		msg.ResultError(out, e)
		cmd.GetUIExecuter().Exec(func() {
			if successCB != nil {
				successCB(out)
			}
		})
		return
	}

	// this will be used as next requestID
	req.RandomID = domain.SequentialUniqueID()

	// TODO :
	// 0. fix database and add PendingMessages table : Done

	// 1. insert into pending messages, id is negative nano timestamp and save RandomID too : Done
	dbID := -time.Now().UnixNano()
	res, err := repo.Ctx().PendingMessages.Save(dbID, r.ConnInfo.UserID, req)

	if err != nil {
		e := new(msg.Error)
		e.Code = "n/a"
		e.Items = "Failed to save to pendingMessages : " + err.Error()
		msg.ResultError(out, e)
		cmd.GetUIExecuter().Exec(func() {
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
	//   exist in pendindTable we remove it and insert new message with new id to message table
	//   invoke new OnUpdate with new protobuff to inform ui that pending message got delivered
	cmd.GetUIExecuter().Exec(func() {
		if successCB != nil {
			successCB(out)
		}
	}) //successCB(out)
}

func (r *River) messageGetHistory(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	log.LOG_Debug("River::messageGetHistory()")
	req := new(msg.MessagesGetHistory)
	if err := req.Unmarshal(in.Message); err != nil {
		log.LOG_Debug("River::messageGetHistory()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
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
		log.LOG_Debug("River::messageGetHistory()-> GetMessageHistory() nothing found in cacheDB pass request to server")
		r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
		return
	}

	// result
	out.RequestID = in.RequestID
	out.Constructor = msg.C_MessagesMany
	out.Message, _ = res.Marshal()

	cmd.GetUIExecuter().Exec(func() {
		if successCB != nil {
			successCB(out)
		}
	}) //successCB(out)
}

func (r *River) contactGet(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	log.LOG_Debug("River::contactGet()")
	req := new(msg.ContactsGet)
	if err := req.Unmarshal(in.Message); err != nil {
		log.LOG_Debug("River::contactGet()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
		return
	}

	res := new(msg.ContactsMany)
	res.Users, res.Contacts = repo.Ctx().Users.GetContacts()

	// if didn't find anything send request to server
	if len(res.Users) == 0 || len(res.Contacts) == 0 {
		r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
		return
	}

	out.Constructor = msg.C_ContactsMany
	out.Message, _ = res.Marshal()

	cmd.GetUIExecuter().Exec(func() {
		if successCB != nil {
			successCB(out)
		}
	}) //successCB(out)
}

func (r *River) messageReadHistory(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	log.LOG_Debug("River::messageReadHistory()")
	req := new(msg.MessagesReadHistory)
	if err := req.Unmarshal(in.Message); err != nil {
		log.LOG_Debug("River::messageReadHistory()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
	}

	dialog := repo.Ctx().Dialogs.GetDialog(req.Peer.ID, int32(req.Peer.Type))
	if dialog == nil {
		return
	}
	if dialog.ReadInboxMaxID > req.MaxID {
		return
	}

	err := repo.Ctx().Dialogs.UpdateReadInboxMaxID(r.ConnInfo.UserID, req.Peer.ID, int32(req.Peer.Type), req.MaxID)
	if err != nil {
		log.LOG_Debug("River::messageReadHistory()-> UpdateReadInboxMaxID()",
			zap.String("Error", err.Error()),
		)
	}
	// send the request to server

	log.LOG_Debug("River::messageReadHistory() Pass request to server")
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
}

func (r *River) usersGet(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	log.LOG_Debug("River::usersGet()")
	req := new(msg.UsersGet)
	if err := req.Unmarshal(in.Message); err != nil {
		log.LOG_Debug("River::usersGet()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
		return
	}
	userIDs := domain.MInt64B{}
	for _, v := range req.Users {
		userIDs[v.UserID] = true
	}

	users := repo.Ctx().Users.GetAnyUsers(userIDs.ToArray())

	if len(users) == len(userIDs) {
		res := new(msg.UsersMany)
		res.Users = users

		out.Constructor = msg.C_UsersMany
		out.Message, _ = res.Marshal()
		cmd.GetUIExecuter().Exec(func() {
			if successCB != nil {
				successCB(out)
			}
		}) //successCB(out)
	} else {
		log.LOG_Debug("River::messageGetHistory()-> GetAnyUsers() cacheDB is not up to date pass request to server")
		// send the request to server
		r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
	}
}

func (r *River) messagesGet(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	log.LOG_Debug("River::messagesGet()")
	req := new(msg.MessagesGet)
	if err := req.Unmarshal(in.Message); err != nil {
		log.LOG_Debug("River::messagesGet()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
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
		cmd.GetUIExecuter().Exec(func() {
			if successCB != nil {
				successCB(out)
			}
		}) //successCB(out)
	} else {
		log.LOG_Debug("River::messagesGet() -> GetManyMessages() cacheDB is not up to date pass request to server")
		// send the request to server
		r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
	}
}

func (r *River) accountUpdateUsername(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	log.LOG_Debug("River::accountUpdateUsername()")
	req := new(msg.AccountUpdateUsername)
	if err := req.Unmarshal(in.Message); err != nil {
		log.LOG_Debug("River::accountUpdateUsername()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
		return
	}

	r.ConnInfo.Username = req.Username
	r.ConnInfo.saveConfig()

	log.LOG_Debug("River::accountUpdateUsername()-> pass request to server")
	// send the request to server
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
}

func (r *River) accountUpdateProfile(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	log.LOG_Debug("River::accountUpdateProfile()")
	req := new(msg.AccountUpdateProfile)
	if err := req.Unmarshal(in.Message); err != nil {
		log.LOG_Debug("River::accountUpdateProfile()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
		return
	}

	r.ConnInfo.FirstName = req.FirstName
	r.ConnInfo.LastName = req.LastName
	r.ConnInfo.saveConfig()

	log.LOG_Debug("River::accountUpdateProfile()-> pass request to server")
	// send the request to server
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
}

// save token to DB
func (r *River) accountRegisterDevice(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.AccountRegisterDevice)
	if err := req.Unmarshal(in.Message); err != nil {
		log.LOG_Debug("River::accountRegisterDevice()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
		return
	}
	r.DeviceToken = req

	val, err := json.Marshal(req)
	if err != nil {
		log.LOG_Debug("River::accountRegisterDevice()-> Json Marshal()",
			zap.String("Error", err.Error()),
		)
		return
	}
	err = repo.Ctx().System.SaveString(domain.CN_DEVICE_TOKEN, string(val))
	if err != nil {
		log.LOG_Debug("River::accountRegisterDevice()-> SaveString()",
			zap.String("Error", err.Error()),
		)
		return
	}
	// send the request to server
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
}

// remove token from DB
func (r *River) accountUnregisterDevice(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.AccountUnregisterDevice)
	if err := req.Unmarshal(in.Message); err != nil {
		log.LOG_Debug("River::accountUnregisterDevice()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
		return
	}
	r.DeviceToken = new(msg.AccountRegisterDevice)

	val, err := json.Marshal(r.DeviceToken)
	if err != nil {
		log.LOG_Debug("River::accountUnregisterDevice()-> Json Marshal()",
			zap.String("Error", err.Error()),
		)
		return
	}
	err = repo.Ctx().System.SaveString(domain.CN_DEVICE_TOKEN, string(val))
	if err != nil {
		log.LOG_Debug("River::accountUnregisterDevice()-> SaveString()",
			zap.String("Error", err.Error()),
		)
		return
	}
	// send the request to server
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
}

func (r *River) accountSetNotifySettings(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.AccountSetNotifySettings)
	if err := req.Unmarshal(in.Message); err != nil {
		log.LOG_Debug("River::accountSetNotifySettings()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
		return
	}

	dialog := repo.Ctx().Dialogs.GetDialog(req.Peer.ID, int32(req.Peer.Type))
	if dialog == nil {
		log.LOG_Debug("River::accountSetNotifySettings()-> GetDialog()",
			zap.String("Error", "Dialog is null"),
		)
		return
	}

	dialog.NotifySettings = req.Settings
	err := repo.Ctx().Dialogs.SaveDialog(dialog, 0)
	if err != nil {
		log.LOG_Debug("River::accountSetNotifySettings()-> SaveDialog()",
			zap.String("Error", err.Error()),
		)
		return
	}

	// send the request to server
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)

}

func (r *River) groupsEditTitle(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.GroupsEditTitle)
	if err := req.Unmarshal(in.Message); err != nil {
		log.LOG_Debug("River::messagesEditGroupTitle()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	err := repo.Ctx().Groups.UpdateGroupTitle(req.GroupID, req.Title)
	if err != nil {
		log.LOG_Debug("River::messagesEditGroupTitle()-> UpdateGroupTitle()",
			zap.String("Error", err.Error()),
		)
	}

	// send the request to server
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)

}

func (r *River) messagesClearHistory(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesClearHistory)
	if err := req.Unmarshal(in.Message); err != nil {
		log.LOG_Debug("River::messagesClearHistory()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}
	// this will be handled on mesage update appliers too
	err := repo.Ctx().Messages.DeleteDialogMessage(req.Peer.ID, int32(req.Peer.Type), req.MaxID)
	if err != nil {
		log.LOG_Debug("River::messagesClearHistory()-> DeleteDialogMessage()",
			zap.String("Error", err.Error()),
		)
	}

	// send the request to server
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)

}

func (r *River) messagesDelete(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesDelete)
	if err := req.Unmarshal(in.Message); err != nil {
		log.LOG_Debug("River::messagesDelete()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	err := repo.Ctx().Messages.DeleteMany(req.MessageIDs)
	if err != nil {
		log.LOG_Debug("River::messagesDelete()-> DeleteMany()",
			zap.String("Error", err.Error()),
		)
	}

	// send the request to server
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)

}

func (r *River) groupAddUser(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.GroupsAddUser)
	if err := req.Unmarshal(in.Message); err != nil {
		log.LOG_Debug("River::groupAddUser()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	gp := &msg.GroupParticipant{
		Date:      time.Now().Unix(),
		InviterID: r.ConnInfo.UserID,
		Type:      msg.ParticipantType_Member,
		UserID:    req.User.UserID,
	}
	err := repo.Ctx().Groups.SaveParticipants(req.GroupID, gp)
	// TODO : Increase group ParticipantCount
	if err != nil {
		log.LOG_Debug("River::groupAddUser()-> SaveParticipants()",
			zap.String("Error", err.Error()),
		)
	}

	// send the request to server
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)

}

func (r *River) groupDeleteUser(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.GroupsDeleteUser)
	if err := req.Unmarshal(in.Message); err != nil {
		log.LOG_Debug("River::groupDeleteUser()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	err := repo.Ctx().Groups.DeleteGroupMember(req.GroupID, req.User.UserID)
	// TODO : Decrease group ParticipantCount
	if err != nil {
		log.LOG_Debug("River::groupDeleteUser()-> SaveParticipants()",
			zap.String("Error", err.Error()),
		)
	}

	// send the request to server
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
}

// Warn : until implementation of GroupFull.Participants we dont get info of users that added to group when group created
func (r *River) groupsGetFull(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.GroupsGetFull)
	if err := req.Unmarshal(in.Message); err != nil {
		log.LOG_Debug("River::groupsGetFull()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	res := new(msg.GroupFull)
	// Group
	group, err := repo.Ctx().Groups.GetGroup(req.GroupID)
	if err != nil {
		log.LOG_Warn("River::groupsGetFull()-> GetGroup() Sending Request To Server !!!",
			zap.String("Error", err.Error()),
		)
		r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
		return
	}
	res.Group = group

	// Participants
	participents, err := repo.Ctx().Groups.GetParticipants(req.GroupID)
	if err != nil {
		log.LOG_Warn("River::groupsGetFull()-> GetParticipants() Sending Request To Server !!!",
			zap.String("Error", err.Error()),
		)
		r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
		return
	}
	res.Participants = participents

	// NotifySettings
	dlg := repo.Ctx().Dialogs.GetDialog(req.GroupID, int32(msg.PeerType_PeerGroup))
	if dlg == nil {
		log.LOG_Warn("River::groupsGetFull()-> GetDialog() Sending Request To Server !!!")

		r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
		return
	}
	res.NotifySettings = dlg.NotifySettings

	// Users
	userIDs := domain.MInt64B{}
	for _, v := range participents {
		userIDs[v.UserID] = true
	}
	users := repo.Ctx().Users.GetAnyUsers(userIDs.ToArray())
	if users == nil || len(participents) != len(users) || len(users) <= 0 {
		log.LOG_Warn("River::groupsGetFull()-> GetAnyUsers() Sending Request To Server !!!",
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
