package minirepo

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"github.com/boltdb/bolt"
	"github.com/ronaksoft/rony/tools"
	"strings"
)

/*
   Creation Time: 2021 - May - 07
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

var (
	bucketGroups = []byte("GRP")
)

type repoGroups struct {
	*repository
}

func newGroup(r *repository) *repoGroups {
	rd := &repoGroups{
		repository: r,
	}
	return rd
}

func (d *repoGroups) Save(groups ...*msg.Group) error {
	alloc := tools.NewAllocator()
	defer alloc.ReleaseAll()

	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketGroups)
		for _, group := range groups {
			err := b.Put(
				alloc.Gen(group.TeamID, group.ID),
				alloc.Marshal(group),
			)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (d *repoGroups) Delete(teamID int64, groupID int64) error {
	alloc := tools.NewAllocator()
	defer alloc.ReleaseAll()

	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketGroups)
		err := b.Delete(alloc.Gen(teamID, groupID))
		if err != nil {
			return err
		}
		return nil
	})
}

func (d *repoGroups) Read(teamID int64, groupID int64) (*msg.Group, error) {
	alloc := tools.NewAllocator()
	defer alloc.ReleaseAll()

	group := &msg.Group{}
	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketGroups)
		v := b.Get(alloc.Gen(teamID, groupID))
		if len(v) > 0 {
			return group.Unmarshal(v)
		}
		return domain.ErrNotFound
	})
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (d *repoGroups) ReadMany(teamID int64, groupIDs ...int64) ([]*msg.Group, error) {
	alloc := tools.NewAllocator()
	defer alloc.ReleaseAll()

	groups := make([]*msg.Group, 0, len(groupIDs))
	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketGroups)
		for _, groupID := range groupIDs {
			group := &msg.Group{}
			v := b.Get(alloc.Gen(teamID, groupID))
			if len(v) > 0 {
				_ = group.Unmarshal(v)
				groups = append(groups, group)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (d *repoGroups) Search(teamID int64, phrase string, limit int) []*msg.Group {
	alloc := tools.NewAllocator()
	defer alloc.ReleaseAll()

	groups := make([]*msg.Group, 0, limit)
	_ = d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketGroups)
		_ = b.ForEach(func(k, v []byte) error {
			g := &msg.Group{}
			_ = g.Unmarshal(v)
			if g.TeamID != teamID {
				return nil
			}
			if strings.Contains(strings.ToLower(g.Title), phrase) {
				groups = append(groups, g)
				limit--
			}
			if limit < 0 {
				return domain.ErrLimitReached
			}
			return nil
		})
		return nil
	})
	return groups
}
