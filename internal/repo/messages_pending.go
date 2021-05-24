package repo

import (
	"fmt"
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/z"
	"github.com/dgraph-io/badger/v2"
	"github.com/ronaksoft/rony/pools"
	"github.com/ronaksoft/rony/tools"
	"time"
)

const (
	prefixPMessagesByID       = "PMSG_ID"
	prefixPMessagesByRandomID = "PMSG_RND"
	prefixPMessagesByRealID   = "PMSG_RID"
)

type repoMessagesPending struct {
	*repository
}

func getPendingMessageKey(msgID int64) []byte {
	sb := pools.AcquireStringsBuilder()
	sb.WriteString(prefixPMessagesByID)
	sb.WriteRune('.')
	z.AppendStrInt64(sb, z.AbsInt64(msgID))
	id := tools.StrToByte(sb.String())
	pools.ReleaseStringsBuilder(sb)
	return id
}

func getPendingMessageRandomKey(randomID int64) []byte {
	sb := pools.AcquireStringsBuilder()
	sb.WriteString(prefixPMessagesByRandomID)
	sb.WriteRune('.')
	z.AppendStrInt64(sb, z.AbsInt64(randomID))
	id := tools.StrToByte(sb.String())
	pools.ReleaseStringsBuilder(sb)
	return id
}

func getPendingMessageRealKey(msgID int64) []byte {
	sb := pools.AcquireStringsBuilder()
	sb.WriteString(prefixPMessagesByRealID)
	sb.WriteRune('.')
	z.AppendStrInt64(sb, msgID)
	id := tools.StrToByte(sb.String())
	pools.ReleaseStringsBuilder(sb)
	return id
}

func getPendingMessageByID(txn *badger.Txn, msgID int64) (*msg.ClientPendingMessage, error) {
	pm := &msg.ClientPendingMessage{}
	item, err := txn.Get(getPendingMessageKey(msgID))
	switch err {
	case nil:
	case badger.ErrKeyNotFound:
		return nil, domain.ErrNotFound
	default:
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

func (r *repoMessagesPending) Save(
	teamID int64, teamAccess uint64, msgID int64, senderID int64, message *msg.MessagesSend,
) (*msg.ClientPendingMessage, error) {
	if message == nil {
		return nil, domain.ErrNotFound
	}
	pm := &msg.ClientPendingMessage{
		TeamID:         teamID,
		TeamAccessHash: teamAccess,
		AccessHash:     message.Peer.AccessHash,
		Body:           message.Body,
		PeerID:         message.Peer.ID,
		PeerType:       int32(message.Peer.Type),
		ReplyTo:        message.ReplyTo,
		RequestID:      message.RandomID,
		Entities:       message.Entities,
		ClearDraft:     message.ClearDraft,
		// Filled by SDK
		CreatedOn: domain.Now().Unix(),
		SenderID:  senderID,
		ID:        msgID,
	}

	bytes, _ := pm.Marshal()
	_ = badgerUpdate(func(txn *badger.Txn) error {
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

	_ = updateDialogLastUpdate(pm.TeamID, pm.PeerID, pm.PeerType, pm.CreatedOn)

	return pm, nil
}

func (r *repoMessagesPending) SaveClientMessageMedia(
	teamID int64, teamAccess uint64,
	msgID, senderID, requestID, fileID, thumbID int64, msgMedia *msg.ClientSendMessageMedia, fileSha256 []byte,
) (*msg.ClientPendingMessage, error) {
	if msgMedia == nil {
		return nil, domain.ErrNotFound
	}

	msgMedia.FileTotalParts = 0

	pm := &msg.ClientPendingMessage{
		ID:             msgID,
		TeamID:         teamID,
		TeamAccessHash: teamAccess,
		PeerID:         msgMedia.Peer.ID,
		PeerType:       int32(msgMedia.Peer.Type),
		AccessHash:     msgMedia.Peer.AccessHash,
		Body:           msgMedia.Caption,
		ReplyTo:        msgMedia.ReplyTo,
		ClearDraft:     msgMedia.ClearDraft,
		MediaType:      msgMedia.MediaType,
		SenderID:       senderID,
		CreatedOn:      domain.Now().Unix(),
		RequestID:      requestID,
		FileUploadID:   fmt.Sprintf("%d", fileID),
		FileID:         fileID,
		Sha256:         fileSha256,
		TinyThumb:      msgMedia.TinyThumb,
	}
	pm.Media, _ = msgMedia.Marshal()

	if thumbID > 0 {
		pm.ThumbID = thumbID
		pm.ThumbUploadID = fmt.Sprintf("%d", thumbID)
	}

	bytes, _ := pm.Marshal()
	err := badgerUpdate(func(txn *badger.Txn) error {
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

	if err != nil {
		return nil, err
	}

	_ = updateDialogLastUpdate(pm.TeamID, pm.PeerID, pm.PeerType, pm.CreatedOn)

	return pm, nil
}

func (r *repoMessagesPending) UpdateClientMessageMedia(pm *msg.ClientPendingMessage, totalParts int32, mediaType msg.InputMediaType, fileLocation *msg.FileLocation) error {
	csmm := new(msg.ClientSendMessageMedia)
	_ = csmm.Unmarshal(pm.Media)
	csmm.FileTotalParts = totalParts
	pm.Media, _ = csmm.Marshal()
	pm.ServerFile = fileLocation
	pm.MediaType = mediaType

	bytes, _ := pm.Marshal()
	return badgerUpdate(func(txn *badger.Txn) error {
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

func (r *repoMessagesPending) SaveMessageMedia(
	teamID int64, teamAccess uint64, msgID int64, senderID int64, msgMedia *msg.MessagesSendMedia,
) (*msg.ClientPendingMessage, error) {
	if msgMedia == nil {
		return nil, domain.ErrNotFound
	}

	pm := &msg.ClientPendingMessage{
		PeerID:         msgMedia.Peer.ID,
		PeerType:       int32(msgMedia.Peer.Type),
		AccessHash:     msgMedia.Peer.AccessHash,
		ReplyTo:        msgMedia.ReplyTo,
		ClearDraft:     msgMedia.ClearDraft,
		MediaType:      msgMedia.MediaType,
		Media:          msgMedia.MediaData,
		ID:             msgID,
		TeamID:         teamID,
		TeamAccessHash: teamAccess,
		SenderID:       senderID,
		CreatedOn:      time.Now().Unix(),
		RequestID:      msgMedia.RandomID,
	}

	bytes, _ := pm.Marshal()
	_ = badgerUpdate(func(txn *badger.Txn) error {
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

	_ = updateDialogLastUpdate(pm.TeamID, pm.PeerID, pm.PeerType, pm.CreatedOn)

	return pm, nil
}

func (r *repoMessagesPending) GetByRealID(msgID int64) *msg.ClientPendingMessage {
	pm := new(msg.ClientPendingMessage)
	err := badgerView(func(txn *badger.Txn) error {
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
	err := badgerView(func(txn *badger.Txn) error {
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

func (r *repoMessagesPending) GetByID(msgID int64) (pm *msg.ClientPendingMessage, err error) {
	err = badgerView(func(txn *badger.Txn) error {
		pm, err = getPendingMessageByID(txn, msgID)
		return err
	})
	return
}

func (r *repoMessagesPending) GetMany(messageIDs []int64) []*msg.UserMessage {
	userMessages := make([]*msg.UserMessage, 0, len(messageIDs))
	_ = badgerView(func(txn *badger.Txn) error {
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

func (r *repoMessagesPending) GetByPeer(teamID int64, peerID int64, peerType int32) []*msg.UserMessage {
	userMessages := make([]*msg.UserMessage, 0, 10)
	_ = badgerUpdate(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		opt.Prefix = tools.StrToByte(fmt.Sprintf("%s.", prefixPMessagesByID))
		it := txn.NewIterator(opt)
		for it.Rewind(); it.Valid(); it.Next() {
			_ = it.Item().Value(func(val []byte) error {
				pm := new(msg.ClientPendingMessage)
				_ = pm.Unmarshal(val)
				if pm.PeerID == peerID && pm.PeerType == peerType && pm.TeamID == teamID {
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
	_ = badgerUpdate(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		opt.Prefix = tools.StrToByte(fmt.Sprintf("%s.", prefixPMessagesByID))
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
	_ = badgerUpdate(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		opt.Prefix = tools.StrToByte(fmt.Sprintf("%s.", prefixPMessagesByID))
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
	return badgerUpdate(func(txn *badger.Txn) error {
		return deletePendingMessage(txn, msgID)
	})
}

func (r *repoMessagesPending) DeleteByRealID(msgID int64) {
	_ = badgerUpdate(func(txn *badger.Txn) error {
		_ = txn.Delete(getPendingMessageRealKey(msgID))
		return nil
	})

}

func (r *repoMessagesPending) DeleteMany(msgIDs []int64) {
	_ = badgerUpdate(func(txn *badger.Txn) error {
		for _, msgID := range msgIDs {
			_ = deletePendingMessage(txn, msgID)
		}
		return nil
	})

}

func (r *repoMessagesPending) GetManyRequestIDs(msgIDs []int64) []int64 {
	requestIDs := make([]int64, 0, len(msgIDs))
	for _, msgID := range msgIDs {
		pm, _ := r.GetByID(msgID)
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
	_ = badgerUpdate(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		opt.Prefix = tools.StrToByte(fmt.Sprintf("%s.", prefixPMessagesByID))
		it := txn.NewIterator(opt)
		for it.Rewind(); it.Valid(); it.Next() {
			_ = it.Item().Value(func(val []byte) error {
				pm := new(msg.ClientPendingMessage)
				_ = pm.Unmarshal(val)
				if pm.PeerID == peerID && pm.PeerType == peerType {
					res.MessageIDs = append(res.MessageIDs, pm.ID)
					_ = r.Delete(pm.ID)
				}
				return nil
			})
		}
		it.Close()
		return nil
	})

	return res
}

func (r *repoMessagesPending) SaveByRealID(randomID, realMsgID int64) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		pm := new(msg.ClientPendingMessage)
		item, err := txn.Get(getPendingMessageRandomKey(randomID))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			return pm.Unmarshal(val)
		})
		if err != nil {
			return err
		}
		bytes, _ := pm.Marshal()
		return txn.SetEntry(badger.NewEntry(getPendingMessageRealKey(realMsgID), bytes))
	})

}

const (
	_ClientSendMessageMediaType        = -1
	_ClientSendMessageContactType      = -2
	_ClientSendMessageGeoLocationType  = -3
	_ClientSendMessageInputMessageType = -4
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
	case msg.InputMediaType_InputMediaTypeUploadedDocument:
		v.MessageType = _ClientSendMessageMediaType
	case msg.InputMediaType_InputMediaTypeContact:
		v.MessageType = _ClientSendMessageContactType
	case msg.InputMediaType_InputMediaTypeGeoLocation:
		v.MessageType = _ClientSendMessageGeoLocationType
	case msg.InputMediaType_InputMediaTypeMessageDocument:
		v.MessageType = _ClientSendMessageInputMessageType
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
	v := &msg.MessagesSendMedia{
		RandomID: m.RequestID,
		Peer: &msg.InputPeer{
			ID:         m.PeerID,
			Type:       msg.PeerType(m.PeerType),
			AccessHash: m.AccessHash,
		},
		ClearDraft: m.ClearDraft,
		ReplyTo:    m.ReplyTo,
		MediaType:  m.MediaType,
	}

	switch m.MediaType {
	case msg.InputMediaType_InputMediaTypeUploadedDocument:
		csmm := new(msg.ClientSendMessageMedia)
		_ = csmm.Unmarshal(m.Media)

		if csmm.FileTotalParts == 0 {
			// If FileTotalParts is still zero it means we have not upload the document yet
			return nil
		}

		uploadedDocument := &msg.InputMediaUploadedDocument{
			File: &msg.InputFile{
				FileID:      csmm.FileID,
				TotalParts:  csmm.FileTotalParts,
				FileName:    csmm.FileName,
				MD5Checksum: "",
			},
			Thumbnail:  nil,
			MimeType:   csmm.FileMIME,
			Caption:    csmm.Caption,
			Stickers:   nil,
			Attributes: csmm.Attributes,
			Entities:   nil,
		}

		if csmm.ThumbID != 0 {
			uploadedDocument.Thumbnail = &msg.InputFile{
				FileID:      csmm.ThumbID,
				TotalParts:  0,
				FileName:    "",
				MD5Checksum: "",
			}
		}
		v.MediaData, _ = uploadedDocument.Marshal()
	case msg.InputMediaType_InputMediaTypeDocument:
		csmm := new(msg.ClientSendMessageMedia)
		_ = csmm.Unmarshal(m.Media)
		doc := &msg.InputMediaDocument{
			Caption:    csmm.Caption,
			Attributes: csmm.Attributes,
			Document: &msg.InputDocument{
				ID:         m.ServerFile.FileID,
				AccessHash: m.ServerFile.AccessHash,
				ClusterID:  m.ServerFile.ClusterID,
			},
		}
		if csmm.ThumbID != 0 {
			doc.Thumbnail = &msg.InputFile{
				FileID:      csmm.ThumbID,
				TotalParts:  0,
				FileName:    "",
				MD5Checksum: "",
			}
		}
		v.MediaData, _ = doc.Marshal()
	default:
		v.MediaData = m.Media
	}

	return v
}
