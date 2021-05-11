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
)

const (
	RetryInterval    = 10000
	RetryLimit       = 6
	ReconnectTry     = 3
	ReconnectTimeout = 15000

	TempCallID = int64(-27001)
)

type call struct {
	module.Base

	peerConnections map[int32]*msg.CallConnection
	peer            *msg.InputPeer
	activeCallID    int64
	callInfo        map[int64]*Info
	iceServer       []*msg.IceServer
	userID          int64

	teamInput  teamInput
	deviceType msg.CallDeviceType

	callback *Callback
}

func New(callback *Callback) *call {
	r := &call{
		peerConnections: nil,
		peer:            nil,
		activeCallID:    0,
		callInfo:        make(map[int64]*Info),
		iceServer:       nil,
		userID:          0,
		teamInput: teamInput{
			teamID:     domain.GetCurrTeamID(),
			teamAccess: domain.GetCurrTeamAccess(),
		},
		deviceType: msg.CallDeviceType_CallDeviceUnknown,
		callback:   callback,
	}

	r.RegisterUpdateAppliers(map[int64]domain.UpdateApplier{
		msg.C_UpdatePhoneCall: r.updatePhoneCall,
	})

	return r
}

func (c *call) Name() string {
	return module.Call
}
