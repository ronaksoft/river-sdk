package repo

import (
	"encoding/json"
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo/dto"
	"github.com/dgraph-io/badger"
)

const (
	prefixMessageExtra = "MSG_EX"
)

type repoMessagesExtra struct {
	*repository
}

func (r *repoMessagesExtra) getKey(peerID int64, peerType int32) []byte {
	return domain.StrToByte(fmt.Sprintf("%s.%021d.%d", prefixMessageExtra, peerID, peerType))
}

func (r *repoMessagesExtra) get(peerID int64, peerType int32) *dto.MessagesExtra {
	message := new(dto.MessagesExtra)
	_ = badgerView(func(txn *badger.Txn) error {
		item, err := txn.Get(r.getKey(peerID, peerType))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, message)
		})
	})
	return message
}

func (r *repoMessagesExtra) save(key []byte, m *dto.MessagesExtra) {
	bytes, _ := json.Marshal(m)
	_ = badgerUpdate(func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry(key, bytes))
	})
}

func (r *repoMessagesExtra) SaveScrollID(peerID int64, peerType int32, msgID int64) {
	key := r.getKey(peerID, peerType)
	m := r.get(peerID, peerType)
	if m == nil {
		return
	}
	m.ScrollID = msgID
	r.save(key, m)
}

func (r *repoMessagesExtra) GetScrollID(peerID int64, peerType int32) int64 {
	m := r.get(peerID, peerType)
	if m == nil {
		return 0
	}
	return m.ScrollID
}

func (r *repoMessagesExtra) SaveHoles(peerID int64, peerType int32, data []byte) {
	key := r.getKey(peerID, peerType)
	m := r.get(peerID, peerType)
	if m == nil {
		return
	}
	m.Holes = data
	r.save(key, m)
}

func (r *repoMessagesExtra) GetHoles(peerID int64, peerType int32) []byte {
	m := r.get(peerID, peerType)
	if m == nil {
		return nil
	}
	return m.Holes
}
