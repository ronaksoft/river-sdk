package main

import (
	"encoding/hex"
	"fmt"
	"mime"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/logs"
	"git.ronaksoftware.com/ronak/riversdk/repo"
	"go.uber.org/zap"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"github.com/kr/pretty"

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
				logs.Error("ExecuteCommand failed", zap.Error(err))
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
				logs.Error("ExecuteCommand failed", zap.Error(err))
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
				logs.Error("ExecuteCommand failed", zap.Error(err))
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
			logs.Error("ExecuteCommand failed", zap.Error(err))
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
			logs.Error("ExecuteCommand failed", zap.Error(err))
		} else {
			reqDelegate.RequestID = reqID
		}

		_SDK.RemoveRealTimeRequest(msg.C_MessageContainer)
	},
}

var DebuncerStatus = &ishell.Cmd{
	Name: "DebuncerStatus",
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
			logs.Error("ExecuteCommand failed", zap.Error(err))
		} else {
			reqDelegate.RequestID = reqID
		}

		time.Sleep(200 * time.Millisecond)

	},
}

var SearchInDialogs = &ishell.Cmd{
	Name: "SearchInDialogs",
	Func: func(c *ishell.Context) {
		searchPhrase := fnGetTitle(c)
		reqDelegate := new(RequestDelegate)
		_SDK.SearchInDialogs(domain.SequentialUniqueID(), searchPhrase, reqDelegate)
	},
}

var GetGroupInputPeer = &ishell.Cmd{
	Name: "GetGroupInputPeer",
	Func: func(c *ishell.Context) {
		groupID := fnGetGroupID(c)
		userID := fnGetPeerID(c)
		reqDelegate := new(RequestDelegate)
		_SDK.GetGroupInputUser(domain.SequentialUniqueID(), groupID, userID, reqDelegate)
	},
}

var UpdateNewMessageHexString = &ishell.Cmd{
	Name: "UpdateNewMessageHexString",
	Func: func(c *ishell.Context) {
		str := fnGetUpdateNewMessageHexString(c)
		buff, err := hex.DecodeString(str)
		if err != nil {
			c.Println("Error : ", err)
			return
		}

		udp := new(msg.UpdateNewMessage)
		err = udp.Unmarshal(buff)
		if err != nil {
			c.Println("Error : ", err)
			return
		}

		fmt.Printf("\r\n\r\n\r\n%# v\r\n\r\n\r\n", pretty.Formatter(udp))
		switch udp.Message.MediaType {
		case msg.MediaTypeDocument:
			x := new(msg.MediaDocument)
			err = x.Unmarshal(udp.Message.Media)
			if err == nil {
				fmt.Printf("\r\n%# v\r\n", pretty.Formatter(x))

				for _, att := range x.Doc.Attributes {
					switch att.Type {
					case msg.AttributeTypeAudio:
						attrib := new(msg.DocumentAttributeAudio)
						err = attrib.Unmarshal(att.Data)
						if err == nil {
							fmt.Printf("\r\n%# v\r\n", pretty.Formatter(attrib))
						}
					case msg.AttributeTypeVideo:
						attrib := new(msg.DocumentAttributeVideo)
						err = attrib.Unmarshal(att.Data)
						if err == nil {
							fmt.Printf("\r\n%# v\r\n", pretty.Formatter(attrib))
						}
					case msg.AttributeTypePhoto:
						attrib := new(msg.DocumentAttributePhoto)
						err = attrib.Unmarshal(att.Data)
						if err == nil {
							fmt.Printf("\r\n%# v\r\n", pretty.Formatter(attrib))
						}
					case msg.AttributeTypeFile:
						attrib := new(msg.DocumentAttributeFile)
						err = attrib.Unmarshal(att.Data)
						if err == nil {
							fmt.Printf("\r\n%# v\r\n", pretty.Formatter(attrib))
						}
					}
				}
			}
		case msg.MediaTypePhoto:
			x := new(msg.MediaPhoto)
			err = x.Unmarshal(udp.Message.Media)
			if err == nil {

			}
		}
	},
}

var MimeToExt = &ishell.Cmd{
	Name: "MimeToExt",
	Func: func(c *ishell.Context) {

		mimeType := fnGetMime(c)
		exts, err := mime.ExtensionsByType(mimeType)
		if err != nil {
			c.Println(err)
			return
		}
		for _, ext := range exts {
			c.Println(ext)
		}
	},
}
var PrintMessage = &ishell.Cmd{
	Name: "PrintMessageMedia",
	Func: func(c *ishell.Context) {

		msgID := fnGetMessageID(c)

		m := repo.Ctx().Messages.GetMessage(msgID)

		if m == nil {
			logs.Error("Message Is nil")
			return
		}

		if m.MediaType == msg.MediaTypeDocument {
			// x := new(msg.MediaDocument)
			// err := x.Unmarshal(m.Media)
			// if err != nil {
			// 	log.Error("Pars MediaDocument Failed", zap.Error(err))
			// 	return
			// }
			fmt.Printf("\r\n\r\n\r\n%# v\r\n\r\n\r\n", pretty.Formatter(m))
			x := new(msg.MediaDocument)
			err := x.Unmarshal(m.Media)
			if err == nil {
				fmt.Printf("\r\n%# v\r\n", pretty.Formatter(x))

				for _, att := range x.Doc.Attributes {
					switch att.Type {
					case msg.AttributeTypeAudio:
						attrib := new(msg.DocumentAttributeAudio)
						err = attrib.Unmarshal(att.Data)
						if err == nil {
							fmt.Printf("\r\n%# v\r\n", pretty.Formatter(attrib))
						}
					case msg.AttributeTypeVideo:
						attrib := new(msg.DocumentAttributeVideo)
						err = attrib.Unmarshal(att.Data)
						if err == nil {
							fmt.Printf("\r\n%# v\r\n", pretty.Formatter(attrib))
						}
					case msg.AttributeTypePhoto:
						attrib := new(msg.DocumentAttributePhoto)
						err = attrib.Unmarshal(att.Data)
						if err == nil {
							fmt.Printf("\r\n%# v\r\n", pretty.Formatter(attrib))
						}
					case msg.AttributeTypeFile:
						attrib := new(msg.DocumentAttributeFile)
						err = attrib.Unmarshal(att.Data)
						if err == nil {
							fmt.Printf("\r\n%# v\r\n", pretty.Formatter(attrib))
						}
					}
				}
			}
		}

	},
}

func init() {
	Debug.AddCmd(SendTyping)
	Debug.AddCmd(MessageSendByNetwork)
	Debug.AddCmd(MessageSendByQueue)
	Debug.AddCmd(ContactImportByNetwork)
	Debug.AddCmd(MessageSendBulk)
	Debug.AddCmd(DebuncerStatus)
	Debug.AddCmd(GetTopMessageID)
	Debug.AddCmd(ContactImportMany)

	Debug.AddCmd(SearchInDialogs)
	Debug.AddCmd(GetGroupInputPeer)

	Debug.AddCmd(UpdateNewMessageHexString)

	Debug.AddCmd(MimeToExt)
	Debug.AddCmd(PrintMessage)
}
