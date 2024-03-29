package call

import (
    "sync"
    "time"

    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/river-sdk/internal/domain"
    "github.com/ronaksoft/river-sdk/internal/request"
    "github.com/ronaksoft/rony"
    "github.com/ronaksoft/rony/registry"
    "go.uber.org/zap"
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

    // Timeout Callback
    timeoutCallback := func() {
        err = domain.ErrRequestTimeout
    }

    // Success Callback
    successCallback := func(x *rony.MessageEnvelope) {
        switch x.Constructor {
        case msg.C_PhoneInit:
            xx := &msg.PhoneInit{}
            innerErr := xx.Unmarshal(x.Message)
            if innerErr != nil {
                err = innerErr
            } else {
                res = xx
            }
        case rony.C_Error:
            xx := &rony.Error{}
            innerErr := xx.Unmarshal(x.Message)
            if innerErr == nil {
                c.Log().Warn("got error on server request PhoneInitCall", zap.Error(err))
            }
            if xx.Code == "E02" && xx.Items == "CALL" {
                err = domain.ErrInvalidCall
            } else {
                err = domain.ErrInvalidData
            }
        default:
            c.Log().Warn("received unknown response for PhoneInitCall", zap.String("C", registry.ConstructorName(x.Constructor)))
            err = domain.ErrInvalidData
        }
    }

    c.executeRemoteCommand(msg.C_PhoneInitCall, reqBytes, timeoutCallback, successCallback, true, callID)

    return
}

func (c *call) apiRequest(
        peer *msg.InputPeer, randomID int64, initiator bool, participants []*msg.PhoneParticipantSDP, callID int64, video, batch bool,
) (res *msg.PhoneCall, err error) {
    req := msg.PhoneRequestCall{
        Peer:         peer,
        RandomID:     randomID,
        Initiator:    initiator,
        Participants: participants,
        CallID:       callID,
        DeviceType:   c.deviceType,
        Video:        video,
    }

    reqBytes, err := req.Marshal()
    if err != nil {
        return
    }

    // Timeout Callback
    timeoutCallback := func() {
        err = domain.ErrRequestTimeout
    }

    // Success Callback
    successCallback := func(x *rony.MessageEnvelope) {
        switch x.Constructor {
        case msg.C_PhoneCall:
            xx := &msg.PhoneCall{}
            innerErr := xx.Unmarshal(x.Message)
            if innerErr != nil {
                err = innerErr
            } else {
                res = xx
            }
        case rony.C_Error:
            xx := &rony.Error{}
            innerErr := xx.Unmarshal(x.Message)
            if innerErr == nil {
                c.Log().Warn("got error on server request PhoneRequestCall", zap.Error(err))
            }
            err = xx
        default:
            c.Log().Warn("received unknown response for PhoneRequestCall", zap.String("C", registry.ConstructorName(x.Constructor)))
            err = domain.ErrInvalidData
        }
    }

    c.executeRemoteCommand(msg.C_PhoneRequestCall, reqBytes, timeoutCallback, successCallback, !batch, callID)
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

    // Timeout Callback
    timeoutCallback := func() {
        err = domain.ErrRequestTimeout
    }

    // Success Callback
    successCallback := func(x *rony.MessageEnvelope) {
        switch x.Constructor {
        case msg.C_PhoneCall:
            xx := &msg.PhoneCall{}
            innerErr := xx.Unmarshal(x.Message)
            if innerErr != nil {
                err = innerErr
            } else {
                res = xx
            }
        case rony.C_Error:
            xx := &rony.Error{}
            innerErr := xx.Unmarshal(x.Message)
            if innerErr == nil {
                c.Log().Warn("got error on server request PhoneAcceptCall", zap.Error(err))
            }
            err = xx
        default:
            c.Log().Warn("received unknown response for PhoneAcceptCall", zap.String("C", registry.ConstructorName(x.Constructor)))
            err = domain.ErrInvalidData
        }
    }

    c.executeRemoteCommand(msg.C_PhoneAcceptCall, reqBytes, timeoutCallback, successCallback, true, callID)
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

    // Timeout Callback
    timeoutCallback := func() {
        err = domain.ErrRequestTimeout
    }

    // Success Callback
    successCallback := func(x *rony.MessageEnvelope) {
        switch x.Constructor {
        case msg.C_Bool:
            xx := &msg.Bool{}
            innerErr := xx.Unmarshal(x.Message)
            if innerErr != nil {
                err = innerErr
            } else {
                res = xx
            }
        case rony.C_Error:
            xx := &rony.Error{}
            innerErr := xx.Unmarshal(x.Message)
            if innerErr == nil {
                c.Log().Warn("got error on server request PhoneDiscardCall", zap.Error(err))
            }
            err = xx
        default:
            c.Log().Warn("received unknown response for PhoneDiscardCall", zap.String("C", registry.ConstructorName(x.Constructor)))
            err = domain.ErrInvalidData
        }
    }

    c.executeRemoteCommand(msg.C_PhoneDiscardCall, reqBytes, timeoutCallback, successCallback, true, callID)
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

    // Timeout Callback
    timeoutCallback := func() {
        err = domain.ErrRequestTimeout
    }

    // Success Callback
    successCallback := func(x *rony.MessageEnvelope) {
        switch x.Constructor {
        case msg.C_PhoneParticipants:
            xx := &msg.PhoneParticipants{}
            innerErr := xx.Unmarshal(x.Message)
            if innerErr != nil {
                err = innerErr
            } else {
                res = xx
            }
        case rony.C_Error:
            xx := &rony.Error{}
            innerErr := xx.Unmarshal(x.Message)
            if innerErr == nil {
                c.Log().Warn("got error on server request PhoneJoinCall", zap.Error(err))
            }
            err = xx
        default:
            c.Log().Warn("received unknown response for PhoneJoinCall", zap.String("C", registry.ConstructorName(x.Constructor)))
            err = domain.ErrInvalidData
        }
    }

    c.executeRemoteCommand(msg.C_PhoneJoinCall, reqBytes, timeoutCallback, successCallback, true, callID)
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

    // Timeout Callback
    timeoutCallback := func() {
        err = domain.ErrRequestTimeout
    }

    // Success Callback
    successCallback := func(x *rony.MessageEnvelope) {
        switch x.Constructor {
        case msg.C_PhoneParticipants:
            xx := &msg.PhoneParticipants{}
            innerErr := xx.Unmarshal(x.Message)
            if innerErr != nil {
                err = innerErr
            } else {
                res = xx
            }
        case rony.C_Error:
            xx := &rony.Error{}
            innerErr := xx.Unmarshal(x.Message)
            if innerErr == nil {
                c.Log().Warn("got error on server request PhoneAddParticipant", zap.Error(err))
            }
            err = xx
        default:
            c.Log().Warn("received unknown response for PhoneAddParticipant", zap.String("C", registry.ConstructorName(x.Constructor)))
            err = domain.ErrInvalidData
        }
    }

    c.executeRemoteCommand(msg.C_PhoneAddParticipant, reqBytes, timeoutCallback, successCallback, true, callID)
    return
}

func (c *call) apiRemoveParticipant(peer *msg.InputPeer, callID int64, participants []*msg.InputUser, isTimout bool) (res *msg.Bool,
        err error,
) {
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

    // Timeout Callback
    timeoutCallback := func() {
        err = domain.ErrRequestTimeout
    }

    // Success Callback
    successCallback := func(x *rony.MessageEnvelope) {
        switch x.Constructor {
        case msg.C_Bool:
            xx := &msg.Bool{}
            innerErr := xx.Unmarshal(x.Message)
            if innerErr != nil {
                err = innerErr
            } else {
                res = xx
            }
        case rony.C_Error:
            xx := &rony.Error{}
            innerErr := xx.Unmarshal(x.Message)
            if innerErr == nil {
                c.Log().Warn("got error on server request PhoneRemoveParticipant", zap.Error(err))
            }
            err = xx
        default:
            c.Log().Warn("received unknown response for PhoneRemoveParticipants", zap.String("C", registry.ConstructorName(x.Constructor)))
            err = domain.ErrInvalidData
        }
    }

    c.executeRemoteCommand(msg.C_PhoneRemoveParticipant, reqBytes, timeoutCallback, successCallback, true, callID)
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

    // Timeout Callback
    timeoutCallback := func() {
        err = domain.ErrRequestTimeout
    }

    // Success Callback
    successCallback := func(x *rony.MessageEnvelope) {
        switch x.Constructor {
        case msg.C_PhoneParticipants:
            xx := &msg.PhoneParticipants{}
            innerErr := xx.Unmarshal(x.Message)
            if innerErr != nil {
                err = innerErr
            } else {
                res = xx
            }
        case rony.C_Error:
            xx := &rony.Error{}
            innerErr := xx.Unmarshal(x.Message)
            if innerErr == nil {
                c.Log().Warn("got error on server request PhoneGetParticipants", zap.Error(err))
            }
            err = xx
        default:
            c.Log().Warn("received unknown response for PhoneGetParticipants", zap.String("C", registry.ConstructorName(x.Constructor)))
            err = domain.ErrInvalidData
        }
    }

    c.executeRemoteCommand(msg.C_PhoneGetParticipants, reqBytes, timeoutCallback, successCallback, true, callID)

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

    // Timeout Callback
    timeoutCallback := func() {
        err = domain.ErrRequestTimeout
    }

    // Success Callback
    successCallback := func(x *rony.MessageEnvelope) {
        switch x.Constructor {
        case msg.C_Bool:
            xx := &msg.Bool{}
            innerErr := xx.Unmarshal(x.Message)
            if innerErr != nil {
                err = innerErr
            } else {
                res = xx
            }
        case rony.C_Error:
            xx := &rony.Error{}
            innerErr := xx.Unmarshal(x.Message)
            if innerErr == nil {
                c.Log().Warn("got error on server request PhoneUpdateAdmin", zap.Error(err))
            }
            err = xx
        default:
            c.Log().Warn("received unknown response for PhoneUpdateAdmin", zap.String("C", registry.ConstructorName(x.Constructor)))
            err = domain.ErrInvalidData
        }
    }

    c.executeRemoteCommand(msg.C_PhoneUpdateAdmin, reqBytes, timeoutCallback, successCallback, true, callID)

    return
}

func (c *call) apiSendUpdate(
        peer *msg.InputPeer, callID int64, participants []*msg.InputUser, action msg.PhoneCallAction, actionData []byte, instant bool,
) (res *msg.Bool, err error) {
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

    // Timeout Callback
    timeoutCallback := func() {
        err = domain.ErrRequestTimeout
    }

    // Success Callback
    successCallback := func(x *rony.MessageEnvelope) {
        switch x.Constructor {
        case msg.C_Bool:
            xx := &msg.Bool{}
            innerErr := xx.Unmarshal(x.Message)
            if innerErr != nil {
                err = innerErr
            } else {
                res = xx
            }
        case rony.C_Error:
            xx := &rony.Error{}
            innerErr := xx.Unmarshal(x.Message)
            if innerErr == nil {
                c.Log().Warn("got error on server request PhoneUpdateCall", zap.Error(err))
            }
            err = xx
        default:
            c.Log().Warn("received unknown response for PhoneUpdateCall", zap.String("C", registry.ConstructorName(x.Constructor)))
            err = domain.ErrInvalidData
        }
    }

    c.executeRemoteCommand(msg.C_PhoneUpdateCall, reqBytes, timeoutCallback, successCallback, instant, callID)
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
        instant bool, callID int64,
) {
    c.Log().Debug("Execute command",
        zap.String("C", registry.ConstructorName(constructor)),
        zap.Bool("Instant", instant),
        zap.Int64("CallID", callID),
    )

    rdt := request.Realtime
    if instant {
        rdt |= request.SkipFlusher
    } else {
        rdt |= request.Batch
    }

    wg := sync.WaitGroup{}

    retry := 0
    var innerTimeoutCB domain.TimeoutCallback
    var innerSuccessCB domain.MessageHandler
    var executeFn func()

    innerTimeoutCB = func() {
        if retry < 1 {
            go func() {
                time.Sleep(time.Duration(1) * time.Second)
                executeFn()
            }()
        } else {
            timeoutCB()
            wg.Done()
        }
    }

    innerSuccessCB = func(m *rony.MessageEnvelope) {
        successCB(m.Clone())
        c.removeCallRequestID(callID, int64(m.RequestID))
        wg.Done()
    }

    executeFn = func() {
        reqID := domain.NextRequestID()
        cb := request.NewCallbackFromBytes(
            c.teamInput.teamID,
            c.teamInput.teamAccess,
            reqID, constructor, commandBytes, innerTimeoutCB, innerSuccessCB, nil, false,
            rdt, 10*time.Second,
        )

        retry++
        err := c.SDK().Execute(cb)
        if err == nil {
            c.appendCallRequestID(callID, int64(reqID))
        }
    }

    wg.Add(1)
    executeFn()
    wg.Wait()
}
