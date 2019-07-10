package repo

import (
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo/dto"
	"go.uber.org/zap"
)

type repoSystem struct {
	*repository
}

// LoadInt
func (r *repoSystem) LoadInt(keyName string) (keyValue int, err error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	keyValue = 0
	row := new(dto.System)
	err = r.db.Where("KeyName = ?", keyName).First(row).Error

	keyValue = int(row.IntValue)
	return
}

// LoadString
func (r *repoSystem) LoadString(keyName string) (keyValue string, err error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	row := new(dto.System)
	err = r.db.Where("KeyName = ?", keyName).First(row).Error

	keyValue = row.StrValue
	return
}

// SaveInt
func (r *repoSystem) SaveInt(keyName string, keyValue int32) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	s := dto.System{}

	err := r.db.Where("KeyName = ?", keyName).First(&s).Error
	if err != nil {
		logs.Error("System::SaveInt()-> fetch system entity",
			zap.String("Key", keyName),
			zap.Int32("Value", keyValue),
			zap.Error(err),
		)
	}

	s.KeyName = keyName
	s.StrValue = ""
	s.IntValue = keyValue

	logs.Debug("System::SaveInt()",
		zap.Int32(keyName, keyValue),
	)

	return r.db.Save(s).Error
}

// SaveString
func (r *repoSystem) SaveString(keyName string, keyValue string) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	s := dto.System{}

	err := r.db.Where("KeyName = ?", keyName).First(&s).Error
	if err != nil {
		logs.Error("System::SaveString()-> fetch system entity", zap.Error(err))
	}

	s.KeyName = keyName
	s.StrValue = keyValue
	s.IntValue = 0

	return r.db.Save(s).Error
}

// SaveSalt
func (r *repoSystem) RemoveSalt() error {
	r.mx.Lock()
	defer r.mx.Unlock()

	s := new(dto.System)

	keyName := "Salt"
	err := r.db.Where("KeyName = ?", keyName).First(s).Error
	if err == nil {
		return r.db.Delete(s).Error
	} else {
		return err
	}
}
