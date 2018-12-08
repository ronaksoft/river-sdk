package repo

import (
	"fmt"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/log"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/repo/dto"
	"go.uber.org/zap"
)

type RepoUsers interface {
	SaveUser(user *msg.User) error
	SaveContactUser(user *msg.ContactUser) error
	GetManyUsers(userIDs []int64) []*msg.User
	GetManyContactUsers(userIDs []int64) []*msg.ContactUser
	GetContacts() ([]*msg.ContactUser, []*msg.PhoneContact)
	SearchContacts(searchPhrase string) ([]*msg.ContactUser, []*msg.PhoneContact)
	GetAccessHash(userID int64) (uint64, error)
	UpdateAccessHash(accessHash int64, peerID int64, peerType int32) error
	GetUser(userID int64) *msg.User
	GetAnyUsers(userIDs []int64) []*msg.User
	SaveMany(users []*msg.User) error
	UpdateContactinfo(userID int64, firstName, lastName string) error
}

type repoUsers struct {
	*repository
}

// Save
func (r *repoUsers) SaveUser(user *msg.User) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if user == nil {
		log.LOG_Debug("RepoRepoUsers::SaveUser()",
			zap.String("User", "user is null"),
		)
		return domain.ErrNotFound
	}
	log.LOG_Info("RepoRepoUsers::SaveUser()",
		zap.String("Name", fmt.Sprintf("%s %s", user.FirstName, user.LastName)),
		zap.Int64("UserID", user.ID),
	)
	u := new(dto.Users)
	u.MapFromUser(user)

	eu := new(dto.Users)
	r.db.Find(eu, u.ID)
	if eu.ID == 0 {
		return r.db.Create(u).Error
	}
	return nil
	// if user not exist just add it no need to update existing user
	//return r.db.Table(u.TableName()).Where("ID=?", u.ID).Update(u).Error
}

// SaveContactUser
func (r *repoUsers) SaveContactUser(user *msg.ContactUser) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if user == nil {
		log.LOG_Debug("RepoRepoUsers::SaveContactUser()",
			zap.String("User", "user is null"),
		)
		return domain.ErrNotFound
	}

	log.LOG_Info("RepoRepoUsers::SaveContactUser()",
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
	}).Error
}

// GetManyUsers
func (r *repoUsers) GetManyUsers(userIDs []int64) []*msg.User {
	r.mx.Lock()
	defer r.mx.Unlock()

	log.LOG_Debug("RepoUsers::GetManyUsers()",
		zap.Int64s("UserIDs", userIDs),
	)

	pbUsers := make([]*msg.User, 0, len(userIDs))
	users := make([]dto.Users, 0, len(userIDs))

	err := r.db.Where("IsContact = 0 AND ID in (?)", userIDs).Find(&users).Error
	if err != nil {
		log.LOG_Debug("RepoUsers::GetManyUsers()-> fetch user entity",
			zap.String("Error", err.Error()),
		)
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

	log.LOG_Debug("RepoUsers::GetManyContactUsers()",
		zap.Int64s("UserIDs", userIDs),
	)
	pbUsers := make([]*msg.ContactUser, 0, len(userIDs))
	users := make([]dto.Users, 0, len(userIDs))

	err := r.db.Where("IsContact = 1 AND ID in (?)", userIDs).Find(&users).Error
	if err != nil {
		log.LOG_Debug("RepoUsers::GetManyContactUsers()-> fetch user entities",
			zap.String("Error", err.Error()),
		)
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

	log.LOG_Debug("RepoUsers::GetContacts()")

	users := make([]dto.Users, 0)

	err := r.db.Where("AccessHash <> 0").Find(&users).Error
	if err != nil {
		log.LOG_Debug("RepoUsers::GetContacts()-> fetch user entities",
			zap.String("Error", err.Error()),
		)
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

	log.LOG_Debug("RepoUsers::SearchContacts()")

	p := "%" + searchPhrase + "%"
	users := make([]dto.Users, 0)
	err := r.db.Where("AccessHash <> 0 AND (FirstName LIKE ? OR LastName LIKE ? OR Phone LIKE ? OR Username LIKE ?)", p, p, p, p).Find(&users).Error
	if err != nil {
		log.LOG_Debug("RepoUsers::SearchContacts()-> fetch user entities",
			zap.String("Error", err.Error()),
		)
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

	log.LOG_Debug("RepoUsers::GetUser()",
		zap.Int64("UserID", userID),
	)

	pbUser := new(msg.User)
	user := new(dto.Users)

	err := r.db.Find(user, userID).Error
	if err != nil {
		log.LOG_Debug("RepoUsers::GetUser()-> fetch user entity",
			zap.String("Error", err.Error()),
		)
		return nil //, err
	}

	user.MapToUser(pbUser)

	return pbUser
}

// GetAnyUsers
func (r *repoUsers) GetAnyUsers(userIDs []int64) []*msg.User {
	r.mx.Lock()
	defer r.mx.Unlock()

	log.LOG_Debug("RepoUsers::GetAnyUsers()",
		zap.Int64s("UserIDs", userIDs),
	)

	pbUsers := make([]*msg.User, 0, len(userIDs))
	users := make([]dto.Users, 0, len(userIDs))

	err := r.db.Where("ID in (?)", userIDs).Find(&users).Error
	if err != nil {
		log.LOG_Debug("RepoUsers::GetAnyUsers()-> fetch user entity",
			zap.String("Error", err.Error()),
		)
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
		log.LOG_Debug("RepoUsers::SaveMany()-> fetch groups entity",
			zap.String("Error", err.Error()),
		)
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
			log.LOG_Debug("RepoUsers::SaveMany()-> save group entity",
				zap.Int64("ID", v.ID),
				zap.String("FirstName", v.FirstName),
				zap.String("Lastname", v.LastName),
				zap.String("UserName", v.Username),
				zap.String("Error", err.Error()),
			)
			break
		}
	}
	return err
}

func (r *repoUsers) UpdateContactinfo(userID int64, firstName, lastName string) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	log.LOG_Debug("RepoUsers::UpdateContactinfo()",
		zap.Int64("UserID", userID),
	)
	dtoUser := new(dto.Users)
	err := r.db.Table(dtoUser.TableName()).Where("ID=?", userID).Updates(map[string]interface{}{
		"FirstName": firstName,
		"LasttName": lastName,
	}).Error

	return err
}
