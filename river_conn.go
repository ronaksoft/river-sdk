package riversdk

import (
	"encoding/json"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/logs"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/repo"
	"go.uber.org/zap"
)

// PublicKey ...
// easyjson:json
type PublicKey struct {
	N           string
	FingerPrint int64
	E           uint32
}

// DHGroup ...
// easyjson:json
type DHGroup struct {
	Prime       string
	Gen         int32
	FingerPrint int64
}

// ServerKeys ...
// easyjson:json
type ServerKeys struct {
	PublicKeys []PublicKey
	DHGroups   []DHGroup
}

// GetPublicKey ...
func (v *ServerKeys) GetPublicKey(keyFP int64) (PublicKey, error) {
	for _, pk := range v.PublicKeys {
		if pk.FingerPrint == keyFP {
			return pk, nil
		}
	}
	return PublicKey{}, domain.ErrNotFound
}

// GetDhGroup ...
func (v *ServerKeys) GetDhGroup(keyFP int64) (DHGroup, error) {
	for _, dh := range v.DHGroups {
		if dh.FingerPrint == keyFP {
			return dh, nil
		}
	}
	return DHGroup{}, domain.ErrNotFound
}

// RiverConnection connection info
// easyjson:json
type RiverConnection struct {
	AuthID    int64
	AuthKey   [256]byte
	UserID    int64
	Username  string
	Phone     string
	FirstName string
	LastName  string
	Bio       string
}

// loadSystemConfig get system config from local DB
func (r *River) loadSystemConfig() {
	r.ConnInfo = new(RiverConnection)
	if err := r.ConnInfo.loadConfig(); err != nil {
		r.ConnInfo.saveConfig()
	}
}

// clearSystemConfig reset config
func (r *River) clearSystemConfig() {
	r.ConnInfo.FirstName = ""
	r.ConnInfo.LastName = ""
	r.ConnInfo.Phone = ""
	r.ConnInfo.UserID = 0
	r.ConnInfo.Username = ""
	r.ConnInfo.Save()
	r.DeviceToken = new(msg.AccountRegisterDevice)
	r.saveDeviceToken()
}

// saveDeviceToken save DeviceToken to DB
func (r *River) saveDeviceToken() {
	val, err := json.Marshal(r.DeviceToken)
	if err != nil {
		logs.Debug("River::saveDeviceToken()-> Json Marshal()",
			zap.String("Error", err.Error()),
		)
		return
	}
	err = repo.Ctx().System.SaveString(domain.ColumnDeviceToken, string(val))
	if err != nil {
		logs.Debug("River::saveDeviceToken()-> SaveString()",
			zap.String("Error", err.Error()),
		)
		return
	}
}

// saveConfig save to DB
func (v *RiverConnection) saveConfig() {
	logs.Debug("RiverConnection::Save()")
	if bytes, err := v.MarshalJSON(); err != nil {
		logs.Info("RiverConnection::Save()->MarshalJSON()",
			zap.String("Error", err.Error()),
		)
	} else if err := repo.Ctx().System.SaveString(domain.ColumnConnectionInfo, string(bytes)); err != nil {
		logs.Info("RiverConnection::Save()->SaveString()",
			zap.String("Error", err.Error()),
		)
	}
}

// loadConfig load from DB
func (v *RiverConnection) loadConfig() error {
	logs.Debug("RiverConnection::Load()")
	if kv, err := repo.Ctx().System.LoadString(domain.ColumnConnectionInfo); err != nil {
		logs.Info("RiverConnection::Load()->LoadString()",
			zap.String("Error", err.Error()),
		)
		return err
	} else if err := v.UnmarshalJSON([]byte(kv)); err != nil {
		logs.Info("RiverConnection::Load()->UnmarshalJSON()",
			zap.String("Error", err.Error()),
		)
		return err
	}
	return nil
}

// Save RiverConfiger interface func
func (v *RiverConnection) Save() { v.saveConfig() }

// ChangeAuthID RiverConfiger interface func
func (v *RiverConnection) ChangeAuthID(authID int64) { v.AuthID = authID }

// ChangeAuthKey RiverConfiger interface func
func (v *RiverConnection) ChangeAuthKey(authKey [256]byte) { v.AuthKey = authKey }

// ChangeUserID RiverConfiger interface func
func (v *RiverConnection) ChangeUserID(userID int64) { v.UserID = userID }

// ChangeUsername RiverConfiger interface func
func (v *RiverConnection) ChangeUsername(username string) { v.Username = username }

// ChangePhone RiverConfiger interface func
func (v *RiverConnection) ChangePhone(phone string) { v.Phone = phone }

// ChangeFirstName RiverConfiger interface func
func (v *RiverConnection) ChangeFirstName(firstName string) { v.FirstName = firstName }

// ChangeLastName RiverConfiger interface func
func (v *RiverConnection) ChangeLastName(lastName string) { v.LastName = lastName }

// ChangeBio RiverConfiger interface func
func (v *RiverConnection) ChangeBio(bio string) { v.Bio = bio }

// Load RiverConfiger interface func
func (v *RiverConnection) Load() error { return v.loadConfig() }

// PickupAuthID RiverConfiger interface func
func (v *RiverConnection) PickupAuthID() int64 { return v.AuthID }

// PickupAuthKey RiverConfiger interface func
func (v *RiverConnection) PickupAuthKey() [256]byte { return v.AuthKey }

// PickupUserID RiverConfiger interface func
func (v *RiverConnection) PickupUserID() int64 { return v.UserID }

// PickupUsername RiverConfiger interface func
func (v *RiverConnection) PickupUsername() string { return v.Username }

// PickupPhone RiverConfiger interface func
func (v *RiverConnection) PickupPhone() string { return v.Phone }

// PickupFirstName RiverConfiger interface func
func (v *RiverConnection) PickupFirstName() string { return v.FirstName }

// PickupLastName RiverConfiger interface func
func (v *RiverConnection) PickupLastName() string { return v.LastName }

// PickupBio RiverConfiger interface func
func (v *RiverConnection) PickupBio() string { return v.Bio }
