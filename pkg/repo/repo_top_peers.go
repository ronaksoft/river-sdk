package repo

import (
	"fmt"
	"git.ronaksoftware.com/river/msg/msg"
	"git.ronaksoftware.com/ronak/riversdk/internal/logs"
	"git.ronaksoftware.com/ronak/riversdk/internal/pools"
	"git.ronaksoftware.com/ronak/riversdk/internal/tools"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"github.com/dgraph-io/badger/v2"
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

func getTopPeerKey(cat msg.TopPeerCategory, teamID, peerID int64, peerType int32) []byte {
	sb := pools.AcquireStringsBuilder()
	sb.WriteString(prefixTopPeers)
	sb.WriteRune('.')
	tools.AppendStrInt64(sb, teamID)
	tools.AppendStrInt32(sb, int32(cat))
	tools.AppendStrInt64(sb, peerID)
	tools.AppendStrInt32(sb, peerType)
	id := tools.StrToByte(sb.String())
	pools.ReleaseStringsBuilder(sb)
	return id
}

func saveTopPeer(txn *badger.Txn, cat msg.TopPeerCategory, teamID int64, tp *msg.TopPeer) error {
	if tp.Peer == nil {
		logs.Warn("Could not save top peer, peer is nit", zap.Any("TP", tp))
		return domain.ErrDoesNotExists
	}
	b, _ := tp.Marshal()
	return txn.SetEntry(badger.NewEntry(
		getTopPeerKey(cat, teamID, tp.Peer.ID, tp.Peer.Type),
		b,
	))
}

func getTopPeer(txn *badger.Txn, cat msg.TopPeerCategory, teamID, peerID int64, peerType int32) (*msg.TopPeer, error) {
	item, err := txn.Get(getTopPeerKey(cat, teamID, peerID, peerType))
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

func deleteTopPeer(txn *badger.Txn, cat msg.TopPeerCategory, teamID, peerID int64, peerType int32) error {
	err := txn.Delete(getTopPeerKey(cat, teamID, peerID, peerType))
	switch err {
	case nil, badger.ErrKeyNotFound:
		return nil
	default:
		return err
	}
}

func (r *repoTopPeers) updateIndex(cat msg.TopPeerCategory, teamID, peerID int64, peerType int32, rate float32) error {
	var indexName string
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
	return r.bunt.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(
			fmt.Sprintf("%s.%d.%d.%d", indexName, teamID, peerID, peerType),
			fmt.Sprintf("%f", rate),
			nil,
		)
		return err
	})
}

func (r *repoTopPeers) Save(cat msg.TopPeerCategory, userID, teamID int64, tps ...*msg.TopPeer) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		for _, tp := range tps {
			if tp.Peer != nil && tp.Peer.ID == userID {
				continue
			}
			err := saveTopPeer(txn, cat, teamID, tp)
			if err != nil {
				return err
			}
			err = r.updateIndex(cat, teamID, tp.Peer.ID, tp.Peer.Type, tp.Rate)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *repoTopPeers) Delete(cat msg.TopPeerCategory, teamID, peerID int64, peerType int32) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		return deleteTopPeer(txn, cat, teamID, peerID, peerType)
	})
}

func (r *repoTopPeers) Update(cat msg.TopPeerCategory, teamID, peerID int64, peerType int32, userID int64) error {
	if peerID == userID {
		return nil
	}
	return badgerUpdate(func(txn *badger.Txn) error {
		accessTime := domain.Now().Unix()
		tp, _ := getTopPeer(txn, cat, teamID, peerID, peerType)
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
		err := saveTopPeer(txn, cat, teamID, tp)
		if err != nil {
			return err
		}

		return r.updateIndex(cat, teamID, peerID, peerType, tp.Rate)
	})
}

func (r *repoTopPeers) List(teamID int64, cat msg.TopPeerCategory, offset, limit int32) ([]*msg.TopPeer, error) {
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
				peerTeamID, peer := getPeerFromIndexKey(key)
				if peerTeamID != teamID {
					return true
				}
				topPeer, err := getTopPeer(txn, cat, peerTeamID, peer.ID, peer.Type)
				if err == nil && topPeer != nil {
					topPeers = append(topPeers, topPeer)
				}
				return true
			})
		})
	})
	return topPeers, err
}
