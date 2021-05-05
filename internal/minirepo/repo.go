package minirepo

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/tidwall/buntdb"
	"os"
	"path/filepath"
)

/*
   Creation Time: 2021 - May - 05
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

var (
	r        *repository
	Dialogs  *repoDialogs
	Contacts *repoContacts
)

type repository struct {
	db    *bolt.DB
	index *buntdb.DB
}

func MustInit(dbPath string) {
	err := Init(dbPath)
	if err != nil {
		panic(err)
	}
}

func Init(dbPath string) (err error) {
	if r != nil {
		return nil
	}

	r = &repository{}

	boltPath := filepath.Join(dbPath, "db")
	_ = os.MkdirAll(boltPath, os.ModePerm)
	if boldDB, err := bolt.Open(filepath.Join(boltPath, "db.dat"), 0666, bolt.DefaultOptions); err != nil {
		return err
	} else {
		r.db = boldDB
	}
	_ = r.db.Update(func(tx *bolt.Tx) error {
		_, _ = tx.CreateBucketIfNotExists(bucketDialogs)
		return nil
	})

	// Initialize BuntDB Indexer
	buntPath := filepath.Join(dbPath, "index")
	_ = os.MkdirAll(buntPath, os.ModePerm)
	if buntIndex, err := buntdb.Open(filepath.Join(buntPath, "index.dat")); err != nil {
		return err
	} else {
		r.index = buntIndex
	}

	_ = r.index.Update(func(tx *buntdb.Tx) error {
		_ = tx.CreateIndex(indexDialogs, fmt.Sprintf("%s.*", prefixDialogs), buntdb.IndexBinary)
		_ = tx.CreateIndex(indexContacts, fmt.Sprintf("%s.*", prefixContacts), buntdb.IndexBinary)
		return nil
	})

	Dialogs = newDialog(r)
	Contacts = newContact(r)

	return
}
