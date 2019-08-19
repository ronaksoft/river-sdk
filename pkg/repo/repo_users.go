package repo

import (
	"fmt"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
	"github.com/dgraph-io/badger"
	"go.uber.org/zap"
	"strings"
	"time"
)

const (
	prefixUsers    = "USERS"
	prefixContacts = "CONTACTS"
)

type repoUsers struct {
	*repository
}

func (r *repoUsers) getUserKey(userID int64) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%021d", prefixUsers, userID))
}

func (r *repoUsers) getUserByKey(userKey []byte) *msg.User {
	user := new(msg.User)
	err := r.badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get(userKey)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return user.Unmarshal(val)
		})
	})
	if err != nil {
		return nil
	}
	return user
}

func (r *repoUsers) getContactKey(userID int64) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%021d", prefixContacts, userID))
}

func (r *repoUsers) getContactByKey(contactKey []byte) *msg.ContactUser {
	contactUser := new(msg.ContactUser)
	err := r.badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get(contactKey)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return contactUser.Unmarshal(val)
		})
	})
	if err != nil {
		return nil
	}
	return contactUser
}

func (r *repoUsers) readFromDb(userID int64) *msg.User {
	user := new(msg.User)
	err := r.badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get(r.getUserKey(userID))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return user.Unmarshal(val)
		})
	})
	if err != nil {
		return nil
	}
	return user
}

func (r *repoUsers) readFromCache(userID int64) *msg.User {
	user := new(msg.User)
	keyID := fmt.Sprintf("OBJ.USER.{%d}", userID)

	if jsonGroup, err := lCache.Get(keyID); err != nil || len(jsonGroup) == 0 {
		user := r.readFromDb(userID)
		if user == nil {
			return nil
		}
		jsonGroup, _ = user.Marshal()
		_ = lCache.Set(keyID, jsonGroup)
		return user
	} else {
		_ = user.Unmarshal(jsonGroup)
	}
	return user
}

func (r *repoUsers) readManyFromCache(userIDs []int64) []*msg.User {
	users := make([]*msg.User, 0, len(userIDs))
	for _, userID := range userIDs {
		if user := r.readFromCache(userID); user != nil {
			delta := time.Now().Unix() - user.LastSeen
			switch {
			case delta < domain.Minute:
				user.Status = msg.UserStatusOnline
			case delta < domain.Week:
				user.Status = msg.UserStatusRecently
			case delta < domain.Month:
				user.Status = msg.UserStatusLastWeek
			case delta < domain.TwoMonth:
				user.Status = msg.UserStatusLastMonth
			default:
				user.Status = msg.UserStatusOffline
			}
			users = append(users, user)
		}
	}
	return users
}

func (r *repoUsers) deleteFromCache(userIDs ...int64) {
	for _, userID := range userIDs {
		_ = lCache.Delete(fmt.Sprintf("OBJ.USER.{%d}", userID))
	}

}

func (r *repoUsers) Get(userID int64) *msg.User {
	return r.readFromCache(userID)
}

func (r *repoUsers) GetMany(userIDs []int64) []*msg.User {
	return r.readManyFromCache(userIDs)
}

func (r *repoUsers) GetContact(userID int64) *msg.ContactUser {
	contactUser := new(msg.ContactUser)
	err := r.badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get(r.getContactKey(userID))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return contactUser.Unmarshal(val)
		})
	})
	if err != nil {
		return nil
	}
	return contactUser
}

func (r *repoUsers) GetManyContactUsers(userIDs []int64) []*msg.ContactUser {
	contactUsers := make([]*msg.ContactUser, 0, len(userIDs))
	for _, userID := range userIDs {
		contactUsers = append(contactUsers, r.GetContact(userID))
	}

	return contactUsers
}

func (r *repoUsers) GetContacts() ([]*msg.ContactUser, []*msg.PhoneContact) {
	contactUsers := make([]*msg.ContactUser, 0, 100)
	phoneContacts := make([]*msg.PhoneContact, 0, 100)

	_ = r.badger.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = ronak.StrToByte(fmt.Sprintf("%s.", prefixContacts))
		it := txn.NewIterator(opts)
		for it.Seek(r.getContactKey(0)); it.ValidForPrefix(opts.Prefix); it.Next() {
			contactUser := new(msg.ContactUser)
			phoneContact := new(msg.PhoneContact)
			_ = it.Item().Value(func(val []byte) error {
				err := contactUser.Unmarshal(val)
				if err != nil {
					return err
				}
				return nil
			})
			contactUsers = append(contactUsers, contactUser)
			phoneContact.FirstName = contactUser.FirstName
			phoneContact.LastName = contactUser.LastName
			phoneContact.Phone = contactUser.Phone
			phoneContact.ClientID = contactUser.ClientID
			phoneContacts = append(phoneContacts, phoneContact)
		}
		it.Close()
		return nil
	})

	return contactUsers, phoneContacts

}

func (r *repoUsers) GetPhoto(userID, photoID int64) *msg.UserPhoto {

	user := r.Get(userID)
	return user.Photo
}

func (r *repoUsers) Save(user *msg.User) {
	if user == nil {
		return
	}
	if alreadySaved(fmt.Sprintf("U.%d", user.ID), user) {
		return
	}

	userKey := r.getUserKey(user.ID)
	userBytes, _ := user.Marshal()
	_ = r.badger.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry(
			userKey, userBytes,
		))
	})
	_ = r.peerSearch.Index(ronak.ByteToStr(userKey), UserSearch{
		Type:      "user",
		FirstName: user.FirstName,
		LastName:  user.LastName,
		PeerID:    user.ID,
		Username:  user.Username,
	})
}

func (r *repoUsers) SaveMany(users []*msg.User) {

	userIDs := domain.MInt64B{}
	for _, v := range users {
		if alreadySaved(fmt.Sprintf("U.%d", v.ID), v) {
			continue
		}
		userIDs[v.ID] = true
	}
	defer r.deleteFromCache(userIDs.ToArray()...)

	for idx := range users {
		r.Save(users[idx])
	}

	return
}

func (r *repoUsers) SaveContact(contactUser *msg.ContactUser) {
	if contactUser == nil {
		return
	}
	userBytes, _ := contactUser.Marshal()
	contactKey := r.getContactKey(contactUser.ID)
	_ = r.badger.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry(
			contactKey, userBytes,
		))
	})

	_ = r.peerSearch.Index(ronak.ByteToStr(contactKey), ContactSearch{
		Type:      "contact",
		FirstName: contactUser.FirstName,
		LastName:  contactUser.LastName,
		Username:  contactUser.Username,
	})
}

func (r *repoUsers) UpdateAccessHash(accessHash uint64, peerID int64, peerType int32) {
	defer r.deleteFromCache(peerID)

	user := r.Get(peerID)
	if user == nil {
		return
	}
	user.AccessHash = accessHash
	Dialogs.updateAccessHash(accessHash, peerID, peerType)
	return
}

func (r *repoUsers) GetAccessHash(userID int64) (uint64, error) {
	user := r.Get(userID)
	if user == nil {
		return 0, domain.ErrDoesNotExists
	}
	return user.AccessHash, nil
}

func (r *repoUsers) UpdatePhoto(userPhoto *msg.UpdateUserPhoto) {
	if alreadySaved(fmt.Sprintf("UPHOTO.%d", userPhoto.UserID), userPhoto) {
		return
	}

	defer r.deleteFromCache(userPhoto.UserID)

	user := r.Get(userPhoto.UserID)
	if user == nil {
		return
	}
	user.Photo = userPhoto.Photo
	r.Save(user)
}

func (r *repoUsers) RemovePhoto(userID int64) {
	defer r.deleteFromCache(userID)
	user := r.Get(userID)
	if user == nil {
		return
	}
	user.Photo = nil
	r.Save(user)
}

func (r *repoUsers) UpdateProfile(userID int64, firstName, lastName, username, bio string) {
	user := r.Get(userID)
	if user == nil {
		return
	}
	defer r.deleteFromCache(userID)
	user.FirstName = firstName
	user.LastName = lastName
	user.Username = username
	user.Bio = bio

	r.Save(user)
}

func (r *repoUsers) UpdateContactInfo(userID int64, firstName, lastName string) {

	defer r.deleteFromCache(userID)

	contact := r.GetContact(userID)
	contact.FirstName = firstName
	contact.LastName = lastName
	r.SaveContact(contact)
}

func (r *repoUsers) SearchContacts(searchPhrase string) ([]*msg.ContactUser, []*msg.PhoneContact) {
	t1 := bleve.NewTermQuery("contact")
	t1.SetField("type")
	qs := make([]query.Query, 0, 2)
	for _, term := range strings.Fields(searchPhrase) {
		qs = append(qs, bleve.NewPrefixQuery(term), bleve.NewMatchQuery(term))
	}
	t2 := bleve.NewDisjunctionQuery(qs...)
	searchRequest := bleve.NewSearchRequest(bleve.NewConjunctionQuery(t1, t2))
	searchResult, _ := r.peerSearch.Search(searchRequest)
	contactUsers := make([]*msg.ContactUser, 0, 100)
	phoneContacts := make([]*msg.PhoneContact, 0, 100)
	for _, hit := range searchResult.Hits {
		contactUser := r.getContactByKey(ronak.StrToByte(hit.ID))
		if contactUser != nil {
			phoneContacts = append(phoneContacts, &msg.PhoneContact{
				ClientID:  contactUser.ClientID,
				FirstName: contactUser.FirstName,
				LastName:  contactUser.LastName,
				Phone:     contactUser.Phone,
			})
			contactUsers = append(contactUsers, contactUser)
		}
	}
	return contactUsers, phoneContacts
}

func (r *repoUsers) SearchNonContacts(searchPhrase string) []*msg.ContactUser {
	t1 := bleve.NewTermQuery("user")
	t1.SetField("type")
	qs := make([]query.Query, 0)
	for _, term := range strings.Fields(searchPhrase) {
		qs = append(qs, bleve.NewPrefixQuery(term), bleve.NewMatchQuery(term))
	}
	t2 := bleve.NewDisjunctionQuery(qs...)
	searchRequest := bleve.NewSearchRequest(bleve.NewConjunctionQuery(t1, t2))
	searchResult, _ := r.peerSearch.Search(searchRequest)
	contactUsers := make([]*msg.ContactUser, 0, 100)
	for _, hit := range searchResult.Hits {
		user := r.getUserByKey(ronak.StrToByte(hit.ID))
		if user != nil {
			contactUsers = append(contactUsers, &msg.ContactUser{
				ID:        user.ID,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Username:  user.Username,
			})
		}
	}
	return contactUsers
}

func (r *repoUsers) SearchUsers(searchPhrase string) []*msg.User {
	t1 := bleve.NewTermQuery("user")
	t1.SetField("type")
	qs := make([]query.Query, 0)
	for _, term := range strings.Fields(searchPhrase) {
		qs = append(qs, bleve.NewPrefixQuery(term), bleve.NewMatchQuery(term))
	}
	t2 := bleve.NewDisjunctionQuery(qs...)
	searchRequest := bleve.NewSearchRequest(bleve.NewConjunctionQuery(t1, t2))
	searchResult, _ := r.peerSearch.Search(searchRequest)
	users := make([]*msg.User, 0, 100)
	for _, hit := range searchResult.Hits {
		user := r.getUserByKey(ronak.StrToByte(hit.ID))
		if user != nil {
			users = append(users, user)
		}
	}
	return users
}

func (r *repoUsers) ReIndex() {
	err := r.badger.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = ronak.StrToByte(prefixUsers)
		it := txn.NewIterator(opts)
		for it.Rewind(); it.Valid(); it.Next() {
			_ = it.Item().Value(func(val []byte) error {
				user := new(msg.User)
				_ = user.Unmarshal(val)
				_ = r.peerSearch.Index(ronak.ByteToStr(r.getUserKey(user.ID)), UserSearch{
					Type:      "user",
					FirstName: user.FirstName,
					LastName:  user.LastName,
					PeerID:    user.ID,
					Username:  user.Username,
				})
				return nil
			})
		}
		it.Close()
		return nil
	})
	if err != nil {
		logs.Warn("Error On ReIndex Users", zap.Error(err))
	}
	err = r.badger.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = ronak.StrToByte(prefixContacts)
		it := txn.NewIterator(opts)
		for it.Rewind(); it.Valid(); it.Next() {
			_ = it.Item().Value(func(val []byte) error {
				contactUser := new(msg.ContactUser)
				_ = contactUser.Unmarshal(val)
				_ = r.peerSearch.Index(ronak.ByteToStr(r.getContactKey(contactUser.ID)), ContactSearch{
					Type:      "contact",
					FirstName: contactUser.FirstName,
					LastName:  contactUser.LastName,
					Username:  contactUser.Username,
				})
				return nil
			})
		}
		it.Close()
		return nil
	})
	if err != nil {
		logs.Warn("Error On ReIndex Contacts", zap.Error(err))
	}
}
