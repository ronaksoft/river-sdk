package repo

import (
	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/log"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/repo/dto"
	"go.uber.org/zap"
)

type RepoMessages interface {
	SaveNewMessage(message *msg.UserMessage, dialog *msg.Dialog, userID int64) error
	SaveMessage(message *msg.UserMessage) error
	SaveSelfMessage(message *msg.UserMessage, dialog *msg.Dialog) error
	GetManyMessages(messageIDs []int64) []*msg.UserMessage
	GetMessageHistory(peerID int64, peerType int32, minID, maxID int64, limit int32) ([]*msg.UserMessage, []*msg.User)
	GetUnreadMessageCount(peerID int64, peerType int32, userID, maxID int64) int32
	DeleteDialogMessage(peerID int64, peerType int32, maxID int64) error
	DeleteMany(IDs []int64) error
}

type repoMessages struct {
	*repository
}

// SaveNewMessage
func (r *repoMessages) SaveNewMessage(message *msg.UserMessage, dialog *msg.Dialog, userID int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if message == nil {
		log.LOG_Debug("RepoRepoMessages::SaveNewMessage()",
			zap.String("Error", "message is null"),
		)
		return domain.ErrNotFound
	}

	log.LOG_Debug("RepoRepoMessages::SaveNewMessage()",
		zap.Int64("MessageID", message.ID),
		zap.Int64("DialogPeerID", dialog.PeerID),
	)

	m := new(dto.Messages)
	m.Map(message)

	dlg := new(dto.Dialogs)
	dlg.Map(dialog)

	em := new(dto.Messages)
	isNewMsg := false
	r.db.Find(em, m.ID)
	if em.ID == 0 {
		isNewMsg = true
		r.db.Create(m)
	} else {
		r.db.Table(m.TableName()).Where("ID=?", m.ID).Update(m)
	}

	// calculate unread count

	dtoDlg := new(dto.Dialogs)
	err := r.db.Where("PeerID = ? AND PeerType = ?", m.PeerID, m.PeerType).First(dtoDlg).Error
	if err != nil {
		log.LOG_Debug("RepoRepoMessages::SaveNewMessage()-> fetch dialog entity",
			zap.String("Error", err.Error()),
		)
		return err
	}
	unreadCount := dtoDlg.UnreadCount
	// newMessage : m.ID > topMessage
	// isNewMsg : if message delivered unordered or late
	if m.SenderID != userID && (isNewMsg || m.ID > dtoDlg.TopMessageID) {
		unreadCount++
	}
	// var unreadCount int
	// maxID := dtoDlg.ReadInboxMaxID
	// err = r.db.Table(em.TableName()).Where("SenderID <> ? AND PeerID = ? AND PeerType = ? AND ID > ? ", userID, m.PeerID, m.PeerType, maxID).Count(&unreadCount).Error
	// if err != nil {
	// 	log.LOG_Debug("RepoRepoMessages::SaveNewMessage()-> fetch messages unread count",
	// 		zap.String("Error", err.Error()),
	// 	)
	// 	return err
	// }

	err = r.db.Table(em.TableName()).Where("PeerID=? AND PeerType=?", m.PeerID, m.PeerType).Limit(1).Order("ID DESC").Find(em).Error
	if err != nil {
		log.LOG_Debug("RepoRepoMessages::SaveNewMessage()-> fetch top message",
			zap.String("Error", err.Error()),
		)
		return err
	}

	topMessageID := m.ID
	if em.ID > m.ID {
		topMessageID = em.ID
	}

	ed := new(dto.Dialogs)
	r.db.LogMode(true)
	err = r.db.Table(ed.TableName()).Where("PeerID=? AND PeerType=?", m.PeerID, m.PeerType).Updates(map[string]interface{}{
		"TopMessageID": topMessageID,
		"LastUpdate":   message.CreatedOn,
		"UnreadCount":  unreadCount, //gorm.Expr("UnreadCount + ?", unreadCount), // in snapshot mode if unread message lefted
	}).Error
	r.db.LogMode(false)
	if err != nil {
		log.LOG_Debug("RepoRepoMessages::SaveNewMessage()-> update dialog entity",
			zap.String("Error", err.Error()),
		)
		return err
	}

	return nil

}

// SaveSelfMessage saves new message and set TopMessageID but do not increase unreadcount cuz sender is myself
func (r *repoMessages) SaveSelfMessage(message *msg.UserMessage, dialog *msg.Dialog) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if message == nil {
		log.LOG_Debug("RepoRepoMessages::SaveSelfMessage()",
			zap.String("Error", "message is null"),
		)
		return domain.ErrNotFound
	}

	log.LOG_Debug("RepoRepoMessages::SaveSelfMessage()",
		zap.Int64("MessageID", message.ID),
		zap.Int64("DialogPeerID", dialog.PeerID),
	)

	m := new(dto.Messages)
	m.Map(message)

	dlg := new(dto.Dialogs)
	dlg.Map(dialog)

	em := new(dto.Messages)
	r.db.Find(em, m.ID)
	if em.ID == 0 {
		r.db.Create(m)
	} else {
		r.db.Table(m.TableName()).Where("ID=?", m.ID).Update(m)
	}

	if dialog.TopMessageID < message.ID {

		ed := new(dto.Dialogs)
		err := r.db.Table(ed.TableName()).Where("PeerID=? AND PeerType=?", m.PeerID, m.PeerType).Updates(map[string]interface{}{
			"TopMessageID": message.ID,
			"LastUpdate":   message.CreatedOn,
		}).Error

		return err
	}

	return nil

}

// SaveMessage
func (r *repoMessages) SaveMessage(message *msg.UserMessage) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if message == nil {
		log.LOG_Debug("RepoRepoMessages::SaveMessage()",
			zap.String("Error", "message is null"),
		)
		return domain.ErrNotFound
	}

	// log.LOG_Debug("RepoMessages::SaveMessage()",
	// 	zap.Int64("MessageID", message.ID),
	// )

	m := new(dto.Messages)
	m.Map(message)

	em := new(dto.Messages)
	r.db.Find(em, m.ID)
	if em.ID == 0 {
		return r.db.Create(m).Error
	}

	return r.db.Table(m.TableName()).Where("id=?", m.ID).Update(m).Error
}

// GetManyMessages
func (r *repoMessages) GetManyMessages(messageIDs []int64) []*msg.UserMessage {
	r.mx.Lock()
	defer r.mx.Unlock()

	log.LOG_Debug("RepoMessages::GetManyMessages()",
		zap.Int64s("MessageIDs", messageIDs),
	)
	messages := make([]*msg.UserMessage, 0, len(messageIDs))
	dtoMsgs := make([]dto.Messages, 0, len(messageIDs))
	err := r.db.Where("ID in (?)", messageIDs).Find(&dtoMsgs).Error
	if err != nil {
		log.LOG_Debug("RepoRepoMessages::GetManyMessages()-> fetch messages",
			zap.String("Error", err.Error()),
		)
		return nil
	}

	for _, v := range dtoMsgs {

		tmp := new(msg.UserMessage)
		v.MapTo(tmp)
		messages = append(messages, tmp)
	}

	return messages
}

// GetMessageHistory
func (r *repoMessages) GetMessageHistory(peerID int64, peerType int32, minID, maxID int64, limit int32) ([]*msg.UserMessage, []*msg.User) {
	r.mx.Lock()
	defer r.mx.Unlock()

	log.LOG_Debug("RepoMessages::GetMessageHistory()",
		zap.Int64("PeerID", peerID),
		zap.Int32("PeerType", peerType),
		zap.Int64("MinID", minID),
		zap.Int64("MaxID", maxID),
		zap.Int32("Limit", limit),
	)

	messages := make([]*msg.UserMessage, 0, limit)
	dtoMsgs := make([]dto.Messages, 0, limit)
	dtoPendings := make([]dto.PendingMessages, 0, limit)

	var err error
	if minID == 0 && maxID == 0 {
		err = r.db.Order("ID DESC").Limit(limit).Where("PeerID = ? AND PeerType = ? ", peerID, peerType).Find(&dtoMsgs).Error
	} else {
		err = r.db.Order("ID DESC").Limit(limit).Where("PeerID = ? AND PeerType = ? AND messages.ID > ? AND messages.ID < ?", peerID, peerType, minID, maxID).Find(&dtoMsgs).Error
	}

	if err != nil {
		log.LOG_Debug("RepoRepoMessages::GetMessageHistory()-> fetch messages",
			zap.String("Error", err.Error()),
		)
		return nil, nil
	}

	dtoResult := make([]dto.Messages, 0, limit)

	// get all pending message for this user
	err = r.db.Order("ID ASC").Limit(limit).Where("PeerID = ? AND PeerType = ? ", peerID, peerType).Find(&dtoPendings).Error
	if err == nil {
		for _, v := range dtoPendings {
			tmp := new(dto.Messages)
			v.MapToDtoMessage(tmp)
			dtoResult = append(dtoResult, *tmp)
			// dtoMsgs = append(dtoMsgs, *tmp)
		}
	}
	dtoResult = append(dtoResult, dtoMsgs...)

	userIDs := domain.MInt64B{}
	for _, v := range dtoResult {
		tmp := new(msg.UserMessage)
		v.MapTo(tmp)
		messages = append(messages, tmp)
		userIDs[v.SenderID] = true
		// // pending messages
		// if v.ID < 0 {
		// 	userIDs[v.PeerID] = true
		// }
	}

	// Get users <rewrite it here to remove coupling>
	pbUsers := make([]*msg.User, 0, len(userIDs))
	users := make([]dto.Users, 0, len(userIDs))

	err = r.db.Where("ID in (?)", userIDs.ToArray()).Find(&users).Error
	if err != nil {
		log.LOG_Debug("RepoRepoMessages::GetMessageHistory()-> fetch users",
			zap.String("Error", err.Error()),
		)
		return nil, nil // , err
	}

	for _, v := range users {
		tmp := new(msg.User)
		v.MapToUser(tmp)
		pbUsers = append(pbUsers, tmp)

	}
	return messages, pbUsers
}

func (r *repoMessages) GetUnreadMessageCount(peerID int64, peerType int32, userID, maxID int64) int32 {
	r.mx.Lock()
	defer r.mx.Unlock()

	count := 0
	r.db.Model(&dto.Messages{}).Where("PeerID = ? AND PeerType = ? AND SenderID <> ? AND ID>?", peerID, peerType, userID, maxID).Count(&count)
	return int32(count)
}

// save message batch mode
func (r *repoMessages) SaveMessageMany() {

}

func (r *repoMessages) DeleteDialogMessage(peerID int64, peerType int32, maxID int64) error {

	return r.db.Where("PeerID=? AND PeerType=? AND ID <= ?", peerID, peerType, maxID).Delete(dto.Messages{}).Error
}
func (r *repoMessages) DeleteMany(IDs []int64) error {

	return r.db.Where("ID IN (?)", IDs).Delete(dto.Messages{}).Error
}
