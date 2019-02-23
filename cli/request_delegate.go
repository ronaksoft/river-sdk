package main

import (
	"git.ronaksoftware.com/ronak/riversdk/logs"
	"git.ronaksoftware.com/ronak/riversdk/msg"
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
		logs.Error("Failed to unmarshal", zap.Error(err))
		return
	}
	logs.Info("Callback OnComplete()", zap.Int64("ReqID", d.RequestID), zap.String("Constructor", msg.ConstructorNames[d.Envelope.Constructor]))

	MessagePrinter(&d.Envelope)
	return
}

func (d *RequestDelegate) OnTimeout(err error) {
	logs.Error("Callback OnTimeout()", zap.Int64("ReqID", d.RequestID), zap.Error(err))
}
