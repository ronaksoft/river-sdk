package synchronizer

import (
	"time"

	"git.ronaksoftware.com/ronak/riversdk/cmd"
	"git.ronaksoftware.com/ronak/riversdk/configs"
	"git.ronaksoftware.com/ronak/riversdk/delegates"
	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/log"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/repo"
	"go.uber.org/zap"
)

// authAuthorization
func (ctrl *SyncController) authAuthorization(e *msg.MessageEnvelope) {
	x := new(msg.AuthAuthorization)
	if err := x.Unmarshal(e.Message); err != nil {
		log.LOG.Debug(err.Error())
		return
	}

	conf := configs.Get()

	conf.FirstName = x.User.FirstName
	conf.LastName = x.User.LastName
	conf.UserID = x.User.ID
	conf.Save()

	ctrl.SetUserID(x.User.ID)

	go ctrl.sync()
}

func (ctrl *SyncController) authSentCode(e *msg.MessageEnvelope) {
	x := new(msg.AuthSentCode)
	if err := x.Unmarshal(e.Message); err != nil {
		log.LOG.Debug(err.Error())
		return
	}
	conf := configs.Get()
	conf.Phone = x.Phone
	// no need to save it here its gonna be saved on authAuthorization
	// conf.Save()
}

// contactsImported
func (ctrl *SyncController) contactsImported(e *msg.MessageEnvelope) {
	x := new(msg.ContactsImported)
	if err := x.Unmarshal(e.Message); err != nil {
		log.LOG.Debug(err.Error())
		return
	}
	for _, u := range x.Users {
		repo.Ctx().Users.SaveContactUser(u)
	}
}

// contactsMany
func (ctrl *SyncController) contactsMany(e *msg.MessageEnvelope) {
	x := new(msg.ContactsMany)
	if err := x.Unmarshal(e.Message); err != nil {
		log.LOG.Debug(err.Error())
		return
	}
	for _, u := range x.Users {
		repo.Ctx().Users.SaveContactUser(u)
	}
}

// messageDialogs
func (ctrl *SyncController) messagesDialogs(e *msg.MessageEnvelope) {
	x := new(msg.MessagesDialogs)
	if err := x.Unmarshal(e.Message); err != nil {
		log.LOG.Debug(err.Error())
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
			log.LOG.Debug(domain.ErrNotFound.Error(),
				zap.Int64("MessageID", dialog.TopMessageID),
			)
			continue
		}
		repo.Ctx().Dialogs.SaveDialog(dialog, topMessage.CreatedOn)
	}
	for _, user := range x.Users {
		repo.Ctx().Users.SaveUser(user)
	}
}

// Check pending messages and notify UI
func (ctrl *SyncController) messageSent(e *msg.MessageEnvelope) {
	pmsg, err := repo.Ctx().PendingMessages.GetPendingMessageByRequestID(int64(e.RequestID))
	if err != nil {
		log.LOG.Debug(err.Error())
		return
	}

	sent := new(msg.MessagesSent)
	sent.Unmarshal(e.Message)

	message := new(msg.UserMessage)
	message.ID = sent.MessageID
	message.CreatedOn = sent.CreatedOn //pmsg.CreatedOn

	//message.RequestID = pmsg.RequestID
	message.PeerID = pmsg.PeerID
	message.SenderID = pmsg.SenderID
	message.PeerType = pmsg.PeerType
	//message.AccessHash = pmsg.AccessHash
	message.Body = pmsg.Body
	message.ReplyTo = pmsg.ReplyTo

	// save message
	err = repo.Ctx().Messages.SaveMessage(message)
	if err != nil {
		log.LOG.Error("faile to move PendingMessage to Messages " + err.Error())
		return
	}

	// delete pending mesage
	err = repo.Ctx().PendingMessages.DeletePendingMessage(pmsg.ID)
	if err != nil {
		log.LOG.Error("faile to delete PendingMessage " + err.Error())
		return
	}

	//Update doaligs
	err = repo.Ctx().Dialogs.UpdateTopMesssageID(message.CreatedOn, message.PeerID, message.PeerType)
	if err != nil {
		log.LOG.Error("faile to update doalogs " + err.Error())
		return
	}

	// TODO : Notify UI that the pending message delivered to server
	e.Constructor = msg.C_ClientPendingMessageDelivery
	pbcpm := new(msg.ClientPendingMessage)
	pmsg.MapTo(pbcpm)

	out := msg.ClientPendingMessageDelivery{
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
	cmd.GetUIExecuter().Exec(func() { delegates.Get().OnUpdates(msg.C_UpdateEnvelope, buff) })
}

func (ctrl *SyncController) usersMany(e *msg.MessageEnvelope) {
	u := new(msg.UsersMany)
	err := u.Unmarshal(e.Message)
	if err != nil {
		log.LOG.Debug(err.Error())
		return
	}
	for _, v := range u.Users {
		repo.Ctx().Users.SaveUser(v)
	}
}

func (ctrl *SyncController) messagesMany(e *msg.MessageEnvelope) {
	u := new(msg.MessagesMany)
	err := u.Unmarshal(e.Message)
	if err != nil {
		log.LOG.Debug(err.Error())
		return
	}
	for _, v := range u.Users {
		repo.Ctx().Users.SaveUser(v)
	}

	for _, v := range u.Messages {
		repo.Ctx().Messages.SaveMessage(v)
	}
}
