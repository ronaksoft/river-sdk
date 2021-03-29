package repo

import (
	"git.ronaksoft.com/river/sdk/internal/z"
	"github.com/dgraph-io/badger/v2"
	"github.com/ronaksoft/rony/pools"
	"github.com/ronaksoft/rony/tools"
)

/**
 * @created 03/10/2020 - 14:52
 * @project sdk
 * @author reza
 */

type repoReactions struct {
	*repository
}

const (
	prefixReactionsUseCount = "REACTIONS_USE_CNT"
)

func getReactionUseCountKey(reaction string) []byte {
	sb := pools.AcquireStringsBuilder()
	sb.WriteString(prefixReactionsUseCount)
	sb.WriteRune('.')
	sb.WriteString(reaction)
	id := tools.StrToByte(sb.String())
	pools.ReleaseStringsBuilder(sb)
	return id
}

func getReactionUseCount(txn *badger.Txn, reaction string) (uint32, error) {
	var useCount uint32
	item, err := txn.Get(getReactionUseCountKey(reaction))
	switch err {
	case nil:
	case badger.ErrKeyNotFound:
		return 0, nil
	default:
		return 0, err
	}

	err = item.Value(func(val []byte) error {
		useCount = z.ByteToUInt32(val)
		return nil
	})
	return useCount, err
}

func saveReactionUseCount(txn *badger.Txn, reaction string, useCount uint32) error {
	return txn.SetEntry(badger.NewEntry(
		getReactionUseCountKey(reaction),
		z.UInt32ToByte(useCount),
	))
}

func (r *repoReactions) GetReactionUseCount(reaction string) (useCount uint32, err error) {
	err = badgerView(func(txn *badger.Txn) error {
		useCount, err = getReactionUseCount(txn, reaction)
		return err
	})
	return
}

func (r *repoReactions) IncrementReactionUseCount(reaction string, cnt int32) error {
	err := badgerUpdate(func(txn *badger.Txn) error {
		useCount, err := getReactionUseCount(txn, reaction)
		if err != nil {
			return err
		}
		if cnt < 0 {
			useCount -= uint32(z.AbsInt32(cnt))
		} else {
			useCount += uint32(z.AbsInt32(cnt))
		}
		return saveReactionUseCount(txn, reaction, useCount)
	})
	return err
}
