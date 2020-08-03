package syncCtrl

import (
	mon "git.ronaksoftware.com/ronak/riversdk/internal/monitoring"
	messageHole "git.ronaksoftware.com/ronak/riversdk/pkg/message_hole"
	"git.ronaksoftware.com/ronak/riversdk/pkg/uiexec"
	"os"
	"sync"
	"time"

	"git.ronaksoftware.com/river/msg/msg"
	"git.ronaksoftware.com/ronak/riversdk/internal/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"go.uber.org/zap"
)

func (ctrl *Controller) updateNewMessage(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateNewMessage)
	err := x.Unmarshal(u.Update)
	if err != nil {
		logs.Error("UpdateApplier couldn't unmarshal UpdateNewMessage", zap.Error(err))
		return nil, err
	}

	logs.Info("SyncCtrl applies UpdateNewMessage",
		zap.Int64("MessageID", x.Message.ID),
		zap.Int64("UpdateID", x.UpdateID),
	)

	// used messageType to identify client & server messages on Media thingy
	x.Message.MessageType = 1
	repo.MessagesExtra.SaveScrollID(x.Message.TeamID, x.Message.PeerID, x.Message.PeerType, 0)

	dialog, _ := repo.Dialogs.Get(x.Message.TeamID, x.Message.PeerID, x.Message.PeerType)
	if dialog == nil {
		unreadCount := int32(0)
		if x.Sender.ID != ctrl.GetUserID() {
			unreadCount = 1
		}
		// make sure to created the message hole b4 creating dialog
		dialog = &msg.Dialog{
			PeerID:         x.Message.PeerID,
			PeerType:       x.Message.PeerType,
			TopMessageID:   x.Message.ID,
			UnreadCount:    unreadCount,
			MentionedCount: 0,
			AccessHash:     x.AccessHash,
		}
		err := repo.Dialogs.SaveNew(dialog, x.Message.CreatedOn)
		if err != nil {
			return nil, err
		}
	}
	// save user if does not exist
	repo.Users.Save(x.Sender)
	err = repo.Messages.SaveNew(x.Message, ctrl.GetUserID())
	if err != nil {
		return nil, err
	}
	messageHole.InsertFill(dialog.TeamID, dialog.PeerID, dialog.PeerType, dialog.TopMessageID, x.Message.ID)

	// If sender is me, check for pending
	if x.Message.SenderID == ctrl.GetUserID() {
		pm := repo.PendingMessages.GetByRealID(x.Message.ID)

		if pm != nil {
			ctrl.handlePendingMessage(x)
			_ = repo.PendingMessages.Delete(pm.ID)
			repo.PendingMessages.DeleteByRealID(x.Message.ID)
		}

	}

	// handle Message's Action
	res := []*msg.UpdateEnvelope{u}
	ctrl.handleMessageAction(x, u, res)

	// update monitoring && top peer && gif
	if x.Message.SenderID == ctrl.GetUserID() {
		if x.Message.PeerID != x.Message.SenderID {
			if x.Message.FwdSenderID != 0 {
				repo.TopPeers.Update(msg.TopPeerCategory_Forwards, x.Message.TeamID, x.Message.PeerID, x.Message.PeerType, ctrl.GetUserID())
			} else {
				switch msg.PeerType(x.Message.PeerType) {
				case msg.PeerUser:
					p, _ := repo.Users.Get(x.Message.PeerID)
					if p == nil || !p.IsBot {
						repo.TopPeers.Update(msg.TopPeerCategory_Users, x.Message.TeamID, x.Message.PeerID, x.Message.PeerType, ctrl.GetUserID())
					} else {
						repo.TopPeers.Update(msg.TopPeerCategory_BotsMessage, x.Message.TeamID, x.Message.PeerID, x.Message.PeerType, ctrl.GetUserID())
					}
				case msg.PeerGroup:
					repo.TopPeers.Update(msg.TopPeerCategory_Groups, x.Message.TeamID, x.Message.PeerID, x.Message.PeerType, ctrl.GetUserID())
				}
			}
		}

		// We check if the file is GIF then check if it is a saved gif, if true, then update last access
		switch x.Message.MediaType {
		case msg.MediaTypeDocument:
			xx := &msg.MediaDocument{}
			_ = xx.Unmarshal(x.Message.Media)
			for _, attr := range xx.Doc.Attributes {
				if attr.Type == msg.AttributeTypeAnimated {
					if repo.Gifs.IsSaved(xx.Doc.ClusterID, xx.Doc.ID) {
						_ = repo.Gifs.UpdateLastAccess(xx.Doc.ClusterID, xx.Doc.ID, x.Message.CreatedOn)
						forceUpdateUI(ctrl, false, false, true)
					}
				}
			}
		}

		switch x.Message.MediaType {
		case msg.MediaTypeEmpty:
			mon.IncMessageSent()
		default:
			mon.IncMediaSent()
		}
	} else {
		switch x.Message.MediaType {
		case msg.MediaTypeEmpty:
			mon.IncMessageReceived()
		default:
			mon.IncMediaReceived()
		}
	}

	return res, nil
}
func (ctrl *Controller) handleMessageAction(x *msg.UpdateNewMessage, u *msg.UpdateEnvelope, res []*msg.UpdateEnvelope) {
	switch x.Message.MessageAction {
	case domain.MessageActionContactRegistered:
		go func() {
			waitGroup := &sync.WaitGroup{}
			waitGroup.Add(1)
			ctrl.GetContacts(waitGroup)
			waitGroup.Wait()
		}()
	case domain.MessageActionGroupDeleteUser:
		act := new(msg.MessageActionGroupDeleteUser)
		err := act.Unmarshal(x.Message.MessageActionData)
		if err != nil {
			logs.Error("SyncCtrl couldn't unmarshal MessageActionGroupDeleteUser", zap.Error(err))
		}

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

		// remove from top peers
		err = repo.TopPeers.Delete(msg.TopPeerCategory_Groups, x.Message.TeamID, x.Message.PeerID, x.Message.PeerType)
		if err != nil {
			logs.Error("UpdateApplier couldn't delete group from top peers", zap.Error(err))
		}

		// Delete PendingMessage
		deletedMsgs := repo.PendingMessages.DeletePeerAllMessages(x.Message.PeerID, x.Message.PeerType)
		if deletedMsgs != nil {
			buff, err := deletedMsgs.Marshal()
			if err != nil {
				logs.Error("UpdateApplier couldn't marshal ClientUpdateMessagesDeleted", zap.Error(err))
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
		_ = repo.Messages.ClearHistory(ctrl.GetUserID(), x.Message.TeamID, x.Message.PeerID, x.Message.PeerType, act.MaxID)

		// 2. Delete Scroll Position
		repo.MessagesExtra.SaveScrollID(x.Message.TeamID, x.Message.PeerID, x.Message.PeerType, 0)

		if act.Delete {
			// 3. Delete Dialog
			repo.Dialogs.Delete(x.Message.TeamID, x.Message.PeerID, x.Message.PeerType)
		} else {
			// 3. Get dialog and create first hole
			dialog, _ := repo.Dialogs.Get(x.Message.TeamID, x.Message.PeerID, x.Message.PeerType)
			if dialog != nil {
				messageHole.InsertFill(dialog.TeamID, dialog.PeerID, dialog.PeerType, dialog.TopMessageID, dialog.TopMessageID)
			}
		}
	}
}
func (ctrl *Controller) handlePendingMessage(x *msg.UpdateNewMessage) {
	pmsg := repo.PendingMessages.GetByRealID(x.Message.ID)
	if pmsg == nil {
		return
	}

	if pmsg.MediaType != msg.InputMediaTypeMessageDocument {
		// if it was file upload request
		switch x.Message.MediaType {
		case msg.MediaTypeDocument:
			// save to local files and delete file status
			clientSendMedia := new(msg.ClientSendMessageMedia)
			unmarshalErr := clientSendMedia.Unmarshal(pmsg.Media)
			if unmarshalErr == nil { // TODO!!! fix this with some flag in pending message
				clientFile, err := repo.Files.GetMediaDocument(x.Message)
				logs.WarnOnErr("Error On GetMediaDocument", err)

				err = os.Rename(clientSendMedia.FilePath, repo.Files.GetFilePath(clientFile))
				if err != nil {
					logs.Error("Error On HandlePendingMessage (Rename)", zap.Error(err))
					return
				}
				_ = repo.Files.UnmarkAsUploaded(clientSendMedia.FileID)
			} else {
				logs.Error("Error On HandlePendingMessage", zap.Error(unmarshalErr))
			}
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

	uiexec.ExecUpdate(ctrl.updateReceivedCallback, msg.C_UpdateEnvelope, udpMsg)
}

func (ctrl *Controller) updateReadHistoryInbox(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateReadHistoryInbox)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	dialog, err := repo.Dialogs.Get(x.TeamID, x.Peer.ID, x.Peer.Type)
	if dialog == nil {
		logs.Error("SyncCtrl got error on UpdateReadHistoryInbox",
			zap.Int64("PeerID", x.Peer.ID),
			zap.Int32("PeerType", x.Peer.Type),
		)
	}

	logs.Info("SyncCtrl applies UpdateReadHistoryInbox",
		zap.Int64("MaxID", x.MaxID),
		zap.Int64("UpdateID", x.UpdateID),
		zap.Int64("PeerID", x.Peer.ID),
	)

	repo.Dialogs.UpdateReadInboxMaxID(ctrl.GetUserID(), x.TeamID, x.Peer.ID, x.Peer.Type, x.MaxID)
	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

func (ctrl *Controller) updateReadHistoryOutbox(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateReadHistoryOutbox)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	dialog, err := repo.Dialogs.Get(x.TeamID, x.Peer.ID, x.Peer.Type)
	if dialog == nil {
		logs.Error("SyncCtrl got error on UpdateReadHistoryOutbox",
			zap.Int64("PeerID", x.Peer.ID),
			zap.Int32("PeerType", x.Peer.Type),
		)
	}

	logs.Info("SyncCtrl applies UpdateReadHistoryOutbox",
		zap.Int64("MaxID", x.MaxID),
		zap.Int64("UpdateID", x.UpdateID),
		zap.Int64("PeerID", x.Peer.ID),
	)

	repo.Dialogs.UpdateReadOutboxMaxID(x.TeamID, x.Peer.ID, x.Peer.Type, x.MaxID)
	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

func (ctrl *Controller) updateMessageEdited(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateMessageEdited)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Info("SyncCtrl applies UpdateMessageEdited",
		zap.Int64("MessageID", x.Message.ID),
		zap.Int64("UpdateID", x.UpdateID),
	)

	repo.Messages.Save(x.Message)

	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

func (ctrl *Controller) updateMessageID(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	res := make([]*msg.UpdateEnvelope, 0)
	x := new(msg.UpdateMessageID)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Info("SyncCtrl applied UpdateMessageID",
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

	userMessage, _ := repo.Messages.Get(sent.MessageID)
	if userMessage != nil {
		logs.Info("SyncCtrl received UpdateMessageID after UpdateNewMessage",
			zap.Int64("MID", x.MessageID),
			zap.Int64("RandomID", x.RandomID),
			zap.String("Body", userMessage.Body),
		)

		if pm == nil {
			logs.Warn("But we couldn't find Pending Message",
				zap.Int64("RandomID", sent.RandomID),
				zap.Int64("RealID", sent.MessageID),
			)
			return res, nil
		}
		ctrl.deletePendingMessage(pm)
		return res, nil
	}

	if pm == nil {
		return nil, nil
	}
	logs.Info("SyncCtrl received UpdateMessageID before UpdateNewMessage",
		zap.Int64("MID", x.MessageID),
		zap.Int64("PendingID", pm.ID),
	)
	logs.Debug("SyncCtrl is going to save by realID")
	repo.PendingMessages.SaveByRealID(sent.RandomID, sent.MessageID)
	logs.Debug("SyncCtrl saved by realID")

	return res, nil
}
func (ctrl *Controller) deletePendingMessage(pm *msg.ClientPendingMessage) {
	// It means we have received the NewMessage
	update := &msg.UpdateMessagesDeleted{
		Peer:       &msg.Peer{ID: pm.PeerID, Type: pm.PeerType},
		MessageIDs: []int64{pm.ID},
	}
	bytes, _ := update.Marshal()

	updateEnvelope := &msg.UpdateEnvelope{
		Constructor: msg.C_UpdateMessagesDeleted,
		Update:      bytes,
		UpdateID:    0,
		Timestamp:   time.Now().Unix(),
	}

	uiexec.ExecUpdate(ctrl.updateReceivedCallback, msg.C_UpdateEnvelope, updateEnvelope)

	_ = repo.PendingMessages.Delete(pm.ID)
}

func (ctrl *Controller) updateNotifySettings(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateNotifySettings)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Info("SyncCtrl applies UpdateNotifySettings",
		zap.Int64("UpdateID", x.UpdateID),
	)

	repo.Dialogs.UpdateNotifySetting(x.TeamID, x.NotifyPeer.ID, x.NotifyPeer.Type, x.Settings)

	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

func (ctrl *Controller) updateDialogPinned(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateDialogPinned)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Info("SyncCtrl applies UpdateDialogPinned",
		zap.Int64("UpdateID", x.UpdateID),
	)

	repo.Dialogs.UpdatePinned(x)
	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

func (ctrl *Controller) updateUsername(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateUsername)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Info("SyncCtrl applies UpdateUsername",
		zap.Int64("UpdateID", x.UpdateID),
	)

	if x.UserID == ctrl.GetUserID() {
		ctrl.connInfo.ChangeUserID(x.UserID)
		ctrl.connInfo.ChangeUsername(x.Username)
		ctrl.connInfo.ChangeFirstName(x.FirstName)
		ctrl.connInfo.ChangeLastName(x.LastName)
		ctrl.connInfo.ChangePhone(x.Phone)
		ctrl.connInfo.ChangeBio(x.Bio)
		ctrl.connInfo.Save()
	}

	err = repo.Users.UpdateProfile(x.UserID, x.FirstName, x.LastName, x.Username, x.Bio, x.Phone)
	if err != nil {
		return nil, err
	}

	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

func (ctrl *Controller) updateMessagesDeleted(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateMessagesDeleted)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Info("SyncCtrl applies UpdateMessagesDeleted",
		zap.Int64("UpdateID", x.UpdateID),
	)

	repo.Messages.Delete(ctrl.GetUserID(), x.TeamID, x.Peer.ID, x.Peer.Type, x.MessageIDs...)

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

func (ctrl *Controller) updateGroupParticipantAdmin(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := &msg.UpdateGroupParticipantAdmin{}
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Info("SyncCtrl applies UpdateGroupParticipantAdmin",
		zap.Int64("UpdateID", x.UpdateID),
	)

	res := []*msg.UpdateEnvelope{u}
	repo.Groups.UpdateMemberType(x.GroupID, x.UserID, x.IsAdmin)
	return res, nil
}

func (ctrl *Controller) updateReadMessagesContents(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := &msg.UpdateReadMessagesContents{}
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Info("SyncCtrl applies UpdateReadMessagesContents",
		zap.Int64s("MessageIDs", x.MessageIDs),
		zap.Int64("UpdateID", x.UpdateID),
	)

	repo.Messages.SetContentRead(x.Peer.ID, int32(x.Peer.Type), x.MessageIDs)

	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

func (ctrl *Controller) updateUserPhoto(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateUserPhoto)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Info("SyncCtrl applies UpdateUserPhoto",
		zap.Int64("UpdateID", x.UpdateID),
	)

	if x.Photo != nil {
		_ = repo.Users.UpdatePhoto(x.UserID, x.Photo)
		repo.Users.SavePhotoGallery(x.UserID, x.Photo)
	}

	for _, photoID := range x.DeletedPhotoIDs {
		repo.Users.RemovePhotoGallery(x.UserID, photoID)
	}

	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

func (ctrl *Controller) updateGroupPhoto(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateGroupPhoto)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Info("SyncCtrl applies UpdateGroupPhoto",
		zap.Int64("GroupID", x.GroupID),
		zap.Int64("UpdateID", x.UpdateID),
	)

	if x.Photo != nil {
		repo.Groups.UpdatePhoto(x.GroupID, x.Photo)
	}

	if x.PhotoID != 0 {
		if x.Photo != nil && x.Photo.PhotoSmall.FileID != 0 {
			repo.Groups.SavePhotoGallery(x.GroupID, x.Photo)
		} else {
			repo.Groups.RemovePhotoGallery(x.GroupID, x.PhotoID)
		}
	}

	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

func (ctrl *Controller) updateGroupAdmins(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := &msg.UpdateGroupAdmins{}
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Info("SyncCtrl applies UpdateGroupAdmins",
		zap.Int64("GroupID", x.GroupID),
		zap.Int64("UpdateID", x.UpdateID),
	)

	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

func (ctrl *Controller) updateAccountPrivacy(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateAccountPrivacy)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Info("SyncCtrl applies UpdateAccountPrivacy",
		zap.Int64("UpdateID", x.UpdateID),
	)

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

	logs.Info("SyncCtrl applies UpdateDraftMessage",
		zap.Int64("UpdateID", x.UpdateID),
	)

	dialog, _ := repo.Dialogs.Get(x.Message.TeamID, x.Message.PeerID, int32(x.Message.PeerType))
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

	logs.Info("SyncCtrl applies UpdateDraftMessageCleared",
		zap.Int64("UpdateID", x.UpdateID),
	)

	dialog, _ := repo.Dialogs.Get(x.TeamID, x.Peer.ID, x.Peer.Type)

	if dialog != nil {
		dialog.Draft = nil
		repo.Dialogs.Save(dialog)
	}

	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

func (ctrl *Controller) updateLabelItemsAdded(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := &msg.UpdateLabelItemsAdded{}
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Info("SyncCtrl applies UpdateLabelItemsAdded",
		zap.Int64("UpdateID", x.UpdateID),
	)

	if len(x.MessageIDs) != 0 {
		err := repo.Labels.AddLabelsToMessages(x.LabelIDs, x.TeamID, x.Peer.ID, x.Peer.Type, x.MessageIDs)
		if err != nil {
			return nil, err
		}
		for _, labelID := range x.LabelIDs {
			bar := repo.Labels.GetFilled(labelID)
			for _, msgID := range x.MessageIDs {
				if msgID > bar.MaxID {
					repo.Labels.Fill(labelID, bar.MaxID, msgID)
				} else if msgID < bar.MinID {
					repo.Labels.Fill(labelID, msgID, bar.MinID)
				}
			}
		}

	}

	err = repo.Labels.Save(x.Labels...)
	if err != nil {
		return nil, err
	}
	return []*msg.UpdateEnvelope{u}, nil
}

func (ctrl *Controller) updateLabelItemsRemoved(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := &msg.UpdateLabelItemsRemoved{}
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Info("SyncCtrl applies UpdateLabelItemsRemoved",
		zap.Int64("UpdateID", x.UpdateID),
	)

	if len(x.MessageIDs) != 0 {
		err := repo.Labels.RemoveLabelsFromMessages(x.LabelIDs, x.TeamID, x.Peer.ID, x.Peer.Type, x.MessageIDs)
		if err != nil {
			return nil, err
		}
	}

	err = repo.Labels.Save(x.Labels...)
	if err != nil {
		return nil, err
	}
	return []*msg.UpdateEnvelope{u}, nil
}

func (ctrl *Controller) updateLabelSet(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := &msg.UpdateLabelSet{}
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Info("SyncCtrl applies UpdateLabelSet",
		zap.Int64("UpdateID", x.UpdateID),
	)

	err = repo.Labels.Save(x.Labels...)
	if err != nil {
		return nil, err
	}
	return []*msg.UpdateEnvelope{u}, nil
}

func (ctrl *Controller) updateLabelDeleted(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := &msg.UpdateLabelDeleted{}
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Info("SyncCtrl applies UpdateLabelDeleted",
		zap.Int64("UpdateID", x.UpdateID),
	)

	err = repo.Labels.Delete(x.LabelIDs...)
	if err != nil {
		return nil, err
	}
	return []*msg.UpdateEnvelope{u}, nil
}

func (ctrl *Controller) updateUserBlocked(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := &msg.UpdateUserBlocked{}
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Info("SyncCtrl applies UpdateUserBlocked",
		zap.Int64("UpdateID", x.UpdateID),
	)

	err = repo.Users.UpdateBlocked(x.UserID, x.Blocked)
	if err != nil {
		return nil, err
	}
	return []*msg.UpdateEnvelope{u}, nil
}
