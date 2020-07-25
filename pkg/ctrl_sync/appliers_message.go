package syncCtrl

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"git.ronaksoftware.com/river/msg/msg"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/internal/logs"
	messageHole "git.ronaksoftware.com/ronak/riversdk/pkg/message_hole"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"go.uber.org/zap"
	"hash/crc32"
	"sort"
	"time"
)

// authAuthorization
func (ctrl *Controller) authAuthorization(e *msg.MessageEnvelope) {
	x := new(msg.AuthAuthorization)
	if err := x.Unmarshal(e.Message); err != nil {
		logs.Error("SyncCtrl couldn't unmarshal AuthAuthorization", zap.Error(err))
		return
	}
	logs.Info("SyncCtrl applies AuthAuthorization",
		zap.String("FirstName", x.User.FirstName),
		zap.String("LastName", x.User.LastName),
		zap.Int64("UserID", x.User.ID),
		zap.String("Bio", x.User.Bio),
		zap.String("Username", x.User.Username),
	)
	ctrl.connInfo.ChangeFirstName(x.User.FirstName)
	ctrl.connInfo.ChangeLastName(x.User.LastName)
	ctrl.connInfo.ChangeUserID(x.User.ID)
	ctrl.connInfo.ChangeBio(x.User.Bio)
	ctrl.connInfo.ChangeUsername(x.User.Username)
	if x.User.Phone != "" {
		ctrl.connInfo.ChangePhone(x.User.Phone)
	}
	ctrl.connInfo.Save()

	ctrl.SetUserID(x.User.ID)

	repo.SetSelfUserID(x.User.ID)

	domain.WindowLog(fmt.Sprintf("Authorized: %s", time.Now().Sub(domain.StartTime)))
	go func() {
		ctrl.Sync()
		domain.WindowLog(fmt.Sprintf("Synced: %s", time.Now().Sub(domain.StartTime)))
	}()
}

// authSentCode
func (ctrl *Controller) authSentCode(e *msg.MessageEnvelope) {
	x := new(msg.AuthSentCode)
	if err := x.Unmarshal(e.Message); err != nil {
		logs.Error("SyncCtrl couldn't unmarshal AuthSentCode", zap.Error(err))
		return
	}
	logs.Info("SyncCtrl applies AuthSentCode")
	ctrl.connInfo.ChangePhone(x.Phone)
}

// contactsImported
func (ctrl *Controller) contactsImported(e *msg.MessageEnvelope) {
	x := new(msg.ContactsImported)
	if err := x.Unmarshal(e.Message); err != nil {
		logs.Error("SyncCtrl couldn't unmarshal ContactsImported", zap.Error(err))
		return
	}
	logs.Info("SyncCtrl applies contactsImported")
	_ = repo.Users.SaveContact(x.ContactUsers...)
	repo.Users.Save(x.Users...)
}

// contactsMany
func (ctrl *Controller) contactsMany(e *msg.MessageEnvelope) {
	x := new(msg.ContactsMany)
	if err := x.Unmarshal(e.Message); err != nil {
		logs.Error("SyncCtrl couldn't unmarshal ContactsMany", zap.Error(err))
		return
	}
	logs.Info("SyncCtrl applies contactsMany",
		zap.Int("Users", len(x.Users)),
		zap.Int("Contacts", len(x.Contacts)),
	)

	// If contacts are modified in server, then first clear all the contacts and rewrite the new ones
	if x.Modified == true {
		_ = repo.Users.DeleteAllContacts()
	}

	// Sort the contact users by their ids
	sort.Slice(x.ContactUsers, func(i, j int) bool { return x.ContactUsers[i].ID < x.ContactUsers[j].ID })

	repo.Users.SaveContact(x.ContactUsers...)
	repo.Users.Save(x.Users...)
	// server
	if len(x.ContactUsers) > 0 {
		buff := bytes.Buffer{}
		b := make([]byte, 8)
		for _, contactUser := range x.ContactUsers {
			binary.BigEndian.PutUint64(b, uint64(contactUser.ID))
			buff.Write(b)
		}
		crc32Hash := crc32.ChecksumIEEE(buff.Bytes())
		err := repo.System.SaveInt(domain.SkContactsGetHash, uint64(crc32Hash))
		if err != nil {
			logs.Error("SyncCtrl couldn't save ContactsHash in to the db", zap.Error(err))
		}
		forceUpdateUI(ctrl, false, true, false)
	}
}

// messageDialogs
func (ctrl *Controller) messagesDialogs(e *msg.MessageEnvelope) {
	x := new(msg.MessagesDialogs)
	if err := x.Unmarshal(e.Message); err != nil {
		logs.Error("SyncCtrl couldn't unmarshal MessagesDialogs", zap.Error(err))
		return
	}
	logs.Info("SyncCtrl applies MessagesDialogs",
		zap.Int("Dialogs", len(x.Dialogs)),
		zap.Int64("UpdateID", x.UpdateID),
		zap.Int32("Count", x.Count),
	)

	mMessages := make(map[int64]*msg.UserMessage)
	for _, message := range x.Messages {
		mMessages[message.ID] = message
	}
	for _, dialog := range x.Dialogs {
		topMessage, _ := mMessages[dialog.TopMessageID]
		if topMessage == nil {
			logs.Error("SyncCtrl got dialog with nil top message", zap.Int64("MessageID", dialog.TopMessageID))
			err := repo.Dialogs.Save(dialog)
			logs.WarnOnErr("SyncCtrl got error on save dialog", err)
		} else {
			err := repo.Dialogs.SaveNew(dialog, topMessage.CreatedOn)
			logs.WarnOnErr("SyncCtrl got error on save new dialog", err)
			messageHole.InsertFill(dialog.TeamID, dialog.PeerID, dialog.PeerType, dialog.TopMessageID, dialog.TopMessageID)
		}
	}
	repo.Users.Save(x.Users...)
	repo.Groups.Save(x.Groups...)
	repo.Messages.Save(x.Messages...)
}

// usersMany
func (ctrl *Controller) usersMany(e *msg.MessageEnvelope) {
	u := new(msg.UsersMany)
	err := u.Unmarshal(e.Message)
	if err != nil {
		logs.Error("SyncCtrl couldn't unmarshal UsersMany", zap.Error(err))
		return
	}
	logs.Info("SyncCtrl applies usersMany",
		zap.Int("Users", len(u.Users)),
	)
	repo.Users.Save(u.Users...)
}

// messagesMany
func (ctrl *Controller) messagesMany(e *msg.MessageEnvelope) {
	u := new(msg.MessagesMany)
	err := u.Unmarshal(e.Message)
	if err != nil {
		logs.Error("SyncCtrl couldn't unmarshal MessagesMany", zap.Error(err))
		return
	}

	// save Groups & Users & Messages
	repo.Users.Save(u.Users...)
	repo.Groups.Save(u.Groups...)
	repo.Messages.Save(u.Messages...)

	logs.Info("SyncCtrl applies MessagesMany",
		zap.Bool("Continues", u.Continuous),
		zap.Int("Messages", len(u.Messages)),
	)
}

// groupFull
func (ctrl *Controller) groupFull(e *msg.MessageEnvelope) {
	u := new(msg.GroupFull)
	err := u.Unmarshal(e.Message)
	if err != nil {
		logs.Error("SyncCtrl couldn't unmarshal GroupFull", zap.Error(err))
		return
	}
	logs.Info("SyncCtrl applies GroupFull",
		zap.Int64("GroupID", u.Group.ID),
	)

	// save GroupSearch
	err = repo.Groups.SaveFull(u)
	if err != nil {
		logs.Error("SyncCtrl couldn't save GroupFull", zap.Error(err))
	}
	err = repo.Groups.Save(u.Group)
	if err != nil {
		logs.Error("SyncCtrl couldn't save GroupFull's Group", zap.Error(err))
	}

	// save Users
	repo.Users.Save(u.Users...)

	repo.Groups.SavePhotoGallery(u.Group.ID, u.PhotoGallery...)

	// Update NotifySettings
	repo.Dialogs.UpdateNotifySetting(u.Group.TeamID, u.Group.ID, int32(msg.PeerGroup), u.NotifySettings)
}

// labelsMany
func (ctrl *Controller) labelsMany(e *msg.MessageEnvelope) {
	u := &msg.LabelsMany{}
	err := u.Unmarshal(e.Message)
	if err != nil {
		logs.Error("SyncCtrl couldn't unmarshal LabelsMany", zap.Error(err))
		return
	}
	logs.Info("SyncCtrl applies LabelsMany")

	err = repo.Labels.Save(u.Labels...)
	logs.WarnOnErr("SyncCtrl got error on applying LabelsMany", err)

	return
}

// labelItems
func (ctrl *Controller) labelItems(e *msg.MessageEnvelope) {
	u := &msg.LabelItems{}
	err := u.Unmarshal(e.Message)
	if err != nil {
		logs.Error("SyncCtrl couldn't unmarshal LabelItems", zap.Error(err))
		return
	}
	logs.Info("SyncCtrl applies LabelItems")

	repo.Messages.Save(u.Messages...)
	repo.Users.Save(u.Users...)
	repo.Groups.Save(u.Groups...)
}

// systemConfig
func (ctrl *Controller) systemConfig(e *msg.MessageEnvelope) {
	u := &msg.SystemConfig{}
	err := u.Unmarshal(e.Message)
	if err != nil {
		logs.Error("SyncCtrl couldn't unmarshal SystemConfig", zap.Error(err))
		return
	}
	logs.Info("SyncCtrl applies SystemConfig")

	sysConfBytes, _ := u.Marshal()
	err = repo.System.SaveBytes("SysConfig", sysConfBytes)
	if err != nil {
		logs.Error("SyncCtrl got error on saving SystemConfig", zap.Error(err))
	}
}

// contactsTopPeers
func (ctrl *Controller) contactsTopPeers(e *msg.MessageEnvelope) {
	u := &msg.ContactsTopPeers{}
	err := u.Unmarshal(e.Message)
	if err != nil {
		logs.Error("SyncCtrl couldn't unmarshal ContactsTopPeers", zap.Error(err))
		return
	}
	logs.Info("SyncCtrl applies ContactsTopPeers",
		zap.Int("L", len(u.Peers)),
		zap.String("Cat", u.Category.String()),
	)

	err = repo.TopPeers.Save(u.Category, ctrl.userID, u.Peers...)
	if err != nil {
		logs.Error("SyncCtrl got error on saving ContactsTopPeers", zap.Error(err))
	}
}

// wallpapersMany
func (ctrl *Controller) wallpapersMany(e *msg.MessageEnvelope) {
	u := &msg.WallPapersMany{}
	err := u.Unmarshal(e.Message)
	if err != nil {
		logs.Error("SyncCtrl couldn't unmarshal wallpapersMany", zap.Error(err))
		return
	}
	err = repo.Wallpapers.SaveWallpapers(u)
	if err != nil {
		logs.Error("SyncCtrl got error on saving wallpapersMany", zap.Error(err))
	}
}

// savedGifs
func (ctrl *Controller) savedGifs(e *msg.MessageEnvelope) {
	u := &msg.SavedGifs{}
	err := u.Unmarshal(e.Message)
	if err != nil {
		logs.Error("SyncCtrl couldn't unmarshal savedGifs", zap.Error(err))
		return
	}
	accessTime := domain.Now().Unix()
	for _, d := range u.Docs {
		err = repo.Files.SaveGif(d)
		if err != nil {
			logs.Warn("SyncCtrl got error on applying SavedGifs (Save File)", zap.Error(err))
		}
		if !repo.Gifs.IsSaved(d.Doc.ClusterID, d.Doc.ID) {
			err = repo.Gifs.Save(d)
			if err != nil {
				logs.Warn("SyncCtrl got error on applying SavedGifs (Save Gif)", zap.Error(err))
			}
			err = repo.Gifs.UpdateLastAccess(d.Doc.ClusterID, d.Doc.ID, accessTime)
			if err != nil {
				logs.Warn("SyncCtrl got error on applying SavedGifs (Update Access Time)", zap.Error(err))
			}
		}
	}
	oldHash, _ := repo.System.LoadInt(domain.SkGifHash)
	err = repo.System.SaveInt(domain.SkGifHash, uint64(u.Hash))
	if err != nil {
		logs.Warn("SyncCtrl got error on saving GifHash", zap.Error(err))
	}
	if oldHash != uint64(u.Hash) {
		forceUpdateUI(ctrl, false, false, true)
	}
}

func (ctrl *Controller) botResults(e *msg.MessageEnvelope) {
	br := &msg.BotResults{}
	err := br.Unmarshal(e.Message)
	if err != nil {
		logs.Error("SyncCtrl couldn't unmarshal BotResults", zap.Error(err))
		return
	}

	for _, m := range br.Results {

		if m == nil || m.Message == nil || m.Type != msg.MediaTypeDocument || m.Message.MediaData == nil {
			logs.Info("SyncCtrl botResults message or media is nil or not mediaDocument", zap.Error(err))
			continue
		}

		md := &msg.MediaDocument{}
		err := md.Unmarshal(m.Message.MediaData)
		if err != nil {
			logs.Error("SyncCtrl couldn't unmarshal BotResults MediaDocument", zap.Error(err))
			continue
		}

		err = repo.Files.SaveMessageMediaDocument(md)

		if err != nil {
			logs.Error("SyncCtrl couldn't save botResults media document", zap.Error(err))
		}
	}
}
