package call

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/registry"
	"go.uber.org/zap"
	"sync"
)

/*
   Creation Time: 2021 - May - 19
   Created by:  (Hamidrezakk)
   Maintainers:
      1.  Hamidrezakk
   Auditor: Hamidrezakk
   Copyright Ronak Software Group 2021
*/

type teamInput struct {
	teamID     int64
	teamAccess uint64
}

func (c *call) apiInit(peer *msg.InputPeer, callID int64) (res *msg.PhoneInit, err error) {
	req := msg.PhoneInitCall{
		Peer:   peer,
		CallID: callID,
	}
	reqBytes, err := req.Marshal()
	if err != nil {
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	// Timeout Callback
	timeoutCallback := func() {
		err = domain.ErrRequestTimeout
		wg.Done()
	}

	// Success Callback
	successCallback := func(x *rony.MessageEnvelope) {
		defer wg.Done()
		switch x.Constructor {
		case msg.C_PhoneInit:
			xx := &msg.PhoneInit{}
			innerErr := xx.Unmarshal(x.Message)
			if innerErr != nil {
				err = innerErr
			} else {
				res = xx
			}
		default:
			logs.Debug("exception", zap.String("C", registry.ConstructorName(x.Constructor)))
			err = domain.ErrRequestTimeout
		}
	}

	c.executeRemoteCommand(msg.C_PhoneInitCall, reqBytes, timeoutCallback, successCallback, true)
	wg.Wait()
	return
}

func (c *call) apiRequest(peer *msg.InputPeer, randomID int64, initiator bool, participants []*msg.PhoneParticipantSDP, callID int64, batch bool) (res *msg.PhoneCall, err error) {
	req := msg.PhoneRequestCall{
		Peer:         peer,
		RandomID:     randomID,
		Initiator:    initiator,
		Participants: participants,
		CallID:       callID,
		DeviceType:   c.deviceType,
	}

	reqBytes, err := req.Marshal()
	if err != nil {
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	// Timeout Callback
	timeoutCallback := func() {
		err = domain.ErrRequestTimeout
		wg.Done()
	}

	// Success Callback
	successCallback := func(x *rony.MessageEnvelope) {
		defer wg.Done()
		switch x.Constructor {
		case msg.C_PhoneCall:
			xx := &msg.PhoneCall{}
			innerErr := xx.Unmarshal(x.Message)
			if innerErr != nil {
				err = innerErr
			} else {
				res = xx
			}
		default:
			logs.Debug("exception", zap.String("C", registry.ConstructorName(x.Constructor)))
			err = domain.ErrRequestTimeout
		}
	}

	c.executeRemoteCommand(msg.C_PhoneRequestCall, reqBytes, timeoutCallback, successCallback, !batch)
	wg.Wait()
	return
}

func (c *call) apiAccept(peer *msg.InputPeer, callID int64, participants []*msg.PhoneParticipantSDP) (res *msg.PhoneCall, err error) {
	req := msg.PhoneAcceptCall{
		Peer:         peer,
		Participants: participants,
		CallID:       callID,
		DeviceType:   c.deviceType,
	}

	reqBytes, err := req.Marshal()
	if err != nil {
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	// Timeout Callback
	timeoutCallback := func() {
		err = domain.ErrRequestTimeout
		wg.Done()
	}

	// Success Callback
	successCallback := func(x *rony.MessageEnvelope) {
		defer wg.Done()
		switch x.Constructor {
		case msg.C_PhoneCall:
			xx := &msg.PhoneCall{}
			innerErr := xx.Unmarshal(x.Message)
			if innerErr != nil {
				err = innerErr
			} else {
				res = xx
			}
		default:
			logs.Debug("exception", zap.String("C", registry.ConstructorName(x.Constructor)))
			err = domain.ErrRequestTimeout
		}
	}

	c.executeRemoteCommand(msg.C_PhoneAcceptCall, reqBytes, timeoutCallback, successCallback, true)
	wg.Wait()
	return
}

func (c *call) apiReject(peer *msg.InputPeer, callID int64, reason msg.DiscardReason, duration int32) (res *msg.Bool, err error) {
	req := msg.PhoneDiscardCall{
		Peer:     peer,
		CallID:   callID,
		Duration: duration,
		Reason:   reason,
	}

	reqBytes, err := req.Marshal()
	if err != nil {
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	// Timeout Callback
	timeoutCallback := func() {
		err = domain.ErrRequestTimeout
		wg.Done()
	}

	// Success Callback
	successCallback := func(x *rony.MessageEnvelope) {
		defer wg.Done()
		switch x.Constructor {
		case msg.C_Bool:
			xx := &msg.Bool{}
			innerErr := xx.Unmarshal(x.Message)
			if innerErr != nil {
				err = innerErr
			} else {
				res = xx
			}
		default:
			logs.Debug("exception", zap.String("C", registry.ConstructorName(x.Constructor)))
			err = domain.ErrRequestTimeout
		}
	}

	c.executeRemoteCommand(msg.C_PhoneDiscardCall, reqBytes, timeoutCallback, successCallback, true)
	wg.Wait()
	return
}

func (c *call) apiJoin(peer *msg.InputPeer, callID int64) (res *msg.PhoneParticipants, err error) {
	req := msg.PhoneJoinCall{
		Peer:   peer,
		CallID: callID,
	}

	reqBytes, err := req.Marshal()
	if err != nil {
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	// Timeout Callback
	timeoutCallback := func() {
		err = domain.ErrRequestTimeout
		wg.Done()
	}

	// Success Callback
	successCallback := func(x *rony.MessageEnvelope) {
		defer wg.Done()
		switch x.Constructor {
		case msg.C_PhoneParticipants:
			xx := &msg.PhoneParticipants{}
			innerErr := xx.Unmarshal(x.Message)
			if innerErr != nil {
				err = innerErr
			} else {
				res = xx
			}
		default:
			logs.Debug("exception", zap.String("C", registry.ConstructorName(x.Constructor)))
			err = domain.ErrRequestTimeout
		}
	}

	c.executeRemoteCommand(msg.C_PhoneJoinCall, reqBytes, timeoutCallback, successCallback, true)
	wg.Wait()
	return
}

func (c *call) apiAddParticipant(peer *msg.InputPeer, callID int64, participants []*msg.InputUser) (res *msg.PhoneParticipants, err error) {
	req := msg.PhoneAddParticipant{
		Peer:         peer,
		CallID:       callID,
		Participants: participants,
	}

	reqBytes, err := req.Marshal()
	if err != nil {
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	// Timeout Callback
	timeoutCallback := func() {
		err = domain.ErrRequestTimeout
		wg.Done()
	}

	// Success Callback
	successCallback := func(x *rony.MessageEnvelope) {
		defer wg.Done()
		switch x.Constructor {
		case msg.C_PhoneParticipants:
			xx := &msg.PhoneParticipants{}
			innerErr := xx.Unmarshal(x.Message)
			if innerErr != nil {
				err = innerErr
			} else {
				res = xx
			}
		default:
			logs.Debug("exception", zap.String("C", registry.ConstructorName(x.Constructor)))
			err = domain.ErrRequestTimeout
		}
	}

	c.executeRemoteCommand(msg.C_PhoneAddParticipant, reqBytes, timeoutCallback, successCallback, true)
	wg.Wait()
	return
}

func (c *call) apiRemoveParticipant(peer *msg.InputPeer, callID int64, participants []*msg.InputUser, isTimout bool) (res *msg.Bool, err error) {
	req := msg.PhoneRemoveParticipant{
		Peer:         peer,
		CallID:       callID,
		Participants: participants,
		Timeout:      isTimout,
	}

	reqBytes, err := req.Marshal()
	if err != nil {
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	// Timeout Callback
	timeoutCallback := func() {
		err = domain.ErrRequestTimeout
		wg.Done()
	}

	// Success Callback
	successCallback := func(x *rony.MessageEnvelope) {
		defer wg.Done()
		switch x.Constructor {
		case msg.C_Bool:
			xx := &msg.Bool{}
			innerErr := xx.Unmarshal(x.Message)
			if innerErr != nil {
				err = innerErr
			} else {
				res = xx
			}
		default:
			logs.Debug("exception", zap.String("C", registry.ConstructorName(x.Constructor)))
			err = domain.ErrRequestTimeout
		}
	}

	c.executeRemoteCommand(msg.C_PhoneRemoveParticipant, reqBytes, timeoutCallback, successCallback, true)
	wg.Wait()
	return
}

func (c *call) apiGetParticipant(peer *msg.InputPeer, callID int64) (res *msg.PhoneParticipants, err error) {
	req := msg.PhoneGetParticipants{
		Peer:   peer,
		CallID: callID,
	}

	reqBytes, err := req.Marshal()
	if err != nil {
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	// Timeout Callback
	timeoutCallback := func() {
		err = domain.ErrRequestTimeout
		wg.Done()
	}

	// Success Callback
	successCallback := func(x *rony.MessageEnvelope) {
		defer wg.Done()
		switch x.Constructor {
		case msg.C_PhoneParticipants:
			xx := &msg.PhoneParticipants{}
			innerErr := xx.Unmarshal(x.Message)
			if innerErr != nil {
				err = innerErr
			} else {
				res = xx
			}
		default:
			logs.Debug("exception", zap.String("C", registry.ConstructorName(x.Constructor)))
			err = domain.ErrRequestTimeout
		}
	}

	c.executeRemoteCommand(msg.C_PhoneGetParticipants, reqBytes, timeoutCallback, successCallback, true)
	wg.Wait()
	return
}

func (c *call) apiUpdateAdmin(peer *msg.InputPeer, callID int64, inputUser *msg.InputUser, admin bool) (res *msg.Bool, err error) {
	req := msg.PhoneUpdateAdmin{
		Peer:   peer,
		CallID: callID,
		User:   inputUser,
		Admin:  admin,
	}

	reqBytes, err := req.Marshal()
	if err != nil {
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	// Timeout Callback
	timeoutCallback := func() {
		err = domain.ErrRequestTimeout
		wg.Done()
	}

	// Success Callback
	successCallback := func(x *rony.MessageEnvelope) {
		defer wg.Done()
		switch x.Constructor {
		case msg.C_Bool:
			xx := &msg.Bool{}
			innerErr := xx.Unmarshal(x.Message)
			if innerErr != nil {
				err = innerErr
			} else {
				res = xx
			}
		default:
			logs.Debug("exception", zap.String("C", registry.ConstructorName(x.Constructor)))
			err = domain.ErrRequestTimeout
		}
	}

	c.executeRemoteCommand(msg.C_PhoneUpdateAdmin, reqBytes, timeoutCallback, successCallback, true)
	wg.Wait()
	return
}

func (c *call) apiSendUpdate(peer *msg.InputPeer, callID int64, participants []*msg.InputUser, action msg.PhoneCallAction, actionData []byte, instant bool) (res *msg.Bool, err error) {
	req := msg.PhoneUpdateCall{
		Peer:         peer,
		CallID:       callID,
		Participants: participants,
		Action:       action,
		ActionData:   actionData,
	}

	reqBytes, err := req.Marshal()
	if err != nil {
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	// Timeout Callback
	timeoutCallback := func() {
		err = domain.ErrRequestTimeout
		wg.Done()
	}

	// Success Callback
	successCallback := func(x *rony.MessageEnvelope) {
		defer wg.Done()
		switch x.Constructor {
		case msg.C_Bool:
			xx := &msg.Bool{}
			innerErr := xx.Unmarshal(x.Message)
			if innerErr != nil {
				err = innerErr
			} else {
				res = xx
			}
		default:
			logs.Debug("exception", zap.String("C", registry.ConstructorName(x.Constructor)))
			err = domain.ErrRequestTimeout
		}
	}

	c.executeRemoteCommand(msg.C_PhoneUpdateCall, reqBytes, timeoutCallback, successCallback, instant)
	wg.Wait()
	return
}

func (c *call) setTeamInput(teamId int64, teamAccess uint64) {
	c.teamInput = teamInput{
		teamID:     teamId,
		teamAccess: teamAccess,
	}
}

func (c *call) executeRemoteCommand(
	constructor int64, commandBytes []byte,
	timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler,
	instant bool) {
	logs.Debug("Execute command",
		zap.String("C", registry.ConstructorName(constructor)),
	)

	requestID := uint64(domain.SequentialUniqueID())
	teamID := c.teamInput.teamID
	teamAccess := c.teamInput.teamAccess
	//if c.tempTeamInput != nil {
	//	teamID = c.tempTeamInput.teamID
	//	teamAccess = c.tempTeamInput.teamAccess
	//	c.tempTeamInput = nil
	//}

	header := domain.TeamHeader(teamID, teamAccess)
	// If the constructor is a realtime command, then just send it to the server
	if instant {
		c.SDK().NetCtrl().WebsocketCommand(&rony.MessageEnvelope{
			Header:      header,
			Constructor: constructor,
			RequestID:   requestID,
			Message:     commandBytes,
		}, timeoutCB, successCB, true, true)
	} else {
		c.SDK().QueueCtrl().EnqueueCommand(
			&rony.MessageEnvelope{
				Header:      header,
				Constructor: constructor,
				RequestID:   requestID,
				Message:     commandBytes,
			},
			timeoutCB, successCB, true,
		)
	}
}
