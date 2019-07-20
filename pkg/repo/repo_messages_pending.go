package repo

import (
	"fmt"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/dgraph-io/badger"
	"time"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
)

const (
	prefixPMessagesByID       = "PMSG_ID"
	prefixPMessagesByRandomID = "PMSG_RND"
)

type repoMessagesPending struct {
	*repository
}

func (r *repoMessagesPending) getKey(msgID int64) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%012d", prefixPMessagesByID, msgID))
}

func (r *repoMessagesPending) getRandomKey(randomID int64) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%012d", prefixPMessagesByRandomID, randomID))
}

func (r *repoMessagesPending) Save(msgID int64, senderID int64, message *msg.MessagesSend) (*msg.ClientPendingMessage, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	if message == nil {
		return nil, domain.ErrNotFound
	}

	pm := &msg.ClientPendingMessage{
		AccessHash: message.Peer.AccessHash,
		Body:       message.Body,
		PeerID:     message.Peer.ID,
		PeerType:   int32(message.Peer.Type),
		ReplyTo:    message.ReplyTo,
		RequestID:  message.RandomID,
		Entities:   message.Entities,
		ClearDraft: message.ClearDraft,
		// Filled by SDK
		CreatedOn: time.Now().Unix(),
		SenderID:  senderID,
		ID:        msgID,
	}

	bytes, _ := pm.Marshal()
	_ = r.badger.Update(func(txn *badger.Txn) error {
		err := txn.SetEntry(badger.NewEntry(
			r.getKey(pm.ID), bytes),
		)
		if err != nil {
			return err
		}
		return txn.SetEntry(badger.NewEntry(
			r.getRandomKey(pm.RequestID), bytes),
		)
	})

	_ = Dialogs.updateLastUpdate(pm.PeerID, pm.PeerType, pm.CreatedOn)

	return pm, nil
}

func (r *repoMessagesPending) SaveClientMessageMedia(msgID, senderID, randomID int64, msgMedia *msg.ClientSendMessageMedia) (*msg.ClientPendingMessage, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	if msgMedia == nil {
		return nil, domain.ErrNotFound
	}

	pm := new(msg.ClientPendingMessage)
	pm.PeerID = msgMedia.Peer.ID
	pm.PeerType = int32(msgMedia.Peer.Type)
	pm.AccessHash = msgMedia.Peer.AccessHash
	pm.Body = msgMedia.Caption
	pm.ReplyTo = msgMedia.ReplyTo
	pm.ClearDraft = msgMedia.ClearDraft
	pm.MediaType = msgMedia.MediaType
	pm.Media, _ = msgMedia.Marshal()
	pm.ID = msgID
	pm.SenderID = senderID
	pm.CreatedOn = time.Now().Unix()
	pm.RequestID = randomID

	bytes, _ := pm.Marshal()
	_ = r.badger.Update(func(txn *badger.Txn) error {
		err := txn.SetEntry(badger.NewEntry(
			r.getKey(pm.ID), bytes),
		)
		if err != nil {
			return err
		}
		return txn.SetEntry(badger.NewEntry(
			r.getRandomKey(pm.RequestID), bytes),
		)
	})

	_ = Dialogs.updateLastUpdate(pm.PeerID, pm.PeerType, pm.CreatedOn)

	return pm, nil
}

func (r *repoMessagesPending) SaveMessageMedia(msgID int64, senderID int64, msgMedia *msg.MessagesSendMedia) (*msg.ClientPendingMessage, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	if msgMedia == nil {
		return nil, domain.ErrNotFound
	}

	pm := new(msg.ClientPendingMessage)
	pm.PeerID = msgMedia.Peer.ID
	pm.PeerType = int32(msgMedia.Peer.Type)
	pm.AccessHash = msgMedia.Peer.AccessHash
	pm.ReplyTo = msgMedia.ReplyTo
	pm.ClearDraft = msgMedia.ClearDraft
	pm.MediaType = msgMedia.MediaType
	pm.Media, _ = msgMedia.Marshal()
	pm.ID = msgID
	pm.SenderID = senderID
	pm.CreatedOn = time.Now().Unix()
	pm.RequestID = msgMedia.RandomID

	bytes, _ := pm.Marshal()
	_ = r.badger.Update(func(txn *badger.Txn) error {
		err := txn.SetEntry(badger.NewEntry(
			r.getKey(pm.ID), bytes),
		)
		if err != nil {
			return err
		}
		return txn.SetEntry(badger.NewEntry(
			r.getRandomKey(pm.RequestID), bytes),
		)
	})

	_ = Dialogs.updateLastUpdate(pm.PeerID, pm.PeerType, pm.CreatedOn)

	return pm, nil
}

func (r *repoMessagesPending) GetByRandomID(randomID int64) (*msg.ClientPendingMessage, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	pm := new(msg.ClientPendingMessage)
	err := r.badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get(r.getRandomKey(randomID))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return pm.Unmarshal(val)
		})
	})
	if err != nil {
		return nil, err
	}

	return pm, nil
}

func (r *repoMessagesPending) GetByID(id int64) *msg.ClientPendingMessage {
	r.mx.Lock()
	defer r.mx.Unlock()
	pm := new(msg.ClientPendingMessage)
	err := r.badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get(r.getKey(id))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return pm.Unmarshal(val)
		})
	})
	if err != nil {
		return nil
	}

	return pm

}

func (r *repoMessagesPending) GetMany(messageIDs []int64) []*msg.UserMessage {
	r.mx.Lock()
	defer r.mx.Unlock()

	userMessages := make([]*msg.UserMessage, 0, len(messageIDs))
	for _, msgID := range messageIDs {
		pm := r.GetByID(msgID)
		if pm != nil {
			userMessages = append(userMessages, r.ToUserMessage(pm))
		}
	}
	return userMessages
}

func (r *repoMessagesPending) GetByPeer(peerID int64, peerType int32) []*msg.UserMessage {
	r.mx.Lock()
	defer r.mx.Unlock()

	userMessages := make([]*msg.UserMessage, 0, 10)
	_ = r.badger.Update(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		opt.Prefix = ronak.StrToByte(fmt.Sprintf("%s.", prefixPMessagesByID))
		it := txn.NewIterator(opt)
		for it.Rewind(); it.Valid(); it.Next() {
			_ = it.Item().Value(func(val []byte) error {
				pm := new(msg.ClientPendingMessage)
				_ = pm.Unmarshal(val)
				if pm.PeerID == peerID && pm.PeerType == peerType {
					userMessages = append(userMessages, r.ToUserMessage(pm))
				}
				return nil
			})
		}
		return nil
	})
	return userMessages
}


func (r *repoMessagesPending) Delete(msgID int64) {
	r.mx.Lock()
	defer r.mx.Unlock()

	pm := r.GetByID(msgID)
	if pm == nil {
		return
	}

	_ = r.badger.Update(func(txn *badger.Txn) error {
		_ = txn.Delete(r.getKey(pm.ID))
		_ = txn.Delete(r.getRandomKey(pm.RequestID))
		return nil
	})
}

func (r *repoMessagesPending) DeleteMany(msgIDs []int64) {
	r.mx.Lock()
	defer r.mx.Unlock()

	for _, msgID := range msgIDs {
		r.Delete(msgID)
	}
}

func (r *repoMessagesPending) GetManyRequestIDs(msgIDs []int64) []int64 {
	r.mx.Lock()
	defer r.mx.Unlock()

	requestIDs := make([]int64, 0, len(msgIDs))
	for _, msgID := range msgIDs {
		pm := r.GetByID(msgID)
		if pm == nil {
			continue
		}
		requestIDs = append(requestIDs, pm.RequestID)
	}
	return requestIDs
}

func (r *repoMessagesPending) DeletePeerAllMessages(peerID int64, peerType int32) *msg.ClientUpdateMessagesDeleted {

	r.mx.Lock()
	defer r.mx.Unlock()

	res := new(msg.ClientUpdateMessagesDeleted)
	res.PeerID = peerID
	res.PeerType = peerType
	res.MessageIDs = make([]int64, 0)
	_ = r.badger.Update(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		opt.Prefix = ronak.StrToByte(fmt.Sprintf("%s.", prefixPMessagesByID))
		it := txn.NewIterator(opt)
		for it.Rewind(); it.Valid(); it.Next() {
			_ = it.Item().Value(func(val []byte) error {
				pm := new(msg.ClientPendingMessage)
				_ = pm.Unmarshal(val)
				if pm.PeerID == peerID && pm.PeerType == peerType {
					res.MessageIDs = append(res.MessageIDs, pm.ID)
					r.Delete(pm.ID)
				}
				return nil
			})
		}
		return nil
	})

	return res
}


const (
	_ClientSendMessageMediaType       = -1
	_ClientSendMessageContactType     = -2
	_ClientSendMessageGeoLocationType = -3
)

func (r *repoMessagesPending) ToUserMessage(m *msg.ClientPendingMessage) *msg.UserMessage {
	v := new(msg.UserMessage)
	v.ID = m.ID
	v.PeerID = m.PeerID
	v.PeerType = m.PeerType
	v.CreatedOn = m.CreatedOn
	v.Body = m.Body
	v.SenderID = m.SenderID
	v.ReplyTo = m.ReplyTo
	v.Entities = m.Entities
	switch msg.InputMediaType(m.MediaType) {
	case msg.InputMediaTypeUploadedDocument:
		v.MessageType = _ClientSendMessageMediaType
	case msg.InputMediaTypeContact:
		v.MessageType = _ClientSendMessageContactType
	case msg.InputMediaTypeGeoLocation:
		v.MessageType = _ClientSendMessageGeoLocationType
	}
	v.MediaType = msg.MediaType(m.MediaType)
	v.Media = m.Media
	return v
}
