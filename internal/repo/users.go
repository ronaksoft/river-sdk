package repo

import (
	"fmt"
	"git.ronaksoft.com/river/msg/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/tools"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
	"github.com/dgraph-io/badger/v2"
	"go.uber.org/zap"
	"strings"
	"time"
)

const (
	prefixUsers             = "USERS"
	prefixContacts          = "CONTACTS"
	prefixPhoneContacts     = "PH_CONTACTS"
	prefixUsersPhotoGallery = "USERS_PHG"
)

type repoUsers struct {
	*repository
}

func getUserKey(userID int64) []byte {
	return domain.StrToByte(fmt.Sprintf("%s.%021d", prefixUsers, userID))
}

func getUserByKey(txn *badger.Txn, userKey []byte) (*msg.User, error) {
	user := new(msg.User)
	item, err := txn.Get(userKey)
	if err != nil {
		return nil, err
	}
	err = item.Value(func(val []byte) error {
		return user.Unmarshal(val)
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

func getContactKey(teamID, userID int64) []byte {
	return domain.StrToByte(fmt.Sprintf("%s.%021d.%021d", prefixContacts, teamID, userID))
}

func getContactTeamKey(teamID int64) []byte {
	return domain.StrToByte(fmt.Sprintf("%s.%021d", prefixContacts, teamID))
}

func getContactByKey(txn *badger.Txn, contactKey []byte) (*msg.ContactUser, error) {
	contactUser := new(msg.ContactUser)
	item, err := txn.Get(contactKey)
	if err != nil {
		return nil, err
	}
	err = item.Value(func(val []byte) error {
		return contactUser.Unmarshal(val)
	})
	if err != nil {
		return nil, err
	}
	return contactUser, nil
}

func getPhoneContactKey(phone string) []byte {
	return domain.StrToByte(fmt.Sprintf("%s.%s", prefixPhoneContacts, phone))
}

func getUserPhotoGalleryKey(userID, photoID int64) []byte {
	return domain.StrToByte(fmt.Sprintf("%s.%021d.%021d", prefixUsersPhotoGallery, userID, photoID))
}

func getUserPhotoGalleryPrefix(userID int64) []byte {
	return domain.StrToByte(fmt.Sprintf("%s.%021d.", prefixUsersPhotoGallery, userID))
}

func saveUser(txn *badger.Txn, user *msg.User) error {
	userKey := getUserKey(user.ID)
	if user.Photo == nil {
		_ = deleteAllUserPhotos(txn, user.ID)
	} else if user.Photo != nil && len(user.PhotoGallery) == 0 {
		// then it not full object
		currentUser, _ := getUserByKey(txn, userKey)
		if currentUser != nil && len(currentUser.PhotoGallery) > 0 {
			user.PhotoGallery = currentUser.PhotoGallery
		}
		err := saveUserPhotos(txn, user.ID, user.Photo)
		if err != nil {
			return err
		}
	} else if len(user.PhotoGallery) > 0 {
		err := saveUserPhotos(txn, user.ID, user.PhotoGallery...)
		if err != nil {
			return err
		}
	}

	userBytes, _ := user.Marshal()
	err := txn.SetEntry(badger.NewEntry(userKey, userBytes))
	if err != nil {
		return err
	}

	if user.ID == r.selfUserID {
		indexPeer(
			domain.ByteToStr(userKey),
			UserSearch{
				Type:      "user",
				FirstName: "Saved",
				LastName:  "Messages",
				PeerID:    user.ID,
				Username:  "savedmessages",
			},
		)
	} else {
		indexPeer(
			domain.ByteToStr(userKey),
			UserSearch{
				Type:      "user",
				FirstName: user.FirstName,
				LastName:  user.LastName,
				PeerID:    user.ID,
				Username:  user.Username,
			},
		)
	}

	return nil
}

func saveContact(txn *badger.Txn, teamID int64, contactUser *msg.ContactUser) error {
	userBytes, _ := contactUser.Marshal()
	contactKey := getContactKey(teamID, contactUser.ID)
	err := txn.SetEntry(badger.NewEntry(
		contactKey, userBytes,
	))
	if err != nil {
		return err
	}
	indexPeer(
		domain.ByteToStr(contactKey),
		ContactSearch{
			Type:      "contact",
			FirstName: contactUser.FirstName,
			LastName:  contactUser.LastName,
			Username:  contactUser.Username,
			Phone:     contactUser.Phone,
			TeamID:    fmt.Sprintf("%d", teamID),
		},
	)
	err = saveUserPhotos(txn, contactUser.ID, contactUser.Photo)
	if err != nil {
		return err
	}
	return nil
}

func savePhoneContact(txn *badger.Txn, phoneContact *msg.PhoneContact) error {
	bytes, _ := phoneContact.Marshal()
	dbKey := getPhoneContactKey(phoneContact.Phone)
	err := txn.SetEntry(badger.NewEntry(
		dbKey, bytes,
	))
	if err != nil {
		return err
	}
	return nil
}

func (r *repoUsers) Get(userID int64) (user *msg.User, err error) {
	err = badgerView(func(txn *badger.Txn) error {
		user, err = getUserByKey(txn, getUserKey(userID))
		if err != nil {
			return err
		}
		updateStatus(user)
		return nil
	})
	return
}

func (r *repoUsers) GetMany(userIDs []int64) ([]*msg.User, error) {
	users := make([]*msg.User, 0, len(userIDs))
	err := badgerView(func(txn *badger.Txn) error {
		for _, userID := range userIDs {
			if userID == 0 {
				continue
			}
			user, err := getUserByKey(txn, getUserKey(userID))
			switch err {
			case nil:
			case badger.ErrKeyNotFound:
				continue
			default:
				return err
			}
			updateStatus(user)
			users = append(users, user)
		}
		return nil
	})
	return users, err
}

func updateStatus(u *msg.User) {
	delta := time.Now().Unix() - u.LastSeen
	if u.Status == msg.UserStatusOnline && delta < domain.Minute {
		return
	}
	switch {
	case delta < domain.Hour:
		u.Status = msg.UserStatusRecently
	case delta < domain.Week:
		u.Status = msg.UserStatusLastWeek
	case delta < domain.Month:
		u.Status = msg.UserStatusLastMonth
	default:
		u.Status = msg.UserStatusOffline
	}
}

func (r *repoUsers) Save(users ...*msg.User) error {
	userIDs := domain.MInt64B{}
	for _, v := range users {
		if strings.TrimSpace(v.FirstName) == "" && strings.TrimSpace(v.LastName) == "" {
			continue
		}
		userIDs[v.ID] = true
	}
	return badgerUpdate(func(txn *badger.Txn) error {
		for idx := range users {
			err := saveUser(txn, users[idx])
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *repoUsers) UpdateBlocked(peerID int64, blocked bool) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		user, err := getUserByKey(txn, getUserKey(peerID))
		switch err {
		case nil:
		case badger.ErrKeyNotFound:
			return nil
		default:
			return err
		}
		user.Blocked = blocked
		return saveUser(txn, user)
	})
}

func (r *repoUsers) GetAccessHash(userID int64) (accessHash uint64, err error) {
	err = badgerView(func(txn *badger.Txn) error {
		user, err := getUserByKey(txn, getUserKey(userID))
		if err != nil {
			return err
		}
		if user != nil {
			accessHash = user.AccessHash
		}
		return nil
	})
	return
}

func (r *repoUsers) UpdateProfile(userID int64, firstName, lastName, username, bio, phone string) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		user, err := getUserByKey(txn, getUserKey(userID))
		switch err {
		case nil:
		case badger.ErrKeyNotFound:
			return nil
		default:
			return err
		}
		user.FirstName = firstName
		user.LastName = lastName
		user.Username = username
		user.Phone = phone
		user.Bio = bio
		return saveUser(txn, user)
	})
}

func (r *repoUsers) SearchUsers(searchPhrase string) []*msg.User {
	users := make([]*msg.User, 0, 100)
	if r.peerSearch == nil {
		return users
	}
	t1 := bleve.NewTermQuery("user")
	t1.SetField("type")
	qs := make([]query.Query, 0)
	for _, term := range strings.Fields(searchPhrase) {
		qs = append(qs, bleve.NewPrefixQuery(term), bleve.NewMatchQuery(term))
	}
	t2 := bleve.NewDisjunctionQuery(qs...)
	searchRequest := bleve.NewSearchRequest(bleve.NewConjunctionQuery(t1, t2))
	searchResult, _ := r.peerSearch.Search(searchRequest)

	err := badgerView(func(txn *badger.Txn) error {
		for _, hit := range searchResult.Hits {
			user, err := getUserByKey(txn, domain.StrToByte(hit.ID))
			if err == nil && user != nil {
				users = append(users, user)
			}
		}
		return nil
	})
	logs.ErrorOnErr("RepoUser got error on search users", err,
		zap.String("Phrase", searchPhrase),
	)
	return users
}

func (r *repoUsers) GetContact(teamID, userID int64) (*msg.ContactUser, error) {
	contactUser := new(msg.ContactUser)
	err := badgerView(func(txn *badger.Txn) error {
		item, err := txn.Get(getContactKey(teamID, userID))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return contactUser.Unmarshal(val)
		})
	})

	return contactUser, err
}

func (r *repoUsers) GetContacts(teamID int64) ([]*msg.ContactUser, []*msg.PhoneContact) {
	contactUsers := make([]*msg.ContactUser, 0, 100)
	phoneContacts := make([]*msg.PhoneContact, 0, 100)

	_ = badgerView(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = domain.StrToByte(fmt.Sprintf("%s.%021d.", prefixContacts, teamID))
		it := txn.NewIterator(opts)
		for it.Rewind(); it.ValidForPrefix(opts.Prefix); it.Next() {
			contactUser := &msg.ContactUser{}
			phoneContact := &msg.PhoneContact{}
			_ = it.Item().Value(func(val []byte) error {
				err := contactUser.Unmarshal(val)
				if err != nil {
					return err
				}
				return nil
			})
			user, _ := getUserByKey(txn, getUserKey(contactUser.ID))
			if user != nil {
				contactUser.Photo = user.Photo
			}
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

func (r *repoUsers) SearchContacts(teamID int64, searchPhrase string) ([]*msg.ContactUser, []*msg.PhoneContact) {
	contactUsers := make([]*msg.ContactUser, 0, 100)
	phoneContacts := make([]*msg.PhoneContact, 0, 100)
	if r.peerSearch == nil {
		return contactUsers, phoneContacts
	}
	t1 := bleve.NewTermQuery("contact")
	t1.SetField("type")
	qs := make([]query.Query, 0, 2)
	for _, term := range strings.Fields(searchPhrase) {
		qs = append(qs, bleve.NewPrefixQuery(term), bleve.NewMatchQuery(term))
	}
	t2 := bleve.NewDisjunctionQuery(qs...)
	t3 := bleve.NewTermQuery(fmt.Sprintf("%d", tools.AbsInt64(teamID)))
	t3.SetField("team_id")
	searchRequest := bleve.NewSearchRequest(bleve.NewConjunctionQuery(t1, t2, t3))
	searchResult, _ := r.peerSearch.Search(searchRequest)

	_ = badgerView(func(txn *badger.Txn) error {
		for _, hit := range searchResult.Hits {
			contactUser, err := getContactByKey(txn, domain.StrToByte(hit.ID))
			if err == nil && contactUser != nil {
				phoneContacts = append(phoneContacts, &msg.PhoneContact{
					ClientID:  contactUser.ClientID,
					FirstName: contactUser.FirstName,
					LastName:  contactUser.LastName,
					Phone:     contactUser.Phone,
				})
				contactUsers = append(contactUsers, contactUser)
			}
		}
		return nil
	})
	return contactUsers, phoneContacts
}

func (r *repoUsers) SearchNonContacts(searchPhrase string) []*msg.ContactUser {
	contactUsers := make([]*msg.ContactUser, 0, 100)
	if r.peerSearch == nil {
		return contactUsers
	}
	t1 := bleve.NewTermQuery("user")
	t1.SetField("type")
	qs := make([]query.Query, 0)
	for _, term := range strings.Fields(searchPhrase) {
		qs = append(qs, bleve.NewPrefixQuery(term), bleve.NewMatchQuery(term))
	}
	t2 := bleve.NewDisjunctionQuery(qs...)
	searchRequest := bleve.NewSearchRequest(bleve.NewConjunctionQuery(t1, t2))
	searchResult, _ := r.peerSearch.Search(searchRequest)

	_ = badgerView(func(txn *badger.Txn) error {
		for _, hit := range searchResult.Hits {
			user, _ := getUserByKey(txn, domain.StrToByte(hit.ID))
			if user != nil {
				contactUsers = append(contactUsers, &msg.ContactUser{
					ID:        user.ID,
					FirstName: user.FirstName,
					LastName:  user.LastName,
					Username:  user.Username,
				})
			}
		}
		return nil
	})
	return contactUsers
}

func (r *repoUsers) UpdateContactInfo(teamID int64, userID int64, firstName, lastName string) error {
	err := badgerUpdate(func(txn *badger.Txn) error {
		contact, err := getContactByKey(txn, getContactKey(teamID, userID))
		if err != nil {
			return err
		}
		contact.FirstName = firstName
		contact.LastName = lastName
		return saveContact(txn, teamID, contact)
	})
	logs.ErrorOnErr("RepoUser got error on update contact info", err)
	return err
}

func (r *repoUsers) SaveContact(teamID int64, contactUsers ...*msg.ContactUser) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		for _, contactUser := range contactUsers {
			err := saveContact(txn, teamID, contactUser)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *repoUsers) SavePhoneContact(phoneContacts ...*msg.PhoneContact) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		for _, phoneContact := range phoneContacts {
			err := savePhoneContact(txn, phoneContact)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *repoUsers) DeletePhoneContact(phoneContacts ...*msg.PhoneContact) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		for _, phoneContact := range phoneContacts {
			_ = txn.Delete(getPhoneContactKey(phoneContact.Phone))
		}
		return nil
	})
}

func (r *repoUsers) GetPhoneContacts(limit int) ([]*msg.PhoneContact, error) {
	phoneContacts := make([]*msg.PhoneContact, 0, limit)
	err := badgerView(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = domain.StrToByte(fmt.Sprintf("%s.", prefixPhoneContacts))
		it := txn.NewIterator(opts)
		for it.Rewind(); it.ValidForPrefix(opts.Prefix); it.Next() {
			if limit <= 0 {
				break
			}
			limit--
			phoneContact := &msg.PhoneContact{}
			_ = it.Item().Value(func(val []byte) error {
				err := phoneContact.Unmarshal(val)
				if err != nil {
					return err
				}
				return nil
			})
			phoneContacts = append(phoneContacts, phoneContact)
		}
		it.Close()
		return nil
	})
	return phoneContacts, err
}

func (r *repoUsers) DeleteContact(teamID int64, contactIDs ...int64) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		for _, contactID := range contactIDs {
			err := txn.Delete(getContactKey(teamID, contactID))
			logs.ErrorOnErr("DeleteContact : error on delete contact", err, zap.Int64("ContactID", contactID))
		}
		return nil
	})
}

func (r *repoUsers) DeleteAllContacts(teamID int64) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = domain.StrToByte(fmt.Sprintf("%s.%021d.", prefixContacts, teamID))
		it := txn.NewIterator(opts)
		for it.Rewind(); it.ValidForPrefix(opts.Prefix); it.Next() {
			_ = txn.Delete(it.Item().KeyCopy(nil))
		}
		it.Close()
		return nil
	})
}

func (r *repoUsers) UpdatePhoto(userID int64, userPhoto *msg.UserPhoto) error {
	err := badgerUpdate(func(txn *badger.Txn) error {
		user, err := getUserByKey(txn, getUserKey(userID))
		switch err {
		case nil:
		case badger.ErrKeyNotFound:
			logs.Warn("We got error on update user's photo, but user does not exists",
				zap.Int64("UserID", userID),
			)
			return nil
		default:
			return err
		}
		user.Photo = userPhoto
		return saveUser(txn, user)
	})
	logs.ErrorOnErr("RepoUser got error on update photo", err)
	return err
}

func (r *repoUsers) SavePhotoGallery(userID int64, photos ...*msg.UserPhoto) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		for _, photo := range photos {
			if photo != nil {
				key := getUserPhotoGalleryKey(userID, photo.PhotoID)
				bytes, _ := photo.Marshal()
				_ = txn.SetEntry(badger.NewEntry(key, bytes))
			}
		}
		return nil
	})
}

func (r *repoUsers) RemovePhotoGallery(userID int64, photoIDs ...int64) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		for _, photoID := range photoIDs {
			_ = txn.Delete(getUserPhotoGalleryKey(userID, photoID))
		}
		return nil
	})
}

func (r *repoUsers) GetPhotoGallery(userID int64) []*msg.UserPhoto {
	photoGallery := make([]*msg.UserPhoto, 0, 4)
	_ = badgerView(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = getUserPhotoGalleryPrefix(userID)
		it := txn.NewIterator(opts)
		for it.Rewind(); it.ValidForPrefix(opts.Prefix); it.Next() {
			_ = it.Item().Value(func(val []byte) error {
				userPhoto := new(msg.UserPhoto)
				err := userPhoto.Unmarshal(val)
				if err != nil {
					return err
				}
				photoGallery = append(photoGallery, userPhoto)
				return nil
			})
		}
		it.Close()
		return nil
	})
	return photoGallery
}

func (r *repoUsers) ReIndex(teamID int64) {
	err := domain.Try(10, time.Second, func() error {
		if r.peerSearch == nil {
			return domain.ErrDoesNotExists
		}
		return nil
	})
	if err != nil {
		return
	}

	_ = badgerView(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = domain.StrToByte(prefixUsers)
		it := txn.NewIterator(opts)
		for it.Rewind(); it.Valid(); it.Next() {
			_ = it.Item().Value(func(val []byte) error {
				user := new(msg.User)
				_ = user.Unmarshal(val)
				key := domain.ByteToStr(getUserKey(user.ID))
				if d, _ := r.peerSearch.Document(key); d == nil {
					if user.ID == r.repository.selfUserID {
						indexPeer(
							key,
							UserSearch{
								Type:      "user",
								FirstName: "Saved",
								LastName:  "Messages",
								PeerID:    user.ID,
								Username:  "savedmessages",
							},
						)
					} else {
						indexPeer(
							key,
							UserSearch{
								Type:      "user",
								FirstName: user.FirstName,
								LastName:  user.LastName,
								PeerID:    user.ID,
								Username:  user.Username,
							},
						)
					}
				}

				return nil
			})
		}
		it.Close()

		opts = badger.DefaultIteratorOptions
		opts.Prefix = domain.StrToByte(prefixContacts)
		it = txn.NewIterator(opts)
		for it.Rewind(); it.Valid(); it.Next() {
			_ = it.Item().Value(func(val []byte) error {
				contactUser := new(msg.ContactUser)
				_ = contactUser.Unmarshal(val)
				indexPeer(
					domain.ByteToStr(getContactKey(teamID, contactUser.ID)),
					ContactSearch{
						Type:      "contact",
						FirstName: contactUser.FirstName,
						LastName:  contactUser.LastName,
						Username:  contactUser.Username,
						TeamID:    fmt.Sprintf("%d", teamID),
					},
				)
				return nil
			})
		}
		it.Close()
		return nil
	})
}
