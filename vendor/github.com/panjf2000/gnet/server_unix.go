// Copyright (c) 2019 Andy Pan
// Copyright (c) 2018 Joshua J Baker
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

// +build linux freebsd dragonfly darwin

package gnet

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/panjf2000/gnet/errors"
	"github.com/panjf2000/gnet/internal/netpoll"
)

type server struct {
	ln           *listener          // the listener for accepting new connections
	lb           loadBalancer       // event-loops for handling events
	wg           sync.WaitGroup     // event-loop close WaitGroup
	opts         *Options           // options with server
	once         sync.Once          // make sure only signalShutdown once
	cond         *sync.Cond         // shutdown signaler
	codec        ICodec             // codec for TCP stream
	mainLoop     *eventloop         // main event-loop for accepting connections
	inShutdown   int32              // whether the server is in shutdown
	tickerCtx    context.Context    // context for ticker
	cancelTicker context.CancelFunc // function to stop the ticker
	eventHandler EventHandler       // user eventHandler
}

func (svr *server) isInShutdown() bool {
	return atomic.LoadInt32(&svr.inShutdown) == 1
}

// waitForShutdown waits for a signal to shutdown.
func (svr *server) waitForShutdown() {
	svr.cond.L.Lock()
	svr.cond.Wait()
	svr.cond.L.Unlock()
}

// signalShutdown signals the server to shut down.
func (svr *server) signalShutdown() {
	svr.once.Do(func() {
		svr.cond.L.Lock()
		svr.cond.Signal()
		svr.cond.L.Unlock()
	})
}

func (svr *server) startEventLoops() {
	svr.lb.iterate(func(i int, el *eventloop) bool {
		svr.wg.Add(1)
		go func() {
			el.loopRun(svr.opts.LockOSThread)
			svr.wg.Done()
		}()
		return true
	})
}

func (svr *server) closeEventLoops() {
	svr.lb.iterate(func(i int, el *eventloop) bool {
		_ = el.poller.Close()
		return true
	})
}

func (svr *server) startSubReactors() {
	svr.lb.iterate(func(i int, el *eventloop) bool {
		svr.wg.Add(1)
		go func() {
			svr.activateSubReactor(el, svr.opts.LockOSThread)
			svr.wg.Done()
		}()
		return true
	})
}

func (svr *server) activateEventLoops(numEventLoop int) (err error) {
	var striker *eventloop
	// Create loops locally and bind the listeners.
	for i := 0; i < numEventLoop; i++ {
		l := svr.ln
		if i > 0 && svr.opts.ReusePort {
			if l, err = initListener(svr.ln.network, svr.ln.addr, svr.opts); err != nil {
				return
			}
		}

		var p *netpoll.Poller
		if p, err = netpoll.OpenPoller(); err == nil {
			el := new(eventloop)
			el.ln = l
			el.svr = svr
			el.poller = p
			el.buffer = make([]byte, svr.opts.ReadBufferCap)
			el.connections = make(map[int]*conn)
			el.eventHandler = svr.eventHandler
			_ = el.poller.AddRead(el.ln.fd)
			svr.lb.register(el)

			// Start the ticker.
			if el.idx == 0 && svr.opts.Ticker {
				striker = el
			}
		} else {
			return
		}
	}

	// Start event-loops in background.
	svr.startEventLoops()

	go striker.loopTicker(svr.tickerCtx)

	return
}

func (svr *server) activateReactors(numEventLoop int) error {
	for i := 0; i < numEventLoop; i++ {
		if p, err := netpoll.OpenPoller(); err == nil {
			el := new(eventloop)
			el.ln = svr.ln
			el.svr = svr
			el.poller = p
			el.buffer = make([]byte, svr.opts.ReadBufferCap)
			el.connections = make(map[int]*conn)
			el.eventHandler = svr.eventHandler
			svr.lb.register(el)
		} else {
			return err
		}
	}

	// Start sub reactors in background.
	svr.startSubReactors()

	if p, err := netpoll.OpenPoller(); err == nil {
		el := new(eventloop)
		el.ln = svr.ln
		el.idx = -1
		el.svr = svr
		el.poller = p
		el.eventHandler = svr.eventHandler
		_ = el.poller.AddRead(el.ln.fd)
		svr.mainLoop = el

		// Start main reactor in background.
		svr.wg.Add(1)
		go func() {
			svr.activateMainReactor(svr.opts.LockOSThread)
			svr.wg.Done()
		}()
	} else {
		return err
	}

	// Start the ticker.
	if svr.opts.Ticker {
		go svr.mainLoop.loopTicker(svr.tickerCtx)
	}

	return nil
}

func (svr *server) start(numEventLoop int) error {
	if svr.opts.ReusePort || svr.ln.network == "udp" {
		return svr.activateEventLoops(numEventLoop)
	}

	return svr.activateReactors(numEventLoop)
}

func (svr *server) stop(s Server) {
	// Wait on a signal for shutdown
	svr.waitForShutdown()

	svr.eventHandler.OnShutdown(s)

	// Notify all loops to close by closing all listeners
	svr.lb.iterate(func(i int, el *eventloop) bool {
		err := el.poller.UrgentTrigger(func(_ interface{}) error { return errors.ErrServerShutdown }, nil)
		if err != nil {
			svr.opts.Logger.Errorf("failed to call UrgentTrigger on sub event-loop when stopping server")
		}
		return true
	})

	if svr.mainLoop != nil {
		svr.ln.close()
		err := svr.mainLoop.poller.UrgentTrigger(func(_ interface{}) error { return errors.ErrServerShutdown }, nil)
		if err != nil {
			svr.opts.Logger.Errorf("failed to call UrgentTrigger on main event-loop when stopping server")
		}
	}

	// Wait on all loops to complete reading events
	svr.wg.Wait()

	svr.closeEventLoops()

	if svr.mainLoop != nil {
		err := svr.mainLoop.poller.Close()
		if err != nil {
			svr.opts.Logger.Errorf("failed to close poller when stopping server")
		}
	}

	// Stop the ticker.
	if svr.opts.Ticker {
		svr.cancelTicker()
	}

	atomic.StoreInt32(&svr.inShutdown, 1)
}

func serve(eventHandler EventHandler, listener *listener, options *Options, protoAddr string) error {
	// Figure out the proper number of event-loops/goroutines to run.
	numEventLoop := 1
	if options.Multicore {
		numEventLoop = runtime.NumCPU()
	}
	if options.NumEventLoop > 0 {
		numEventLoop = options.NumEventLoop
	}

	svr := new(server)
	svr.opts = options
	svr.eventHandler = eventHandler
	svr.ln = listener

	switch options.LB {
	case RoundRobin:
		svr.lb = new(roundRobinLoadBalancer)
	case LeastConnections:
		svr.lb = new(leastConnectionsLoadBalancer)
	case SourceAddrHash:
		svr.lb = new(sourceAddrHashLoadBalancer)
	}

	svr.cond = sync.NewCond(&sync.Mutex{})
	if svr.opts.Ticker {
		svr.tickerCtx, svr.cancelTicker = context.WithCancel(context.Background())
	}
	svr.codec = func() ICodec {
		if options.Codec == nil {
			return new(BuiltInFrameCodec)
		}
		return options.Codec
	}()

	server := Server{
		svr:          svr,
		Multicore:    options.Multicore,
		Addr:         listener.lnaddr,
		NumEventLoop: numEventLoop,
		ReusePort:    options.ReusePort,
		TCPKeepAlive: options.TCPKeepAlive,
	}
	switch svr.eventHandler.OnInitComplete(server) {
	case None:
	case Shutdown:
		return nil
	}

	if err := svr.start(numEventLoop); err != nil {
		svr.closeEventLoops()
		svr.opts.Logger.Errorf("gnet server is stopping with error: %v", err)
		return err
	}
	defer svr.stop(server)

	allServers.Store(protoAddr, svr)

	return nil
}
