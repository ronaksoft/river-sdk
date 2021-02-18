package repo

import (
	"bytes"
	"context"
	"fmt"
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/pools"
	"git.ronaksoft.com/river/sdk/internal/tools"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/search/query"
	"github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/badger/v2/pb"
	"go.uber.org/zap"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	prefixMessages     = "MSG"
	prefixUserMessages = "UMSG"
)

type repoMessages struct {
	*repository
}

func getMessageKey(teamID, peerID int64, peerType int32, msgID int64) []byte {
	sb := pools.AcquireStringsBuilder()
	sb.WriteString(prefixMessages)
	sb.WriteRune('.')
	tools.AppendStrInt64(sb, teamID)
	tools.AppendStrInt64(sb, peerID)
	tools.AppendStrInt32(sb, peerType)
	tools.AppendStrInt64(sb, msgID)
	id := tools.StrToByte(sb.String())
	pools.ReleaseStringsBuilder(sb)
	return id
}

func getMessagePrefix(teamID, peerID int64, peerType int32) []byte {
	sb := pools.AcquireStringsBuilder()
	sb.WriteString(prefixMessages)
	sb.WriteRune('.')
	tools.AppendStrInt64(sb, teamID)
	tools.AppendStrInt64(sb, peerID)
	tools.AppendStrInt32(sb, peerType)
	id := tools.StrToByte(sb.String())
	pools.ReleaseStringsBuilder(sb)
	return id
}

func getUserMessageKey(msgID int64) []byte {
	sb := pools.AcquireStringsBuilder()
	sb.WriteString(prefixUserMessages)
	sb.WriteRune('.')
	tools.AppendStrInt64(sb, msgID)
	id := tools.StrToByte(sb.String())
	pools.ReleaseStringsBuilder(sb)
	return id
}

func getMessageByID(txn *badger.Txn, msgID int64) (*msg.UserMessage, error) {
	message := new(msg.UserMessage)
	item, err := txn.Get(getUserMessageKey(msgID))
	if err != nil {
		return nil, err
	}
	err = item.Value(func(val []byte) error {
		parts := strings.Split(domain.ByteToStr(val), ".")
		if len(parts) != 3 {
			return domain.ErrInvalidUserMessageKey
		}
		itemMessage, err := txn.Get(getMessageKey(domain.StrToInt64(parts[0]), domain.StrToInt64(parts[1]), domain.StrToInt32(parts[2]), msgID))
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
	docType := msg.ClientMediaType_ClientMediaNone

	switch message.MediaType {
	case msg.MediaType_MediaTypeDocument:
		doc := &msg.MediaDocument{}
		_ = doc.Unmarshal(message.Media)
		if doc.Doc == nil {
			logs.Error("RepoMessage got error on save message, Document is Nil",
				zap.Int64("MessageID", message.ID),
				zap.Int64("SenderID", message.SenderID),
			)
			return nil
		}
		for _, da := range doc.Doc.Attributes {
			if docType == msg.ClientMediaType_ClientMediaGif {
				break
			}
			switch da.Type {
			case msg.DocumentAttributeType_AttributeTypeAudio:
				a := &msg.DocumentAttributeAudio{}
				_ = a.Unmarshal(da.Data)
				if a.Voice {
					docType = msg.ClientMediaType_ClientMediaVoice
				} else {
					docType = msg.ClientMediaType_ClientMediaAudio
				}
			case msg.DocumentAttributeType_AttributeTypeVideo, msg.DocumentAttributeType_AttributeTypePhoto:
				docType = msg.ClientMediaType_ClientMediaMedia
			case msg.DocumentAttributeType_AttributeTypeAnimated:
				docType = msg.ClientMediaType_ClientMediaGif
			case msg.DocumentAttributeType_AttributeTypeFile:
				if docType == msg.ClientMediaType_ClientMediaNone {
					docType = msg.ClientMediaType_ClientMediaFile
				}
			}
		}
	case msg.MediaType_MediaTypeWebDocument:
		webDoc := &msg.MediaWebDocument{}
		for _, da := range webDoc.Attributes {
			if docType == msg.ClientMediaType_ClientMediaGif {
				break
			}
			switch da.Type {
			case msg.DocumentAttributeType_AttributeTypeAudio:
				a := new(msg.DocumentAttributeAudio)
				_ = a.Unmarshal(da.Data)
				if a.Voice {
					docType = msg.ClientMediaType_ClientMediaVoice
				} else {
					docType = msg.ClientMediaType_ClientMediaAudio
				}
			case msg.DocumentAttributeType_AttributeTypeVideo, msg.DocumentAttributeType_AttributeTypePhoto:
				docType = msg.ClientMediaType_ClientMediaMedia
			case msg.DocumentAttributeType_AttributeTypeAnimated:
				docType = msg.ClientMediaType_ClientMediaGif
			case msg.DocumentAttributeType_AttributeTypeFile:
				if docType == msg.ClientMediaType_ClientMediaNone {
					docType = msg.ClientMediaType_ClientMediaFile
				}
			}
		}
	default:
		// Do nothing
	}

	// 1. Write Message
	err := txn.SetEntry(
		badger.NewEntry(
			getMessageKey(message.TeamID, message.PeerID, message.PeerType, message.ID),
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
			domain.StrToByte(fmt.Sprintf("%d.%d.%d", message.TeamID, message.PeerID, message.PeerType)),
		).WithMeta(byte(docType)),
	)
	if err != nil {
		return err
	}

	indexMessage(
		domain.ByteToStr(getMessageKey(message.TeamID, message.PeerID, message.PeerType, message.ID)),
		MessageSearch{
			Type:     "msg",
			Body:     message.Body,
			PeerID:   fmt.Sprintf("%d", message.PeerID),
			SenderID: fmt.Sprintf("%d", message.SenderID),
			TeamID:   fmt.Sprintf("%d", message.TeamID),
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

		dialog, err := getDialog(txn, message.TeamID, message.PeerID, message.PeerType)
		switch err {
		case nil:
		case badger.ErrKeyNotFound:
			logs.Info("We got new message but we don't have the dialog",
				zap.Int64("TeamID", message.TeamID),
				zap.Int64("MsgID", message.ID),
			)
			return nil
		default:
			return err

		}
		if message.ID > dialog.TopMessageID {
			dialog.TopMessageID = message.ID
			if !dialog.Pinned {
				err = updateDialogLastUpdate(message.TeamID, message.PeerID, message.PeerType, message.CreatedOn)
				if err != nil {
					logs.Info("We got error on update dialog last update",
						zap.Int64("TeamID", message.TeamID),
						zap.Int64("PeerID", message.PeerID),
						zap.Int64("MsgID", message.ID),
					)
				}
			}
			// Update counters if necessary
			if message.SenderID != userID {
				dialog.UnreadCount += 1
				for _, entity := range message.Entities {
					if (entity.Type == msg.MessageEntityType_MessageEntityTypeMention && entity.UserID == userID) ||
						(entity.Type == msg.MessageEntityType_MessageEntityTypeMentionAll) {
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

func (r *repoMessages) UpdateReactionCounter(messageID int64, reactions []*msg.ReactionCounter, yourReactions []string) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		um, err := getMessageByID(txn, messageID)
		if err != nil {
			return nil
		}
		um.Reactions = reactions
		um.YourReactions = yourReactions

		return saveMessage(txn, um)
	})
}

func (r *repoMessages) GetMessageHistory(
	teamID, peerID int64, peerType int32, minID, maxID int64, limit int32, filters ...msg.ClientMediaType,
) (userMessages []*msg.UserMessage, users []*msg.User, groups []*msg.Group) {
	userMessages = make([]*msg.UserMessage, 0, limit)
	switch {
	case maxID == 0 && minID == 0:
		dialog, err := Dialogs.Get(teamID, peerID, peerType)
		if err != nil {
			logs.Error("RepoMessage got error on GetHistory",
				zap.Error(err),
				zap.Int64("TeamID", teamID),
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
			opts.Prefix = getMessagePrefix(teamID, peerID, peerType)
			opts.Reverse = true
			it := txn.NewIterator(opts)
			it.Seek(getMessageKey(teamID, peerID, peerType, maxID))
			stopWatch1 = time.Now()
			for ; it.ValidForPrefix(opts.Prefix); it.Next() {
				if limit--; limit < 0 {
					break
				}
				_ = it.Item().Value(func(val []byte) error {
					userMessage := &msg.UserMessage{}
					err := userMessage.Unmarshal(val)
					if err != nil {
						return err
					}
					if len(filters) > 0 {
						if bypassFilter(it.Item().UserMeta(), filters...) {
							// increase the limit counter since we are not going to use this message
							limit++
							return nil
						}
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
			opts.Prefix = getMessagePrefix(teamID, peerID, peerType)
			opts.Reverse = false
			it := txn.NewIterator(opts)
			it.Seek(getMessageKey(teamID, peerID, peerType, minID))
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
					if len(filters) > 0 {
						if bypassFilter(it.Item().UserMeta(), filters...) {
							// increase the limit counter since we are not going to use this message
							limit++
							return nil
						}
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
			opts.Prefix = getMessagePrefix(teamID, peerID, peerType)
			opts.Reverse = true
			it := txn.NewIterator(opts)
			it.Seek(getMessageKey(teamID, peerID, peerType, maxID))
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
					if len(filters) > 0 {
						if bypassFilter(it.Item().UserMeta(), filters...) {
							// increase the limit counter since we are not going to use this message
							limit++
							return nil
						}
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

	users, groups = extractMessages(userMessages...)
	return
}

func bypassFilter(userMeta byte, filters ...msg.ClientMediaType) bool {
	byPass := true
	for _, f := range filters {
		if userMeta == byte(f) {
			byPass = false
			break
		}
	}
	return byPass
}
func extractMessages(msgs ...*msg.UserMessage) (users []*msg.User, groups []*msg.Group) {
	userIDs := domain.MInt64B{}
	groupIDs := domain.MInt64B{}
	for _, m := range msgs {
		if m.PeerType == int32(msg.PeerType_PeerSelf) || m.PeerType == int32(msg.PeerType_PeerUser) {
			userIDs[m.PeerID] = true
		}
		if m.PeerType == int32(msg.PeerType_PeerGroup) {
			groupIDs[m.PeerID] = true
		}
		if m.SenderID > 0 {
			userIDs[m.SenderID] = true
		}
		if m.FwdSenderID > 0 {
			userIDs[m.FwdSenderID] = true
		}
		for _, userID := range domain.ExtractActionUserIDs(m.MessageAction, m.MessageActionData) {
			userIDs[userID] = true
		}
	}
	users, _ = Users.GetMany(userIDs.ToArray())
	groups, _ = Groups.GetMany(groupIDs.ToArray())
	return
}

func (r *repoMessages) Delete(userID int64, teamID, peerID int64, peerType int32, msgIDs ...int64) {
	sort.Slice(msgIDs, func(i, j int) bool {
		return msgIDs[i] < msgIDs[j]
	})
	err := badgerUpdate(func(txn *badger.Txn) error {
		// Update the Dialog if necessary
		dialog, err := getDialog(txn, teamID, peerID, peerType)
		if err != nil {
			return err
		}

		for _, msgID := range msgIDs {
			if m, _ := getMessageByID(txn, msgID); m != nil {
				for _, labelID := range m.LabelIDs {
					_ = removeLabelFromMessage(txn, labelID, msgID)
					_ = decreaseLabelItemCount(txn, teamID, labelID)
				}
			}

			// delete from messages
			_ = txn.Delete(getMessageKey(teamID, peerID, peerType, msgID))

			// delete from user messages
			_ = txn.Delete(getUserMessageKey(msgID))
		}

		msgID := msgIDs[len(msgIDs)-1]
		if dialog.TopMessageID == msgID {
			opts := badger.DefaultIteratorOptions
			opts.Prefix = getMessagePrefix(dialog.TeamID, dialog.PeerID, dialog.PeerType)
			opts.Reverse = true
			it := txn.NewIterator(opts)
			it.Seek(getMessageKey(dialog.TeamID, dialog.PeerID, dialog.PeerType, dialog.TopMessageID))
			if it.ValidForPrefix(opts.Prefix) {
				userMessage := new(msg.UserMessage)
				_ = it.Item().Value(func(val []byte) error {
					return userMessage.Unmarshal(val)
				})
				dialog.TopMessageID = userMessage.ID
			}
			it.Close()
			if dialog.TopMessageID == msgID {
				_ = txn.Delete(getDialogKey(teamID, peerID, peerType))
				indexMessageRemove(domain.ByteToStr(getMessageKey(teamID, peerID, peerType, msgID)))
				return nil
			}
		}

		dialog.UnreadCount, dialog.MentionedCount, err = countDialogUnread(txn, dialog.TeamID, dialog.PeerID, dialog.PeerType, userID, dialog.ReadInboxMaxID+1)
		if err != nil {
			return err
		}
		err = saveDialog(txn, dialog)
		if err != nil {
			return err
		}

		indexMessageRemove(domain.ByteToStr(getMessageKey(teamID, peerID, peerType, msgID)))
		return nil
	})
	if err != nil {
		logs.Warn("RepoMessage got error on delete",
			zap.Error(err),
			zap.Int64("UserID", userID),
			zap.Int64("TeamID", teamID),
			zap.Int64("PeerID", peerID),
			zap.Int64s("MsgIDs", msgIDs),
		)
	}
}

func (r *repoMessages) ClearHistory(userID int64, teamID, peerID int64, peerType int32, maxID int64) error {
	err := badgerUpdate(func(txn *badger.Txn) error {
		st := r.badger.NewStream()
		st.Prefix = getMessagePrefix(teamID, peerID, peerType)
		st.NumGo = 10
		maxKey := getMessageKey(teamID, peerID, peerType, maxID)
		st.ChooseKey = func(item *badger.Item) bool {
			return bytes.Compare(item.Key(), maxKey) <= 0
		}
		st.Send = func(kvList *pb.KVList) error {
			for _, kv := range kvList.Kv {
				um, _ := getMessageByKey(txn, kv.Key)
				if um != nil {
					for _, labelID := range um.LabelIDs {
						_ = removeLabelFromMessage(txn, labelID, um.ID)
						_ = decreaseLabelItemCount(txn, teamID, labelID)
					}
				}
				err := txn.Delete(kv.Key)
				if err != nil {
					logs.Warn("RepoMessage got error on ClearHistory", zap.Error(err), zap.String("Key", domain.ByteToStr(kv.Key)))
				}
			}
			return nil
		}
		err := st.Orchestrate(context.Background())
		if err != nil {
			return err
		}
		dialog, err := getDialog(txn, teamID, peerID, peerType)
		if err != nil {
			return err
		}
		dialog.UnreadCount, dialog.MentionedCount, err = countDialogUnread(txn, teamID, peerID, peerType, userID, maxID)
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

func (r *repoMessages) GetTopMessageID(teamID, peerID int64, peerType int32) (int64, error) {
	topMessageID := int64(0)
	err := badgerView(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = getMessagePrefix(teamID, peerID, peerType)
		opts.Reverse = true
		it := txn.NewIterator(opts)
		it.Seek(getMessageKey(teamID, peerID, peerType, 2<<31))
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

func (r *repoMessages) SearchText(teamID int64, text string, limit int32) []*msg.UserMessage {
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
	t3 := bleve.NewTermQuery(fmt.Sprintf("%d", tools.AbsInt64(teamID)))
	t3.SetField("team_id")
	searchRequest := bleve.NewSearchRequest(bleve.NewConjunctionQuery(t1, t2, t3))
	searchResult, _ := r.msgSearch.Search(searchRequest)
	searchRequest.Size = int(limit)
	_ = badgerView(func(txn *badger.Txn) error {
		for _, hit := range searchResult.Hits {
			userMessage, _ := getMessageByKey(txn, domain.StrToByte(hit.ID))
			if userMessage != nil && userMessage.TeamID == teamID {
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

func (r *repoMessages) SearchTextByPeerID(teamID int64, text string, peerID int64, limit int32) []*msg.UserMessage {
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
	t3 := bleve.NewTermQuery(fmt.Sprintf("%d", tools.AbsInt64(peerID)))
	t3.SetField("peer_id")
	t4 := bleve.NewTermQuery(fmt.Sprintf("%d", tools.AbsInt64(teamID)))
	t4.SetField("team_id")
	searchRequest := bleve.NewSearchRequest(bleve.NewConjunctionQuery(t1, t2, t3, t4))
	searchResult, _ := r.msgSearch.Search(searchRequest)
	_ = badgerView(func(txn *badger.Txn) error {
		for _, hit := range searchResult.Hits {
			userMessage, _ := getMessageByKey(txn, domain.StrToByte(hit.ID))
			if userMessage != nil && userMessage.TeamID == teamID {
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

func (r *repoMessages) SearchByLabels(teamID int64, labelIDs []int32, peerID int64, limit int32) []*msg.UserMessage {
	userMessages := make([]*msg.UserMessage, 0, limit)
	_ = badgerView(func(txn *badger.Txn) error {
		st := r.badger.NewStream()
		st.Prefix = domain.StrToByte(prefixMessages)
		st.ChooseKey = func(item *badger.Item) bool {
			m := &msg.UserMessage{}
			err := item.Value(func(val []byte) error {
				return m.Unmarshal(val)
			})
			if err != nil {
				return false
			}
			if m.TeamID != teamID {
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

func (r *repoMessages) GetAllMedia(documentType msg.ClientMediaType) ([]*msg.UserMessage, error) {
	limit := 500
	msgMtx := sync.Mutex{}
	userMessages := make([]*msg.UserMessage, 0, limit)

	stream := r.badger.NewStream()
	stream.Prefix = domain.StrToByte(fmt.Sprintf("%s.", prefixMessages))
	stream.ChooseKey = func(item *badger.Item) bool {
		if item.UserMeta() == byte(documentType) {
			m := &msg.UserMessage{}
			err := item.Value(func(val []byte) error {
				return m.Unmarshal(val)
			})
			if err != nil {
				return false
			}

			msgMtx.Lock()
			userMessages = append(userMessages, m)
			msgMtx.Unlock()
			return true
		}

		return false
	}

	stream.Send = func(list *pb.KVList) error {
		return nil
	}

	_ = stream.Orchestrate(context.Background())

	return userMessages, nil
}

func (r *repoMessages) SearchBySender(teamID int64, text string, senderID int64, peerID int64, limit int32) []*msg.UserMessage {
	userMessages := make([]*msg.UserMessage, 0, limit)

	if r.msgSearch == nil {
		return userMessages
	}

	t1 := bleve.NewTermQuery("msg")
	t1.SetField("type")

	var t2 *query.DisjunctionQuery
	if len(text) != 0 {
		qs := make([]query.Query, 0)
		for _, term := range strings.Fields(text) {
			qs = append(qs, bleve.NewMatchQuery(term), bleve.NewPrefixQuery(term), bleve.NewFuzzyQuery(term))
		}
		t2 = bleve.NewDisjunctionQuery(qs...)
	}

	t3 := bleve.NewTermQuery(fmt.Sprintf("%d", tools.AbsInt64(peerID)))
	t3.SetField("peer_id")

	t4 := bleve.NewTermQuery(fmt.Sprintf("%d", tools.AbsInt64(senderID)))
	t4.SetField("sender_id")

	t5 := bleve.NewTermQuery(fmt.Sprintf("%d", tools.AbsInt64(teamID)))
	t5.SetField("team_id")

	var searchRequest *bleve.SearchRequest
	if t2 != nil {
		searchRequest = bleve.NewSearchRequest(bleve.NewConjunctionQuery(t1, t2, t3, t4, t5))
	} else {
		searchRequest = bleve.NewSearchRequest(bleve.NewConjunctionQuery(t1, t3, t4, t5))
	}

	searchRequest.Size = int(limit)
	searchRequest.SortBy([]string{"_id"})
	searchResult, _ := r.msgSearch.Search(searchRequest)
	_ = badgerView(func(txn *badger.Txn) error {
		for _, hit := range searchResult.Hits {
			userMessage, _ := getMessageByKey(txn, domain.StrToByte(hit.ID))
			if userMessage != nil && userMessage.TeamID == teamID {
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

func (r *repoMessages) GetLastBotKeyboard(teamID, peerID int64, peerType int32) (*msg.UserMessage, error) {
	limit := 1000

	var keyboardMessage *msg.UserMessage
	stop := false
	_ = badgerView(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		opts.Prefix = getMessagePrefix(teamID, peerID, peerType)
		opts.Reverse = true
		it := txn.NewIterator(opts)
		for it.Seek(getMessageKey(teamID, peerID, peerType, 1<<31)); it.ValidForPrefix(opts.Prefix); it.Next() {
			if limit--; limit < 0 || stop {
				break
			}

			_ = it.Item().Value(func(val []byte) error {
				userMessage := new(msg.UserMessage)
				err := userMessage.Unmarshal(val)
				if err != nil {
					return err
				}

				if userMessage.SenderID == userMessage.PeerID {
					if userMessage.ReplyMarkup == msg.C_ReplyKeyboardMarkup || userMessage.ReplyMarkup == msg.C_ReplyKeyboardForceReply {
						keyboardMessage = userMessage
					}
					stop = true
				}
				return nil
			})
		}
		it.Close()
		return nil
	})

	return keyboardMessage, nil
}

func (r *repoMessages) ReIndex() {
	err := domain.Try(10, time.Second, func() error {
		if r.msgSearch == nil {
			return domain.ErrDoesNotExists
		}
		return nil
	})
	if err != nil {
		return
	}
	err = badgerView(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = domain.StrToByte(prefixMessages)
		it := txn.NewIterator(opts)
		for it.Rewind(); it.Valid(); it.Next() {
			_ = it.Item().Value(func(val []byte) error {
				message := &msg.UserMessage{}
				_ = message.Unmarshal(val)
				msgKey := domain.ByteToStr(getMessageKey(message.TeamID, message.PeerID, message.PeerType, message.ID))
				if d, _ := r.msgSearch.Document(msgKey); d == nil {
					indexMessage(
						msgKey,
						MessageSearch{
							Type:     "msg",
							Body:     message.Body,
							PeerID:   fmt.Sprintf("%d", message.PeerID),
							SenderID: fmt.Sprintf("%d", message.SenderID),
							TeamID:   fmt.Sprintf("%d", message.TeamID),
						},
					)
				}
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
