package repo

import (
	"context"
	"encoding/binary"
	"fmt"
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/z"
	"github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/badger/v2/pb"
	"github.com/ronaksoft/rony/pools"
	"github.com/ronaksoft/rony/tools"
	"go.uber.org/zap"
	"sort"
	"strings"
)

/*
   Creation Time: 2019 - Dec - 08
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

const (
	prefixLabel         = "LBL"
	prefixLabelMessages = "LBLM"
	prefixLabelFill     = "LBLF"
	prefixLabelCount    = "LBLC"
)

type repoLabels struct {
	*repository
}

func getLabelKey(labelID int32) []byte {
	sb := pools.AcquireStringsBuilder()
	sb.WriteString(prefixLabel)
	sb.WriteRune('.')
	z.AppendStrInt32(sb, labelID)
	id := tools.StrToByte(sb.String())
	pools.ReleaseStringsBuilder(sb)
	return id
}

func getLabelCountKey(teamID int64, labelID int32) []byte {
	sb := pools.AcquireStringsBuilder()
	sb.WriteString(prefixLabelCount)
	sb.WriteRune('.')
	z.AppendStrInt32(sb, labelID)
	z.AppendStrInt64(sb, teamID)
	id := tools.StrToByte(sb.String())
	pools.ReleaseStringsBuilder(sb)
	return id
}

func getLabelCountPrefix(labelID int32) []byte {
	sb := pools.AcquireStringsBuilder()
	sb.WriteString(prefixLabelCount)
	sb.WriteRune('.')
	z.AppendStrInt32(sb, labelID)
	id := tools.StrToByte(sb.String())
	pools.ReleaseStringsBuilder(sb)
	return id
}

func getLabelMessageKey(labelID int32, msgID int64) []byte {
	sb := pools.AcquireStringsBuilder()
	sb.WriteString(prefixLabelMessages)
	sb.WriteRune('.')
	z.AppendStrInt32(sb, labelID)
	z.AppendStrInt64(sb, msgID)
	id := tools.StrToByte(sb.String())
	pools.ReleaseStringsBuilder(sb)
	return id
}

func getLabelMessagePrefix(labelID int32) []byte {
	sb := pools.AcquireStringsBuilder()
	sb.WriteString(prefixLabelMessages)
	sb.WriteRune('.')
	z.AppendStrInt32(sb, labelID)
	id := tools.StrToByte(sb.String())
	pools.ReleaseStringsBuilder(sb)
	return id
}

func getLabelByID(txn *badger.Txn, teamID int64, labelID int32) (*msg.Label, error) {
	l, err := getLabelByKey(txn, getLabelKey(labelID))
	if err != nil {
		return nil, err
	}
	l.Count = getLabelCount(txn, teamID, labelID)
	return l, nil
}

func getLabelByKey(txn *badger.Txn, key []byte) (*msg.Label, error) {
	label := &msg.Label{}
	item, err := txn.Get(key)
	if err != nil {
		return nil, err
	}
	err = item.Value(func(val []byte) error {
		return label.Unmarshal(val)
	})
	if err != nil {
		return nil, err
	}
	return label, nil
}

func saveLabel(txn *badger.Txn, label *msg.Label) error {
	labelBytes, _ := label.Marshal()
	err := txn.SetEntry(badger.NewEntry(
		getLabelKey(label.ID),
		labelBytes,
	))
	return err
}

func saveLabelCount(txn *badger.Txn, teamID int64, labelID int32, cnt int32) error {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(cnt))
	return txn.Set(getLabelCountKey(teamID, labelID), b)
}

func getLabelCount(txn *badger.Txn, teamID int64, labelID int32) int32 {
	var cnt int32
	item, err := txn.Get(getLabelCountKey(teamID, labelID))
	if err != nil {
		return 0
	}
	_ = item.Value(func(val []byte) error {
		cnt = int32(binary.BigEndian.Uint32(val))
		return nil
	})
	return cnt
}

func deleteLabel(txn *badger.Txn, labelID int32) error {
	opts := badger.DefaultIteratorOptions
	opts.Prefix = getLabelCountPrefix(labelID)
	it := txn.NewIterator(opts)
	for it.Rewind(); it.ValidForPrefix(opts.Prefix); it.Next() {
		_ = txn.Delete(it.Item().KeyCopy(nil))
	}
	it.Close()
	_ = txn.Delete(getLabelKey(labelID))
	return nil
}

func addLabelToMessage(txn *badger.Txn, labelID int32, peerType int32, peerID int64, msgID int64) error {
	err := txn.SetEntry(badger.NewEntry(
		getLabelMessageKey(labelID, msgID),
		tools.StrToByte(fmt.Sprintf("%d.%d.%d", peerType, peerID, msgID)),
	))
	return err
}

func removeLabelFromMessage(txn *badger.Txn, labelID int32, msgID int64) error {
	err := txn.Delete(getLabelMessageKey(labelID, msgID))
	switch err {
	case badger.ErrKeyNotFound:
		return nil
	}
	return err
}

func decreaseLabelItemCount(txn *badger.Txn, teamID int64, labelID int32) error {
	cnt := getLabelCount(txn, teamID, labelID)
	if cnt == 0 {
		logs.Warn("RepoLabel tried to decrement counter but it is zero")
		return nil
	}
	cnt--
	return saveLabelCount(txn, teamID, labelID, cnt)
}

func (r *repoLabels) Set(labels ...*msg.Label) error {
	err := badgerUpdate(func(txn *badger.Txn) error {
		for _, l := range labels {
			err := saveLabel(txn, l)
			if err != nil {
				return err
			}
		}
		return nil
	})
	logs.ErrorOnErr("RepoLabel got error on Set", err)
	return err
}

func (r *repoLabels) Save(teamID int64, labels ...*msg.Label) error {
	err := badgerUpdate(func(txn *badger.Txn) error {
		for _, l := range labels {
			err := saveLabel(txn, l)
			if err != nil {
				return err
			}
			err = saveLabelCount(txn, teamID, l.ID, l.Count)
			if err != nil {
				return err
			}
		}
		return nil
	})
	logs.ErrorOnErr("RepoLabel got error on Save", err)
	return err
}

func (r *repoLabels) Delete(labelIDs ...int32) error {
	err := badgerUpdate(func(txn *badger.Txn) error {
		for _, labelID := range labelIDs {
			err := deleteLabel(txn, labelID)
			if err != nil {
				return err
			}
			stream := r.badger.NewStream()
			stream.Prefix = getLabelMessagePrefix(labelID)
			stream.Send = func(list *pb.KVList) error {
				for _, kv := range list.Kv {
					parts := strings.Split(tools.ByteToStr(kv.Value), ".")
					if len(parts) != 3 {
						return domain.ErrInvalidData
					}
					msgID := tools.StrToInt64(parts[2])
					_ = removeLabelFromMessage(txn, labelID, msgID)
				}
				return nil
			}
			err = stream.Orchestrate(context.Background())
			if err != nil {
				return err
			}
		}
		return nil
	})
	logs.ErrorOnErr("RepoLabel got error on Delete", err)
	return err

}

func (r *repoLabels) GetMany(teamID int64, labelIDs ...int32) []*msg.Label {
	labels := make([]*msg.Label, 0, len(labelIDs))
	_ = badgerView(func(txn *badger.Txn) error {
		for _, labelID := range labelIDs {
			l, err := getLabelByID(txn, teamID, labelID)
			if err == nil {
				labels = append(labels, l)
			}
		}
		return nil
	})
	return labels
}

func (r *repoLabels) GetAll(teamID int64) []*msg.Label {
	labels := make([]*msg.Label, 0, 20)
	err := badgerView(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = tools.StrToByte(prefixLabel)
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.ValidForPrefix(opts.Prefix); it.Next() {
			_ = it.Item().Value(func(val []byte) error {
				l := &msg.Label{}
				err := l.Unmarshal(val)
				if err != nil {
					return err
				}

				l.Count = getLabelCount(txn, teamID, l.ID)
				labels = append(labels, l)
				return nil
			})
		}

		return nil
	})
	logs.ErrorOnErr("RepoLabels got error on GetAll", err)
	return labels
}

func (r *repoLabels) ListMessages(labelID int32, teamID int64, limit int32, minID, maxID int64) ([]*msg.UserMessage, []*msg.User, []*msg.Group) {
	userMessages := make([]*msg.UserMessage, 0, limit)
	userIDs := make(domain.MInt64B, limit)
	groupIDs := make(domain.MInt64B, limit)

	opts := badger.DefaultIteratorOptions
	opts.Prefix = getLabelMessagePrefix(labelID)
	switch {
	case maxID == 0 && minID == 0:
		fallthrough
	case maxID > 0:
		if maxID > 0 {
			opts.Reverse = true
		}
		err := badgerView(func(txn *badger.Txn) error {
			it := txn.NewIterator(opts)
			defer it.Close()
			if maxID > 0 {
				it.Seek(getLabelMessageKey(labelID, maxID))
			} else {
				it.Rewind()
			}
			for ; it.ValidForPrefix(opts.Prefix); it.Next() {
				if limit--; limit < 0 {
					break
				}
				err := it.Item().Value(func(val []byte) error {
					return extractMessage(txn, val, teamID, &userMessages, userIDs, groupIDs)
				})
				logs.WarnOnErr("RepoLabels got error on ListMessage for getting message", err)
			}
			return nil
		})
		logs.WarnOnErr("RepoLabels got error on ListMessages", err)
		logs.Debug("RepoLabels got list", zap.Int64("MinID", minID), zap.Int64("MaxID", maxID))
	case minID != 0:
		_ = badgerView(func(txn *badger.Txn) error {
			it := txn.NewIterator(opts)
			defer it.Close()
			it.Seek(getLabelMessageKey(labelID, minID))
			for ; it.ValidForPrefix(opts.Prefix); it.Next() {
				if limit--; limit < 0 {
					break
				}
				_ = it.Item().Value(func(val []byte) error {
					return extractMessage(txn, val, teamID, &userMessages, userIDs, groupIDs)
				})
			}
			return nil
		})

	default:
	}
	sort.Slice(userMessages, func(i, j int) bool {
		return userMessages[i].ID < userMessages[j].ID
	})
	users, _ := Users.GetMany(userIDs.ToArray())
	groups, _ := Groups.GetMany(groupIDs.ToArray())
	return userMessages, users, groups
}
func extractMessage(txn *badger.Txn, val []byte, teamID int64, userMessages *[]*msg.UserMessage, userIDs, groupIDs domain.MInt64B ) error {
	parts := strings.Split(tools.ByteToStr(val), ".")
	if len(parts) != 3 {
		return domain.ErrInvalidData
	}
	msgID := tools.StrToInt64(parts[2])

	um, err := getMessageByID(txn, msgID)
	if err != nil {
		return err
	}
	if um.TeamID != teamID {
		return nil
	}
	userIDs.Add(um.SenderID)
	if um.FwdSenderID != 0 {
		userIDs.Add(um.FwdSenderID)
	}
	switch msg.PeerType(um.PeerType) {
	case msg.PeerType_PeerUser:
		userIDs.Add(um.PeerID)
	case msg.PeerType_PeerGroup:
		groupIDs.Add(um.PeerID)
	}

	userIDs.Add(domain.ExtractActionUserIDs(um.MessageAction, um.MessageActionData)...)

	*userMessages = append(*userMessages, um)
	return nil
}

func (r *repoLabels) AddLabelsToMessages(labelIDs []int32, teamID, peerID int64, peerType int32, msgIDs []int64) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		for _, labelID := range labelIDs {
			for _, msgID := range msgIDs {
				err := addLabelToMessage(txn, labelID, peerType, peerID, msgID)
				if err != nil {
					return err
				}
			}
		}
		for _, msgID := range msgIDs {
			um, err := getMessageByKey(txn, getMessageKey(teamID, peerID, peerType, msgID))
			if err != nil {
				switch err {
				case badger.ErrKeyNotFound:
					continue
				default:
					return err
				}
			}
			m := domain.MInt32B{}
			m.Add(um.LabelIDs...)
			m.Add(labelIDs...)
			um.LabelIDs = m.ToArray()
			err = saveMessage(txn, um)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *repoLabels) RemoveLabelsFromMessages(labelIDs []int32, teamID, peerID int64, peerType int32, msgIDs []int64) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		for _, labelID := range labelIDs {
			for _, msgID := range msgIDs {
				err := removeLabelFromMessage(txn, labelID, msgID)
				if err != nil {
					return err
				}
			}
		}
		for _, msgID := range msgIDs {
			um, err := getMessageByKey(txn, getMessageKey(teamID, peerID, peerType, msgID))
			if err != nil {
				switch err {
				case badger.ErrKeyNotFound:
					continue
				default:
					return err
				}
			}
			m := domain.MInt32B{}
			m.Add(um.LabelIDs...)
			m.Remove(labelIDs...)
			um.LabelIDs = m.ToArray()
			err = saveMessage(txn, um)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

type LabelBar struct {
	MinID int64
	MaxID int64
}

func getLabelBarMaxKey(teamID int64, labelID int32) []byte {
	return tools.StrToByte(fmt.Sprintf("%s.%021d.03%d.MAXID", prefixLabelFill, teamID, labelID))
}

func getLabelBarMinKey(teamID int64, labelID int32) []byte {
	return tools.StrToByte(fmt.Sprintf("%s.%021d.03%d.MINID", prefixLabelFill, teamID, labelID))
}

func (r *repoLabels) Fill(teamID int64, labelID int32, minID, maxID int64) error {
	minIDb := make([]byte, 8)
	binary.BigEndian.PutUint64(minIDb, uint64(minID))
	maxIDb := make([]byte, 8)
	binary.BigEndian.PutUint64(maxIDb, uint64(maxID))
	bar := r.GetFilled(teamID, labelID)
	if maxID > bar.MaxID {
		_ = badgerUpdate(func(txn *badger.Txn) error {
			return txn.SetEntry(badger.NewEntry(
				getLabelBarMaxKey(teamID, labelID),
				maxIDb,
			))
		})
	}

	if bar.MinID == 0 || minID < bar.MinID {
		_ = badgerUpdate(func(txn *badger.Txn) error {
			return txn.SetEntry(badger.NewEntry(
				getLabelBarMinKey(teamID, labelID),
				minIDb,
			))
		})
	}

	return nil
}

func (r *repoLabels) GetFilled(teamID int64, labelID int32) LabelBar {
	bar := LabelBar{}
	_ = badgerView(func(txn *badger.Txn) error {
		minIDItem, err := txn.Get(getLabelBarMinKey(teamID, labelID))
		if err != nil {
			return err
		}
		_ = minIDItem.Value(func(val []byte) error {
			bar.MinID = int64(binary.BigEndian.Uint64(val))
			return nil
		})
		maxIDItem, err := txn.Get(getLabelBarMaxKey(teamID, labelID))
		if err != nil {
			return err
		}
		_ = maxIDItem.Value(func(val []byte) error {
			bar.MaxID = int64(binary.BigEndian.Uint64(val))
			return nil
		})
		return nil
	})
	return bar
}

func (r *repoLabels) GetLowerFilled(teamID int64, labelID int32, maxID int64) (bool, LabelBar) {
	b := r.GetFilled(teamID, labelID)
	if b.MinID == 0 && b.MaxID == 0 {
		return false, b
	}
	if maxID > b.MaxID || maxID < b.MinID {
		return false, b
	}
	b.MaxID = maxID
	return true, b
}

func (r *repoLabels) GetUpperFilled(teamID int64, labelID int32, minID int64) (bool, LabelBar) {
	b := r.GetFilled(teamID, labelID)
	if b.MinID == 0 && b.MaxID == 0 {
		return false, b
	}
	if minID < b.MinID || minID > b.MaxID {
		return false, b
	}
	b.MinID = minID
	return true, b
}
