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
	GetAccessHash(userID int64) (uint64, error)
	UpdateAccessHash(accessHash int64, peerID int64, peerType int32) error
	GetUser(userID int64) *msg.User
	GetAnyUsers(userIDs []int64) []*msg.User
}

type repoUsers struct {
	*repository
}

// Save
func (r *repoUsers) SaveUser(user *msg.User) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if user == nil {
		log.LOG.Debug("RepoRepoUsers::SaveUser()",
			zap.String("User", "user is null"),
		)
		return domain.ErrNotFound
	}
	log.LOG.Info("RepoRepoUsers::SaveUser()",
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
		log.LOG.Debug("RepoRepoUsers::SaveContactUser()",
			zap.String("User", "user is null"),
		)
		return domain.ErrNotFound
	}

	log.LOG.Info("RepoRepoUsers::SaveContactUser()",
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
	return r.db.Table(u.TableName()).Where("ID=?", u.ID).Update(u).Error
}

// GetManyUsers
func (r *repoUsers) GetManyUsers(userIDs []int64) []*msg.User {
	r.mx.Lock()
	defer r.mx.Unlock()

	log.LOG.Debug("RepoUsers::GetManyUsers()",
		zap.Int64s("UserIDs", userIDs),
	)

	pbUsers := make([]*msg.User, 0, len(userIDs))
	users := make([]dto.Users, 0, len(userIDs))

	err := r.db.Where("IsContact = 0 AND ID in (?)", userIDs).Find(&users).Error
	if err != nil {
		log.LOG.Debug("RepoUsers::GetManyUsers()-> fetch user entity",
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

	log.LOG.Debug("RepoUsers::GetManyContactUsers()",
		zap.Int64s("UserIDs", userIDs),
	)
	pbUsers := make([]*msg.ContactUser, 0, len(userIDs))
	users := make([]dto.Users, 0, len(userIDs))

	err := r.db.Where("IsContact = 1 AND ID in (?)", userIDs).Find(&users).Error
	if err != nil {
		log.LOG.Debug("RepoUsers::GetManyContactUsers()-> fetch user entities",
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

	log.LOG.Debug("RepoUsers::GetContacts()")

	users := make([]dto.Users, 0)

	err := r.db.Where("AccessHash <> 0").Find(&users).Error
	if err != nil {
		log.LOG.Debug("RepoUsers::GetContacts()-> fetch user entities",
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

	log.LOG.Debug("RepoUsers::GetUser()",
		zap.Int64("UserID", userID),
	)

	pbUser := new(msg.User)
	user := new(dto.Users)

	err := r.db.Find(user, userID).Error
	if err != nil {
		log.LOG.Debug("RepoUsers::GetUser()-> fetch user entity",
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

	log.LOG.Debug("RepoUsers::GetAnyUsers()",
		zap.Int64s("UserIDs", userIDs),
	)

	pbUsers := make([]*msg.User, 0, len(userIDs))
	users := make([]dto.Users, 0, len(userIDs))

	err := r.db.Where("ID in (?)", userIDs).Find(&users).Error
	if err != nil {
		log.LOG.Debug("RepoUsers::GetAnyUsers()-> fetch user entity",
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
