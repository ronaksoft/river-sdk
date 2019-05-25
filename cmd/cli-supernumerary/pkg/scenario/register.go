package scenario

import (
	"strings"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-supernumerary/pkg/shared"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
)

// Register Scenario
type Register struct {
	Scenario
}

// NewRegister create new instance
func NewRegister(isFinal bool) shared.Screenwriter {
	s := new(Register)
	s.isFinal = isFinal
	return s
}

// Play execute Register scenario
func (s *Register) Play(act shared.Actor) {
	if act.GetUserID() > 0 {
		s.log(act, "Actor already have userID", 0, 0)
		return
	}

	if act.GetAuthID() == 0 {
		s.log(act, "AuthID is ZERO, Scenario Failed", 0, 0)
		return
	}

	s.AddJobs(1)
	act.ExecuteRequest(s.sendCode(act))
}

// sendCode : Step 1
func (s *Register) sendCode(act shared.Actor) (*msg.MessageEnvelope, shared.SuccessCallback, shared.TimeoutCallback) {
	envReq := AuthSendCode(act.GetPhone())
	timeoutCB := func(requestID uint64, elapsed time.Duration) {
		// Reporter failed
		act.SetTimeout(msg.C_AuthSendCode, elapsed)
		s.failed(act, elapsed, requestID, "sendCode() Timeout")
	}
	successCB := func(resp *msg.MessageEnvelope, elapsed time.Duration) {
		act.SetSuccess(msg.C_AuthSendCode, elapsed)
		if s.isErrorResponse(act, elapsed, resp, "sendCode()") {
			return
		}
		if resp.Constructor == msg.C_AuthSentCode {
			x := new(msg.AuthSentCode)
			_ = x.Unmarshal(resp.Message)
			act.ExecuteRequest(s.register(x, act))
		} else {
			s.failed(act, elapsed, resp.RequestID, "sendCode() SuccessCB response is not AuthSentCode")
		}
	}

	return envReq, successCB, timeoutCB
}

// register : Step 2
func (s *Register) register(resp *msg.AuthSentCode, act shared.Actor) (*msg.MessageEnvelope, shared.SuccessCallback, shared.TimeoutCallback) {
	if strings.HasPrefix(resp.Phone, "237400") {
		code := resp.Phone[len(resp.Phone)-4:]
		envReq := AuthRegister(resp.Phone, code, resp.PhoneCodeHash)

		timeoutCB := func(requestID uint64, elapsed time.Duration) {
			// Reporter failed
			act.SetTimeout(msg.C_AuthRegister, elapsed)
			s.failed(act, elapsed, requestID, "register() TimedOut")
		}
		successCB := func(resp *msg.MessageEnvelope, elapsed time.Duration) {
			act.SetSuccess(msg.C_AuthRegister, elapsed)

			if s.isErrorResponse(act, elapsed, resp, "register()") {
				return
			}
			if resp.Constructor == msg.C_AuthAuthorization {
				x := new(msg.AuthAuthorization)
				_ = x.Unmarshal(resp.Message)

				act.SetUserInfo(x.User.ID, x.User.Username, x.User.FirstName+" "+x.User.LastName)
				if s.isFinal {
					err := act.Save()
					if err != nil {
						s.log(act, "register() Actor.Save(), Err : "+err.Error(), elapsed, resp.RequestID)
					}
				}
				s.completed(act, elapsed, resp.RequestID, "register() Success")
			} else {
				s.failed(act, elapsed, resp.RequestID, "sendCode() SuccessCB response is not AuthAuthorization")
			}
		}

		return envReq, successCB, timeoutCB
	}

	s.failed(act, -1, 0, "login() phone number does not start with 237400")

	return nil, nil, nil
}
