package repo

import (
	"fmt"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo/dto"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/dgraph-io/badger"
	"go.uber.org/zap"
)

const (
	prefixUsers = "USERS"
)

type repoUsers struct {
	*repository
}

func (r *repoUsers) getUserKey(userID int64) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%021d", prefixUsers, userID))
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

func (r *repoUsers) GetPhoto(userID, photoID int64) *msg.UserPhoto {
	r.mx.Lock()
	defer r.mx.Unlock()

	user := r.Get(userID)
	return user.Photo
}

func (r *repoUsers) Save(user *msg.User) {
	if alreadySaved(fmt.Sprintf("U.%d", user.ID), user) {
		return
	}
	r.mx.Lock()
	defer r.mx.Unlock()

	if user == nil {
		return
	}

	userBytes, _ := user.Marshal()
	_ = r.badger.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry(
			r.getUserKey(user.ID), userBytes,
		))
	})
}

func (r *repoUsers) SaveMany(users []*msg.User) {
	r.mx.Lock()
	defer r.mx.Unlock()

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

func (r *repoUsers) UpdateAccessHash(accessHash uint64, peerID int64, peerType int32) error {
	defer r.deleteFromCache(peerID)

	r.mx.Lock()
	defer r.mx.Unlock()

	user := r.Get(peerID)
	if user == nil {
		return domain.ErrDoesNotExists
	}
	user.AccessHash = accessHash
	return Dialogs.updateAccessHash(accessHash, peerID, peerType)
}

func (r *repoUsers) GetAccessHash(userID int64) (uint64, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

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
	r.mx.Lock()
	defer r.mx.Unlock()

	defer r.deleteFromCache(userPhoto.UserID)

	user := r.Get(userPhoto.UserID)
	if user == nil {
		return
	}
	user.Photo = userPhoto.Photo
	r.Save(user)
}

func (r *repoUsers) RemovePhoto(userID int64)  {
	r.mx.Lock()
	defer r.mx.Unlock()
	defer r.deleteFromCache(userID)
	user := r.Get(userID)
	if user == nil {
		return
	}
	user.Photo = nil
	r.Save(user)
}

func (r *repoUsers) UpdateProfile(userID int64, firstName, lastName, username, bio string) {
	r.mx.Lock()
	defer r.mx.Unlock()

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





// OLD

func (r *repoUsers) SaveContact(user *msg.ContactUser) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if user == nil {
		logs.Debug("RepoUsers::SaveContactUser()",
			zap.String("User", "user is null"),
		)
		return domain.ErrNotFound
	}

	logs.Debug("RepoUsers::SaveContactUser()",
		zap.String("Name", fmt.Sprintf("%s %s", user.FirstName, user.LastName)),
		zap.Int64("userID", user.ID),
	)

	u := new(dto.Users)
	u.MapFromContactUser(user)

	eu := new(dto.Users)
	r.db.Find(eu, u.ID)

	if eu.ID == 0 {

		return r.db.Create(u).Error
	}

	// WARNING when update with struct, GORM will only update those fields that with non blank value
	// fix bug : if we had user and later we add its contact with fname or lname empty the fname and lname from user will remain unchanged
	return r.db.Table(u.TableName()).Where("ID=?", u.ID).Updates(map[string]interface{}{
		"FirstName":  u.FirstName,
		"LastName":   u.LastName,
		"AccessHash": u.AccessHash,
		"Phone":      u.Phone,
		"Username":   u.Username,
		"ClientID":   u.ClientID,
		"IsContact":  u.IsContact,
		"Photo":      u.Photo,
	}).Error
}

func (r *repoUsers) UpdatePhoneContact(user *msg.PhoneContact) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if user == nil {
		logs.Debug("RepoUsers::SaveContactUser()",
			zap.String("User", "user is null"),
		)
		return domain.ErrNotFound
	}
	dtoUser := new(dto.Users)
	// just need to edit existing contacts info
	return r.db.Table(dtoUser.TableName()).Where("Phone=?", user.Phone).Updates(map[string]interface{}{
		"FirstName": user.FirstName,
		"LastName":  user.LastName,
	}).Error

}

func (r *repoUsers) GetManyContactUsers(userIDs []int64) []*msg.ContactUser {
	r.mx.Lock()
	defer r.mx.Unlock()

	logs.Debug("Users::GetManyContactUsers()",
		zap.Int64s("UserIDs", userIDs),
	)
	pbUsers := make([]*msg.ContactUser, 0, len(userIDs))
	users := make([]dto.Users, 0, len(userIDs))

	err := r.db.Where("IsContact = 1 AND ID in (?)", userIDs).Find(&users).Error
	if err != nil {
		logs.Error("Users::GetManyContactUsers()-> fetch user entities", zap.Error(err))
		return nil
	}

	for _, v := range users {
		tmp := new(msg.ContactUser)
		v.MapToContactUser(tmp)
		pbUsers = append(pbUsers, tmp)

	}

	return pbUsers
}

func (r *repoUsers) GetContacts() ([]*msg.ContactUser, []*msg.PhoneContact) {
	r.mx.Lock()
	defer r.mx.Unlock()

	logs.Debug("Users::GetContacts()")

	users := make([]dto.Users, 0)

	err := r.db.Where("IsContact = 1").Find(&users).Error
	if err != nil {
		logs.Error("Users::GetContacts()-> fetch user entities", zap.Error(err))
		return nil, nil
	}
	pbUsers := make([]*msg.ContactUser, 0)
	pbContacts := make([]*msg.PhoneContact, 0)
	for _, v := range users {
		tmpUser := new(msg.ContactUser)
		tmpContact := new(msg.PhoneContact)
		v.MapToContactUser(tmpUser)
		v.MapToPhoneContact(tmpContact)
		pbUsers = append(pbUsers, tmpUser)
		pbContacts = append(pbContacts, tmpContact)
	}

	return pbUsers, pbContacts

}

func (r *repoUsers) SearchContacts(searchPhrase string) ([]*msg.ContactUser, []*msg.PhoneContact) {
	r.mx.Lock()
	defer r.mx.Unlock()

	logs.Debug("Users::SearchContacts()")

	p := "%" + searchPhrase + "%"
	users := make([]dto.Users, 0)
	err := r.db.Where("AccessHash <> 0 And IsContact = 1 AND (FirstName LIKE ? OR LastName LIKE ? OR Phone LIKE ? OR Username LIKE ?)", p, p, p, p).Find(&users).Error
	if err != nil {
		logs.Error("Users::SearchContacts()-> fetch user entities", zap.Error(err))
		return nil, nil // , err
	}
	pbUsers := make([]*msg.ContactUser, 0)
	pbContacts := make([]*msg.PhoneContact, 0)
	for _, v := range users {
		tmpUser := new(msg.ContactUser)
		tmpContact := new(msg.PhoneContact)
		v.MapToContactUser(tmpUser)
		v.MapToPhoneContact(tmpContact)
		pbUsers = append(pbUsers, tmpUser)
		pbContacts = append(pbContacts, tmpContact)
	}

	return pbUsers, pbContacts

}

func (r *repoUsers) UpdateContactInfo(userID int64, firstName, lastName string) error {
	defer r.deleteFromCache(userID)
	r.mx.Lock()
	defer r.mx.Unlock()

	logs.Debug("Users::UpdateContactInfo()",
		zap.Int64("userID", userID),
	)
	dtoUser := new(dto.Users)
	err := r.db.Table(dtoUser.TableName()).Where("ID=?", userID).Updates(map[string]interface{}{
		"FirstName": firstName,
		"LasttName": lastName,
	}).Error

	return err
}

func (r *repoUsers) SearchUsers(searchPhrase string) []*msg.User {
	r.mx.Lock()
	defer r.mx.Unlock()

	logs.Debug("Users::SearchContacts()")

	p := "%" + searchPhrase + "%"
	users := make([]dto.Users, 0)
	err := r.db.Where("FirstName LIKE ? OR LastName LIKE ? OR Phone LIKE ? OR Username LIKE ?", p, p, p, p).Find(&users).Error
	if err != nil {
		logs.Warn("Users::SearchUsers()-> fetch user entities", zap.Error(err))
		return nil
	}
	pbUsers := make([]*msg.User, 0)
	for _, v := range users {
		tmpUser := new(msg.User)
		v.MapToUser(tmpUser)
		pbUsers = append(pbUsers, tmpUser)
	}

	return pbUsers
}

func (r *repoUsers) UpdateAccountPhotoPath(userID, photoID int64, isBig bool, filePath string) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	e := new(dto.UsersPhoto)

	if isBig {
		return r.db.Table(e.TableName()).Where("userID = ? AND PhotoID = ?", userID, photoID).Updates(map[string]interface{}{
			"BigFilePath": filePath,
		}).Error
	}

	return r.db.Table(e.TableName()).Where("userID = ? AND PhotoID = ?", userID, photoID).Updates(map[string]interface{}{
		"SmallFilePath": filePath,
	}).Error

}

func (r *repoUsers) SearchNonContactsWithIDs(ids []int64, searchPhrase string) []*msg.ContactUser {
	r.mx.Lock()
	defer r.mx.Unlock()

	logs.Debug("Users::SearchNonContactsWithIDs()")

	p := "%" + searchPhrase + "%"
	users := make([]dto.Users, 0)
	err := r.db.Where("IsContact = 0 AND ID in (?) AND (FirstName LIKE ? OR LastName LIKE ? OR Phone LIKE ? OR Username LIKE ?)", ids, p, p, p, p).Find(&users).Error
	if err != nil {
		logs.Error("Users::SearchNonContactsWithIDs()-> fetch user entities", zap.Error(err))
		return nil
	}
	pbUsers := make([]*msg.ContactUser, 0)
	for _, v := range users {
		tmpUser := new(msg.ContactUser)
		v.MapToContactUser(tmpUser)
		pbUsers = append(pbUsers, tmpUser)
	}

	return pbUsers
}
