package riversdk

import (
	mon "git.ronaksoftware.com/ronak/riversdk/pkg/monitoring"
	"time"
)

/*
   Creation Time: 2019 - Jun - 25
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

func (r *River) StartNetwork() {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("StartNetwork", time.Now().Sub(startTime))
	}()
	r.networkCtrl.Start()
	r.networkCtrl.Connect()
}

func (r *River) StopNetwork() {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("StopNetwork", time.Now().Sub(startTime))
	}()
	r.networkCtrl.Stop()
}

func (r *River) GetNetworkStatus() int32 {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("GetNetworkStatus", time.Now().Sub(startTime))
	}()
	return int32(r.networkCtrl.GetQuality())
}

// UnderlyingNetworkChange
func (r *River) UnderlyingNetworkChange(connected bool) {
	if connected {
		r.networkCtrl.Reconnect()
	} else {
		r.networkCtrl.Disconnect()
	}
}

// UnderlyingNetworkSpeedChange
func (r *River) UnderlyingNetworkSpeedChange(fast bool) {
	// TODO:: update timeout values
}
