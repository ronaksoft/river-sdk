package synchronizer

import (
	"os"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"git.ronaksoftware.com/ronak/riversdk/pkg/uiexec"
	"go.uber.org/zap"
)

// authAuthorization
func (ctrl *Controller) authAuthorization(e *msg.MessageEnvelope) {
	logs.Info("authAuthorization() applier")
	x := new(msg.AuthAuthorization)
	if err := x.Unmarshal(e.Message); err != nil {
		logs.Error("authAuthorization()-> Unmarshal()", zap.Error(err))
		return
	}

	ctrl.connInfo.ChangeFirstName(x.User.FirstName)
	ctrl.connInfo.ChangeLastName(x.User.LastName)
	ctrl.connInfo.ChangeUsername(x.User.Username)
	ctrl.connInfo.ChangeUserID(x.User.ID)
	ctrl.connInfo.Save()

	ctrl.SetUserID(x.User.ID)

	go ctrl.sync()
}

// authSentCode
func (ctrl *Controller) authSentCode(e *msg.MessageEnvelope) {
	logs.Info("authSentCode() applier")
	x := new(msg.AuthSentCode)
	if err := x.Unmarshal(e.Message); err != nil {
		logs.Error("authSentCode()-> Unmarshal()", zap.Error(err))
		return
	}
	ctrl.connInfo.ChangePhone(x.Phone)
	// No need to save it here its gonna be saved on authAuthorization
	// conf.Save()
}

// contactsImported
func (ctrl *Controller) contactsImported(e *msg.MessageEnvelope) {
	logs.Info("contactsImported() applier")
	x := new(msg.ContactsImported)
	if err := x.Unmarshal(e.Message); err != nil {
		logs.Error("contactsImported()-> Unmarshal()", zap.Error(err))
		return
	}
	for _, u := range x.Users {
		repo.Ctx().Users.SaveContactUser(u)
	}
}

// contactsMany
func (ctrl *Controller) contactsMany(e *msg.MessageEnvelope) {
	logs.Info("contactsMany() applier")
	x := new(msg.ContactsMany)
	if err := x.Unmarshal(e.Message); err != nil {
		logs.Error("contactsMany()-> Unmarshal()", zap.Error(err))
		return
	}

	userIDs := domain.MInt64B{}
	for _, u := range x.Users {
		userIDs[u.ID] = true
		repo.Ctx().Users.SaveContactUser(u)
	}
	// server
	if len(userIDs) > 0 {
		// calculate contactsGethash and save
		crc32Hash := domain.CalculateContactsGetHash(userIDs.ToArray())
		err := repo.Ctx().System.SaveInt(domain.ColumnContactsGetHash, int32(crc32Hash))
		if err != nil {
			logs.Error("contactsMany() failed to save ContactsGetHash to DB", zap.Error(err))
		}
	}
}

// messageDialogs
func (ctrl *Controller) messagesDialogs(e *msg.MessageEnvelope) {
	logs.Info("messagesDialogs() applier")
	x := new(msg.MessagesDialogs)
	if err := x.Unmarshal(e.Message); err != nil {
		logs.Error("messagesDialogs()-> Unmarshal()", zap.Error(err))
		return
	}

	mMessages := make(map[int64]*msg.UserMessage)
	for _, message := range x.Messages {
		repo.Ctx().Messages.SaveMessage(message)
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
		repo.Ctx().Dialogs.SaveDialog(dialog, topMessage.CreatedOn)
	}
	for _, user := range x.Users {
		repo.Ctx().Users.SaveUser(user)
	}
	for _, group := range x.Groups {
		repo.Ctx().Groups.Save(group)
	}

	logs.Debug("messagesDialogs()",
		zap.Int("DialogsCount", len(x.Dialogs)),
		zap.Int("GroupsCount", len(x.Groups)),
		zap.Int("UsersCount", len(x.Users)),
		zap.Int("MessagesCount", len(x.Messages)),
	)
}

// Check pending messages and notify UI
func (ctrl *Controller) messageSent(e *msg.MessageEnvelope) {

	logs.Info("messageSent() applier")

	sent := new(msg.MessagesSent)
	err := sent.Unmarshal(e.Message)
	if err != nil {
		logs.Error("MessageSent() failed to unamarshal", zap.Error(err))
		return
	}

	// Add delivered message to prevent invoking newMessage event later
	ctrl.addToDeliveredMessageList(sent.MessageID)

	pmsg, err := repo.Ctx().PendingMessages.GetPendingMessageByRequestID(int64(e.RequestID))
	if err != nil {
		logs.Error("messageSent()-> GetPendingMessageByRequestID()", zap.Uint64("RequestID", e.RequestID), zap.Error(err))
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
			err = repo.Ctx().Files.MoveUploadedFileToFiles(clientSendMedia, int32(fileSize), sent)
			if err != nil {
				logs.Error("messageSent()-> MoveUploadedFileToLocalFile() failed ", zap.Error(err))
			}
			// delete file status
			err = repo.Ctx().Files.DeleteFileStatus(pmsg.ID)
			if err != nil {
				logs.Error("messageSent()-> DeleteFileStatus() failed to delete FileStatus", zap.Error(err))
			}
		}
	}

	message := new(msg.UserMessage)
	pmsg.MapToUserMessage(message)
	message.ID = sent.MessageID
	message.CreatedOn = sent.CreatedOn

	// save message
	err = repo.Ctx().Messages.SaveMessage(message)
	if err != nil {
		logs.Error("messageSent()-> SaveMessage() failed to move pendingMessage to message table", zap.Error(err))
		return
	}

	// delete pending mesage
	err = repo.Ctx().PendingMessages.DeletePendingMessage(pmsg.ID)
	if err != nil {
		logs.Error("messageSent()-> DeletePendingMessage() failed to delete pendingMessage", zap.Error(err))
	}

	//Update doaligs
	err = repo.Ctx().Dialogs.UpdateTopMessageID(message.CreatedOn, message.PeerID, message.PeerType)
	if err != nil {
		logs.Error("messageSent()-> UpdateTopMessageID() failed to update doalogs", zap.Error(err))
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

// usersMany
func (ctrl *Controller) usersMany(e *msg.MessageEnvelope) {
	logs.Info("usersMany() applier")
	u := new(msg.UsersMany)
	err := u.Unmarshal(e.Message)
	if err != nil {
		logs.Error("usersMany()-> Unmarshal()", zap.Error(err))
		return
	}
	for _, v := range u.Users {
		repo.Ctx().Users.SaveUser(v)
	}
}

// messagesMany
func (ctrl *Controller) messagesMany(e *msg.MessageEnvelope) {
	logs.Info("messagesMany() applier")
	u := new(msg.MessagesMany)
	err := u.Unmarshal(e.Message)
	if err != nil {
		logs.Error("messagesMany()-> Unmarshal()", zap.Error(err))
		return
	}
	for _, v := range u.Users {
		repo.Ctx().Users.SaveUser(v)
	}

	// extract MessageHole info
	peerMessageMinID := make(map[int64]int64)
	peerMessageMaxID := make(map[int64]int64)
	peerMessages := make(map[int64]bool)
	for _, v := range u.Messages {
		repo.Ctx().Messages.SaveMessage(v)
		if _, ok := peerMessages[v.PeerID]; ok {
			if v.ID < peerMessageMinID[v.PeerID] {
				peerMessageMinID[v.PeerID] = v.ID
			}
			if v.ID > peerMessageMaxID[v.PeerID] {
				peerMessageMaxID[v.PeerID] = v.ID
			}
		} else {
			peerMessages[v.PeerID] = true
			peerMessageMaxID[v.PeerID] = v.ID
			peerMessageMinID[v.PeerID] = v.ID
		}
	}

	// handle Media message
	go handleMediaMessage(u.Messages...)

	// fill MessageHoles
	for peerID := range peerMessages {
		err := fillMessageHoles(peerID, peerMessageMinID[peerID], peerMessageMaxID[peerID])
		if err != nil {
			logs.Error("updateMessageHole()", zap.Error(err))
		}
	}

}

// groupFull
func (ctrl *Controller) groupFull(e *msg.MessageEnvelope) {
	logs.Info("groupFull() applier")
	u := new(msg.GroupFull)
	err := u.Unmarshal(e.Message)
	if err != nil {
		logs.Error("groupFull()-> Unmarshal()", zap.Error(err))
		return
	}
	// Save Group
	repo.Ctx().Groups.Save(u.Group)

	// Save Group Members
	for _, v := range u.Participants {
		repo.Ctx().Groups.SaveParticipants(u.Group.ID, v)
	}

	// Save Users
	for _, v := range u.Users {
		repo.Ctx().Users.SaveUser(v)
	}

	// Update NotifySettings
	repo.Ctx().Dialogs.UpdatePeerNotifySettings(u.Group.ID, int32(msg.PeerGroup), u.NotifySettings)
}
