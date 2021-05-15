/*
   Creation Time: 2021 - April - 04
   Created by:  (hamidrezakk)
   Maintainers:
      1.  HamidrezaKK (hamidrezakks@gmail.com)
   Auditor: HamidrezaKK
   Copyright Ronak Software Group 2021
*/

package call

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/module"
	"sync"
)

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
	AuthID     int64
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
	iceServer       []*msg.IceServer
	userID          int64
	authID          int64

	teamInput  teamInput
	deviceType msg.CallDeviceType

	callback *Callback
}

func New(config *Config) *call {
	c := &call{
		mu:              nil,
		peerConnections: make(map[int32]*Connection),
		peer:            nil,
		activeCallID:    0,
		callInfo:        make(map[int64]*Info),
		iceServer:       nil,
		userID:          config.UserID,
		authID:          config.AuthID,
		teamInput: teamInput{
			teamID:     config.TeamID,
			teamAccess: config.TeamAccess,
		},
		deviceType: config.DeviceType,
		callback:   config.Callback,
	}

	c.RegisterHandlers(
		map[int64]domain.LocalHandler{
			msg.C_ClientCallToggleVideo:             c.toggleVideoHandler,
			msg.C_ClientCallToggleAudio:             c.toggleAudioHandler,
			msg.C_ClientCallTryReconnect:            c.tryReconnectHandler,
			msg.C_ClientCallDestroy:                 c.destroyHandler,
			msg.C_ClientCallAreAllAudio:             c.areAllAudioHandler,
			msg.C_ClientCallSendIceCandidate:        c.iceCandidateHandler,
			msg.C_ClientCallSendIceConnectionStatus: c.iceConnectionStatusChangeHandler,
			msg.C_ClientCallSendMediaSettings:       c.mediaSettingsChangeHandler,
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
		msg.C_UpdatePhoneCall: c.updatePhoneCall,
	})

	return c
}

func (c *call) Name() string {
	return module.Call
}
