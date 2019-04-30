package riversdk

import (
	"encoding/json"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
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
	AuthKey   []byte
	UserID    int64
	Username  string
	Phone     string
	FirstName string
	LastName  string
	Bio       string
	Delegate  ConnInfoDelegate
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

// Get deviceToken
func (r *River) loadDeviceToken() {
	r.DeviceToken = new(msg.AccountRegisterDevice)
	str, err := repo.Ctx().System.LoadString(domain.ColumnDeviceToken)
	if err != nil {
		logs.Error("River::loadDeviceToken() failed to fetch DeviceToken",
			zap.String("Error", err.Error()),
		)
		return
	}
	err = json.Unmarshal([]byte(str), r.DeviceToken)
	if err != nil {
		logs.Error("River::loadDeviceToken() failed to unmarshal DeviceToken",
			zap.String("Error", err.Error()),
		)
	}
}

// Save RiverConfig interface func
func (v *RiverConnection) Save() {
	b, _ := v.MarshalJSON()
	v.Delegate.SaveConnInfo(b)
//	v.saveConfig()
}
