package synchronizer

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
	logs.Info("SyncController::updateNewMessage()")
	x := new(msg.UpdateNewMessage)
	err := x.Unmarshal(u.Update)
	if err != nil {
		logs.Error("updateNewMessage()-> Unmarshal()", zap.Error(err))
		return []*msg.UpdateEnvelope{}
	}

	// used messageType to identify client & server messages on Media thingy
	x.Message.MessageType = 1

	dialog := repo.Dialogs.GetDialog(x.Message.PeerID, x.Message.PeerType)
	if dialog == nil {
		_ = messageHole.InsertHole(x.Message.PeerID, x.Message.PeerType, 0, x.Message.ID-1)

		// make sure to created the message hole b4 creating dialog
		dialog = &msg.Dialog{
			PeerID:       x.Message.PeerID,
			PeerType:     x.Message.PeerType,
			TopMessageID: x.Message.ID,
			UnreadCount:  0,
			AccessHash:   x.AccessHash,
		}
		err := repo.Dialogs.SaveDialog(dialog, x.Message.CreatedOn)
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
	_ = repo.Users.SaveUser(x.Sender)

	// if the sender is not myself increase dialog counter else just save message
	if x.Message.SenderID != ctrl.userID {
		err := repo.Messages.SaveNewMessage(x.Message, dialog, ctrl.connInfo.PickupUserID())
		if err != nil {
			logs.Error("updateNewMessage()-> SaveNewMessage()", zap.Error(err))
		}
	} else {
		err := repo.Messages.SaveSelfMessage(x.Message, dialog)
		if err != nil {
			logs.Error("updateNewMessage()-> SaveSelfMessage()", zap.Error(err))
		}
	}

	_ = messageHole.SetUpperFilled(x.Message.PeerID, x.Message.PeerType, x.Message.ID)

	// bug : sometime server do not sends access hash
	if x.AccessHash > 0 {
		// update users access hash
		err := repo.Users.UpdateAccessHash(int64(x.AccessHash), x.Message.PeerID, x.Message.PeerType)
		if err != nil {
			logs.Error("updateNewMessage() -> Users.UpdateAccessHash()", zap.Error(err))
		}
		err = repo.Dialogs.UpdateAccessHash(int64(x.AccessHash), x.Message.PeerID, x.Message.PeerType)
		if err != nil {
			logs.Error("updateNewMessage() -> Dialogs.UpdateAccessHash()", zap.Error(err))
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
		go extractMessagesMedia(x.Message)
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
		err = repo.Groups.DeleteGroupMemberMany(x.Message.PeerID, act.UserIDs)
		if err != nil {
			logs.Error("updateNewMessage() -> DeleteGroupMemberMany() Failed", zap.Error(err))
		}

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
			err = repo.Groups.DeleteAllGroupMember(x.Message.PeerID)
			if err != nil {
				logs.Error("updateNewMessage() -> Groups.DeleteAllGroupMember() Failed", zap.Error(err))
			}
		} else {
			// get dialog and create first hole
			dtoDlg := repo.Dialogs.GetDialog(x.Message.PeerID, x.Message.PeerType)
			if dtoDlg != nil {
				_ = messageHole.InsertHole(dtoDlg.PeerID, dtoDlg.PeerType, 0, dtoDlg.TopMessageID-1)

			}
		}
	}
}

// updateReadHistoryInbox
func (ctrl *Controller) updateReadHistoryInbox(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	logs.Info("SyncController::updateReadHistoryInbox()")
	x := new(msg.UpdateReadHistoryInbox)
	x.Unmarshal(u.Update)
	dialog := repo.Dialogs.GetDialog(x.Peer.ID, x.Peer.Type)
	if dialog == nil {
		return []*msg.UpdateEnvelope{}
	}

	err := repo.Dialogs.UpdateReadInboxMaxID(ctrl.userID, x.Peer.ID, x.Peer.Type, x.MaxID)
	if err != nil {
		logs.Error("updateReadHistoryInbox() -> UpdateReadInboxMaxID()", zap.Error(err))
	}
	res := []*msg.UpdateEnvelope{u}
	return res
}

// updateReadHistoryOutbox
func (ctrl *Controller) updateReadHistoryOutbox(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	logs.Info("SyncController::updateReadHistoryOutbox()")
	x := new(msg.UpdateReadHistoryOutbox)
	x.Unmarshal(u.Update)
	err := repo.Dialogs.UpdateReadOutboxMaxID(x.Peer.ID, x.Peer.Type, x.MaxID)
	if err != nil {
		logs.Error("updateReadHistoryOutbox() -> UpdateReadOutboxMaxID()", zap.Error(err))
	}
	res := []*msg.UpdateEnvelope{u}
	return res
}

// updateMessageEdited
func (ctrl *Controller) updateMessageEdited(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	logs.Info("SyncController::updateMessageEdited()")
	x := new(msg.UpdateMessageEdited)
	x.Unmarshal(u.Update)
	err := repo.Messages.SaveMessage(x.Message)
	if err != nil {
		logs.Error("updateMessageEdited() -> SaveMessage()", zap.Error(err))
	}
	res := []*msg.UpdateEnvelope{u}
	return res
}

// updateMessageID
func (ctrl *Controller) updateMessageID(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	logs.Info("SyncController::updateMessageID()")

	x := new(msg.UpdateMessageID)
	x.Unmarshal(u.Update)

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

	// used message randomID as requestID of pending message se we can retrive it here
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
	x.Unmarshal(u.Update)

	err := repo.Dialogs.UpdateNotifySetting(x)
	if err != nil {
		logs.Error("updateNotifySettings() -> Dialogs.UpdateNotifySettings()", zap.Error(err))
	}
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

	err := repo.Users.UpdateUsername(x)
	if err != nil {
		logs.Error("updateUsername() error save to DB", zap.Error(err))
	}

	res := []*msg.UpdateEnvelope{u}
	return res
}

// updateMessagesDeleted
func (ctrl *Controller) updateMessagesDeleted(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	logs.Info("SyncController::updateMessagesDeleted()")

	x := new(msg.UpdateMessagesDeleted)
	x.Unmarshal(u.Update)

	udps, err := repo.Messages.DeleteManyAndReturnClientUpdate(x.MessageIDs)
	if err != nil {
		logs.Error("updateMessagesDeleted() -> DeleteMany()", zap.Error(err))
	}

	res := []*msg.UpdateEnvelope{u}

	for _, v := range udps {
		udp := new(msg.UpdateEnvelope)
		udp.Constructor = msg.C_ClientUpdateMessagesDeleted
		udp.Timestamp = u.Timestamp
		udp.UCount = 1
		udp.UpdateID = u.UpdateID
		udp.Update, _ = v.Marshal()
		res = append(res, udp)
	}

	return res
}

// updateGroupParticipantAdmin
func (ctrl *Controller) updateGroupParticipantAdmin(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	logs.Info("SyncController::updateGroupParticipantAdmin()")

	x := new(msg.UpdateGroupParticipantAdmin)
	x.Unmarshal(u.Update)

	res := []*msg.UpdateEnvelope{u}

	err := repo.Groups.UpdateGroupMemberType(x.GroupID, x.UserID, x.IsAdmin)
	if err != nil {
		logs.Error("updateGroupParticipantAdmin()-> UpdateGroupMemberType()", zap.Error(err))
	}
	return res
}

// updateReadMessagesContents
func (ctrl *Controller) updateReadMessagesContents(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	logs.Info("SyncController::updateReadMessagesContents()")

	x := new(msg.UpdateReadMessagesContents)
	x.Unmarshal(u.Update)

	err := repo.Messages.SetContentRead(x.MessageIDs)
	if err != nil {
		logs.Error("updateReadMessagesContents()-> SetContentRead()", zap.Error(err))
	}

	res := []*msg.UpdateEnvelope{u}
	return res
}

// updateUserPhoto
func (ctrl *Controller) updateUserPhoto(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	logs.Info("SyncController::updateUserPhoto()")

	x := new(msg.UpdateUserPhoto)
	x.Unmarshal(u.Update)

	err := repo.Users.SaveUserPhoto(x)
	if err != nil {
		logs.Error("updateUserPhoto()-> SaveUserPhoto()", zap.Error(err))
	}

	res := []*msg.UpdateEnvelope{u}
	return res
}

// updateGroupPhoto
func (ctrl *Controller) updateGroupPhoto(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	logs.Info("SyncController::updateGroupPhoto()")

	x := new(msg.UpdateGroupPhoto)
	x.Unmarshal(u.Update)

	err := repo.Groups.UpdateGroupPhoto(x)
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
