package uiexec

import (
	"sync"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
)

var (
	mx             sync.Mutex
	exec           *UIExecutor
	uiExecInterval = 100 * time.Millisecond
)

// UIExecutor layer
type UIExecutor struct {
	chUIExecutor chan func()
	chStop       chan bool
}

func init() {
	Ctx()
}

// InitUIExec added for clarify cuz init will initialize this
func InitUIExec() {
	Ctx()
}

// Ctx singleton
func Ctx() *UIExecutor {
	if exec == nil {
		mx.Lock()
		defer mx.Unlock()
		if exec == nil {
			exec = &UIExecutor{
				chUIExecutor: make(chan func(), 100),
				chStop:       make(chan bool),
			}
			exec.Start()
		}
	}
	return exec
}

// Start starts UIExecutor listener
func (c *UIExecutor) Start() {
	go c.UIExecutor()
}

// Stop sent stop signal
func (c *UIExecutor) Stop() {
	logs.Debug("StopServices-UIExecutor::Stop() called")

	select {
	case c.chStop <- true:
		logs.Debug("CMD::Stop() sent stop signal")
	case <-time.After(uiExecInterval):
		logs.Debug("CMD::Stop() cmd is not started")
	}
	exec = nil
}

// Exec pass given function to UIExecutor buffered channel
func (c *UIExecutor) Exec(fn func()) {
	select {
	case c.chUIExecutor <- fn:
		logs.Debug("CMD::Exec() sent to channel")
	case <-time.After(uiExecInterval):
		logs.Warn("CMD::Exec() cmd is not started")
	}
}

// UIExecutor Pass responses to external handler (UI) one by one
func (c *UIExecutor) UIExecutor() {
	for {
		select {
		case fn := <-c.chUIExecutor:
			fn()
		case <-c.chStop:
			return
		}
	}
}
