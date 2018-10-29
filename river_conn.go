package riversdk

import (
	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/log"
	"git.ronaksoftware.com/ronak/riversdk/repo"
	"go.uber.org/zap"
)

// easyjson:json
// publicKey
type PublicKey struct {
	N           string
	FingerPrint int64
	E           uint32
}

// easyjson:json
// DHGroup
type DHGroup struct {
	Prime       string
	Gen         int32
	FingerPrint int64
}

// easyjson:json
// ServerKeys
type ServerKeys struct {
	PublicKeys []PublicKey
	DHGroups   []DHGroup
}

// getPublicKey
func (v *ServerKeys) GetPublicKey(keyFP int64) (PublicKey, error) {
	for _, pk := range v.PublicKeys {
		if pk.FingerPrint == keyFP {
			return pk, nil
		}
	}
	return PublicKey{}, domain.ErrNotFound
}

// getDhGroup
func (v *ServerKeys) GetDhGroup(keyFP int64) (DHGroup, error) {
	for _, dh := range v.DHGroups {
		if dh.FingerPrint == keyFP {
			return dh, nil
		}
	}
	return DHGroup{}, domain.ErrNotFound
}

// easyjson:json
// RiverConnection
type RiverConnection struct {
	AuthID    int64
	AuthKey   [256]byte
	UserID    int64
	Username  string
	Phone     string
	FirstName string
	LastName  string
}

// Get
func (r *River) loadSystemConfig() {
	r.ConnInfo = new(RiverConnection)
	if err := r.ConnInfo.loadConfig(); err != nil {
		r.ConnInfo.saveConfig()
	}
}

// Clear
func (r *River) clearSystemConfig() {
	r.ConnInfo = &RiverConnection{}
}

// Save
func (v *RiverConnection) saveConfig() {
	log.LOG.Debug("RiverConnection::Save()")
	if bytes, err := v.MarshalJSON(); err != nil {
		log.LOG.Info("RiverConnection::Save()->MarshalJSON()",
			zap.String("Error", err.Error()),
		)
	} else if err := repo.Ctx().System.SaveString(domain.CN_CONN_INFO, string(bytes)); err != nil {
		log.LOG.Info("RiverConnection::Save()->SaveString()",
			zap.String("Error", err.Error()),
		)
	}
}

// Load
func (v *RiverConnection) loadConfig() error {
	log.LOG.Debug("RiverConnection::Load()")
	if kv, err := repo.Ctx().System.LoadString(domain.CN_CONN_INFO); err != nil {
		log.LOG.Info("RiverConnection::Load()->LoadString()",
			zap.String("Error", err.Error()),
		)
		return err
	} else if err := v.UnmarshalJSON([]byte(kv)); err != nil {
		log.LOG.Info("RiverConnection::Load()->UnmarshalJSON()",
			zap.String("Error", err.Error()),
		)
		return err
	}
	return nil
}

func (v *RiverConnection) Save()                            { v.saveConfig() }
func (v *RiverConnection) ChangeAuthID(authID int64)        { v.AuthID = authID }
func (v *RiverConnection) ChangeAuthKey(authKey [256]byte)  { v.AuthKey = authKey }
func (v *RiverConnection) ChangeUserID(userID int64)        { v.UserID = userID }
func (v *RiverConnection) ChangeUsername(username string)   { v.Username = username }
func (v *RiverConnection) ChangePhone(phone string)         { v.Phone = phone }
func (v *RiverConnection) ChangeFirstName(firstName string) { v.FirstName = firstName }
func (v *RiverConnection) ChangeLastName(lastName string)   { v.LastName = lastName }
func (v *RiverConnection) Load() error                      { return v.loadConfig() }
func (v *RiverConnection) PickupAuthID() int64              { return v.AuthID }
func (v *RiverConnection) PickupAuthKey() [256]byte         { return v.AuthKey }
func (v *RiverConnection) PickupUserID() int64              { return v.UserID }
func (v *RiverConnection) PickupUsername() string           { return v.Username }
func (v *RiverConnection) PickupPhone() string              { return v.Phone }
func (v *RiverConnection) PickupFirstName() string          { return v.FirstName }
func (v *RiverConnection) PickupLastName() string           { return v.LastName }
