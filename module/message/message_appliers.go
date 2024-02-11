package message

import (
    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/river-sdk/internal/hole"
    "github.com/ronaksoft/river-sdk/internal/repo"
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
        r.Log().Error("couldn't unmarshal MessagesDialogs", zap.Error(err))
        return
    }
    r.Log().Debug("applies MessagesDialogs",
        zap.Int("Dialogs", len(x.Dialogs)),
        zap.Int64("GetUpdateID", x.UpdateID),
        zap.Int32("Count", x.Count),
    )

    mMessages := make(map[int64]*msg.UserMessage)
    for _, message := range x.Messages {
        mMessages[message.ID] = message
    }

    waitGroup := pools.AcquireWaitGroup()
    for _, dialog := range x.Dialogs {
        waitGroup.Add(1)
        go func(dialog *msg.Dialog) {
            topMessage := mMessages[dialog.TopMessageID]
            if topMessage == nil {
                r.Log().Error("got dialog with nil top message", zap.Int64("MessageID", dialog.TopMessageID))
                err := repo.Dialogs.Save(dialog)
                r.Log().WarnOnErr("got error on save dialog", err)
            } else {
                err := repo.Dialogs.SaveNew(dialog, topMessage.CreatedOn)
                r.Log().WarnOnErr("got error on save new dialog", err)
                hole.InsertFill(dialog.TeamID, dialog.PeerID, dialog.PeerType, msg.MediaCategory_MediaCategoryNone, dialog.TopMessageID, dialog.TopMessageID)
                hole.InsertFill(dialog.TeamID, dialog.PeerID, dialog.PeerType, msg.MediaCategory_MediaCategoryAudio, dialog.TopMessageID, dialog.TopMessageID)
                hole.InsertFill(dialog.TeamID, dialog.PeerID, dialog.PeerType, msg.MediaCategory_MediaCategoryVoice, dialog.TopMessageID, dialog.TopMessageID)
                hole.InsertFill(dialog.TeamID, dialog.PeerID, dialog.PeerType, msg.MediaCategory_MediaCategoryMedia, dialog.TopMessageID, dialog.TopMessageID)
                hole.InsertFill(dialog.TeamID, dialog.PeerID, dialog.PeerType, msg.MediaCategory_MediaCategoryFile, dialog.TopMessageID, dialog.TopMessageID)
                hole.InsertFill(dialog.TeamID, dialog.PeerID, dialog.PeerType, msg.MediaCategory_MediaCategoryGif, dialog.TopMessageID, dialog.TopMessageID)
                hole.InsertFill(dialog.TeamID, dialog.PeerID, dialog.PeerType, msg.MediaCategory_MediaCategoryWeb, dialog.TopMessageID, dialog.TopMessageID)
                hole.InsertFill(dialog.TeamID, dialog.PeerID, dialog.PeerType, msg.MediaCategory_MediaCategoryContact, dialog.TopMessageID, dialog.TopMessageID)
                hole.InsertFill(dialog.TeamID, dialog.PeerID, dialog.PeerType, msg.MediaCategory_MediaCategoryLocation, dialog.TopMessageID, dialog.TopMessageID)
            }
            waitGroup.Done()
        }(dialog)

    }
    // save Groups & Users & Messages

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
        r.Log().Error("couldn't unmarshal MessagesMany", zap.Error(err))
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

    r.Log().Info("applies MessagesMany",
        zap.Bool("Continues", x.Continuous),
        zap.Int("Messages", len(x.Messages)),
    )
}

func (r *message) reactionList(e *rony.MessageEnvelope) {
    tm := &msg.MessagesReactionList{}
    err := tm.Unmarshal(e.Message)
    if err != nil {
        r.Log().Error("couldn't unmarshal MessagesReactionList", zap.Error(err))
        return
    }

}
