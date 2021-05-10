package syncCtrl

import (
	"bytes"
	"encoding/binary"
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	messageHole "git.ronaksoft.com/river/sdk/internal/message_hole"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"git.ronaksoft.com/river/sdk/internal/uiexec"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/pools"
	"go.uber.org/zap"
	"hash/crc32"
	"sort"
)

func (ctrl *Controller) authAuthorization(e *rony.MessageEnvelope) {
	x := new(msg.AuthAuthorization)
	if err := x.Unmarshal(e.Message); err != nil {
		logs.Error("SyncCtrl couldn't unmarshal AuthAuthorization", zap.Error(err))
		return
	}
	logs.Debug("SyncCtrl applies AuthAuthorization",
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

	go func() {
		ctrl.Sync()
	}()
}

func (ctrl *Controller) authSentCode(e *rony.MessageEnvelope) {
	x := new(msg.AuthSentCode)
	if err := x.Unmarshal(e.Message); err != nil {
		logs.Error("SyncCtrl couldn't unmarshal AuthSentCode", zap.Error(err))
		return
	}

	logs.Debug("SyncCtrl applies AuthSentCode")

	ctrl.connInfo.ChangePhone(x.Phone)
}

func (ctrl *Controller) contactsImported(e *rony.MessageEnvelope) {
	x := new(msg.ContactsImported)
	if err := x.Unmarshal(e.Message); err != nil {
		logs.Error("SyncCtrl couldn't unmarshal ContactsImported", zap.Error(err))
		return
	}

	logs.Debug("SyncCtrl applies contactsImported")

	_ = repo.Users.SaveContact(domain.GetTeamID(e), x.ContactUsers...)
	repo.Users.Save(x.Users...)
}

func (ctrl *Controller) contactsMany(e *rony.MessageEnvelope) {
	x := new(msg.ContactsMany)
	if err := x.Unmarshal(e.Message); err != nil {
		logs.Error("SyncCtrl couldn't unmarshal ContactsMany", zap.Error(err))
		return
	}
	logs.Debug("SyncCtrl applies contactsMany",
		zap.Int("Users", len(x.Users)),
		zap.Int("Contacts", len(x.Contacts)),
	)

	// If contacts are modified in server, then first clear all the contacts and rewrite the new ones
	if x.Modified == true {
		_ = repo.Users.DeleteAllContacts(domain.GetTeamID(e))
	}

	// Sort the contact users by their ids
	sort.Slice(x.ContactUsers, func(i, j int) bool { return x.ContactUsers[i].ID < x.ContactUsers[j].ID })

	_ = repo.Users.SaveContact(domain.GetTeamID(e), x.ContactUsers...)
	_ = repo.Users.Save(x.Users...)

	if len(x.ContactUsers) > 0 {
		buff := bytes.Buffer{}
		b := make([]byte, 8)
		for _, contactUser := range x.ContactUsers {
			binary.BigEndian.PutUint64(b, uint64(contactUser.ID))
			buff.Write(b)
		}
		crc32Hash := crc32.ChecksumIEEE(buff.Bytes())
		err := repo.System.SaveInt(domain.GetContactsGetHashKey(domain.GetTeamID(e)), uint64(crc32Hash))
		if err != nil {
			logs.Error("SyncCtrl couldn't save ContactsHash in to the db", zap.Error(err))
		}
		uiexec.ExecDataSynced(false, true, false)
	}
}

func (ctrl *Controller) messagesDialogs(e *rony.MessageEnvelope) {
	x := new(msg.MessagesDialogs)
	if err := x.Unmarshal(e.Message); err != nil {
		logs.Error("SyncCtrl couldn't unmarshal MessagesDialogs", zap.Error(err))
		return
	}
	logs.Debug("SyncCtrl applies MessagesDialogs",
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
			logs.Error("SyncCtrl got dialog with nil top message", zap.Int64("MessageID", dialog.TopMessageID))
			err := repo.Dialogs.Save(dialog)
			logs.WarnOnErr("SyncCtrl got error on save dialog", err)
		} else {
			err := repo.Dialogs.SaveNew(dialog, topMessage.CreatedOn)
			logs.WarnOnErr("SyncCtrl got error on save new dialog", err)
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

func (ctrl *Controller) usersMany(e *rony.MessageEnvelope) {
	x := new(msg.UsersMany)
	err := x.Unmarshal(e.Message)
	if err != nil {
		logs.Error("SyncCtrl couldn't unmarshal UsersMany", zap.Error(err))
		return
	}
	logs.Debug("SyncCtrl applies usersMany",
		zap.Int("Users", len(x.Users)),
	)
	_ = repo.Users.Save(x.Users...)
}

func (ctrl *Controller) messagesMany(e *rony.MessageEnvelope) {
	x := new(msg.MessagesMany)
	err := x.Unmarshal(e.Message)
	if err != nil {
		logs.Error("SyncCtrl couldn't unmarshal MessagesMany", zap.Error(err))
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

	logs.Info("SyncCtrl applies MessagesMany",
		zap.Bool("Continues", x.Continuous),
		zap.Int("Messages", len(x.Messages)),
	)
}

func (ctrl *Controller) groupFull(e *rony.MessageEnvelope) {
	u := new(msg.GroupFull)
	err := u.Unmarshal(e.Message)
	if err != nil {
		logs.Error("SyncCtrl couldn't unmarshal GroupFull", zap.Error(err))
		return
	}
	logs.Debug("SyncCtrl applies GroupFull",
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

	// Save Users, and notify settings
	_ = repo.Users.Save(u.Users...)
	_ = repo.Groups.SavePhotoGallery(u.Group.ID, u.PhotoGallery...)
	_ = repo.Dialogs.UpdateNotifySetting(u.Group.TeamID, u.Group.ID, int32(msg.PeerType_PeerGroup), u.NotifySettings)
}

func (ctrl *Controller) labelsMany(e *rony.MessageEnvelope) {
	u := &msg.LabelsMany{}
	err := u.Unmarshal(e.Message)
	if err != nil {
		logs.Error("SyncCtrl couldn't unmarshal LabelsMany", zap.Error(err))
		return
	}

	logs.Debug("SyncCtrl applies LabelsMany", zap.Any("TeamID", e.Get("TeamID", "0")))

	err = repo.Labels.Save(domain.GetTeamID(e), u.Labels...)
	logs.WarnOnErr("SyncCtrl got error on applying LabelsMany", err)

	return
}

func (ctrl *Controller) labelItems(e *rony.MessageEnvelope) {
	u := &msg.LabelItems{}
	err := u.Unmarshal(e.Message)
	if err != nil {
		logs.Error("SyncCtrl couldn't unmarshal LabelItems", zap.Error(err))
		return
	}

	logs.Debug("SyncCtrl applies LabelItems")

	_ = repo.Messages.Save(u.Messages...)
	_ = repo.Users.Save(u.Users...)
	_ = repo.Groups.Save(u.Groups...)
}

func (ctrl *Controller) systemConfig(e *rony.MessageEnvelope) {
	u := &msg.SystemConfig{}
	err := u.Unmarshal(e.Message)
	if err != nil {
		logs.Error("SyncCtrl couldn't unmarshal SystemConfig", zap.Error(err))
		return
	}

	logs.Debug("SyncCtrl applies SystemConfig")

	sysConfBytes, _ := u.Marshal()
	domain.SysConfig.Reactions = domain.SysConfig.Reactions[:0]
	err = domain.SysConfig.Unmarshal(sysConfBytes)
	if err != nil {
		logs.Error("SyncCtrl got error on unmarshalling SystemConfig", zap.Error(err))
	}
	err = repo.System.SaveBytes("SysConfig", sysConfBytes)
	if err != nil {
		logs.Error("SyncCtrl got error on saving SystemConfig", zap.Error(err))
	}
}

func (ctrl *Controller) contactsTopPeers(e *rony.MessageEnvelope) {
	u := &msg.ContactsTopPeers{}
	err := u.Unmarshal(e.Message)
	if err != nil {
		logs.Error("SyncCtrl couldn't unmarshal ContactsTopPeers", zap.Error(err))
		return
	}

	logs.Debug("SyncCtrl applies ContactsTopPeers",
		zap.Int("L", len(u.Peers)),
		zap.String("Cat", u.Category.String()),
	)
	err = repo.TopPeers.Save(u.Category, ctrl.userID, domain.GetTeamID(e), u.Peers...)
	if err != nil {
		logs.Error("SyncCtrl got error on saving ContactsTopPeers", zap.Error(err))
	}
}

func (ctrl *Controller) wallpapersMany(e *rony.MessageEnvelope) {
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

func (ctrl *Controller) savedGifs(e *rony.MessageEnvelope) {
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
		uiexec.ExecDataSynced(false, false, true)
	}
}

func (ctrl *Controller) botResults(e *rony.MessageEnvelope) {
	br := &msg.BotResults{}
	err := br.Unmarshal(e.Message)
	if err != nil {
		logs.Error("SyncCtrl couldn't unmarshal BotResults", zap.Error(err))
		return
	}

	for _, m := range br.Results {
		if m == nil || m.Message == nil || m.Type != msg.MediaType_MediaTypeDocument || m.Message.MediaData == nil {
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

func (ctrl *Controller) teamsMany(e *rony.MessageEnvelope) {
	tm := &msg.TeamsMany{}
	err := tm.Unmarshal(e.Message)
	if err != nil {
		logs.Error("SyncCtrl couldn't unmarshal TeamsMany", zap.Error(err))
		return
	}

	err = repo.Users.Save(tm.Users...)
	logs.ErrorOnErr("SyncCtrl couldn't save teamsMany users", err)

	err = repo.Teams.Clear()
	logs.ErrorOnErr("SyncCtrl couldn't clear saved teams", err)

	err = repo.Teams.Save(tm.Teams...)
	logs.ErrorOnErr("SyncCtrl couldn't save teamsMany teams", err)
}

func (ctrl *Controller) teamMembers(e *rony.MessageEnvelope) {
	tm := &msg.TeamMembers{}
	err := tm.Unmarshal(e.Message)
	if err != nil {
		logs.Error("SyncCtrl couldn't unmarshal TeamMembers", zap.Error(err))
		return
	}

	err = repo.Users.Save(tm.Users...)
	logs.ErrorOnErr("SyncCtrl couldn't save teamMembers users", err)
}

func (ctrl *Controller) reactionList(e *rony.MessageEnvelope) {
	tm := &msg.MessagesReactionList{}
	err := tm.Unmarshal(e.Message)
	if err != nil {
		logs.Error("SyncCtrl couldn't unmarshal MessagesReactionList", zap.Error(err))
		return
	}

}
