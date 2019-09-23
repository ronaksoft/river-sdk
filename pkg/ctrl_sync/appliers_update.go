package syncCtrl

import (
	"fmt"
	fileCtrl "git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_file"
	messageHole "git.ronaksoftware.com/ronak/riversdk/pkg/message_hole"
	"git.ronaksoftware.com/ronak/riversdk/pkg/uiexec"
	"io"
	"os"
	"time"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"go.uber.org/zap"
)

// updateNewMessage
func (ctrl *Controller) updateNewMessage(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateNewMessage)
	err := x.Unmarshal(u.Update)
	if err != nil {
		logs.Error("updateNewMessage()-> Unmarshal()", zap.Error(err))
		return nil, err
	}

	logs.Info("SyncController::updateNewMessage",
		zap.Int64("MessageID", x.Message.ID),
		zap.Int64("UpdateID", x.UpdateID),
	)

	// used messageType to identify client & server messages on Media thingy
	x.Message.MessageType = 1

	dialog := repo.Dialogs.Get(x.Message.PeerID, x.Message.PeerType)
	if dialog == nil {
		// make sure to created the message hole b4 creating dialog
		dialog = &msg.Dialog{
			PeerID:         x.Message.PeerID,
			PeerType:       x.Message.PeerType,
			TopMessageID:   x.Message.ID,
			UnreadCount:    0,
			MentionedCount: 0,
			AccessHash:     x.AccessHash,
		}
		repo.Dialogs.SaveNew(dialog, x.Message.CreatedOn)
	}
	// save user if does not exist
	repo.Users.Save(x.Sender)
	repo.Messages.SaveNew(x.Message, dialog, ctrl.userID)
	messageHole.InsertFill(dialog.PeerID, dialog.PeerType, dialog.TopMessageID, x.Message.ID)

	// bug : sometime server do not sends access hash
	if x.AccessHash > 0 {
		// update users access hash
		repo.Users.UpdateAccessHash(x.AccessHash, x.Message.PeerID, x.Message.PeerType)
	}

	// If sender is me, check for pending
	if x.Message.SenderID == ctrl.userID {
		pm := repo.PendingMessages.GetByRealID(x.Message.ID)
		if pm != nil {
			ctrl.handlePendingMessage(x)
			repo.PendingMessages.Delete(pm.ID)
			repo.PendingMessages.DeleteByRealID(x.Message.ID)
		}
	}

	// handle Message's Action
	res := []*msg.UpdateEnvelope{u}
	ctrl.handleMessageAction(x, u, res)

	return res, nil
}
func (ctrl *Controller) handleMessageAction(x *msg.UpdateNewMessage, u *msg.UpdateEnvelope, res []*msg.UpdateEnvelope) {
	switch x.Message.MessageAction {
	case domain.MessageActionContactRegistered:
		go ctrl.ContactImportFromServer()
	case domain.MessageActionGroupDeleteUser:
		act := new(msg.MessageActionGroupAddUser)
		err := act.Unmarshal(x.Message.MessageActionData)
		if err != nil {
			logs.Error("updateNewMessage() -> MessageActionGroupDeleteUser Failed to Parse", zap.Error(err))
		}

		repo.Groups.DeleteMemberMany(x.Message.PeerID, act.UserIDs)

		// Check if user left (deleted him/her self from group) remove its GroupSearch, Dialog and its MessagesPending
		selfUserID := ctrl.connInfo.PickupUserID()
		userLeft := false
		for _, v := range act.UserIDs {
			if v == selfUserID {
				userLeft = true
				break
			}
		}

		// Don't go further if the user itself has not been left or removed
		if !userLeft {
			break
		}

		// Delete GroupSearch		NOT REQUIRED
		// Delete Dialog			NOT REQUIRED
		// Delete PendingMessage
		deletedMsgs := repo.PendingMessages.DeletePeerAllMessages(x.Message.PeerID, x.Message.PeerType)
		if deletedMsgs != nil {
			buff, err := deletedMsgs.Marshal()
			if err != nil {
				logs.Error("River::groupDeleteUser()-> Unmarshal ClientUpdateMessagesDeleted", zap.Error(err))
				break
			}
			udp := new(msg.UpdateEnvelope)
			udp.Constructor = msg.C_ClientUpdateMessagesDeleted
			udp.Update = buff
			udp.UCount = 1
			udp.Timestamp = u.Timestamp
			udp.UpdateID = u.UpdateID
			res = append(res, udp)

		}
	case domain.MessageActionClearHistory:
		act := new(msg.MessageActionClearHistory)
		_ = act.Unmarshal(x.Message.MessageActionData)

		// 1. Delete All Messages < x.MessageID
		repo.Messages.DeleteAll(ctrl.userID, x.Message.PeerID, x.Message.PeerType, act.MaxID-1)

		// Delete Scroll Position
		repo.MessagesExtra.SaveScrollID(x.Message.PeerID, x.Message.PeerType, 0)

		if act.Delete {
			// Delete Dialog
			repo.Dialogs.Delete(x.Message.PeerID, x.Message.PeerType)

			// Delete GroupSearch
			repo.Groups.Delete(x.Message.PeerID)
		} else {
			// get dialog and create first hole
			dtoDlg := repo.Dialogs.Get(x.Message.PeerID, x.Message.PeerType)
			if dtoDlg != nil {
				messageHole.InsertFill(dtoDlg.PeerID, dtoDlg.PeerType, dtoDlg.TopMessageID, dtoDlg.TopMessageID)
			}
		}
	}
}
func (ctrl *Controller) handlePendingMessage(x *msg.UpdateNewMessage) {
	pmsg := repo.PendingMessages.GetByRealID(x.Message.ID)
	if pmsg == nil {
		return
	}

	// if it was file upload request
	switch x.Message.MediaType {
	case msg.MediaTypeDocument:
		// save to local files and delete file status
		clientSendMedia := new(msg.ClientSendMessageMedia)
		err := clientSendMedia.Unmarshal(pmsg.Media)
		if err != nil {
			return
		}

		clientFile, err := repo.Files.GetMediaDocument(x.Message)
		logs.WarnOnErr("Error On GetMediaDocument", err)
		_, err = copyUploadedFile(clientSendMedia.FilePath, fileCtrl.GetFilePath(clientFile))
		if err != nil {
			return
		}
	}

	clientUpdate := new(msg.ClientUpdatePendingMessageDelivery)
	clientUpdate.Messages = x.Message
	clientUpdate.PendingMessage = pmsg
	clientUpdate.Success = true

	bytes, _ := clientUpdate.Marshal()

	udpMsg := new(msg.UpdateEnvelope)
	udpMsg.Constructor = msg.C_ClientUpdatePendingMessageDelivery
	udpMsg.Update = bytes
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
func copyUploadedFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

// updateReadHistoryInbox
func (ctrl *Controller) updateReadHistoryInbox(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateReadHistoryInbox)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	dialog := repo.Dialogs.Get(x.Peer.ID, x.Peer.Type)
	if dialog == nil {
		return nil, err
	}

	logs.Info("SyncController::updateReadHistoryInbox",
		zap.Int64("MaxID", x.MaxID),
		zap.Int64("UpdateID", x.UpdateID),
		zap.Int64("PeerID", x.Peer.ID),
	)

	repo.Dialogs.UpdateReadInboxMaxID(ctrl.userID, x.Peer.ID, x.Peer.Type, x.MaxID)
	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

// updateReadHistoryOutbox
func (ctrl *Controller) updateReadHistoryOutbox(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateReadHistoryOutbox)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Info("SyncController::updateReadHistoryOutbox",
		zap.Int64("MaxID", x.MaxID),
		zap.Int64("UpdateID", x.UpdateID),
		zap.Int64("PeerID", x.Peer.ID),
	)

	repo.Dialogs.UpdateReadOutboxMaxID(x.Peer.ID, x.Peer.Type, x.MaxID)
	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

// updateMessageEdited
func (ctrl *Controller) updateMessageEdited(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateMessageEdited)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Info("SyncController::updateMessageEdited",
		zap.Int64("MessageID", x.Message.ID),
	)

	repo.Messages.Save(x.Message)

	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

// updateMessageID
func (ctrl *Controller) updateMessageID(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	res := make([]*msg.UpdateEnvelope, 0)
	x := new(msg.UpdateMessageID)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Info("SyncController::updateMessageID",
		zap.Int64("RandomID", x.RandomID),
		zap.Int64("MessageID", x.MessageID),
	)

	msgEnvelop := new(msg.MessageEnvelope)
	msgEnvelop.Constructor = msg.C_MessageEnvelope

	sent := new(msg.MessagesSent)
	sent.MessageID = x.MessageID
	sent.RandomID = x.RandomID
	sent.CreatedOn = time.Now().Unix()

	// If we are here, it means we receive UpdateNewMessage before UpdateMessageID / MessagesSent
	// so we create a fake UpdateMessageDelete to remove the pending from the view
	pm, _ := repo.PendingMessages.GetByRandomID(sent.RandomID)

	userMessage := repo.Messages.Get(sent.MessageID)
	if userMessage != nil {
		logs.Info("Pending Message:: UpdateMessageID after UpdateNewMessage",
			zap.Int64("MID", x.MessageID),
			zap.Int64("RandomID", x.RandomID),
			zap.String("Body", userMessage.Body),
		)

		if pm == nil {
			logs.Warn("Pending Message NOT FOUND",
				zap.Int64("RandomID", sent.RandomID),
				zap.Int64("RealID", sent.MessageID),
			)
			return res, nil
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

		_ = repo.PendingMessages.Delete(pm.ID)
		return res, nil
	} else {
		logs.Info("Pending Message:: UpdateMessageID before UpdateNewMessage",
			zap.Int64("MID", x.MessageID),
			zap.Int64("PendingID", pm.ID),
		)
	}

	if pm == nil {
		return nil, nil
	}
	repo.PendingMessages.SaveByRealID(sent.RandomID, sent.MessageID)

	return res, nil
}

// updateNotifySettings
func (ctrl *Controller) updateNotifySettings(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateNotifySettings)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Info("SyncController::updateNotifySettings()")

	repo.Dialogs.UpdateNotifySetting(x.NotifyPeer.ID, x.NotifyPeer.Type, x.Settings)

	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

// updateDialogPinned
func (ctrl *Controller) updateDialogPinned(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateDialogPinned)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Info("SyncController::updateDialogPinned()")

	repo.Dialogs.UpdatePinned(x)
	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

// updateUsername
func (ctrl *Controller) updateUsername(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateUsername)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Info("SyncController::updateUsername()")

	if x.UserID == ctrl.userID {
		ctrl.connInfo.ChangeUserID(x.UserID)
		ctrl.connInfo.ChangeUsername(x.Username)
		ctrl.connInfo.ChangeFirstName(x.FirstName)
		ctrl.connInfo.ChangeLastName(x.LastName)
		ctrl.connInfo.ChangeBio(x.Bio)
		ctrl.connInfo.Save()
	}

	repo.Users.UpdateProfile(x.UserID, x.FirstName, x.LastName, x.Username, x.Bio)

	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

// updateMessagesDeleted
func (ctrl *Controller) updateMessagesDeleted(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateMessagesDeleted)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Info("SyncController::updateMessagesDeleted()")

	for _, msgID := range x.MessageIDs {
		repo.Messages.Delete(ctrl.userID, x.Peer.ID, x.Peer.Type, msgID)
	}

	update := new(msg.ClientUpdateMessagesDeleted)
	update.PeerID = x.Peer.ID
	update.PeerType = x.Peer.Type
	update.MessageIDs = x.MessageIDs

	res := []*msg.UpdateEnvelope{u}

	updateEnvelope := new(msg.UpdateEnvelope)
	updateEnvelope.Constructor = msg.C_ClientUpdateMessagesDeleted
	updateEnvelope.Timestamp = u.Timestamp
	updateEnvelope.UCount = 1
	updateEnvelope.UpdateID = u.UpdateID
	updateEnvelope.Update, _ = update.Marshal()
	res = append(res, updateEnvelope)
	return res, nil
}

// updateGroupParticipantAdmin
func (ctrl *Controller) updateGroupParticipantAdmin(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateGroupParticipantAdmin)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Info("SyncController::updateGroupParticipantAdmin()")

	res := []*msg.UpdateEnvelope{u}
	repo.Groups.UpdateMemberType(x.GroupID, x.UserID, x.IsAdmin)
	return res, nil
}

// updateReadMessagesContents
func (ctrl *Controller) updateReadMessagesContents(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateReadMessagesContents)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Info("SyncController::updateReadMessagesContents",
		zap.Int64s("MessageIDs", x.MessageIDs),
	)

	repo.Messages.SetContentRead(x.Peer.ID, int32(x.Peer.Type), x.MessageIDs)

	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

// updateUserPhoto
func (ctrl *Controller) updateUserPhoto(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateUserPhoto)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Info("SyncController::updateUserPhoto()")

	repo.Users.UpdatePhoto(x)

	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

// updateGroupPhoto
func (ctrl *Controller) updateGroupPhoto(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateGroupPhoto)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Info("SyncController::updateGroupPhoto",
		zap.Int64("GroupID", x.GroupID),
	)

	repo.Groups.UpdatePhoto(x)

	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

// updateTooLong indicate that client updates exceed from server cache size so we should re-sync client with server
func (ctrl *Controller) updateTooLong(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	logs.Info("SyncController::updateTooLong()")

	go ctrl.sync()

	res := make([]*msg.UpdateEnvelope, 0)
	return res, nil
}

func (ctrl *Controller) updateAccountPrivacy(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateAccountPrivacy)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Info("SyncController::updateAccountPrivacy")

	_ = repo.Account.SetPrivacy(msg.PrivacyKeyChatInvite, x.ChatInvite)
	_ = repo.Account.SetPrivacy(msg.PrivacyKeyLastSeen, x.LastSeen)
	_ = repo.Account.SetPrivacy(msg.PrivacyKeyPhoneNumber, x.PhoneNumber)
	_ = repo.Account.SetPrivacy(msg.PrivacyKeyProfilePhoto, x.ProfilePhoto)
	_ = repo.Account.SetPrivacy(msg.PrivacyKeyForwardedMessage, x.ForwardedMessage)
	_ = repo.Account.SetPrivacy(msg.PrivacyKeyCall, x.Call)

	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

func (ctrl *Controller) updateDraftMessage(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateDraftMessage)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Info("SyncController::updateDraftMessage")

	dialog := repo.Dialogs.Get(x.Message.PeerID, int32(x.Message.PeerType))

	if dialog != nil {
		dialog.Draft = x.Message
		repo.Dialogs.Save(dialog)
	}

	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

func (ctrl *Controller) updateDraftMessageCleared(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateDraftMessageCleared)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Info("SyncController::updateDraftMessageCleared")

	dialog := repo.Dialogs.Get(x.Peer.ID, int32(x.Peer.Type))

	if dialog != nil {

		dialog.Draft = nil

		repo.Dialogs.Save(dialog)
	}

	res := []*msg.UpdateEnvelope{u}
	return res, nil
}
