package main

import (
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
	_Log.Debug("OnComplete Called")
	d.Envelope.Unmarshal(b)
	_Shell.Println(_GREEN("RequestID: %d", d.RequestID))
	MessagePrinter(&d.Envelope)
	return
}

func (d *RequestDelegate) OnTimeout(err error) {
	_Log.Debug("OnTimeout Called",
		zap.Error(err),
	)

}
