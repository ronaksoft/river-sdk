package main

import (
	msg "git.ronaksoftware.com/river/msg/chat"
	"go.uber.org/zap"
)

/*
   Creation Time: 2018 - Sep - 05
   Created by:  (ehsan)
   Maintainers:
       1.  (ehsan)
   Auditor: Ehsan N. Moosa
   Copyright Ronak Software Group 2018
*/

type RequestDelegate struct {
	RequestID int64
	Envelope  msg.MessageEnvelope
}

func (d *RequestDelegate) OnComplete(b []byte) {
	err := d.Envelope.Unmarshal(b)
	if err != nil {
		_Log.Error("Failed to unmarshal", zap.Error(err))
		return
	}
	_Log.Info("Callback OnComplete()", zap.Int64("ReqID", d.RequestID), zap.String("C", msg.ConstructorNames[d.Envelope.Constructor]))

	MessagePrinter(&d.Envelope)
	return
}

func (d *RequestDelegate) OnTimeout(err error) {
	_Log.Error("Callback OnTimeout()", zap.Int64("ReqID", d.RequestID), zap.Error(err))
}

func (d *RequestDelegate) Flags() int32 {
	return 0
}

type CustomRequestDelegate struct {
	OnCompleteFunc func(b []byte)
	OnTimeoutFunc  func(err error)
	FlagsFunc      func() int32
}

func (c CustomRequestDelegate) OnComplete(b []byte) {
	c.OnCompleteFunc(b)
}

func (c CustomRequestDelegate) OnTimeout(err error) {
	c.OnTimeoutFunc(err)
}

func (c CustomRequestDelegate) Flags() int32 {
	return c.FlagsFunc()
}

func NewCustomDelegate() *CustomRequestDelegate {
	c := &CustomRequestDelegate{}
	d := &RequestDelegate{}
	c.OnCompleteFunc = d.OnComplete
	c.OnTimeoutFunc = d.OnTimeout
	c.FlagsFunc = d.Flags
	return c
}
