package call

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"go.uber.org/zap"
)

func (c *call) updatePhoneCall(u *msg.UpdateEnvelope) (res []*msg.UpdateEnvelope, err error) {
	if u.Constructor != msg.C_UpdatePhoneCall {
		return
	}

	x := &msg.UpdatePhoneCall{}
	err = x.Unmarshal(u.Update)
	if err != nil {
		return
	}

	logs.Info("updatePhoneCall", zap.Int32("action", int32(x.Action)))

	now := domain.Now().Unix()
	if !(x.Timestamp == 0 || now-x.Timestamp < 60) {
		return
	}

	data, err := parseCallAction(x.Action, x.ActionData)
	if err != nil {
		logs.Debug("parseCallAction", zap.Error(err))
		return
	}

	update := &UpdatePhoneCall{
		UpdatePhoneCall: x,
		Data:            data,
	}

	switch x.Action {
	case msg.PhoneCallAction_PhoneCallRequested:
		c.callRequested(update)
	case msg.PhoneCallAction_PhoneCallAccepted:
		c.callAccepted(update)
	case msg.PhoneCallAction_PhoneCallDiscarded:
		c.callDiscarded(update)
	case msg.PhoneCallAction_PhoneCallIceExchange:
		c.iceExchange(update)
	case msg.PhoneCallAction_PhoneCallMediaSettingsChanged:
		c.mediaSettingsUpdated(update)
	case msg.PhoneCallAction_PhoneCallSDPOffer:
		c.sdpOfferUpdated(update)
	case msg.PhoneCallAction_PhoneCallSDPAnswer:
		c.sdpAnswerUpdated(update)
	case msg.PhoneCallAction_PhoneCallAck:
		c.callAcknowledged(update)
	case msg.PhoneCallAction_PhoneCallParticipantAdded:
		c.participantAdded(update)
	case msg.PhoneCallAction_PhoneCallParticipantRemoved:
		c.participantRemoved(update)
	case msg.PhoneCallAction_PhoneCallAdminUpdated:
		c.adminUpdated(update)
	case msg.PhoneCallAction_PhoneCallJoinRequested:
		c.joinRequested(update)
	case msg.PhoneCallAction_PhoneCallScreenShare:
		c.screenShareUpdated(update)
	case msg.PhoneCallAction_PhoneCallPicked:
		c.callPicked(update)
	case msg.PhoneCallAction_PhoneCallRestarted:
		c.callRestarted(update)
	}

	return
}
