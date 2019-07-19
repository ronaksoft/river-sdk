package repo

import (
	"fmt"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo/dto"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/dgraph-io/badger"
	"github.com/tidwall/buntdb"
	"go.uber.org/zap"
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

func (r *repoMessages) SaveNewMessage(message *msg.UserMessage, dialog *msg.Dialog, userID int64) error {
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
		dialogKey := Dialogs.getKey(message.PeerID, message.PeerType)
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
					// update the list order
					err = r.bunt.Update(func(tx *buntdb.Tx) error {
						_, _, err := tx.Set(
							ronak.ByteToStr(Dialogs.getKey(message.PeerID, message.PeerType)),
							fmt.Sprintf("%021d", message.CreatedOn),
							nil,
						)
						return err
					})
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

				dialogBytes, _ := dialog.Marshal()
				return txn.SetEntry(badger.NewEntry(dialogKey, dialogBytes))
			}

			return nil
		})
	})
}

func (r *repoMessages) SaveMessage(message *msg.UserMessage) error {
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
		return txn.SetEntry(
			badger.NewEntry(
				r.getUserMessageKey(message.ID),
				ronak.StrToByte(fmt.Sprintf("%d.%d", message.PeerID, message.PeerType)),
			).WithMeta(byte(message.MediaType)),
		)
	})
}

func (r *repoMessages) GetManyMessages(messageIDs []int64) []*msg.UserMessage {
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

func (r *repoMessages) GetMessageHistoryWithPendingMessages(peerID int64, peerType int32, minID, maxID int64, limit int32) (protoMsgs []*msg.UserMessage, protoUsers []*msg.User) {
	r.mx.Lock()
	defer r.mx.Unlock()

	dtoMsgs := make([]dto.Messages, 0, limit)
	dtoPendings := make([]dto.MessagesPending, 0, limit)

	var err error
	if maxID == 0 && minID == 0 {
		err = r.db.Order("ID DESC").Limit(limit).Where("PeerID = ? AND PeerType = ?", peerID, peerType).Find(&dtoMsgs).Error
	} else if minID == 0 && maxID != 0 {
		err = r.db.Order("ID DESC").Limit(limit).Where("PeerID = ? AND PeerType = ? AND messages.ID <= ?", peerID, peerType, maxID).Find(&dtoMsgs).Error
	} else if minID != 0 && maxID == 0 {
		err = r.db.Order("ID ASC").Limit(limit).Where("PeerID = ? AND PeerType = ? AND messages.ID >= ?", peerID, peerType, minID).Find(&dtoMsgs).Error
		// sort DESC again
		if err == nil {
			sort.Slice(dtoMsgs, func(i, j int) bool {
				return dtoMsgs[i].ID > dtoMsgs[j].ID
			})
		}
	} else {
		err = r.db.Order("ID DESC").Limit(limit).Where("PeerID = ? AND PeerType = ? AND messages.ID >= ? AND messages.ID <= ?", peerID, peerType, minID, maxID).Find(&dtoMsgs).Error
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

	userMessages := make([]*msg.UserMessage, 0, limit)
	switch {
	case maxID == 0 && minID == 0:
		maxID = Dialogs.Get(peerID, peerType).TopMessageID
		fallthrough
	case maxID != 0 && minID == 0:
		_ = r.badger.View(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.Prefix = ronak.StrToByte(fmt.Sprintf("%s.%d.%d.", prefixMessages, peerID, peerType))
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
					userMessages = append(userMessages, userMessage)
					return nil
				})
			}
			return nil
		})
	case maxID == 0 && minID != 0:
		_ = r.badger.View(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.Prefix = ronak.StrToByte(fmt.Sprintf("%s.%d.%d.", prefixMessages, peerID, peerType))
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
					userMessages = append(userMessages, userMessage)
					return nil
				})
			}
			sort.Slice(userMessages, func(i, j int) bool {
				return userMessages[i].ID > userMessages[j].ID
			})
			return nil
		})
	default:
		_ = r.badger.View(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.Prefix = ronak.StrToByte(fmt.Sprintf("%s.%d.%d.", prefixMessages, peerID, peerType))
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
					userMessages = append(userMessages, userMessage)
					return nil
				})
				if userMessage.ID <= minID {
					break
				}
			}
			return nil
		})

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

	count := int32(0)
	_ = r.badger.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = ronak.StrToByte(fmt.Sprintf("%s.%d.%d.", prefixMessages, peerID, peerType))
		opts.Reverse = false
		it := txn.NewIterator(opts)
		for it.Seek(r.getMessageKey(peerID, peerType, maxID)); it.Valid(); it.Next() {

			_ = it.Item().Value(func(val []byte) error {
				userMessage := new(msg.UserMessage)
				err := userMessage.Unmarshal(val)
				if err != nil {
					return err
				}
				if userMessage.SenderID != userID {
					count++
				}
				return nil
			})
		}
		return nil
	})
	return count
}

func (r *repoMessages) DeleteDialogMessage(peerID int64, peerType int32, msgID int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	err := r.badger.Update(func(txn *badger.Txn) error {
		err := txn.Delete(r.getMessageKey(peerID, peerType, msgID))
		if err != nil {
			return err
		}
		return txn.Delete(r.getUserMessageKey(msgID))
	})
	if err != nil {
		return err
 	}

	err = r.badger.Update(func(txn *badger.Txn) error {
		dialog := Dialogs.Get(peerID, peerType)
		if dialog == nil {
			return nil
		}
		if dialog.TopMessageID == msgID {

		}

	})
	return err
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
