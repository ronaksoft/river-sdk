package repo

import (
	"fmt"
	"sort"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo/dto"
	"go.uber.org/zap"
)

type repoMessages struct {
	*repository
}

func (r *repoMessages) SaveNewMessage(message *msg.UserMessage, dialog *msg.Dialog, userID int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if message == nil {
		return domain.ErrNotFound
	}

	m := new(dto.Messages)
	m.Map(message)

	dlg := new(dto.Dialogs)
	dlg.Map(dialog)

	em := new(dto.Messages)
	if err := r.db.Create(m).Error; err != nil {
		logs.Error("Repo::SaveNewMessage()-> save message", zap.Error(err))
	}

	// calculate unread count
	dtoDlg := new(dto.Dialogs)
	err := r.db.Where("PeerID = ? AND PeerType = ?", m.PeerID, m.PeerType).First(dtoDlg).Error
	if err != nil {
		logs.Warn("Repo::SaveNewMessage()-> fetch dialog entity", zap.Error(err))
		return err
	}
	unreadCount := dtoDlg.UnreadCount
	mentionedCount := dtoDlg.MentionedCount
	// newMessage : m.ID > topMessage
	// isNewMsg : if message delivered unordered or late
	if m.SenderID != userID && m.ID > dtoDlg.TopMessageID {
		unreadCount++
		for _, entity := range message.Entities {
			if entity.Type == msg.MessageEntityTypeMention && entity.UserID == userID {
				mentionedCount++
			}
		}
	}

	err = r.db.Table(em.TableName()).Where("PeerID=? AND PeerType=?", m.PeerID, m.PeerType).Limit(1).Order("ID DESC").Find(em).Error
	if err != nil {
		logs.Warn("Repo::SaveNewMessage()-> fetch top message", zap.Error(err))
		return err
	}

	topMessageID := m.ID
	if em.ID > m.ID {
		topMessageID = em.ID
	}

	ed := new(dto.Dialogs)
	err = r.db.Table(ed.TableName()).Where("PeerID=? AND PeerType=?", m.PeerID, m.PeerType).Updates(map[string]interface{}{
		"TopMessageID":   topMessageID,
		"LastUpdate":     message.CreatedOn,
		"UnreadCount":    unreadCount,
		"MentionedCount": mentionedCount,
	}).Error
	if err != nil {
		logs.Error("Repo::SaveNewMessage()-> update dialog entity", zap.Error(err))
		return err
	}

	return nil

}

func (r *repoMessages) SaveSelfMessage(message *msg.UserMessage, dialog *msg.Dialog) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if message == nil {
		return domain.ErrNotFound
	}

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

func (r *repoMessages) SaveMessage(message *msg.UserMessage) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if message == nil {
		return domain.ErrNotFound
	}

	m := new(dto.Messages)
	m.Map(message)

	em := new(dto.Messages)
	r.db.Find(em, m.ID)
	if em.ID == 0 {
		return r.db.Create(m).Error
	}

	return r.db.Table(m.TableName()).Where("id=?", m.ID).Update(m).Error
}

func (r *repoMessages) GetManyMessages(messageIDs []int64) []*msg.UserMessage {
	r.mx.Lock()
	defer r.mx.Unlock()

	messages := make([]*msg.UserMessage, 0, len(messageIDs))
	dtoMsgs := make([]dto.Messages, 0, len(messageIDs))
	err := r.db.Where("ID in (?)", messageIDs).Find(&dtoMsgs).Error
	if err != nil {
		logs.Warn("Repo::GetManyMessages()-> fetch messages", zap.Error(err))
		return nil
	}

	for _, v := range dtoMsgs {

		tmp := new(msg.UserMessage)
		v.MapTo(tmp)
		messages = append(messages, tmp)
	}

	return messages
}

func (r *repoMessages) GetMessageHistoryWithPendingMessages(peerID int64, peerType int32, minID, maxID int64, limit int32) (protoMsgs []*msg.UserMessage, protoUsers []*msg.User) {
	r.mx.Lock()
	defer r.mx.Unlock()

	dtoMsgs := make([]dto.Messages, 0, limit)
	dtoPendings := make([]dto.MessagesPending, 0, limit)

	var err error
	if minID == 0 && maxID == 0 {
		err = r.db.Order("ID DESC").Limit(limit).Where("PeerID = ? AND PeerType = ? ", peerID, peerType).Find(&dtoMsgs).Error
	} else {
		// there is no way this will happens :/
	}

	if err != nil {
		logs.Warn("Repo::GetMessageHistory()-> fetch messages", zap.Error(err))
		return
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
		protoMsgs = append(protoMsgs, tmp)
		userIDs[v.SenderID] = true
		userIDs[v.FwdSenderID] = true
		// load MessageActionData users
		actionUserIds := domain.ExtractActionUserIDs(v.MessageAction, v.MessageActionData)
		for _, id := range actionUserIds {
			userIDs[id] = true
		}
	}

	// Get users <rewrite it here to remove coupling>
	users := make([]dto.Users, 0, len(userIDs))

	err = r.db.Where("ID in (?)", userIDs.ToArray()).Find(&users).Error
	if err != nil {
		logs.Warn("Repo::GetMessageHistory()-> fetch users", zap.Error(err))
		return
	}

	for _, v := range users {
		tmp := new(msg.User)
		v.MapToUser(tmp)
		protoUsers = append(protoUsers, tmp)

	}
	return
}

func (r *repoMessages) GetMessageHistory(peerID int64, peerType int32, minID, maxID int64, limit int32) (protoMsgs []*msg.UserMessage, protoUsers []*msg.User) {
	r.mx.Lock()
	defer r.mx.Unlock()

	dtoResult := make([]dto.Messages, 0, limit)

	var err error
	if maxID == 0 && minID == 0 {
		err = r.db.Order("ID DESC").Limit(limit).Where("PeerID = ? AND PeerType = ?", peerID, peerType).Find(&dtoResult).Error
	} else if minID == 0 && maxID != 0 {
		err = r.db.Order("ID DESC").Limit(limit).Where("PeerID = ? AND PeerType = ? AND messages.ID <= ?", peerID, peerType, maxID).Find(&dtoResult).Error
	} else if minID != 0 && maxID == 0 {
		err = r.db.Order("ID ASC").Limit(limit).Where("PeerID = ? AND PeerType = ? AND messages.ID >= ?", peerID, peerType, minID).Find(&dtoResult).Error
		// sort DESC again
		if err == nil {
			sort.Slice(dtoResult, func(i, j int) bool {
				return dtoResult[i].ID > dtoResult[j].ID
			})
		}
	} else {
		err = r.db.Order("ID DESC").Limit(limit).Where("PeerID = ? AND PeerType = ? AND messages.ID >= ? AND messages.ID <= ?", peerID, peerType, minID, maxID).Find(&dtoResult).Error
	}

	if err != nil {
		logs.Warn("Repo::GetMessageHistory()-> fetch messages", zap.Error(err))
		return
	}
	userIDs := domain.MInt64B{}
	for _, v := range dtoResult {
		tmp := new(msg.UserMessage)
		v.MapTo(tmp)
		protoMsgs = append(protoMsgs, tmp)
		userIDs[v.SenderID] = true
		userIDs[v.FwdSenderID] = true
		// load MessageActionData users
		actionUserIds := domain.ExtractActionUserIDs(v.MessageAction, v.MessageActionData)
		for _, id := range actionUserIds {
			userIDs[id] = true
		}
	}

	// Get users <rewrite it here to remove coupling>
	users := make([]dto.Users, 0, len(userIDs))

	err = r.db.Where("ID in (?)", userIDs.ToArray()).Find(&users).Error
	if err != nil {
		logs.Warn("Repo::GetMessageHistory()-> fetch users", zap.Error(err))
		return
	}

	for _, v := range users {
		tmp := new(msg.User)
		v.MapToUser(tmp)
		protoUsers = append(protoUsers, tmp)

	}
	return
}

func (r *repoMessages) GetUnreadMessageCount(peerID int64, peerType int32, userID, maxID int64) int32 {
	r.mx.Lock()
	defer r.mx.Unlock()

	count := 0
	r.db.Model(&dto.Messages{}).Where("PeerID = ? AND PeerType = ? AND SenderID <> ? AND ID>?", peerID, peerType, userID, maxID).Count(&count)
	return int32(count)
}

func (r *repoMessages) SaveMessageMany() {

}

func (r *repoMessages) DeleteDialogMessage(peerID int64, peerType int32, maxID int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	return r.db.Where("PeerID=? AND PeerType=? AND ID <= ?", peerID, peerType, maxID).Delete(dto.Messages{}).Error
}

func (r *repoMessages) DeleteMany(IDs []int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	// Get dialogs that their top message is going to be removed
	dtoDialogs := make([]dto.Dialogs, 0)
	err := r.db.Where("TopMessageID in (?)", IDs).Find(&dtoDialogs).Error
	if err != nil {
		logs.Warn("Messages::DeleteMany() fetch dialogs", zap.Error(err))
	}

	// remove message
	err = r.db.Where("ID IN (?)", IDs).Delete(dto.Messages{}).Error
	if err != nil {
		logs.Warn("Messages::DeleteMany() delete from Messages", zap.Error(err))
	}

	// fetch last message and set it as dialog top message
	for _, d := range dtoDialogs {
		dtoMsg := dto.Messages{}
		err := r.db.Table(dtoMsg.TableName()).Where("PeerID =? AND PeerType= ?", d.PeerID, d.PeerType).Last(&dtoMsg).Error
		if err == nil && dtoMsg.ID != 0 {
			d.TopMessageID = dtoMsg.ID
			d.LastUpdate = dtoMsg.CreatedOn
			r.db.Save(d)
		}
	}

	return err
}

func (r *repoMessages) DeleteManyAndReturnClientUpdate(IDs []int64) ([]*msg.ClientUpdateMessagesDeleted, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	// Get dialogs that their top message is going to be removed
	dtoDialogs := make([]dto.Dialogs, 0)
	err := r.db.Where("TopMessageID in (?)", IDs).Find(&dtoDialogs).Error
	if err != nil {
		logs.Warn("Messages::DeleteMany() fetch dialogs", zap.Error(err))
	}

	res := make([]*msg.ClientUpdateMessagesDeleted, 0)
	msgs := make([]dto.Messages, 0)
	mpeer := make(map[int64]*msg.ClientUpdateMessagesDeleted)
	err = r.db.Where("ID in (?)", IDs).Find(&msgs).Error
	if err == nil {
		for _, v := range msgs {

			if udp, ok := mpeer[v.PeerID]; ok {
				udp.MessageIDs = append(udp.MessageIDs, v.ID)
			} else {
				tmp := new(msg.ClientUpdateMessagesDeleted)
				tmp.PeerID = v.PeerID
				tmp.PeerType = v.PeerType
				tmp.MessageIDs = []int64{v.ID}
				mpeer[v.PeerID] = tmp
			}
		}
	}

	for _, v := range mpeer {
		// Update Dialog Counter on delete message
		dtoDlg := new(dto.Dialogs)
		err := r.db.Where("PeerID = ? AND PeerType = ?", v.PeerID, v.PeerType).First(dtoDlg).Error
		if err == nil {
			removedUnreadCount := int32(0)
			for _, msgID := range v.MessageIDs {
				if msgID > dtoDlg.ReadInboxMaxID {
					removedUnreadCount++
				}
			}
			if removedUnreadCount > 0 && removedUnreadCount <= dtoDlg.UnreadCount {
				err = r.db.Table(dtoDlg.TableName()).Where("PeerID=? AND PeerType=?", v.PeerID, v.PeerType).Updates(map[string]interface{}{
					"UnreadCount": (dtoDlg.UnreadCount - removedUnreadCount), // gorm.Expr("UnreadCount - ?", removedUnreadCount),
				}).Error
			}
		}
		res = append(res, v)
	}

	err = r.db.Where("ID IN (?)", IDs).Delete(dto.Messages{}).Error

	// fetch last message and set it as dialog top message
	for _, d := range dtoDialogs {
		dtoMsg := dto.Messages{}
		err := r.db.Table(dtoMsg.TableName()).Where("PeerID =? AND PeerType= ?", d.PeerID, d.PeerType).Last(&dtoMsg).Error
		if err == nil && dtoMsg.ID != 0 {
			err = r.db.Table(d.TableName()).Where("PeerID=? AND PeerType=?", d.PeerID, d.PeerType).Updates(map[string]interface{}{
				"TopMessageID": dtoMsg.ID,
				"LastUpdate":   dtoMsg.CreatedOn,
			}).Error
		}
	}

	return res, err
}

func (r *repoMessages) GetTopMessageID(peerID int64, peerType int32) (int64, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	dtoMsg := dto.Messages{}
	err := r.db.Table(dtoMsg.TableName()).Where("PeerID =? AND PeerType= ?", peerID, peerType).Last(&dtoMsg).Error
	if err != nil {
		logs.Warn("Repo::GetTopMessageID()-> fetch message", zap.Error(err))
		return -1, err
	}
	return dtoMsg.ID, nil
}

func (r *repoMessages) GetMessage(messageID int64) *msg.UserMessage {
	r.mx.Lock()
	defer r.mx.Unlock()

	message := new(msg.UserMessage)
	dtoMsg := new(dto.Messages)
	err := r.db.Where("ID = ?", messageID).Find(&dtoMsg).Error
	if err != nil {
		logs.Warn("Repo::GetMessage()-> fetch messages", zap.Error(err))
		return nil
	}

	dtoMsg.MapTo(message)

	return message
}

func (r *repoMessages) SetContentRead(messageIDs []int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	mdl := dto.Messages{}
	return r.db.Table(mdl.TableName()).Where("ID in (?)", messageIDs).Updates(map[string]interface{}{
		"ContentRead": true,
	}).Error
}

func (r *repoMessages) GetDialogMessageCount(peerID int64, peerType int32) int32 {
	r.mx.Lock()
	defer r.mx.Unlock()

	count := int32(0)
	mdl := &dto.Messages{}
	r.db.Table(mdl.TableName()).Where("PeerID =? AND PeerType= ?", peerID, peerType).Count(&count)

	return count
}

func (r *repoMessages) SearchText(text string) []*msg.UserMessage {
	r.mx.Lock()
	defer r.mx.Unlock()
	var userMsgs []*msg.UserMessage
	msgs := make([]dto.Messages, 0)
	if err := r.db.Where("Body LIKE ?", "%"+fmt.Sprintf("%s", text)+"%").Find(&msgs).Error; err != nil {
		logs.Warn("SearchText::error", zap.String("", err.Error()))
		return userMsgs
	}

	for _, dtoMsg := range msgs {
		message := new(msg.UserMessage)
		dtoMsg.MapTo(message)
		userMsgs = append(userMsgs, message)
	}
	return userMsgs
}

func (r *repoMessages) SearchTextByPeerID(text string, peerID int64) []*msg.UserMessage {
	r.mx.Lock()
	defer r.mx.Unlock()
	var userMsgs []*msg.UserMessage
	msgs := make([]dto.Messages, 0)
	if err := r.db.Where("Body LIKE ? AND PeerID = ?", "%"+fmt.Sprintf("%s", text)+"%", peerID).Find(&msgs).Error; err != nil {
		logs.Warn("SearchText::error", zap.String("", err.Error()))
		return userMsgs
	}

	for _, dtoMsg := range msgs {
		message := new(msg.UserMessage)
		dtoMsg.MapTo(message)
		userMsgs = append(userMsgs, message)
	}
	return userMsgs
}
