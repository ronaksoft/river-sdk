package main

import (
	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/toolbox"
	"gopkg.in/abiosoft/ishell.v2"
)

var Message = &ishell.Cmd{
	Name: "Messages",
}

// MessageSend
var MessageSend = &ishell.Cmd{
	Name: "Send",
	Func: func(c *ishell.Context) {
		req := msg.MessagesSend{}
		req.RandomID = ronak.RandomInt64(0)
		req.Peer = &msg.InputPeer{}
		req.Peer.Type = fnGetPeerType(c)
		req.Peer.ID = fnGetPeerID(c)
		req.Peer.AccessHash = fnGetAccessHash(c)
		req.Body = fnGetBody(c)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesSend, reqBytes, reqDelegate, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

// MessageGetDialogs
var MessageGetDialogs = &ishell.Cmd{
	Name: "GetDialogs",
	Func: func(c *ishell.Context) {
		req := msg.MessagesGetDialogs{}
		req.Limit = int32(100)
		req.Offset = int32(0)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesGetDialogs, reqBytes, reqDelegate, false); err != nil {
			_Log.Debug(err.Error())
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
		if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesGetDialog, reqBytes, reqDelegate, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

// MessageGetHistory History
var MessageGetHistory = &ishell.Cmd{
	Name: "MessageGetHistory",
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
		if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesGetHistory, reqBytes, reqDelegate, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var MessageReadHistory = &ishell.Cmd{
	Name: "MessageReadHistory",
	Func: func(c *ishell.Context) {
		req := msg.MessagesReadHistory{}
		req.Peer = &msg.InputPeer{}
		req.Peer.Type = fnGetPeerType(c)
		req.Peer.ID = fnGetPeerID(c)
		req.Peer.AccessHash = fnGetAccessHash(c)
		req.MaxID = fnGetMaxID(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesReadHistory, reqBytes, reqDelegate, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var MessageSetTyping = &ishell.Cmd{
	Name: "MessageSetTyping",
	Func: func(c *ishell.Context) {
		req := msg.MessagesSetTyping{}
		req.Peer = &msg.InputPeer{}
		req.Peer.Type = fnGetPeerType(c)
		req.Peer.ID = fnGetPeerID(c)
		req.Peer.AccessHash = fnGetAccessHash(c)
		req.Action = fnGetTypingAction(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesSetTyping, reqBytes, reqDelegate, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var MessagesGet = &ishell.Cmd{
	Name: "MessagesGet",
	Func: func(c *ishell.Context) {
		req := msg.MessagesGet{}
		req.Peer = &msg.InputPeer{}
		req.Peer.Type = fnGetPeerType(c)

		req.Peer.ID = fnGetPeerID(c)
		req.Peer.AccessHash = fnGetAccessHash(c)
		req.MessagesIDs = fnGetMessageIDs(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesGet, reqBytes, reqDelegate, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var MessagesClearHistory = &ishell.Cmd{
	Name: "MessagesClearHistory",
	Func: func(c *ishell.Context) {
		req := msg.MessagesClearHistory{}
		req.Peer = &msg.InputPeer{}
		req.Peer.Type = fnGetPeerType(c)
		req.Peer.ID = fnGetPeerID(c)
		req.MaxID = fnGetMaxID(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesClearHistory, reqBytes, reqDelegate, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var MessagesDelete = &ishell.Cmd{
	Name: "MessagesDelete",
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
		if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesDelete, reqBytes, reqDelegate, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var MessagesEdit = &ishell.Cmd{
	Name: "MessagesEdit",
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
		if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesEdit, reqBytes, reqDelegate, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var MessagesForward = &ishell.Cmd{
	Name: "MessagesForward",
	Func: func(c *ishell.Context) {
		req := msg.MessagesForward{}
		req.FromPeer = &msg.InputPeer{}
		req.ToPeer = &msg.InputPeer{}

		req.RandomID = domain.SequentialUniqueID()
		req.Silence = fnGetSilence(c)

		c.Print("***** From Peer :")
		req.FromPeer.Type = fnGetPeerType(c)
		req.FromPeer.ID = fnGetPeerID(c)
		req.FromPeer.AccessHash = fnGetAccessHash(c)

		c.Print("***** To Peer :")
		req.ToPeer.Type = fnGetPeerType(c)
		req.ToPeer.ID = fnGetPeerID(c)
		req.ToPeer.AccessHash = fnGetAccessHash(c)

		req.MessageIDs = fnGetMessageIDs(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesForward, reqBytes, reqDelegate, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

func init() {
	Message.AddCmd(MessageGetDialogs)
	Message.AddCmd(MessageGetDialog)
	Message.AddCmd(MessageSend)
	// Message.AddCmd(MessageReadHistory)
	Message.AddCmd(MessageGetHistory)
	Message.AddCmd(MessageReadHistory)
	Message.AddCmd(MessageSetTyping)
	Message.AddCmd(MessagesGet)
	Message.AddCmd(MessagesClearHistory)
	Message.AddCmd(MessagesDelete)
	Message.AddCmd(MessagesEdit)
	Message.AddCmd(MessagesForward)
}
