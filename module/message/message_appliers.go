package message

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/logs"
	messageHole "git.ronaksoft.com/river/sdk/internal/message_hole"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/pools"
	"go.uber.org/zap"
)

/*
   Creation Time: 2021 - May - 10
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func (r *message) messagesDialogs(e *rony.MessageEnvelope) {
	x := new(msg.MessagesDialogs)
	if err := x.Unmarshal(e.Message); err != nil {
		logs.Error("MessageModule couldn't unmarshal MessagesDialogs", zap.Error(err))
		return
	}
	logs.Debug("MessageModule applies MessagesDialogs",
		zap.Int("Dialogs", len(x.Dialogs)),
		zap.Int64("GetUpdateID", x.UpdateID),
		zap.Int32("Count", x.Count),
	)

	mMessages := make(map[int64]*msg.UserMessage)
	for _, message := range x.Messages {
		mMessages[message.ID] = message
	}
	for _, dialog := range x.Dialogs {
		topMessage, _ := mMessages[dialog.TopMessageID]
		if topMessage == nil {
			logs.Error("MessageModule got dialog with nil top message", zap.Int64("MessageID", dialog.TopMessageID))
			err := repo.Dialogs.Save(dialog)
			logs.WarnOnErr("MessageModule got error on save dialog", err)
		} else {
			err := repo.Dialogs.SaveNew(dialog, topMessage.CreatedOn)
			logs.WarnOnErr("MessageModule got error on save new dialog", err)
			messageHole.InsertFill(dialog.TeamID, dialog.PeerID, dialog.PeerType, 0, dialog.TopMessageID, dialog.TopMessageID)
		}
	}
	// save Groups & Users & Messages
	waitGroup := pools.AcquireWaitGroup()
	waitGroup.Add(3)
	go func() {
		_ = repo.Users.Save(x.Users...)
		waitGroup.Done()
	}()
	go func() {
		_ = repo.Groups.Save(x.Groups...)
		waitGroup.Done()
	}()
	go func() {
		_ = repo.Messages.Save(x.Messages...)
		waitGroup.Done()
	}()
	waitGroup.Wait()
	pools.ReleaseWaitGroup(waitGroup)
}

func (r *message) messagesMany(e *rony.MessageEnvelope) {
	x := new(msg.MessagesMany)
	err := x.Unmarshal(e.Message)
	if err != nil {
		logs.Error("MessageModule couldn't unmarshal MessagesMany", zap.Error(err))
		return
	}

	// save Groups & Users & Messages
	waitGroup := pools.AcquireWaitGroup()
	waitGroup.Add(3)
	go func() {
		_ = repo.Users.Save(x.Users...)
		waitGroup.Done()
	}()
	go func() {
		_ = repo.Groups.Save(x.Groups...)
		waitGroup.Done()
	}()
	go func() {
		_ = repo.Messages.Save(x.Messages...)
		waitGroup.Done()
	}()
	waitGroup.Wait()
	pools.ReleaseWaitGroup(waitGroup)

	logs.Info("MessageModule applies MessagesMany",
		zap.Bool("Continues", x.Continuous),
		zap.Int("Messages", len(x.Messages)),
	)
}

func (r *message) reactionList(e *rony.MessageEnvelope) {
	tm := &msg.MessagesReactionList{}
	err := tm.Unmarshal(e.Message)
	if err != nil {
		logs.Error("MessageModule couldn't unmarshal MessagesReactionList", zap.Error(err))
		return
	}

}
