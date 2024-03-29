package call

import (
    "sync"

    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/river-sdk/internal/domain"
    "github.com/ronaksoft/river-sdk/internal/request"
    "github.com/ronaksoft/river-sdk/module"
)

/*
   Creation Time: 2021 - May - 19
   Created by:  (Hamidrezakk)
   Maintainers:
      1.  Hamidrezakk
   Auditor: Hamidrezakk
   Copyright Ronak Software Group 2021
*/

const (
    RetryInterval    = 10
    RetryLimit       = 6
    ReconnectTry     = 3
    ReconnectTimeout = 15

    TempCallID = int64(-27001)
)

type Config struct {
    TeamID     int64
    TeamAccess uint64
    UserID     int64
    DeviceType msg.CallDeviceType
    Callback   *Callback
}

type call struct {
    module.Base

    mu              *sync.RWMutex
    peerConnections map[int32]*Connection
    peer            *msg.InputPeer
    activeCallID    int64
    callInfo        map[int64]*Info
    callDuration    map[int64]*Duration
    rejectedCallIDs []int64

    iceServer []*msg.IceServer
    userID    int64

    teamInput  teamInput
    deviceType msg.CallDeviceType

    callback *Callback
}

func New(config *Config) *call {
    c := &call{
        mu:              &sync.RWMutex{},
        peerConnections: make(map[int32]*Connection),
        peer:            nil,
        activeCallID:    0,
        callInfo:        make(map[int64]*Info),
        callDuration:    make(map[int64]*Duration),
        rejectedCallIDs: nil,
        iceServer:       nil,
        userID:          config.UserID,
        teamInput: teamInput{
            teamID:     config.TeamID,
            teamAccess: config.TeamAccess,
        },
        deviceType: config.DeviceType,
        callback:   config.Callback,
    }

    /*
    	C_ClientCallSendIceCandidate
    	C_ClientCallStart
    	C_ClientCallSendMediaSettings
    	C_ClientCallDestroy
    	C_ClientCallReject
    	C_ClientCallAccept
    */
    c.RegisterHandlers(
        map[int64]request.LocalHandler{
            msg.C_ClientCallToggleVideo:             c.toggleVideoHandler,
            msg.C_ClientCallToggleAudio:             c.toggleAudioHandler,
            msg.C_ClientCallTryReconnect:            c.tryReconnectHandler,
            msg.C_ClientCallDestroy:                 c.destroyHandler,
            msg.C_ClientCallAreAllAudio:             c.areAllAudioHandler,
            msg.C_ClientCallGetDuration:             c.durationHandler,
            msg.C_ClientCallSendIceCandidate:        c.iceCandidateHandler,
            msg.C_ClientCallSendIceConnectionStatus: c.iceConnectionStatusChangeHandler,
            msg.C_ClientCallSendMediaSettings:       c.mediaSettingsChangeHandler,
            msg.C_ClientCallSendTrack:               c.trackUpdateHandler,
            msg.C_ClientCallSendAck:                 c.ackHandler,
            msg.C_ClientCallStart:                   c.startHandler,
            msg.C_ClientCallJoin:                    c.joinHandler,
            msg.C_ClientCallAccept:                  c.acceptHandler,
            msg.C_ClientCallReject:                  c.rejectHandler,
            msg.C_ClientCallGetParticipantByUserID:  c.getParticipantByUserIDHandler,
            msg.C_ClientCallGetParticipantByConnId:  c.getParticipantByConnIdHandler,
            msg.C_ClientCallGetParticipantList:      c.getParticipantListHandler,
            msg.C_ClientCallMuteParticipant:         c.muteParticipantHandler,
            msg.C_ClientCallGroupAddParticipant:     c.groupAddParticipantHandler,
            msg.C_ClientCallGroupRemoveParticipant:  c.groupRemoveParticipantHandler,
            msg.C_ClientCallGroupUpdateAdmin:        c.groupUpdateAdminHandler,
        },
    )

    c.RegisterUpdateAppliers(map[int64]domain.UpdateApplier{
        msg.C_UpdatePhoneCall:        c.updatePhoneCall,
        msg.C_UpdatePhoneCallStarted: c.updatePhoneCallStarted,
        msg.C_UpdatePhoneCallEnded:   c.updatePhoneCallEnded,
    })

    return c
}

func (c *call) Name() string {
    return module.Call
}
