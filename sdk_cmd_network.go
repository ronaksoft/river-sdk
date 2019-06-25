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
	r.networkCtrl.Connect(true)
}

func (r *River) StopNetwork() {
	r.networkCtrl.Disconnect()
}
