package main

import (
	"fmt"

	"go.uber.org/zap"

	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
)

func UpdatePrinter(envelope *msg.UpdateEnvelope) {
	// constructorName, _ := msg.ConstructorNames[envelope.Constructor]
	// _Shell.Println(_MAGNETA("ConstructorName: %s (0x%X)", constructorName, envelope.Constructor))
	// _Shell.Println(_MAGNETA("UpdateID: %10d, UCount: %4d", envelope.UpdateID, envelope.UCount))
	switch envelope.Constructor {
	case msg.C_UpdateNewMessage:
		x := new(msg.UpdateNewMessage)
		x.Unmarshal(envelope.Update)
		logs.Message(fmt.Sprintf("UpdateNewMessage \t MsgID:%d, PeerID:%d , SenderID:%d , Body:%s",
			x.Message.ID, x.Message.PeerID, x.Message.SenderID, x.Message.Body))
	case msg.C_UpdateReadHistoryInbox:
		x := new(msg.UpdateReadHistoryInbox)
		x.Unmarshal(envelope.Update)
		logs.Message(fmt.Sprintf("UpdateReadHistoryInbox \t PeerID:%d , MaxID:%d", x.Peer.ID, x.MaxID))
	case msg.C_UpdateReadHistoryOutbox:
		x := new(msg.UpdateReadHistoryOutbox)
		x.Unmarshal(envelope.Update)
		logs.Message(fmt.Sprintf("UpdateReadHistoryOutbox \t PeerID:%d , MaxID:%d", x.Peer.ID, x.MaxID))
	case msg.C_UpdateUserTyping:
		x := new(msg.UpdateUserTyping)
		x.Unmarshal(envelope.Update)
		logs.Message(fmt.Sprintf("UpdateUserTyping \t UserID:%d , Action:%s", x.UserID, x.Action.String()))

	case msg.C_ClientUpdatePendingMessageDelivery:
		x := new(msg.ClientUpdatePendingMessageDelivery)
		err := x.Unmarshal(envelope.Update)
		if err != nil {
			logs.Error("Failed to unmarshal", zap.Error(err))
			return
		}
		logs.Message(fmt.Sprintf("#UPDATE PendingMessageDelivery: %v", x.Success))
		logs.Message(fmt.Sprintf("PendingMessage: %v", x.PendingMessage))
		logs.Message(fmt.Sprintf("Messages: %v", x.Messages))
	case msg.C_UpdateContainer:
		x := new(msg.UpdateContainer)
		err := x.Unmarshal(envelope.Update)
		if err != nil {
			logs.Error("Failed to unmarshal", zap.Error(err))
			return
		}
		for _, u := range x.Updates {
			UpdatePrinter(u)
		}

	default:
		logs.Message("Received Update", zap.String("Constructor", msg.ConstructorNames[envelope.Constructor]))
	}
}
