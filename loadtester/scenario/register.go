package scenario

import (
	"strings"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/actor"
	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"

	"git.ronaksoftware.com/ronak/riversdk/msg"
)

// Register Scenario
type Register struct {
}

// NewRegister create new instance
func NewRegister() *Register {
	s := new(Register)
	return s
}

// Execute Register scenario
func (s *Register) Execute(act *actor.Actor) {

	//check if actor does not have AuthID it means it didn't run CreateAuthKey scenario
	if act.AuthID == 0 {
		createAuthKey := NewCreateAuthKey()
		createAuthKey.Execute(act)
	}

	act.ExecuteRequest(s.sendCode(act))
}

//sendCode : Step 1
func (s *Register) sendCode(act *actor.Actor) (*msg.MessageEnvelope, shared.SuccessCallback, shared.TimeoutCallback) {
	envReq := AuthSendCode(act.Phone)
	timeoutCB := func(requestID uint64, elapsed time.Duration) {
		// TODO : Reporter failed
	}
	successCB := func(resp *msg.MessageEnvelope, elapsed time.Duration) {
		// TODO : chain next request here
		if resp.Constructor == msg.C_AuthSentCode {
			x := new(msg.AuthSentCode)
			x.Unmarshal(resp.Message)
			act.ExecuteRequest(s.register(x, act))
		} else {
			// TODO : Reporter failed
		}
	}

	return envReq, successCB, timeoutCB
}

// register : Step 2
func (s *Register) register(resp *msg.AuthSentCode, act *actor.Actor) (*msg.MessageEnvelope, shared.SuccessCallback, shared.TimeoutCallback) {
	if strings.HasSuffix(resp.Phone, "237400") {
		code := resp.Phone[len(resp.Phone)-4:]
		envReq := AuthRegister(resp.Phone, code, resp.PhoneCodeHash)

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

		return envReq, successCB, timeoutCB
	}

	// TODO : Reporter failed
	return nil, nil, nil
}
