package riversdk

import (
	"encoding/json"

	msg "git.ronaksoftware.com/river/msg/chat"
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
	logs.Info("River::CreateAuthKey() Check GetPublicKeys",
		zap.Any("Public Keys", v.PublicKeys),
		zap.Int64("keyFP", keyFP),
	)

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
	Delegate  ConnInfoDelegate `json:"-"`
	Version   int
}

// saveDeviceToken save DeviceToken to DB
func (r *River) saveDeviceToken() {
	val, err := json.Marshal(r.DeviceToken)
	if err != nil {
		logs.Error("River::saveDeviceToken()-> Json Marshal()", zap.Error(err))
		return
	}
	err = repo.System.SaveString(domain.SkDeviceToken, string(val))
	if err != nil {
		logs.Error("River::saveDeviceToken()-> SaveString()", zap.Error(err))
		return
	}
}

// Get deviceToken
func (r *River) loadDeviceToken() {
	r.DeviceToken = new(msg.AccountRegisterDevice)
	str, err := repo.System.LoadString(domain.SkDeviceToken)
	if err != nil {
		logs.Warn("River::loadDeviceToken() failed to fetch DeviceToken",
			zap.String("Error", err.Error()),
		)
		return
	}
	err = json.Unmarshal([]byte(str), r.DeviceToken)
	if err != nil {
		logs.Warn("River::loadDeviceToken() failed to unmarshal DeviceToken",
			zap.String("Error", err.Error()),
		)
	}
}

// save RiverConfig interface func
func (v *RiverConnection) Save() {
	logs.Debug("ConnInfo saved.")
	b, _ := v.MarshalJSON()
	v.Delegate.SaveConnInfo(b)
}

// ChangeAuthID RiverConfig interface func
func (v *RiverConnection) ChangeAuthID(authID int64) { v.AuthID = authID }

// ChangeAuthKey RiverConfig interface func
func (v *RiverConnection) ChangeAuthKey(authKey []byte) {
	copy(v.AuthKey[:], authKey[:256])
}

func (v *RiverConnection) GetAuthKey() []byte {
	var bytes = make([]byte, 256)
	copy(bytes, v.AuthKey[0:256])
	return bytes
}

// ChangeUserID RiverConfig interface func
func (v *RiverConnection) ChangeUserID(userID int64) { v.UserID = userID }

// ChangeUsername RiverConfig interface func
func (v *RiverConnection) ChangeUsername(username string) { v.Username = username }

// ChangePhone RiverConfig interface func
func (v *RiverConnection) ChangePhone(phone string) {
	v.Phone = phone
	domain.ClientPhone = phone
}

// ChangeFirstName RiverConfig interface func
func (v *RiverConnection) ChangeFirstName(firstName string) { v.FirstName = firstName }

// ChangeLastName RiverConfig interface func
func (v *RiverConnection) ChangeLastName(lastName string) { v.LastName = lastName }

// ChangeBio RiverConfig interface func
func (v *RiverConnection) ChangeBio(bio string) { v.Bio = bio }

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

func (v *RiverConnection) GetKey(key string) string {
	return v.Delegate.Get(key)
}

func (v *RiverConnection) SetKey(key, value string) {
	v.Delegate.Set(key, value)
}