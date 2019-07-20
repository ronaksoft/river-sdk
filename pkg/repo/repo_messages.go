package repo

import (
	"fmt"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/blevesearch/bleve"
	"github.com/dgraph-io/badger"
	"github.com/scylladb/go-set/i64set"
	"sort"
	"strings"
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
		return nil, err
	}
	return message, nil
}

func (r *repoMessages) Get(messageID int64) *msg.UserMessage {
	r.mx.Lock()
	defer r.mx.Unlock()

	userMessage, err := r.getUserMessage(messageID)
	if err != nil {
		return nil
	}
	return userMessage
}

func (r *repoMessages) GetMany(messageIDs []int64) []*msg.UserMessage {
	r.mx.Lock()
	defer r.mx.Unlock()

	userMessages := make([]*msg.UserMessage, 0, len(messageIDs))
	for _, messageID := range messageIDs {
		userMessage, err := r.getUserMessage(messageID)
		if err == nil {
			userMessages = append(userMessages, userMessage)
		}

	}
	return userMessages
}

func (r *repoMessages) SaveNew(message *msg.UserMessage, dialog *msg.Dialog, userID int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if message == nil {
		return domain.ErrNotFound
	}

	messageBytes, _ := message.Marshal()
	return r.badger.Update(func(txn *badger.Txn) error {
		// 1. Write Message
		err := txn.SetEntry(
			badger.NewEntry(
				r.getMessageKey(message.PeerID, message.PeerType, message.ID),
				messageBytes,
			).WithMeta(byte(message.MediaType)),
		)
		if err != nil {
			return err
		}

		// 2. WriteUserMessage
		err = txn.SetEntry(
			badger.NewEntry(
				r.getUserMessageKey(message.ID),
				ronak.StrToByte(fmt.Sprintf("%d.%d", message.PeerID, message.PeerType)),
			).WithMeta(byte(message.MediaType)),
		)
		if err != nil {
			return err
		}

		// 3. Read Dialog
		dialogKey := Dialogs.getDialogKey(message.PeerID, message.PeerType)
		item, err := txn.Get(dialogKey)
		if err != nil {
			return err
		}
		dialog := new(msg.Dialog)
		return item.Value(func(val []byte) error {
			_ = dialog.Unmarshal(val)
			// Update Dialog if it is a new message
			if message.ID > dialog.TopMessageID {
				dialog.TopMessageID = message.ID
				if !dialog.Pinned {
					err = Dialogs.updateLastUpdate(message.PeerID, message.PeerType, message.CreatedOn)
					if err != nil {
						return err
					}
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
				return Dialogs.Save(dialog)
			}
			return nil
		})
	})
}

func (r *repoMessages) Save(message *msg.UserMessage) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if message == nil {
		return domain.ErrNotFound
	}

	messageBytes, _ := message.Marshal()
	err := r.badger.Update(func(txn *badger.Txn) error {
		// 1. Write Message

		err := txn.SetEntry(
			badger.NewEntry(
				r.getMessageKey(message.PeerID, message.PeerType, message.ID),
				messageBytes,
			).WithMeta(byte(message.MediaType)),
		)
		if err != nil {
			return err
		}

		// 2. WriteUserMessage
		return txn.SetEntry(
			badger.NewEntry(
				r.getUserMessageKey(message.ID),
				ronak.StrToByte(fmt.Sprintf("%d.%d", message.PeerID, message.PeerType)),
			).WithMeta(byte(message.MediaType)),
		)
	})

	_ = r.searchIndex.Index(ronak.ByteToStr(r.getUserMessageKey(message.ID)), MessageSearch{
		Type:   "msg",
		Body:   message.Body,
		PeerID: message.PeerID,
	})

	return err
}

func (r *repoMessages) GetMessageHistory(peerID int64, peerType int32, minID, maxID int64, limit int32) (userMessages []*msg.UserMessage, users []*msg.User) {
	r.mx.Lock()
	defer r.mx.Unlock()

	userMessages = make([]*msg.UserMessage, 0, limit)
	userIDs := i64set.New()
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
			for it.Seek(r.getMessageKey(peerID, peerType, maxID)); it.Valid(); it.Next() {
				if limit--; limit < 0 {
					break
				}
				_ = it.Item().Value(func(val []byte) error {
					userMessage := new(msg.UserMessage)
					err := userMessage.Unmarshal(val)
					if err != nil {
						return err
					}
					userIDs.Add(userMessage.SenderID)
					if userMessage.FwdSenderID != 0 {
						userIDs.Add(userMessage.FwdSenderID)
					}
					userIDs.Add(domain.ExtractActionUserIDs(userMessage.MessageAction, userMessage.MessageActionData)...)
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
			for it.Seek(r.getMessageKey(peerID, peerType, minID)); it.Valid(); it.Next() {
				if limit--; limit < 0 {
					break
				}
				_ = it.Item().Value(func(val []byte) error {
					userMessage := new(msg.UserMessage)
					err := userMessage.Unmarshal(val)
					if err != nil {
						return err
					}
					userIDs.Add(userMessage.SenderID)
					if userMessage.FwdSenderID != 0 {
						userIDs.Add(userMessage.FwdSenderID)
					}
					userIDs.Add(domain.ExtractActionUserIDs(userMessage.MessageAction, userMessage.MessageActionData)...)
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
			for it.Seek(r.getMessageKey(peerID, peerType, maxID)); it.Valid(); it.Next() {
				if limit--; limit < 0 {
					break
				}

				userMessage := new(msg.UserMessage)
				_ = it.Item().Value(func(val []byte) error {
					err := userMessage.Unmarshal(val)
					if err != nil {
						return err
					}
					userIDs.Add(userMessage.SenderID)
					if userMessage.FwdSenderID != 0 {
						userIDs.Add(userMessage.FwdSenderID)
					}
					userIDs.Add(domain.ExtractActionUserIDs(userMessage.MessageAction, userMessage.MessageActionData)...)
					userMessages = append(userMessages, userMessage)
					return nil
				})
				if userMessage.ID <= minID {
					break
				}
			}
			it.Close()
			return nil
		})

	}

	users = Users.GetMany(userIDs.List())
	return
}

func (r *repoMessages) DeleteDialogMessage(peerID int64, peerType int32, msgID int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

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
		return err
	}

	// 2. Update the Dialog if necessary
	dialog := Dialogs.Get(peerID, peerType)
	if dialog == nil {
		return nil
	}
	if dialog.TopMessageID == msgID {
		Dialogs.updateTopMessageID(peerID, peerType)
	}
	return err
}

func (r *repoMessages) SetContentRead(peerID int64, peerType int32, messageIDs []int64) {
	r.mx.Lock()
	defer r.mx.Unlock()

	for _, msgID := range messageIDs {
		_ = r.badger.Update(func(txn *badger.Txn) error {
			userMessage, err := r.getUserMessage(msgID)
			if err != nil {
				return err
			}
			userMessage.ContentRead = true
			return r.Save(userMessage)
		})

	}
	return
}

func (r *repoMessages) GetTopMessageID(peerID int64, peerType int32) (int64, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	topMessageID := int64(0)
	err := r.badger.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = r.getPrefix(peerID, peerType)
		opts.Reverse = true
		it := txn.NewIterator(opts)
		it.Seek(Messages.getMessageKey(peerID, peerType, 2<<31))
		if it.Valid() {
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
	r.mx.Lock()
	defer r.mx.Unlock()

	textTerm := bleve.NewTermQuery(text)
	textTerm.SetField("Body")
	searchRequest := bleve.NewSearchRequest(textTerm)
	searchResult, _ := r.searchIndex.Search(searchRequest)
	userMessages := make([]*msg.UserMessage, 0, 100)
	for _, hit := range searchResult.Hits {
		userMessage := r.Get(ronak.StrToInt64(hit.Fields["Body"].(string)))
		if userMessage != nil {
			userMessages = append(userMessages, userMessage)
		}

	}
	return userMessages
}

func (r *repoMessages) SearchTextByPeerID(text string, peerID int64) []*msg.UserMessage {
	r.mx.Lock()
	defer r.mx.Unlock()
	textTerm := bleve.NewTermQuery(text)
	textTerm.SetField("Body")
	peerTerm := bleve.NewTermQuery(fmt.Sprintf("%d", peerID))
	peerTerm.SetField("PeerID")
	q := bleve.NewConjunctionQuery(textTerm, peerTerm)
	searchRequest := bleve.NewSearchRequest(q)
	searchResult, _ := r.searchIndex.Search(searchRequest)
	userMessages := make([]*msg.UserMessage, 0, 100)
	for _, hit := range searchResult.Hits {
		userMessage := r.Get(ronak.StrToInt64(hit.Fields["Body"].(string)))
		if userMessage != nil {
			userMessages = append(userMessages, userMessage)
		}

	}
	return userMessages
}

func (r *repoMessages) ClearAll() error {
	err := r.badger.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = ronak.StrToByte(fmt.Sprintf("%s.", prefixMessages))
		opts.Reverse = true
		it := txn.NewIterator(opts)
		count := 0
		for it.Rewind(); it.Valid(); it.Next() {
			count++
			err := txn.Delete(it.Item().Key())
			if err != nil {
				return err
			}
			if count%100 == 0 {
				_ = txn.Commit()
			}
		}
		it.Close()
		return nil
	})
	if err != nil {
		return err
	}

	err = r.badger.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = ronak.StrToByte(fmt.Sprintf("%s.", prefixUserMessages))
		opts.Reverse = true
		it := txn.NewIterator(opts)
		count := 0
		for it.Rewind(); it.Valid(); it.Next() {
			count++
			err := txn.Delete(it.Item().Key())
			if err != nil {
				return err
			}
			if count%100 == 0 {
				_ = txn.Commit()
			}
		}
		it.Close()
		return nil
	})

	return err
}

// OLD

// func (r *repoMessages) GetMessageHistoryWithPendingMessages(peerID int64, peerType int32, minID, maxID int64, limit int32) (protoMsgs []*msg.UserMessage, protoUsers []*msg.User) {
// 	r.mx.Lock()
// 	defer r.mx.Unlock()
//
// 	dtoMsgs := make([]dto.Messages, 0, limit)
// 	dtoPendings := make([]dto.MessagesPending, 0, limit)
//
// 	var err error
// 	if maxID == 0 && minID == 0 {
// 		err = r.db.Order("ID DESC").Limit(limit).Where("PeerID = ? AND PeerType = ?", peerID, peerType).Find(&dtoMsgs).Error
// 	} else if minID == 0 && maxID != 0 {
// 		err = r.db.Order("ID DESC").Limit(limit).Where("PeerID = ? AND PeerType = ? AND messages.ID <= ?", peerID, peerType, maxID).Find(&dtoMsgs).Error
// 	} else if minID != 0 && maxID == 0 {
// 		err = r.db.Order("ID ASC").Limit(limit).Where("PeerID = ? AND PeerType = ? AND messages.ID >= ?", peerID, peerType, minID).Find(&dtoMsgs).Error
// 		// sort DESC again
// 		if err == nil {
// 			sort.Slice(dtoMsgs, func(i, j int) bool {
// 				return dtoMsgs[i].ID > dtoMsgs[j].ID
// 			})
// 		}
// 	} else {
// 		err = r.db.Order("ID DESC").Limit(limit).Where("PeerID = ? AND PeerType = ? AND messages.ID >= ? AND messages.ID <= ?", peerID, peerType, minID, maxID).Find(&dtoMsgs).Error
// 	}
//
// 	if err != nil {
// 		logs.Warn("Repo::GetMessageHistory()-> fetch messages", zap.Error(err))
// 		return
// 	}
//
// 	dtoResult := make([]dto.Messages, 0, limit)
//
// 	// get all pending message for this user
// 	err = r.db.Order("ID ASC").Limit(limit).Where("PeerID = ? AND PeerType = ? ", peerID, peerType).Find(&dtoPendings).Error
// 	if err == nil {
// 		for _, v := range dtoPendings {
// 			tmp := new(dto.Messages)
// 			v.MapToDtoMessage(tmp)
// 			dtoResult = append(dtoResult, *tmp)
// 		}
// 	}
// 	dtoResult = append(dtoResult, dtoMsgs...)
//
// 	userIDs := domain.MInt64B{}
// 	for _, v := range dtoResult {
// 		tmp := new(msg.UserMessage)
// 		v.MapTo(tmp)
// 		protoMsgs = append(protoMsgs, tmp)
// 		userIDs[v.SenderID] = true
// 		userIDs[v.FwdSenderID] = true
// 		// load MessageActionData users
// 		actionUserIds := domain.ExtractActionUserIDs(v.MessageAction, v.MessageActionData)
// 		for _, id := range actionUserIds {
// 			userIDs[id] = true
// 		}
// 	}
//
// 	// Get users <rewrite it here to remove coupling>
// 	users := make([]dto.Users, 0, len(userIDs))
//
// 	err = r.db.Where("ID in (?)", userIDs.ToArray()).Find(&users).Error
// 	if err != nil {
// 		logs.Warn("Repo::GetMessageHistory()-> fetch users", zap.Error(err))
// 		return
// 	}
//
// 	for _, v := range users {
// 		tmp := new(msg.User)
// 		v.MapToUser(tmp)
// 		protoUsers = append(protoUsers, tmp)
//
// 	}
// 	return
// }
//
//
//
