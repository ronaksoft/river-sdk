package repo

import (
	"errors"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/logs"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/repo/dto"
	"go.uber.org/zap"
)

// Dialogs repoDialogs interface
type Dialogs interface {
	SaveDialog(dialog *msg.Dialog, lastUpdate int64) error
	GetDialogs(offset, limit int32) []*msg.Dialog
	GetDialog(peerID int64, peerType int32) *msg.Dialog
	CountDialogs() int32
	UpdateReadInboxMaxID(userID, peerID int64, peerType int32, maxID int64) error
	UpdateReadOutboxMaxID(peerID int64, peerType int32, maxID int64) error
	UpdateDialogUnreadCount(peerID int64, peerTyep, unreadCount int32) error
	UpdateAccessHash(accessHash int64, peerID int64, peerType int32) error
	UpdateTopMesssageID(createdOn, peerID int64, peerType int32) error
	UpdateNotifySetting(msg *msg.UpdateNotifySettings) error
	UpdatePeerNotifySettings(peerID int64, peerType int32, notifySetting *msg.PeerNotifySettings) error
	Delete(groupID int64, peerType int32) error
	GetManyDialog(peerIDs []int64) []*msg.Dialog
}

type repoDialogs struct {
	*repository
}

// UpdateTopMesssageID
func (r *repoDialogs) UpdateTopMesssageID(createdOn, peerID int64, peerType int32) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	em := dto.Messages{}
	err := r.db.Table(em.TableName()).Where("PeerID=? AND PeerType=?", peerID, peerType).Limit(1).Order("ID DESC").Find(&em).Error
	if err != nil {
		logs.Error("Dialogs::UpdateTopMesssageID() TopMessage", zap.Error(err))
		return err
	}

	topMessageID := em.ID
	ed := new(dto.Dialogs)
	err = r.db.Table(ed.TableName()).Where("PeerID=? AND PeerType=?", peerID, peerType).Updates(map[string]interface{}{
		"TopMessageID": topMessageID,
		"LastUpdate":   createdOn,
	}).Error

	return err
}

// SaveDialog
func (r *repoDialogs) SaveDialog(dialog *msg.Dialog, lastUpdate int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if dialog == nil {
		logs.Debug("Dialogs::SaveDialog()",
			zap.String("Error", "dialog is null"),
		)
		return domain.ErrNotFound
	}

	logs.Debug("Dialogs::SaveDialog",
		zap.Int64("PeerID", dialog.PeerID),
		zap.Uint64("AccessHash", dialog.AccessHash),
		zap.Int32("UnreadCount", dialog.UnreadCount),
	)
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

// UpdateDialogUnreadCount
func (r *repoDialogs) UpdateDialogUnreadCount(peerID int64, peerTyep, unreadCount int32) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	logs.Debug("Dialogs::UpdateDialogUnreadCount()",
		zap.Int64("PeerID", peerID),
		zap.Int32("PeerType", peerTyep),
		zap.Int32("UnreadCount", unreadCount),
	)
	ed := new(dto.Dialogs)
	return r.db.Table(ed.TableName()).Where("PeerID=? AND PeerType=?", peerID, peerTyep).Updates(map[string]interface{}{"UnreadCount": unreadCount}).Error
}

// GetDialogs
func (r *repoDialogs) GetDialogs(offset, limit int32) []*msg.Dialog {
	r.mx.Lock()
	defer r.mx.Unlock()

	logs.Debug("Dialogs::GetDialogs()",
		zap.Int32("Offset", offset),
		zap.Int32("Limit", limit),
	)

	dtoDlgs := make([]dto.Dialogs, 0, limit)

	err := r.db.Limit(limit).Offset(offset).Order("LastUpdate DESC").Find(&dtoDlgs).Error
	if err != nil {
		logs.Error("Dialogs::GetDialogs()", zap.Error(err))
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

// GetDialog
func (r *repoDialogs) GetDialog(peerID int64, peerType int32) *msg.Dialog {
	r.mx.Lock()
	defer r.mx.Unlock()

	logs.Debug("Dialogs::GetDialog()",
		zap.Int64("PeerID", peerID),
		zap.Int32("PeerType", peerType),
	)

	dtoDlg := new(dto.Dialogs)
	err := r.db.Where("PeerID = ? AND PeerType = ?", peerID, peerType).First(dtoDlg).Error
	if err != nil {
		logs.Error("Dialogs::GetDialog()->fetch dialog entity", zap.Error(err))
		return nil
	}

	dialog := new(msg.Dialog)
	dtoDlg.MapTo(dialog)

	return dialog
}

// CountDialogs
func (r *repoDialogs) CountDialogs() int32 {
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

// UpdateReadInboxMaxID
func (r *repoDialogs) UpdateReadInboxMaxID(userID, peerID int64, peerType int32, maxID int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	logs.Debug("Dialogs::UpdateReadInboxMaxID()",
		zap.Int64("PeerID", peerID),
		zap.Int32("PeerType", peerType),
		zap.Int64("MaxID", maxID),
	)

	// get dialog
	dtoDlg := new(dto.Dialogs)
	err := r.db.Where("PeerID = ? AND PeerType = ?", peerID, peerType).First(dtoDlg).Error
	if err != nil {
		logs.Error("Dialogs::UpdateReadInboxMaxID()-> fetch dialog entity", zap.Error(err))
		return nil
	}

	// current maxID is newer so skip updating dialog unread counts
	if dtoDlg.ReadInboxMaxID > maxID {
		logs.Debug("repoDialogs::UpdateReadInboxMaxID() skipped dialogs unreadCount update",
			zap.Int64("Current MaxID", dtoDlg.ReadInboxMaxID),
			zap.Int64("Server MaxID", maxID),
		)
		return nil
	}

	// calculate unread count
	// currently when we enter dialog the max unreaded message ID will be sent
	var unreadCount int

	em := new(dto.Messages)
	err = r.db.Table(em.TableName()).Where("SenderID <> ? AND PeerID = ? AND PeerType = ? AND ID > ? ", userID, peerID, peerType, maxID).Count(&unreadCount).Error
	if err != nil {
		logs.Error("Dialogs::UpdateReadInboxMaxID()-> fetch messages unread count", zap.Error(err))
		return err
	}

	ed := new(dto.Dialogs)
	err = r.db.Table(ed.TableName()).Where("PeerID = ? AND PeerType = ?", peerID, peerType).Updates(map[string]interface{}{
		"ReadInboxMaxID": maxID,
		"UnreadCount":    unreadCount, //gorm.Expr("UnreadCount + ?", unreadCount), // in snapshot mode if unread message lefted
		"MentionedCount": 0,           // Hotfix :: on each ReadHistoryInbox set mentioned count to zero
	}).Error

	if err != nil {
		logs.Error("Dialogs::UpdateReadInboxMaxID()-> update dialog entity", zap.Error(err))
		return err
	}
	return nil
}

// UpdateReadOutboxMaxID
func (r *repoDialogs) UpdateReadOutboxMaxID(peerID int64, peerType int32, maxID int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	logs.Debug("Dialogs::UpdateReadOutboxMaxID",
		zap.Int64("PeerID", peerID),
		zap.Int32("PeerType", peerType),
		zap.Int64("MaxID", maxID),
	)

	ed := new(dto.Dialogs)
	err := r.db.Table(ed.TableName()).Where("PeerID = ? AND PeerType = ?", peerID, peerType).Updates(map[string]interface{}{
		"ReadOutboxMaxID": maxID,
	}).Error

	if err != nil {
		logs.Error("Dialogs::UpdateReadOutboxMaxID()-> update dialog entity", zap.Error(err))
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

func (r *repoDialogs) UpdateNotifySetting(msg *msg.UpdateNotifySettings) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if msg.NotifyPeer == nil {
		return errors.New("Dialogs::UpdateNotifySetting() => msg.NotifyPeer is null")
	}
	if msg.Settings == nil {
		return errors.New("Dialogs::UpdateNotifySetting() => msg.Settings is null")
	}

	logs.Debug("Dialogs::UpdateNotifySetting()",
		zap.Int64("UserId", msg.UserID),
		zap.Int64("PeerID", msg.NotifyPeer.ID),
		zap.Uint64("AccessHash", msg.NotifyPeer.AccessHash),
		zap.Int32("Flag", msg.Settings.Flags),
		zap.Int64("MuteUntil", msg.Settings.MuteUntil),
		zap.String("Sound", msg.Settings.Sound),
	)

	dtoDlg := new(dto.Dialogs)
	err := r.db.Where("PeerID = ? AND PeerType = ?", msg.NotifyPeer.ID, msg.NotifyPeer.Type).First(dtoDlg).Error
	if err != nil {
		logs.Error("Dialogs::UpdateNotifySetting()->fetch dialog entity", zap.Error(err))
		return err
	}
	dtoDlg.NotifyFlags = msg.Settings.Flags
	dtoDlg.NotifyMuteUntil = msg.Settings.MuteUntil
	dtoDlg.NotifySound = msg.Settings.Sound

	return r.db.Save(dtoDlg).Error
}

// UpdatePeerNotifySettings
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

// GetManyDialog
func (r *repoDialogs) GetManyDialog(peerIDs []int64) []*msg.Dialog {
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
