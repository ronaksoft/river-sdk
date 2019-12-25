package scenario

import (
	"fmt"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-supernumerary/pkg/shared"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/chat"
)

type CreateGroup struct {
	Scenario
}

// NewSendMessage create new instance
func NewCreateGroup(isFinal bool) shared.Screenwriter {
	s := new(CreateGroup)
	s.isFinal = isFinal
	return s
}

// Play execute SendMessage scenario
func (s *CreateGroup) Play(act shared.Actor) {
	if act.GetAuthID() == 0 {
		s.log(act, "AuthID is ZERO, Scenario Failed", 0, 0)
		return
	}
	if act.GetUserID() == 0 {
		s.log(act, "UserID is ZERO, Scenario Failed", 0, 0)
		return
	}

	if len(act.GetPeers()) == 0 {
		s.AddJobs(1)
		success := Play(act, NewImportContact(false))
		if !success {
			s.failed(act, 0, 0, "Play() : failed at pre requested scenario Create Group")
			return
		}
		s.wait.Done()
	}

	peers := act.GetPeers()
	s.AddJobs(len(peers))
	act.ExecuteRequest(s.createGroup(act, peers))
}

// messageSend : Step 1
func (s *CreateGroup) createGroup(act shared.Actor, peers []*shared.PeerInfo) (*msg.MessageEnvelope, shared.SuccessCallback, shared.TimeoutCallback) {
	reqEnv := GroupsCreate(fmt.Sprintf("Group Created By %s", act.GetPhone()), peers)
	timeoutCB := func(requestID uint64, elapsed time.Duration) {
		// Reporter failed
		act.SetTimeout(msg.C_GroupsCreate, elapsed)
		s.failed(act, elapsed, requestID, "createGroup() Timeout")
	}

	successCB := func(resp *msg.MessageEnvelope, elapsed time.Duration) {
		act.SetSuccess(msg.C_GroupsCreate, elapsed)
		if s.isErrorResponse(act, elapsed, resp, "createGroup()") {
			return
		}
		if resp.Constructor == msg.C_Group {
			x := new(msg.Group)
			_ = x.Unmarshal(resp.Message)
			s.completed(act, elapsed, resp.RequestID, "createGroup() Success")
		} else {
			s.failed(act, elapsed, resp.RequestID, "createGroup() SuccessCB response is not MessagesSent, Constructor :"+msg.ConstructorNames[resp.Constructor])
		}
	}

	return reqEnv, successCB, timeoutCB
}
