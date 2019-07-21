package main

import (
	"encoding/hex"
	"fmt"
	"mime"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"go.uber.org/zap"

	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"github.com/kr/pretty"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
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
				_Log.Error("ExecuteCommand failed", zap.Error(err))
			} else {
				reqDelegate.RequestID = reqID
			}
		}

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
			_Log.Error("ExecuteCommand failed", zap.Error(err))
		} else {
			reqDelegate.RequestID = reqID
		}

		time.Sleep(200 * time.Millisecond)

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

var GetSDKSalt = &ishell.Cmd{
	Name: "GetSDKSalt",
	Func: func(c *ishell.Context) {
		_Log.Info("SDK salt: ", zap.Int64("_SDK.GetSDKSalt", _SDK.GetSDKSalt()))
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

		m := repo.Messages.Get(msgID)

		if m == nil {
			_Log.Error("Message Is nil")
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
	Debug.AddCmd(ContactImportMany)

	Debug.AddCmd(GetGroupInputPeer)

	Debug.AddCmd(UpdateNewMessageHexString)

	Debug.AddCmd(MimeToExt)
	Debug.AddCmd(PrintMessage)
	Debug.AddCmd(GetSDKSalt)
}
