package callCtrl

import (
	"git.ronaksoft.com/river/msg/go/msg"
	networkCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_network"
	queueCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_queue"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/registry"
	"go.uber.org/zap"
	"sync"
)

type teamInput struct {
	teamID     int64
	teamAccess uint64
}

type CallAPI interface {
	Init(peer *msg.InputPeer, callID int64) (res *msg.PhoneInit, err error)
	Request(peer *msg.InputPeer, randomID int64, initiator bool, participants []*msg.PhoneParticipantSDP, callID int64, batch bool) (res *msg.PhoneCall, err error)
	Accept(peer *msg.InputPeer, callID int64, participants []*msg.PhoneParticipantSDP) (res *msg.PhoneCall, err error)
	Reject(peer *msg.InputPeer, callID int64, reason msg.DiscardReason, duration int32) (res *msg.Bool, err error)
	Join(peer *msg.InputPeer, callID int64) (res *msg.PhoneParticipants, err error)
	AddParticipant(peer *msg.InputPeer, callID int64, participants []*msg.InputUser) (res *msg.PhoneParticipants, err error)
	RemoveParticipant(peer *msg.InputPeer, callID int64, participants []*msg.InputUser, isTimout bool) (res *msg.Bool, err error)
	GetParticipant(peer *msg.InputPeer, callID int64) (res *msg.PhoneParticipants, err error)
	UpdateAdmin(peer *msg.InputPeer, callID int64, inputUser *msg.InputUser, admin bool) (res *msg.Bool, err error)
	SendUpdate(peer *msg.InputPeer, callID int64, participants []*msg.InputUser, action msg.PhoneCallAction, actionData []byte, instant bool) (res *msg.Bool, err error)

	SetTeamInput(teamId int64, teamAccess uint64)
	SetTempTeamInput(teamId int64, teamAccess uint64)
}

func NewCallAPI() CallAPI {
	return &callAPI{
		networkCtrl: nil,
		queueCtrl:   nil,
		teamInput: teamInput{
			teamID:     domain.GetCurrTeamID(),
			teamAccess: domain.GetCurrTeamAccess(),
		},
		deviceType:    msg.CallDeviceType_CallDeviceUnknown,
		tempTeamInput: nil,
	}
}

type callAPI struct {
	networkCtrl *networkCtrl.Controller
	queueCtrl   *queueCtrl.Controller

	teamInput  teamInput
	deviceType msg.CallDeviceType

	tempTeamInput *teamInput
}

func (c *callAPI) Init(peer *msg.InputPeer, callID int64) (res *msg.PhoneInit, err error) {
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

func (c *callAPI) Request(peer *msg.InputPeer, randomID int64, initiator bool, participants []*msg.PhoneParticipantSDP, callID int64, batch bool) (res *msg.PhoneCall, err error) {
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

func (c *callAPI) Accept(peer *msg.InputPeer, callID int64, participants []*msg.PhoneParticipantSDP) (res *msg.PhoneCall, err error) {
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

func (c *callAPI) Reject(peer *msg.InputPeer, callID int64, reason msg.DiscardReason, duration int32) (res *msg.Bool, err error) {
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

func (c *callAPI) Join(peer *msg.InputPeer, callID int64) (res *msg.PhoneParticipants, err error) {
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

func (c *callAPI) AddParticipant(peer *msg.InputPeer, callID int64, participants []*msg.InputUser) (res *msg.PhoneParticipants, err error) {
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

func (c *callAPI) RemoveParticipant(peer *msg.InputPeer, callID int64, participants []*msg.InputUser, isTimout bool) (res *msg.Bool, err error) {
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

func (c *callAPI) GetParticipant(peer *msg.InputPeer, callID int64) (res *msg.PhoneParticipants, err error) {
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

func (c *callAPI) UpdateAdmin(peer *msg.InputPeer, callID int64, inputUser *msg.InputUser, admin bool) (res *msg.Bool, err error) {
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

func (c *callAPI) SendUpdate(peer *msg.InputPeer, callID int64, participants []*msg.InputUser, action msg.PhoneCallAction, actionData []byte, instant bool) (res *msg.Bool, err error) {
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

func (c *callAPI) SetTeamInput(teamId int64, teamAccess uint64) {
	c.teamInput = teamInput{
		teamID:     teamId,
		teamAccess: teamAccess,
	}
}

func (c *callAPI) SetTempTeamInput(teamId int64, teamAccess uint64) {
	c.tempTeamInput = &teamInput{
		teamID:     teamId,
		teamAccess: teamAccess,
	}
}

func (c *callAPI) executeRemoteCommand(
	constructor int64, commandBytes []byte,
	timeoutCB domain.TimeoutCallback, successCB domain.MessageHandler,
	instant bool) {
	logs.Debug("Execute command",
		zap.String("C", registry.ConstructorName(constructor)),
	)

	requestID := uint64(domain.SequentialUniqueID())
	teamID := c.teamInput.teamID
	teamAccess := c.teamInput.teamAccess
	if c.tempTeamInput != nil {
		teamID = c.tempTeamInput.teamID
		teamAccess = c.tempTeamInput.teamAccess
		c.tempTeamInput = nil
	}

	// If the constructor is a realtime command, then just send it to the server
	if instant {
		c.networkCtrl.WebsocketCommand(&rony.MessageEnvelope{
			Header:      domain.TeamHeader(teamID, teamAccess),
			Constructor: constructor,
			RequestID:   requestID,
			Message:     commandBytes,
		}, timeoutCB, successCB, true, true)
	} else {
		c.queueCtrl.EnqueueCommand(
			&rony.MessageEnvelope{
				Header:      domain.TeamHeader(teamID, teamAccess),
				Constructor: constructor,
				RequestID:   requestID,
				Message:     commandBytes,
			},
			timeoutCB, successCB, true,
		)
	}
}
