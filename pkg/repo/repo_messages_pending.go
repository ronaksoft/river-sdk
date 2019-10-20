package repo

import (
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/dgraph-io/badger"
	"math"
	"strings"
	"time"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
)

const (
	prefixPMessagesByID       = "PMSG_ID"
	prefixPMessagesByRandomID = "PMSG_RND"
	prefixPMessagesByRealID   = "PMSG_RID"
)

type repoMessagesPending struct {
	*repository
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func getPendingMessageKey(msgID int64) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%012d", prefixPMessagesByID, int64(math.Abs(float64(msgID)))))
}

func getPendingMessageRandomKey(randomID int64) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%012d", prefixPMessagesByRandomID, abs(randomID)))
}

func getPendingMessageRealKey(msgID int64) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%012d", prefixPMessagesByRealID, msgID))
}

func getPendingMessageByID(txn *badger.Txn, msgID int64) (*msg.ClientPendingMessage, error) {
	pm := &msg.ClientPendingMessage{}
	item, err := txn.Get(getPendingMessageKey(msgID))
	if err != nil {
		return nil, err
	}

	err = item.Value(func(val []byte) error {
		return pm.Unmarshal(val)
	})
	if err != nil {
		return nil, err
	}

	return pm, nil
}

func deletePendingMessage(txn *badger.Txn, msgID int64) error {
	pm, err := getPendingMessageByID(txn, msgID)
	if err != nil {
		return err
	}
	err = txn.Delete(getPendingMessageKey(pm.ID))
	if err != nil {
		return err
	}
	err = txn.Delete(getPendingMessageRandomKey(pm.RequestID))
	return err
}

func (r *repoMessagesPending) Save(msgID int64, senderID int64, message *msg.MessagesSend) (*msg.ClientPendingMessage, error) {
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
		CreatedOn: domain.Now().Unix(),
		SenderID:  senderID,
		ID:        msgID,
	}

	bytes, _ := pm.Marshal()
	_ = r.badger.Update(func(txn *badger.Txn) error {
		err := txn.SetEntry(badger.NewEntry(
			getPendingMessageKey(pm.ID), bytes),
		)
		if err != nil {
			return err
		}
		return txn.SetEntry(badger.NewEntry(
			getPendingMessageRandomKey(pm.RequestID), bytes),
		)
	})

	Dialogs.updateLastUpdate(pm.PeerID, pm.PeerType, pm.CreatedOn)

	return pm, nil
}

func (r *repoMessagesPending) SaveClientMessageMedia(msgID, senderID, requestID, fileID, thumbID int64, msgMedia *msg.ClientSendMessageMedia) (*msg.ClientPendingMessage, error) {
	if msgMedia == nil {
		return nil, domain.ErrNotFound
	}
	switch msgMedia.MediaType {
	case msg.InputMediaTypeUploadedDocument, msg.InputMediaTypeUploadedPhoto:
	default:
		panic("Invalid MediaInputType")
	}

	// support IOS file path
	if strings.HasPrefix(msgMedia.FilePath, "file://") {
		msgMedia.FilePath = msgMedia.FilePath[7:]
	}

	msgMedia.FileTotalParts = 0

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
	pm.CreatedOn = domain.Now().Unix()
	pm.RequestID = requestID

	pm.FileUploadID = fmt.Sprintf("%d", fileID)
	pm.FileID = fileID
	if thumbID > 0 {
		pm.ThumbID = thumbID
		pm.ThumbUploadID = fmt.Sprintf("%d", thumbID)
	}

	bytes, _ := pm.Marshal()
	_ = r.badger.Update(func(txn *badger.Txn) error {
		// 1. Save PendingMessage by ID
		err := txn.SetEntry(badger.NewEntry(
			getPendingMessageKey(pm.ID), bytes),
		)
		if err != nil {
			return err
		}

		// 2. Save PendingMessage by RequestID/RandomID
		return txn.SetEntry(badger.NewEntry(
			getPendingMessageRandomKey(pm.RequestID), bytes),
		)
	})

	Dialogs.updateLastUpdate(pm.PeerID, pm.PeerType, pm.CreatedOn)

	return pm, nil
}

func (r *repoMessagesPending) UpdateClientMessageMedia(pm *msg.ClientPendingMessage, totalParts int32) {
	switch pm.MediaType {
	case msg.InputMediaTypeUploadedDocument:
	default:
		panic("invalid input MediaType in pending message for ClientMessageMedia")
	}
	csmm := new(msg.ClientSendMessageMedia)
	_ = csmm.Unmarshal(pm.Media)
	csmm.FileTotalParts = totalParts
	pm.Media, _ = csmm.Marshal()

	bytes, _ := pm.Marshal()
	_ = r.badger.Update(func(txn *badger.Txn) error {
		err := txn.SetEntry(badger.NewEntry(
			getPendingMessageKey(pm.ID), bytes),
		)
		if err != nil {
			return err
		}
		return txn.SetEntry(badger.NewEntry(
			getPendingMessageRandomKey(pm.RequestID), bytes),
		)
	})

}

func (r *repoMessagesPending) SaveMessageMedia(msgID int64, senderID int64, msgMedia *msg.MessagesSendMedia) (*msg.ClientPendingMessage, error) {
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
	pm.Media = msgMedia.MediaData
	pm.ID = msgID
	pm.SenderID = senderID
	pm.CreatedOn = time.Now().Unix()
	pm.RequestID = msgMedia.RandomID

	bytes, _ := pm.Marshal()
	_ = r.badger.Update(func(txn *badger.Txn) error {
		err := txn.SetEntry(badger.NewEntry(
			getPendingMessageKey(pm.ID), bytes),
		)
		if err != nil {
			return err
		}
		return txn.SetEntry(badger.NewEntry(
			getPendingMessageRandomKey(pm.RequestID), bytes),
		)
	})

	Dialogs.updateLastUpdate(pm.PeerID, pm.PeerType, pm.CreatedOn)

	return pm, nil
}

func (r *repoMessagesPending) GetByRealID(msgID int64) *msg.ClientPendingMessage {
	pm := new(msg.ClientPendingMessage)
	err := r.badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get(getPendingMessageRealKey(msgID))
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

func (r *repoMessagesPending) GetByRandomID(randomID int64) (*msg.ClientPendingMessage, error) {
	pm := new(msg.ClientPendingMessage)
	err := r.badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get(getPendingMessageRandomKey(randomID))
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

func (r *repoMessagesPending) GetByID(msgID int64) (pm *msg.ClientPendingMessage) {
	err := r.badger.View(func(txn *badger.Txn) error {
		var err error
		pm, err = getPendingMessageByID(txn, msgID)
		return err
	})
	logs.ErrorOnErr("RepoPending got error on get by id", err)
	return pm
}

func (r *repoMessagesPending) GetMany(messageIDs []int64) []*msg.UserMessage {
	userMessages := make([]*msg.UserMessage, 0, len(messageIDs))
	_ = r.badger.View(func(txn *badger.Txn) error {
		for _, msgID := range messageIDs {
			pm, _ := getPendingMessageByID(txn, msgID)
			if pm != nil {
				userMessages = append(userMessages, r.ToUserMessage(pm))
			}
		}
		return nil
	})
	return userMessages
}

func (r *repoMessagesPending) GetByPeer(peerID int64, peerType int32) []*msg.UserMessage {
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
		it.Close()
		return nil
	})
	return userMessages
}

func (r *repoMessagesPending) GetAndConvertAll() []*msg.UserMessage {
	userMessages := make([]*msg.UserMessage, 0, 10)
	_ = r.badger.Update(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		opt.Prefix = ronak.StrToByte(fmt.Sprintf("%s.", prefixPMessagesByID))
		it := txn.NewIterator(opt)
		for it.Rewind(); it.Valid(); it.Next() {
			_ = it.Item().Value(func(val []byte) error {
				pm := new(msg.ClientPendingMessage)
				_ = pm.Unmarshal(val)
				userMessages = append(userMessages, r.ToUserMessage(pm))
				return nil
			})
		}
		it.Close()
		return nil
	})
	return userMessages
}

func (r *repoMessagesPending) GetAll() []*msg.ClientPendingMessage {
	pendingMessages := make([]*msg.ClientPendingMessage, 0, 10)
	_ = r.badger.Update(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		opt.Prefix = ronak.StrToByte(fmt.Sprintf("%s.", prefixPMessagesByID))
		it := txn.NewIterator(opt)
		for it.Rewind(); it.Valid(); it.Next() {
			_ = it.Item().Value(func(val []byte) error {
				pm := new(msg.ClientPendingMessage)
				_ = pm.Unmarshal(val)
				pendingMessages = append(pendingMessages, pm)
				return nil
			})
		}
		it.Close()
		return nil
	})
	return pendingMessages
}

func (r *repoMessagesPending) Delete(msgID int64) error {
	return r.badger.Update(func(txn *badger.Txn) error {
		return deletePendingMessage(txn, msgID)
	})
}

func (r *repoMessagesPending) DeleteByRealID(msgID int64) {
	_ = r.badger.Update(func(txn *badger.Txn) error {
		_ = txn.Delete(getPendingMessageRealKey(msgID))
		return nil
	})
}

func (r *repoMessagesPending) DeleteMany(msgIDs []int64) {
	err := r.badger.Update(func(txn *badger.Txn) error {
		for _, msgID := range msgIDs {
			err  :=  deletePendingMessage(txn, msgID)
			if err != nil {
				return err
			}
		}
		return nil
	})
	logs.ErrorOnErr("RepoPending got error on delete many", err)
}

func (r *repoMessagesPending) GetManyRequestIDs(msgIDs []int64) []int64 {
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
		it.Close()
		return nil
	})

	return res
}

func (r *repoMessagesPending) SaveByRealID(randomID, realMsgID int64) {
	// TODO:: Remove old keys
	pm, err := r.GetByRandomID(randomID)
	if err != nil {
		return
	}
	bytes, _ := pm.Marshal()
	_ = r.badger.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry(
			getPendingMessageRealKey(realMsgID), bytes),
		)
	})
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
	switch m.MediaType {
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

func (r *repoMessagesPending) ToMessagesSend(m *msg.ClientPendingMessage) *msg.MessagesSend {
	v := new(msg.MessagesSend)
	v.RandomID = m.RequestID
	v.Body = m.Body
	v.Peer = &msg.InputPeer{
		ID:         m.PeerID,
		Type:       msg.PeerType(m.PeerType),
		AccessHash: m.AccessHash,
	}
	v.ClearDraft = m.ClearDraft
	v.Entities = m.Entities
	v.ReplyTo = m.ReplyTo
	return v
}

func (r *repoMessagesPending) ToMessagesSendMedia(m *msg.ClientPendingMessage) *msg.MessagesSendMedia {
	v := new(msg.MessagesSendMedia)
	v.RandomID = m.RequestID
	v.Peer = &msg.InputPeer{
		ID:         m.PeerID,
		Type:       msg.PeerType(m.PeerType),
		AccessHash: m.AccessHash,
	}
	v.ClearDraft = m.ClearDraft
	v.ReplyTo = m.ReplyTo
	v.MediaType = m.MediaType

	switch m.MediaType {
	case msg.InputMediaTypeUploadedDocument:
		csmm := new(msg.ClientSendMessageMedia)
		_ = csmm.Unmarshal(m.Media)
		uploadedDocument := new(msg.InputMediaUploadedDocument)

		if csmm.FileTotalParts == 0 {
			// If FileTotalParts is still zero it means we have not upload the document yet
			return nil
		}
		uploadedDocument.File = &msg.InputFile{
			FileID:      csmm.FileID,
			TotalParts:  csmm.FileTotalParts,
			FileName:    csmm.FileName,
			MD5Checksum: "",
		}
		if csmm.ThumbID != 0 {
			uploadedDocument.Thumbnail = &msg.InputFile{
				FileID:      csmm.ThumbID,
				TotalParts:  0,
				FileName:    "",
				MD5Checksum: "",
			}
		}
		uploadedDocument.MimeType = csmm.FileMIME
		uploadedDocument.Attributes = csmm.Attributes
		uploadedDocument.Caption = csmm.Caption
		v.MediaData, _ = uploadedDocument.Marshal()
	default:
		v.MediaData = m.Media
	}

	return v
}
