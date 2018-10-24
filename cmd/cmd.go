package cmd

import (
	"sync"

	"git.ronaksoftware.com/ronak/riversdk/log"
)

var (
	mx   sync.Mutex
	exec *UIExecuter
)

// UIExecuter layer
type UIExecuter struct {
	chUIExecuter chan func()
	chStop       chan bool
}

func InitUIExecuter() {
	GetUIExecuter()
}

// GetUIUIExecuter
func GetUIExecuter() *UIExecuter {
	if exec == nil {
		mx.Lock()
		defer mx.Unlock()
		if exec == nil {
			exec = &UIExecuter{
				chUIExecuter: make(chan func()),
				chStop:       make(chan bool),
			}
			exec.Start()
		}
	}
	return exec
}

func (c *UIExecuter) Start() {
	go c.UIExecuter()
}

func (c *UIExecuter) Stop() {
	select {
	case c.chStop <- true:
		log.LOG.Debug("CMD::Stop() sent stop signal")
	default:
		log.LOG.Debug("CMD::Stop() cmd is not started")
	}
}
func (c *UIExecuter) Exec(fn func()) {
	select {
	case c.chUIExecuter <- fn:
		log.LOG.Debug("CMD::Exec() sent to channel")
	default:
		log.LOG.Debug("CMD::Exec() cmd is not started")
	}
}

// Pass responses to external handler (UI) one by one
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
