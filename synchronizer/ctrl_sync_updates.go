package synchronizer

import (
	"time"

	"git.ronaksoftware.com/ronak/riversdk/configs"
	"git.ronaksoftware.com/ronak/riversdk/log"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/repo"
	"go.uber.org/zap"
)

// updateNewMessage
func (ctrl *SyncController) updateNewMessage(u *msg.UpdateEnvelope) (passToExternalhandler bool) {
	x := new(msg.UpdateNewMessage)
	x.Unmarshal(u.Update)
	dialog := repo.Ctx().Dialogs.GetDialog(x.Message.PeerID, x.Message.PeerType)
	if dialog == nil {
		unreadCount := int32(0)
		if x.Message.SenderID != ctrl.UserID {
			unreadCount = 1
		}
		dialog = &msg.Dialog{
			PeerID:       x.Message.PeerID,
			PeerType:     x.Message.PeerType,
			TopMessageID: x.Message.ID,
			UnreadCount:  unreadCount,
			AccessHash:   x.AccessHash,
		}
		repo.Ctx().Dialogs.SaveDialog(dialog, x.Message.CreatedOn)
	}
	// save user if does not exist
	repo.Ctx().Users.SaveUser(x.Sender)

	// if the sender is not myself increase dialog counter else just save message
	if x.Message.SenderID != ctrl.UserID {
		repo.Ctx().Messages.SaveNewMessage(x.Message, dialog, configs.Get().UserID)
	} else {
		repo.Ctx().Messages.SaveSelfMessage(x.Message, dialog)
	}

	// bug : sometime server do not sends accesshash
	if x.AccessHash > 0 {
		// update users access hash
		err := repo.Ctx().Users.UpdateAccessHash(int64(x.AccessHash), x.Message.PeerID, x.Message.PeerType)
		if err != nil {
			log.LOG.Error("updateNewMessage::Users.SaveUser()",
				zap.String("Error", err.Error()),
			)
		}
		err = repo.Ctx().Dialogs.UpdateAccessHash(int64(x.AccessHash), x.Message.PeerID, x.Message.PeerType)

		if err != nil {
			log.LOG.Error("updateNewMessage::Dialogs.UpdateAccessHash()",
				zap.String("Error", err.Error()),
			)
		}
	}

	// Perevent calling external delegate
	if ctrl.isDeliveredMessage(x.Message.ID) {
		passToExternalhandler = false
	} else {
		passToExternalhandler = true
	}
	return
}

// updateReadHistoryInbox
func (ctrl *SyncController) updateReadHistoryInbox(u *msg.UpdateEnvelope) (passToExternalhandler bool) {
	x := new(msg.UpdateReadHistoryInbox)
	x.Unmarshal(u.Update)
	dialog := repo.Ctx().Dialogs.GetDialog(x.Peer.ID, x.Peer.Type)
	if dialog == nil {
		return
	}

	repo.Ctx().Dialogs.UpdateReadInboxMaxID(ctrl.UserID, x.Peer.ID, x.Peer.Type, x.MaxID)

	passToExternalhandler = true
	return
}

// updateReadHistoryOutbox
func (ctrl *SyncController) updateReadHistoryOutbox(u *msg.UpdateEnvelope) (passToExternalhandler bool) {
	log.LOG.Debug("NativeConnector::updateReadHistoryOutbox() start applier")
	x := new(msg.UpdateReadHistoryOutbox)
	x.Unmarshal(u.Update)
	repo.Ctx().Dialogs.UpdateReadOutboxMaxID(x.Peer.ID, x.Peer.Type, x.MaxID)

	passToExternalhandler = true
	log.LOG.Debug("NativeConnector::updateReadHistoryOutbox() end applier",
		zap.Bool("passToExternalhandler", passToExternalhandler),
	)
	return
}

// updateMessageEdited
func (ctrl *SyncController) updateMessageEdited(u *msg.UpdateEnvelope) (passToExternalhandler bool) {
	x := new(msg.UpdateMessageEdited)
	x.Unmarshal(u.Update)
	repo.Ctx().Messages.SaveMessage(x.Message)

	passToExternalhandler = true
	return
}

func (ctrl *SyncController) updateMessageID(u *msg.UpdateEnvelope) (passToExternalhandler bool) {

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
	passToExternalhandler = false
	return
}
