package riversdk

/*
   Creation Time: 2019 - Jun - 25
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

func (r *River) StartNetwork() {
	r.networkCtrl.Connect()
}

func (r *River) StopNetwork() {
	r.networkCtrl.Disconnect()
}

func (r *River) GetNetworkStatus() int32 {
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
