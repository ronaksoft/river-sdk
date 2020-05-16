package repo

import (
	"encoding/binary"
	"fmt"
	msg "git.ronaksoftware.com/river/msg/chat"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"github.com/dgraph-io/badger"
	"math"
)

/*
   Creation Time: 2020 - May - 16
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

type repoTopPeers struct {
	*repository
}

const (
	prefixTopPeers = "T_PEER"
)

func getTopPeerKey(cat msg.TopPeerCategory, peerID int64, peerType int32) []byte {
	return domain.StrToByte(fmt.Sprintf("%s.%02d.%021d.%d", prefixTopPeers, cat, peerID, peerType))
}

func saveTopPeer(txn *badger.Txn, cat msg.TopPeerCategory, tp *msg.TopPeer) error {
	b, _ := tp.Marshal()
	binary.BigEndian.PutUint32(b, math.Float32bits(tp.Rate))
	return txn.SetEntry(badger.NewEntry(
		getTopPeerKey(cat, tp.Peer.ID, tp.Peer.Type),
		b,
	))
}

func getTopPeer(txn *badger.Txn, cat msg.TopPeerCategory, peerID int64, peerType int32) (*msg.TopPeer, error) {
	item, err := txn.Get(getTopPeerKey(cat, peerID, peerType))
	switch err {
	case nil:
	case badger.ErrKeyNotFound:
		return nil, nil
	default:
		return nil, err
	}

	tp := &msg.TopPeer{}
	err = item.Value(func(val []byte) error {
		return tp.Unmarshal(val)
	})
	return tp, err
}


func updateTopPeerRate(txn *badger.Txn, cat msg.TopPeerCategory, peerID int64, peerType int32, accessTime int64) error {
	tp, _ := getTopPeer(txn, cat, peerID, peerType)
	tp.Rate += float32(math.Min(math.Exp(float64(float32(accessTime-tp.LastUpdate)/domain.SysConfig.TopPeerDecayRate)), float64(domain.SysConfig.TopPeerMaxStep)))
	return saveTopPeer(txn, cat, tp)
}

func (r *repoTopPeers) Save(cat msg.TopPeerCategory, tps ...*msg.TopPeer) error {
	return nil
}
