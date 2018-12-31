package scenario

import (
	"strings"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"

	"git.ronaksoftware.com/ronak/riversdk/msg"
)

// Register Scenario
type Register struct {
	Scenario
}

// NewRegister create new instance
func NewRegister() *Register {
	s := new(Register)
	return s
}

// Play execute Register scenario
func (s *Register) Play(act shared.Acter) {
	if act.GetAuthID() == 0 {
		Play(act, NewCreateAuthKey())
	}
	s.wait.Add(1)
	act.ExecuteRequest(s.sendCode(act))
}

//sendCode : Step 1
func (s *Register) sendCode(act shared.Acter) (*msg.MessageEnvelope, shared.SuccessCallback, shared.TimeoutCallback) {
	envReq := AuthSendCode(act.GetPhone())
	timeoutCB := func(requestID uint64, elapsed time.Duration) {
		// TODO : Reporter failed
		s.failed("sendCode() Timeout")
	}
	successCB := func(resp *msg.MessageEnvelope, elapsed time.Duration) {
		if s.isErrorResponse(resp) {
			return
		}
		// TODO : chain next request here
		if resp.Constructor == msg.C_AuthSentCode {
			x := new(msg.AuthSentCode)
			x.Unmarshal(resp.Message)
			act.ExecuteRequest(s.register(x, act))
		} else {
			// TODO : Reporter failed
			s.failed("sendCode() SuccessCB response is not AuthSentCode")
		}
	}

	return envReq, successCB, timeoutCB
}

// register : Step 2
func (s *Register) register(resp *msg.AuthSentCode, act shared.Acter) (*msg.MessageEnvelope, shared.SuccessCallback, shared.TimeoutCallback) {
	if strings.HasPrefix(resp.Phone, "237400") {
		code := resp.Phone[len(resp.Phone)-4:]
		envReq := AuthRegister(resp.Phone, code, resp.PhoneCodeHash)

		timeoutCB := func(requestID uint64, elapsed time.Duration) {
			// TODO : Reporter failed
			s.failed("register() TimedOut")
		}
		successCB := func(resp *msg.MessageEnvelope, elapsed time.Duration) {
			if s.isErrorResponse(resp) {
				return
			}
			if resp.Constructor == msg.C_AuthAuthorization {
				x := new(msg.AuthAuthorization)
				x.Unmarshal(resp.Message)

				// TODO : Complete Scenario
				act.SetUserInfo(x.User.ID, x.User.Username, x.User.FirstName+" "+x.User.LastName)
				s.completed("register() Success")

			} else {
				// TODO : Reporter failed
				s.failed("sendCode() SuccessCB response is not AuthAuthorization")
			}
		}

		return envReq, successCB, timeoutCB
	}

	// TODO : Reporter failed
	s.failed("login() phone number does not start with 237400")

	return nil, nil, nil
}
