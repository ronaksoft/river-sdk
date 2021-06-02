package minirepo

import (
	"encoding/binary"
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

func (d *repoUsers) getUser(alloc *tools.Allocator, b *bolt.Bucket, userID int64) (*msg.User, error) {
	v := b.Get(alloc.Gen(userID))
	if len(v) == 0 {
		return nil, domain.ErrNotFound
	}
	u := &msg.User{}
	_ = u.Unmarshal(v)
	return u, nil
}

func (d *repoUsers) saveContact(alloc *tools.Allocator, tx *bolt.Tx, teamID int64, contact *msg.ContactUser, lastSeen int64) error {
	b := tx.Bucket(bucketContacts)
	err := b.Put(
		alloc.Gen(teamID, contact.ID),
		alloc.Marshal(contact),
	)
	if err != nil {
		return err
	}
	err = d.index.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(
			fmt.Sprintf("%s.%d.%d", prefixContacts, teamID, contact.ID),
			tools.Int64ToStr(lastSeen),
			nil,
		)
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

func (d *repoUsers) DeleteContact(teamID int64, userID int64) {
	alloc := tools.NewAllocator()
	defer alloc.ReleaseAll()

	_ = d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketContacts)
		_ = b.Delete(alloc.Gen(teamID, userID))
		_ = d.index.Update(func(tx *buntdb.Tx) error {
			_, err := tx.Delete(fmt.Sprintf("%s.%d.%d", prefixContacts, teamID, userID))
			return err
		})
		return nil
	})

}

func (d *repoUsers) SaveUser(teamID int64, users ...*msg.User) error {
	alloc := tools.NewAllocator()
	defer alloc.ReleaseAll()

	return d.db.Update(func(tx *bolt.Tx) error {
		b1 := tx.Bucket(bucketUsers)
		for _, user := range users {
			err := b1.Put(alloc.Gen(user.ID), alloc.Marshal(user))
			if err != nil {
				return err
			}

			b2 := tx.Bucket(bucketContacts)
			v1 := b2.Get(alloc.Gen(teamID, user.ID))
			if len(v1) > 0 {
				err = d.index.Update(func(tx *buntdb.Tx) error {
					_, _, err := tx.Set(
						fmt.Sprintf("%s.%d.%d", prefixContacts, teamID, user.ID),
						tools.Int64ToStr(user.LastSeen),
						nil,
					)
					return err
				})
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (d *repoUsers) SaveAllContacts(teamID int64, newContacts *msg.ContactsMany) error {
	if !newContacts.Modified {
		return nil
	}
	newContactsMap := make(map[int64]int64, len(newContacts.ContactUsers))
	for idx := range newContacts.Users {
		newContactsMap[newContacts.ContactUsers[idx].ID] = newContacts.Users[idx].LastSeen
	}
	oldContacts, err := d.ReadAllContacts(teamID)
	if err == nil {
		oldContactsMap := make(map[int64]bool, len(oldContacts.ContactUsers))
		for idx := range oldContacts.ContactUsers {
			oldContactsMap[oldContacts.ContactUsers[idx].ID] = true
		}
		for userID := range oldContactsMap {
			if _, ok := newContactsMap[userID]; !ok {
				d.DeleteContact(teamID, userID)
			}
		}
	}

	alloc := tools.NewAllocator()
	defer alloc.ReleaseAll()
	err = d.db.Update(func(tx *bolt.Tx) error {
		for _, cu := range newContacts.ContactUsers {
			err = d.saveContact(alloc, tx, teamID, cu, newContactsMap[cu.ID])
		}
		return nil
	})
	return nil
}

func (d *repoUsers) ReadAllContacts(teamID int64) (*msg.ContactsMany, error) {
	res := &msg.ContactsMany{}

	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketContacts)
		_ = b.ForEach(func(k, v []byte) error {
			if binary.BigEndian.Uint64(k[1:9]) != uint64(teamID) {
				return nil
			}

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

	alloc := tools.NewAllocator()
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
	alloc := tools.NewAllocator()
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

func (d *repoUsers) ReadMany(userIDs ...int64) ([]*msg.User, error) {
	alloc := tools.NewAllocator()
	defer alloc.ReleaseAll()

	var (
		users = make([]*msg.User, 0, len(userIDs))
		err   error
	)
	err = d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketUsers)
		for _, userID := range userIDs {
			u, _ := d.getUser(alloc, b, userID)
			if u != nil {
				users = append(users, u)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (d *repoUsers) Search(phrase string, limit int) []*msg.User {
	users := make([]*msg.User, 0, limit)
	_ = d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketUsers)
		_ = b.ForEach(func(k, v []byte) error {
			u := &msg.User{}
			_ = u.Unmarshal(v)

			if strings.Contains(strings.ToLower(u.FirstName), phrase) ||
				strings.Contains(strings.ToLower(u.LastName), phrase) ||
				strings.Contains(strings.ToLower(u.Username), phrase) {
				users = append(users, u)
				limit--
			}
			if limit < 0 {
				return domain.ErrLimitReached
			}
			return nil
		})
		return nil
	})
	return users
}

func (d *repoUsers) SearchContacts(teamID int64, phrase string, limit int) []*msg.ContactUser {
	contacts := make([]*msg.ContactUser, 0, limit)
	_ = d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketContacts)
		_ = b.ForEach(func(k, v []byte) error {
			if binary.BigEndian.Uint64(k[1:9]) != uint64(teamID) {
				return nil
			}

			c := &msg.ContactUser{}
			_ = c.Unmarshal(v)
			if strings.Contains(strings.ToLower(c.FirstName), phrase) ||
				strings.Contains(strings.ToLower(c.LastName), phrase) ||
				strings.Contains(strings.ToLower(c.Username), phrase) {
				contacts = append(contacts, c)
				limit--
			}
			if limit < 0 {
				return domain.ErrLimitReached
			}
			return nil
		})
		return nil
	})
	return contacts
}
