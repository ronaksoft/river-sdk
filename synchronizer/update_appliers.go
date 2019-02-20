package synchronizer

import (
	"time"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/logs"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/repo"
	"go.uber.org/zap"
)

// updateNewMessage
func (ctrl *Controller) updateNewMessage(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	logs.Debug("updateNewMessage() applier")
	x := new(msg.UpdateNewMessage)
	err := x.Unmarshal(u.Update)
	if err != nil {
		logs.Debug("updateNewMessage()-> Unmarshal()",
			zap.String("Error", err.Error()),
		)
		return []*msg.UpdateEnvelope{}
	}

	// used messageType to identify client & server messages on Media thingy
	x.Message.MessageType = 1

	dialog := repo.Ctx().Dialogs.GetDialog(x.Message.PeerID, x.Message.PeerType)
	if dialog == nil {
		// // this will be handled by repo
		// unreadCount := int32(0)
		// if x.Message.SenderID != ctrl.UserID {
		// 	unreadCount = 1
		// }
		dialog = &msg.Dialog{
			PeerID:       x.Message.PeerID,
			PeerType:     x.Message.PeerType,
			TopMessageID: x.Message.ID,
			UnreadCount:  0,
			AccessHash:   x.AccessHash,
		}
		err := repo.Ctx().Dialogs.SaveDialog(dialog, x.Message.CreatedOn)
		if err != nil {
			logs.Debug("updateNewMessage()-> SaveDialog()",
				zap.String("Error", err.Error()),
			)
		}
	}
	// save user if does not exist
	repo.Ctx().Users.SaveUser(x.Sender)

	// if the sender is not myself increase dialog counter else just save message
	if x.Message.SenderID != ctrl.UserID {
		err := repo.Ctx().Messages.SaveNewMessage(x.Message, dialog, ctrl.connInfo.PickupUserID())
		if err != nil {
			logs.Debug("updateNewMessage()-> SaveNewMessage()",
				zap.String("Error", err.Error()),
			)
		}
	} else {
		err := repo.Ctx().Messages.SaveSelfMessage(x.Message, dialog)
		if err != nil {
			logs.Debug("updateNewMessage()-> SaveSelfMessage()",
				zap.String("Error", err.Error()),
			)
		}
	}

	logs.Warn("UpdateNewMessage AccessHash ",
		zap.Uint64("AccessHash Uint64", x.AccessHash),
		zap.Int64("AccessHash Int64", int64(x.AccessHash)),
		zap.String("Body", x.Message.Body),
	)

	// bug : sometime server do not sends accesshash
	if x.AccessHash > 0 {
		// update users access hash
		err := repo.Ctx().Users.UpdateAccessHash(int64(x.AccessHash), x.Message.PeerID, x.Message.PeerType)
		if err != nil {
			logs.Debug("updateNewMessage() -> Users.UpdateAccessHash()",
				zap.String("Error", err.Error()),
			)
		}
		err = repo.Ctx().Dialogs.UpdateAccessHash(int64(x.AccessHash), x.Message.PeerID, x.Message.PeerType)

		if err != nil {
			logs.Debug("updateNewMessage() -> Dialogs.UpdateAccessHash()",
				zap.String("Error", err.Error()),
			)
		}
	}
	res := make([]*msg.UpdateEnvelope, 0)
	// Perevent calling external delegate
	if !ctrl.isDeliveredMessage(x.Message.ID) {
		res = append(res, u)
	}

	// Parse message action and call required appliers
	switch x.Message.MessageAction {
	case domain.MessageActionNope:
		// Do nothing
	case domain.MessageActionContactRegistered:
		// Not implemented
	case domain.MessageActionGroupCreated:
		// this will be handled by upper level on UpdateContainer
	case domain.MessageActionGroupAddUser:
		// TODO : this should be implemented

	case domain.MessageActionGroupDeleteUser:
		act := new(msg.MessageActionGroupAddUser)
		err := act.Unmarshal(x.Message.MessageActionData)
		if err != nil {
			logs.Debug("updateNewMessage() -> MessageActionGroupDeleteUser Failed to Parse", zap.String("Error", err.Error()))
		}
		err = repo.Ctx().Groups.DeleteGroupMemberMany(x.Message.PeerID, act.UserIDs)
		if err != nil {
			logs.Debug("updateNewMessage() -> DeleteGroupMemberMany() Failed", zap.String("Error", err.Error()))
		}

		// Check if user left (deleted him/her self from group) remove its Group, Dialog and its PendingMessages
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
			deletedMsgs, err := repo.Ctx().PendingMessages.DeletePeerAllMessages(x.Message.PeerID, x.Message.PeerType)
			if err != nil {
				logs.Debug("River::groupDeleteUser()-> DeleteGroupPendingMessage()",
					zap.String("Error", err.Error()),
				)
			} else {
				buff, err := deletedMsgs.Marshal()
				if err != nil {
					logs.Debug("River::groupDeleteUser()-> Unmarshal ClientUpdateMessagesDeleted",
						zap.String("Error", err.Error()),
					)
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
			logs.Debug("updateNewMessage() -> MessageActionClearHistory Failed to Parse", zap.String("Error", err.Error()))
		}

		err = repo.Ctx().Messages.DeleteDialogMessage(x.Message.PeerID, x.Message.PeerType, act.MaxID)
		if err != nil {
			logs.Debug("updateNewMessage() -> DeleteDialogMessage() Failed", zap.String("Error", err.Error()))
		}
		if act.Delete {
			// Delete Dialog
			repo.Ctx().Dialogs.Delete(x.Message.PeerID, x.Message.PeerType)
			// Delete Group
			repo.Ctx().Groups.Delete(x.Message.PeerID)
			// Delete Participants
			repo.Ctx().Groups.DeleteAllGroupMember(x.Message.PeerID)
			//delete MessageHole
			deleteMessageHole(x.Message.PeerID)
		} else {
			// get dialog and create first hole
			dtoDlg := repo.Ctx().Dialogs.GetDialog(x.Message.PeerID, x.Message.PeerType)
			if dtoDlg != nil {
				deleteMessageHole(x.Message.PeerID)
				createMessageHole(dtoDlg.PeerID, 0, dtoDlg.TopMessageID-1)
			}
		}
	}

	// handle Media message
	if int32(x.Message.MediaType) > 0 {
		go handleMediaMessage(x.Message)
	}
	return res
}

// updateReadHistoryInbox
func (ctrl *Controller) updateReadHistoryInbox(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	logs.Debug("updateReadHistoryInbox() applier")
	x := new(msg.UpdateReadHistoryInbox)
	x.Unmarshal(u.Update)
	dialog := repo.Ctx().Dialogs.GetDialog(x.Peer.ID, x.Peer.Type)
	if dialog == nil {
		return []*msg.UpdateEnvelope{}
	}

	err := repo.Ctx().Dialogs.UpdateReadInboxMaxID(ctrl.UserID, x.Peer.ID, x.Peer.Type, x.MaxID)
	if err != nil {
		logs.Debug("updateReadHistoryInbox() -> UpdateReadInboxMaxID()",
			zap.String("Error", err.Error()),
		)
	}
	res := []*msg.UpdateEnvelope{u}
	return res
}

// updateReadHistoryOutbox
func (ctrl *Controller) updateReadHistoryOutbox(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	logs.Debug("updateReadHistoryOutbox() applier")
	x := new(msg.UpdateReadHistoryOutbox)
	x.Unmarshal(u.Update)
	err := repo.Ctx().Dialogs.UpdateReadOutboxMaxID(x.Peer.ID, x.Peer.Type, x.MaxID)
	if err != nil {
		logs.Debug("updateReadHistoryOutbox() -> UpdateReadOutboxMaxID()",
			zap.String("Error", err.Error()),
		)
	}
	res := []*msg.UpdateEnvelope{u}
	return res
}

// updateMessageEdited
func (ctrl *Controller) updateMessageEdited(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	logs.Debug("updateMessageEdited() applier")
	x := new(msg.UpdateMessageEdited)
	x.Unmarshal(u.Update)
	err := repo.Ctx().Messages.SaveMessage(x.Message)
	if err != nil {
		logs.Debug("updateMessageEdited() -> SaveMessage()",
			zap.String("Error", err.Error()),
		)
	}
	res := []*msg.UpdateEnvelope{u}
	return res
}

// updateMessageID
func (ctrl *Controller) updateMessageID(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	logs.Debug("updateMessageID() applier")
	x := new(msg.UpdateMessageID)
	x.Unmarshal(u.Update)

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
	logs.Debug("updateNotifySettings() applier")
	x := new(msg.UpdateNotifySettings)
	x.Unmarshal(u.Update)

	err := repo.Ctx().Dialogs.UpdateNotifySetting(x)
	if err != nil {
		logs.Debug("updateNotifySettings() -> Dialogs.UpdateNotifySettings()",
			zap.String("Error", err.Error()),
		)
	}
	res := []*msg.UpdateEnvelope{u}
	return res
}

// updateUsername
func (ctrl *Controller) updateUsername(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	logs.Debug("updateUsername() applier")
	x := new(msg.UpdateUsername)
	x.Unmarshal(u.Update)

	ctrl.connInfo.ChangeUserID(x.UserID)
	ctrl.connInfo.ChangeUsername(x.Username)
	ctrl.connInfo.ChangeFirstName(x.FirstName)
	ctrl.connInfo.ChangeLastName(x.LastName)
	ctrl.connInfo.ChangeBio(x.Bio)
	ctrl.connInfo.Save()

	err := repo.Ctx().Users.UpdateUsername(x)
	if err != nil {
		logs.Debug("updateUsername() error save to DB", zap.Error(err))
	}

	res := []*msg.UpdateEnvelope{u}
	return res
}

// updateMessagesDeleted
func (ctrl *Controller) updateMessagesDeleted(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	logs.Debug("updateMessagesDeleted() applier")

	x := new(msg.UpdateMessagesDeleted)
	x.Unmarshal(u.Update)

	udps, err := repo.Ctx().Messages.DeleteManyAndReturnClientUpdate(x.MessageIDs)
	if err != nil {
		logs.Debug("updateMessagesDeleted() -> DeleteMany()",
			zap.String("Error", err.Error()),
		)
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
	logs.Debug("updateGroupParticipantAdmin() applier")

	x := new(msg.UpdateGroupParticipantAdmin)
	x.Unmarshal(u.Update)

	res := []*msg.UpdateEnvelope{u}

	err := repo.Ctx().Groups.UpdateGroupMemberType(x.GroupID, x.UserID, x.IsAdmin)
	if err != nil {
		logs.Debug("updateGroupParticipantAdmin()-> UpdateGroupMemberType()",
			zap.String("Error", err.Error()),
		)
	}
	return res
}

// updateReadMessagesContents
func (ctrl *Controller) updateReadMessagesContents(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	logs.Debug("updateReadMessagesContents() applier")

	x := new(msg.UpdateReadMessagesContents)
	x.Unmarshal(u.Update)

	err := repo.Ctx().Messages.SetContentRead(x.MessageIDs)
	if err != nil {
		logs.Debug("updateReadMessagesContents()-> SetContentRead()",
			zap.String("Error", err.Error()),
		)
	}

	res := []*msg.UpdateEnvelope{u}
	return res
}

// updateUserPhoto
func (ctrl *Controller) updateUserPhoto(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	logs.Debug("updateUserPhoto() applier")

	x := new(msg.UpdateUserPhoto)
	x.Unmarshal(u.Update)

	err := repo.Ctx().Users.SaveUserPhoto(x)
	if err != nil {
		logs.Debug("updateUserPhoto()-> SaveUserPhoto()",
			zap.String("Error", err.Error()),
		)
	}

	res := []*msg.UpdateEnvelope{u}
	return res
}

// updateGroupPhoto
func (ctrl *Controller) updateGroupPhoto(u *msg.UpdateEnvelope) []*msg.UpdateEnvelope {
	logs.Debug("updateGroupPhoto() applier")

	x := new(msg.UpdateGroupPhoto)
	x.Unmarshal(u.Update)

	err := repo.Ctx().Groups.UpdateGroupPhoto(x)
	if err != nil {
		logs.Debug("updateGroupPhoto()-> UpdateGroupPhoto()",
			zap.String("Error", err.Error()),
		)
	}

	res := []*msg.UpdateEnvelope{u}
	return res
}
