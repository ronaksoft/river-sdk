package main

import (
	"fmt"
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"go.uber.org/zap/zapcore"
	"gopkg.in/abiosoft/ishell.v2"
	"io/ioutil"
	"mime"
	"os"
	"strings"
	"sync"
	"time"
)

var Debug = &ishell.Cmd{
	Name: "Debug",
}

func init() {
	Debug.AddCmd(DebugSendTyping)
	Debug.AddCmd(DebugContactImportMany)
	Debug.AddCmd(DebugMimeToExt)
	Debug.AddCmd(DebugLogoutLoop)
	Debug.AddCmd(SetUpdateState)
	Debug.AddCmd(DebugConcurrent)
	Debug.AddCmd(DebugReconnect)
}

var DebugConcurrent = &ishell.Cmd{
	Name: "Concurrent",
	Func: func(c *ishell.Context) {
		for j := 0; j < 20; j++ {
			wg := sync.WaitGroup{}
			for i := 0; i < 20; i++ {
				wg.Add(1)
				go func() {
					req := msg.ContactsGetTopPeers{
						Offset:   0,
						Limit:    100,
						Category: msg.TopPeerCategory_Users,
					}
					reqBytes, _ := req.Marshal()
					reqDelegate := new(RequestDelegate)
					if reqID, err := _SDK.ExecuteCommand(msg.C_ContactsGetTopPeers, reqBytes, reqDelegate); err != nil {
						c.Println("Command Failed:", err)
					} else {
						reqDelegate.RequestID = reqID
					}
					wg.Done()
				}()
			}
			wg.Wait()
		}
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
		_Log.SetLogLevel(int(zapcore.WarnLevel))
		phone := fnGetPhone(c)
		for {
			if _SDK.ConnInfo.UserID == 0 {
				c.Println("Sending Code")
				sendCode(c, phone)
				c.Println("Sending Login")
				login(c, phone)
				time.Sleep(time.Second * 3)
			}
			recall()
			c.Println("Sending Logout")
			logout()
			if _SDK.ConnInfo.UserID != 0 {
				c.Println("WRONG!!!!!!!!!! We are not logout correctly")
			}
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
	reqDelegate.FlagsVal = domain.RequestBlocking
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
	reqDelegate.FlagsVal = domain.RequestBlocking
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

var DebugMimeToExt = &ishell.Cmd{
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

var SetUpdateState = &ishell.Cmd{
	Name: "SetUpdateState",
	Func: func(c *ishell.Context) {
		updateID := fnGetFromUpdateID(c)
		_SDK.SetUpdateState(updateID)
	},
}
