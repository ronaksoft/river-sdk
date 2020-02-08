// +build linux

package networkCtrl

import (
	"net"
)

/*
   Creation Time: 2020 - Feb - 08
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

func (ctrl *Controller) ignoreSIGPIPE(c net.Conn) {}
