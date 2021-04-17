package main

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"go.uber.org/zap"
)

func UpdatePrinter(envelope *msg.UpdateEnvelope) {
	switch envelope.Constructor {
	case msg.C_UpdateNewMessage:
		x := new(msg.UpdateNewMessage)
		x.Unmarshal(envelope.Update)
	case msg.C_UpdateReadHistoryInbox:
		x := new(msg.UpdateReadHistoryInbox)
		x.Unmarshal(envelope.Update)
	case msg.C_UpdateReadHistoryOutbox:
		x := new(msg.UpdateReadHistoryOutbox)
		x.Unmarshal(envelope.Update)
	case msg.C_UpdateUserTyping:
		x := new(msg.UpdateUserTyping)
		x.Unmarshal(envelope.Update)
	case msg.C_ClientUpdatePendingMessageDelivery:
		x := new(msg.ClientUpdatePendingMessageDelivery)
		err := x.Unmarshal(envelope.Update)
		if err != nil {
			_Shell.Println("Failed to unmarshal", zap.Error(err))
			return
		}
	case msg.C_UpdateMessageID:
	case msg.C_UpdateContainer:
		x := new(msg.UpdateContainer)
		err := x.Unmarshal(envelope.Update)
		if err != nil {
			_Shell.Println("Failed to unmarshal", zap.Error(err))
			return
		}
		for _, u := range x.Updates {
			UpdatePrinter(u)
		}

	default:
		// _Shell.Println("Received Update", zap.String("C", registry.ConstructorName(envelope.Constructor)))
	}
}
