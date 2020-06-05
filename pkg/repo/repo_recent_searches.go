package repo

import (
	"fmt"
	"git.ronaksoftware.com/river/msg/msg"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"github.com/dgraph-io/badger"
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

func GetRecentSearchKey(peerID int64, peerType int32) []byte {
	return domain.StrToByte(fmt.Sprintf("%s.%021d.%d", prefixRecentSearch, peerID, peerType))
}

func (r *repoRecentSearches) List(limit int32) []*msg.RecentSearch {
	recentSearches := make([]*msg.RecentSearch, 0, limit)
	err := badgerView(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = domain.StrToByte(prefixRecentSearch)
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

func (r *repoRecentSearches) Put(recentSearch *msg.RecentSearch) error {
	err := badgerUpdate(func(txn *badger.Txn) error {
		recentSearchBytes, _ := recentSearch.Marshal()
		recentSearchKey := GetRecentSearchKey(recentSearch.Peer.ID, recentSearch.Peer.Type)
		err := txn.SetEntry(badger.NewEntry(
			recentSearchKey, recentSearchBytes,
		))

		return err
	})
	return err
}

func (r *repoRecentSearches) Delete(peer *msg.InputPeer) error {
	err := badgerUpdate(func(txn *badger.Txn) error {
		recentSearchKey := GetRecentSearchKey(peer.ID, int32(peer.Type))
		err := txn.Delete(recentSearchKey)
		return err
	})
	return err
}

func (r *repoRecentSearches) Clear() error {
	err := badgerUpdate(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = domain.StrToByte(fmt.Sprintf("%s.", prefixRecentSearch))
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
