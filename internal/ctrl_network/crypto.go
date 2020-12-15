package networkCtrl

import (
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"go.uber.org/zap"
)

/*
   Creation Time: 2020 - Dec - 15
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

// PublicKey ...
type PublicKey struct {
	N           string
	FingerPrint int64
	E           uint32
}

// DHGroup ...
type DHGroup struct {
	Prime       string
	Gen         int32
	FingerPrint int64
}

// ServerKeys ...
type ServerKeys struct {
	PublicKeys []PublicKey
	DHGroups   []DHGroup
}

// GetPublicKey ...
func (v *ServerKeys) GetPublicKey(keyFP int64) (PublicKey, error) {
	logs.Info("Public Keys loaded",
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
