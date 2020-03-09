package scenario

import (
	"time"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-supernumerary/pkg/shared"
	msg "git.ronaksoftware.com/river/msg/chat"
)

// Login scenario
type ResetAuthorizations struct {
	Scenario
}

// NewTerminateSessions create new instance
func NewResetAuthorizations(isFinal bool) shared.Screenwriter {
	s := new(ResetAuthorizations)
	s.isFinal = isFinal
	return s
}

// Play execute Login scenario
func (s *ResetAuthorizations) Play(act shared.Actor) {
	if act.GetUserID() == 0 {
		s.log(act, "Actor has no userID", 0, 0)
		return
	}
	if act.GetAuthID() == 0 {
		s.log(act, "AuthID is ZERO, Scenario Failed", 0, 0)
		return
	}

	s.AddJobs(1)
	act.ExecuteRequest(s.resetAuthorization(act))
}

// sendCode : Step 1
func (s *ResetAuthorizations) resetAuthorization(act shared.Actor) (*msg.MessageEnvelope, shared.SuccessCallback, shared.TimeoutCallback) {
	reqEnv := AccountResetAuthorizations()

	timeoutCB := func(requestID uint64, elapsed time.Duration) {
		// Reporter failed
		act.SetTimeout(msg.C_AccountResetAuthorization, elapsed)
		s.failed(act, elapsed, requestID, "resetAuthorizations() Timeout")
	}

	successCB := func(resp *msg.MessageEnvelope, elapsed time.Duration) {
		act.SetSuccess(msg.C_AccountResetAuthorization, elapsed)
		if s.isErrorResponse(act, elapsed, resp, "resetAuthorizations()") {
			return
		}
		if resp.Constructor == msg.C_Bool {
			x := new(msg.Bool)
			x.Unmarshal(resp.Message)

			s.completed(act, elapsed, resp.RequestID, "resetAuthorizations() Success")
		} else {
			s.failed(act, elapsed, resp.RequestID, "resetAuthorizations() SuccessCB response is not AuthSentCode , Constructor :"+msg.ConstructorNames[resp.Constructor])
		}
	}

	return reqEnv, successCB, timeoutCB
}
