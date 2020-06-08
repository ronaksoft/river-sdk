package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"mime"
	"os"
	"strings"
	"sync"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"github.com/kr/pretty"

	"git.ronaksoftware.com/river/msg/msg"
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
			if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesSetTyping, reqBytes, reqDelegate); err != nil {
				c.Println("Command Failed:", err)
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
		if reqID, err := _SDK.ExecuteCommand(msg.C_ContactsImport, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}

		time.Sleep(200 * time.Millisecond)

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

var LogoutLoop = &ishell.Cmd{
	Name: "LogoutLoop",
	Func: func(c *ishell.Context) {
		phone := fnGetPhone(c)
		for {
			sendCode(c, phone)
			login(c, phone)
			time.Sleep(time.Second * 3)
			wg := sync.WaitGroup{}
			for i := 0 ; i < 3; i++ {
				wg.Add(1)
				go func() {
					logout()
					wg.Done()
				}()
			}
			wg.Wait()
			time.Sleep(time.Second * 3)
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

func init() {
	Debug.AddCmd(SendTyping)
	Debug.AddCmd(ContactImportMany)
	Debug.AddCmd(UpdateNewMessageHexString)
	Debug.AddCmd(MimeToExt)
	Debug.AddCmd(PrintMessage)
	Debug.AddCmd(LogoutLoop)
}
