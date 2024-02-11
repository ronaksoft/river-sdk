package repo

import (
    "github.com/dgraph-io/badger/v2"
    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/river-sdk/internal/z"
    "github.com/ronaksoft/rony/pools"
    "github.com/ronaksoft/rony/tools"
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
    z.AppendStrInt64(sb, teamID)
    z.AppendStrInt64(sb, peerID)
    z.AppendStrInt32(sb, peerType)
    id := tools.StrToByte(sb.String())
    pools.ReleaseStringsBuilder(sb)
    return id
}

func getRecentSearchPrefix(teamID int64) []byte {
    sb := pools.AcquireStringsBuilder()
    sb.WriteString(prefixRecentSearch)
    sb.WriteRune('.')
    z.AppendStrInt64(sb, teamID)
    id := tools.StrToByte(sb.String())
    pools.ReleaseStringsBuilder(sb)
    return id
}

func (r *repoRecentSearches) List(teamID int64, limit int32) []*msg.ClientRecentSearch {
    recentSearches := make([]*msg.ClientRecentSearch, 0, limit)
    _ = badgerView(func(txn *badger.Txn) error {
        opts := badger.DefaultIteratorOptions
        opts.Prefix = getRecentSearchPrefix(teamID)
        it := txn.NewIterator(opts)

        defer it.Close()
        for it.Rewind(); it.ValidForPrefix(opts.Prefix); it.Next() {
            _ = it.Item().Value(func(val []byte) error {
                r := &msg.ClientRecentSearch{}
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

    return recentSearches
}

func (r *repoRecentSearches) Put(teamID int64, recentSearch *msg.ClientRecentSearch) error {
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
