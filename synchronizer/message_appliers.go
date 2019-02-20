package synchronizer

import (
	"os"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/filemanager"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/logs"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/repo"
	"git.ronaksoftware.com/ronak/riversdk/uiexec"
	"go.uber.org/zap"
)

// authAuthorization
func (ctrl *Controller) authAuthorization(e *msg.MessageEnvelope) {
	logs.Debug("authAuthorization() applier")
	x := new(msg.AuthAuthorization)
	if err := x.Unmarshal(e.Message); err != nil {
		logs.Debug("authAuthorization()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
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
	logs.Debug("authSentCode() applier")
	x := new(msg.AuthSentCode)
	if err := x.Unmarshal(e.Message); err != nil {
		logs.Debug("authSentCode()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
		return
	}
	ctrl.connInfo.ChangePhone(x.Phone)
	// No need to save it here its gonna be saved on authAuthorization
	// conf.Save()
}

// contactsImported
func (ctrl *Controller) contactsImported(e *msg.MessageEnvelope) {
	logs.Debug("contactsImported() applier")
	x := new(msg.ContactsImported)
	if err := x.Unmarshal(e.Message); err != nil {
		logs.Debug("contactsImported()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
		return
	}
	for _, u := range x.Users {
		repo.Ctx().Users.SaveContactUser(u)
	}
}

// contactsMany
func (ctrl *Controller) contactsMany(e *msg.MessageEnvelope) {
	logs.Debug("contactsMany() applier")
	x := new(msg.ContactsMany)
	if err := x.Unmarshal(e.Message); err != nil {
		logs.Debug("contactsMany()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
		return
	}
	for _, u := range x.Users {
		repo.Ctx().Users.SaveContactUser(u)
	}
}

// messageDialogs
func (ctrl *Controller) messagesDialogs(e *msg.MessageEnvelope) {
	logs.Debug("messagesDialogs() applier")
	x := new(msg.MessagesDialogs)
	if err := x.Unmarshal(e.Message); err != nil {
		logs.Debug("messagesDialogs()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
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
	logs.Debug("messageSent() applier")
	pmsg, err := repo.Ctx().PendingMessages.GetPendingMessageByRequestID(int64(e.RequestID))
	if err != nil {
		logs.Debug("messageSent()-> GetPendingMessageByRequestID()",
			zap.String("Error", err.Error()),
		)
		return
	}

	sent := new(msg.MessagesSent)
	sent.Unmarshal(e.Message)

	message := new(msg.UserMessage)

	pmsg.MapToUserMessage(message)

	message.ID = sent.MessageID
	message.CreatedOn = sent.CreatedOn

	// save message
	err = repo.Ctx().Messages.SaveMessage(message)
	if err != nil {
		logs.Debug("messageSent()-> SaveMessage() failed to move pendingMessage to message table",
			zap.String("Error", err.Error()),
		)
		return
	}

	// delete pending mesage
	err = repo.Ctx().PendingMessages.DeletePendingMessage(pmsg.ID)
	if err != nil {
		logs.Debug("messageSent()-> DeletePendingMessage() failed to delete pendingMessage",
			zap.String("Error", err.Error()),
		)
	}
	// if it was file upload request
	if pmsg.MediaType == int32(msg.InputMediaTypeUploadedDocument) {
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
		}
		// save to local files
		err = repo.Ctx().Files.MoveUploadedFileToFiles(clientSendMedia, int32(fileSize), sent)
		filemanager.Ctx().DeleteFromQueue(pmsg.ID)
		if err != nil {
			logs.Debug("messageSent()-> MoveUploadedFileToLocalFile() failed ",
				zap.String("Error", err.Error()),
			)
		}
		// delete file status
		err = repo.Ctx().Files.DeleteFileStatus(pmsg.ID)
		if err != nil {
			logs.Debug("messageSent()-> DeleteFileStatus() failed to delete FileStatus",
				zap.String("Error", err.Error()),
			)
		}
	}

	//Update doaligs
	err = repo.Ctx().Dialogs.UpdateTopMesssageID(message.CreatedOn, message.PeerID, message.PeerType)
	if err != nil {
		logs.Debug("messageSent()-> UpdateTopMesssageID() failed to update doalogs",
			zap.String("Error", err.Error()),
		)
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
	// Add delivered message to prevent invoking newMessage event later
	ctrl.addToDeliveredMessageList(message.ID)

	// call external handler
	uiexec.Ctx().Exec(func() {
		if ctrl.onUpdateMainDelegate != nil {
			ctrl.onUpdateMainDelegate(msg.C_UpdateEnvelope, buff)
		}
	})
}

// usersMany
func (ctrl *Controller) usersMany(e *msg.MessageEnvelope) {
	logs.Debug("usersMany() applier")
	u := new(msg.UsersMany)
	err := u.Unmarshal(e.Message)
	if err != nil {
		logs.Debug("usersMany()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
		return
	}
	for _, v := range u.Users {
		repo.Ctx().Users.SaveUser(v)
	}
}

// messagesMany
func (ctrl *Controller) messagesMany(e *msg.MessageEnvelope) {
	logs.Debug("messagesMany() applier")
	u := new(msg.MessagesMany)
	err := u.Unmarshal(e.Message)
	if err != nil {
		logs.Debug("messagesMany()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
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
			logs.Debug("updateMessageHole()",
				zap.String("Error", err.Error()),
			)
		}
	}

}

// groupFull
func (ctrl *Controller) groupFull(e *msg.MessageEnvelope) {
	logs.Debug("groupFull() applier")
	u := new(msg.GroupFull)
	err := u.Unmarshal(e.Message)
	if err != nil {
		logs.Debug("groupFull()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
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
