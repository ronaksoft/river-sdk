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
	r.networkCtrl.Connect(true)
}

func (r *River) StopNetwork() {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("StopNetwork", time.Now().Sub(startTime))
	}()
	r.networkCtrl.Disconnect()
}
