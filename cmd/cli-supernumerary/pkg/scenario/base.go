package scenario

import (
	"fmt"
	log "git.ronaksoftware.com/ronak/toolbox/logger"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-supernumerary/pkg/shared"
)

var (
	_Log log.Logger
)

func init() {
	_Log = log.NewConsoleLogger()
}

func SetLogger(l log.Logger) {
	_Log = l
}

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

// AddJobs add delta to wait group
func (s *Scenario) AddJobs(delta int) {
	s.wait.Add(delta)
}

func (s *Scenario) failed(act shared.Actor, elapsed time.Duration, reqID uint64, str string) {
	s.result = false
	act.SetActorSucceed(false)

	// ----
	shared.SetFailedRequest(reqID, act.GetAuthID(), act.GetPhone(), "Timeout Request "+str)
	// ----

	_Log.Error("Scenario Failed",
		zap.Uint64("ReqID", reqID),
		zap.String("Phone", act.GetPhone()),
		zap.String("Str", str),
		zap.Duration("Elapsed", elapsed),
	)
	s.wait.Done()
}

func (s *Scenario) completed(act shared.Actor, elapsed time.Duration, reqID uint64, str string) {
	s.result = true
	act.SetActorSucceed(true)
	_Log.Info("Scenario Completed",
		zap.Uint64("ReqID", reqID),
		zap.String("Phone", act.GetPhone()),
		zap.String("Str", str),
		zap.Duration("Elapsed", elapsed),
	)
	s.wait.Done()
}

func (s *Scenario) log(act shared.Actor, str string, elapsed time.Duration, reqID uint64) {
	if elapsed > 0 {
		_Log.Warn(fmt.Sprintf("%s: \t ReqID: %d \t %s", act.GetPhone(), reqID, str),
			zap.Duration("Duration", elapsed),
		)
	} else {
		_Log.Warn(fmt.Sprintf("%s: \t ReqID: %d \t %s", act.GetPhone(), reqID, str))
	}
}

func (s *Scenario) isErrorResponse(act shared.Actor, elapsed time.Duration, resp *msg.MessageEnvelope, cbName string) bool {
	if resp.Constructor == msg.C_Error {
		act.ReceivedErrorResponse()
		x := new(msg.Error)
		x.Unmarshal(resp.Message)
		s.result = false
		act.SetActorSucceed(false)

		// ----
		shared.SetFailedRequest(resp.RequestID, act.GetAuthID(), act.GetPhone(), "Error Response CallBackName :"+cbName+"\tErr:"+x.String())
		// ----

		_Log.Error(act.GetPhone()+"\tReqID:"+strconv.FormatUint(resp.RequestID, 10)+"\t isErrorResponse(): ", zap.String("CallBackName", cbName), zap.String("Err", x.String()), zap.Duration("elapsed", elapsed))
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