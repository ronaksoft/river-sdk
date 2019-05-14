package scenario

import (
	"strconv"
	"sync"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-loadtester/logs"

	"go.uber.org/zap"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-loadtester/shared"
)

// Scenario is something that screenwriter create
type Scenario struct {
	isFinal bool // final scenario will stop the actors when its finished
	wait    sync.WaitGroup
	result  bool // true if completed false if failed
}

// Play execute scenario
func (s *Scenario) Play(act shared.Actor) {
	panic("Not implemented")
}

// Wait until play overs
func (s *Scenario) Wait(act shared.Actor) bool {
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

func (s *Scenario) failed(act shared.Actor, elapsed time.Duration, reqID uint64, str string) {
	s.result = false
	act.SetActorSucceed(false)

	//----
	shared.SetFailedRequest(reqID, act.GetAuthID(), act.GetPhone(), "Timeout Request "+str)
	//----

	logs.Error(act.GetPhone()+"\tReqID:"+strconv.FormatUint(reqID, 10)+"\t failed() : "+str, zap.Duration("elapsed", elapsed))
	s.wait.Done()
}

func (s *Scenario) completed(act shared.Actor, elapsed time.Duration, reqID uint64, str string) {
	s.result = true
	act.SetActorSucceed(true)
	logs.Info(act.GetPhone()+"\tReqID:"+strconv.FormatUint(reqID, 10)+"\t completed() : "+str, zap.Duration("elapsed", elapsed))
	s.wait.Done()
}

func (s *Scenario) log(act shared.Actor, str string, elapsed time.Duration, reqID uint64) {
	if elapsed > 0 {
		logs.Warn(act.GetPhone()+"\tReqID:"+strconv.FormatUint(reqID, 10)+"\t log() : "+str, zap.Duration("elapsed", elapsed))
	} else {
		logs.Warn(act.GetPhone() + "\tReqID:" + strconv.FormatUint(reqID, 10) + "\t log() : " + str)
	}
}

func (s *Scenario) isErrorResponse(act shared.Actor, elapsed time.Duration, resp *msg.MessageEnvelope, cbName string) bool {
	if resp.Constructor == msg.C_Error {
		act.ReceivedErrorResponse()
		x := new(msg.Error)
		x.Unmarshal(resp.Message)
		s.result = false
		act.SetActorSucceed(false)

		//----
		shared.SetFailedRequest(resp.RequestID, act.GetAuthID(), act.GetPhone(), "Error Response CallBackName :"+cbName+"\tErr:"+x.String())
		//----

		logs.Error(act.GetPhone()+"\tReqID:"+strconv.FormatUint(resp.RequestID, 10)+"\t isErrorResponse(): ", zap.String("CallBackName", cbName), zap.String("Err", x.String()), zap.Duration("elapsed", elapsed))
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
func Play(act shared.Actor, scenario shared.Screenwriter) bool {
	scenario.Play(act)
	scenario.Wait(act)
	return scenario.GetResult()
}
