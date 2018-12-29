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
}

// NewLogin create new instance
func NewLogin() *Login {
	s := new(Login)
	return s
}

// Execute Login scenario
func (s *Login) Execute(act *actor.Actor) {

	//check if actor does not have AuthID it means it didn't run CreateAuthKey scenario
	if act.AuthID == 0 {
		createAuthKey := NewCreateAuthKey()
		createAuthKey.Execute(act)
	}

	act.ExecuteRequest(s.sendCode(act))
}

//sendCode : Step 1
func (s *Login) sendCode(act *actor.Actor) (*msg.MessageEnvelope, shared.SuccessCallback, shared.TimeoutCallback) {
	envReq := AuthSendCode(act.Phone)

	timeoutCB := func(requestID uint64, elapsed time.Duration) {
		// TODO : Reporter failed
	}

	successCB := func(resp *msg.MessageEnvelope, elapsed time.Duration) {
		if resp.Constructor == msg.C_AuthSentCode {
			x := new(msg.AuthSentCode)
			x.Unmarshal(resp.Message)
			// chain next request here
			act.ExecuteRequest(s.login(x, act))
		} else {
			// TODO : Reporter failed
		}
	}

	return envReq, successCB, timeoutCB
}

// login Step : 2
func (s *Login) login(resp *msg.AuthSentCode, act *actor.Actor) (*msg.MessageEnvelope, shared.SuccessCallback, shared.TimeoutCallback) {
	if strings.HasSuffix(resp.Phone, "237400") {
		code := resp.Phone[len(resp.Phone)-4:]

		reqEnv := AuthLogin(resp.Phone, code, resp.PhoneCodeHash)

		timeoutCB := func(requestID uint64, elapsed time.Duration) {
			// TODO : Reporter failed
		}

		successCB := func(resp *msg.MessageEnvelope, elapsed time.Duration) {
			if resp.Constructor == msg.C_AuthAuthorization {
				x := new(msg.AuthAuthorization)
				x.Unmarshal(resp.Message)

				// TODO : Complete Scenario
				// x.User

			} else {
				// TODO : Reporter failed
			}
		}

		return reqEnv, successCB, timeoutCB
	}

	// TODO : Reporter failed
	return nil, nil, nil
}
