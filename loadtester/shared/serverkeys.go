package shared

import (
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
)

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
