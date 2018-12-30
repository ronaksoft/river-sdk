package scenario

import (
	"strings"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/actor"
	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"
	"git.ronaksoftware.com/ronak/riversdk/msg"
)

// Login scenario
type Login struct {
	Scenario
}

// NewLogin create new instance
func NewLogin() *Login {
	s := new(Login)
	return s
}

// Execute Login scenario
func (s *Login) Execute(act *actor.Actor) {
	s.wait.Add(1)
	act.ExecuteRequest(s.sendCode(act))
}

//sendCode : Step 1
func (s *Login) sendCode(act *actor.Actor) (*msg.MessageEnvelope, shared.SuccessCallback, shared.TimeoutCallback) {
	reqEnv := AuthSendCode(act.Phone)

	timeoutCB := func(requestID uint64, elapsed time.Duration) {
		// TODO : Reporter failed
		s.failed("sendCode() Timeout")
	}

	successCB := func(resp *msg.MessageEnvelope, elapsed time.Duration) {
		if s.isErrorResponse(resp) {
			return
		}
		if resp.Constructor == msg.C_AuthSentCode {
			x := new(msg.AuthSentCode)
			x.Unmarshal(resp.Message)
			// chain next request here
			act.ExecuteRequest(s.login(x, act))
		} else {
			// TODO : Reporter failed
			s.failed("sendCode() SuccessCB response is not AuthSentCode")
		}
	}

	return reqEnv, successCB, timeoutCB
}

// login Step : 2
func (s *Login) login(resp *msg.AuthSentCode, act *actor.Actor) (*msg.MessageEnvelope, shared.SuccessCallback, shared.TimeoutCallback) {
	if strings.HasPrefix(resp.Phone, "237400") {
		code := resp.Phone[len(resp.Phone)-4:]

		reqEnv := AuthLogin(resp.Phone, code, resp.PhoneCodeHash)

		timeoutCB := func(requestID uint64, elapsed time.Duration) {
			// TODO : Reporter failed
			s.failed("login() TimedOut")
		}

		successCB := func(resp *msg.MessageEnvelope, elapsed time.Duration) {
			if s.isErrorResponse(resp) {
				return
			}
			if resp.Constructor == msg.C_AuthAuthorization {
				x := new(msg.AuthAuthorization)
				x.Unmarshal(resp.Message)

				// TODO : Complete Scenario
				act.UserID = x.User.ID
				act.UserName = x.User.Username
				act.UserFullName = x.User.FirstName + " " + x.User.LastName

				s.completed("login() Success")

			} else {
				// TODO : Reporter failed
				s.failed("sendCode() SuccessCB response is not AuthAuthorization")
			}
		}

		return reqEnv, successCB, timeoutCB
	}

	// TODO : Reporter failed
	s.failed("login() phone number does not start with 237400")

	return nil, nil, nil
}
