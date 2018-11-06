package main

import (
	"git.ronaksoftware.com/ronak/riversdk/msg"
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
		_Shell.Println(_RED("Callback OnComplete() Unmarshal Error: %v", err.Error()))
	} else {
		_Shell.Println(_RED("Callback OnComplete() Constructor: %v", msg.ConstructorNames[d.Envelope.Constructor]))
	}
	_Shell.Println(_GREEN("RequestID: %d", d.RequestID))
	MessagePrinter(&d.Envelope)
	return
}

func (d *RequestDelegate) OnTimeout(err error) {
	_Shell.Println(_RED("Callback OnTimeout() Error: %v", err.Error()))
}
