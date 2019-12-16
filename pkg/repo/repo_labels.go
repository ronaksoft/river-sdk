package repo

import (
	"context"
	"encoding/binary"
	"fmt"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/pb"
	"go.uber.org/zap"
	"sort"
	"strings"
	"time"
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
	prefixLabelHoles    = "LBLH"
)

type repoLabels struct {
	*repository
}

func getLabelKey(labelID int32) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%03d", prefixLabel, labelID))
}

func getLabelMessageKey(labelID int32, msgID int64) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%03d.%021d", prefixLabelMessages, labelID, msgID))
}

func getLabelByID(txn *badger.Txn, labelID int32) (*msg.Label, error) {
	return getLabelByKey(txn, getLabelKey(labelID))
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

func deleteLabel(txn *badger.Txn, labelID int32) error {
	return txn.Delete(getLabelKey(labelID))
}

func addLabelToMessage(txn *badger.Txn, labelID int32, peerType int32, peerID int64, msgID int64) error {
	err := txn.SetEntry(badger.NewEntry(
		getLabelMessageKey(labelID, msgID),
		ronak.StrToByte(fmt.Sprintf("%d.%d.%d", peerType, peerID, msgID)),
	))
	return err
}

func removeLabelFromMessage(txn *badger.Txn, labelID int32, msgID int64) error {
	err := txn.Delete(ronak.StrToByte(fmt.Sprintf("%s.03%d.021%d", prefixLabelMessages, labelID, msgID)))
	switch err {
	case badger.ErrKeyNotFound:
		return nil
	}
	return err
}

func (r *repoLabels) Save(labels ...*msg.Label) error {
	err := badgerUpdate(func(txn *badger.Txn) error {
		for _, l := range labels {
			err := saveLabel(txn, l)
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
			stream.Prefix = ronak.StrToByte(fmt.Sprintf("%s.%03d.", prefixLabelMessages, labelID))
			stream.Send = func(list *pb.KVList) error {
				for _, kv := range list.Kv {
					parts := strings.Split(ronak.ByteToStr(kv.Value), ".")
					if len(parts) != 3 {
						return domain.ErrInvalidData
					}
					msgID := ronak.StrToInt64(parts[2])
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

func (r *repoLabels) GetMany(labelIDs ...int32) []*msg.Label {
	labels := make([]*msg.Label, 0, len(labelIDs))
	_ = badgerView(func(txn *badger.Txn) error {
		for _, labelID := range labelIDs {
			l, err := getLabelByID(txn, labelID)
			if err == nil {
				labels = append(labels, l)
			}
		}
		return nil
	})
	return labels
}

func (r *repoLabels) GetAll() []*msg.Label {
	labels := make([]*msg.Label, 0, 20)
	err := badgerView(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = ronak.StrToByte(prefixLabel)
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.ValidForPrefix(opts.Prefix); it.Next() {
			_ = it.Item().Value(func(val []byte) error {
				l := &msg.Label{}
				err := l.Unmarshal(val)
				if err != nil {
					return err
				}
				labels = append(labels, l)
				return nil
			})
		}

		return nil
	})
	logs.ErrorOnErr("RepoLabels got error on GetAll", err)
	return labels
}

func (r *repoLabels) ListMessages(labelID int32, limit int32, minID, maxID int64) ([]*msg.UserMessage, []*msg.User, []*msg.Group) {
	userMessages := make([]*msg.UserMessage, 0, limit)
	userIDs := domain.MInt64B{}
	switch {
	case maxID == 0 && minID == 0:
		fallthrough
	case maxID != 0 && minID == 0:
		startTime := time.Now()
		var stopWatch1, stopWatch2 time.Time
		err := badgerView(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.Prefix = ronak.StrToByte(fmt.Sprintf("%s.%03d", prefixLabelMessages, labelID))
			if maxID > 0 {
				opts.Reverse = true
			}
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
				_ = it.Item().Value(func(val []byte) error {
					parts := strings.Split(ronak.ByteToStr(val), ".")
					if len(parts) != 3 {
						return domain.ErrInvalidData
					}
					msgID := ronak.StrToInt64(parts[2])

					um, err := getMessageByID(txn, msgID)
					if err != nil {
						return err
					}
					userIDs.Add(um.SenderID)
					if um.FwdSenderID != 0 {
						userIDs.Add(um.FwdSenderID)
					}

					userIDs.Add(domain.ExtractActionUserIDs(um.MessageAction, um.MessageActionData)...)
					userMessages = append(userMessages, um)
					return nil
				})
			}
			return nil
		})
		logs.WarnOnErr("RepoLabels got error on ListMessages", err)
		logs.Info("RepoLabels got list", zap.Int64("MinID", minID), zap.Int64("MaxID", maxID),
			zap.Duration("SP1", stopWatch1.Sub(startTime)),
			zap.Duration("SP2", stopWatch2.Sub(startTime)),
		)
	case maxID == 0 && minID != 0:
		startTime := time.Now()
		var stopWatch1, stopWatch2, stopWatch3 time.Time
		_ = badgerView(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.Prefix = ronak.StrToByte(fmt.Sprintf("%s.%03d", prefixLabelMessages, labelID))
			it := txn.NewIterator(opts)
			defer it.Close()
			it.Seek(getLabelMessageKey(labelID, minID))
			stopWatch1 = time.Now()
			for ; it.ValidForPrefix(opts.Prefix); it.Next() {
				if limit--; limit < 0 {
					break
				}
				_ = it.Item().Value(func(val []byte) error {
					parts := strings.Split(ronak.ByteToStr(val), ".")
					if len(parts) != 3 {
						return domain.ErrInvalidData
					}
					msgID := ronak.StrToInt64(parts[2])
					um, err := getMessageByID(txn, msgID)
					if err != nil {
						return err
					}
					userIDs.Add(um.SenderID)
					if um.FwdSenderID != 0 {
						userIDs.Add(um.FwdSenderID)
					}
					userIDs.Add(domain.ExtractActionUserIDs(um.MessageAction, um.MessageActionData)...)

					userMessages = append(userMessages, um)
					return nil
				})
			}
			stopWatch2 = time.Now()
			sort.Slice(userMessages, func(i, j int) bool {
				return userMessages[i].ID > userMessages[j].ID
			})
			stopWatch3 = time.Now()
			return nil
		})
		logs.Info("RepoLabels got list", zap.Int64("MinID", minID), zap.Int64("MaxID", maxID),
			zap.Duration("SP1", stopWatch1.Sub(startTime)),
			zap.Duration("SP2", stopWatch2.Sub(startTime)),
			zap.Duration("SP3", stopWatch3.Sub(startTime)),
		)
	default:
	}

	users := Users.GetMany(userIDs.ToArray())
	return userMessages, users
}

func (r *repoLabels) AddLabelsToMessages(labelIDs []int32, peerType int32, peerID int64, msgIDs []int64) error {
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
			um, err := getMessageByKey(txn, getMessageKey(peerID, peerType, msgID))
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

func (r *repoLabels) RemoveLabelsFromMessages(labelIDs []int32, peerType int32, peerID int64, msgIDs []int64) error {
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
			um, err := getMessageByKey(txn, getMessageKey(peerID, peerType, msgID))
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

type Bar struct {
	MinID int64
	MaxID int64
}

func (r *repoLabels) Fill(labelID int32, minID, maxID int64) error {
	minIDb := make([]byte, 8)
	binary.BigEndian.PutUint64(minIDb, uint64(minID))
	maxIDb := make([]byte, 8)
	binary.BigEndian.PutUint64(maxIDb, uint64(maxID))
	bar := r.GetFilled(labelID)
	if maxID > bar.MaxID {
		_ = badgerUpdate(func(txn *badger.Txn) error {
			return txn.SetEntry(badger.NewEntry(
				ronak.StrToByte(fmt.Sprintf("%s.03%d.MAXID", prefixLabelMessages, labelID)),
				maxIDb,
			))
		})
	}

	if bar.MinID == 0 || minID < bar.MinID {
		_ = badgerUpdate(func(txn *badger.Txn) error {
			return txn.SetEntry(badger.NewEntry(
				ronak.StrToByte(fmt.Sprintf("%s.03%d.MINID", prefixLabelMessages, labelID)),
				minIDb,
			))
		})
	}

	return nil
}

func (r *repoLabels) GetFilled(labelID int32) Bar {
	bar := Bar{}
	_ = badgerView(func(txn *badger.Txn) error {
		minIDItem, err := txn.Get(ronak.StrToByte(fmt.Sprintf("%s.03%d.MINID", prefixLabelMessages, labelID)))
		if err != nil {
			return err
		}
		_ = minIDItem.Value(func(val []byte) error {
			bar.MinID = int64(binary.BigEndian.Uint64(val))
			return nil
		})
		maxIDItem, err := txn.Get(ronak.StrToByte(fmt.Sprintf("%s.03%d.MAXID", prefixLabelMessages, labelID)))
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

func (r *repoLabels) GetLowerFilled(labelID int32, maxID int64) (bool, Bar) {
	b := r.GetFilled(labelID)
	if maxID > b.MaxID || maxID < b.MinID {
		return false, Bar{}
	}
	b.MaxID = maxID
	return true, b
}

func (r *repoLabels) GetUpperFilled(labelID int32, minID int64) (bool, Bar) {
	b := r.GetFilled(labelID)
	if minID < b.MinID || minID > b.MaxID {
		return false, Bar{}
	}
	b.MinID = minID
	return true, b
}
