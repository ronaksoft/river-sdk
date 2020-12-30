package main

import (
	"encoding/hex"
	"fmt"
	"git.ronaksoft.com/river/sdk/internal/logs"
	riversdk "git.ronaksoft.com/river/sdk/sdk/prime"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"mime"
	"os"
	"strings"
	"sync"
	"time"

	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"github.com/kr/pretty"

	"git.ronaksoft.com/river/msg/go/msg"
	"gopkg.in/abiosoft/ishell.v2"
)

var Debug = &ishell.Cmd{
	Name: "Debug",
}

func init() {
	Debug.AddCmd(DebugSendTyping)
	Debug.AddCmd(DebugContactImportMany)
	Debug.AddCmd(UpdateNewMessageHexString)
	Debug.AddCmd(MimeToExt)
	Debug.AddCmd(PrintMessage)
	Debug.AddCmd(DebugLogoutLoop)
	Debug.AddCmd(SendMessage)
	Debug.AddCmd(DebugConcurrent)
	Debug.AddCmd(DebugReconnect)
}

var DebugConcurrent = &ishell.Cmd{
	Name: "Concurrent",
	Func: func(c *ishell.Context) {
		wg := sync.WaitGroup{}
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func() {
				req := msg.SystemGetServerTime{}
				reqBytes, _ := req.Marshal()
				reqDelegate := new(RequestDelegate)
				reqDelegate.FlagsVal = riversdk.RequestDontWaitForNetwork
				if reqID, err := _SDK.ExecuteCommand(msg.C_SystemGetServerTime, reqBytes, reqDelegate); err != nil {
					c.Println("Command Failed:", err)
				} else {
					reqDelegate.RequestID = reqID
				}
				wg.Done()
			}()
		}
		wg.Wait()
	},
}

var DebugReconnect = &ishell.Cmd{
	Name: "Reconnect",
	Func: func(c *ishell.Context) {
		for i := 0; i < 10; i++ {
			_SDK.StopNetwork()
			_SDK.StartNetwork("IR")
		}

	},
}

var DebugSendTyping = &ishell.Cmd{
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
				req.Action = msg.TypingAction_TypingActionTyping
			} else {
				req.Action = msg.TypingAction_TypingActionCancel
			}

			reqBytes, _ := req.Marshal()
			reqDelegate := new(RequestDelegate)
			if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesSetTyping, reqBytes, reqDelegate); err != nil {
				c.Println("Command Failed:", err)
			} else {
				reqDelegate.RequestID = reqID
			}
		}

	},
}

var DebugContactImportMany = &ishell.Cmd{
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
		if reqID, err := _SDK.ExecuteCommand(msg.C_ContactsImport, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}

		time.Sleep(200 * time.Millisecond)

	},
}

var DebugLogoutLoop = &ishell.Cmd{
	Name: "LogoutLoop",
	Func: func(c *ishell.Context) {
		logs.SetLogLevel(int(zapcore.WarnLevel))
		phone := fnGetPhone(c)
		for {
			if _SDK.ConnInfo.UserID == 0 {
				c.Println("Sending Code")
				sendCode(c, phone)
				c.Println("Sending Login")
				login(c, phone)
				time.Sleep(time.Second * 3)
			}
			c.Println("Sending Logout")
			go recall()
			wg := sync.WaitGroup{}
			for i := 0; i < 3; i++ {
				wg.Add(1)
				go func() {
					logout()
					wg.Done()
				}()
			}
			wg.Wait()
			if _SDK.ConnInfo.UserID != 0 {
				c.Println("WRONG!!!!!!!!!! We are not logout correctly")
			}
			time.Sleep(time.Second * 3)
		}
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
		case msg.MediaType_MediaTypeDocument:
			x := new(msg.MediaDocument)
			err = x.Unmarshal(udp.Message.Media)
			if err == nil {
				fmt.Printf("\r\n%# v\r\n", pretty.Formatter(x))

				for _, att := range x.Doc.Attributes {
					switch att.Type {
					case msg.DocumentAttributeType_AttributeTypeAudio:
						attrib := new(msg.DocumentAttributeAudio)
						err = attrib.Unmarshal(att.Data)
						if err == nil {
							fmt.Printf("\r\n%# v\r\n", pretty.Formatter(attrib))
						}
					case msg.DocumentAttributeType_AttributeTypeVideo:
						attrib := new(msg.DocumentAttributeVideo)
						err = attrib.Unmarshal(att.Data)
						if err == nil {
							fmt.Printf("\r\n%# v\r\n", pretty.Formatter(attrib))
						}
					case msg.DocumentAttributeType_AttributeTypePhoto:
						attrib := new(msg.DocumentAttributePhoto)
						err = attrib.Unmarshal(att.Data)
						if err == nil {
							fmt.Printf("\r\n%# v\r\n", pretty.Formatter(attrib))
						}
					case msg.DocumentAttributeType_AttributeTypeFile:
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
		m, _ := repo.Messages.Get(msgID)
		if m == nil {
			c.Println("Message is nil")
			return
		}

		if m.MediaType == msg.MediaType_MediaTypeDocument {
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
					case msg.DocumentAttributeType_AttributeTypeAudio:
						attrib := new(msg.DocumentAttributeAudio)
						err = attrib.Unmarshal(att.Data)
						if err == nil {
							fmt.Printf("\r\n%# v\r\n", pretty.Formatter(attrib))
						}
					case msg.DocumentAttributeType_AttributeTypeVideo:
						attrib := new(msg.DocumentAttributeVideo)
						err = attrib.Unmarshal(att.Data)
						if err == nil {
							fmt.Printf("\r\n%# v\r\n", pretty.Formatter(attrib))
						}
					case msg.DocumentAttributeType_AttributeTypePhoto:
						attrib := new(msg.DocumentAttributePhoto)
						err = attrib.Unmarshal(att.Data)
						if err == nil {
							fmt.Printf("\r\n%# v\r\n", pretty.Formatter(attrib))
						}
					case msg.DocumentAttributeType_AttributeTypeFile:
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



func sendCode(c *ishell.Context, phone string) {
	req := msg.AuthSendCode{
		Phone: phone,
	}
	reqBytes, _ := req.Marshal()
	reqDelegate := new(RequestDelegate)
	if reqID, err := _SDK.ExecuteCommand(msg.C_AuthSendCode, reqBytes, reqDelegate); err != nil {
		c.Println("Command Failed:", err)
	} else {
		reqDelegate.RequestID = reqID
	}
}
func login(c *ishell.Context, phone string) {
	req := msg.AuthLogin{}
	phoneFile, err := os.Open("./_phone")
	if err != nil {
		return
	} else {
		b, _ := ioutil.ReadAll(phoneFile)
		req.Phone = string(b)
		if strings.HasPrefix(req.Phone, "2374") {
			File, err := os.Open("./_phoneCodeHash")
			if err != nil {
				return
			} else {
				req.PhoneCode = req.Phone[len(req.Phone)-5:]
				b, _ := ioutil.ReadAll(File)
				req.PhoneCodeHash = string(b)
			}
		}
	}
	reqBytes, _ := req.Marshal()
	reqDelegate := new(RequestDelegate)
	os.Remove("./_phone")
	os.Remove("./_phoneCodeHash")
	if reqID, err := _SDK.ExecuteCommand(msg.C_AuthLogin, reqBytes, reqDelegate); err != nil {
		c.Println("Command Failed:", err)
	} else {
		reqDelegate.RequestID = reqID
	}
}
func logout() {
	_SDK.Logout(true, 0)
}
func recall() {
	req := msg.AuthRecall{}
	reqBytes, _ := req.Marshal()
	reqDelegate := new(RequestDelegate)
	if reqID, err := _SDK.ExecuteCommand(msg.C_AuthRecall, reqBytes, reqDelegate); err == nil {
		reqDelegate.RequestID = reqID
	}
}

var SendMessage = &ishell.Cmd{
	Name: "SendMessage",
	Func: func(c *ishell.Context) {
		for i := 0; i < 10; i++ {
			req := msg.MessagesSend{}
			req.ClearDraft = true
			req.RandomID = domain.RandomInt63()
			req.Peer = &msg.InputPeer{}
			req.Peer.Type = msg.PeerType_PeerUser
			req.Peer.ID = _SDK.ConnInfo.UserID
			req.Peer.AccessHash = 0
			req.Body = fmt.Sprintf("Text: %s", domain.RandomID(32))
			req.Entities = nil
			reqBytes, _ := req.Marshal()
			reqDelegate := NewCustomDelegate()
			if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesSend, reqBytes, reqDelegate); err != nil {
				c.Println("Command Failed:", err)
			} else {
				reqDelegate.RequestID = reqID
			}
			time.Sleep(time.Second)
		}
	},
}
