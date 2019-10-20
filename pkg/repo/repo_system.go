package repo

import (
	"encoding/binary"
	"fmt"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/dgraph-io/badger"
)

const (
	systemPrefix = "SYS"
)

type repoSystem struct {
	*repository
}

func (r *repoSystem) getKey(keyName string) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%s", systemPrefix, keyName))
}

// LoadInt
func (r *repoSystem) LoadInt(keyName string) (uint64, error) {
	keyValue := uint64(0)
	err := r.badger.View(func(txn *badger.Txn) error {
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
	if err != nil {
		return 0, err
	}

	return keyValue, nil
}

// LoadString
func (r *repoSystem) LoadString(keyName string) (string, error) {

	var v []byte
	err := r.badger.View(func(txn *badger.Txn) error {
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

// LoadBytes
func (r *repoSystem) LoadBytes(keyName string) ([]byte, error) {
	var v []byte
	err := r.badger.View(func(txn *badger.Txn) error {
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

// SaveInt
func (r *repoSystem) SaveInt(keyName string, keyValue uint64) error {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, keyValue)
	return badgerUpdate(func(txn *badger.Txn) error {
		return txn.SetEntry(
			badger.NewEntry(r.getKey(keyName), b),
		)
	})
}

// SaveString
func (r *repoSystem) SaveString(keyName string, keyValue string) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		return txn.SetEntry(
			badger.NewEntry(r.getKey(keyName), ronak.StrToByte(keyValue)),
		)
	})
}

// SaveBytes
func (r *repoSystem) SaveBytes(keyName string, keyValue []byte) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		return txn.SetEntry(
			badger.NewEntry(r.getKey(keyName), keyValue),
		)
	})
}
