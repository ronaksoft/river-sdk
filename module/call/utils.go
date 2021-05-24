package call

import (
	"errors"
	"git.ronaksoft.com/river/msg/go/msg"
	"sync"
	"time"
)

/*
   Creation Time: 2021 - May - 19
   Created by:  (Hamidrezakk)
   Maintainers:
      1.  Hamidrezakk
   Auditor: Hamidrezakk
   Copyright Ronak Software Group 2021
*/

var (
	ErrActionNotFound             = errors.New("action not found")
	ErrInvalidCallId              = errors.New("invalid call id")
	ErrInvalidConnId              = errors.New("invalid conn id")
	ErrInvalidPeerInput           = errors.New("invalid peer input")
	ErrInvalidCallRequest         = errors.New("invalid call request")
	ErrNoActiveCall               = errors.New("no active call")
	ErrNoSDP                      = errors.New("no sdp")
	ErrCannotSetAnswerSDP         = errors.New("cannot set answer sdp")
	ErrNoCallRequest              = errors.New("no call request")
	ErrCannotInitStream           = errors.New("cannot init stream")
	ErrCannotInitConnection       = errors.New("cannot init connection")
	ErrCannotCloseConnection      = errors.New("cannot close connection")
	ErrInvalidRequest             = errors.New("invalid request")
	ErrInvalidResponse            = errors.New("invalid response")
	ErrCallbacksAreNotInitialized = errors.New("callbacks are not initialized")
)

type UpdatePhoneCall struct {
	*msg.UpdatePhoneCall
	Data interface{}
}

type Connection struct {
	msg.CallConnection
	mu              *sync.RWMutex
	connectTicker   *time.Ticker
	reconnectTimout *time.Timer
}

type MediaSettingsIn struct {
	Audio       *bool
	ScreenShare *bool
	Video       *bool
}

type MediaSettings struct {
	Audio       bool
	ScreenShare bool
	Video       bool
}

type Info struct {
	acceptedParticipantIds []int64
	acceptedParticipants   []int32
	allConnected           bool
	dialed                 bool
	mediaSettings          *msg.CallMediaSettings
	participantMap         map[int64]int32
	participants           map[int32]*msg.CallParticipant
	requestParticipantIds  []int64
	requests               []*UpdatePhoneCall
	iceServer              *msg.IceServer
	requestMap             map[int64]struct{}
	mu                     *sync.RWMutex
}

func parseCallAction(constructor msg.PhoneCallAction, data []byte) (out interface{}, err error) {
	switch constructor {
	case msg.PhoneCallAction_PhoneCallRequested:
		t1 := &msg.PhoneActionRequested{}
		err = t1.Unmarshal(data)
		if err != nil {
			return
		}
		return t1, nil
	case msg.PhoneCallAction_PhoneCallAccepted:
		t2 := &msg.PhoneActionAccepted{}
		err = t2.Unmarshal(data)
		if err != nil {
			return
		}
		return t2, nil
	case msg.PhoneCallAction_PhoneCallDiscarded:
		t3 := &msg.PhoneActionDiscarded{}
		err = t3.Unmarshal(data)
		if err != nil {
			return
		}
		return t3, nil
	case msg.PhoneCallAction_PhoneCallCallWaiting:
		t4 := &msg.PhoneActionCallWaiting{}
		err = t4.Unmarshal(data)
		if err != nil {
			return
		}
		return t4, nil
	case msg.PhoneCallAction_PhoneCallIceExchange:
		t5 := &msg.PhoneActionIceExchange{}
		err = t5.Unmarshal(data)
		if err != nil {
			return
		}
		return t5, nil
	case msg.PhoneCallAction_PhoneCallEmpty:
		t6 := &msg.PhoneActionCallEmpty{}
		err = t6.Unmarshal(data)
		if err != nil {
			return
		}
		return t6, nil
	case msg.PhoneCallAction_PhoneCallMediaSettingsChanged:
		t7 := &msg.PhoneActionMediaSettingsUpdated{}
		err = t7.Unmarshal(data)
		if err != nil {
			return
		}
		return t7, nil
	case msg.PhoneCallAction_PhoneCallSDPOffer:
		t8 := &msg.PhoneActionSDPOffer{}
		err = t8.Unmarshal(data)
		if err != nil {
			return
		}
		return t8, nil
	case msg.PhoneCallAction_PhoneCallSDPAnswer:
		t9 := &msg.PhoneActionSDPAnswer{}
		err = t9.Unmarshal(data)
		if err != nil {
			return
		}
		return t9, nil
	case msg.PhoneCallAction_PhoneCallAck:
		t10 := &msg.PhoneActionSDPAnswer{}
		err = t10.Unmarshal(data)
		if err != nil {
			return
		}
		return t10, nil
	case msg.PhoneCallAction_PhoneCallParticipantAdded:
		t11 := &msg.PhoneActionParticipantAdded{}
		err = t11.Unmarshal(data)
		if err != nil {
			return
		}
		return t11, nil
	case msg.PhoneCallAction_PhoneCallParticipantRemoved:
		t12 := &msg.PhoneActionParticipantRemoved{}
		err = t12.Unmarshal(data)
		if err != nil {
			return
		}
		return t12, nil
	case msg.PhoneCallAction_PhoneCallJoinRequested:
		t13 := &msg.PhoneActionJoinRequested{}
		err = t13.Unmarshal(data)
		if err != nil {
			return
		}
		return t13, nil
	case msg.PhoneCallAction_PhoneCallAdminUpdated:
		t14 := &msg.PhoneActionAdminUpdated{}
		err = t14.Unmarshal(data)
		if err != nil {
			return
		}
		return t14, nil
	case msg.PhoneCallAction_PhoneCallScreenShare:
		t15 := &msg.PhoneActionScreenShare{}
		err = t15.Unmarshal(data)
		if err != nil {
			return
		}
		return t15, nil
	case msg.PhoneCallAction_PhoneCallPicked:
		t16 := &msg.PhoneActionPicked{}
		err = t16.Unmarshal(data)
		if err != nil {
			return
		}
		return t16, nil
	case msg.PhoneCallAction_PhoneCallRestarted:
		t17 := &msg.PhoneActionRestarted{}
		err = t17.Unmarshal(data)
		if err != nil {
			return
		}
		return t17, nil
	}
	return nil, ErrActionNotFound
}
