package repo

import (
	"bytes"
	"context"
	"fmt"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/chat"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/pb"
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

func getMessageKey(peerID int64, peerType int32, msgID int64) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%021d.%d.%012d", prefixMessages, peerID, peerType, msgID))
}

func getMessagePrefix(peerID int64, peerType int32) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%021d.%d.", prefixMessages, peerID, peerType))
}

func getUserMessageKey(msgID int64) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%012d", prefixUserMessages, msgID))
}

func getMessageByID(txn *badger.Txn, msgID int64) (*msg.UserMessage, error) {
	message := new(msg.UserMessage)
	item, err := txn.Get(getUserMessageKey(msgID))
	if err != nil {
		return nil, err
	}
	err = item.Value(func(val []byte) error {
		parts := strings.Split(ronak.ByteToStr(val), ".")
		if len(parts) != 2 {
			return domain.ErrInvalidUserMessageKey
		}
		itemMessage, err := txn.Get(getMessageKey(ronak.StrToInt64(parts[0]), ronak.StrToInt32(parts[1]), msgID))
		if err != nil {
			return err
		}
		return itemMessage.Value(func(val []byte) error {
			return message.Unmarshal(val)
		})
	})

	if err != nil {
		return nil, err
	}
	return message, nil
}

func getMessageByKey(txn *badger.Txn, msgKey []byte) (*msg.UserMessage, error) {
	message := &msg.UserMessage{}
	item, err := txn.Get(msgKey)
	if err != nil {
		return nil, err
	}
	err = item.Value(func(val []byte) error {
		return message.Unmarshal(val)
	})
	if err != nil {
		return nil, err
	}
	return message, nil
}

func saveMessage(txn *badger.Txn, message *msg.UserMessage) error {
	messageBytes, _ := message.Marshal()
	docType := msg.ClientMediaNone
	switch message.MediaType {
	case msg.MediaTypeDocument:
		doc := new(msg.MediaDocument)
		_ = doc.Unmarshal(message.Media)
		if doc.Doc == nil {
			logs.Error("RepoMessage got error on save message, Document is Nil",
				zap.Int64("MessageID", message.ID),
				zap.Int64("SenderID", message.SenderID),
			)
			return nil
		}
		for _, da := range doc.Doc.Attributes {
			switch da.Type {
			case msg.AttributeTypeAudio:
				a := new(msg.DocumentAttributeAudio)
				_ = a.Unmarshal(da.Data)
				if a.Voice {
					docType = msg.ClientMediaVoice
				} else {
					docType = msg.ClientMediaAudio
				}
			case msg.AttributeTypeVideo, msg.AttributeTypePhoto:
				docType = msg.ClientMediaMedia
			case msg.AttributeTypeFile:
				if docType == msg.ClientMediaNone {
					docType = msg.ClientMediaFile
				}
			}
		}
	default:
		// Do nothing
	}
	// 1. Write Message
	err := txn.SetEntry(
		badger.NewEntry(
			getMessageKey(message.PeerID, message.PeerType, message.ID),
			messageBytes,
		).WithMeta(byte(docType)),
	)
	if err != nil {
		return err
	}

	// 2. Write UserMessage
	err = txn.SetEntry(
		badger.NewEntry(
			getUserMessageKey(message.ID),
			ronak.StrToByte(fmt.Sprintf("%d.%d", message.PeerID, message.PeerType)),
		).WithMeta(byte(docType)),
	)
	if err != nil {
		return err
	}

	indexMessage(
		ronak.ByteToStr(getMessageKey(message.PeerID, message.PeerType, message.ID)),
		MessageSearch{
			Type:   "msg",
			Body:   message.Body,
			PeerID: fmt.Sprintf("%d", message.PeerID),
		},
	)

	return nil
}

func (r *repoMessages) Get(messageID int64) (um *msg.UserMessage, err error) {
	err = badgerView(func(txn *badger.Txn) error {
		um, err = getMessageByID(txn, messageID)
		return err
	})
	return
}

func (r *repoMessages) GetMany(messageIDs []int64) []*msg.UserMessage {
	userMessages := make([]*msg.UserMessage, 0, len(messageIDs))
	err := badgerView(func(txn *badger.Txn) error {
		for _, messageID := range messageIDs {
			userMessage, err := getMessageByID(txn, messageID)
			if err != nil {
				switch err {
				case badger.ErrKeyNotFound:
					logs.Warn("RepoMessage got error on get many (key not found)",
						zap.Int64("MsgID", messageID),
					)
				default:
					logs.Warn("RepoMessage got error on get many",
						zap.Error(err),
						zap.Int64("MsgID", messageID),
					)
				}
			}
			if err == nil {
				userMessages = append(userMessages, userMessage)
			}
		}
		return nil
	})
	logs.ErrorOnErr("RepoMessage got error on get many", err,
		zap.Int64s("MsgIDs", messageIDs),
	)
	return userMessages
}

func (r *repoMessages) SaveNew(message *msg.UserMessage, userID int64) error {
	if message == nil {
		return nil
	}
	err := badgerUpdate(func(txn *badger.Txn) error {
		err := saveMessage(txn, message)
		if err != nil {
			return err
		}

		err = saveMessageMedia(txn, message)
		if err != nil {
			return err
		}

		dialog, err := getDialog(txn, message.PeerID, message.PeerType)
		if err != nil {
			return err
		}
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
		return saveDialog(txn, dialog)
	})
	logs.ErrorOnErr("RepoMessage got error on save new message", err)

	return err
}

func (r *repoMessages) Save(messages ...*msg.UserMessage) {
	err := badgerUpdate(func(txn *badger.Txn) error {
		for _, message := range messages {
			err := saveMessage(txn, message)
			if err != nil {
				return err
			}
			err = saveMessageMedia(txn, message)
			if err != nil {
				return err
			}

			for _, labelID := range message.LabelIDs {
				err = addLabelToMessage(txn, labelID, message.PeerType, message.PeerID, message.ID)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	logs.ErrorOnErr("RepoMessage got error on save", err)
}

func (r *repoMessages) GetMessageHistory(peerID int64, peerType int32, minID, maxID int64, limit int32) (userMessages []*msg.UserMessage, users []*msg.User) {
	userMessages = make([]*msg.UserMessage, 0, limit)
	userIDs := domain.MInt64B{}
	switch {
	case maxID == 0 && minID == 0:
		dialog, err := Dialogs.Get(peerID, peerType)
		if err != nil {
			logs.Error("RepoMessage got error on GetHistory",
				zap.Error(err),
				zap.Int64("PeerID", peerID),
			)
			return
		}
		maxID = dialog.TopMessageID
		fallthrough
	case maxID != 0 && minID == 0:
		startTime := time.Now()
		var stopWatch1, stopWatch2 time.Time
		_ = badgerView(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.Prefix = getMessagePrefix(peerID, peerType)
			opts.Reverse = true
			it := txn.NewIterator(opts)
			it.Seek(getMessageKey(peerID, peerType, maxID))
			stopWatch1 = time.Now()
			for ; it.ValidForPrefix(opts.Prefix); it.Next() {
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
			stopWatch2 = time.Now()
			return nil
		})
		logs.Info("RepoMessage got history", zap.Int64("MinID", minID), zap.Int64("MaxID", maxID),
			zap.Duration("SP1", stopWatch1.Sub(startTime)),
			zap.Duration("SP2", stopWatch2.Sub(startTime)),
		)
	case maxID == 0 && minID != 0:
		startTime := time.Now()
		var stopWatch1, stopWatch2, stopWatch3 time.Time
		_ = badgerView(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.Prefix = getMessagePrefix(peerID, peerType)
			opts.Reverse = false
			it := txn.NewIterator(opts)
			it.Seek(getMessageKey(peerID, peerType, minID))
			stopWatch1 = time.Now()
			for ; it.ValidForPrefix(opts.Prefix); it.Next() {
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
			stopWatch2 = time.Now()
			sort.Slice(userMessages, func(i, j int) bool {
				return userMessages[i].ID > userMessages[j].ID
			})
			stopWatch3 = time.Now()
			return nil
		})
		logs.Info("RepoMessage got history", zap.Int64("MinID", minID), zap.Int64("MaxID", maxID),
			zap.Duration("SP1", stopWatch1.Sub(startTime)),
			zap.Duration("SP2", stopWatch2.Sub(startTime)),
			zap.Duration("SP3", stopWatch3.Sub(startTime)),
		)
	default:
		startTime := time.Now()
		var stopWatch1, stopWatch2 time.Time
		_ = badgerView(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.Prefix = getMessagePrefix(peerID, peerType)
			opts.Reverse = true
			it := txn.NewIterator(opts)
			it.Seek(getMessageKey(peerID, peerType, maxID))
			stopWatch1 = time.Now()
			for ; it.ValidForPrefix(opts.Prefix); it.Next() {
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
			stopWatch2 = time.Now()
			logs.Info("RepoMessage got history", zap.Int64("MinID", minID), zap.Int64("MaxID", maxID),
				zap.Duration("SP1", stopWatch1.Sub(startTime)),
				zap.Duration("SP2", stopWatch2.Sub(startTime)),
			)
			return nil
		})

	}

	users = Users.GetMany(userIDs.ToArray())
	return
}

func (r *repoMessages) Delete(userID int64, peerID int64, peerType int32, msgIDs ...int64) {
	sort.Slice(msgIDs, func(i, j int) bool {
		return msgIDs[i] < msgIDs[j]
	})
	err := badgerUpdate(func(txn *badger.Txn) error {
		// 2. Update the Dialog if necessary
		dialog, err := getDialog(txn, peerID, peerType)
		if err != nil {
			return err
		}

		for _, msgID := range msgIDs {
			// delete from messages
			_ = txn.Delete(getMessageKey(peerID, peerType, msgID))

			// delete from user messages
			_ = txn.Delete(getUserMessageKey(msgID))
		}

		msgID := msgIDs[len(msgIDs)-1]
		if dialog.TopMessageID == msgID {
			opts := badger.DefaultIteratorOptions
			opts.Prefix = getMessagePrefix(dialog.PeerID, dialog.PeerType)
			opts.Reverse = true
			it := txn.NewIterator(opts)
			it.Seek(getMessageKey(dialog.PeerID, dialog.PeerType, dialog.TopMessageID))
			if it.ValidForPrefix(opts.Prefix) {
				userMessage := new(msg.UserMessage)
				_ = it.Item().Value(func(val []byte) error {
					return userMessage.Unmarshal(val)
				})
				dialog.TopMessageID = userMessage.ID
			}
			it.Close()
			if dialog.TopMessageID == msgID {
				_ = txn.Delete(getDialogKey(peerID, peerType))
				indexMessageRemove(ronak.ByteToStr(getMessageKey(peerID, peerType, msgID)))
				return nil
			}
		}

		dialog.UnreadCount, dialog.MentionedCount, err = countDialogUnread(txn, dialog.PeerID, dialog.PeerType, userID, dialog.ReadInboxMaxID+1)
		if err != nil {
			return err
		}
		err = saveDialog(txn, dialog)
		if err != nil {
			return err
		}
		indexMessageRemove(ronak.ByteToStr(getMessageKey(peerID, peerType, msgID)))
		return nil
	})
	logs.ErrorOnErr("RepoMessage got error on delete", err)
}

func (r *repoMessages) ClearHistory(userID int64, peerID int64, peerType int32, maxID int64) error {
	err := badgerUpdate(func(txn *badger.Txn) error {
		st := r.badger.NewStream()
		st.Prefix = getMessagePrefix(peerID, peerType)
		st.NumGo = 10
		maxKey := getMessageKey(peerID, peerType, maxID)
		st.ChooseKey = func(item *badger.Item) bool {
			return bytes.Compare(item.Key(), maxKey) <= 0
		}
		st.Send = func(kvList *pb.KVList) error {
			for _, kv := range kvList.Kv {
				err := txn.Delete(kv.Key)
				if err != nil {
					logs.Warn("RepoMessage got error on delete all", zap.Error(err), zap.String("Key", ronak.ByteToStr(kv.Key)))
				}
			}
			return nil
		}
		err := st.Orchestrate(context.Background())
		if err != nil {
			return err
		}
		dialog, err := getDialog(txn, peerID, peerType)
		if err != nil {
			return err
		}
		dialog.UnreadCount, dialog.MentionedCount, err = countDialogUnread(txn, peerID, peerType, userID, maxID)
		if err != nil {
			return err
		}
		return saveDialog(txn, dialog)
	})
	logs.ErrorOnErr("RepoMessage got error on delete all", err)
	return err
}

func (r *repoMessages) SetContentRead(peerID int64, peerType int32, messageIDs []int64) {
	err := badgerUpdate(func(txn *badger.Txn) error {
		for _, msgID := range messageIDs {
			userMessage, err := getMessageByID(txn, msgID)
			if err != nil {
				return err
			}
			userMessage.ContentRead = true
			err = saveMessage(txn, userMessage)
			if err != nil {
				return err
			}
		}
		return nil
	})
	logs.ErrorOnErr("RepoMessage got error on set content read", err)
	return
}

func (r *repoMessages) GetTopMessageID(peerID int64, peerType int32) (int64, error) {

	topMessageID := int64(0)
	err := badgerView(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = getMessagePrefix(peerID, peerType)
		opts.Reverse = true
		it := txn.NewIterator(opts)
		it.Seek(getMessageKey(peerID, peerType, 2<<31))
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

func (r *repoMessages) SearchText(text string, limit int32) []*msg.UserMessage {
	userMessages := make([]*msg.UserMessage, 0, limit)
	if r.msgSearch == nil {
		return userMessages
	}
	t1 := bleve.NewTermQuery("msg")
	t1.SetField("type")
	qs := make([]query.Query, 0)
	for _, term := range strings.Fields(text) {
		qs = append(qs, bleve.NewMatchQuery(term), bleve.NewPrefixQuery(term), bleve.NewFuzzyQuery(term))
	}
	t2 := bleve.NewDisjunctionQuery(qs...)
	searchRequest := bleve.NewSearchRequest(bleve.NewConjunctionQuery(t1, t2))
	searchResult, _ := r.msgSearch.Search(searchRequest)
	_ = badgerView(func(txn *badger.Txn) error {
		for _, hit := range searchResult.Hits {
			userMessage, _ := getMessageByKey(txn, ronak.StrToByte(hit.ID))
			if userMessage != nil {
				userMessages = append(userMessages, userMessage)
			}
		}
		return nil
	})
	if len(userMessages) > int(limit) {
		userMessages = userMessages[:limit]
	}
	return userMessages
}

func (r *repoMessages) SearchTextByPeerID(text string, peerID int64, limit int32) []*msg.UserMessage {
	userMessages := make([]*msg.UserMessage, 0, limit)
	if r.msgSearch == nil {
		return userMessages
	}

	t1 := bleve.NewTermQuery("msg")
	t1.SetField("type")
	qs := make([]query.Query, 0)
	for _, term := range strings.Fields(text) {
		qs = append(qs, bleve.NewMatchQuery(term), bleve.NewPrefixQuery(term), bleve.NewFuzzyQuery(term))
	}
	t2 := bleve.NewDisjunctionQuery(qs...)
	t3 := bleve.NewTermQuery(fmt.Sprintf("%d", abs(peerID)))
	t3.SetField("peer_id")
	searchRequest := bleve.NewSearchRequest(bleve.NewConjunctionQuery(t1, t2, t3))
	searchResult, _ := r.msgSearch.Search(searchRequest)
	_ = badgerView(func(txn *badger.Txn) error {
		for _, hit := range searchResult.Hits {
			userMessage, _ := getMessageByKey(txn, ronak.StrToByte(hit.ID))
			if userMessage != nil {
				userMessages = append(userMessages, userMessage)
			}
		}
		return nil
	})

	if len(userMessages) > int(limit) {
		userMessages = userMessages[:limit]
	}
	return userMessages
}

func (r *repoMessages) SearchByLabels(labelIDs []int32, peerID int64, limit int32) []*msg.UserMessage {
	userMessages := make([]*msg.UserMessage, 0, limit)
	_ = badgerView(func(txn *badger.Txn) error {
		st := r.badger.NewStream()
		st.Prefix = ronak.StrToByte(prefixMessages)
		st.ChooseKey = func(item *badger.Item) bool {
			m := &msg.UserMessage{}
			err := item.Value(func(val []byte) error {
				return m.Unmarshal(val)
			})
			if err != nil {
				return false
			}
			if len(m.LabelIDs) < len(labelIDs) {
				return false
			}
			var found bool
			for _, li := range labelIDs {
				found = false
				for _, lj := range m.LabelIDs {
					if li == lj {
						found = true
						break
					}
				}
				if !found {
					return false
				}
			}
			return true
		}
		st.Send = func(list *pb.KVList) error {
			if int32(len(userMessages)) > limit {
				return nil
			}
			for _, kv := range list.Kv {
				m := &msg.UserMessage{}
				if err := m.Unmarshal(kv.Value); err == nil {
					if peerID == 0 {
						userMessages = append(userMessages, m)
					} else if m.PeerID == peerID {
						userMessages = append(userMessages, m)
					}
				}
			}
			return nil
		}
		return st.Orchestrate(context.Background())
	})
	return userMessages

}

func (r *repoMessages) GetSharedMedia(peerID int64, peerType int32, documentType msg.ClientMediaType) ([]*msg.UserMessage, error) {
	limit := 500
	userMessages := make([]*msg.UserMessage, 0, limit)
	_ = badgerView(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		opts.Prefix = getMessagePrefix(peerID, peerType)
		opts.Reverse = true
		it := txn.NewIterator(opts)
		for it.Seek(getMessageKey(peerID, peerType, 1<<31)); it.ValidForPrefix(opts.Prefix); it.Next() {
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
	err := badgerView(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = ronak.StrToByte(prefixMessages)
		it := txn.NewIterator(opts)
		for it.Rewind(); it.Valid(); it.Next() {
			_ = it.Item().Value(func(val []byte) error {
				message := new(msg.UserMessage)
				_ = message.Unmarshal(val)
				indexMessage(
					ronak.ByteToStr(getMessageKey(message.PeerID, message.PeerType, message.ID)),
					MessageSearch{
						Type:   "msg",
						Body:   message.Body,
						PeerID: fmt.Sprintf("%d", message.PeerID),
					},
				)
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
