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
func NewSendMessage() *SendMessage {
	s := new(SendMessage)
	return s
}

// Play execute SendMessage scenario
func (s *SendMessage) Play(act shared.Acter) {
	if act.GetAuthID() == 0 {
		Play(act, NewCreateAuthKey())
	}
	if act.GetUserID() == 0 {
		Play(act, NewLogin())
	}
	if len(act.GetPeers()) == 0 {
		Play(act, NewImportContact())
	}

	for _, p := range act.GetPeers() {
		s.wait.Add(1)
		act.ExecuteRequest(s.messageSend(act, p))
	}
}

// messageSend : Step 1
func (s *SendMessage) messageSend(act shared.Acter, peer *shared.PeerInfo) (*msg.MessageEnvelope, shared.SuccessCallback, shared.TimeoutCallback) {
	reqEnv := MessageSend(peer)

	timeoutCB := func(requestID uint64, elapsed time.Duration) {
		// TODO : Reporter failed
		s.failed("messageSend() Timeout")
	}

	successCB := func(resp *msg.MessageEnvelope, elapsed time.Duration) {
		if s.isErrorResponse(resp) {
			return
		}
		if resp.Constructor == msg.C_MessagesSent {
			x := new(msg.MessagesSent)
			x.Unmarshal(resp.Message)

			// TODO : Complete Scenario
			s.completed("messageSend() Success")
		} else {
			// TODO : Reporter failed
			s.failed("messageSend() SuccessCB response is not MessagesSent")
		}
	}

	return reqEnv, successCB, timeoutCB
}
