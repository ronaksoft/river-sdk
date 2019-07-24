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
func (r *repoSystem) LoadInt(keyName string) (int, error) {

	keyValue := uint32(0)

	err := r.badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get(r.getKey(keyName))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			keyValue = binary.BigEndian.Uint32(val)
			return nil
		})
	})
	if err != nil {
		return 0, err
	}

	return int(keyValue), nil
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

// SaveInt
func (r *repoSystem) SaveInt(keyName string, keyValue int32) error {

	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(keyValue))
	return r.badger.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(
			badger.NewEntry(r.getKey(keyName), b),
		)
	})
}

// SaveString
func (r *repoSystem) SaveString(keyName string, keyValue string) error {

	return r.badger.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(
			badger.NewEntry(r.getKey(keyName), ronak.StrToByte(keyValue)),
		)
	})
}
