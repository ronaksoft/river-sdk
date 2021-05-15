package call

import (
	"errors"
	"git.ronaksoft.com/river/msg/go/msg"
	"sync"
	"time"
)

var (
	ErrActionNotFound     = errors.New("action not found")
	ErrInvalidCallId      = errors.New("invalid call id")
	ErrInvalidConnId      = errors.New("invalid conn id")
	ErrInvalidPeerInput   = errors.New("invalid peer input")
	ErrInvalidCallRequest = errors.New("invalid call request")
	ErrNoActiveCall       = errors.New("no active call")
	ErrNoSDP              = errors.New("no sdp")
	ErrNoCallRequest      = errors.New("no call request")
)

type UpdatePhoneCall struct {
	*msg.UpdatePhoneCall
	Data interface{}
}

type Connection struct {
	msg.CallConnection
	mu              *sync.RWMutex
	connectTimout   *time.Timer
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
	mediaSettings          MediaSettings
	participantMap         map[int64]int32
	participants           map[int32]*msg.CallParticipant
	requestParticipantIds  []int64
	requests               []*UpdatePhoneCall
	iceServer              *msg.IceServer
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
