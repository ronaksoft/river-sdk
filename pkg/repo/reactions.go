package repo

import (
	"fmt"
	"git.ronaksoft.com/river/sdk/internal/tools"
	"github.com/dgraph-io/badger/v2"
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
	prefixReactions = "REACTIONS"
)

func getReactionUseCountKey(reaction string) string {
	return fmt.Sprintf("%s.%s.useCount.", prefixReactions, reaction)
}

func (r *repoReactions) GetReactionUseCount(reaction string) (error, uint32) {
	var useCount uint32 = 0

	err := badgerView(func(txn *badger.Txn) error {
		item, err := txn.Get(tools.StrToByte(getReactionUseCountKey(reaction)))
		switch err {
		case nil:
		case badger.ErrKeyNotFound:
			useCount = 0
			return nil
		default:
			return err
		}

		err = item.Value(func(val []byte) error {
			useCount = tools.ByteToInt(val)
			return nil
		})

		return nil
	})

	return err, useCount
}

func (r *repoReactions) IncreaseReactionUseCount(reaction string) error {
	err, useCount := r.GetReactionUseCount(reaction)

	if err != nil {
		return err
	}

	useCount++

	err = badgerUpdate(func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry(
			tools.StrToByte(getReactionUseCountKey(reaction)),
			tools.IntToByte(useCount),
		))
	})

	return err
}

func (r *repoReactions) DecreaseReactionUseCount(reaction string) error {
	err, useCount := r.GetReactionUseCount(reaction)

	if err != nil {
		return err
	}

	useCount--

	err = badgerUpdate(func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry(
			tools.StrToByte(getReactionUseCountKey(reaction)),
			tools.IntToByte(useCount),
		))
	})

	return err
}
