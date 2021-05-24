package call

import (
	"errors"
	"git.ronaksoft.com/river/msg/go/msg"
	"github.com/ronaksoft/rony"
)

/*
   Creation Time: 2021 - May - 19
   Created by:  (Hamidrezakk)
   Maintainers:
      1.  Hamidrezakk
   Auditor: Hamidrezakk
   Copyright Ronak Software Group 2021
*/

type Callback struct {
	OnUpdate             func(updateType int32, b []byte)
	InitStream           func(audio, video bool) bool
	InitConnection       func(connId int32, b []byte) int64
	CloseConnection      func(connId int32) bool
	GetOfferSDP          func(connId int32) (out []byte)
	SetOfferGetAnswerSDP func(connId int32, req []byte) (out []byte)
	SetAnswerSDP         func(connId int32, b []byte) bool
	AddIceCandidate      func(connId int32, b []byte) bool
}

func (c *call) CallbackInitConnection(connId int32, initData *msg.PhoneInit) int64 {
	if c.callback.InitConnection == nil {
		c.Log().Error("callbacks are not initialized")
		return -1
	}

	phoneInitByte, err := initData.Marshal()
	if err != nil {
		return -1
	}

	inMe := &rony.MessageEnvelope{
		Constructor: msg.C_PhoneInit,
		RequestID:   0,
		Message:     phoneInitByte,
	}

	inMeByte, err := inMe.Marshal()
	if err != nil {
		return -1
	}

	return c.callback.InitConnection(connId, inMeByte)
}

func (c *call) CallbackGetOfferSDP(connId int32) (offerSdp *msg.PhoneActionSDPOffer, err error) {
	if c.callback.GetOfferSDP == nil {
		err = ErrCallbacksAreNotInitialized
		return
	}

	res := c.callback.GetOfferSDP(connId)
	me := &rony.MessageEnvelope{}
	err = me.Unmarshal(res)
	if err != nil {
		return
	}

	switch me.Constructor {
	case msg.C_PhoneActionSDPOffer:
		offerSdp = &msg.PhoneActionSDPOffer{}
		err = offerSdp.Unmarshal(me.Message)
		if err != nil {
			return
		}
	case msg.C_ClientError:
		errObj := &msg.ClientError{}
		err = errObj.Unmarshal(me.Message)
		if err != nil {
			return
		}
		err = errors.New(errObj.Error)
	default:
		err = ErrInvalidResponse
	}
	return
}

func (c *call) CallbackSetOfferGetAnswerSDP(connId int32, offerSdp *msg.PhoneActionSDPOffer) (answerSdp *msg.PhoneActionSDPAnswer, err error) {
	if c.callback.SetOfferGetAnswerSDP == nil {
		err = ErrCallbacksAreNotInitialized
		return
	}

	offerSdpByte, err := offerSdp.Marshal()
	if err != nil {
		return
	}

	inMe := &rony.MessageEnvelope{
		Constructor: msg.C_PhoneActionSDPOffer,
		RequestID:   0,
		Message:     offerSdpByte,
	}

	inMeByte, err := inMe.Marshal()
	if err != nil {
		return
	}

	res := c.callback.SetOfferGetAnswerSDP(connId, inMeByte)
	me := &rony.MessageEnvelope{}
	err = me.Unmarshal(res)
	if err != nil {
		return
	}

	switch me.Constructor {
	case msg.C_PhoneActionSDPAnswer:
		answerSdp = &msg.PhoneActionSDPAnswer{}
		err = answerSdp.Unmarshal(me.Message)
		if err != nil {
			return
		}
	case msg.C_ClientError:
		errObj := &msg.ClientError{}
		err = errObj.Unmarshal(me.Message)
		if err != nil {
			return
		}
		err = errors.New(errObj.Error)
	default:
		err = ErrInvalidResponse
	}
	return
}

func (c *call) CallbackSetAnswerSDP(connId int32, answerSdp *msg.PhoneActionSDPAnswer) bool {
	if c.callback.SetAnswerSDP == nil {
		c.Log().Error("callbacks are not initialized")
		return false
	}

	answerSdpByte, err := answerSdp.Marshal()
	if err != nil {
		return false
	}

	inMe := &rony.MessageEnvelope{
		Constructor: msg.C_PhoneActionSDPAnswer,
		RequestID:   0,
		Message:     answerSdpByte,
	}

	inMeByte, err := inMe.Marshal()
	if err != nil {
		return false
	}

	return c.callback.SetAnswerSDP(connId, inMeByte)
}

func (c *call) CallbackAddIceCandidate(connId int32, candidate *msg.CallRTCIceCandidate) bool {
	if c.callback.AddIceCandidate == nil {
		c.Log().Error("callbacks are not initialized")
		return false
	}

	candidateByte, err := candidate.Marshal()
	if err != nil {
		return false
	}

	inMe := &rony.MessageEnvelope{
		Constructor: msg.C_CallRTCIceCandidate,
		RequestID:   0,
		Message:     candidateByte,
	}

	inMeByte, err := inMe.Marshal()
	if err != nil {
		return false
	}

	return c.callback.AddIceCandidate(connId, inMeByte)
}
