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
			c.callAccepted(data)
		case msg.PhoneCallAction_PhoneCallDiscarded:
			c.callDiscarded(data)
		case msg.PhoneCallAction_PhoneCallIceExchange:
			c.iceExchange(data)
		case msg.PhoneCallAction_PhoneCallMediaSettingsChanged:
			c.mediaSettingsUpdate(data)
		case msg.PhoneCallAction_PhoneCallSDPOffer:
			c.sdpOfferUpdated(data)
		case msg.PhoneCallAction_PhoneCallSDPAnswer:
			c.sdpAnswerUpdated(data)
		case msg.PhoneCallAction_PhoneCallAck:
			c.callAcknowledged(data)
		case msg.PhoneCallAction_PhoneCallParticipantAdded:
			c.participantAdded(data)
		case msg.PhoneCallAction_PhoneCallParticipantRemoved:
			c.participantRemoved(data)
		case msg.PhoneCallAction_PhoneCallAdminUpdated:
			c.adminUpdated(data)
		case msg.PhoneCallAction_PhoneCallJoinRequested:
			c.joinRequested(data)
		case msg.PhoneCallAction_PhoneCallScreenShare:
			c.screenShareUpdated(data)
		case msg.PhoneCallAction_PhoneCallPicked:
			c.callPicked(data)
		case msg.PhoneCallAction_PhoneCallRestarted:
			c.callRestarted(data)
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

func (c *callController) callAccepted(in interface{}) {
	data := in.(*msg.PhoneActionAccepted)
	fmt.Println(data)
}

func (c *callController) callDiscarded(in interface{}) {
	data := in.(*msg.PhoneActionDiscarded)
	fmt.Println(data)
}

func (c *callController) iceExchange(in interface{}) {
	data := in.(*msg.PhoneActionIceExchange)
	fmt.Println(data)
}

func (c *callController) mediaSettingsUpdate(in interface{}) {
	data := in.(*msg.PhoneActionMediaSettingsUpdated)
	fmt.Println(data)
}

func (c *callController) sdpOfferUpdated(in interface{}) {
	data := in.(*msg.PhoneActionSDPOffer)
	fmt.Println(data)
}

func (c *callController) sdpAnswerUpdated(in interface{}) {
	data := in.(*msg.PhoneActionSDPAnswer)
	fmt.Println(data)
}

func (c *callController) callAcknowledged(in interface{}) {
	data := in.(*msg.PhoneActionAck)
	fmt.Println(data)
}

func (c *callController) participantAdded(in interface{}) {
	data := in.(*msg.PhoneActionParticipantAdded)
	fmt.Println(data)
}

func (c *callController) participantRemoved(in interface{}) {
	data := in.(*msg.PhoneActionParticipantRemoved)
	fmt.Println(data)
}

func (c *callController) adminUpdated(in interface{}) {
	data := in.(*msg.PhoneActionAdminUpdated)
	fmt.Println(data)
}

func (c *callController) joinRequested(in interface{}) {
	data := in.(*msg.PhoneActionJoinRequested)
	fmt.Println(data)
}

func (c *callController) screenShareUpdated(in interface{}) {
	data := in.(*msg.PhoneActionScreenShare)
	fmt.Println(data)
}

func (c *callController) callPicked(in interface{}) {
	data := in.(*msg.PhoneActionPicked)
	fmt.Println(data)
}

func (c *callController) callRestarted(in interface{}) {
	data := in.(*msg.PhoneActionRestarted)
	fmt.Println(data)
}
