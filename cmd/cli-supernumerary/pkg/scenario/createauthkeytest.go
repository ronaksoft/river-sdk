package scenario

import (
	"time"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-supernumerary/pkg/shared"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
)

// CreateAuthKeyTest scenario
type CreateAuthKeyTest struct {
	Scenario
}

// NewCreateAuthKey initiate CreateAuthKey scenario
func NewCreateAuthKeyTest(isFinal bool) shared.Screenwriter {
	s := new(CreateAuthKeyTest)
	s.isFinal = isFinal

	return s
}

// Play execute CreateAuthKey scenario
func (s *CreateAuthKeyTest) Play(act shared.Actor) {
	if act.GetAuthID() != 0 {
		s.log(act, "Actor already have AuthID", 0, 0)
		return
	}
	s.AddJobs(1)
	act.ExecuteRequest(s.initConnectTest(act))
}

func (s *CreateAuthKeyTest) initConnectTest(act shared.Actor) (*msg.MessageEnvelope, shared.SuccessCallback, shared.TimeoutCallback) {
	reqEnv := InitConnectTest()

	timeoutCB := func(requestID uint64, elapsed time.Duration) {
		// Reporter failed
		act.SetTimeout(msg.C_InitConnectTest, elapsed)
		s.failed(act, elapsed, requestID, "initConnectTest() Timeout")
	}

	successCB := func(resp *msg.MessageEnvelope, elapsed time.Duration) {
		act.SetSuccess(msg.C_InitConnectTest, elapsed)
		if s.isErrorResponse(act, elapsed, resp, "initConnectTest()") {
			return
		}
		if resp.Constructor == msg.C_InitTestAuth {
			x := new(msg.InitTestAuth)
			x.Unmarshal(resp.Message)

			// save authKey && authID
			act.SetAuthInfo(x.AuthID, x.AuthKey)
			if s.isFinal {
				err := act.Save()
				if err != nil {
					s.log(act, "initConnectTest() Actor.save(), Err : "+err.Error(), elapsed, resp.RequestID)
				}
			}
			s.completed(act, elapsed, resp.RequestID, "initConnectTest() Success")
		} else {
			s.failed(act, elapsed, resp.RequestID, "initConnectTest() successCB response type is not InitTestAuth , Constructor :"+msg.ConstructorNames[resp.Constructor])
		}
	}

	return reqEnv, successCB, timeoutCB
}
