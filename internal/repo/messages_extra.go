package repo

import (
    "encoding/json"

    "github.com/dgraph-io/badger/v2"
    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/river-sdk/internal/z"
    "github.com/ronaksoft/rony/pools"
    "github.com/ronaksoft/rony/tools"
)

const (
    prefixMessageExtra = "MSG_EX"
)

type MessagesExtraItem struct {
    PeerID   int64  `json:"PeerID"`
    PeerType int32  `json:"PeerType"`
    ScrollID int64  `json:"ScrollID"`
    Holes    []byte `json:"Holes"`
}

type repoMessagesExtra struct {
    *repository
}

func (r *repoMessagesExtra) getKey(teamID, peerID int64, peerType int32, cat msg.MediaCategory) []byte {
    sb := pools.AcquireStringsBuilder()
    sb.WriteString(prefixMessageExtra)
    sb.WriteRune('.')
    z.AppendStrInt64(sb, teamID)
    z.AppendStrInt64(sb, peerID)
    z.AppendStrInt32(sb, peerType)
    z.AppendStrInt32(sb, int32(cat))
    id := tools.StrToByte(sb.String())
    pools.ReleaseStringsBuilder(sb)
    return id
}

func (r *repoMessagesExtra) get(teamID, peerID int64, peerType int32, cat msg.MediaCategory) *MessagesExtraItem {
    message := &MessagesExtraItem{}
    _ = badgerView(func(txn *badger.Txn) error {
        item, err := txn.Get(r.getKey(teamID, peerID, peerType, cat))
        if err != nil {
            return err
        }
        return item.Value(func(val []byte) error {
            return json.Unmarshal(val, message)
        })
    })
    return message
}

func (r *repoMessagesExtra) save(key []byte, m *MessagesExtraItem) {
    bytes, _ := json.Marshal(m)
    _ = badgerUpdate(func(txn *badger.Txn) error {
        return txn.SetEntry(badger.NewEntry(key, bytes))
    })
}

func (r *repoMessagesExtra) SaveScrollID(teamID, peerID int64, peerType int32, cat msg.MediaCategory, msgID int64) {
    key := r.getKey(teamID, peerID, peerType, cat)
    m := r.get(teamID, peerID, peerType, cat)
    if m == nil {
        return
    }
    m.ScrollID = msgID
    r.save(key, m)
}

func (r *repoMessagesExtra) GetScrollID(teamID, peerID int64, peerType int32, cat msg.MediaCategory) int64 {
    m := r.get(teamID, peerID, peerType, cat)
    if m == nil {
        return 0
    }
    return m.ScrollID
}

func (r *repoMessagesExtra) SaveHoles(teamID, peerID int64, peerType int32, cat msg.MediaCategory, data []byte) {
    key := r.getKey(teamID, peerID, peerType, cat)
    m := r.get(teamID, peerID, peerType, cat)
    if m == nil {
        return
    }
    m.Holes = data
    r.save(key, m)
}

func (r *repoMessagesExtra) GetHoles(teamID, peerID int64, peerType int32, cat msg.MediaCategory) []byte {
    m := r.get(teamID, peerID, peerType, cat)
    if m == nil {
        return nil
    }
    return m.Holes
}
