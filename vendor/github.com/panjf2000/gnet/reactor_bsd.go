// Copyright (c) 2019 Andy Pan
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// +build freebsd dragonfly darwin

package gnet

import (
	"runtime"

	"github.com/panjf2000/gnet/internal/netpoll"
)

func (svr *server) activateMainReactor(lockOSThread bool) {
	if lockOSThread {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
	}

	defer svr.signalShutdown()

	err := svr.mainLoop.poller.Polling(func(fd int, filter int16) error { return svr.acceptNewConnection(fd) })
	svr.logger.Infof("Main reactor is exiting due to error: %v", err)
}

func (svr *server) activateSubReactor(el *eventloop, lockOSThread bool) {
	if lockOSThread {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
	}

	defer func() {
		el.closeAllConns()
		svr.signalShutdown()
	}()

	err := el.poller.Polling(func(fd int, filter int16) (err error) {
		if c, ack := el.connections[fd]; ack {
			switch filter {
			case netpoll.EVFilterSock:
				err = el.loopCloseConn(c, nil)
			case netpoll.EVFilterWrite:
				err = el.loopWrite(c)
			case netpoll.EVFilterRead:
				err = el.loopRead(c)
			}
		}
		return
	})
	svr.logger.Infof("Event-loop(%d) is exiting normally on the signal error: %v", el.idx, err)
}
