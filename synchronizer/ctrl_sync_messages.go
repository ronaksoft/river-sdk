package synchronizer

import (
	"time"

	"git.ronaksoftware.com/ronak/riversdk/filemanager"

	"git.ronaksoftware.com/ronak/riversdk/cmd"
	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/log"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/repo"
	"go.uber.org/zap"
)

// authAuthorization
func (ctrl *SyncController) authAuthorization(e *msg.MessageEnvelope) {
	log.LOG_Debug("SyncController::authAuthorization() applier")
	x := new(msg.AuthAuthorization)
	if err := x.Unmarshal(e.Message); err != nil {
		log.LOG_Debug("SyncController::authAuthorization()-> Unmarshal()",
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

func (ctrl *SyncController) authSentCode(e *msg.MessageEnvelope) {
	log.LOG_Debug("SyncController::authSentCode() applier")
	x := new(msg.AuthSentCode)
	if err := x.Unmarshal(e.Message); err != nil {
		log.LOG_Debug("SyncController::authSentCode()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
		return
	}
	ctrl.connInfo.ChangePhone(x.Phone)
	// No need to save it here its gonna be saved on authAuthorization
	// conf.Save()
}

// contactsImported
func (ctrl *SyncController) contactsImported(e *msg.MessageEnvelope) {
	log.LOG_Debug("SyncController::contactsImported() applier")
	x := new(msg.ContactsImported)
	if err := x.Unmarshal(e.Message); err != nil {
		log.LOG_Debug("SyncController::contactsImported()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
		return
	}
	for _, u := range x.Users {
		repo.Ctx().Users.SaveContactUser(u)
	}
}

// contactsMany
func (ctrl *SyncController) contactsMany(e *msg.MessageEnvelope) {
	log.LOG_Debug("SyncController::contactsMany() applier")
	x := new(msg.ContactsMany)
	if err := x.Unmarshal(e.Message); err != nil {
		log.LOG_Debug("SyncController::contactsMany()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
		return
	}
	for _, u := range x.Users {
		repo.Ctx().Users.SaveContactUser(u)
	}
}

// messageDialogs
func (ctrl *SyncController) messagesDialogs(e *msg.MessageEnvelope) {
	log.LOG_Debug("SyncController::messagesDialogs() applier")
	x := new(msg.MessagesDialogs)
	if err := x.Unmarshal(e.Message); err != nil {
		log.LOG_Debug("SyncController::messagesDialogs()-> Unmarshal()",
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
			log.LOG_Debug(domain.ErrNotFound.Error(),
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

	log.LOG_Debug("SyncController::messagesDialogs()",
		zap.Int("DialogsCount", len(x.Dialogs)),
		zap.Int("GroupsCount", len(x.Groups)),
		zap.Int("UsersCount", len(x.Users)),
		zap.Int("MessagesCount", len(x.Messages)),
	)
}

// Check pending messages and notify UI
func (ctrl *SyncController) messageSent(e *msg.MessageEnvelope) {
	log.LOG_Debug("SyncController::messageSent() applier")
	pmsg, err := repo.Ctx().PendingMessages.GetPendingMessageByRequestID(int64(e.RequestID))
	if err != nil {
		log.LOG_Debug("SyncController::messageSent()-> GetPendingMessageByRequestID()",
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
		log.LOG_Debug("SyncController::messageSent()-> SaveMessage() failed to move pendingMessage to message table",
			zap.String("Error", err.Error()),
		)
		return
	}

	// delete pending mesage
	err = repo.Ctx().PendingMessages.DeletePendingMessage(pmsg.ID)
	if err != nil {
		log.LOG_Debug("SyncController::messageSent()-> DeletePendingMessage() failed to delete pendingMessage",
			zap.String("Error", err.Error()),
		)
	}
	// if it was file upload request
	if pmsg.MediaType > 0 {
		// save to local files and delete file status
		clientMendMedia := new(msg.ClientSendMessageMedia)
		clientMendMedia.Unmarshal(pmsg.Media)
		// save to local files
		err := repo.Ctx().Files.MoveUploadedFileToLocalFile(clientMendMedia, sent)
		filemanager.Ctx().DeleteFromQueue(pmsg.ID)
		if err != nil {
			log.LOG_Debug("SyncController::messageSent()-> MoveUploadedFileToLocalFile() failed ",
				zap.String("Error", err.Error()),
			)
		}
		// delete file status
		err = repo.Ctx().Files.DeleteFileStatus(pmsg.ID)
		if err != nil {
			log.LOG_Debug("SyncController::messageSent()-> DeleteFileStatus() failed to delete FileStatus",
				zap.String("Error", err.Error()),
			)
		}
	}

	//Update doaligs
	err = repo.Ctx().Dialogs.UpdateTopMesssageID(message.CreatedOn, message.PeerID, message.PeerType)
	if err != nil {
		log.LOG_Debug("SyncController::messageSent()-> UpdateTopMesssageID() failed to update doalogs",
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
	cmd.GetUIExecuter().Exec(func() {
		if ctrl.onUpdateMainDelegate != nil {
			ctrl.onUpdateMainDelegate(msg.C_UpdateEnvelope, buff)
		}
	})
}

func (ctrl *SyncController) usersMany(e *msg.MessageEnvelope) {
	log.LOG_Debug("SyncController::usersMany() applier")
	u := new(msg.UsersMany)
	err := u.Unmarshal(e.Message)
	if err != nil {
		log.LOG_Debug("SyncController::usersMany()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
		return
	}
	for _, v := range u.Users {
		repo.Ctx().Users.SaveUser(v)
	}
}

func (ctrl *SyncController) messagesMany(e *msg.MessageEnvelope) {
	log.LOG_Debug("SyncController::messagesMany() applier")
	u := new(msg.MessagesMany)
	err := u.Unmarshal(e.Message)
	if err != nil {
		log.LOG_Debug("SyncController::messagesMany()-> Unmarshal()",
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

	// fill MessageHoles
	for peerID := range peerMessages {
		err := fillMessageHoles(peerID, peerMessageMinID[peerID], peerMessageMaxID[peerID])
		if err != nil {
			log.LOG_Debug("SyncController::updateMessageHole()",
				zap.String("Error", err.Error()),
			)
		}
	}

}

func (ctrl *SyncController) groupFull(e *msg.MessageEnvelope) {
	log.LOG_Debug("SyncController::groupFull() applier")
	u := new(msg.GroupFull)
	err := u.Unmarshal(e.Message)
	if err != nil {
		log.LOG_Debug("SyncController::groupFull()-> Unmarshal()",
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
