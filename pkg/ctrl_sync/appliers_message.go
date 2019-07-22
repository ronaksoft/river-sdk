package syncCtrl

import (
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	messageHole "git.ronaksoftware.com/ronak/riversdk/pkg/message_hole"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"git.ronaksoftware.com/ronak/riversdk/pkg/uiexec"
	"go.uber.org/zap"
	"time"
)

// authAuthorization
func (ctrl *Controller) authAuthorization(e *msg.MessageEnvelope) {
	x := new(msg.AuthAuthorization)
	if err := x.Unmarshal(e.Message); err != nil {
		logs.Warn("authAuthorization()-> Unmarshal()", zap.Error(err))
		return
	}
	logs.Info("SyncController::authAuthorization",
		zap.String("FirstName", x.User.FirstName),
		zap.String("LastName", x.User.LastName),
		zap.Int64("UserID", x.User.ID),
		zap.String("Bio", x.User.Bio),
		zap.String("Username", x.User.Username),
	)
	ctrl.connInfo.ChangeFirstName(x.User.FirstName)
	ctrl.connInfo.ChangeLastName(x.User.LastName)
	ctrl.connInfo.ChangeUserID(x.User.ID)
	ctrl.connInfo.ChangeBio(x.User.Bio)
	ctrl.connInfo.ChangeUsername(x.User.Username)
	ctrl.connInfo.Save()

	ctrl.SetUserID(x.User.ID)

	go ctrl.sync()
}

// authSentCode
func (ctrl *Controller) authSentCode(e *msg.MessageEnvelope) {
	x := new(msg.AuthSentCode)
	if err := x.Unmarshal(e.Message); err != nil {
		logs.Error("authSentCode()-> Unmarshal()", zap.Error(err))
		return
	}
	logs.Info("SyncController::authSentCode")
	ctrl.connInfo.ChangePhone(x.Phone)
	// No need to save it here its gonna be saved on authAuthorization
	// conf.save()
}

// contactsImported
func (ctrl *Controller) contactsImported(e *msg.MessageEnvelope) {
	x := new(msg.ContactsImported)
	if err := x.Unmarshal(e.Message); err != nil {
		logs.Error("contactsImported()-> Unmarshal()", zap.Error(err))
		return
	}
	logs.Info("SyncController::contactsImported")
	for _, u := range x.Users {
		repo.Users.SaveContact(u)
	}
}

// contactsMany
func (ctrl *Controller) contactsMany(e *msg.MessageEnvelope) {
	x := new(msg.ContactsMany)
	if err := x.Unmarshal(e.Message); err != nil {
		logs.Error("contactsMany()-> Unmarshal()", zap.Error(err))
		return
	}
	logs.Info("SyncController::contactsMany",
		zap.Int("Users", len(x.Users)),
		zap.Int("Contacts", len(x.Contacts)),
	)

	userIDs := domain.MInt64B{}
	for _, u := range x.Users {
		userIDs[u.ID] = true
		repo.Users.SaveContact(u)
	}
	// server
	if len(userIDs) > 0 {
		// calculate contactsGethash and save
		crc32Hash := domain.CalculateContactsGetHash(userIDs.ToArray())
		err := repo.System.SaveInt(domain.ColumnContactsGetHash, int32(crc32Hash))
		if err != nil {
			logs.Error("contactsMany() failed to save ContactsGetHash to DB", zap.Error(err))
		}
	}
}

// messageDialogs
func (ctrl *Controller) messagesDialogs(e *msg.MessageEnvelope) {
	x := new(msg.MessagesDialogs)
	if err := x.Unmarshal(e.Message); err != nil {
		logs.Error("messagesDialogs()-> Unmarshal()", zap.Error(err))
		return
	}
	logs.Info("SyncController::messagesDialogs",
		zap.Int("Dialogs", len(x.Dialogs)),
		zap.Int64("UpdateID", x.UpdateID),
		zap.Int32("Count", x.Count),
	)

	mMessages := make(map[int64]*msg.UserMessage)
	for _, message := range x.Messages {
		repo.Messages.Save(message)
		mMessages[message.ID] = message
	}
	for _, dialog := range x.Dialogs {
		topMessage, _ := mMessages[dialog.TopMessageID]
		if topMessage == nil {
			logs.Error("Top Message Is Nil",
				zap.Int64("MessageID", dialog.TopMessageID),
			)
			continue
		}
		repo.Dialogs.SaveNew(dialog, topMessage.CreatedOn)
	}
	for _, user := range x.Users {
		repo.Users.Save(user)
	}
	for _, group := range x.Groups {
		repo.Groups.Save(group)
	}

	logs.Debug("SyncController::messagesDialogs()",
		zap.Int("DialogsCount", len(x.Dialogs)),
		zap.Int("GroupsCount", len(x.Groups)),
		zap.Int("UsersCount", len(x.Users)),
		zap.Int("MessagesCount", len(x.Messages)),
	)
}

// Check pending messages and notify UI
func (ctrl *Controller) messageSent(e *msg.MessageEnvelope) {
	sent := new(msg.MessagesSent)
	err := sent.Unmarshal(e.Message)
	if err != nil {
		logs.Error("MessageSent()-> Unmarshal", zap.Error(err))
		return
	}
	logs.Info("SyncController::messageSent",
		zap.Int64("MessageID", sent.MessageID),
		zap.Int64("RandomID", sent.RandomID),
	)

	userMessage := repo.Messages.Get(sent.MessageID)
	if userMessage != nil {
		// If we are here, it means we receive UpdateNewMessage before UpdateMessageID / MessagesSent
		pm, _ := repo.PendingMessages.GetByRandomID(int64(e.RequestID))
		if pm == nil {
			return
		}
		// It means we have received the NewMessage
		update := new(msg.UpdateMessagesDeleted)
		update.Peer = &msg.Peer{ID: pm.PeerID, Type: pm.PeerType}
		update.MessageIDs = []int64{pm.ID}
		bytes, _ := update.Marshal()

		updateEnvelope := new(msg.UpdateEnvelope)
		updateEnvelope.Constructor = msg.C_UpdateMessagesDeleted
		updateEnvelope.Update = bytes
		updateEnvelope.UpdateID = 0
		updateEnvelope.Timestamp = time.Now().Unix()

		buff, _ := updateEnvelope.Marshal()

		// call external handler
		uiexec.Ctx().Exec(func() {
			if ctrl.onUpdateMainDelegate != nil {
				ctrl.onUpdateMainDelegate(msg.C_UpdateEnvelope, buff)
			}
		})
		return
	}

	_, err = repo.PendingMessages.GetByRandomID(int64(e.RequestID))
	if err != nil {
		return
	}
	repo.PendingMessages.SaveByRealID(int64(e.RequestID), sent.MessageID)
}

// usersMany
func (ctrl *Controller) usersMany(e *msg.MessageEnvelope) {
	u := new(msg.UsersMany)
	err := u.Unmarshal(e.Message)
	if err != nil {
		logs.Error("usersMany()-> Unmarshal()", zap.Error(err))
		return
	}
	logs.Info("SyncController::usersMany",
		zap.Int("Users", len(u.Users)),
	)
	for _, v := range u.Users {
		repo.Users.Save(v)
	}
}

// messagesMany
func (ctrl *Controller) messagesMany(e *msg.MessageEnvelope) {
	u := new(msg.MessagesMany)
	err := u.Unmarshal(e.Message)
	if err != nil {
		logs.Error("messagesMany()-> Unmarshal()", zap.Error(err))
		return
	}

	// save Groups & Users & Messages
	repo.Users.SaveMany(u.Users)
	repo.Groups.SaveMany(u.Groups)
	repo.Messages.SaveMany(u.Messages)

	minID := int64(0)
	maxID := int64(0)
	for _, v := range u.Messages {
		if v.ID < minID || minID == 0 {
			minID = v.ID
		}
		if v.ID > maxID {
			maxID = v.ID
		}
	}

	// handle Media message
	ctrl.extractMessagesMedia(u.Messages...)

	logs.Info("SyncController::messagesMany",
		zap.Int("Messages", len(u.Messages)),
		zap.Bool("Continues", u.Continuous),
		zap.Int64("MinID", minID),
		zap.Int64("MaxID", maxID),
	)

	if u.Continuous && minID != 0 && minID != maxID {
		peerID := u.Messages[0].PeerID
		peerType := u.Messages[0].PeerType
		messageHole.InsertFill(peerID, peerType, minID, maxID)
	}
}

// groupFull
func (ctrl *Controller) groupFull(e *msg.MessageEnvelope) {
	u := new(msg.GroupFull)
	err := u.Unmarshal(e.Message)
	if err != nil {
		logs.Error("groupFull()-> Unmarshal()", zap.Error(err))
		return
	}
	logs.Info("SyncController::groupFull",
		zap.Int64("GroupID", u.Group.ID),
	)

	// save GroupSearch
	repo.Groups.Save(u.Group)

	// save GroupSearch Members
	for _, v := range u.Participants {
		repo.Groups.SaveParticipant(u.Group.ID, v)
	}

	// save Users
	for _, v := range u.Users {
		repo.Users.Save(v)
	}

	// Update NotifySettings
	repo.Dialogs.UpdateNotifySetting(u.Group.ID, int32(msg.PeerGroup), u.NotifySettings)
}
