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
	"errors"
	"git.ronaksoft.com/river/msg/go/msg"
)

const (
	RetryInterval = 10000
	RetryLimit    = 6
)

var ErrActionNotFound = errors.New("action not found")

type UpdatePhoneCall struct {
	msg.UpdatePhoneCall
	Data interface{}
}

type CallParticipant struct {
	msg.PhoneParticipant
	DeviceType    msg.CallDeviceType
	MediaSettings msg.CallMediaSettings
	Started       bool
}

type RTCIceCandidate struct {
	Candidate        string
	SDPMid           string
	SDPMLineIndex    int64
	UsernameFragment string
}

type CallConnection struct {
	Accepted bool
	IceQueue []RTCIceCandidate
	Interval interface{}
	Try      int64
}

type CallInfo struct {
	AcceptedParticipantIds []int64
	AcceptedParticipants   []int
	AllConnected           bool
	Dialed                 bool
	MediaSettings          msg.CallMediaSettings
	ParticipantMap         map[int64]int64
	Participants           map[int64]CallParticipant
	RequestParticipantIds  []int64
	Requests               []UpdatePhoneCall
}

func parseCallAction(constructor msg.PhoneCallAction, data []byte) (out interface{}, err error) {
	switch constructor {
	case msg.PhoneCallAction_PhoneCallRequested:
		t1 := &msg.PhoneActionRequested{}
		err = t1.Unmarshal(data)
		if err != nil {
			return
		}
		return out, nil
	case msg.PhoneCallAction_PhoneCallAccepted:
		t2 := &msg.PhoneActionAccepted{}
		err = t2.Unmarshal(data)
		if err != nil {
			return
		}
		return out, nil
	case msg.PhoneCallAction_PhoneCallDiscarded:
		t3 := &msg.PhoneActionDiscarded{}
		err = t3.Unmarshal(data)
		if err != nil {
			return
		}
		return out, nil
	case msg.PhoneCallAction_PhoneCallCallWaiting:
		t4 := &msg.PhoneActionCallWaiting{}
		err = t4.Unmarshal(data)
		if err != nil {
			return
		}
		return out, nil
	case msg.PhoneCallAction_PhoneCallIceExchange:
		t5 := &msg.PhoneActionIceExchange{}
		err = t5.Unmarshal(data)
		if err != nil {
			return
		}
		return out, nil
	case msg.PhoneCallAction_PhoneCallEmpty:
		t6 := &msg.PhoneActionCallEmpty{}
		err = t6.Unmarshal(data)
		if err != nil {
			return
		}
		return out, nil
	case msg.PhoneCallAction_PhoneCallMediaSettingsChanged:
		t7 := &msg.PhoneActionMediaSettingsUpdated{}
		err = t7.Unmarshal(data)
		if err != nil {
			return
		}
		return out, nil
	case msg.PhoneCallAction_PhoneCallSDPOffer:
		t8 := &msg.PhoneActionSDPOffer{}
		err = t8.Unmarshal(data)
		if err != nil {
			return
		}
		return out, nil
	case msg.PhoneCallAction_PhoneCallSDPAnswer:
		t9 := &msg.PhoneActionSDPAnswer{}
		err = t9.Unmarshal(data)
		if err != nil {
			return
		}
		return out, nil
	case msg.PhoneCallAction_PhoneCallParticipantAdded:
		t10 := &msg.PhoneActionParticipantAdded{}
		err = t10.Unmarshal(data)
		if err != nil {
			return
		}
		return out, nil
	case msg.PhoneCallAction_PhoneCallParticipantRemoved:
		t11 := &msg.PhoneActionParticipantRemoved{}
		err = t11.Unmarshal(data)
		if err != nil {
			return
		}
		return out, nil
	case msg.PhoneCallAction_PhoneCallJoinRequested:
		t12 := &msg.PhoneActionJoinRequested{}
		err = t12.Unmarshal(data)
		if err != nil {
			return
		}
		return out, nil
	case msg.PhoneCallAction_PhoneCallAdminUpdated:
		t13 := &msg.PhoneActionAdminUpdated{}
		err = t13.Unmarshal(data)
		if err != nil {
			return
		}
		return out, nil
	case msg.PhoneCallAction_PhoneCallScreenShare:
		t14 := &msg.PhoneActionScreenShare{}
		err = t14.Unmarshal(data)
		if err != nil {
			return
		}
		return out, nil
	case msg.PhoneCallAction_PhoneCallPicked:
		t15 := &msg.PhoneActionPicked{}
		err = t15.Unmarshal(data)
		if err != nil {
			return
		}
		return out, nil
	case msg.PhoneCallAction_PhoneCallRestarted:
		t15 := &msg.PhoneActionRestarted{}
		err = t15.Unmarshal(data)
		if err != nil {
			return
		}
		return out, nil
	}
	return nil, ErrActionNotFound
}

type callController struct {
	peerConnections map[int64]CallConnection
	peer            *msg.InputPeer
	activeCallID    int64
	callInfo        map[int64]CallInfo
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
