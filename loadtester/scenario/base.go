package scenario

import (
	"sync"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/log"

	"go.uber.org/zap"

	"git.ronaksoftware.com/ronak/riversdk/msg"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"
)

// Scenario is something that screenwriter create
type Scenario struct {
	wait sync.WaitGroup
}

// Play execute scenario
func (s *Scenario) Play(act shared.Acter) {
	panic("Not implemented")
}

// Wait until play overs
func (s *Scenario) Wait() {
	s.wait.Wait()
}

func (s *Scenario) failed(act shared.Acter, elapsed time.Duration, str string) {
	log.LOG_Warn("failed() : "+str, zap.Duration("elapsed", elapsed))
	s.wait.Done()
	act.Stop()
}

func (s *Scenario) completed(act shared.Acter, elapsed time.Duration, str string) {
	log.LOG_Info("completed() : "+str, zap.Duration("elapsed", elapsed))
	s.wait.Done()
	act.Stop()
}

func (s *Scenario) log(str string, elapsed time.Duration) {
	if elapsed > 0 {
		log.LOG_Info("log() : "+str, zap.Duration("elapsed", elapsed))
	}
	log.LOG_Warn("log() : " + str)
}

func (s *Scenario) isErrorResponse(act shared.Acter, elapsed time.Duration, resp *msg.MessageEnvelope) bool {
	if resp.Constructor == msg.C_Error {
		x := new(msg.Error)
		x.Unmarshal(resp.Message)
		log.LOG_Warn("isErrorResponse(): ", zap.String("Err", x.String()), zap.Duration("elapsed", elapsed))
		s.wait.Done()
		act.Stop()
		return true
	}
	return false
}

// Play puts actor in scenario to play its role and wait to show overs
func Play(act shared.Acter, scenario shared.Screenwriter) {
	scenario.Play(act)
	scenario.Wait()
}
