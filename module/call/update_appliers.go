package call

import (
    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/river-sdk/internal/domain"
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

func (c *call) updatePhoneCall(u *msg.UpdateEnvelope) (res []*msg.UpdateEnvelope, err error) {
    // TODO: Add worker
    go func() {
        x := &msg.UpdatePhoneCall{}
        err = x.Unmarshal(u.Update)
        if err != nil {
            return
        }

        c.Log().Info("applies UpdatePhoneCall", zap.String("Action", x.Action.String()))

        now := domain.Now().Unix()
        if x.Timestamp != 0 && now-x.Timestamp > 60 {
            return
        }

        data, err := parseCallAction(x.Action, x.ActionData)
        if err != nil {
            c.Log().Debug("parseCallAction", zap.Error(err))
            return
        }

        update := &UpdatePhoneCall{
            UpdatePhoneCall: x,
            Data:            data,
        }

        if data == nil {
            c.Log().Debug("Update data is nil")
            return
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
    }()

    res = []*msg.UpdateEnvelope{u}
    return
}
