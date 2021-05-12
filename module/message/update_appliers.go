package message

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	messageHole "git.ronaksoft.com/river/sdk/internal/message_hole"
	mon "git.ronaksoft.com/river/sdk/internal/monitoring"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"git.ronaksoft.com/river/sdk/internal/uiexec"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/pools"
	"github.com/ronaksoft/rony/tools"
	"go.uber.org/zap"
	"os"
	"sync"
	"time"
)

/*
   Creation Time: 2021 - May - 10
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func (r *message) updateNewMessage(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := &msg.UpdateNewMessage{}
	err := x.Unmarshal(u.Update)
	if err != nil {
		logs.Error("SyncCtrl couldn't unmarshal UpdateNewMessage", zap.Error(err))
		return nil, err
	}

	logs.Debug("SyncCtrl applies UpdateNewMessage",
		zap.Int64("MessageID", x.Message.ID),
		zap.Int64("UpdateID", x.UpdateID),
	)

	dialog, _ := repo.Dialogs.Get(x.Message.TeamID, x.Message.PeerID, x.Message.PeerType)
	if dialog == nil {
		unreadCount := int32(0)
		if x.Sender.ID != r.SDK().SyncCtrl().GetUserID() {
			unreadCount = 1
		}
		// make sure to created the message hole b4 creating dialog
		dialog = &msg.Dialog{
			TeamID:         x.Message.TeamID,
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

	waitGroup := pools.AcquireWaitGroup()
	waitGroup.Add(1)
	go func() {
		// used messageType to identify client & server messages on Media thingy
		x.Message.MessageType = 1
		repo.MessagesExtra.SaveScrollID(x.Message.TeamID, x.Message.PeerID, x.Message.PeerType, 0, 0)
		waitGroup.Done()
	}()

	waitGroup.Add(1)
	go func() {
		// save user if does not exist
		err = repo.Users.Save(x.Sender)
		logs.WarnOnErr("SyncCtrl got error on saving user while applying new message", err, zap.Int64("SenderID", x.Sender.ID))
		waitGroup.Done()
	}()

	waitGroup.Add(1)
	go func() {
		err = repo.Messages.SaveNew(x.Message, r.SDK().SyncCtrl().GetUserID())
		logs.WarnOnErr("SyncCtrl got error on saving new message while applying new message", err, zap.Int64("SenderID", x.Sender.ID))
		messageHole.InsertFill(dialog.TeamID, dialog.PeerID, dialog.PeerType, 0, dialog.TopMessageID, x.Message.ID)
		waitGroup.Done()
	}()

	waitGroup.Wait()
	pools.ReleaseWaitGroup(waitGroup)

	// handle Message's Action
	res := []*msg.UpdateEnvelope{u}
	r.handleMessageAction(x, u, res)

	// If sender is me, check for pending
	if x.Message.SenderID == r.SDK().SyncCtrl().GetUserID() {
		pm := repo.PendingMessages.GetByRealID(x.Message.ID)
		if pm != nil {
			r.handlePendingMessage(x)
			_ = repo.PendingMessages.Delete(pm.ID)
			repo.PendingMessages.DeleteByRealID(x.Message.ID)
		}

		if x.Message.PeerID != x.Message.SenderID {
			if x.Message.FwdSenderID != 0 {
				_ = repo.TopPeers.Update(msg.TopPeerCategory_Forwards, r.SDK().SyncCtrl().GetUserID(), x.Message.TeamID, x.Message.PeerID, x.Message.PeerType)
			} else {
				switch msg.PeerType(x.Message.PeerType) {
				case msg.PeerType_PeerUser:
					p, _ := repo.Users.Get(x.Message.PeerID)
					if p == nil || !p.IsBot {
						_ = repo.TopPeers.Update(msg.TopPeerCategory_Users, r.SDK().SyncCtrl().GetUserID(), x.Message.TeamID, x.Message.PeerID, x.Message.PeerType)
					} else {
						_ = repo.TopPeers.Update(msg.TopPeerCategory_BotsMessage, r.SDK().SyncCtrl().GetUserID(), x.Message.TeamID, x.Message.PeerID, x.Message.PeerType)
					}
				case msg.PeerType_PeerGroup:
					_ = repo.TopPeers.Update(msg.TopPeerCategory_Groups, r.SDK().SyncCtrl().GetUserID(), x.Message.TeamID, x.Message.PeerID, x.Message.PeerType)
				}
			}
		}

		// We check if the file is GIF then check if it is a saved gif, if true, then update last access
		switch x.Message.MediaType {
		case msg.MediaType_MediaTypeDocument:
			xx := &msg.MediaDocument{}
			_ = xx.Unmarshal(x.Message.Media)
			for _, attr := range xx.Doc.Attributes {
				if attr.Type == msg.DocumentAttributeType_AttributeTypeAnimated {
					if repo.Gifs.IsSaved(xx.Doc.ClusterID, xx.Doc.ID) {
						_ = repo.Gifs.UpdateLastAccess(xx.Doc.ClusterID, xx.Doc.ID, x.Message.CreatedOn)

						uiexec.ExecDataSynced(false, false, true)
					}
				}
			}
			mon.IncMediaSent()
		case msg.MediaType_MediaTypeEmpty:
			mon.IncMessageSent()
		default:
			mon.IncMediaSent()
		}
	} else {
		switch x.Message.MediaType {
		case msg.MediaType_MediaTypeEmpty:
			mon.IncMessageReceived()
		default:
			mon.IncMediaReceived()
		}
	}

	return res, nil
}
func (r *message) handleMessageAction(x *msg.UpdateNewMessage, u *msg.UpdateEnvelope, res []*msg.UpdateEnvelope) {
	switch x.Message.MessageAction {
	case domain.MessageActionContactRegistered:
		go func() {
			waitGroup := &sync.WaitGroup{}
			waitGroup.Add(1)
			r.SDK().SyncCtrl().GetContacts(waitGroup, 0, 0)
			waitGroup.Wait()
		}()
	case domain.MessageActionGroupDeleteUser:
		act := new(msg.MessageActionGroupDeleteUser)
		err := act.Unmarshal(x.Message.MessageActionData)
		if err != nil {
			logs.Error("SyncCtrl couldn't unmarshal MessageActionGroupDeleteUser", zap.Error(err))
		}

		// Check if user left (deleted him/her self from group) remove its GroupSearch, Dialog and its MessagesPending
		selfUserID := r.SDK().GetConnInfo().PickupUserID()
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
			logs.Error("SyncCtrl couldn't delete group from top peers", zap.Error(err))
		}

		// Delete PendingMessage
		deletedMsgs := repo.PendingMessages.DeletePeerAllMessages(x.Message.PeerID, x.Message.PeerType)
		if deletedMsgs != nil {
			buff, err := deletedMsgs.Marshal()
			if err != nil {
				logs.Error("SyncCtrl couldn't marshal ClientUpdateMessagesDeleted", zap.Error(err))
				break
			}

			udp := &msg.UpdateEnvelope{
				Constructor: msg.C_ClientUpdateMessagesDeleted,
				Update:      buff,
				UCount:      1,
				Timestamp:   u.Timestamp,
				UpdateID:    u.UpdateID,
			}

			res = append(res, udp)

		}
	case domain.MessageActionClearHistory:
		act := new(msg.MessageActionClearHistory)
		_ = act.Unmarshal(x.Message.MessageActionData)

		// 1. Delete All Messages < x.MessageID
		_ = repo.Messages.ClearHistory(r.SDK().SyncCtrl().GetUserID(), x.Message.TeamID, x.Message.PeerID, x.Message.PeerType, act.MaxID)

		// 2. Delete Scroll Position
		repo.MessagesExtra.SaveScrollID(x.Message.TeamID, x.Message.PeerID, x.Message.PeerType, 0, 0)

		if act.Delete {
			// 3. Delete Dialog
			_ = repo.Dialogs.Delete(x.Message.TeamID, x.Message.PeerID, x.Message.PeerType)
		} else {
			// 3. Get dialog and create first hole
			dialog, _ := repo.Dialogs.Get(x.Message.TeamID, x.Message.PeerID, x.Message.PeerType)
			if dialog != nil {
				messageHole.InsertFill(dialog.TeamID, dialog.PeerID, dialog.PeerType, 0, dialog.TopMessageID, dialog.TopMessageID)
			}
		}
	}
}
func (r *message) handlePendingMessage(x *msg.UpdateNewMessage) {
	pmsg := repo.PendingMessages.GetByRealID(x.Message.ID)
	if pmsg == nil {
		return
	}

	if pmsg.MediaType != msg.InputMediaType_InputMediaTypeMessageDocument {
		// if it was file upload request
		switch x.Message.MediaType {
		case msg.MediaType_MediaTypeDocument:
			// save to local files and delete file status
			clientSendMedia := new(msg.ClientSendMessageMedia)
			unmarshalErr := clientSendMedia.Unmarshal(pmsg.Media)
			if unmarshalErr == nil { // TODO!!! fix this with some flag in pending message
				clientFile, err := repo.Files.GetMediaDocument(x.Message)
				logs.WarnOnErr("Error On GetMediaDocument", err)

				err = os.Rename(clientSendMedia.FilePath, repo.Files.GetFilePath(clientFile))
				if err != nil {
					logs.Error("Error On HandlePendingMessage (Move) we try Copy", zap.Error(err))
					err = domain.CopyFile(clientSendMedia.FilePath, repo.Files.GetFilePath(clientFile))
					if err != nil {
						logs.Error("Error On HandlePendingMessage (Copy) we give up!", zap.Error(err))
					}
					return
				}
			} else {
				logs.Error("Error On HandlePendingMessage", zap.Error(unmarshalErr))
			}
		}
	}

	clientUpdate := &msg.ClientUpdatePendingMessageDelivery{
		Messages:       x.Message,
		PendingMessage: pmsg,
		Success:        true,
	}

	bytes, _ := clientUpdate.Marshal()
	udpMsg := &msg.UpdateEnvelope{
		Constructor: msg.C_ClientUpdatePendingMessageDelivery,
		Update:      bytes,
		UpdateID:    0,
		Timestamp:   tools.TimeUnix(),
	}

	uiexec.ExecUpdate(msg.C_UpdateEnvelope, udpMsg)
}

func (r *message) updateReadHistoryInbox(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateReadHistoryInbox)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Debug("SyncCtrl applies UpdateReadHistoryInbox",
		zap.Int64("MaxID", x.MaxID),
		zap.Int64("UpdateID", x.UpdateID),
		zap.Int64("PeerID", x.Peer.ID),
	)

	_ = repo.Dialogs.UpdateReadInboxMaxID(r.SDK().SyncCtrl().GetUserID(), x.TeamID, x.Peer.ID, x.Peer.Type, x.MaxID)
	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

func (r *message) updateReadHistoryOutbox(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateReadHistoryOutbox)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Debug("SyncCtrl applies UpdateReadHistoryOutbox",
		zap.Int64("MaxID", x.MaxID),
		zap.Int64("UpdateID", x.UpdateID),
		zap.Int64("PeerID", x.Peer.ID),
	)

	_ = repo.Dialogs.UpdateReadOutboxMaxID(x.TeamID, x.Peer.ID, x.Peer.Type, x.MaxID)
	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

func (r *message) updateMessageEdited(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateMessageEdited)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Debug("SyncCtrl applies UpdateMessageEdited",
		zap.Int64("MessageID", x.Message.ID),
		zap.Int64("UpdateID", x.UpdateID),
	)

	repo.Messages.Save(x.Message)

	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

func (r *message) updateMessageID(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	res := make([]*msg.UpdateEnvelope, 0)
	x := new(msg.UpdateMessageID)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Debug("SyncCtrl applied UpdateMessageID",
		zap.Int64("RandomID", x.RandomID),
		zap.Int64("MessageID", x.MessageID),
	)

	msgEnvelop := new(rony.MessageEnvelope)
	msgEnvelop.Constructor = rony.C_MessageEnvelope

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
			zap.Int64("RandomID", x.RandomID),
			zap.Int64("MID", x.MessageID),
			zap.String("Body", userMessage.Body),
		)

		if pm == nil {
			logs.Warn("But we couldn't find Pending Message",
				zap.Int64("RandomID", sent.RandomID),
				zap.Int64("RealID", sent.MessageID),
			)
			return res, nil
		}
		r.deletePendingMessage(pm)
		return res, nil
	}

	if pm == nil {
		return nil, nil
	}
	logs.Info("SyncCtrl received UpdateMessageID before UpdateNewMessage",
		zap.Int64("RandomID", x.RandomID),
		zap.Int64("MID", x.MessageID),
		zap.Int64("PendingID", pm.ID),
	)

	repo.PendingMessages.SaveByRealID(sent.RandomID, sent.MessageID)
	return res, nil
}
func (r *message) deletePendingMessage(pm *msg.ClientPendingMessage) {
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

	uiexec.ExecUpdate(msg.C_UpdateEnvelope, updateEnvelope)

	_ = repo.PendingMessages.Delete(pm.ID)
}

func (r *message) updateDraftMessage(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateDraftMessage)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Debug("SyncCtrl applies UpdateDraftMessage",
		zap.Int64("UpdateID", x.UpdateID),
	)

	dialog, _ := repo.Dialogs.Get(x.Message.TeamID, x.Message.PeerID, x.Message.PeerType)
	if dialog != nil {
		dialog.Draft = x.Message
		repo.Dialogs.Save(dialog)
	}

	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

func (r *message) updateDraftMessageCleared(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateDraftMessageCleared)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Debug("SyncCtrl applies UpdateDraftMessageCleared",
		zap.Int64("UpdateID", x.UpdateID),
	)

	dialog, _ := repo.Dialogs.Get(x.TeamID, x.Peer.ID, x.Peer.Type)
	if dialog != nil {
		dialog.Draft = nil
		_ = repo.Dialogs.Save(dialog)
	}

	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

func (r *message) updateReaction(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := &msg.UpdateReaction{}
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Debug("SyncCtrl applies UpdateReaction",
		zap.Int64("UpdateID", x.UpdateID),
	)

	err = repo.Messages.UpdateReactionCounter(x.MessageID, x.Counter, x.YourReactions)
	if err != nil {
		return nil, err
	}

	return []*msg.UpdateEnvelope{u}, nil
}

func (r *message) updateMessagePinned(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := &msg.UpdateMessagePinned{}
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Debug("SyncCtrl applies UpdateMessagePinned",
		zap.Int64("UpdateID", x.UpdateID),
	)

	err = repo.Dialogs.UpdatePinMessageID(x.TeamID, x.Peer.ID, x.Peer.Type, x.MsgID)
	if err != nil {
		return nil, err
	}

	return []*msg.UpdateEnvelope{u}, nil
}

func (r *message) updateReadMessagesContents(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := &msg.UpdateReadMessagesContents{}
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Debug("SyncCtrl applies UpdateReadMessagesContents",
		zap.Int64s("MessageIDs", x.MessageIDs),
		zap.Int64("UpdateID", x.UpdateID),
	)

	repo.Messages.SetContentRead(x.Peer.ID, x.Peer.Type, x.MessageIDs)

	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

func (r *message) updateMessagesDeleted(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateMessagesDeleted)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Debug("SyncCtrl applies UpdateMessagesDeleted",
		zap.Int64("UpdateID", x.UpdateID),
	)

	repo.Messages.Delete(r.SDK().SyncCtrl().GetUserID(), x.TeamID, x.Peer.ID, x.Peer.Type, x.MessageIDs...)

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

func (r *message) updateNotifySettings(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateNotifySettings)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Debug("SyncCtrl applies UpdateNotifySettings",
		zap.Int64("UpdateID", x.UpdateID),
	)

	repo.Dialogs.UpdateNotifySetting(x.TeamID, x.NotifyPeer.ID, x.NotifyPeer.Type, x.Settings)

	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

func (r *message) updateDialogPinned(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	x := new(msg.UpdateDialogPinned)
	err := x.Unmarshal(u.Update)
	if err != nil {
		return nil, err
	}

	logs.Debug("SyncCtrl applies UpdateDialogPinned",
		zap.Int64("UpdateID", x.UpdateID),
	)

	repo.Dialogs.UpdatePinned(x)
	res := []*msg.UpdateEnvelope{u}
	return res, nil
}
