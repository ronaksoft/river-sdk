package scenario

import (
	"strings"
	"time"

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

// Play execute Login scenario
func (s *Login) Play(act shared.Acter) {
	if act.GetUserID() > 0 {
		s.log("Actor already have UserID", 0)
		return
	}
	if act.GetAuthID() == 0 {
		Play(act, NewCreateAuthKey())
	}
	s.wait.Add(1)
	act.ExecuteRequest(s.sendCode(act))
}

//sendCode : Step 1
func (s *Login) sendCode(act shared.Acter) (*msg.MessageEnvelope, shared.SuccessCallback, shared.TimeoutCallback) {
	reqEnv := AuthSendCode(act.GetPhone())

	timeoutCB := func(requestID uint64, elapsed time.Duration) {
		// Reporter failed
		act.SetTimeout(msg.C_AuthSendCode, elapsed)
		s.failed(act, elapsed, "sendCode() Timeout")
	}

	successCB := func(resp *msg.MessageEnvelope, elapsed time.Duration) {
		act.SetSuccess(msg.C_AuthSendCode, elapsed)
		if s.isErrorResponse(act, elapsed, resp) {
			return
		}
		if resp.Constructor == msg.C_AuthSentCode {
			x := new(msg.AuthSentCode)
			x.Unmarshal(resp.Message)
			// chain next request here
			act.ExecuteRequest(s.login(x, act))
		} else {
			// TODO : Reporter failed
			s.failed(act, elapsed, "sendCode() SuccessCB response is not AuthSentCode")
		}
	}

	return reqEnv, successCB, timeoutCB
}

// login Step : 2
func (s *Login) login(resp *msg.AuthSentCode, act shared.Acter) (*msg.MessageEnvelope, shared.SuccessCallback, shared.TimeoutCallback) {
	if strings.HasPrefix(resp.Phone, "237400") {
		code := resp.Phone[len(resp.Phone)-4:]

		reqEnv := AuthLogin(resp.Phone, code, resp.PhoneCodeHash)

		timeoutCB := func(requestID uint64, elapsed time.Duration) {
			//Reporter failed
			act.SetTimeout(msg.C_AuthLogin, elapsed)
			s.failed(act, elapsed, "login() TimedOut")
		}

		successCB := func(resp *msg.MessageEnvelope, elapsed time.Duration) {
			act.SetSuccess(msg.C_AuthLogin, elapsed)
			if s.isErrorResponse(act, elapsed, resp) {
				return
			}
			if resp.Constructor == msg.C_AuthAuthorization {
				x := new(msg.AuthAuthorization)
				x.Unmarshal(resp.Message)

				// TODO : Complete Scenario
				act.SetUserInfo(x.User.ID, x.User.Username, x.User.FirstName+" "+x.User.LastName)
				err := act.Save()
				if err != nil {
					s.log("contactImport() Actor.Save(), Err : "+err.Error(), elapsed)
				}

				s.completed(act, elapsed, "login() Success")

			} else {
				// TODO : Reporter failed
				s.failed(act, elapsed, "sendCode() SuccessCB response is not AuthAuthorization")
			}
		}

		return reqEnv, successCB, timeoutCB
	}

	// TODO : Reporter failed
	s.failed(act, -1, "login() phone number does not start with 237400")

	return nil, nil, nil
}
