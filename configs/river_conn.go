package configs

import (
	"sync"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/log"
	"git.ronaksoftware.com/ronak/riversdk/repo"
	"go.uber.org/zap"
)

var (
	mx        sync.Mutex
	riverConf *RiverConnection
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
func Get() *RiverConnection {
	if riverConf == nil {
		mx.Lock()
		defer mx.Unlock()
		if riverConf == nil {
			riverConf = new(RiverConnection)
			if err := riverConf.Load(); err != nil {
				riverConf.Save()
			}
		}
	}
	return riverConf
}

// Clear
func Clear() {
	riverConf = &RiverConnection{}
}

// Save
func (v *RiverConnection) Save() {
	log.LOG.Debug("RiverConnection::Save()")
	if bytes, err := v.MarshalJSON(); err != nil {
		log.LOG.Info(err.Error(),
			zap.String(domain.LK_FUNC_NAME, "RiverConnection::Save()"),
		)
	} else if err := repo.Ctx().System.SaveString(domain.CN_CONN_INFO, string(bytes)); err != nil {
		log.LOG.Info(err.Error(),
			zap.String(domain.LK_FUNC_NAME, "RiverConnection::Save()"),
		)
	}
}

// Load
func (v *RiverConnection) Load() error {
	log.LOG.Debug("RiverConnection::Load()")
	if kv, err := repo.Ctx().System.LoadString(domain.CN_CONN_INFO); err != nil {
		log.LOG.Info(err.Error())
		return err
	} else if err := v.UnmarshalJSON([]byte(kv)); err != nil {
		log.LOG.Error(err.Error(),
			zap.String(domain.LK_FUNC_NAME, "RiverConnection::Load()"),
		)
		return err
	}
	return nil
}
