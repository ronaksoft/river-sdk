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
	isFinal bool // final scenario will stop the actors when its finished
	wait    sync.WaitGroup
	result  bool // true if completed false if failed
}

// Play execute scenario
func (s *Scenario) Play(act shared.Acter) {
	panic("Not implemented")
}

// Wait until play overs
func (s *Scenario) Wait(act shared.Acter) bool {
	s.wait.Wait()
	if s.isFinal {
		act.Stop()
	}
	return s.GetResult()
}

//AddJobs add delta to wait group
func (s *Scenario) AddJobs(delta int) {
	s.wait.Add(delta)
}

func (s *Scenario) failed(act shared.Acter, elapsed time.Duration, str string) {
	s.result = false
	log.LOG_Error(act.GetPhone()+"\t failed() : "+str, zap.Duration("elapsed", elapsed))
	s.wait.Done()
}

func (s *Scenario) completed(act shared.Acter, elapsed time.Duration, str string) {
	s.result = true
	log.LOG_Info(act.GetPhone()+"\t completed() : "+str, zap.Duration("elapsed", elapsed))
	s.wait.Done()
}

func (s *Scenario) log(act shared.Acter, str string, elapsed time.Duration) {
	if elapsed > 0 {
		log.LOG_Warn(act.GetPhone()+"\t log() : "+str, zap.Duration("elapsed", elapsed))
	} else {
		log.LOG_Warn(act.GetPhone() + "\t log() : " + str)
	}
}

func (s *Scenario) isErrorResponse(act shared.Acter, elapsed time.Duration, resp *msg.MessageEnvelope, cbName string) bool {
	if resp.Constructor == msg.C_Error {
		act.ReceivedErrorResponse()
		x := new(msg.Error)
		x.Unmarshal(resp.Message)
		s.result = false
		log.LOG_Error(act.GetPhone()+"\t isErrorResponse(): ", zap.String("CallBackName", cbName), zap.String("Err", x.String()), zap.Duration("elapsed", elapsed))
		s.wait.Done()
		return true
	}
	return false
}

// SetFinal set final state in order to clean up actor when scenario ends
func (s *Scenario) SetFinal(isFinal bool) {
	s.isFinal = isFinal
}

// IsFinal get final state
func (s *Scenario) IsFinal() bool {
	return s.isFinal
}

// GetResult returns scenario result
func (s *Scenario) GetResult() bool {
	return s.result
}

// Play puts actor in scenario to play its role and wait to show overs
func Play(act shared.Acter, scenario shared.Screenwriter) bool {
	scenario.Play(act)
	scenario.Wait(act)
	return scenario.GetResult()
}