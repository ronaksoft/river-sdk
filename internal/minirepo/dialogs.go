package minirepo

import (
	"fmt"
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"github.com/boltdb/bolt"
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

func (d *repoDialogs) Save(dialogs ...*msg.Dialog) error {
	alloc := tools.NewAllocator()
	defer alloc.ReleaseAll()

	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketDialogs)
		for _, dialog := range dialogs {
			err := b.Put(
				alloc.Gen(dialog.TeamID, dialog.PeerID, dialog.PeerType),
				alloc.Marshal(dialog),
			)
			if err != nil {
				return err
			}
			err = d.index.Update(func(tx *buntdb.Tx) error {
				_, _, err := tx.Set(
					fmt.Sprintf("%s.%d.%d.%d", prefixDialogs, dialog.TeamID, dialog.PeerID, dialog.PeerType),
					tools.Int64ToStr(dialog.TopMessageID),
					nil,
				)
				return err
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (d *repoDialogs) Delete(teamID int64, peerID int64, peerType int32) error {
	alloc := tools.NewAllocator()
	defer alloc.ReleaseAll()

	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketDialogs)
		err := b.Delete(alloc.Gen(teamID, peerID, peerType))
		if err != nil {
			return err
		}
		return d.index.Update(func(tx *buntdb.Tx) error {
			_, _ = tx.Delete(fmt.Sprintf("%s.%d.%d.%d", prefixDialogs, teamID, peerID, peerType))
			return nil
		})
	})
}

func (d *repoDialogs) Read(teamID int64, peerID int64, peerType int32) (*msg.Dialog, error) {
	alloc := tools.NewAllocator()
	defer alloc.ReleaseAll()

	dialog := &msg.Dialog{}

	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketDialogs)
		v := b.Get(alloc.Gen(teamID, peerID, peerType))
		if len(v) > 0 {
			return dialog.Unmarshal(v)
		}
		return domain.ErrNotFound
	})
	if err != nil {
		return nil, err
	}
	return dialog, nil
}

func (d *repoDialogs) List(teamID int64, offset, limit int32) ([]*msg.Dialog, error) {
	dialogs := make([]*msg.Dialog, 0, limit)
	alloc := tools.NewAllocator()
	defer alloc.ReleaseAll()

	err := d.index.View(func(tx *buntdb.Tx) error {
		return tx.Descend(indexDialogs, func(key, value string) bool {
			if offset--; offset >= 0 {
				return true
			}
			if limit--; limit < 0 {
				return false
			}
			parts := strings.Split(key, ".")
			if tools.StrToInt64(parts[1]) != teamID {
				return true
			}
			peerID := tools.StrToInt64(parts[2])
			peerType := tools.StrToInt32(parts[3])
			_ = d.db.View(func(tx *bolt.Tx) error {
				b := tx.Bucket(bucketDialogs)
				v := b.Get(alloc.Gen(teamID, peerID, peerType))
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
