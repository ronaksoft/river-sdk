package repo

import (
	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/log"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/repo/dto"
	"github.com/kataras/iris/core/errors"
	"go.uber.org/zap"
)

type RepoDialogs interface {
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
		log.LOG.Debug("RepoDialogs::UpdateTopMesssageID() TopMessage",
			zap.String("Error", err.Error()),
		)
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
		log.LOG.Debug("RepoDialogs::SaveDialog()",
			zap.String("Error", "dialog is null"),
		)
		return domain.ErrNotFound
	}

	log.LOG.Debug("RepoDialogs::SaveDialog",
		zap.Int64("PeerID", dialog.PeerID),
		zap.Uint64("AccessHash", dialog.AccessHash),
		zap.Int32("UnreadCount", dialog.UnreadCount),
	)
	d := new(dto.Dialogs)
	d.Map(dialog)

	d.LastUpdate = lastUpdate

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

	log.LOG.Debug("RepoDialogs::UpdateDialogUnreadCount()",
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

	log.LOG.Debug("RepoDialogs::GetDialogs()",
		zap.Int32("Offset", offset),
		zap.Int32("Limit", limit),
	)

	dtoDlgs := make([]dto.Dialogs, 0, limit)

	err := r.db.Limit(limit).Offset(offset).Find(&dtoDlgs).Error
	if err != nil {
		log.LOG.Debug("RepoDialogs::GetDialogs()",
			zap.String("Error", err.Error()),
		)
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

	log.LOG.Debug("RepoDialogs::GetDialog()",
		zap.Int64("PeerID", peerID),
		zap.Int32("PeerType", peerType),
	)

	dtoDlg := new(dto.Dialogs)
	err := r.db.Where("PeerID = ? AND PeerType = ?", peerID, peerType).First(dtoDlg).Error
	if err != nil {
		log.LOG.Debug("RepoDialogs::GetDialog()->fetch dialog entity",
			zap.String("Error", err.Error()),
		)
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

	log.LOG.Debug("RepoDialogs::UpdateReadInboxMaxID()",
		zap.Int64("PeerID", peerID),
		zap.Int32("PeerType", peerType),
		zap.Int64("MaxID", maxID),
	)

	// get dialog
	dtoDlg := new(dto.Dialogs)
	err := r.db.Where("PeerID = ? AND PeerType = ?", peerID, peerType).First(dtoDlg).Error
	if err != nil {
		log.LOG.Debug("RepoDialogs::UpdateReadInboxMaxID()-> fetch dialog entity",
			zap.String("Error", err.Error()),
		)
		return nil
	}

	// current maxID is newer so skip updating dialog unread counts
	if dtoDlg.ReadInboxMaxID > maxID {
		log.LOG.Debug("repoDialogs::UpdateReadInboxMaxID() skipped dialogs unreadCount update",
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
		log.LOG.Debug("RepoDialogs::UpdateReadInboxMaxID()-> fetch messages unread count",
			zap.String("Error", err.Error()),
		)
		return err
	}

	ed := new(dto.Dialogs)
	err = r.db.Table(ed.TableName()).Where("PeerID = ? AND PeerType = ?", peerID, peerType).Updates(map[string]interface{}{
		"ReadInboxMaxID": maxID,
		"UnreadCount":    unreadCount, //gorm.Expr("UnreadCount + ?", unreadCount), // in snapshot mode if unread message lefted
	}).Error

	if err != nil {
		log.LOG.Debug("RepoDialogs::UpdateReadInboxMaxID()-> update dialog entity",
			zap.String("Error", err.Error()),
		)
		return err
	}
	return nil
}

// UpdateReadOutboxMaxID
func (r *repoDialogs) UpdateReadOutboxMaxID(peerID int64, peerType int32, maxID int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	log.LOG.Debug("RepoDialogs::UpdateReadOutboxMaxID",
		zap.Int64("PeerID", peerID),
		zap.Int32("PeerType", peerType),
		zap.Int64("MaxID", maxID),
	)

	ed := new(dto.Dialogs)
	err := r.db.Table(ed.TableName()).Where("PeerID = ? AND PeerType = ?", peerID, peerType).Updates(map[string]interface{}{
		"ReadOutboxMaxID": maxID,
	}).Error

	if err != nil {
		log.LOG.Debug("RepoDialogs::UpdateReadOutboxMaxID()-> update dialog entity",
			zap.String("Error", err.Error()),
		)
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
		return errors.New("RepoDialogs::UpdateNotifySetting() => msg.NotifyPeer is null")
	}
	if msg.Settings == nil {
		return errors.New("RepoDialogs::UpdateNotifySetting() => msg.Settings is null")
	}

	log.LOG.Debug("RepoDialogs::UpdateNotifySetting()",
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
		log.LOG.Debug("RepoDialogs::UpdateNotifySetting()->fetch dialog entity",
			zap.String("Error", err.Error()),
		)
		return err
	}
	dtoDlg.NotifyFlags = msg.Settings.Flags
	dtoDlg.NotifyMuteUntil = msg.Settings.MuteUntil
	dtoDlg.NotifySound = msg.Settings.Sound

	return r.db.Save(dtoDlg).Error
}
