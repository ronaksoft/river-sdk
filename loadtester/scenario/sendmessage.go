package scenario

import (
	"time"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/actor"
	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"
	"git.ronaksoftware.com/ronak/riversdk/msg"
)

type SendMessage struct {
}

// NewSendMessage create new instance
func NewSendMessage() *SendMessage {
	s := new(SendMessage)
	return s
}

// Execute SendMessage scenario
func (s *SendMessage) Execute(act *actor.Actor) {

	//check if actor does not have AuthID it means it didn't run CreateAuthKey scenario
	if act.AuthID == 0 {
		createAuthKey := NewCreateAuthKey()
		createAuthKey.Execute(act)
	}

	for _, p := range act.Peers {
		act.ExecuteRequest(s.messageSend(act, p))
	}
}

func (s *SendMessage) messageSend(act *actor.Actor, peer *shared.PeerInfo) (*msg.MessageEnvelope, shared.SuccessCallback, shared.TimeoutCallback) {
	reqEnv := MessageSend(peer)

	timeoutCB := func(requestID uint64, elapsed time.Duration) {
		// TODO : Reporter failed
	}

	successCB := func(resp *msg.MessageEnvelope, elapsed time.Duration) {
		if resp.Constructor == msg.C_MessagesSent {
			x := new(msg.MessagesSent)
			x.Unmarshal(resp.Message)

			// TODO : Complete Scenario

		} else {
			// TODO : Reporter failed
		}
	}

	return reqEnv, successCB, timeoutCB
}
