package scenario

import (
	"time"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"
	"git.ronaksoftware.com/ronak/riversdk/msg"
)

type SendMessage struct {
	Scenario
}

// NewSendMessage create new instance
func NewSendMessage(isFinal bool) shared.Screenwriter {
	s := new(SendMessage)
	s.isFinal = isFinal
	return s
}

// Play execute SendMessage scenario
func (s *SendMessage) Play(act shared.Acter) {
	if act.GetAuthID() == 0 {
		Play(act, NewCreateAuthKey(false))
	}
	if act.GetUserID() == 0 {
		Play(act, NewLogin(false))
	}
	if len(act.GetPeers()) == 0 {
		Play(act, NewImportContact(false))
	}
	peers := act.GetPeers()
	s.AddJobs(len(peers))
	for _, p := range peers {
		act.ExecuteRequest(s.messageSend(act, p))
	}
}

// messageSend : Step 1
func (s *SendMessage) messageSend(act shared.Acter, peer *shared.PeerInfo) (*msg.MessageEnvelope, shared.SuccessCallback, shared.TimeoutCallback) {
	reqEnv := MessageSend(peer)

	timeoutCB := func(requestID uint64, elapsed time.Duration) {
		// Reporter failed
		act.SetTimeout(msg.C_MessagesSend, elapsed)
		s.failed(act, elapsed, "messageSend() Timeout")
	}

	successCB := func(resp *msg.MessageEnvelope, elapsed time.Duration) {
		act.SetSuccess(msg.C_MessagesSend, elapsed)
		if s.isErrorResponse(act, elapsed, resp) {
			return
		}
		if resp.Constructor == msg.C_MessagesSent {
			x := new(msg.MessagesSent)
			x.Unmarshal(resp.Message)

			// TODO : Complete Scenario
			s.completed(act, elapsed, "messageSend() Success")
		} else {
			// TODO : Reporter failed
			s.failed(act, elapsed, "messageSend() SuccessCB response is not MessagesSent")
		}
	}

	return reqEnv, successCB, timeoutCB
}
