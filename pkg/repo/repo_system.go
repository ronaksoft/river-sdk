package repo

import (
	"encoding/json"
	"git.ronaksoftware.com/ronak/riversdk/msg/ext"
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

	logs.Debug("System::SaveString()",
		zap.String(keyName, keyValue),
	)
	return r.db.Save(s).Error
}

// SaveSalt
func (r *repoSystem) SaveSalt(serverSlats *msg.SystemSalts) error {
	r.mx.Lock()
	defer r.mx.Unlock()
	logs.Info("repoSystem()::SaveSalt", zap.Int("salt length", len(serverSlats.Salts)))
	dtoSalts := make([]dto.ServerSalt,0)
	for _, salt := range serverSlats.Salts {
		s := dto.ServerSalt{}
		s.Timestamp = salt.Timestamp
		s.Salt = salt.Value
		dtoSalts = append(dtoSalts, s)
	}
	s := new(dto.System)
	keyName := "Salt"
	r.db.Where("KeyName = ?", keyName).First(s)
	//if len(s.Salt) > 0 {
	//	r.db.Delete(s)
	//}
	logs.Info("repoSystem()::Save", zap.Any("salts", dtoSalts))
	s.KeyName = "Salt"
	b, _ := json.Marshal(dtoSalts)
	s.Salt = string(b)

	return r.db.Save(s).Error
}

// GetSalt
func (r *repoSystem) GetSalt() (salt msg.SystemSalts ,err error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	salt = msg.SystemSalts{}

	logs.Info("repoSystem()::GetSalt")
	s := dto.System{}
	keyName := "Salt"
	err = r.db.Where("KeyName = ?", keyName).First(&s).Error
	if err != nil {
		return salt, err
	}
	var dtoSalts []dto.ServerSalt
	err = json.Unmarshal([]byte(s.Salt), &dtoSalts)
	if err != nil {
		logs.Warn("repoSalt()::GetSalt", zap.String("error", err.Error()))
		return
	}
	for _, dtoSalt := range dtoSalts {
		sysSalt := new(msg.Salt)
		sysSalt.Value = dtoSalt.Salt
		sysSalt.Timestamp = dtoSalt.Timestamp
		salt.Salts = append(salt.Salts, sysSalt)
	}
	logs.Debug("repoSalt()::GetSalt", zap.Any("salt", len(salt.Salts)))
	return
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
