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

func init() {
	GetUIExecuter()
}

// InitUIExecuter added for clearify cuz init will initialize this
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
				chUIExecuter: make(chan func(), 100),
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
	// uiExecTimeout := 50 * time.Millisecond

	for {
		select {
		case fn := <-c.chUIExecuter:
			fn()
			// // prevent (UI) external handler block our callbacks forever
			// chDone := make(chan bool)
			// go func(ch chan bool) {
			// 	fn()
			// 	ch <- true
			// }(chDone)
			// select {
			// case <-time.After(uiExecTimeout):
			// 	log.LOG.Debug("cmd::UIExecuter() execute fn() timedout")
			// case <-chDone:
			// 	log.LOG.Debug("cmd::UIExecuter() execute fn() successfully finished")
			// }
			// chDone = nil

			// case <-c.chStop:
			// 	return
		}
	}
}
