package minirepo

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"github.com/boltdb/bolt"
	"github.com/ronaksoft/rony/store"
	"github.com/ronaksoft/rony/tools"
	"github.com/tidwall/buntdb"
	"strings"
)

/*
   Creation Time: 2021 - May - 05
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

const (
	prefixDialogs = "DLG"
	indexDialogs
)

var (
	bucketDialogs = []byte("DLG")
)

type repoDialogs struct {
	*repository
}

func newDialog(r *repository) *repoDialogs {
	rd := &repoDialogs{
		repository: r,
	}
	return rd
}

func (d *repoDialogs) Save(dialog *msg.Dialog) error {
	alloc := store.NewAllocator()
	defer alloc.ReleaseAll()

	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketDialogs)
		err := b.Put(
			alloc.Gen(dialog.TeamID, dialog.PeerID, dialog.PeerType),
			alloc.Marshal(dialog),
		)
		if err != nil {
			return err
		}
		err = d.index.Update(func(tx *buntdb.Tx) error {
			_, _, err := tx.Set(
				tools.B2S(alloc.Gen(prefixDialogs, dialog.TeamID, dialog.PeerID, dialog.PeerType)),
				tools.Int64ToStr(dialog.TopMessageID),
				nil,
			)
			return err
		})
		if err != nil {
			return err
		}
		return nil
	})
}

func (d *repoDialogs) List(offset, limit int32) ([]*msg.Dialog, error) {
	dialogs := make([]*msg.Dialog, 0, limit)
	alloc := store.NewAllocator()
	defer alloc.ReleaseAll()

	err := d.index.View(func(tx *buntdb.Tx) error {
		return tx.Descend(indexDialogs, func(key, value string) bool {
			if offset--; offset >= 0 {
				return true
			}
			if limit--; limit < 0 {
				return false
			}

			_ = d.db.View(func(tx *bolt.Tx) error {
				b := tx.Bucket(bucketDialogs)
				parts := strings.Split(key, ".")
				v := b.Get(alloc.Gen(prefixDialogs, tools.StrToInt64(parts[1]), tools.StrToInt64(parts[2]), tools.StrToInt32(parts[3])))
				if len(v) > 0 {
					dialog := &msg.Dialog{}
					_ = dialog.Unmarshal(v)
					dialogs = append(dialogs, dialog)
				}
				return nil
			})
			return true
		})
	})

	if err != nil {
		return nil, err
	}
	return dialogs, nil
}
