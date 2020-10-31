// +build darwin

package networkCtrl

import (
	"git.ronaksoft.com/river/sdk/internal/logs"
	"go.uber.org/zap"
	"net"
	"syscall"
)

/*
   Creation Time: 2020 - Feb - 08
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

func (ctrl *Controller) ignoreSIGPIPE(c net.Conn) {
	if c == nil {
		return
	}
	s, ok := c.(syscall.Conn)
	if !ok {
		return
	}
	r, e := s.SyscallConn()
	if e != nil {
		logs.Error("Failed to get SyscallConn", zap.Error(e))
		return
	}
	e = r.Control(func(fd uintptr) {
		intfd := int(fd)
		if e := syscall.SetsockoptInt(intfd, syscall.SOL_SOCKET, syscall.SO_NOSIGPIPE, 1); e != nil {
			logs.Error("Failed to set SO_NOSIGPIPE", zap.Error(e))
		}
	})
	if e != nil {
		logs.Error("Failed to set SO_NOSIGPIPE", zap.Error(e))
	}
}
