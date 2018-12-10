package main

import (
	"fmt"
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
		req.Peer.ID = fnGetPeerID(c)
		req.Peer.AccessHash = fnGetAccessHash(c)

		count := fnGetTries(c)
		interval := fnGetInterval(c)

		for i := 0; i < count; i++ {
			time.Sleep(interval)
			if i%2 == 0 {
				req.Action = msg.TypingActionTyping
			} else {
				req.Action = msg.TypingActionCancel
			}

			reqBytes, _ := req.Marshal()
			reqDelegate := new(RequestDelegate)
			if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesSetTyping, reqBytes, reqDelegate, false, false); err != nil {
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
		req.Peer.Type = fnGetPeerType(c)
		req.Peer.ID = fnGetPeerID(c)
		req.Peer.AccessHash = fnGetAccessHash(c)

		count := fnGetTries(c)
		interval := fnGetInterval(c)

		_SDK.AddRealTimeRequest(msg.C_MessagesSend)
		for i := 0; i < count; i++ {
			time.Sleep(interval)
			req.RandomID = ronak.RandomInt64(0)
			req.Body = fmt.Sprintf("Test Msg [%v]", i)

			reqBytes, _ := req.Marshal()
			reqDelegate := new(RequestDelegate)
			if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesSend, reqBytes, reqDelegate, false, false); err != nil {
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
		req.Peer.Type = fnGetPeerType(c)
		req.Peer.ID = fnGetPeerID(c)
		req.Peer.AccessHash = fnGetAccessHash(c)

		count := fnGetTries(c)
		interval := fnGetInterval(c)

		// make sure
		_SDK.RemoveRealTimeRequest(msg.C_MessagesSend)
		for i := 0; i < count; i++ {
			time.Sleep(interval)
			req.RandomID = ronak.RandomInt64(0)
			req.Body = fmt.Sprintf("Test Msg [%v]", i)

			reqBytes, _ := req.Marshal()
			reqDelegate := new(RequestDelegate)
			if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesSend, reqBytes, reqDelegate, false, false); err != nil {
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
		contact.FirstName = fnGetFirstName(c)
		contact.LastName = fnGetLastName(c)
		contact.Phone = fnGetPhone(c)

		contact.ClientID = ronak.RandomInt64(0)
		req.Contacts = append(req.Contacts, &contact)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)

		// requestID := fnGetRequestID(c)

		_SDK.RemoveRealTimeRequest(msg.C_ContactsImport)

		if reqID, err := _SDK.ExecuteCommand(msg.C_ContactsImport, reqBytes, reqDelegate, true, false); err != nil {
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

		peerType := fnGetPeerType(c)
		peerID := fnGetPeerID(c)
		accessHash := fnGetAccessHash(c)
		count := fnGetTries(c)

		_SDK.AddRealTimeRequest(msg.C_MessageContainer)

		msgs := make([]*msg.MessageEnvelope, count)
		for i := 0; i < count; i++ {
			req := &msg.MessagesSend{}
			req.Peer = &msg.InputPeer{}
			req.Peer.ID = peerID
			req.Peer.AccessHash = accessHash
			req.Peer.Type = peerType
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
		if reqID, err := _SDK.ExecuteCommand(msg.C_MessageContainer, reqBytes, reqDelegate, false, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}

		_SDK.RemoveRealTimeRequest(msg.C_MessageContainer)
	},
}

var TestORM = &ishell.Cmd{
	Name: "TestORM",
	Func: func(c *ishell.Context) {
		tries := fnGetTries(c)
		_SDK.TestORM(tries)
	},
}
var TestRAW = &ishell.Cmd{
	Name: "TestRAW",
	Func: func(c *ishell.Context) {
		tries := fnGetTries(c)
		_SDK.TestORM(tries)
	},
}
var TestBatch = &ishell.Cmd{
	Name: "TestBatch",
	Func: func(c *ishell.Context) {
		tries := fnGetTries(c)
		_SDK.TestBatch(tries)
	},
}

var PrintDebuncerStatus = &ishell.Cmd{
	Name: "PrintDebuncerStatus",
	Func: func(c *ishell.Context) {
		_SDK.PrintDebuncerStatus()
	},
}

var GetTopMessageID = &ishell.Cmd{
	Name: "GetTopMessageID",
	Func: func(c *ishell.Context) {

		peerID := fnGetPeerID(c)
		peerType := fnGetPeerType(c)
		maxID := _SDK.GetRealTopMessageID(peerID, int32(peerType))
		c.Println("XXXXXXXXXXXXXXXXXXXXXXXXXXXXXX MaxID = ", maxID)
	},
}

var ContactImportMany = &ishell.Cmd{
	Name: "ContactImportMany",
	Func: func(c *ishell.Context) {
		req := msg.ContactsImport{}
		req.Replace = true
		for i := 0; i < 80; i++ {
			txt := fmt.Sprintf("237400%d", 23740010+i)
			contact := msg.PhoneContact{}
			contact.FirstName = txt
			contact.LastName = txt
			contact.Phone = txt
			contact.ClientID = domain.SequentialUniqueID()
			req.Contacts = append(req.Contacts, &contact)
		}
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_ContactsImport, reqBytes, reqDelegate, false, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}

		time.Sleep(200 * time.Millisecond)

	},
}

func init() {
	Debug.AddCmd(SendTyping)
	Debug.AddCmd(MessageSendByNetwork)
	Debug.AddCmd(MessageSendByQueue)
	Debug.AddCmd(ContactImportByNetwork)
	Debug.AddCmd(MessageSendBulk)
	Debug.AddCmd(TestORM)
	Debug.AddCmd(TestRAW)
	Debug.AddCmd(TestBatch)
	Debug.AddCmd(PrintDebuncerStatus)
	Debug.AddCmd(GetTopMessageID)
	Debug.AddCmd(ContactImportMany)
}
