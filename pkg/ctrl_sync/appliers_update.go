package syncCtrl

import (
	"fmt"
	messageHole "git.ronaksoftware.com/ronak/riversdk/pkg/message_hole"
	"time"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"go.uber.org/zap"
)

// updateNewMessage
func (ctrl *Controller) updateNewMessage(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	x := new(msg.UpdateNewMessage)
	err := x.Unmarshal(u.Update)
	if err != nil {
		logs.Error("updateNewMessage()-> Unmarshal()", zap.Error(err))
		return []*msg.UpdateEnvelope{}
	}
	logs.Info("SyncController::updateNewMessage",
		zap.Int64("MessageID", x.Message.ID),
		zap.Int64("UpdateID", x.UpdateID),
	)

	// used messageType to identify client & server messages on Media thingy
	x.Message.MessageType = 1

	dialog := repo.Dialogs.Get(x.Message.PeerID, x.Message.PeerType)
	if dialog == nil {
		messageHole.InsertHole(x.Message.PeerID, x.Message.PeerType, 0, x.Message.ID-1)
		messageHole.SetUpperFilled(x.Message.PeerID, x.Message.PeerType, x.Message.ID)

		// make sure to created the message hole b4 creating dialog
		dialog = &msg.Dialog{
			PeerID:         x.Message.PeerID,
			PeerType:       x.Message.PeerType,
			TopMessageID:   x.Message.ID,
			UnreadCount:    0,
			MentionedCount: 0,
			AccessHash:     x.AccessHash,
		}
		err := repo.Dialogs.SaveNew(dialog, x.Message.CreatedOn)
		if err != nil {
			logs.Error("updateNewMessage() -> onSuccessCallback() -> SaveDialog() ",
				zap.String("Error", err.Error()),
				zap.String("Dialog", fmt.Sprintf("%v", dialog)),
			)
		}

		if err != nil {
			logs.Error("updateNewMessage()-> SaveDialog()", zap.Error(err))
		}
	}
	// save user if does not exist
	repo.Users.Save(x.Sender)

	err = repo.Messages.SaveNew(x.Message, dialog, ctrl.connInfo.PickupUserID())
	if err != nil {
		logs.Warn("updateNewMessage()-> SaveNew()", zap.Error(err))
	}
	messageHole.SetUpperFilled(x.Message.PeerID, x.Message.PeerType, x.Message.ID)

	// bug : sometime server do not sends access hash
	if x.AccessHash > 0 {
		// update users access hash
		err := repo.Users.UpdateAccessHash(x.AccessHash, x.Message.PeerID, x.Message.PeerType)
		if err != nil {
			logs.Error("updateNewMessage() -> Users.updateAccessHash()", zap.Error(err))
		}
	}
	res := make([]*msg.UpdateEnvelope, 0)
	// Prevent calling external delegate
	if !ctrl.isDeliveredMessage(x.Message.ID) {
		res = append(res, u)
	}

	// handle Message's Action
	ctrl.handleMessageAction(x, u, res)

	// handle Message's Media
	if int32(x.Message.MediaType) > 0 {
		go ctrl.extractMessagesMedia(x.Message)
	}
	return res
}
func (ctrl *Controller) handleMessageAction(x *msg.UpdateNewMessage, u *msg.UpdateEnvelope, res []*msg.UpdateEnvelope) {
	switch x.Message.MessageAction {
	case domain.MessageActionNope:
		// Do nothing
	case domain.MessageActionContactRegistered:
		go ctrl.ContactImportFromServer()
	case domain.MessageActionGroupCreated:
		// this will be handled by upper level on UpdateContainer
	case domain.MessageActionGroupAddUser:
		// TODO : this should be implemented

	case domain.MessageActionGroupDeleteUser:
		act := new(msg.MessageActionGroupAddUser)
		err := act.Unmarshal(x.Message.MessageActionData)
		if err != nil {
			logs.Error("updateNewMessage() -> MessageActionGroupDeleteUser Failed to Parse", zap.Error(err))
		}

		repo.Groups.DeleteMemberMany(x.Message.PeerID, act.UserIDs)

		// Check if user left (deleted him/her self from group) remove its Group, Dialog and its MessagesPending
		selfUserID := ctrl.connInfo.PickupUserID()
		userLeft := false
		for _, v := range act.UserIDs {
			if v == selfUserID {
				userLeft = true
				break
			}
		}
		if userLeft {
			// Delete Group		NOT REQUIRED
			// Delete Dialog	NOT REQUIRED
			// Delete PendingMessage
			deletedMsgs, err := repo.PendingMessages.DeletePeerAllMessages(x.Message.PeerID, x.Message.PeerType)
			if err != nil {
				logs.Error("River::groupDeleteUser()-> DeleteGroupPendingMessage()", zap.Error(err))
			} else {
				buff, err := deletedMsgs.Marshal()
				if err != nil {
					logs.Error("River::groupDeleteUser()-> Unmarshal ClientUpdateMessagesDeleted", zap.Error(err))
				} else {
					udp := new(msg.UpdateEnvelope)
					udp.Constructor = msg.C_ClientUpdateMessagesDeleted
					udp.Update = buff
					udp.UCount = 1
					udp.Timestamp = u.Timestamp
					udp.UpdateID = u.UpdateID
					res = append(res, udp)
				}
			}

		}

	case domain.MessageActionGroupTitleChanged:
	// this will be handled by upper level on UpdateContainer
	case domain.MessageActionClearHistory:
		act := new(msg.MessageActionClearHistory)
		err := act.Unmarshal(x.Message.MessageActionData)
		if err != nil {
			logs.Error("updateNewMessage() -> MessageActionClearHistory Failed to Parse", zap.Error(err))
		}

		err = repo.Messages.DeleteDialogMessage(x.Message.PeerID, x.Message.PeerType, act.MaxID)
		if err != nil {
			logs.Error("updateNewMessage() -> DeleteDialogMessage() Failed", zap.Error(err))
		}
		// Delete Scroll Position
		err = repo.MessagesExtra.DeleteScrollID(x.Message.PeerID, x.Message.PeerType)
		if err != nil {
			logs.Error("updateNewMessage() -> DeleteScrollID() Failed", zap.Error(err))
		}

		if act.Delete {
			// Delete Dialog
			err = repo.Dialogs.Delete(x.Message.PeerID, x.Message.PeerType)
			if err != nil {
				logs.Error("updateNewMessage() -> Dialogs.Delete() Failed", zap.Error(err))
			}
			// Delete Group
			err = repo.Groups.Delete(x.Message.PeerID)
			if err != nil {
				logs.Error("updateNewMessage() -> Groups.Delete() Failed", zap.Error(err))
			}

			// Delete Participants
			repo.Groups.DeleteAllMembers(x.Message.PeerID)

		} else {
			// get dialog and create first hole
			dtoDlg := repo.Dialogs.Get(x.Message.PeerID, x.Message.PeerType)
			if dtoDlg != nil {
				messageHole.InsertHole(dtoDlg.PeerID, dtoDlg.PeerType, 0, dtoDlg.TopMessageID-1)
				messageHole.SetUpperFilled(dtoDlg.PeerID, dtoDlg.PeerType, dtoDlg.TopMessageID)

			}
		}
	}
}

// updateReadHistoryInbox
func (ctrl *Controller) updateReadHistoryInbox(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	x := new(msg.UpdateReadHistoryInbox)
	_ = x.Unmarshal(u.Update)

	dialog := repo.Dialogs.Get(x.Peer.ID, x.Peer.Type)
	if dialog == nil {
		return []*msg.UpdateEnvelope{}
	}
	logs.Info("SyncController::updateReadHistoryInbox",
		zap.Int64("MaxID", x.MaxID),
		zap.Int64("UpdateID", x.UpdateID),
		zap.Int64("PeerID", x.Peer.ID),
	)

	repo.Dialogs.UpdateReadInboxMaxID(ctrl.userID, x.Peer.ID, x.Peer.Type, x.MaxID)
	res := []*msg.UpdateEnvelope{u}
	return res
}

// updateReadHistoryOutbox
func (ctrl *Controller) updateReadHistoryOutbox(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	x := new(msg.UpdateReadHistoryOutbox)
	_ = x.Unmarshal(u.Update)
	logs.Info("SyncController::updateReadHistoryOutbox",
		zap.Int64("MaxID", x.MaxID),
		zap.Int64("UpdateID", x.UpdateID),
		zap.Int64("PeerID", x.Peer.ID),
	)

	repo.Dialogs.UpdateReadOutboxMaxID(x.Peer.ID, x.Peer.Type, x.MaxID)
	res := []*msg.UpdateEnvelope{u}
	return res
}

// updateMessageEdited
func (ctrl *Controller) updateMessageEdited(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	x := new(msg.UpdateMessageEdited)
	_ = x.Unmarshal(u.Update)
	logs.Info("SyncController::updateMessageEdited",
		zap.Int64("MessageID", x.Message.ID),
	)
	_ = repo.Messages.Save(x.Message)

	res := []*msg.UpdateEnvelope{u}
	return res
}

// updateMessageID
func (ctrl *Controller) updateMessageID(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	x := new(msg.UpdateMessageID)
	_ = x.Unmarshal(u.Update)
	logs.Info("SyncController::updateMessageID",
		zap.Int64("RandomID", x.RandomID),
		zap.Int64("MessageID", x.MessageID),
	)

	if ctrl.isDeliveredMessage(x.MessageID) {
		logs.Debug("updateMessageID() message is already delivered")
		return []*msg.UpdateEnvelope{}
	}

	msgEnvelop := new(msg.MessageEnvelope)
	msgEnvelop.Constructor = msg.C_MessageEnvelope

	sent := new(msg.MessagesSent)
	sent.MessageID = x.MessageID
	sent.RandomID = x.RandomID
	sent.CreatedOn = time.Now().Unix()

	// used message randomID as requestID of pending message se we can retrieve it here
	msgEnvelop.RequestID = uint64(x.RandomID)
	msgEnvelop.Message, _ = sent.Marshal()
	ctrl.messageSent(msgEnvelop)
	res := []*msg.UpdateEnvelope{u}
	return res
}

// updateNotifySettings
func (ctrl *Controller) updateNotifySettings(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	logs.Info("SyncController::updateNotifySettings()")
	x := new(msg.UpdateNotifySettings)
	_ = x.Unmarshal(u.Update)

	repo.Dialogs.UpdateNotifySetting(x.NotifyPeer.ID, x.NotifyPeer.Type, x.Settings)

	res := []*msg.UpdateEnvelope{u}
	return res
}

// updateDialogPinned
func (ctrl *Controller) updateDialogPinned(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	logs.Info("SyncController::updateDialogPinned()")
	x := new(msg.UpdateDialogPinned)
	_ = x.Unmarshal(u.Update)
	repo.Dialogs.UpdatePinned(x)
	res := []*msg.UpdateEnvelope{u}
	return res
}

// updateUsername
func (ctrl *Controller) updateUsername(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	logs.Info("SyncController::updateUsername()")
	x := new(msg.UpdateUsername)
	x.Unmarshal(u.Update)

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
	return res
}

// updateMessagesDeleted
func (ctrl *Controller) updateMessagesDeleted(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	logs.Info("SyncController::updateMessagesDeleted()")

	x := new(msg.UpdateMessagesDeleted)
	_ = x.Unmarshal(u.Update)

	for _, msgID := range x.MessageIDs {
		_ = repo.Messages.DeleteDialogMessage(x.Peer.ID, x.Peer.Type, msgID)
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
	return res
}

// updateGroupParticipantAdmin
func (ctrl *Controller) updateGroupParticipantAdmin(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	logs.Info("SyncController::updateGroupParticipantAdmin()")

	x := new(msg.UpdateGroupParticipantAdmin)
	x.Unmarshal(u.Update)

	res := []*msg.UpdateEnvelope{u}

	err := repo.Groups.UpdateMemberType(x.GroupID, x.UserID, x.IsAdmin)
	if err != nil {
		logs.Error("updateGroupParticipantAdmin()-> UpdateGroupMemberType()", zap.Error(err))
	}
	return res
}

// updateReadMessagesContents
func (ctrl *Controller) updateReadMessagesContents(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	x := new(msg.UpdateReadMessagesContents)
	_ = x.Unmarshal(u.Update)
	logs.Info("SyncController::updateReadMessagesContents",
		zap.Int64s("MessageIDs", x.MessageIDs),
	)

	repo.Messages.SetContentRead(x.Peer.ID, int32(x.Peer.Type), x.MessageIDs)

	res := []*msg.UpdateEnvelope{u}
	return res
}

// updateUserPhoto
func (ctrl *Controller) updateUserPhoto(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	logs.Info("SyncController::updateUserPhoto()")

	x := new(msg.UpdateUserPhoto)
	_ = x.Unmarshal(u.Update)

	repo.Users.UpdatePhoto(x)

	res := []*msg.UpdateEnvelope{u}
	return res
}

// updateGroupPhoto
func (ctrl *Controller) updateGroupPhoto(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	x := new(msg.UpdateGroupPhoto)
	_ = x.Unmarshal(u.Update)

	logs.Info("SyncController::updateGroupPhoto",
		zap.Int64("GroupID", x.GroupID),
	)

	err := repo.Groups.UpdatePhoto(x)
	if err != nil {
		logs.Error("updateGroupPhoto()-> UpdateGroupPhoto()", zap.Error(err))
	}

	res := []*msg.UpdateEnvelope{u}
	return res
}

// updateTooLong indicate that client updates exceed from server cache size so we should re-sync client with server
func (ctrl *Controller) updateTooLong(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	logs.Info("SyncController::updateTooLong()")

	go ctrl.sync()

	res := make([]*msg.UpdateEnvelope, 0)
	return res
}
