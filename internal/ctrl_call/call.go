/*
   Creation Time: 2021 - April - 04
   Created by:  (hamidrezakk)
   Maintainers:
      1.  HamidrezaKK (hamidrezakks@gmail.com)
   Auditor: HamidrezaKK
   Copyright Ronak Software Group 2021
*/

package callCtrl

import (
	"fmt"
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"go.uber.org/zap"
)

const (
	RetryInterval = 10000
	RetryLimit    = 6
)

type CallController interface {
	ParseUpdate(update *msg.UpdateEnvelope)
}

func NewCallController() CallController {
	return &callController{
		peerConnections: nil,
		peer:            nil,
		activeCallID:    0,
		callInfo:        nil,
		callAPI:         nil,
	}
}

type callController struct {
	peerConnections map[int64]CallConnection
	peer            *msg.InputPeer
	activeCallID    int64
	callInfo        map[int64]CallInfo

	callAPI *CallAPI
}

func (c *callController) ParseUpdate(update *msg.UpdateEnvelope) {
	go func() {
		if update.Constructor != msg.C_UpdatePhoneCall {
			return
		}
		x := &msg.UpdatePhoneCall{}
		err := x.Unmarshal(update.Update)
		if err != nil {
			return
		}

		now := domain.Now().Unix()
		if !(x.Timestamp == 0 || now-x.Timestamp < 60) {
			return
		}

		data, err := parseCallAction(x.Action, x.ActionData)
		if err != nil {
			logs.Debug("parseCallAction", zap.Error(err))
			return
		}

		switch x.Action {
		case msg.PhoneCallAction_PhoneCallRequested:
			c.callRequested(data)
		case msg.PhoneCallAction_PhoneCallAccepted:
			//
		case msg.PhoneCallAction_PhoneCallDiscarded:
			//
		case msg.PhoneCallAction_PhoneCallIceExchange:
			//
		case msg.PhoneCallAction_PhoneCallMediaSettingsChanged:
			//
		case msg.PhoneCallAction_PhoneCallSDPOffer:
			//
		case msg.PhoneCallAction_PhoneCallSDPAnswer:
			//
		case msg.PhoneCallAction_PhoneCallAck:
			//
		case msg.PhoneCallAction_PhoneCallParticipantAdded:
			//
		case msg.PhoneCallAction_PhoneCallParticipantRemoved:
			//
		case msg.PhoneCallAction_PhoneCallAdminUpdated:
			//
		case msg.PhoneCallAction_PhoneCallJoinRequested:
			//
		case msg.PhoneCallAction_PhoneCallScreenShare:
			//
		case msg.PhoneCallAction_PhoneCallPicked:
			//
		case msg.PhoneCallAction_PhoneCallRestarted:
			//
		}
	}()
}

func (c *callController) ToggleVide(enable bool) {

}

func (c *callController) ToggleAudio(enable bool) {

}

func (c *callController) TryReconnect(connId int32) {

}

func (c *callController) CallStart(peer *msg.InputPeer, participants []*msg.InputUser, callID int64) {
	c.peer = peer

}

func (c *callController) callRequested(in interface{}) {
	data := in.(*msg.PhoneActionRequested)
	fmt.Println(data)
}
