package scenario

import (
	"time"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-supernumerary/pkg/shared"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
)

type SendGroupMessage struct {
	Scenario
}

// NewSendMessage create new instance
func NewSendGroupMessage(isFinal bool) shared.Screenwriter {
	s := new(SendGroupMessage)
	s.isFinal = isFinal
	return s
}

// Play execute SendMessage scenario
func (s *SendGroupMessage) Play(act shared.Actor) {
	if act.GetAuthID() == 0 {
		s.log(act, "AuthID is ZERO, Scenario Failed", 0, 0)
		return
	}
	if act.GetUserID() == 0 {
		s.log(act, "UserID is ZERO, Scenario Failed", 0, 0)
		return
	}


	peers := act.GetPeers()
	s.AddJobs(len(peers))
	for _, p := range peers {
		act.ExecuteRequest(s.messageSend(act, p))
	}
}

// messageSend : Step 1
func (s *SendGroupMessage) messageSend(act shared.Actor, peer *shared.PeerInfo) (*msg.MessageEnvelope, shared.SuccessCallback, shared.TimeoutCallback) {
	reqEnv := MessageSend(peer)
	timeoutCB := func(requestID uint64, elapsed time.Duration) {
		// Reporter failed
		act.SetTimeout(msg.C_MessagesSend, elapsed)
		s.failed(act, elapsed, requestID, "messageSend() Timeout")
	}

	successCB := func(resp *msg.MessageEnvelope, elapsed time.Duration) {
		act.SetSuccess(msg.C_MessagesSend, elapsed)
		if s.isErrorResponse(act, elapsed, resp, "messageSend()") {
			return
		}
		if resp.Constructor == msg.C_MessagesSent {
			x := new(msg.MessagesSent)
			_ = x.Unmarshal(resp.Message)
			s.completed(act, elapsed, resp.RequestID, "messageSend() Success")
		} else {
			s.failed(act, elapsed, resp.RequestID, "messageSend() SuccessCB response is not MessagesSent, Constructor :"+msg.ConstructorNames[resp.Constructor])
		}
	}

	return reqEnv, successCB, timeoutCB
}
