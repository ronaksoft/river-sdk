package syncCtrl

import (
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	messageHole "git.ronaksoftware.com/ronak/riversdk/pkg/message_hole"
	"os"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"git.ronaksoftware.com/ronak/riversdk/pkg/uiexec"
	"go.uber.org/zap"
)

// authAuthorization
func (ctrl *Controller) authAuthorization(e *msg.MessageEnvelope) {
	x := new(msg.AuthAuthorization)
	if err := x.Unmarshal(e.Message); err != nil {
		logs.Warn("authAuthorization()-> Unmarshal()", zap.Error(err))
		return
	}
	logs.Info("SyncController::authAuthorization",
		zap.String("FirstName",x.User.FirstName),
		zap.String("LastName",x.User.LastName),
		zap.Int64("UserID", x.User.ID),
		zap.String("Bio",x.User.Bio),
		zap.String("Username",x.User.Username),
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
	// conf.Save()
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
		_ = repo.Messages.SaveMessage(message)
		mMessages[message.ID] = message
	}
	for _, dialog := range x.Dialogs {
		topMessage, _ := mMessages[dialog.TopMessageID]
		if topMessage == nil {
			logs.Debug(domain.ErrNotFound.Error(),
				zap.Int64("MessageID", dialog.TopMessageID),
			)
			continue
		}
		_ = repo.Dialogs.Save(dialog, topMessage.CreatedOn)
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

	// Add delivered message to prevent invoking newMessage event later
	ctrl.addToDeliveredMessageList(sent.MessageID)

	pmsg, err := repo.PendingMessages.GetPendingMessageByRequestID(int64(e.RequestID))
	if err != nil {
		return
	}

	// if it was file upload request
	if pmsg.MediaType > 0 {
		// save to local files and delete file status
		clientSendMedia := new(msg.ClientSendMessageMedia)
		err := clientSendMedia.Unmarshal(pmsg.Media)
		// get file size
		fileSize := int64(0)
		if err == nil {
			f, err := os.Open(clientSendMedia.FilePath)
			if err == nil {
				fstate, err := f.Stat()
				if err == nil {
					fileSize = fstate.Size()
				}
			}
			// save to local files
			err = repo.Files.MoveUploadedFileToFiles(clientSendMedia, int32(fileSize), sent)
			if err != nil {
				logs.Warn("messageSent()-> MoveUploadedFileToLocalFile() failed ", zap.Error(err))
			}
			// delete file status
			err = repo.Files.DeleteFileStatus(pmsg.ID)
			if err != nil {
				logs.Warn("messageSent()-> DeleteFileStatus() failed to delete FileStatus", zap.Error(err))
			}
		}
	}

	message := new(msg.UserMessage)
	pmsg.MapToUserMessage(message)
	message.ID = sent.MessageID
	message.CreatedOn = sent.CreatedOn

	// save message
	err = repo.Messages.SaveMessage(message)
	if err != nil {
		logs.Warn("messageSent()-> SaveMessage() failed to move pendingMessage to message table", zap.Error(err))
		return
	}

	// delete pending mesage
	err = repo.PendingMessages.DeletePendingMessage(pmsg.ID)
	if err != nil {
		logs.Warn("messageSent()-> DeletePendingMessage() failed to delete pendingMessage", zap.Error(err))
	}

	// Update dialogs
	err = repo.Dialogs.UpdateTopMessageID(message.CreatedOn, message.PeerID, message.PeerType)
	if err != nil {
		logs.Warn("messageSent()-> UpdateTopMessageID() failed to update dialogs", zap.Error(err))
	}

	// TODO : Notify UI that the pending message delivered to server
	e.Constructor = msg.C_ClientUpdatePendingMessageDelivery
	pbcpm := new(msg.ClientPendingMessage)
	pmsg.MapTo(pbcpm)

	out := msg.ClientUpdatePendingMessageDelivery{
		Messages:       message,
		PendingMessage: pbcpm,
		Success:        true,
	}

	e.Message, _ = out.Marshal()

	udpMsg := new(msg.UpdateEnvelope)
	udpMsg.Constructor = e.Constructor
	udpMsg.Update = e.Message
	udpMsg.UpdateID = 0
	udpMsg.Timestamp = time.Now().Unix()

	buff, _ := udpMsg.Marshal()

	// call external handler
	uiexec.Ctx().Exec(func() {
		if ctrl.onUpdateMainDelegate != nil {
			ctrl.onUpdateMainDelegate(msg.C_UpdateEnvelope, buff)
		}
	})
}
func (ctrl *Controller) addToDeliveredMessageList(id int64) {
	ctrl.deliveredMessagesMutex.Lock()
	ctrl.deliveredMessages[id] = true
	ctrl.deliveredMessagesMutex.Unlock()
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
	logs.Info("SyncController::messagesMany",
		zap.Int("Messages", len(u.Messages)),
		zap.Bool("Continues", u.Continuous),
	)

	// Save Groups & Users
	_ = repo.Users.SaveMany(u.Users)
	_ = repo.Groups.SaveMany(u.Groups)

	// handle Media message
	go extractMessagesMedia(u.Messages...)

	minID := int64(0)
	maxID := int64(0)
	for _, v := range u.Messages {
		_ = repo.Messages.SaveMessage(v)
		if minID == 0 || v.ID < minID {
			minID = v.ID
		}
		if maxID == 0 || v.ID > maxID {
			maxID = v.ID
		}
	}

	if u.Continuous && minID != 0 && minID != maxID {
		peerID := u.Messages[0].PeerID
		peerType := u.Messages[0].PeerType
		_ = messageHole.InsertFill(peerID, peerType, minID, maxID)
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

	// Save Group
	_ = repo.Groups.Save(u.Group)

	// Save Group Members
	for _, v := range u.Participants {
		_ = repo.Groups.SaveParticipants(u.Group.ID, v)
	}

	// Save Users
	for _, v := range u.Users {
		_ = repo.Users.Save(v)
	}

	// Update NotifySettings
	repo.Dialogs.UpdatePeerNotifySettings(u.Group.ID, int32(msg.PeerGroup), u.NotifySettings)
}
