package main

import (
	"fmt"
	"strconv"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/domain"

	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/toolbox"

	"gopkg.in/abiosoft/ishell.v2"
)

var Debug = &ishell.Cmd{
	Name: "Debug",
}

var SendTyping = &ishell.Cmd{
	Name: "SendTyping",
	Func: func(c *ishell.Context) {
		// for just one user
		req := msg.MessagesSetTyping{}
		req.Peer = &msg.InputPeer{}
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

		var count int
		for {
			c.Print("Tries : ")
			tmp, err := strconv.ParseInt(c.ReadLine(), 10, 32)
			if err == nil {
				count = int(tmp)
				break
			} else {
				c.Println(err.Error())
			}
		}
		var interval time.Duration
		for {
			c.Print("Interval: ")
			tmp, err := strconv.ParseInt(c.ReadLine(), 10, 32)

			if err == nil {
				interval = time.Duration(tmp) * time.Millisecond
				break
			} else {
				c.Println(err.Error())
			}
		}
		for i := 0; i < count; i++ {
			time.Sleep(interval)
			if i%2 == 0 {
				req.Action = msg.TypingAction_Typing
			} else {
				req.Action = msg.TypingAction_Cancel
			}

			reqBytes, _ := req.Marshal()
			reqDelegate := new(RequestDelegate)
			if reqID, err := _SDK.ExecuteCommand(int64(msg.C_MessagesSetTyping), reqBytes, reqDelegate, false, true); err != nil {
				_Log.Debug(err.Error())
			} else {
				reqDelegate.RequestID = reqID
			}
		}

	},
}

var MessageSendByNetwork = &ishell.Cmd{
	Name: "MessageSendByNetwork",
	Func: func(c *ishell.Context) {
		// for just one user
		req := msg.MessagesSend{}
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

		var count int
		for {
			c.Print("Tries : ")
			tmp, err := strconv.ParseInt(c.ReadLine(), 10, 32)
			if err == nil {
				count = int(tmp)
				break
			} else {
				c.Println(err.Error())
			}
		}
		var interval time.Duration
		for {
			c.Print("Interval: ")
			tmp, err := strconv.ParseInt(c.ReadLine(), 10, 32)

			if err == nil {
				interval = time.Duration(tmp) * time.Millisecond
				break
			} else {
				c.Println(err.Error())
			}
		}

		_SDK.AddRealTimeRequest(msg.C_MessagesSend)
		for i := 0; i < count; i++ {
			time.Sleep(interval)
			req.RandomID = ronak.RandomInt64(0)
			req.Body = fmt.Sprintf("Test Msg [%v]", i)

			reqBytes, _ := req.Marshal()
			reqDelegate := new(RequestDelegate)
			if reqID, err := _SDK.ExecuteCommand(int64(msg.C_MessagesSend), reqBytes, reqDelegate, false, true); err != nil {
				_Log.Debug(err.Error())
			} else {
				reqDelegate.RequestID = reqID
			}
		}
		_SDK.RemoveRealTimeRequest(msg.C_MessagesSend)
	},
}

var MessageSendByQueue = &ishell.Cmd{
	Name: "MessageSendByQueue",
	Func: func(c *ishell.Context) {
		// for just one user
		req := msg.MessagesSend{}
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

		var count int
		for {
			c.Print("Tries : ")
			tmp, err := strconv.ParseInt(c.ReadLine(), 10, 32)
			if err == nil {
				count = int(tmp)
				break
			} else {
				c.Println(err.Error())
			}
		}
		var interval time.Duration
		for {
			c.Print("Interval: ")
			tmp, err := strconv.ParseInt(c.ReadLine(), 10, 32)

			if err == nil {
				interval = time.Duration(tmp) * time.Millisecond
				break
			} else {
				c.Println(err.Error())
			}
		}
		// make sure
		_SDK.RemoveRealTimeRequest(msg.C_MessagesSend)
		for i := 0; i < count; i++ {
			time.Sleep(interval)
			req.RandomID = ronak.RandomInt64(0)
			req.Body = fmt.Sprintf("Test Msg [%v]", i)

			reqBytes, _ := req.Marshal()
			reqDelegate := new(RequestDelegate)
			if reqID, err := _SDK.ExecuteCommand(int64(msg.C_MessagesSend), reqBytes, reqDelegate, false, true); err != nil {
				_Log.Debug(err.Error())
			} else {
				reqDelegate.RequestID = reqID
			}
		}
	},
}

var ContactImportByNetwork = &ishell.Cmd{
	Name: "ContactImportByNetwork",
	Func: func(c *ishell.Context) {
		req := msg.ContactsImport{}
		req.Replace = true
		contact := msg.PhoneContact{}
		c.Print("First Name:")
		contact.FirstName = c.ReadLine()
		c.Print("Last Name:")
		contact.LastName = c.ReadLine()
		c.Print("Phone: ")
		contact.Phone = c.ReadLine()
		contact.ClientID = ronak.RandomInt64(0)
		req.Contacts = append(req.Contacts, &contact)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)

		// var requestID int64
		// for {
		// 	c.Print("RequestID : ")
		// 	tmp, err := strconv.ParseInt(c.ReadLine(), 10, 32)
		// 	if err == nil {
		// 		requestID = tmp
		// 		break
		// 	} else {
		// 		c.Println(err.Error())
		// 	}
		// }

		_SDK.RemoveRealTimeRequest(msg.C_ContactsImport)

		if reqID, err := _SDK.ExecuteCommand(int64(msg.C_ContactsImport), reqBytes, reqDelegate, true, true); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var MessageSendBulk = &ishell.Cmd{
	Name: "MessageSendBulk",
	Func: func(c *ishell.Context) {
		// for just one user

		peerID := int64(0)
		for {
			c.Print("Peer ID: ")
			pID, err := strconv.ParseInt(c.ReadLine(), 10, 64)
			if err == nil {
				peerID = pID
				break
			} else {
				c.Println(err.Error())
			}
		}
		accessHash := uint64(0)
		for {
			c.Print("Access Hash: ")
			acHash, err := strconv.ParseUint(c.ReadLine(), 10, 64)
			if err == nil {
				accessHash = acHash
				break
			} else {
				c.Println(err.Error())
			}
		}

		var count int
		for {
			c.Print("Tries : ")
			tmp, err := strconv.ParseInt(c.ReadLine(), 10, 32)
			if err == nil {
				count = int(tmp)
				break
			} else {
				c.Println(err.Error())
			}
		}

		_SDK.AddRealTimeRequest(msg.C_MessageContainer)

		msgs := make([]*msg.MessageEnvelope, count)
		for i := 0; i < count; i++ {
			req := &msg.MessagesSend{}
			req.Peer = &msg.InputPeer{}
			req.Peer.ID = peerID
			req.Peer.AccessHash = accessHash
			req.Peer.Type = msg.PeerType_PeerUser
			req.RandomID = ronak.RandomInt64(0)
			req.Body = fmt.Sprintf("Test Msg [%v]", i)

			buff, _ := req.Marshal()
			msgEnvelop := &msg.MessageEnvelope{
				Constructor: msg.C_MessagesSend,
				Message:     buff,
				RequestID:   uint64(domain.SequentialUniqueID()),
			}
			msgs[i] = msgEnvelop
		}

		msgContainer := new(msg.MessageContainer)
		msgContainer.Envelopes = msgs
		msgContainer.Length = int32(len(msgs))
		reqBytes, _ := msgContainer.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(int64(msg.C_MessageContainer), reqBytes, reqDelegate, false, true); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}

		_SDK.RemoveRealTimeRequest(msg.C_MessageContainer)
	},
}

func init() {
	Debug.AddCmd(SendTyping)
	Debug.AddCmd(MessageSendByNetwork)
	Debug.AddCmd(MessageSendByQueue)
	Debug.AddCmd(ContactImportByNetwork)
	Debug.AddCmd(MessageSendBulk)
}
