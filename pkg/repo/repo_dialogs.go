package repo

import (
	"errors"
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo/dto"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/dgraph-io/badger"
	"github.com/gobwas/pool/pbytes"
	"go.uber.org/zap"
)

const (
	prefixDialogs       = "DLG"
)

type repoDialogs struct {
	*repository
}

func (r *repoDialogs) getKey(peerID int64, peerType int32) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%021d.%d", prefixDialogs, peerID, peerType))
}

func (r *repoDialogs) Save(dialog *msg.Dialog, lastUpdate int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if dialog == nil {
		logs.Debug("Dialogs::SaveDialog()",
			zap.String("Error", "dialog is null"),
		)
		return domain.ErrNotFound
	}

	d := new(dto.Dialogs)
	d.Map(dialog)

	if lastUpdate > 0 {
		d.LastUpdate = lastUpdate
	}

	entity := new(dto.Dialogs)
	r.db.Where(&dto.Dialogs{PeerID: d.PeerID, PeerType: d.PeerType}).Find(entity)
	if entity.PeerID == 0 {
		return r.db.Create(d).Error
	}
	return r.db.Table(d.TableName()).Where("PeerID=? AND PeerType=?", d.PeerID, d.PeerType).Update(d).Error
}

func (r *repoDialogs) UpdateUnreadCount(peerID int64, peerTyep, unreadCount int32) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	ed := new(dto.Dialogs)
	return r.db.Table(ed.TableName()).Where("PeerID=? AND PeerType=?", peerID, peerTyep).Updates(map[string]interface{}{"UnreadCount": unreadCount}).Error
}

func (r *repoDialogs) List(offset, limit int32) []*msg.Dialog {
	r.mx.Lock()
	defer r.mx.Unlock()

	dtoDlgs := make([]dto.Dialogs, 0, limit)

	err := r.db.Limit(limit).Offset(offset).Order("LastUpdate DESC").Find(&dtoDlgs).Error
	if err != nil {
		logs.Warn("Dialogs::GetDialogs()", zap.Error(err))
		return nil
	}

	dialogs := make([]*msg.Dialog, 0, limit)
	for _, v := range dtoDlgs {
		tmp := new(msg.Dialog)
		v.MapTo(tmp)
		dialogs = append(dialogs, tmp)
	}

	return dialogs
}

func (r *repoDialogs) Get(peerID int64, peerType int32) *msg.Dialog {
	r.mx.Lock()
	defer r.mx.Unlock()

	dtoDlg := new(dto.Dialogs)
	err := r.db.Where("PeerID = ? AND PeerType = ?", peerID, peerType).First(dtoDlg).Error
	if err != nil {
		logs.Warn("Dialogs::GetDialog()->fetch dialog entity", zap.Error(err))
		return nil
	}

	dialog := new(msg.Dialog)
	dtoDlg.MapTo(dialog)

	return dialog
}

func (r *repoDialogs) Count() int32 {
	r.mx.Lock()
	defer r.mx.Unlock()

	// TODO:: implement it in ORM
	qry := `SELECT Count(DISTINCT dialogs.*) FROM dialogs
	LEFT OUTER JOIN messages ON dialogs.PeerID = messages.PeerID AND dialogs.PeerType = messages.PeerType
	WHERE messages.ID is not null`
	var count int32
	r.db.Raw(qry).Scan(&count)
	return count
}

func (r *repoDialogs) UpdateReadInboxMaxID(userID, peerID int64, peerType int32, maxID int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	// get dialog
	dtoDlg := new(dto.Dialogs)
	err := r.db.Where("PeerID = ? AND PeerType = ?", peerID, peerType).First(dtoDlg).Error
	if err != nil {
		logs.Error("Dialogs::UpdateReadInboxMaxID()-> fetch dialog entity", zap.Error(err))
		return nil
	}

	// current maxID is newer so skip updating dialog unread counts
	if dtoDlg.ReadInboxMaxID > maxID {
		return nil
	}

	// calculate unread count
	// currently when we enter dialog the max unreaded message ID will be sent
	var unreadCount int

	em := new(dto.Messages)
	err = r.db.Table(em.TableName()).Where("SenderID <> ? AND PeerID = ? AND PeerType = ? AND ID > ? ", userID, peerID, peerType, maxID).Count(&unreadCount).Error
	if err != nil {
		logs.Warn("Dialogs::UpdateReadInboxMaxID()-> fetch messages unread count", zap.Error(err))
		return err
	}

	ed := new(dto.Dialogs)
	err = r.db.Table(ed.TableName()).Where("PeerID = ? AND PeerType = ?", peerID, peerType).Updates(map[string]interface{}{
		"ReadInboxMaxID": maxID,
		"UnreadCount":    unreadCount, // gorm.Expr("UnreadCount + ?", unreadCount), // in snapshot mode if unread message lefted
		"MentionedCount": 0,           // Hotfix :: on each ReadHistoryInbox set mentioned count to zero
	}).Error

	if err != nil {
		logs.Warn("Dialogs::UpdateReadInboxMaxID()-> update dialog entity", zap.Error(err))
		return err
	}
	return nil
}

func (r *repoDialogs) UpdateReadOutboxMaxID(peerID int64, peerType int32, maxID int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	ed := new(dto.Dialogs)
	err := r.db.Table(ed.TableName()).Where("PeerID = ? AND PeerType = ?", peerID, peerType).Updates(map[string]interface{}{
		"ReadOutboxMaxID": maxID,
	}).Error

	if err != nil {
		logs.Warn("Dialogs::UpdateReadOutboxMaxID()-> update dialog entity", zap.Error(err))
		return err
	}
	return nil
}

func (r *repoDialogs) UpdateAccessHash(accessHash int64, peerID int64, peerType int32) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	ed := new(dto.Dialogs)
	err := r.db.Table(ed.TableName()).Where("PeerID=? AND PeerType=?", peerID, peerType).Updates(map[string]interface{}{
		"AccessHash": int64(accessHash),
	}).Error
	return err
}

func (r *repoDialogs) GetPinnedDialogs() []*msg.Dialog {
	r.mx.Lock()
	defer r.mx.Unlock()

	ed := new(dto.Dialogs)
	dtoDlgs := make([]dto.Dialogs, 0)
	err := r.db.Table(ed.TableName()).Where("IsPinned=?", 1).Find(&dtoDlgs).Error

	if err != nil {
		logs.Error("Dialogs::GetPinnedDialogs()->fetch dialogs entity", zap.Error(err))
		return nil
	}

	dialogs := make([]*msg.Dialog, 0)
	for _, v := range dtoDlgs {
		tmp := new(msg.Dialog)
		v.MapTo(tmp)
		dialogs = append(dialogs, tmp)
	}

	return dialogs
}

func (r *repoDialogs) UpdateNotifySetting(msg *msg.UpdateNotifySettings) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if msg.NotifyPeer == nil {
		return errors.New("Dialogs::UpdateNotifySetting() => msg.NotifyPeer is null")
	}
	if msg.Settings == nil {
		return errors.New("Dialogs::UpdateNotifySetting() => msg.Settings is null")
	}

	dtoDlg := new(dto.Dialogs)
	err := r.db.Where("PeerID = ? AND PeerType = ?", msg.NotifyPeer.ID, msg.NotifyPeer.Type).First(dtoDlg).Error
	if err != nil {
		logs.Warn("Dialogs::UpdateNotifySetting()->fetch dialog entity", zap.Error(err))
		return err
	}
	dtoDlg.NotifyFlags = msg.Settings.Flags
	dtoDlg.NotifyMuteUntil = msg.Settings.MuteUntil
	dtoDlg.NotifySound = msg.Settings.Sound

	return r.db.Save(dtoDlg).Error
}

func (r *repoDialogs) UpdatePinned(msg *msg.UpdateDialogPinned) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if msg.Peer == nil {
		return errors.New("Dialogs::UpdateDialogPinned() => msg.Peer is null")
	}

	dtoDlg := new(dto.Dialogs)
	err := r.db.Where("PeerID = ? AND PeerType = ?", msg.Peer.ID, msg.Peer.Type).First(dtoDlg).Error
	if err != nil {
		logs.Error("Dialogs::UpdateDialogPinned()->fetch dialog entity", zap.Error(err))
		return err
	}
	dtoDlg.Pinned = msg.Pinned

	return r.db.Save(dtoDlg).Error
}

func (r *repoDialogs) UpdatePeerNotifySettings(peerID int64, peerType int32, notifySetting *msg.PeerNotifySettings) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	ed := new(dto.Dialogs)
	err := r.db.Table(ed.TableName()).Where("PeerID=? AND PeerType=?", peerID, peerType).Updates(map[string]interface{}{
		"NotifyFlags":     notifySetting.Flags,
		"NotifyMuteUntil": notifySetting.MuteUntil,
		"NotifySound":     notifySetting.Sound,
	}).Error

	return err
}

func (r *repoDialogs) Delete(groupID int64, peerType int32) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	return r.db.Where("PeerID=? AND  PeerType=?", groupID, peerType).Delete(dto.Dialogs{}).Error
}

func (r *repoDialogs) GetMany(peerIDs []int64) []*msg.Dialog {
	r.mx.Lock()
	defer r.mx.Unlock()

	dtoDlgs := make([]dto.Dialogs, 0)
	err := r.db.Where("PeerID IN (?)", peerIDs).Find(&dtoDlgs).Error
	if err != nil {
		logs.Error("Dialogs::GetDialogMany()", zap.Error(err))
		return nil
	}

	dialogs := make([]*msg.Dialog, 0)
	for _, v := range dtoDlgs {
		tmp := new(msg.Dialog)
		v.MapTo(tmp)
		dialogs = append(dialogs, tmp)
	}

	return dialogs
}

func (r *repoDialogs) GetPeerIDs() []int64 {
	r.mx.Lock()
	defer r.mx.Unlock()

	dialogs := make([]dto.Dialogs, 0)
	err := r.db.Where("PeerType = 1").Find(&dialogs).Error
	if err != nil {
		logs.Error("Dialogs::GetPeerIDs()", zap.Error(err))
		return nil
	}
	var ids []int64
	for _, dialog := range dialogs {
		ids = append(ids, dialog.PeerID)
	}
	return ids
}
