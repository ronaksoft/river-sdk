package repo

import (
	"fmt"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/logs"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/repo/dto"
	"go.uber.org/zap"
)

// Users repoUsers interface
type Users interface {
	SaveUser(user *msg.User) error
	SaveContactUser(user *msg.ContactUser) error
	UpdatePhoneContact(user *msg.PhoneContact) error
	GetManyUsers(userIDs []int64) []*msg.User
	GetManyContactUsers(userIDs []int64) []*msg.ContactUser
	GetContacts() ([]*msg.ContactUser, []*msg.PhoneContact)
	SearchContacts(searchPhrase string) ([]*msg.ContactUser, []*msg.PhoneContact)
	GetAccessHash(userID int64) (uint64, error)
	UpdateAccessHash(accessHash int64, peerID int64, peerType int32) error
	GetUser(userID int64) *msg.User
	GetAnyUsers(userIDs []int64) []*msg.User
	SaveMany(users []*msg.User) error
	UpdateContactInfo(userID int64, firstName, lastName string) error
	SearchUsers(searchPhrase string) []*msg.User
	UpdateUserProfile(userID int64, req *msg.AccountUpdateProfile) error
	UpdateUsername(u *msg.UpdateUsername) error
	GetUserPhoto(userID, photoID int64) *dto.UserPhotos
	UpdateAccountPhotoPath(userID, photoID int64, isBig bool, filePath string) error
	SaveUserPhoto(userPhoto *msg.UpdateUserPhoto) error
	RemoveUserPhoto(userID int64) error
}

type repoUsers struct {
	*repository
}

// Save
func (r *repoUsers) SaveUser(user *msg.User) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if user == nil {
		logs.Debug("RepoUsers::SaveUser()",
			zap.String("User", "user is null"),
		)
		return domain.ErrNotFound
	}
	logs.Debug("RepoUsers::SaveUser()",
		zap.String("Name", fmt.Sprintf("%s %s", user.FirstName, user.LastName)),
		zap.Int64("UserID", user.ID),
	)

	// save user Photos
	if user.Photo != nil {

		dtoPhoto := new(dto.UserPhotos)
		r.db.Where(&dto.UserPhotos{UserID: user.ID, PhotoID: user.Photo.PhotoID}).Find(dtoPhoto)
		if dtoPhoto.UserID == 0 || dtoPhoto.PhotoID == 0 {
			dtoPhoto.Map(user.ID, user.Photo)
			r.db.Create(dtoPhoto)
		} else {
			dtoPhoto.Map(user.ID, user.Photo)
			r.db.Table(dtoPhoto.TableName()).Where("UserID=? AND PhotoID=?", user.ID, user.Photo.PhotoID).Update(dtoPhoto)
		}
	}

	u := new(dto.Users)
	u.MapFromUser(user)

	eu := new(dto.Users)
	r.db.Find(eu, u.ID)
	if eu.ID == 0 {
		return r.db.Create(u).Error
	}

	// if user not exist just add it no need to update existing user
	// just update user photo and its last status info
	if user.Photo != nil {
		// update user photo info if exist
		buff, err := user.Photo.Marshal()
		if err == nil {
			return r.db.Table(u.TableName()).Where("ID=?", u.ID).Updates(map[string]interface{}{
				"Photo":      buff,
				"Status":     int32(user.Status),
				"StatusTime": time.Now().Unix(),
			}).Error
		}
	} else {
		return r.db.Table(u.TableName()).Where("ID=?", u.ID).Updates(map[string]interface{}{
			"Photo":      []byte("[]"),
			"Status":     int32(user.Status),
			"StatusTime": time.Now().Unix(),
		}).Error
	}
	return nil
}

// SaveContactUser
func (r *repoUsers) SaveContactUser(user *msg.ContactUser) error {
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
		zap.Int64("UserID", user.ID),
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

// SavePhoneContact
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

// GetManyUsers
func (r *repoUsers) GetManyUsers(userIDs []int64) []*msg.User {
	r.mx.Lock()
	defer r.mx.Unlock()

	logs.Debug("Users::GetManyUsers()",
		zap.Int64s("UserIDs", userIDs),
	)

	pbUsers := make([]*msg.User, 0, len(userIDs))
	users := make([]dto.Users, 0, len(userIDs))

	err := r.db.Where("IsContact = 0 AND ID in (?)", userIDs).Find(&users).Error
	if err != nil {
		logs.Error("Users::GetManyUsers()-> fetch user entity", zap.Error(err))
		return nil //, err
	}

	for _, v := range users {
		tmp := new(msg.User)
		v.MapToUser(tmp)
		pbUsers = append(pbUsers, tmp)

	}

	return pbUsers
}

// GetManyContactUsers
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
		return nil //, err
	}

	for _, v := range users {
		tmp := new(msg.ContactUser)
		v.MapToContactUser(tmp)
		pbUsers = append(pbUsers, tmp)

	}

	return pbUsers
}

// GetContacts
func (r *repoUsers) GetContacts() ([]*msg.ContactUser, []*msg.PhoneContact) {
	r.mx.Lock()
	defer r.mx.Unlock()

	logs.Debug("Users::GetContacts()")

	users := make([]dto.Users, 0)

	err := r.db.Where("AccessHash <> 0").Find(&users).Error
	if err != nil {
		logs.Error("Users::GetContacts()-> fetch user entities", zap.Error(err))
		return nil, nil //, err
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

// SearchContacts
func (r *repoUsers) SearchContacts(searchPhrase string) ([]*msg.ContactUser, []*msg.PhoneContact) {
	r.mx.Lock()
	defer r.mx.Unlock()

	logs.Debug("Users::SearchContacts()")

	p := "%" + searchPhrase + "%"
	users := make([]dto.Users, 0)
	err := r.db.Where("AccessHash <> 0 AND (FirstName LIKE ? OR LastName LIKE ? OR Phone LIKE ? OR Username LIKE ?)", p, p, p, p).Find(&users).Error
	if err != nil {
		logs.Error("Users::SearchContacts()-> fetch user entities", zap.Error(err))
		return nil, nil //, err
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

func (r *repoUsers) GetAccessHash(userID int64) (uint64, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	eu := new(dto.Users)
	r.db.Find(eu, userID)
	if eu.ID > 0 {
		return uint64(eu.AccessHash), nil
	}
	return 0, domain.ErrNotFound
}

func (r *repoUsers) UpdateAccessHash(accessHash int64, peerID int64, peerType int32) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	ed := new(dto.Users)
	err := r.db.Table(ed.TableName()).Where("ID=? ", peerID).Updates(map[string]interface{}{
		"AccessHash": int64(accessHash),
	}).Error
	return err
}

// GetManyUsers
func (r *repoUsers) GetUser(userID int64) *msg.User {
	r.mx.Lock()
	defer r.mx.Unlock()

	logs.Debug("Users::GetUser()",
		zap.Int64("UserID", userID),
	)

	pbUser := new(msg.User)
	user := new(dto.Users)

	err := r.db.Find(user, userID).Error
	if err != nil {
		logs.Error("Users::GetUser()-> fetch user entity", zap.Error(err))
		return nil //, err
	}

	user.MapToUser(pbUser)

	return pbUser
}

// GetAnyUsers
func (r *repoUsers) GetAnyUsers(userIDs []int64) []*msg.User {
	r.mx.Lock()
	defer r.mx.Unlock()

	logs.Debug("Users::GetAnyUsers()",
		zap.Int64s("UserIDs", userIDs),
	)

	pbUsers := make([]*msg.User, 0, len(userIDs))
	users := make([]dto.Users, 0, len(userIDs))

	err := r.db.Where("ID in (?)", userIDs).Find(&users).Error
	if err != nil {
		logs.Error("Users::GetAnyUsers()-> fetch user entity", zap.Error(err))
		return nil //, err
	}

	for _, v := range users {
		tmp := new(msg.User)
		v.MapToUser(tmp)
		pbUsers = append(pbUsers, tmp)

	}

	return pbUsers
}

func (r *repoUsers) SaveMany(users []*msg.User) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	userIDs := domain.MInt64B{}
	for _, v := range users {
		userIDs[v.ID] = true
	}
	mapDTOUsers := make(map[int64]*dto.Users)
	dtoUsers := make([]dto.Users, 0)
	err := r.db.Where("ID in (?)", userIDs.ToArray()).Find(&dtoUsers).Error
	if err != nil {
		logs.Error("Users::SaveMany()-> fetch groups entity", zap.Error(err))
		return err
	}
	count := len(dtoUsers)
	for i := 0; i < count; i++ {
		mapDTOUsers[dtoUsers[i].ID] = &dtoUsers[i]
	}

	for _, v := range users {
		if dtoEntity, ok := mapDTOUsers[v.ID]; ok {
			dtoEntity.MapFromUser(v)
			err = r.db.Table(dtoEntity.TableName()).Where("ID=?", dtoEntity.ID).Update(dtoEntity).Error
		} else {
			dtoEntity := new(dto.Users)
			dtoEntity.MapFromUser(v)
			err = r.db.Create(dtoEntity).Error
		}
		if err != nil {
			logs.Error("Users::SaveMany()-> save group entity",
				zap.Int64("ID", v.ID),
				zap.String("FirstName", v.FirstName),
				zap.String("Lastname", v.LastName),
				zap.String("UserName", v.Username),
				zap.Error(err))
			break
		}
	}
	return err
}

func (r *repoUsers) UpdateContactInfo(userID int64, firstName, lastName string) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	logs.Debug("Users::UpdateContactInfo()",
		zap.Int64("UserID", userID),
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
		logs.Error("Users::SearchUsers()-> fetch user entities", zap.Error(err))
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

func (r *repoUsers) UpdateUserProfile(userID int64, req *msg.AccountUpdateProfile) error {
	r.mx.Lock()
	defer r.mx.Unlock()
	dtoUser := dto.Users{}
	// just need to edit existing contacts info
	return r.db.Table(dtoUser.TableName()).Where("ID=?", userID).Updates(map[string]interface{}{
		"FirstName": req.FirstName,
		"LastName":  req.LastName,
		"Bio":       req.Bio,
	}).Error
}

// SaveContactUser
func (r *repoUsers) UpdateUsername(u *msg.UpdateUsername) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if u == nil {
		logs.Debug("RepoUsers::SaveContactUser()",
			zap.String("User", "user is null"),
		)
		return domain.ErrNotFound
	}
	mdl := dto.Users{}

	return r.db.Table(mdl.TableName()).Where("ID=?", u.UserID).Updates(map[string]interface{}{
		"FirstName": u.FirstName,
		"LastName":  u.LastName,
		"Username":  u.Username,
		"Bio":       u.Bio,
	}).Error
}

func (r *repoUsers) GetUserPhoto(userID, photoID int64) *dto.UserPhotos {
	r.mx.Lock()
	defer r.mx.Unlock()

	logs.Debug("Users::GetUserPhoto()",
		zap.Int64("UserID", userID),
		zap.Int64("PhotoID", photoID),
	)

	dtoPhoto := dto.UserPhotos{}

	err := r.db.Where("UserID = ? AND PhotoID = ?", userID, photoID).First(&dtoPhoto).Error
	if err != nil {
		logs.Error("Users::GetUserPhoto()->fetch UserPhoto entity", zap.Error(err))
		return nil
	}

	return &dtoPhoto
}

func (r *repoUsers) UpdateAccountPhotoPath(userID, photoID int64, isBig bool, filePath string) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	e := new(dto.UserPhotos)

	if isBig {
		return r.db.Table(e.TableName()).Where("UserID = ? AND PhotoID = ?", userID, photoID).Updates(map[string]interface{}{
			"Big_FilePath": filePath,
		}).Error
	}

	return r.db.Table(e.TableName()).Where("UserID = ? AND PhotoID = ?", userID, photoID).Updates(map[string]interface{}{
		"Small_FilePath": filePath,
	}).Error

}

// SaveUserPhoto
func (r *repoUsers) SaveUserPhoto(userPhoto *msg.UpdateUserPhoto) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if userPhoto == nil {
		logs.Debug("RepoUsers::SaveUserPhoto()",
			zap.String("User", "user is null"),
		)
		return domain.ErrNotFound
	}

	var er error
	// save user Photos
	if userPhoto.Photo != nil {

		dtoPhoto := new(dto.UserPhotos)
		r.db.Where(&dto.UserPhotos{UserID: userPhoto.UserID, PhotoID: userPhoto.Photo.PhotoID}).Find(dtoPhoto)
		if dtoPhoto.UserID == 0 || dtoPhoto.PhotoID == 0 {
			dtoPhoto.Map(userPhoto.UserID, userPhoto.Photo)
			er = r.db.Create(dtoPhoto).Error
		} else {
			dtoPhoto.Map(userPhoto.UserID, userPhoto.Photo)
			er = r.db.Table(dtoPhoto.TableName()).Where("UserID=? AND PhotoID=?", userPhoto.UserID, userPhoto.Photo.PhotoID).Update(dtoPhoto).Error
		}
	}
	user := new(dto.Users)
	r.db.Find(user, userPhoto.UserID)
	if user.ID != 0 && userPhoto.Photo != nil {
		// update user photo info if exist
		buff, err := userPhoto.Photo.Marshal()
		if err == nil {
			return r.db.Table(user.TableName()).Where("ID=?", user.ID).Updates(map[string]interface{}{
				"Photo": buff,
			}).Error
		}
	}
	return er
}

// RemoveUserPhoto
func (r *repoUsers) RemoveUserPhoto(userID int64) error {
	r.mx.Lock()
	defer r.mx.Unlock()
	u := dto.Users{}
	r.db.Table(u.TableName()).Where("ID=?", userID).Updates(map[string]interface{}{
		"Photo": []byte(""),
	})
	return r.db.Delete(dto.UserPhotos{}, "UserID = ?", userID).Error
}
