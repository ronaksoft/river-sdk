package repo

import (
	"fmt"
	msg "git.ronaksoftware.com/river/msg/chat"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"github.com/dgraph-io/badger"
	"github.com/tidwall/buntdb"
	"go.uber.org/zap"
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
	prefixTopPeers          = "T_PEER"
	indexTopPeersUser       = "T_PEER_U"
	indexTopPeersGroup      = "T_PEER_G"
	indexTopPeersForward    = "T_PEER_F"
	indexTopPeersBotMessage = "T_PEER_BM"
	indexTopPeersBotInline  = "T_PEER_BI"
)

func getTopPeerKey(cat msg.TopPeerCategory, peerID int64, peerType int32) []byte {
	return domain.StrToByte(fmt.Sprintf("%s_%02d.%021d.%d", prefixTopPeers, cat, peerID, peerType))
}

func saveTopPeer(txn *badger.Txn, cat msg.TopPeerCategory, tp *msg.TopPeer) error {
	if tp.Peer == nil {
		logs.Warn("Could not save top peer, peer is nit", zap.Any("TP", tp))
		return domain.ErrDoesNotExists
	}
	b, _ := tp.Marshal()
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

func (r *repoTopPeers) updateIndex(cat msg.TopPeerCategory, peerID int64, peerType int32, rate float32) error {
	return r.bunt.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(
			domain.ByteToStr(getTopPeerKey(cat, peerID, peerType)),
			fmt.Sprintf("%f", rate),
			nil,
		)
		return err
	})
}

func (r *repoTopPeers) Save(cat msg.TopPeerCategory, tps ...*msg.TopPeer) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		for _, tp := range tps {
			err := saveTopPeer(txn, cat, tp)
			if err != nil {
				return err
			}
			err = r.updateIndex(cat, tp.Peer.ID, tp.Peer.Type, tp.Rate)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *repoTopPeers) Update(cat msg.TopPeerCategory, peerID int64, peerType int32) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		accessTime := domain.Now().Unix()
		tp, _ := getTopPeer(txn, cat, peerID, peerType)
		if tp == nil {
			switch msg.PeerType(peerType) {
			case msg.PeerUser:
				p, err := getUserByKey(txn, getUserKey(peerID))
				if err != nil {
					return err
				}
				tp = &msg.TopPeer{
					Peer: &msg.Peer{
						ID:         p.ID,
						Type:       peerType,
						AccessHash: p.AccessHash,
					},
				}
			case msg.PeerGroup:
				p, err := getGroupByKey(txn, getGroupKey(peerID))
				if err != nil {
					return err
				}
				tp = &msg.TopPeer{
					Peer: &msg.Peer{
						ID:   p.ID,
						Type: peerType,
					},
				}
			default:
				return domain.ErrNotFound
			}
		}
		tp.Rate += float32(math.Min(math.Exp(float64(float32(accessTime-tp.LastUpdate)/domain.SysConfig.TopPeerDecayRate)), float64(domain.SysConfig.TopPeerMaxStep)))
		tp.LastUpdate = accessTime
		err := saveTopPeer(txn, cat, tp)
		if err != nil {
			return err
		}

		return r.updateIndex(cat, peerID, peerType, tp.Rate)
	})
}

func (r *repoTopPeers) List(cat msg.TopPeerCategory, offset, limit int32) ([]*msg.TopPeer, error) {
	var (
		topPeers  = make([]*msg.TopPeer, 0, limit)
		indexName = ""
	)
	switch cat {
	case msg.TopPeerCategory_Users:
		indexName = indexTopPeersUser
	case msg.TopPeerCategory_Groups:
		indexName = indexTopPeersGroup
	case msg.TopPeerCategory_Forwards:
		indexName = indexTopPeersForward
	case msg.TopPeerCategory_BotsMessage:
		indexName = indexTopPeersBotMessage
	case msg.TopPeerCategory_BotsInline:
		indexName = indexTopPeersBotInline
	default:
		panic("BUG! we dont support the top peer category")
	}

	err := badgerView(func(txn *badger.Txn) error {
		return r.bunt.View(func(tx *buntdb.Tx) error {
			return tx.Descend(indexName, func(key, value string) bool {
				if offset--; offset >= 0 {
					return true
				}
				if limit--; limit < 0 {
					return false
				}
				peer := getPeerFromKey(key)
				topPeer, err := getTopPeer(txn, cat, peer.ID, peer.Type)
				if err == nil && topPeer != nil {
					topPeers = append(topPeers, topPeer)
				}
				return true
			})
		})
	})
	return topPeers, err
}