package scenario

import (
	"strings"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-supernumerary/pkg/shared"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
)

// Login scenario
type Login struct {
	Scenario
}

// NewLogin create new instance
func NewLogin(isFinal bool) shared.Screenwriter {
	s := new(Login)
	s.isFinal = isFinal
	return s
}

// Play execute Login scenario
func (s *Login) Play(act shared.Actor) {
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
func (s *Login) sendCode(act shared.Actor) (*msg.MessageEnvelope, shared.SuccessCallback, shared.TimeoutCallback) {
	reqEnv := AuthSendCode(act.GetPhone())

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
			x.Unmarshal(resp.Message)
			// chain next request here
			act.ExecuteRequest(s.login(x, act))
		} else {
			s.failed(act, elapsed, resp.RequestID, "sendCode() SuccessCB response is not AuthSentCode , Constructor :"+msg.ConstructorNames[resp.Constructor])
		}
	}

	return reqEnv, successCB, timeoutCB
}

// login Step : 2
func (s *Login) login(resp *msg.AuthSentCode, act shared.Actor) (*msg.MessageEnvelope, shared.SuccessCallback, shared.TimeoutCallback) {
	if strings.HasPrefix(strings.TrimLeft("+", resp.Phone), "237400") {
		code := resp.Phone[len(resp.Phone)-4:]

		reqEnv := AuthLogin(resp.Phone, code, resp.PhoneCodeHash)

		timeoutCB := func(requestID uint64, elapsed time.Duration) {
			// Reporter failed
			act.SetTimeout(msg.C_AuthLogin, elapsed)
			s.failed(act, elapsed, requestID, "login() TimedOut")
		}

		successCB := func(resp *msg.MessageEnvelope, elapsed time.Duration) {
			act.SetSuccess(msg.C_AuthLogin, elapsed)
			if s.isErrorResponse(act, elapsed, resp, "login()") {
				return
			}
			if resp.Constructor == msg.C_AuthAuthorization {
				x := new(msg.AuthAuthorization)
				x.Unmarshal(resp.Message)

				act.SetUserInfo(x.User.ID, x.User.Username, x.User.FirstName+" "+x.User.LastName)

				if s.isFinal {
					err := act.Save()
					if err != nil {
						s.log(act, "login() Actor.Save(), Err : "+err.Error(), elapsed, resp.RequestID)
					}
				}

				s.completed(act, elapsed, resp.RequestID, "login() Success")

			} else {
				s.failed(act, elapsed, resp.RequestID, "sendCode() SuccessCB response is not AuthAuthorization , Constructor :"+msg.ConstructorNames[resp.Constructor])
			}
		}

		return reqEnv, successCB, timeoutCB
	}

	s.failed(act, -1, 0, "login() phone number does not start with 237400")

	return nil, nil, nil
}
