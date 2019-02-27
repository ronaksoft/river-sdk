package scenario

import (
	"time"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"
	"git.ronaksoftware.com/ronak/riversdk/msg"
)

type SendMessageMany struct {
	Scenario
	startNo int64
	endNo   int64
}

// NewSendMessageMany create new instance
func NewSendMessageMany(isFinal bool, startNo, endNo int64) shared.Screenwriter {
	s := new(SendMessageMany)
	s.isFinal = isFinal
	s.startNo = startNo
	s.endNo = endNo
	return s
}

// Play execute SendMessageMany scenario
func (s *SendMessageMany) Play(act shared.Acter) {
	if act.GetAuthID() == 0 {
		s.AddJobs(1)
		success := Play(act, NewCreateAuthKey(false))
		if !success {
			s.failed(act, 0, 0, "Play() : failed at pre requested scenario CreateAuthKey")
			return
		}
		s.wait.Done()
	}
	if act.GetUserID() == 0 {
		s.AddJobs(1)
		success := Play(act, NewLogin(false))
		if !success {
			s.failed(act, 0, 0, "Play() : failed at pre requested scenario Login")
			return
		}
		s.wait.Done()
	}
	if len(act.GetPeers()) == 0 {
		s.AddJobs(1)
		success := Play(act, NewImportContact(false))
		if !success {
			s.failed(act, 0, 0, "Play() : failed at pre requested scenario ImportContact")
			return
		}
		s.wait.Done()
	}

	act.SetUpdateApplier(msg.C_UpdateNewMessage, s.onReceiveUpdateNewMessage)
	for n := s.startNo; n <= s.endNo; n++ {
		phoneNo := shared.GetPhone(n)
		toActor, ok := shared.GetCachedActorByPhone(phoneNo)
		if ok {
			s.AddJobs(1)
			p := GetPeerInfo(act.GetUserID(), toActor.GetUserID(), msg.PeerUser)
			act.ExecuteRequest(s.messageSend(act, p))
		}
	}
}

// messageSend : Step 1
func (s *SendMessageMany) messageSend(act shared.Acter, peer *shared.PeerInfo) (*msg.MessageEnvelope, shared.SuccessCallback, shared.TimeoutCallback) {
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
			x.Unmarshal(resp.Message)

			// TODO : Complete Scenario
			s.completed(act, elapsed, resp.RequestID, "messageSend() Success")
		} else {
			// TODO : Reporter failed
			s.failed(act, elapsed, resp.RequestID, "messageSend() SuccessCB response is not MessagesSent, Constructor :"+msg.ConstructorNames[resp.Constructor])
		}
	}

	return reqEnv, successCB, timeoutCB
}

func (s *SendMessageMany) onReceiveUpdateNewMessage(act shared.Acter, u *msg.UpdateEnvelope) {
	x := new(msg.UpdateNewMessage)
	_ = x.Unmarshal(u.Update)
	peer := GetPeerInfo(act.GetUserID(), x.Message.SenderID, msg.PeerType(x.Message.PeerType))
	reqEnv := MessageSend(peer)

	timeoutCB := func(requestID uint64, elapsed time.Duration) {
		// Do nothing
	}
	successCB := func(resp *msg.MessageEnvelope, elapsed time.Duration) {
		// Do nothing
	}

	act.ExecuteRequest(reqEnv, successCB, timeoutCB)
}
