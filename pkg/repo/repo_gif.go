package repo

import (
	"fmt"
	"git.ronaksoftware.com/river/msg/msg"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"github.com/dgraph-io/badger"
	"github.com/tidwall/buntdb"
)

/*
   Creation Time: 2020 - Jun - 15
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

const (
	prefixGif = "GIF"
	indexGif  = "GIF"
)

type repoGifs struct {
	*repository
}

func getGifKey(clusterID int32, docID int64) []byte {
	return domain.StrToByte(fmt.Sprintf("%s.%012d.%021d", prefixGif, clusterID, docID))
}

func getGifByID(txn *badger.Txn, clusterID int32, docID int64) (*msg.MediaDocument, error) {
	return getGifByKey(txn, getGifKey(clusterID, docID))
}

func getGifByKey(txn *badger.Txn, key []byte) (*msg.MediaDocument, error) {
	md := &msg.MediaDocument{}
	item, err := txn.Get(key)
	if err != nil {
		return nil, err
	}
	err = item.Value(func(val []byte) error {
		return md.Unmarshal(val)
	})
	if err != nil {
		return nil, err
	}
	return md, nil
}

func saveGif(txn *badger.Txn, md *msg.MediaDocument) error {
	mdBytes, _ := md.Marshal()
	err := txn.SetEntry(badger.NewEntry(
		getGifKey(md.Doc.ClusterID, md.Doc.ID),
		mdBytes,
	))
	return err
}

func deleteGif(txn *badger.Txn, clusterID int32, docID int64) error {
	return txn.Delete(getGifKey(clusterID, docID))
}

func (r *repoGifs) UpdateLastAccess(clusterID int32, docID int64, accessTime int64) error {
	if !r.IsSaved(clusterID, docID) {
		return nil
	}
	return r.bunt.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(
			domain.ByteToStr(getGifKey(clusterID, docID)),
			fmt.Sprintf("%021d", accessTime),
			nil,
		)
		return err
	})
}

func (r *repoGifs) Get(clusterID int32, docID int64) (gif *msg.MediaDocument,err error) {
	err = badgerView(func(txn *badger.Txn) error {
		gif, err = getGifByID(txn, clusterID, docID)
		return err
	})
	return
}

func (r *repoGifs) IsSaved(clusterID int32, docID int64) (found bool) {
	_ = badgerView(func(txn *badger.Txn) error {
		_, err := getGifByID(txn, clusterID, docID)
		switch err {
		case nil:
			found = true
			return nil
		case badger.ErrKeyNotFound:
			found = false
			return nil
		}
		return err
	})
	return
}

func (r *repoGifs) Save(cf *msg.MediaDocument) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		return saveGif(txn, cf)
	})
}

func (r *repoGifs) GetSaved() (*msg.SavedGifs, error) {
	savedGifs := make([]*msg.MediaDocument, 0, 20)
	err := badgerView(func(txn *badger.Txn) error {
		return r.bunt.View(func(tx *buntdb.Tx) error {
			return tx.Descend(indexGif, func(key, value string) bool {
				md, err := getGifByKey(txn, domain.StrToByte(key))
				if err != nil {
					return false
				}
				savedGifs = append(savedGifs, md)
				return true
			})
		})
	})

	if err != nil {
		return nil, err
	}
	return &msg.SavedGifs{
		Docs: savedGifs,
		NotModified: false,
	}, nil
}

func (r *repoGifs) Delete(clusterID int32, docID int64) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		err := deleteGif(txn, clusterID, docID)
		switch err {
		case nil, badger.ErrKeyNotFound:
		default:
			return err
		}
		return r.bunt.Update(func(tx *buntdb.Tx) error {
			_, err := tx.Delete(domain.ByteToStr(getGifKey(clusterID, docID)))
			return err
		})
	})
}
