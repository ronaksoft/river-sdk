package main

import (
	"go.uber.org/zap"

	msg "git.ronaksoftware.com/river/msg/chat"
)

func UpdatePrinter(envelope *msg.UpdateEnvelope) {
	// constructorName, _ := msg.ConstructorNames[envelope.Constructor]
	// _Shell.Println(_MAGNETA("ConstructorName: %s (0x%X)", constructorName, envelope.Constructor))
	// _Shell.Println(_MAGNETA("UpdateID: %10d, UCount: %4d", envelope.UpdateID, envelope.UCount))
	switch envelope.Constructor {
	case msg.C_UpdateNewMessage:
		x := new(msg.UpdateNewMessage)
		x.Unmarshal(envelope.Update)
		// _Shell.Println(fmt.Sprintf("UpdateNewMessage \t MsgID:%d, PeerID:%d , SenderID:%d , Body:%s",
		// 	x.Message.ID, x.Message.PeerID, x.Message.SenderID, x.Message.Body))
	case msg.C_UpdateReadHistoryInbox:
		x := new(msg.UpdateReadHistoryInbox)
		x.Unmarshal(envelope.Update)
		// _Shell.Println(fmt.Sprintf("UpdateReadHistoryInbox \t PeerID:%d , MaxID:%d", x.Peer.ID, x.MaxID))
	case msg.C_UpdateReadHistoryOutbox:
		x := new(msg.UpdateReadHistoryOutbox)
		x.Unmarshal(envelope.Update)
		// _Shell.Println(fmt.Sprintf("UpdateReadHistoryOutbox \t PeerID:%d , MaxID:%d", x.Peer.ID, x.MaxID))
	case msg.C_UpdateUserTyping:
		x := new(msg.UpdateUserTyping)
		x.Unmarshal(envelope.Update)
		// _Shell.Println(fmt.Sprintf("UpdateUserTyping \t userID:%d , Action:%s", x.UserID, x.Action.String()))

	case msg.C_ClientUpdatePendingMessageDelivery:
		x := new(msg.ClientUpdatePendingMessageDelivery)
		err := x.Unmarshal(envelope.Update)
		if err != nil {
			_Log.Error("Failed to unmarshal", zap.Error(err))
			return
		}
		// _Shell.Println(fmt.Sprintf("#UPDATE PendingMessageDelivery: %v", x.Success))
		// _Shell.Println(fmt.Sprintf("PendingMessage: %v", x.PendingMessage))
		// _Shell.Println(fmt.Sprintf("Messages: %v", x.Messages))
	case msg.C_UpdateContainer:
		x := new(msg.UpdateContainer)
		err := x.Unmarshal(envelope.Update)
		if err != nil {
			_Log.Error("Failed to unmarshal", zap.Error(err))
			return
		}
		for _, u := range x.Updates {
			UpdatePrinter(u)
		}

	default:
		// _Shell.Println("Received Update", zap.String("C", msg.ConstructorNames[envelope.Constructor]))
	}
}
