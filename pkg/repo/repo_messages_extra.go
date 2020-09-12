package repo

import (
	"encoding/json"
	"git.ronaksoft.com/river/sdk/internal/pools"
	"git.ronaksoft.com/river/sdk/internal/tools"
	"git.ronaksoft.com/river/sdk/pkg/repo/dto"
	"github.com/dgraph-io/badger/v2"
)

const (
	prefixMessageExtra = "MSG_EX"
)

type repoMessagesExtra struct {
	*repository
}

func (r *repoMessagesExtra) getKey(teamID, peerID int64, peerType int32) []byte {
	sb := pools.AcquireStringsBuilder()
	sb.WriteString(prefixMessageExtra)
	sb.WriteRune('.')
	tools.AppendStrInt64(sb, teamID)
	tools.AppendStrInt64(sb, peerID)
	tools.AppendStrInt32(sb, peerType)
	id := tools.StrToByte(sb.String())
	pools.ReleaseStringsBuilder(sb)
	return id
}

func (r *repoMessagesExtra) get(teamID, peerID int64, peerType int32) *dto.MessagesExtra {
	message := new(dto.MessagesExtra)
	_ = badgerView(func(txn *badger.Txn) error {
		item, err := txn.Get(r.getKey(teamID, peerID, peerType))
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

func (r *repoMessagesExtra) SaveScrollID(teamID, peerID int64, peerType int32, msgID int64) {
	key := r.getKey(teamID, peerID, peerType)
	m := r.get(teamID, peerID, peerType)
	if m == nil {
		return
	}
	m.ScrollID = msgID
	r.save(key, m)
}

func (r *repoMessagesExtra) GetScrollID(teamID, peerID int64, peerType int32) int64 {
	m := r.get(teamID, peerID, peerType)
	if m == nil {
		return 0
	}
	return m.ScrollID
}

func (r *repoMessagesExtra) SaveHoles(teamID, peerID int64, peerType int32, data []byte) {
	key := r.getKey(teamID, peerID, peerType)
	m := r.get(teamID, peerID, peerType)
	if m == nil {
		return
	}
	m.Holes = data
	r.save(key, m)
}

func (r *repoMessagesExtra) GetHoles(teamID, peerID int64, peerType int32) []byte {
	m := r.get(teamID, peerID, peerType)
	if m == nil {
		return nil
	}
	return m.Holes
}
