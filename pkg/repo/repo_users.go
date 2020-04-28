package repo

import (
	"fmt"
	msg "git.ronaksoftware.com/river/msg/chat"
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
	prefixUsers             = "USERS"
	prefixContacts          = "CONTACTS"
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

func getContactKey(userID int64) []byte {
	return domain.StrToByte(fmt.Sprintf("%s.%021d", prefixContacts, userID))
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

	return nil
}

func saveContact(txn *badger.Txn, contactUser *msg.ContactUser) error {
	userBytes, _ := contactUser.Marshal()
	contactKey := getContactKey(contactUser.ID)
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
		},
	)
	err = saveUserPhotos(txn, contactUser.ID, contactUser.Photo)
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
		return nil
	})
	return
}

func (r *repoUsers) GetMany(userIDs []int64) ([]*msg.User, error) {
	timeNow := time.Now().Unix()
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

			delta := timeNow - user.LastSeen
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
		return nil
	})
	return users, err
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

func (r *repoUsers) UpdateAccessHash(accessHash uint64, peerID int64, peerType int32) error {
	if msg.PeerType(peerType) != msg.PeerUser {
		return nil
	}

	return badgerUpdate(func(txn *badger.Txn) error {
		user, err := getUserByKey(txn, getUserKey(peerID))
		if err != nil {
			return err
		}
		user.AccessHash = accessHash
		return updateDialogAccessHash(txn, accessHash, peerID, peerType)
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

func (r *repoUsers) GetContact(userID int64) (*msg.ContactUser, error) {
	contactUser := new(msg.ContactUser)
	err := badgerView(func(txn *badger.Txn) error {
		item, err := txn.Get(getContactKey(userID))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return contactUser.Unmarshal(val)
		})
	})

	return contactUser, err
}

func (r *repoUsers) GetContacts() ([]*msg.ContactUser, []*msg.PhoneContact) {
	contactUsers := make([]*msg.ContactUser, 0, 100)
	phoneContacts := make([]*msg.PhoneContact, 0, 100)

	_ = badgerView(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = domain.StrToByte(fmt.Sprintf("%s.", prefixContacts))
		it := txn.NewIterator(opts)
		for it.Seek(getContactKey(0)); it.ValidForPrefix(opts.Prefix); it.Next() {
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

func (r *repoUsers) SearchContacts(searchPhrase string) ([]*msg.ContactUser, []*msg.PhoneContact) {
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
	searchRequest := bleve.NewSearchRequest(bleve.NewConjunctionQuery(t1, t2))
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

func (r *repoUsers) UpdateContactInfo(userID int64, firstName, lastName string) error {
	err := badgerUpdate(func(txn *badger.Txn) error {
		contact, err := getContactByKey(txn, getContactKey(userID))
		if err != nil {
			return err
		}
		contact.FirstName = firstName
		contact.LastName = lastName
		return saveContact(txn, contact)
	})
	logs.ErrorOnErr("RepoUser got error on update contact info", err)
	return err
}

func (r *repoUsers) SaveContact(contactUsers ...*msg.ContactUser) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		for _, contactUser := range contactUsers {
			err := saveContact(txn, contactUser)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *repoUsers) DeleteContact(contactIDs ...int64) error {
	return badgerUpdate(func(txn *badger.Txn) error {
		for _, contactID := range contactIDs {
			_ = txn.Delete(getContactKey(contactID))
		}
		return nil
	})
}

func (r *repoUsers) DeleteAllContacts() error {
	return badgerUpdate(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = domain.StrToByte(fmt.Sprintf("%s.", prefixContacts))
		it := txn.NewIterator(opts)
		for it.Rewind(); it.ValidForPrefix(opts.Prefix); it.Next() {
			_ = txn.Delete(it.Item().Key())
		}
		it.Close()
		return nil
	})
}

func (r *repoUsers) UpdatePhoto(userID int64, userPhoto *msg.UserPhoto) error {
	err := badgerUpdate(func(txn *badger.Txn) error {
		user, err := getUserByKey(txn, getUserKey(userID))
		if err != nil {
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

func (r *repoUsers) ReIndex() {
	err := ronak.Try(10, time.Second, func() error {
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
					domain.ByteToStr(getContactKey(contactUser.ID)),
					ContactSearch{
						Type:      "contact",
						FirstName: contactUser.FirstName,
						LastName:  contactUser.LastName,
						Username:  contactUser.Username,
					},
				)
				return nil
			})
		}
		it.Close()
		return nil
	})
}
