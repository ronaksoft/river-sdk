package repo

import (
	"fmt"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
	"github.com/dgraph-io/badger"
	"go.uber.org/zap"
	"sort"
	"strings"
	"time"
)

const (
	prefixMessages     = "MSG"
	prefixUserMessages = "UMSG"
)

type repoMessages struct {
	*repository
}

func (r *repoMessages) getMessageKey(peerID int64, peerType int32, msgID int64) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%021d.%d.%012d", prefixMessages, peerID, peerType, msgID))
}

func (r *repoMessages) getPrefix(peerID int64, peerType int32) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%021d.%d.", prefixMessages, peerID, peerType))
}

func (r *repoMessages) getUserMessageKey(msgID int64) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%012d", prefixUserMessages, msgID))
}

func (r *repoMessages) getUserMessage(msgID int64) (*msg.UserMessage, error) {
	message := new(msg.UserMessage)
	err := r.badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get(r.getUserMessageKey(msgID))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			parts := strings.Split(ronak.ByteToStr(val), ".")
			if len(parts) != 2 {
				return domain.ErrInvalidUserMessageKey
			}
			itemMessage, err := txn.Get(r.getMessageKey(ronak.StrToInt64(parts[0]), ronak.StrToInt32(parts[1]), msgID))
			if err != nil {
				return err
			}
			return itemMessage.Value(func(val []byte) error {
				return message.Unmarshal(val)
			})
		})
	})
	if err != nil {
		logs.Warn("Error On getUserMessage", zap.Int64("MsgID", msgID))
		return nil, err
	}
	return message, nil
}

func (r *repoMessages) getByKey(msgKey []byte) *msg.UserMessage {
	message := new(msg.UserMessage)
	err := r.badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get(msgKey)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return message.Unmarshal(val)
		})
	})
	if err != nil {
		return nil
	}
	return message
}

func (r *repoMessages) save(message *msg.UserMessage) {
	if message == nil {
		return
	}

	messageBytes, _ := message.Marshal()
	docType := domain.SharedMediaTypeAll
	switch message.MediaType {
	case msg.MediaTypeDocument:
		doc := new(msg.MediaDocument)
		_ = doc.Unmarshal(message.Media)
		if doc.Doc == nil {
			logs.Error("Error On RepoMessages Save, Document is Nil",
				zap.Int64("MessageID", message.ID),
				zap.Int64("SenderID", message.SenderID),
			)
			return
		}
		for _, da := range doc.Doc.Attributes {
			switch da.Type {
			case msg.AttributeTypeAudio:
				a := new(msg.DocumentAttributeAudio)
				_ = a.Unmarshal(da.Data)
				if a.Voice {
					docType = domain.SharedMediaTypeVoice
				} else {
					docType = domain.SharedMediaTypeAudio
				}
			case msg.AttributeTypeVideo, msg.AttributeTypePhoto:
				docType = domain.SharedMediaTypeMedia
			case msg.AttributeTypeFile:
				if docType == domain.SharedMediaTypeAll {
					docType = domain.SharedMediaTypeFile
				}
			}
		}
	default:
		// Do nothing
	}

	err := ronak.Try(10, 100*time.Millisecond, func() error {
		err := r.badger.Update(func(txn *badger.Txn) error {
			// 1. Write Message
			err := txn.SetEntry(
				badger.NewEntry(
					r.getMessageKey(message.PeerID, message.PeerType, message.ID),
					messageBytes,
				).WithMeta(byte(docType)),
			)
			if err != nil {
				return err
			}

			// 2. WriteUserMessage
			return txn.SetEntry(
				badger.NewEntry(
					r.getUserMessageKey(message.ID),
					ronak.StrToByte(fmt.Sprintf("%d.%d", message.PeerID, message.PeerType)),
				).WithMeta(byte(docType)),
			)
		})
		WarnOnErr("RepoMessage::save", err, zap.Int64("MsgID", message.ID))
		return err
	})
	if err != nil {
		logs.Error("Error In SaveMessage")
	}

	_ = r.msgSearch.Index(ronak.ByteToStr(r.getMessageKey(message.PeerID, message.PeerType, message.ID)), MessageSearch{
		Type:   "msg",
		Body:   message.Body,
		PeerID: message.PeerID,
	})

	return
}

func (r *repoMessages) Get(messageID int64) *msg.UserMessage {
	userMessage, err := r.getUserMessage(messageID)
	if err != nil {
		return nil
	}
	return userMessage
}

func (r *repoMessages) GetMany(messageIDs []int64) []*msg.UserMessage {

	userMessages := make([]*msg.UserMessage, 0, len(messageIDs))
	for _, messageID := range messageIDs {
		userMessage, err := r.getUserMessage(messageID)
		if err == nil {
			userMessages = append(userMessages, userMessage)
		}

	}
	return userMessages
}

func (r *repoMessages) SaveNew(message *msg.UserMessage, dialog *msg.Dialog, userID int64) {
	if message == nil {
		return
	}

	r.save(message)

	dialog = Dialogs.Get(message.PeerID, message.PeerType)
	// Update Dialog if it is a new message
	if message.ID > dialog.TopMessageID {
		dialog.TopMessageID = message.ID
		if !dialog.Pinned {
			Dialogs.updateLastUpdate(message.PeerID, message.PeerType, message.CreatedOn)
		}
		// Update counters if necessary
		if message.SenderID != userID {
			dialog.UnreadCount += 1
			for _, entity := range message.Entities {
				if entity.Type == msg.MessageEntityTypeMention && entity.UserID == userID {
					dialog.MentionedCount += 1
				}
			}
		}
	}
	Dialogs.Save(dialog)

	// This is just to make sure the data has been written
	_ = r.badger.Sync()
	return
}

func (r *repoMessages) Save(message *msg.UserMessage) {
	r.save(message)

	// This is just to make sure the data has been written
	_ = r.badger.Sync()

}

func (r *repoMessages) SaveMany(messages []*msg.UserMessage) {
	for _, message := range messages {
		r.save(message)
	}
	// This is just to make sure the data has been written
	_ = r.badger.Sync()
}

func (r *repoMessages) GetMessageHistory(peerID int64, peerType int32, minID, maxID int64, limit int32) (userMessages []*msg.UserMessage, users []*msg.User) {
	userMessages = make([]*msg.UserMessage, 0, limit)
	userIDs := domain.MInt64B{}
	switch {
	case maxID == 0 && minID == 0:
		maxID = Dialogs.Get(peerID, peerType).TopMessageID
		fallthrough
	case maxID != 0 && minID == 0:
		_ = r.badger.View(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.Prefix = r.getPrefix(peerID, peerType)
			opts.Reverse = true
			it := txn.NewIterator(opts)
			for it.Seek(r.getMessageKey(peerID, peerType, maxID)); it.ValidForPrefix(opts.Prefix); it.Next() {
				if limit--; limit < 0 {
					break
				}
				_ = it.Item().Value(func(val []byte) error {
					userMessage := new(msg.UserMessage)
					err := userMessage.Unmarshal(val)
					if err != nil {
						return err
					}
					userIDs[userMessage.SenderID] = true
					if userMessage.FwdSenderID != 0 {
						userIDs[userMessage.FwdSenderID] = true
					}
					for _, userID := range domain.ExtractActionUserIDs(userMessage.MessageAction, userMessage.MessageActionData) {
						userIDs[userID] = true
					}
					userMessages = append(userMessages, userMessage)
					return nil
				})
			}
			it.Close()
			return nil
		})
	case maxID == 0 && minID != 0:
		_ = r.badger.View(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.Prefix = r.getPrefix(peerID, peerType)
			opts.Reverse = false
			it := txn.NewIterator(opts)
			for it.Seek(r.getMessageKey(peerID, peerType, minID)); it.ValidForPrefix(opts.Prefix); it.Next() {
				if limit--; limit < 0 {
					break
				}
				_ = it.Item().Value(func(val []byte) error {
					userMessage := new(msg.UserMessage)
					err := userMessage.Unmarshal(val)
					if err != nil {
						return err
					}
					userIDs[userMessage.SenderID] = true
					if userMessage.FwdSenderID != 0 {
						userIDs[userMessage.FwdSenderID] = true
					}
					for _, userID := range domain.ExtractActionUserIDs(userMessage.MessageAction, userMessage.MessageActionData) {
						userIDs[userID] = true
					}
					userMessages = append(userMessages, userMessage)
					return nil
				})
			}
			it.Close()
			sort.Slice(userMessages, func(i, j int) bool {
				return userMessages[i].ID > userMessages[j].ID
			})
			return nil
		})
	default:
		_ = r.badger.View(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.Prefix = r.getPrefix(peerID, peerType)
			opts.Reverse = true
			it := txn.NewIterator(opts)
			for it.Seek(r.getMessageKey(peerID, peerType, maxID)); it.ValidForPrefix(opts.Prefix); it.Next() {
				if limit--; limit < 0 {
					break
				}
				_ = it.Item().Value(func(val []byte) error {
					userMessage := new(msg.UserMessage)
					err := userMessage.Unmarshal(val)
					if err != nil {
						return err
					}
					userIDs[userMessage.SenderID] = true
					if userMessage.FwdSenderID != 0 {
						userIDs[userMessage.FwdSenderID] = true
					}
					for _, userID := range domain.ExtractActionUserIDs(userMessage.MessageAction, userMessage.MessageActionData) {
						userIDs[userID] = true
					}
					userMessages = append(userMessages, userMessage)
					if userMessage.ID <= minID {
						limit = 0
					}
					return nil
				})
			}
			it.Close()
			return nil
		})

	}

	users = Users.GetMany(userIDs.ToArray())
	return
}

func (r *repoMessages) Delete(userID int64, peerID int64, peerType int32, msgID int64) {
	// 1. Delete the Message
	err := r.badger.Update(func(txn *badger.Txn) error {
		// delete from messages
		err := txn.Delete(r.getMessageKey(peerID, peerType, msgID))
		if err != nil {
			return err
		}
		// delete from user messages
		return txn.Delete(r.getUserMessageKey(msgID))
	})
	if err != nil {
		return
	}

	// 2. Update the Dialog if necessary
	dialog := Dialogs.Get(peerID, peerType)
	if dialog == nil {
		return
	}

	// 3. Update top message if necessary
	if dialog.TopMessageID == msgID {
		Dialogs.updateTopMessageID(dialog)

	}

	// 4. Update unread/mention counter
	dialog.UnreadCount, dialog.MentionedCount = Dialogs.countUnread(dialog.PeerID, dialog.PeerType, userID, dialog.ReadInboxMaxID+1)
	Dialogs.Save(dialog)

	// 5. Delete the entry from the index
	_ = r.msgSearch.Delete(ronak.ByteToStr(r.getMessageKey(peerID, peerType, msgID)))
	return
}

func (r *repoMessages) DeleteAll(userID int64, peerID int64, peerType int32, maxID int64) {
	_ = r.badger.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = r.getPrefix(peerID, peerType)
		opts.PrefetchValues = false
		opts.Reverse = true
		it := txn.NewIterator(opts)
		for it.Seek(r.getMessageKey(peerID, peerType, maxID)); it.ValidForPrefix(opts.Prefix); it.Next() {
			_ = txn.Delete(it.Item().KeyCopy(nil))
		}
		it.Close()
		return nil
	})

	dialog := Dialogs.Get(peerID, peerType)
	dialog.UnreadCount, dialog.MentionedCount = Dialogs.countUnread(peerID, peerType, userID, maxID)
	Dialogs.Save(dialog)
}

func (r *repoMessages) SetContentRead(peerID int64, peerType int32, messageIDs []int64) {

	for _, msgID := range messageIDs {
		_ = r.badger.Update(func(txn *badger.Txn) error {
			userMessage, err := r.getUserMessage(msgID)
			if err != nil {
				return err
			}
			userMessage.ContentRead = true
			r.save(userMessage)
			return nil
		})

	}
	return
}

func (r *repoMessages) GetTopMessageID(peerID int64, peerType int32) (int64, error) {
	topMessageID := int64(0)
	err := r.badger.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = r.getPrefix(peerID, peerType)
		opts.Reverse = true
		it := txn.NewIterator(opts)
		it.Seek(Messages.getMessageKey(peerID, peerType, 2<<31))
		if it.ValidForPrefix(opts.Prefix) {
			userMessage := new(msg.UserMessage)
			_ = it.Item().Value(func(val []byte) error {
				return userMessage.Unmarshal(val)
			})
			topMessageID = userMessage.ID
		}
		it.Close()
		return nil
	})
	return topMessageID, err
}

func (r *repoMessages) SearchText(text string) []*msg.UserMessage {
	userMessages := make([]*msg.UserMessage, 0, 100)
	if r.msgSearch != nil {
		t1 := bleve.NewTermQuery("msg")
		t1.SetField("type")
		qs := make([]query.Query, 0)
		for _, term := range strings.Fields(text) {
			qs = append(qs, bleve.NewMatchQuery(term), bleve.NewPrefixQuery(term), bleve.NewFuzzyQuery(term))
		}
		t2 := bleve.NewDisjunctionQuery(qs...)
		searchRequest := bleve.NewSearchRequest(bleve.NewConjunctionQuery(t1, t2))
		searchResult, _ := r.msgSearch.Search(searchRequest)
		for _, hit := range searchResult.Hits {
			userMessage := r.getByKey(ronak.StrToByte(hit.ID))
			if userMessage != nil {
				userMessages = append(userMessages, userMessage)
			}
		}
	}

	return userMessages
}

func (r *repoMessages) SearchTextByPeerID(text string, peerID int64) []*msg.UserMessage {
	userMessages := make([]*msg.UserMessage, 0, 100)
	if r.msgSearch != nil {
		t1 := bleve.NewTermQuery("msg")
		t1.SetField("type")
		qs := make([]query.Query, 0)
		for _, term := range strings.Fields(text) {
			qs = append(qs, bleve.NewMatchQuery(term), bleve.NewPrefixQuery(term), bleve.NewFuzzyQuery(term))
		}
		t2 := bleve.NewDisjunctionQuery(qs...)
		t3 := bleve.NewTermQuery(fmt.Sprintf("%d", peerID))
		t3.SetField("peer_id")
		searchRequest := bleve.NewSearchRequest(bleve.NewConjunctionQuery(t1, t2, t3))
		searchResult, _ := r.msgSearch.Search(searchRequest)

		for _, hit := range searchResult.Hits {
			userMessage := r.getByKey(ronak.StrToByte(hit.ID))
			if userMessage != nil {
				userMessages = append(userMessages, userMessage)
			}
		}
	}

	return userMessages
}

func (r *repoMessages) GetSharedMedia(peerID int64, peerType int32, documentType domain.SharedMediaType) ([]*msg.UserMessage, error) {
	limit := 500
	userMessages := make([]*msg.UserMessage, 0, limit)
	_ = r.badger.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		opts.Prefix = r.getPrefix(peerID, peerType)
		opts.Reverse = true
		it := txn.NewIterator(opts)
		for it.Seek(Messages.getMessageKey(peerID, peerType, 1<<31)); it.ValidForPrefix(opts.Prefix); it.Next() {
			if limit--; limit < 0 {
				break
			}
			if it.Item().UserMeta() == byte(documentType) {
				_ = it.Item().Value(func(val []byte) error {
					userMessage := new(msg.UserMessage)
					err := userMessage.Unmarshal(val)
					if err != nil {
						return err
					}
					userMessages = append(userMessages, userMessage)
					return nil
				})
			}
		}
		it.Close()
		return nil
	})

	return userMessages, nil

}

func (r *repoMessages) ReIndex() {
	err := r.badger.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = ronak.StrToByte(prefixMessages)
		it := txn.NewIterator(opts)
		for it.Rewind(); it.Valid(); it.Next() {
			_ = it.Item().Value(func(val []byte) error {
				message := new(msg.UserMessage)
				_ = message.Unmarshal(val)
				_ = r.msgSearch.Index(ronak.ByteToStr(r.getMessageKey(message.PeerID, message.PeerType, message.ID)), MessageSearch{
					Type:   "msg",
					Body:   message.Body,
					PeerID: message.PeerID,
				})
				return nil
			})
		}
		it.Close()
		return nil
	})
	if err != nil {
		logs.Warn("Error On ReIndex", zap.Error(err))
	}
}
