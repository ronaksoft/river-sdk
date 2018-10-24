package main

import (
	"strconv"

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
		req.Peer.Type = msg.PeerType_PeerUser
		for {
			c.Print("Peer ID: ")
			peerID, err := strconv.ParseInt(c.ReadLine(), 10, 64)
			req.Peer.ID = peerID
			if err == nil {
				break
			} else {
				c.Println(err.Error())
			}
		}
		for {
			c.Print("Access Hash: ")
			accessHash, err := strconv.ParseUint(c.ReadLine(), 10, 64)
			req.Peer.AccessHash = uint64(accessHash)
			if err == nil {
				break
			} else {
				c.Println(err.Error())
			}
		}
		c.Print("Body: ")
		req.Body = c.ReadLine()

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(int64(msg.C_MessagesSend), reqBytes, reqDelegate, false); err != nil {
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
		if reqID, err := _SDK.ExecuteCommand(int64(msg.C_MessagesGetDialogs), reqBytes, reqDelegate, false); err != nil {
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
		req.Peer = &msg.InputPeer{
			Type:       msg.PeerType_PeerUser,
			AccessHash: 0,
		}
		for {
			c.Print("Peer ID: ")
			peerID, err := strconv.ParseInt(c.ReadLine(), 10, 64)
			req.Peer.ID = peerID
			if err == nil {
				break
			} else {
				c.Println(err.Error())
			}
		}

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(int64(msg.C_MessagesGetDialog), reqBytes, reqDelegate, false); err != nil {
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
		req.Peer.Type = msg.PeerType_PeerUser
		for {
			c.Print("Peer ID: ")
			peerID, err := strconv.ParseInt(c.ReadLine(), 10, 64)
			req.Peer.ID = peerID
			if err == nil {
				break
			} else {
				c.Println(err.Error())
			}
		}
		for {
			c.Print("Access Hash: ")
			accessHash, err := strconv.ParseUint(c.ReadLine(), 10, 64)
			req.Peer.AccessHash = uint64(accessHash)
			if err == nil {
				break
			} else {
				c.Println(err.Error())
			}
		}
		for {
			c.Print("Max ID: ")
			maxID, err := strconv.ParseInt(c.ReadLine(), 10, 64)
			req.MaxID = maxID
			if err == nil {
				break
			} else {
				c.Println(err.Error())
			}
		}
		for {
			c.Print("Min ID: ")
			minID, err := strconv.ParseInt(c.ReadLine(), 10, 64)
			req.MinID = minID
			if err == nil {
				break
			} else {
				c.Println(err.Error())
			}
		}
		for {
			c.Print("Limit: ")
			limit, err := strconv.ParseInt(c.ReadLine(), 10, 64)
			req.Limit = int32(limit)
			if err == nil {
				break
			} else {
				c.Println(err.Error())
			}
		}

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(int64(msg.C_MessagesGetHistory), reqBytes, reqDelegate, false); err != nil {
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
		req.Peer.Type = msg.PeerType_PeerUser
		for {
			c.Print("Peer ID: ")
			peerID, err := strconv.ParseInt(c.ReadLine(), 10, 64)
			req.Peer.ID = peerID
			if err == nil {
				break
			} else {
				c.Println(err.Error())
			}
		}
		for {
			c.Print("Access Hash: ")
			accessHash, err := strconv.ParseUint(c.ReadLine(), 10, 64)
			req.Peer.AccessHash = uint64(accessHash)
			if err == nil {
				break
			} else {
				c.Println(err.Error())
			}
		}
		for {
			c.Print("Max ID: ")
			maxID, err := strconv.ParseInt(c.ReadLine(), 10, 64)
			req.MaxID = maxID
			if err == nil {
				break
			} else {
				c.Println(err.Error())
			}
		}
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(int64(msg.C_MessagesReadHistory), reqBytes, reqDelegate, false); err != nil {
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
		req.Peer.Type = msg.PeerType_PeerUser
		for {
			c.Print("Peer ID: ")
			peerID, err := strconv.ParseInt(c.ReadLine(), 10, 64)
			req.Peer.ID = peerID
			if err == nil {
				break
			} else {
				c.Println(err.Error())
			}
		}
		for {
			c.Print("Access Hash: ")
			accessHash, err := strconv.ParseUint(c.ReadLine(), 10, 64)
			req.Peer.AccessHash = uint64(accessHash)
			if err == nil {
				break
			} else {
				c.Println(err.Error())
			}
		}
		for {
			c.Print("Action (0:Typing, 4:Cancel): ")
			actionID, err := strconv.ParseInt(c.ReadLine(), 10, 32)
			req.Action = msg.TypingAction(actionID)
			if err == nil {
				break
			} else {
				c.Println(err.Error())
			}
		}
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(int64(msg.C_MessagesSetTyping), reqBytes, reqDelegate, false); err != nil {
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
		req.Peer.Type = msg.PeerType_PeerUser
		req.MessagesIDs = make([]int64, 0)
		for {
			c.Print("Peer ID: ")
			peerID, err := strconv.ParseInt(c.ReadLine(), 10, 64)
			req.Peer.ID = peerID
			if err == nil {
				break
			} else {
				c.Println(err.Error())
			}
		}
		for {
			c.Print("Access Hash: ")
			accessHash, err := strconv.ParseUint(c.ReadLine(), 10, 64)
			req.Peer.AccessHash = uint64(accessHash)
			if err == nil {
				break
			} else {
				c.Println(err.Error())
			}
		}
		for {

			c.Print(len(req.MessagesIDs), "Enter none numeric character to break\r\n")
			c.Print(len(req.MessagesIDs), "MessageID: ")
			msgID, err := strconv.ParseInt(c.ReadLine(), 10, 64)
			if err != nil {
				break
			} else {
				req.MessagesIDs = append(req.MessagesIDs, msgID)
			}
		}
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(int64(msg.C_MessagesGet), reqBytes, reqDelegate, false); err != nil {
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
}
