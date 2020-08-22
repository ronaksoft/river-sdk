package repo

import (
	"git.ronaksoft.com/river/msg/msg"
	"git.ronaksoft.com/ronak/riversdk/internal/logs"
	"git.ronaksoft.com/ronak/riversdk/internal/pools"
	"git.ronaksoft.com/ronak/riversdk/internal/tools"
	"github.com/dgraph-io/badger/v2"
)

/**
 * @created 01/06/2020 - 14:55
 * @project riversdk
 * @author reza
 */

type repoRecentSearches struct {
	*repository
}

const (
	prefixRecentSearch = "RECENT_SEARCH"
)

func getRecentSearchKey(teamID, peerID int64, peerType int32) []byte {
	sb := pools.AcquireStringsBuilder()
	sb.WriteString(prefixRecentSearch)
	sb.WriteRune('.')
	tools.AppendStrInt64(sb, teamID)
	tools.AppendStrInt64(sb, peerID)
	tools.AppendStrInt32(sb, peerType)
	id := tools.StrToByte(sb.String())
	pools.ReleaseStringsBuilder(sb)
	return id
}

func getRecentSearchPrefix(teamID int64) []byte {
	sb := pools.AcquireStringsBuilder()
	sb.WriteString(prefixRecentSearch)
	sb.WriteRune('.')
	tools.AppendStrInt64(sb, teamID)
	id := tools.StrToByte(sb.String())
	pools.ReleaseStringsBuilder(sb)
	return id
}

func (r *repoRecentSearches) List(teamID int64, limit int32) []*msg.RecentSearch {
	recentSearches := make([]*msg.RecentSearch, 0, limit)
	err := badgerView(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = getRecentSearchPrefix(teamID)
		it := txn.NewIterator(opts)

		defer it.Close()
		for it.Rewind(); it.ValidForPrefix(opts.Prefix); it.Next() {
			_ = it.Item().Value(func(val []byte) error {
				r := &msg.RecentSearch{}
				err := r.Unmarshal(val)
				if err != nil {
					return err
				}
				recentSearches = append(recentSearches, r)
				return nil
			})
		}

		return nil
	})
	logs.ErrorOnErr("RepoRecentSearches got error on GetAll", err)
	return recentSearches
}

func (r *repoRecentSearches) Put(teamID int64, recentSearch *msg.RecentSearch) error {
	err := badgerUpdate(func(txn *badger.Txn) error {
		recentSearchBytes, _ := recentSearch.Marshal()
		recentSearchKey := getRecentSearchKey(teamID, recentSearch.Peer.ID, recentSearch.Peer.Type)
		err := txn.SetEntry(badger.NewEntry(
			recentSearchKey, recentSearchBytes,
		))

		return err
	})
	return err
}

func (r *repoRecentSearches) Delete(teamID int64, peer *msg.InputPeer) error {
	err := badgerUpdate(func(txn *badger.Txn) error {
		recentSearchKey := getRecentSearchKey(teamID, peer.ID, int32(peer.Type))
		err := txn.Delete(recentSearchKey)
		return err
	})
	return err
}

func (r *repoRecentSearches) Clear(teamID int64) error {
	err := badgerUpdate(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = getRecentSearchPrefix(teamID)
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		for it.Rewind(); it.ValidForPrefix(opts.Prefix); it.Next() {
			err := txn.Delete(it.Item().KeyCopy(nil))
			if err != nil {
				return err
			}
		}
		it.Close()

		return nil
	})
	return err
}
