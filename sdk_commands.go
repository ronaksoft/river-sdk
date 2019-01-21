package riversdk

import (
	"encoding/json"
	"strings"

	"git.ronaksoftware.com/ronak/riversdk/filemanager"

	"git.ronaksoftware.com/ronak/riversdk/cmd"
	"git.ronaksoftware.com/ronak/riversdk/synchronizer"

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
		if d.PeerType == int32(msg.PeerUser) {
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
		if m.PeerType == int32(msg.PeerUser) {
			mUsers[m.SenderID] = true
		}
		if m.PeerType == int32(msg.PeerGroup) {
			mGroups[m.PeerID] = true
		}
		mUsers[m.FwdSenderID] = true

		// load MessageActionData users
		actUserIDs := domain.ExtractActionUserIDs(m.MessageAction, m.MessageActionData)
		for _, id := range actUserIDs {
			mUsers[id] = true
		}
	}
	res.Groups = repo.Ctx().Groups.GetManyGroups(mGroups.ToArray())
	res.Users = repo.Ctx().Users.GetAnyUsers(mUsers.ToArray())
	out.Constructor = msg.C_MessagesDialogs
	buff, err := res.Marshal()
	if err != nil {
		log.LOG_Debug("River::messagesGetDialogs()-> res.Marshal()", zap.Error(err))
	}
	out.Message = buff
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
	dbID := -domain.SequentialUniqueID()
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
	log.LOG_Warn("River::messageGetHistory()")
	req := new(msg.MessagesGetHistory)
	if err := req.Unmarshal(in.Message); err != nil {
		log.LOG_Warn("River::messageGetHistory()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
		return
	}

	dtoDialog := repo.Ctx().Dialogs.GetDialog(req.Peer.ID, int32(req.Peer.Type))
	if dtoDialog == nil {
		out.Constructor = msg.C_Error
		out.RequestID = in.RequestID
		msg.ResultError(out, &msg.Error{Code: "-1", Items: "dialog does not exist"})
		cmd.GetUIExecuter().Exec(func() {
			if successCB != nil {
				successCB(out)
			}
		})
		return
	}
	// Offline mode
	if r.networkCtrl.Quality() == domain.DISCONNECTED || r.networkCtrl.Quality() == domain.CONNECTING {
		if dtoDialog.TopMessageID < 0 {
			messages, users := repo.Ctx().Messages.GetMessageHistoryWithPendingMessages(req.Peer.ID, int32(req.Peer.Type), req.MinID, req.MaxID, req.Limit)
			fnSendGetMessageHistoryResponse(out, messages, users, in.RequestID, successCB)
		} else {
			messages, users := repo.Ctx().Messages.GetMessageHistoryWithMinMaxID(req.Peer.ID, int32(req.Peer.Type), req.MinID, req.MaxID, req.Limit)
			fnSendGetMessageHistoryResponse(out, messages, users, in.RequestID, successCB)
		}
		return
	}

	if req.MinID == 0 && req.MaxID == 0 {
		// Load type 0 : initial
		if dtoDialog.TopMessageID < 0 {
			log.LOG_Debug("River::messageGetHistory() Load Type 0 : from localDB PendedMessage")
			// fetch messages from localDB cuz there is a pending message it means we are not connected to server
			messages, users := repo.Ctx().Messages.GetMessageHistoryWithPendingMessages(req.Peer.ID, int32(req.Peer.Type), req.MinID, req.MaxID, req.Limit)
			fnSendGetMessageHistoryResponse(out, messages, users, in.RequestID, successCB)
		} else {

			maxID := dtoDialog.TopMessageID - 1

			holes := synchronizer.GetHoles(dtoDialog.PeerID, req.MinID, maxID)
			closestHole := synchronizer.GetMaxClosetHole(maxID, holes)
			if len(holes) > 0 {
				if closestHole != nil {
					messages, users := repo.Ctx().Messages.GetMessageHistoryWithMinMaxID(req.Peer.ID, int32(req.Peer.Type), closestHole.MaxID, dtoDialog.TopMessageID, req.Limit)
					if len(messages) == int(req.Limit) {
						log.LOG_Debug("River::messageGetHistory() Load Type 0:A1 from localDB")
						fnSendGetMessageHistoryResponse(out, messages, users, in.RequestID, successCB)
					} else {
						log.LOG_Debug("River::messageGetHistory() Load Type 0:A2 sent to server")
						r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
					}
				} else {
					log.LOG_Debug("River::messageGetHistory() Load Type 0:A3 sent to server")
					r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
				}
			} else {
				log.LOG_Debug("River::messageGetHistory() Load Type 0:A4 from localDB")
				messages, users := repo.Ctx().Messages.GetMessageHistoryWithMinMaxID(req.Peer.ID, int32(req.Peer.Type), req.MinID, req.MaxID, req.Limit)
				fnSendGetMessageHistoryResponse(out, messages, users, in.RequestID, successCB)
			}

		}

	} else if req.MinID == 0 && req.MaxID != 0 {
		// Load type 1 : scroll to up
		holes := synchronizer.GetHoles(dtoDialog.PeerID, req.MinID, req.MaxID)
		closestHole := synchronizer.GetMaxClosetHole(req.MaxID, holes)

		if len(holes) > 0 {
			if closestHole != nil {
				messages, users := repo.Ctx().Messages.GetMessageHistoryWithMinMaxID(req.Peer.ID, int32(req.Peer.Type), closestHole.MaxID, req.MaxID, req.Limit)
				if len(messages) == int(req.Limit) {
					log.LOG_Debug("River::messageGetHistory() Load Type 1:B1 from localDB")
					fnSendGetMessageHistoryResponse(out, messages, users, in.RequestID, successCB)
				} else {
					log.LOG_Debug("River::messageGetHistory() Load Type 1:B2 sent to server")
					r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
				}
			} else {
				log.LOG_Debug("River::messageGetHistory() Load Type 1:B3 sent to server")
				r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
			}
		} else {
			log.LOG_Debug("River::messageGetHistory() Load Type 1:B4 from localDB")
			messages, users := repo.Ctx().Messages.GetMessageHistoryWithMinMaxID(req.Peer.ID, int32(req.Peer.Type), req.MinID, req.MaxID, req.Limit)
			fnSendGetMessageHistoryResponse(out, messages, users, in.RequestID, successCB)
		}

	} else if req.MinID != 0 && req.MaxID == 0 {
		// Load type 2 : scroll to down
		maxID := dtoDialog.TopMessageID - 1

		holes := synchronizer.GetHoles(dtoDialog.PeerID, req.MinID, maxID)
		closestHole := synchronizer.GetMinClosetHole(req.MinID, holes)
		if len(holes) > 0 {
			if closestHole != nil {
				messages, users := repo.Ctx().Messages.GetMessageHistoryWithMinMaxID(req.Peer.ID, int32(req.Peer.Type), req.MinID, closestHole.MinID.Int64, req.Limit)
				if len(messages) == int(req.Limit) {
					log.LOG_Debug("River::messageGetHistory() Load Type 2:C1 from localDB")
					fnSendGetMessageHistoryResponse(out, messages, users, in.RequestID, successCB)
				} else {
					log.LOG_Debug("River::messageGetHistory() Load Type 2:C2 sent to server")
					r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
				}
			} else {
				log.LOG_Debug("River::messageGetHistory() Load Type 2:C3 sent to server")
				r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
			}
		} else {
			log.LOG_Debug("River::messageGetHistory() Load Type 2:C4 from localDB")
			messages, users := repo.Ctx().Messages.GetMessageHistoryWithMinMaxID(req.Peer.ID, int32(req.Peer.Type), req.MinID, req.MaxID, req.Limit)
			fnSendGetMessageHistoryResponse(out, messages, users, in.RequestID, successCB)
		}

	} else {
		// Load type 3 : exact size
		holes := synchronizer.GetHoles(dtoDialog.PeerID, req.MinID, req.MaxID)
		if len(holes) > 0 {
			log.LOG_Debug("River::messageGetHistory() Load Type 3:C1 sent to server")
			r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
		} else {
			log.LOG_Debug("River::messageGetHistory() Load Type 3:C2 from localDB")
			messages, users := repo.Ctx().Messages.GetMessageHistoryWithMinMaxID(req.Peer.ID, int32(req.Peer.Type), req.MinID, req.MaxID, req.Limit)
			fnSendGetMessageHistoryResponse(out, messages, users, in.RequestID, successCB)
		}
	}
}

func fnSendGetMessageHistoryResponse(out *msg.MessageEnvelope, messages []*msg.UserMessage, users []*msg.User, requestID uint64, successCB domain.MessageHandler) {
	res := new(msg.MessagesMany)
	res.Messages = messages
	res.Users = users
	// result
	out.RequestID = requestID
	out.Constructor = msg.C_MessagesMany
	out.Message, _ = res.Marshal()
	cmd.GetUIExecuter().Exec(func() {
		if successCB != nil {
			successCB(out)
		}
	})
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

func (r *River) contactsImport(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.ContactsImport)
	if err := req.Unmarshal(in.Message); err != nil {
		log.LOG_Debug("River::contactsImport()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	if req.Replace {
		for _, c := range req.Contacts {
			err := repo.Ctx().Users.UpdatePhoneContact(c)
			if err != nil {
				log.LOG_Debug("River::contactsImport()-> SaveParticipants()",
					zap.String("Error", err.Error()),
				)
			}
		}
	}

	// send the request to server
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
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

	messages := repo.Ctx().Messages.GetManyMessages(msgIDs.ToArray())
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
	users := repo.Ctx().Users.GetAnyUsers(mUsers.ToArray())

	// if db allready had all users
	if len(messages) == len(msgIDs) && len(users) > 0 {
		res := new(msg.MessagesMany)
		res.Users = make([]*msg.User, 0)
		res.Messages = messages
		res.Users = append(res.Users, users...)

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

	// Get PendingMessage and cancel its requests
	pendingMessageIDs := make([]int64, 0)
	for _, id := range req.MessageIDs {
		if id < 0 {
			pendingMessageIDs = append(pendingMessageIDs, id)
		}
	}
	if len(pendingMessageIDs) > 0 {
		// remove from queue
		pendedRequestIDs := repo.Ctx().PendingMessages.GetManyPendingMessagesRequestID(pendingMessageIDs)
		for _, reqID := range pendedRequestIDs {
			r.queueCtrl.CancelRequest(reqID)
		}
		// remove from DB
		err := repo.Ctx().PendingMessages.DeleteManyPendingMessage(pendingMessageIDs)
		if err != nil {
			log.LOG_Debug("River::messagesDelete()-> DeletePendingMessage()",
				zap.String("Error", err.Error()),
			)
		}
	}

	// remove message
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
	user := repo.Ctx().Users.GetUser(req.User.UserID)
	if user != nil {
		gp := &msg.GroupParticipant{
			AccessHash: req.User.AccessHash,
			FirstName:  user.FirstName,
			LastName:   user.LastName,
			UserID:     req.User.UserID,
			Type:       msg.ParticipantTypeMember,
		}
		err := repo.Ctx().Groups.SaveParticipants(req.GroupID, gp)
		// TODO : Increase group ParticipantCount
		if err != nil {
			log.LOG_Debug("River::groupAddUser()-> SaveParticipants()",
				zap.String("Error", err.Error()),
			)
		}
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
	dlg := repo.Ctx().Dialogs.GetDialog(req.GroupID, int32(msg.PeerGroup))
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

func (r *River) groupUpdateAdmin(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.GroupsUpdateAdmin)
	if err := req.Unmarshal(in.Message); err != nil {
		log.LOG_Debug("River::groupUpdateAdmin()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	err := repo.Ctx().Groups.UpdateGroupMemberType(req.GroupID, req.User.UserID, req.Admin)
	// TODO : Decrease group ParticipantCount
	if err != nil {
		log.LOG_Debug("River::groupUpdateAdmin()-> UpdateGroupMemberType()",
			zap.String("Error", err.Error()),
		)
	}

	// send the request to server
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
}

func (r *River) clientSendMessageMedia(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	// send Media
	log.LOG_Debug("River::clientSendMessageMedia()")
	reqMedia := new(msg.ClientSendMessageMedia)
	if err := reqMedia.Unmarshal(in.Message); err != nil {
		log.LOG_Debug("River::clientSendMessageMedia()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
		return
	}

	// support IOS file path
	if strings.HasPrefix(reqMedia.FilePath, "file://") {
		reqMedia.FilePath = reqMedia.FilePath[7:]
		in.Message, _ = reqMedia.Marshal()
	}

	// TODO : check if file has been uploaded b4
	dtoFile := repo.Ctx().Files.GetExistingFileDocument(reqMedia.FilePath)
	fileAlreadyUploaded := false
	if dtoFile != nil {
		fileAlreadyUploaded = dtoFile.ClusterID > 0 && dtoFile.DocumentID > 0 && dtoFile.AccessHash > 0
	}
	if fileAlreadyUploaded {
		msgID := -domain.SequentialUniqueID()
		fileID := uint64(domain.SequentialUniqueID())
		res, err := repo.Ctx().PendingMessages.SaveClientMessageMedia(msgID, r.ConnInfo.UserID, int64(fileID), reqMedia)

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

		//TODO :FIX THIS : required behaviour add to pending message and put this to progress pipe line and etc ...
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

		r.onFileUploadCompleted(msgID, int64(fileID), dtoFile.ClusterID, -1, reqMedia)
		// send the request to server
		r.queueCtrl.ExecuteCommand(fileID, in.Constructor, in.Message, nil, nil, false)

	} else {
		// 1. insert into pending messages, id is negative nano timestamp and save RandomID too : Done
		msgID := -domain.SequentialUniqueID()
		fileID := int64(in.RequestID)
		res, err := repo.Ctx().PendingMessages.SaveClientMessageMedia(msgID, r.ConnInfo.UserID, fileID, reqMedia)

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

		// 2. TODO : start file upload and send process
		err = filemanager.Ctx().Upload(fileID, res)
		if err != nil {
			e := new(msg.Error)
			e.Code = "n/a"
			e.Items = "Failed to start Upload : " + err.Error()
			msg.ResultError(out, e)
			cmd.GetUIExecuter().Exec(func() {
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
		//   exist in pendindTable we remove it and insert new message with new id to message table
		//   invoke new OnUpdate with new protobuff to inform ui that pending message got delivered
		cmd.GetUIExecuter().Exec(func() {
			if successCB != nil {
				successCB(out)
			}
		}) //successCB(out)
	}

}

func (r *River) messagesReadContents(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	req := new(msg.MessagesReadContents)
	if err := req.Unmarshal(in.Message); err != nil {
		log.LOG_Debug("River::messagesReadContents()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
		msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
		successCB(out)
		return
	}

	err := repo.Ctx().Messages.SetContentRead(req.MessageIDs)
	if err != nil {
		log.LOG_Debug("River::groupUpdateAdmin()-> UpdateGroupMemberType()",
			zap.String("Error", err.Error()),
		)
	}

	// send the request to server
	r.queueCtrl.ExecuteCommand(in.RequestID, in.Constructor, in.Message, timeoutCB, successCB, true)
}

func (r *River) messagesSendMedia(in, out *msg.MessageEnvelope, timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler) {
	log.LOG_Debug("River::messagesSendMedia()")

	req := new(msg.MessagesSendMedia)
	if err := req.Unmarshal(in.Message); err != nil {
		log.LOG_Debug("River::messagesSendMedia()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
		return
	}

	// this will be used as next requestID
	req.RandomID = domain.SequentialUniqueID()

	// 1. insert into pending messages, id is negative nano timestamp and save RandomID too : Done
	dbID := -domain.SequentialUniqueID()
	res, err := repo.Ctx().PendingMessages.SaveMessageMedia(dbID, r.ConnInfo.UserID, req)

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

	requestBytes, _ := req.Marshal()
	r.queueCtrl.ExecuteCommand(uint64(req.RandomID), msg.C_MessagesSendMedia, requestBytes, timeoutCB, successCB, true)

	// 3. return to CallBack with pending message data : Done
	out.Constructor = msg.C_ClientPendingMessage
	out.Message, _ = res.Marshal()

	cmd.GetUIExecuter().Exec(func() {
		if successCB != nil {
			successCB(out)
		}
	}) //successCB(out)

}
