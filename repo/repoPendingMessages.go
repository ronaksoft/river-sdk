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
	GetManyPendingMessages(messageIDs []int64) []*msg.UserMessage
	GetAllPendingMessages() []*msg.MessagesSend
	DeletePeerAllMessages(peerID int64, peerType int32) (*msg.ClientUpdateMessagesDeleted, error)
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

func (r *repoPendingMessages) GetAllPendingMessages() []*msg.MessagesSend {
	r.mx.Lock()
	defer r.mx.Unlock()

	res := make([]*msg.MessagesSend, 0)
	var dtos []dto.PendingMessages
	r.db.Find(&dtos)

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
