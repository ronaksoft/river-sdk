package repo

import (
	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/repo/dto"
)

type RepoSystem interface {
	LoadInt(keyName string) (keyValue int, err error)
	LoadString(keyName string) (keyValue string, err error)
	SaveInt(keyName string, keyValue int32) error
	SaveString(keyName string, keyValue string) error
}

type repoSystem struct {
	*repository
}

// LoadInt
func (r *repoSystem) LoadInt(keyName string) (keyValue int, err error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	row := new(dto.System)
	err = r.db.Where("KeyName = ?", keyName).First(row).Error

	if row == nil {
		err = domain.ErrDoesNotExists
	}
	keyValue = int(row.IntValue)
	return
}

// LoadString
func (r *repoSystem) LoadString(keyName string) (keyValue string, err error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	row := new(dto.System)
	err = r.db.Where("KeyName = ?", keyName).First(row).Error

	if row == nil {
		err = domain.ErrDoesNotExists
	}
	keyValue = row.StrValue
	return
}

// SaveInt
func (r *repoSystem) SaveInt(keyName string, keyValue int32) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	s := dto.System{}

	r.db.Where("KeyName = ?", keyName).First(&s)

	s.KeyName = keyName
	s.StrValue = ""
	s.IntValue = keyValue

	return r.db.Save(s).Error
}

// SaveString
func (r *repoSystem) SaveString(keyName string, keyValue string) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	s := dto.System{}

	r.db.Where("KeyName = ?", keyName).First(&s)

	s.KeyName = keyName
	s.StrValue = keyValue
	s.IntValue = 0

	return r.db.Save(s).Error
}
