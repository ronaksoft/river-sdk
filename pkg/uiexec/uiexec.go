package uiexec

import (
	"sync"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
)

var (
	mx             sync.Mutex
	exec           *UIExecuter
	uiExecInterval = 100 * time.Millisecond
)

// UIExecuter layer
type UIExecuter struct {
	chUIExecuter chan func()
	chStop       chan bool
}

func init() {
	Ctx()
}

// InitUIExec added for clearify cuz init will initialize this
func InitUIExec() {
	Ctx()
}

// Ctx singletone
func Ctx() *UIExecuter {
	if exec == nil {
		mx.Lock()
		defer mx.Unlock()
		if exec == nil {
			exec = &UIExecuter{
				chUIExecuter: make(chan func(), 100),
				chStop:       make(chan bool),
			}
			exec.Start()
		}
	}
	return exec
}

// Start starts UIExecuter listener
func (c *UIExecuter) Start() {
	go c.UIExecuter()
}

// Stop sent stop signal
func (c *UIExecuter) Stop() {
	select {
	case c.chStop <- true:
		logs.Debug("CMD::Stop() sent stop signal")
	case <-time.After(uiExecInterval):
		logs.Debug("CMD::Stop() cmd is not started")
	}
}

// Exec pass given function to UIExecuter buffered channel
func (c *UIExecuter) Exec(fn func()) {
	select {
	case c.chUIExecuter <- fn:
		logs.Debug("CMD::Exec() sent to channel")
	case <-time.After(uiExecInterval):
		logs.Warn("CMD::Exec() cmd is not started")
	}
}

// UIExecuter Pass responses to external handler (UI) one by one
func (c *UIExecuter) UIExecuter() {
	for {
		select {
		case fn := <-c.chUIExecuter:
			fn()
		case <-c.chStop:
			return
		}
	}
}
