package main

import (
	msg "git.ronaksoftware.com/river/msg/chat"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"gopkg.in/abiosoft/ishell.v2"
	"time"
)

var Message = &ishell.Cmd{
	Name: "Messages",
}

var MessageSend = &ishell.Cmd{
	Name: "SendMessage",
	Func: func(c *ishell.Context) {
		req := msg.MessagesSend{}
		req.RandomID = domain.RandomInt63()
		req.Peer = &msg.InputPeer{}
		req.Peer.Type = fnGetPeerType(c)
		req.Peer.ID = fnGetPeerID(c)
		req.Peer.AccessHash = fnGetAccessHash(c)
		req.Body = fnGetBody(c)
		req.Entities = fnGetEntities(c)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesSend, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var MessageSendToSelf = &ishell.Cmd{
	Name: "SendToMe",
	Func: func(c *ishell.Context) {
		req := msg.MessagesSend{}
		req.RandomID = domain.RandomInt63()
		req.Peer = &msg.InputPeer{}
		req.Peer.Type = msg.PeerUser
		req.Peer.ID = _SDK.ConnInfo.UserID
		req.Peer.AccessHash = 0
		req.Body = fnGetBody(c)
		if len(req.Body) == 0 {
			req.Body = domain.RandomID(32)
		}
		req.Entities = nil
		reqBytes, _ := req.Marshal()
		reqDelegate := NewCustomDelegate()
		startTime := time.Now()
		reqDelegate.OnCompleteFunc = func(b []byte) {
			x := &msg.MessageEnvelope{}
			_ = x.Unmarshal(b)
			MessagePrinter(x)
			_Shell.Println("Execute Time:", time.Now().Sub(startTime))
		}
		if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesSend, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var MessageGetDialogs = &ishell.Cmd{
	Name: "GetDialogs",
	Func: func(c *ishell.Context) {
		req := msg.MessagesGetDialogs{}
		req.Limit = int32(100)
		req.Offset = int32(0)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesGetDialogs, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var MessageGetDialog = &ishell.Cmd{
	Name: "GetDialog",
	Func: func(c *ishell.Context) {
		req := msg.MessagesGetDialog{}
		req.Peer = &msg.InputPeer{}
		req.Peer.Type = fnGetPeerType(c)
		req.Peer.ID = fnGetPeerID(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesGetDialog, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var MessageGetHistory = &ishell.Cmd{
	Name: "GetHistory",
	Func: func(c *ishell.Context) {
		req := msg.MessagesGetHistory{}
		req.Peer = &msg.InputPeer{}
		req.Peer.Type = fnGetPeerType(c)
		req.Peer.ID = fnGetPeerID(c)
		req.Peer.AccessHash = fnGetAccessHash(c)
		req.MaxID = fnGetMaxID(c)
		req.MinID = fnGetMinID(c)
		req.Limit = fnGetLimit(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesGetHistory, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var MessageReadHistory = &ishell.Cmd{
	Name: "ReadHistory",
	Func: func(c *ishell.Context) {
		req := msg.MessagesReadHistory{}
		req.Peer = &msg.InputPeer{}
		req.Peer.Type = fnGetPeerType(c)
		req.Peer.ID = fnGetPeerID(c)
		req.Peer.AccessHash = fnGetAccessHash(c)
		req.MaxID = fnGetMaxID(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesReadHistory, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var MessageSetTyping = &ishell.Cmd{
	Name: "SetTyping",
	Func: func(c *ishell.Context) {
		req := msg.MessagesSetTyping{}
		req.Peer = &msg.InputPeer{}
		req.Peer.Type = fnGetPeerType(c)
		req.Peer.ID = fnGetPeerID(c)
		req.Peer.AccessHash = fnGetAccessHash(c)
		req.Action = fnGetTypingAction(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesSetTyping, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var MessagesGet = &ishell.Cmd{
	Name: "Get",
	Func: func(c *ishell.Context) {
		req := msg.MessagesGet{}
		req.Peer = &msg.InputPeer{}
		req.Peer.Type = fnGetPeerType(c)

		req.Peer.ID = fnGetPeerID(c)
		req.Peer.AccessHash = fnGetAccessHash(c)
		req.MessagesIDs = fnGetMessageIDs(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesGet, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var MessagesClearHistory = &ishell.Cmd{
	Name: "ClearHistory",
	Func: func(c *ishell.Context) {
		req := msg.MessagesClearHistory{}
		req.Peer = &msg.InputPeer{}
		req.Peer.Type = fnGetPeerType(c)
		req.Peer.ID = fnGetPeerID(c)
		req.Peer.AccessHash = fnGetAccessHash(c)
		req.MaxID = fnGetMaxID(c)
		req.Delete = fnGetDelete(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesClearHistory, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var MessagesDelete = &ishell.Cmd{
	Name: "Delete",
	Func: func(c *ishell.Context) {
		req := msg.MessagesDelete{}
		req.Peer = &msg.InputPeer{}
		req.Peer.Type = fnGetPeerType(c)
		req.Peer.ID = fnGetPeerID(c)
		req.Peer.AccessHash = fnGetAccessHash(c)
		req.Revoke = fnGetRevoke(c)
		req.MessageIDs = fnGetMessageIDs(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesDelete, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var MessagesEdit = &ishell.Cmd{
	Name: "Edit",
	Func: func(c *ishell.Context) {
		req := msg.MessagesEdit{}
		req.Peer = &msg.InputPeer{}
		req.Peer.Type = fnGetPeerType(c)
		req.Peer.ID = fnGetPeerID(c)
		req.Peer.AccessHash = fnGetAccessHash(c)
		req.MessageID = fnGetMessageID(c)
		req.Body = fnGetBody(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesEdit, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var MessagesForward = &ishell.Cmd{
	Name: "Forward",
	Func: func(c *ishell.Context) {
		req := msg.MessagesForward{}
		req.FromPeer = &msg.InputPeer{}
		req.ToPeer = &msg.InputPeer{}

		req.RandomID = domain.SequentialUniqueID()
		req.Silence = fnGetSilence(c)

		c.Println("***** From Peer :")
		req.FromPeer.Type = fnGetPeerType(c)
		req.FromPeer.ID = fnGetPeerID(c)
		req.FromPeer.AccessHash = fnGetAccessHash(c)

		c.Println("***** To Peer :")
		req.ToPeer.Type = fnGetPeerType(c)
		req.ToPeer.ID = fnGetPeerID(c)
		req.ToPeer.AccessHash = fnGetAccessHash(c)

		req.MessageIDs = fnGetMessageIDs(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesForward, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var MessagesReadContents = &ishell.Cmd{
	Name: "ReadContents",
	Func: func(c *ishell.Context) {
		req := msg.MessagesReadContents{
			Peer:       new(msg.InputPeer),
			MessageIDs: make([]int64, 0),
		}
		req.Peer.Type = fnGetPeerType(c)
		req.Peer.ID = fnGetPeerID(c)
		req.Peer.AccessHash = fnGetAccessHash(c)
		req.MessageIDs = fnGetMessageIDs(c)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesReadContents, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var MessagesGetDBMediaStatus = &ishell.Cmd{
	Name: "GetDBMediaStatus",
	Func: func(c *ishell.Context) {
		// reqDelegate := new(dbMediaDelegate)
		// _SDK.GetDBStatus(reqDelegate)
	},
}

var MessagesClearMedia = &ishell.Cmd{
	Name: "ClearMedia",
	Func: func(c *ishell.Context) {
		// peerId := fnGetPeerID(c)
		// all := fnClearAll(c)
		// mediaType := fnGetMediaTypes(c)
		// status := _SDK.ClearCache(peerId, mediaType, all)
		// _Log.Debug("MessagesClearMedia::status", zap.Bool("", status))
	},
}

func init() {
	Message.AddCmd(MessageGetDialogs)
	Message.AddCmd(MessageGetDialog)
	Message.AddCmd(MessageSend)
	Message.AddCmd(MessageSendToSelf)
	Message.AddCmd(MessageGetHistory)
	Message.AddCmd(MessageReadHistory)
	Message.AddCmd(MessageSetTyping)
	Message.AddCmd(MessagesGet)
	Message.AddCmd(MessagesClearHistory)
	Message.AddCmd(MessagesDelete)
	Message.AddCmd(MessagesEdit)
	Message.AddCmd(MessagesForward)
	Message.AddCmd(MessagesReadContents)
	Message.AddCmd(MessagesGetDBMediaStatus)
	Message.AddCmd(MessagesClearMedia)
}
