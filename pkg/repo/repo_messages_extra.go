package repo

import (
	"encoding/json"
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo/dto"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/dgraph-io/badger"
)

const (
	prefixMessageExtra = "MSG_EX"
)

type repoMessagesExtra struct {
	*repository
}

func (r *repoMessagesExtra) getKey(peerID int64, peerType int32) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%021d.%d", prefixMessageExtra, peerID, peerType))
}

func (r *repoMessagesExtra) get(peerID int64, peerType int32) *dto.MessagesExtra {
	r.mx.Lock()
	defer r.mx.Unlock()

	message := new(dto.MessagesExtra)
	err := r.badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get(r.getKey(peerID, peerType))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, message)
		})
	})
	if err != nil {
		return nil
	}
	return message
}

func (r *repoMessagesExtra) save(key []byte, m *dto.MessagesExtra) {
	bytes, _ := json.Marshal(m)
	_ = r.badger.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry(key, bytes))
	})
}

func (r *repoMessagesExtra) SaveScrollID(peerID int64, peerType int32, msgID int64) {
	r.mx.Lock()
	defer r.mx.Unlock()

	key := r.getKey(peerID, peerType)
	m := r.get(peerID, peerType)
	if m == nil {
		return
	}
	m.ScrollID = msgID
	r.save(key, m)
}

func (r *repoMessagesExtra) GetScrollID(peerID int64, peerType int32) int64  {
	r.mx.Lock()
	defer r.mx.Unlock()

	m := r.get(peerID, peerType)
	if m == nil {
		return 0
	}
	return m.ScrollID
}

func (r *repoMessagesExtra) SaveHoles(peerID int64, peerType int32, data []byte)  {
	r.mx.Lock()
	defer r.mx.Unlock()

	key := r.getKey(peerID, peerType)
	m := r.get(peerID, peerType)
	if m == nil {
		return
	}
	m.Holes = data
	r.save(key, m)
}

func (r *repoMessagesExtra) GetHoles(peerID int64, peerType int32) []byte {
	r.mx.Lock()
	defer r.mx.Unlock()

	m := r.get(peerID, peerType)
	if m == nil {
		return nil
	}
	return m.Holes
}
