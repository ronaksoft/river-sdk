package repo

import (
	"context"
	"fmt"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/pb"
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
)

type repoLabels struct {
	*repository
}

func getLabelKey(labelID int32) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%012d", prefixLabel, labelID))
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
		ronak.StrToByte(fmt.Sprintf("%s.03%d.021%d", prefixLabelMessages, labelID, msgID)),
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
		for it.Rewind(); it.ValidForPrefix(opts.Prefix); it.Next() {
			err := it.Item().Value(func(val []byte) error {
				l := &msg.Label{}
				err := l.Unmarshal(val)
				if err != nil {
					return err
				}
				labels = append(labels, l)
				return nil
			})
			if err != nil {
				return err
			}
		}
		it.Close()
		return nil
	})
	logs.ErrorOnErr("RepoLabels got error on GetAll", err)
	return labels
}

func (r *repoLabels) ListMessages(labelID int32, limit int32, minID, maxID int64) []*msg.UserMessage {
	messages := make([]*msg.UserMessage, 0, 10)
	badgerView(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = ronak.StrToByte(fmt.Sprintf("%s.%03d.", prefixLabelMessages, labelID))
		it := txn.NewIterator(opts)
		for it.Rewind(); it.ValidForPrefix(opts.Prefix); it.Next() {
			err := it.Item().Value(func(val []byte) error {
				parts := strings.Split(ronak.ByteToStr(val), ".")
				if len(parts) != 3 {
					return domain.ErrInvalidData
				}
				peerType := ronak.StrToInt32(parts[0])
				peerID := ronak.StrToInt64(parts[1])
				msgID := ronak.StrToInt64(parts[2])
				um, err := getMessageByKey(txn, getMessageKey(peerID, peerType, msgID))
				if err != nil {
					return err
				}
				messages = append(messages, um)
				return nil
			})
			if err != nil {
				return err
			}
		}
		it.Close()
		return nil
	})
	return messages
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