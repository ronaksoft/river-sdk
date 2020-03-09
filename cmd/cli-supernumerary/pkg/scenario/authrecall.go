package scenario

import (
	"time"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-supernumerary/pkg/shared"
	msg "git.ronaksoftware.com/river/msg/chat"
)

type AuthRecall struct {
	Scenario
}

// NewAuthRecall create new instance
func NewAuthRecall(isFinal bool) shared.Screenwriter {
	s := new(AuthRecall)
	s.isFinal = isFinal
	return s
}

// Play execute AuthRecall scenario
func (s *AuthRecall) Play(act shared.Actor) {
	s.AddJobs(1)
	act.ExecuteRequest(s.authRecall(act))
}

// authRecall : Step 1
func (s *AuthRecall) authRecall(act shared.Actor) (*msg.MessageEnvelope, shared.SuccessCallback, shared.TimeoutCallback) {
	reqEnv := AuthRecallReq()

	timeoutCB := func(requestID uint64, elapsed time.Duration) {
		// Reporter failed
		act.SetTimeout(msg.C_AuthRecall, elapsed)
		s.failed(act, elapsed, requestID, "authRecall() Timeout")
	}

	successCB := func(resp *msg.MessageEnvelope, elapsed time.Duration) {
		act.SetSuccess(msg.C_AuthRecall, elapsed)
		if s.isErrorResponse(act, elapsed, resp, "authRecall()") {
			return
		}
		if resp.Constructor == msg.C_AuthRecalled {
			x := new(msg.AuthRecalled)
			x.Unmarshal(resp.Message)

			// TODO : Complete Scenario
			s.completed(act, elapsed, resp.RequestID, "authRecall() Success")
		} else {
			// TODO : Reporter failed
			s.failed(act, elapsed, resp.RequestID, "authRecall() SuccessCB response is not AuthRecalled, Constructor :"+msg.ConstructorNames[resp.Constructor])
		}
	}

	return reqEnv, successCB, timeoutCB
}
