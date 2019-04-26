package repo

import (
	"time"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo/dto"
	"go.uber.org/zap"
)

// MessagesPending repoMessagesPending interface
type MessagesPending interface {
	Save(ID int64, senderID int64, message *msg.MessagesSend) (*msg.ClientPendingMessage, error)
	GetPendingMessageByRequestID(requestID int64) (*dto.MessagesPending, error)
	GetPendingMessageByID(id int64) (*dto.MessagesPending, error)
	DeletePendingMessage(ID int64) error
	DeleteManyPendingMessage(IDs []int64) error
	GetManyPendingMessages(messageIDs []int64) []*msg.UserMessage
	GetManyPendingMessagesRequestID(messageIDs []int64) []int64
	GetAllPendingMessages() []*msg.MessagesSend
	DeletePeerAllMessages(peerID int64, peerType int32) (*msg.ClientUpdateMessagesDeleted, error)
	SaveClientMessageMedia(ID, senderID, requestID int64, msgMedia *msg.ClientSendMessageMedia) (*msg.ClientPendingMessage, error)
	GetPendingMessage(messageID int64) *msg.UserMessage
	SaveMessageMedia(ID int64, senderID int64, message *msg.MessagesSendMedia) (*msg.ClientPendingMessage, error)
}

type repoMessagesPending struct {
	*repository
}

func (r *repoMessagesPending) Save(ID int64, senderID int64, message *msg.MessagesSend) (*msg.ClientPendingMessage, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	if message == nil {
		logs.Debug("MessagesPending::Save() :",
			zap.String("Error", "message is null"),
		)
		return nil, domain.ErrNotFound
	}
	logs.Debug("MessagesPending::Save()",
		zap.Int64("PendingMessageID", ID),
	)

	m := new(dto.MessagesPending)
	m.Map(message)
	m.ID = ID
	m.SenderID = senderID
	m.CreatedOn = time.Now().Unix()

	res := new(msg.ClientPendingMessage)
	m.MapTo(res)

	ep := new(dto.MessagesPending)
	r.db.Find(ep, m.ID)

	var err error
	if ep.ID == 0 {
		err = r.db.Create(m).Error
	} else {
		err = r.db.Table(m.TableName()).Where("ID", m.ID).Update(m).Error
	}
	if err != nil {
		return nil, err
	}

	dlg := new(dto.Dialogs)
	r.db.Where(&dto.Dialogs{PeerID: m.PeerID, PeerType: m.PeerType}).Find(&dlg)
	if dlg.PeerID == 0 {
		dlg.LastUpdate = m.CreatedOn
		dlg.PeerID = m.PeerID
		dlg.PeerType = m.PeerType
		dlg.TopMessageID = m.ID
		r.db.Create(dlg)
	} else {
		r.db.Table(dlg.TableName()).Where("PeerID=? AND PeerType=?", m.PeerID, m.PeerType).Updates(map[string]interface{}{
			"TopMessageID": m.ID,
			"LastUpdate":   m.CreatedOn,
		})

	}

	return res, nil
}

func (r *repoMessagesPending) GetPendingMessageByRequestID(requestID int64) (*dto.MessagesPending, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	logs.Debug("MessagesPending::GetPendingMessageByRequestID()",
		zap.Int64("RandomID/RequestID", requestID),
	)

	pmsg := new(dto.MessagesPending)
	err := r.db.Where("RequestID = ?", requestID).First(pmsg).Error
	if err != nil {
		logs.Error("MessagesPending::GetPendingMessageByRequestID()-> fetch pendingMessage entity", zap.Error(err))
		return nil, err
	}

	return pmsg, nil

}

func (r *repoMessagesPending) GetPendingMessageByID(id int64) (*dto.MessagesPending, error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	logs.Debug("MessagesPending::GetPendingMessageByID()",
		zap.Int64("ID", id),
	)

	pmsg := new(dto.MessagesPending)
	err := r.db.Where("ID = ?", id).First(pmsg).Error
	if err != nil {
		logs.Error("MessagesPending::GetPendingMessageByID()-> fetch pendingMessage entity", zap.Error(err))
		return nil, err
	}

	return pmsg, nil

}

func (r *repoMessagesPending) DeletePendingMessage(ID int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	logs.Debug("MessagesPending::DeletePendingMessage()",
		zap.Int64("ID", ID),
	)
	// get pending message
	pmsg := new(dto.MessagesPending)
	err := r.db.Where("ID = ?", ID).First(pmsg).Error
	if err != nil {
		logs.Error("MessagesPending::GetPendingMessageByID()-> fetch pendingMessage entity", zap.Error(err))
		return err
	}
	// get its dialog
	d := new(dto.Dialogs)
	err = r.db.Where("PeerID =? AND PeerType= ?", pmsg.PeerID, pmsg.PeerType).First(d).Error
	if err != nil {
		logs.Error("MessagesPending::GetPendingMessageByID()-> fetch dialog entity", zap.Error(err))
		return err
	}

	err = r.db.Where("ID = ?", ID).Delete(dto.MessagesPending{}).Error
	if err != nil {
		logs.Error("MessagesPending::DeletePendingMessage()-> delete pendingMessage entity", zap.Error(err))
		// return err
	}

	// if this message is this dialog top message id
	if d.TopMessageID == pmsg.ID {
		dtoPend := dto.MessagesPending{}
		err = r.db.Table(dtoPend.TableName()).Where("PeerID =? AND PeerType= ?", d.PeerID, d.PeerType).First(&dtoPend).Error // cuz the pendMsg Ids are negative of nano time so the smallest is latest record
		if err == nil && dtoPend.ID != 0 {
			d.TopMessageID = dtoPend.ID
			d.LastUpdate = dtoPend.CreatedOn
			err = r.db.Save(d).Error
		} else {
			dtoMsg := dto.Messages{}
			err = r.db.Table(dtoMsg.TableName()).Where("PeerID =? AND PeerType= ?", d.PeerID, d.PeerType).Last(&dtoMsg).Error
			if err == nil && dtoMsg.ID != 0 {
				d.TopMessageID = dtoMsg.ID
				d.LastUpdate = dtoMsg.CreatedOn
				err = r.db.Save(d).Error
			}
		}
	}

	return err
}

func (r *repoMessagesPending) DeleteManyPendingMessage(IDs []int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	// Get dialogs that their top message is going to be removed
	dtoDoalogs := make([]dto.Dialogs, 0)
	err := r.db.Where("TopMessageID in (?)", IDs).Find(&dtoDoalogs).Error
	if err != nil {
		logs.Error("MessagesPending::DeleteMany() fetch dialogs", zap.Error(err))
	}

	// remove pending message
	err = r.db.Where("ID IN (?)", IDs).Delete(dto.MessagesPending{}).Error
	if err != nil {
		logs.Error("MessagesPending::DeleteMany() delete from Messages", zap.Error(err))
	}

	// fetch last message and set it as dialog top message
	for _, d := range dtoDoalogs {
		dtoPend := dto.MessagesPending{}
		err := r.db.Table(dtoPend.TableName()).Where("PeerID =? AND PeerType= ?", d.PeerID, d.PeerType).First(&dtoPend).Error // cuz the pendMsg Ids are negative of nano time so the smallest is latest record
		if err == nil && dtoPend.ID != 0 {
			d.TopMessageID = dtoPend.ID
			d.LastUpdate = dtoPend.CreatedOn
			r.db.Save(d)
		} else {
			dtoMsg := dto.Messages{}
			err := r.db.Table(dtoMsg.TableName()).Where("PeerID =? AND PeerType= ?", d.PeerID, d.PeerType).Last(&dtoMsg).Error
			if err == nil && dtoMsg.ID != 0 {
				d.TopMessageID = dtoMsg.ID
				d.LastUpdate = dtoMsg.CreatedOn
				r.db.Save(d)
			}
		}
	}

	return nil
}

func (r *repoMessagesPending) GetManyPendingMessages(messageIDs []int64) []*msg.UserMessage {
	r.mx.Lock()
	defer r.mx.Unlock()

	logs.Debug("MessagesPending::GetManyPendingMessages()",
		zap.Int64s("MessageIDs", messageIDs),
	)
	messages := make([]*msg.UserMessage, 0, len(messageIDs))
	dtoMsgs := make([]dto.MessagesPending, 0, len(messageIDs))
	err := r.db.Where("ID in (?)", messageIDs).Find(&dtoMsgs).Error
	if err != nil {
		logs.Error("MessagesPending::GetManyPendingMessages()-> fetch pendingMessage entities", zap.Error(err))
		return nil
	}

	for _, v := range dtoMsgs {

		tmp := new(msg.UserMessage)
		v.MapToUserMessage(tmp)
		messages = append(messages, tmp)
	}

	return messages
}

func (r *repoMessagesPending) GetManyPendingMessagesRequestID(messageIDs []int64) []int64 {
	r.mx.Lock()
	defer r.mx.Unlock()

	logs.Debug("MessagesPending::GetManyPendingMessagesRequestID()",
		zap.Int64s("MessageIDs", messageIDs),
	)
	dtoMsgs := make([]dto.MessagesPending, 0, len(messageIDs))
	err := r.db.Where("ID in (?)", messageIDs).Find(&dtoMsgs).Error
	if err != nil {
		logs.Error("MessagesPending::GetManyPendingMessages()-> fetch pendingMessage entities", zap.Error(err))
		return nil
	}
	requestIDs := make([]int64, 0)
	for _, v := range dtoMsgs {
		requestIDs = append(requestIDs, v.RequestID)
	}

	return requestIDs
}

func (r *repoMessagesPending) GetAllPendingMessages() []*msg.MessagesSend {
	r.mx.Lock()
	defer r.mx.Unlock()

	res := make([]*msg.MessagesSend, 0)
	var dtos []dto.MessagesPending
	r.db.Where("MediaType=?", 0).Find(&dtos)

	for _, v := range dtos {
		tmp := new(msg.MessagesSend)
		v.MapToMessageSend(tmp)
		res = append(res, tmp)
	}

	return res
}

func (r *repoMessagesPending) DeletePeerAllMessages(peerID int64, peerType int32) (*msg.ClientUpdateMessagesDeleted, error) {

	msgs := make([]dto.MessagesPending, 0)
	err := r.db.Where("PeerID=? AND PeerType=?", peerID, peerType).Find(&msgs).Error
	if err != nil {
		logs.Error("MessagesPending::DeletePeerAllMessages()-> fetch pendingMessage entities", zap.Error(err))
		return nil, err
	}
	res := new(msg.ClientUpdateMessagesDeleted)
	res.PeerID = peerID
	res.PeerType = peerType
	res.MessageIDs = make([]int64, 0)
	for _, v := range msgs {
		res.MessageIDs = append(res.MessageIDs, v.ID)
	}
	err = r.db.Where("PeerID=? AND PeerType=?", peerID, peerType).Delete(dto.MessagesPending{}).Error
	if err != nil {
		logs.Error("MessagesPending::DeletePeerAllMessages()-> delete pendingMessage entities", zap.Error(err))
	}
	return res, err
}

func (r *repoMessagesPending) SaveClientMessageMedia(ID, senderID, requestID int64, msgMedia *msg.ClientSendMessageMedia) (*msg.ClientPendingMessage, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	if msgMedia == nil {
		logs.Debug("MessagesPending::msgMedia() :",
			zap.String("Error", "msgMedia is null"),
		)
		return nil, domain.ErrNotFound
	}
	logs.Debug("MessagesPending::SaveMedia()",
		zap.Int64("PendingMessageID", ID),
	)

	fileID := domain.SequentialUniqueID()
	m := new(dto.MessagesPending)
	m.MapFromClientMessageMedia(fileID, msgMedia)
	m.ID = ID
	m.SenderID = senderID
	m.CreatedOn = time.Now().Unix()
	m.RequestID = requestID

	res := new(msg.ClientPendingMessage)
	m.MapTo(res)
	ep := new(dto.MessagesPending)
	r.db.Find(ep, m.ID)

	var err error
	if ep.ID == 0 {
		err = r.db.Create(m).Error
	} else {
		err = r.db.Table(m.TableName()).Where("ID", m.ID).Update(m).Error
	}

	if err == nil {
		dlg := new(dto.Dialogs)
		r.db.Where(&dto.Dialogs{PeerID: m.PeerID, PeerType: m.PeerType}).Find(&dlg)
		if dlg.PeerID == 0 {
			dlg.LastUpdate = m.CreatedOn
			dlg.PeerID = m.PeerID
			dlg.PeerType = m.PeerType
			dlg.TopMessageID = m.ID
			r.db.Create(dlg)
		} else {
			r.db.Table(dlg.TableName()).Where("PeerID=? AND PeerType=?", m.PeerID, m.PeerType).Updates(map[string]interface{}{
				"TopMessageID": m.ID,
				"LastUpdate":   m.CreatedOn,
			})

		}
	}
	return res, err
}

func (r *repoMessagesPending) GetPendingMessage(messageID int64) *msg.UserMessage {
	r.mx.Lock()
	defer r.mx.Unlock()

	logs.Debug("MessagesPending::GetPendingMessages()",
		zap.Int64("MessageID", messageID),
	)
	message := new(msg.UserMessage)
	dtoMsg := new(dto.MessagesPending)
	err := r.db.Where("ID = ?", messageID).Find(&dtoMsg).Error
	if err != nil {
		logs.Error("MessagesPending::GetPendingMessages()-> fetch messages", zap.Error(err))
		return nil
	}

	dtoMsg.MapToUserMessage(message)

	return message
}

func (r *repoMessagesPending) SaveMessageMedia(ID int64, senderID int64, message *msg.MessagesSendMedia) (*msg.ClientPendingMessage, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	if message == nil {
		logs.Debug("MessagesPending::SaveMessageMedia() :",
			zap.String("Error", "message is null"),
		)
		return nil, domain.ErrNotFound
	}
	logs.Debug("MessagesPending::SaveMessageMedia()",
		zap.Int64("PendingMessageID", ID),
	)

	m := new(dto.MessagesPending)
	m.MapFromMessageMedia(message)
	m.ID = ID
	m.SenderID = senderID
	m.CreatedOn = time.Now().Unix()

	res := new(msg.ClientPendingMessage)
	m.MapTo(res)

	ep := new(dto.MessagesPending)
	r.db.Find(ep, m.ID)

	var err error
	if ep.ID == 0 {
		err = r.db.Create(m).Error
	} else {
		err = r.db.Table(m.TableName()).Where("ID", m.ID).Update(m).Error
	}

	if err == nil {
		dlg := new(dto.Dialogs)
		r.db.Where(&dto.Dialogs{PeerID: m.PeerID, PeerType: m.PeerType}).Find(&dlg)
		if dlg.PeerID == 0 {
			dlg.LastUpdate = m.CreatedOn
			dlg.PeerID = m.PeerID
			dlg.PeerType = m.PeerType
			dlg.TopMessageID = m.ID
			r.db.Create(dlg)
		} else {
			r.db.Table(dlg.TableName()).Where("PeerID=? AND PeerType=?", m.PeerID, m.PeerType).Updates(map[string]interface{}{
				"TopMessageID": m.ID,
				"LastUpdate":   m.CreatedOn,
			})

		}
	}
	return res, err
}
