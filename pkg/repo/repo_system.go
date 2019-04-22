package repo

import (
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo/dto"
	"go.uber.org/zap"
)

// System repoSystem interface
type System interface {
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
	keyValue = 0
	row := new(dto.System)
	err = r.db.Where("KeyName = ?", keyName).First(row).Error

	if row == nil {
		err = domain.ErrDoesNotExists
	}
	keyValue = int(row.IntValue)

	logs.Debug("System::LoadInt()",
		zap.Int(keyName, keyValue),
	)
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

	logs.Debug("System::LoadString()",
		zap.String(keyName, keyValue),
	)

	return
}

// SaveInt
func (r *repoSystem) SaveInt(keyName string, keyValue int32) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	logs.Info("repoSystem()", zap.Int32(keyName, keyValue))

	s := dto.System{}

	err := r.db.Where("KeyName = ?", keyName).First(&s).Error
	if err != nil {
		logs.Error("System::SaveInt()-> fetch system entity", zap.Error(err))
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

	logs.Debug("System::SaveString()",
		zap.String(keyName, keyValue),
	)
	return r.db.Save(s).Error
}
