package repo

import (
	"time"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/log"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/repo/dto"
	"go.uber.org/zap"
)

type RepoPendingMessages interface {
	Save(ID int64, senderID int64, message *msg.MessagesSend) (*msg.ClientPendingMessage, error)
	GetPendingMessageByRequestID(requestID int64) (*dto.PendingMessages, error)
	GetPendingMessageByID(id int64) (*dto.PendingMessages, error)
	DeletePendingMessage(ID int64) error
	DeleteManyPendingMessage(IDs []int64) error
	GetManyPendingMessages(messageIDs []int64) []*msg.UserMessage
	GetManyPendingMessagesRequestID(messageIDs []int64) []int64
	GetAllPendingMessages() []*msg.MessagesSend
	DeletePeerAllMessages(peerID int64, peerType int32) (*msg.ClientUpdateMessagesDeleted, error)
	SaveClientMessageMedia(ID, senderID, requestID int64, msgMedia *msg.ClientSendMessageMedia) (*msg.ClientPendingMessage, error)
	GetPendingMessage(messageID int64) *msg.UserMessage
}

type repoPendingMessages struct {
	*repository
}

// Save
func (r *repoPendingMessages) Save(ID int64, senderID int64, message *msg.MessagesSend) (*msg.ClientPendingMessage, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	if message == nil {
		log.LOG_Debug("RepoPendingMessages::Save() :",
			zap.String("Error", "message is null"),
		)
		return nil, domain.ErrNotFound
	}
	log.LOG_Debug("RepoPendingMessages::Save()",
		zap.Int64("PendingMessageID", ID),
	)

	m := new(dto.PendingMessages)
	m.Map(message)
	m.ID = ID
	m.SenderID = senderID
	m.CreatedOn = time.Now().Unix()

	res := new(msg.ClientPendingMessage)
	m.MapTo(res)

	ep := new(dto.PendingMessages)
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

// GetPendingMessageByRequestID
func (r *repoPendingMessages) GetPendingMessageByRequestID(requestID int64) (*dto.PendingMessages, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	log.LOG_Debug("RepoPendingMessages::GetPendingMessageByRequestID()",
		zap.Int64("RandomID/RequestID", requestID),
	)

	pmsg := new(dto.PendingMessages)
	err := r.db.Where("RequestID = ?", requestID).First(pmsg).Error
	if err != nil {
		log.LOG_Debug("RepoPendingMessages::GetPendingMessageByRequestID()-> fetch pendingMessage entity",
			zap.String("Error", err.Error()),
		)
		return nil, err
	}

	return pmsg, nil

}

// GetPendingMessageByID
func (r *repoPendingMessages) GetPendingMessageByID(id int64) (*dto.PendingMessages, error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	log.LOG_Debug("RepoPendingMessages::GetPendingMessageByID()",
		zap.Int64("ID", id),
	)

	pmsg := new(dto.PendingMessages)
	err := r.db.Where("ID = ?", id).First(pmsg).Error
	if err != nil {
		log.LOG_Debug("RepoPendingMessages::GetPendingMessageByID()-> fetch pendingMessage entity",
			zap.String("Error", err.Error()),
		)
		return nil, err
	}

	return pmsg, nil

}

// DeletePendingMessage
func (r *repoPendingMessages) DeletePendingMessage(ID int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	log.LOG_Debug("RepoPendingMessages::DeletePendingMessage()",
		zap.Int64("ID", ID),
	)

	err := r.db.Where("ID = ?", ID).Delete(dto.PendingMessages{}).Error
	if err != nil {
		log.LOG_Debug("RepoPendingMessages::DeletePendingMessage()-> delete pendingMessage entity",
			zap.String("Error", err.Error()),
		)
		return err
	}
	return nil
}

// DeleteManyPendingMessage
func (r *repoPendingMessages) DeleteManyPendingMessage(IDs []int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	// Get dialogs that their top message is going to be removed
	dtoDoalogs := make([]dto.Dialogs, 0)
	err := r.db.Where("TopMessageID in (?)", IDs).Find(&dtoDoalogs).Error
	if err != nil {
		log.LOG_Debug("RepoPendingMessages::DeleteMany() fetch dialogs", zap.Error(err))
	}

	// remove pending message
	err = r.db.Where("ID IN (?)", IDs).Delete(dto.PendingMessages{}).Error
	if err != nil {
		log.LOG_Debug("RepoPendingMessages::DeleteMany() delete from Messages", zap.Error(err))
	}

	// fetch last message and set it as dialog top message
	for _, d := range dtoDoalogs {
		dtoPend := dto.PendingMessages{}
		err := r.db.Table(dtoPend.TableName()).Where("PeerID =? AND PeerType= ?", d.PeerID, d.PeerType).First(&dtoPend).Error // cuz the pendMsg Ids are negative of nano time so the smallest is latest record
		if err == nil && dtoPend.ID != 0 {
			d.TopMessageID = dtoPend.ID
			r.db.Save(d)
		} else {
			dtoMsg := dto.Messages{}
			err := r.db.Table(dtoMsg.TableName()).Where("PeerID =? AND PeerType= ?", d.PeerID, d.PeerType).Last(&dtoMsg).Error
			if err == nil && dtoMsg.ID != 0 {
				d.TopMessageID = dtoMsg.ID
				r.db.Save(d)
			}
		}
	}

	return nil
}

// GetManyPendingMessages
func (r *repoPendingMessages) GetManyPendingMessages(messageIDs []int64) []*msg.UserMessage {
	r.mx.Lock()
	defer r.mx.Unlock()

	log.LOG_Debug("RepoPendingMessages::GetManyPendingMessages()",
		zap.Int64s("MessageIDs", messageIDs),
	)
	messages := make([]*msg.UserMessage, 0, len(messageIDs))
	dtoMsgs := make([]dto.PendingMessages, 0, len(messageIDs))
	err := r.db.Where("ID in (?)", messageIDs).Find(&dtoMsgs).Error
	if err != nil {
		log.LOG_Debug("RepoPendingMessages::GetManyPendingMessages()-> fetch pendingMessage entities",
			zap.String("Error", err.Error()),
		)
		return nil
	}

	for _, v := range dtoMsgs {

		tmp := new(msg.UserMessage)
		v.MapToUserMessage(tmp)
		messages = append(messages, tmp)
	}

	return messages
}

// GetManyPendingMessages
func (r *repoPendingMessages) GetManyPendingMessagesRequestID(messageIDs []int64) []int64 {
	r.mx.Lock()
	defer r.mx.Unlock()

	log.LOG_Debug("RepoPendingMessages::GetManyPendingMessagesRequestID()",
		zap.Int64s("MessageIDs", messageIDs),
	)
	dtoMsgs := make([]dto.PendingMessages, 0, len(messageIDs))
	err := r.db.Where("ID in (?)", messageIDs).Find(&dtoMsgs).Error
	if err != nil {
		log.LOG_Debug("RepoPendingMessages::GetManyPendingMessages()-> fetch pendingMessage entities",
			zap.String("Error", err.Error()),
		)
		return nil
	}
	requestIDs := make([]int64, 0)
	for _, v := range dtoMsgs {
		requestIDs = append(requestIDs, v.RequestID)
	}

	return requestIDs
}
func (r *repoPendingMessages) GetAllPendingMessages() []*msg.MessagesSend {
	r.mx.Lock()
	defer r.mx.Unlock()

	res := make([]*msg.MessagesSend, 0)
	var dtos []dto.PendingMessages
	r.db.Where("MediaType=?", 0).Find(&dtos)

	for _, v := range dtos {
		tmp := new(msg.MessagesSend)
		v.MapToMessageSend(tmp)
		res = append(res, tmp)
	}

	return res
}

func (r *repoPendingMessages) DeletePeerAllMessages(peerID int64, peerType int32) (*msg.ClientUpdateMessagesDeleted, error) {

	msgs := make([]dto.PendingMessages, 0)
	err := r.db.Where("PeerID=? AND PeerType=?", peerID, peerType).Find(&msgs).Error
	if err != nil {
		log.LOG_Debug("RepoPendingMessages::DeletePeerAllMessages()-> fetch pendingMessage entities",
			zap.String("Error", err.Error()),
		)
		return nil, err
	}
	res := new(msg.ClientUpdateMessagesDeleted)
	res.PeerID = peerID
	res.PeerType = peerType
	res.MessageIDs = make([]int64, 0)
	for _, v := range msgs {
		res.MessageIDs = append(res.MessageIDs, v.ID)
	}
	err = r.db.Where("PeerID=? AND PeerType=?", peerID, peerType).Delete(dto.PendingMessages{}).Error
	if err != nil {
		log.LOG_Debug("RepoPendingMessages::DeletePeerAllMessages()-> delete pendingMessage entities",
			zap.String("Error", err.Error()),
		)
	}
	return res, err
}

func (r *repoPendingMessages) SaveClientMessageMedia(ID, senderID, requestID int64, msgMedia *msg.ClientSendMessageMedia) (*msg.ClientPendingMessage, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	if msgMedia == nil {
		log.LOG_Debug("RepoPendingMessages::msgMedia() :",
			zap.String("Error", "msgMedia is null"),
		)
		return nil, domain.ErrNotFound
	}
	log.LOG_Debug("RepoPendingMessages::SaveMedia()",
		zap.Int64("PendingMessageID", ID),
	)

	fileID := domain.SequentialUniqueID()
	m := new(dto.PendingMessages)
	m.MapFromClientMessageMedia(fileID, msgMedia)
	m.ID = ID
	m.SenderID = senderID
	m.CreatedOn = time.Now().Unix()
	m.RequestID = requestID

	res := new(msg.ClientPendingMessage)
	m.MapTo(res)
	ep := new(dto.PendingMessages)
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

// GetPendingMessages
func (r *repoPendingMessages) GetPendingMessage(messageID int64) *msg.UserMessage {
	r.mx.Lock()
	defer r.mx.Unlock()

	log.LOG_Debug("RepoPendingMessages::GetPendingMessages()",
		zap.Int64("MessageID", messageID),
	)
	message := new(msg.UserMessage)
	dtoMsg := new(dto.PendingMessages)
	err := r.db.Where("ID = ?", messageID).Find(&dtoMsg).Error
	if err != nil {
		log.LOG_Debug("RepoPendingMessages::GetPendingMessages()-> fetch messages",
			zap.String("Error", err.Error()),
		)
		return nil
	}

	dtoMsg.MapToUserMessage(message)

	return message
}

func (r *repoPendingMessages) SaveMessageMedia(ID int64, senderID int64, message *msg.MessagesSendMedia) (*msg.ClientPendingMessage, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	if message == nil {
		log.LOG_Debug("RepoPendingMessages::SaveMessageMedia() :",
			zap.String("Error", "message is null"),
		)
		return nil, domain.ErrNotFound
	}
	log.LOG_Debug("RepoPendingMessages::SaveMessageMedia()",
		zap.Int64("PendingMessageID", ID),
	)

	m := new(dto.PendingMessages)
	m.MapFromMessageMedia(message)
	m.ID = ID
	m.SenderID = senderID
	m.CreatedOn = time.Now().Unix()

	res := new(msg.ClientPendingMessage)
	m.MapTo(res)

	ep := new(dto.PendingMessages)
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
