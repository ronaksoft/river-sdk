package repo

import (
	"encoding/binary"
	"fmt"
	"github.com/dgraph-io/badger/v2"
	"github.com/ronaksoft/rony/tools"
)

const (
	systemPrefix = "SYS"
)

type repoSystem struct {
	*repository
}

func (r *repoSystem) getKey(keyName string) []byte {
	return tools.StrToByte(fmt.Sprintf("%s.%s", systemPrefix, keyName))
}

func (r *repoSystem) LoadInt(keyName string) (uint64, error) {
	keyValue := uint64(0)
	err := badgerView(func(txn *badger.Txn) error {
		item, err := txn.Get(r.getKey(keyName))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			switch len(val) {
			case 4:
				keyValue = uint64(binary.BigEndian.Uint32(val))
			case 8:
				keyValue = binary.BigEndian.Uint64(val)
			}
			return nil
		})
	})
	switch err {
	case nil, badger.ErrKeyNotFound:
		return keyValue, nil
	default:
		return 0, err
	}
}

func (r *repoSystem) LoadInt64(keyName string) (int64, error) {
	x, err := r.LoadInt(keyName)
	if err != nil {
		return 0, err
	}
	return int64(x), nil
}

func (r *repoSystem) LoadString(keyName string) (string, error) {

	var v []byte
	err := badgerView(func(txn *badger.Txn) error {
		item, err := txn.Get(r.getKey(keyName))
		if err != nil {
			return err
		}
		v, err = item.ValueCopy(nil)
		return err
	})
	if err != nil {
		return "", err
	}

	return string(v), nil
}

func (r *repoSystem) LoadBytes(keyName string) ([]byte, error) {
	var v []byte
	err := badgerView(func(txn *badger.Txn) error {
		item, err := txn.Get(r.getKey(keyName))
		if err != nil {
			return err
		}
		v, err = item.ValueCopy(nil)
		return err
	})
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (r *repoSystem) SaveInt(keyName string, keyValue uint64) error {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, keyValue)
	return badgerUpdate(func(txn *badger.Txn) error {
		return txn.SetEntry(
			badger.NewEntry(r.getKey(keyName), b),
		)
	})
}

func (r *repoSystem) SaveString(keyName string, keyValue string) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		return txn.SetEntry(
			badger.NewEntry(r.getKey(keyName), tools.StrToByte(keyValue)),
		)
	})
}

func (r *repoSystem) SaveBytes(keyName string, keyValue []byte) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		return txn.SetEntry(
			badger.NewEntry(r.getKey(keyName), keyValue),
		)
	})
}

func (r *repoSystem) Delete(keyName string) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		return txn.Delete(r.getKey(keyName))
	})
}
