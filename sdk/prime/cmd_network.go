package riversdk

import (
    "github.com/ronaksoft/river-sdk/internal/domain"
    "go.uber.org/zap"
)

/*
   Creation Time: 2019 - Jun - 25
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

func (r *River) StartNetwork(country string) {
    if country != "" {
        r.networkCtrl.UpdateEndpoint(country)
    }
    r.networkCtrl.Connect()
}

func (r *River) StopNetwork() {
    r.networkCtrl.Disconnect()
}

// NetworkChange accepts possible values: cellular (2), wifi (1), none (0)
func (r *River) NetworkChange(connection int) {
    logger.Debug("NetworkChange called", zap.Int("C", connection))
    switch connection {
    case domain.ConnectionNone:
        r.networkCtrl.Disconnect()
    default:
        r.networkCtrl.Reconnect()
    }
}

func (r *River) GetNetworkStatus() int32 {
    return int32(r.networkCtrl.GetStatus())
}
