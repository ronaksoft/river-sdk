package main

import "git.ronaksoftware.com/ronak/riversdk/msg"

func UpdatePrinter(envelope *msg.UpdateEnvelope) {
	constructorName, _ := msg.ConstructorNames[envelope.Constructor]
	_Shell.Println(_MAGNETA("ConstructorName: %s (0x%X)", constructorName, envelope.Constructor))
	_Shell.Println(_MAGNETA("UpdateID: %10d, UCount: %4d", envelope.UpdateID, envelope.UCount))
	switch envelope.Constructor {
	case msg.C_UpdateNewMessage:
		x := new(msg.UpdateNewMessage)
		x.Unmarshal(envelope.Update)
		_Shell.Println(_MAGNETA("PeerID, MessageID, SenderID, Body: %d %d %d %s",
			x.Message.PeerID, x.Message.ID, x.Message.SenderID, x.Message.Body,
		))
	case msg.C_UpdateReadHistoryInbox:
		x := new(msg.UpdateReadHistoryInbox)
		x.Unmarshal(envelope.Update)
	case msg.C_UpdateReadHistoryOutbox:
		x := new(msg.UpdateReadHistoryOutbox)
		x.Unmarshal(envelope.Update)
	case msg.C_UpdateUserTyping:
		x := new(msg.UpdateUserTyping)
		x.Unmarshal(envelope.Update)
		_Shell.Println(_MAGNETA("UserID, Action: %10d, %s", x.UserID, x.Action.String()))

	case msg.C_ClientUpdatePendingMessageDelivery:
		x := new(msg.ClientUpdatePendingMessageDelivery)
		err := x.Unmarshal(envelope.Update)
		if err != nil {
			_Shell.Println(_BLUE("#UPDATE failed to unmarshal: %v", err))
			return
		}
		_Shell.Println(_BLUE("#UPDATE PendingMessageDelivery: %v", x.Success))
		_Shell.Println(_BLUE("PendingMessage: %v", x.PendingMessage))
		_Shell.Println(_BLUE("Messages: %v", x.Messages))

	case msg.C_UpdateMessageEdited:
	case msg.C_UpdateMessageID:

	}
}
