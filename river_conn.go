package riversdk

import (
	"encoding/json"

	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
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
	} else {
		logs.Error("River::loadSystemConfig()", zap.Error(err))
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
		logs.Error("River::saveDeviceToken()-> Json Marshal()", zap.Error(err))
		return
	}
	err = repo.Ctx().System.SaveString(domain.ColumnDeviceToken, string(val))
	if err != nil {
		logs.Error("River::saveDeviceToken()-> SaveString()", zap.Error(err))
		return
	}
}

// saveConfig save to DB
func (v *RiverConnection) saveConfig() {
	logs.Debug("RiverConnection::Save()")
	if bytes, err := v.MarshalJSON(); err != nil {
		logs.Error("RiverConnection::Save()->MarshalJSON()", zap.Error(err))
	} else if err := repo.Ctx().System.SaveString(domain.ColumnConnectionInfo, string(bytes)); err != nil {
		logs.Error("RiverConnection::Save()->SaveString()", zap.Error(err))
	}
}

// loadConfig load from DB
func (v *RiverConnection) loadConfig() error {
	logs.Debug("RiverConnection::Load()")
	if kv, err := repo.Ctx().System.LoadString(domain.ColumnConnectionInfo); err != nil {
		logs.Error("RiverConnection::Load()->LoadString()", zap.Error(err))
		return err
	} else if err := v.UnmarshalJSON([]byte(kv)); err != nil {
		logs.Error("RiverConnection::Load()->UnmarshalJSON()", zap.Error(err))
		return err
	}
	return nil
}

// Save RiverConfig interface func
func (v *RiverConnection) Save() { v.saveConfig() }

// ChangeAuthID RiverConfig interface func
func (v *RiverConnection) ChangeAuthID(authID int64) { v.AuthID = authID }

// ChangeAuthKey RiverConfig interface func
func (v *RiverConnection) ChangeAuthKey(authKey [256]byte) { v.AuthKey = authKey }

// ChangeUserID RiverConfig interface func
func (v *RiverConnection) ChangeUserID(userID int64) { v.UserID = userID }

// ChangeUsername RiverConfig interface func
func (v *RiverConnection) ChangeUsername(username string) { v.Username = username }

// ChangePhone RiverConfig interface func
func (v *RiverConnection) ChangePhone(phone string) { v.Phone = phone }

// ChangeFirstName RiverConfig interface func
func (v *RiverConnection) ChangeFirstName(firstName string) { v.FirstName = firstName }

// ChangeLastName RiverConfig interface func
func (v *RiverConnection) ChangeLastName(lastName string) { v.LastName = lastName }

// ChangeBio RiverConfig interface func
func (v *RiverConnection) ChangeBio(bio string) { v.Bio = bio }

// Load RiverConfig interface func
func (v *RiverConnection) Load() error { return v.loadConfig() }

// PickupAuthID RiverConfig interface func
func (v *RiverConnection) PickupAuthID() int64 { return v.AuthID }

// PickupAuthKey RiverConfig interface func
func (v *RiverConnection) PickupAuthKey() [256]byte { return v.AuthKey }

// PickupUserID RiverConfig interface func
func (v *RiverConnection) PickupUserID() int64 { return v.UserID }

// PickupUsername RiverConfig interface func
func (v *RiverConnection) PickupUsername() string { return v.Username }

// PickupPhone RiverConfig interface func
func (v *RiverConnection) PickupPhone() string { return v.Phone }

// PickupFirstName RiverConfig interface func
func (v *RiverConnection) PickupFirstName() string { return v.FirstName }

// PickupLastName RiverConfig interface func
func (v *RiverConnection) PickupLastName() string { return v.LastName }

// PickupBio RiverConfig interface func
func (v *RiverConnection) PickupBio() string { return v.Bio }
