package minirepo

import (
	"fmt"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"github.com/boltdb/bolt"
	"github.com/tidwall/buntdb"
	"go.uber.org/zap"
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
	r       *repository
	Dialogs *repoDialogs
	Users   *repoUsers
	Groups  *repoGroups
	General *repoGenerals
	logger  *logs.Logger
)

type repository struct {
	db    *bolt.DB
	index *buntdb.DB
}

func init() {
	logger = logs.With("MiniREPO")
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

	boltPath := filepath.Join(dbPath, "bolt")
	_ = os.MkdirAll(boltPath, os.ModePerm)
	if boldDB, err := bolt.Open(filepath.Join(boltPath, "db.dat"), 0666, bolt.DefaultOptions); err != nil {
		return err
	} else {
		r.db = boldDB
	}
	_ = r.db.Update(func(tx *bolt.Tx) error {
		buckets := [][]byte{
			bucketGroups, bucketUsers, bucketGenerals, bucketContacts, bucketDialogs,
		}
		for _, b := range buckets {
			_, err = tx.CreateBucketIfNotExists(b)
			if err != nil {
				logger.Error("MiniRepo got error on creating bucket", zap.Error(err))
			}
		}
		return nil
	})

	// Initialize BuntDB Indexer
	buntPath := filepath.Join(dbPath, "bunty")
	_ = os.MkdirAll(buntPath, os.ModePerm)
	if buntIndex, err := buntdb.Open(filepath.Join(buntPath, "mini.dat")); err != nil {
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
	Users = newUser(r)
	Groups = newGroup(r)
	General = newGeneral(r)

	return
}
