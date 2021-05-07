package minirepo

import (
	"fmt"
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"github.com/boltdb/bolt"
	"github.com/ronaksoft/rony/store"
	"github.com/ronaksoft/rony/tools"
	"github.com/tidwall/buntdb"
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
	prefixUsers = "USERS"
	indexUsers
	prefixContacts = "CONTACTS"
	indexContacts
)

var (
	bucketContacts = []byte("CNTC")
	bucketUsers    = []byte("USR")
)

type repoUsers struct {
	*repository
}

func newUser(r *repository) *repoUsers {
	rd := &repoUsers{
		repository: r,
	}
	return rd
}

func (d *repoUsers) getUser(alloc *store.Allocator, b *bolt.Bucket, userID int64) (*msg.User, error) {
	v := b.Get(alloc.Gen(userID))
	if len(v) == 0 {
		return nil, domain.ErrNotFound
	}
	u := &msg.User{}
	_ = u.Unmarshal(v)
	return u, nil
}

func (d *repoUsers) SaveContact(contact *msg.ContactUser, lastSeen int64) error {
	alloc := store.NewAllocator()
	defer alloc.ReleaseAll()

	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketContacts)
		err := b.Put(
			alloc.Gen(contact.ID),
			alloc.Marshal(contact),
		)
		if err != nil {
			return err
		}
		err = d.index.Update(func(tx *buntdb.Tx) error {
			_, _, err := tx.Set(
				fmt.Sprintf("%s.%d", prefixContacts, contact.ID),
				tools.Int64ToStr(lastSeen),
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

func (d *repoUsers) SaveUser(user *msg.User) error {
	alloc := store.NewAllocator()
	defer alloc.ReleaseAll()

	return d.db.Update(func(tx *bolt.Tx) error {
		b1 := tx.Bucket(bucketUsers)
		err := b1.Put(
			alloc.Gen(user.ID),
			alloc.Marshal(user),
		)
		if err != nil {
			return err
		}

		b2 := tx.Bucket(bucketContacts)
		v1 := b2.Get(alloc.Gen(user.ID))
		if len(v1) > 0 {
			err = d.index.Update(func(tx *buntdb.Tx) error {
				_, _, err := tx.Set(
					fmt.Sprintf("%s.%d", prefixContacts, user.ID),
					tools.Int64ToStr(user.LastSeen),
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

func (d *repoUsers) Delete(teamID int64, peerID int64, peerType int32) error {
	alloc := store.NewAllocator()
	defer alloc.ReleaseAll()

	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketContacts)
		err := b.Delete(alloc.Gen(teamID, peerID, peerType))
		if err != nil {
			return err
		}
		return d.index.Update(func(tx *buntdb.Tx) error {
			_, _ = tx.Delete(fmt.Sprintf("%s.%d.%d.%d", prefixContacts, teamID, peerID, peerType))
			return nil
		})
	})
}

func (d *repoUsers) ReadAllContacts() (*msg.ContactsMany, error) {
	res := &msg.ContactsMany{}

	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketContacts)
		_ = b.ForEach(func(k, v []byte) error {
			c := &msg.ContactUser{}
			_ = c.Unmarshal(v)
			res.ContactUsers = append(res.ContactUsers, c)
			if c.Phone != "" {
				res.Contacts = append(res.Contacts, &msg.PhoneContact{
					ClientID:  c.ClientID,
					FirstName: c.FirstName,
					LastName:  c.LastName,
					Phone:     c.Phone,
				})
			}
			return nil
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	alloc := store.NewAllocator()
	defer alloc.ReleaseAll()

	for _, c := range res.ContactUsers {
		_ = d.db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket(bucketUsers)
			u, _ := d.getUser(alloc, b, c.ID)
			if u != nil {
				res.Users = append(res.Users, u)
			}
			return nil
		})
	}
	return res, nil
}

func (d *repoUsers) ReadUser(userID int64) (*msg.User, error) {
	alloc := store.NewAllocator()
	defer alloc.ReleaseAll()

	var (
		u   *msg.User
		err error
	)
	err = d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketUsers)
		u, err = d.getUser(alloc, b, userID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return u, nil
}
