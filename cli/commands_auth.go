package main

import (
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"gopkg.in/abiosoft/ishell.v2"
)

var Auth = &ishell.Cmd{
	Name: "Auth",
}

var AuthCheckPhone = &ishell.Cmd{
	Name: "CheckPhone",
	Func: func(c *ishell.Context) {
		req := msg.AuthCheckPhone{}
		req.Phone = fnGetPhone(c)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_AuthCheckPhone, reqBytes, reqDelegate, false, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var AuthSendCode = &ishell.Cmd{
	Name: "SendCode",
	Func: func(c *ishell.Context) {
		req := msg.AuthSendCode{}
		req.Phone = fnGetPhone(c)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_AuthSendCode, reqBytes, reqDelegate, false, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var AuthRegister = &ishell.Cmd{
	Name: "Register",
	Func: func(c *ishell.Context) {
		req := msg.AuthRegister{}

		req.Phone = fnGetPhone(c)
		req.PhoneCode = fnGetPhoneCode(c)
		req.PhoneCodeHash = fnGetPhoneCodeHash(c)
		req.FirstName = fnGetFirstName(c)
		req.LastName = fnGetLastName(c)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_AuthRegister, reqBytes, reqDelegate, false, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var AuthLogin = &ishell.Cmd{
	Name: "Login",
	Func: func(c *ishell.Context) {
		req := msg.AuthLogin{}
		req.Phone = fnGetPhone(c)
		req.PhoneCode = fnGetPhoneCode(c)
		req.PhoneCodeHash = fnGetPhoneCodeHash(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_AuthLogin, reqBytes, reqDelegate, false, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var AuthLogout = &ishell.Cmd{
	Name: "Logout",
	Func: func(c *ishell.Context) {
		if _, err := _SDK.Logout(); err != nil {
			_Log.Debug(err.Error())
		}
	},
}

var AuthRecall = &ishell.Cmd{
	Name: "Recall",
	Func: func(c *ishell.Context) {
		req := msg.AuthRecall{}
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_AuthRecall, reqBytes, reqDelegate, false, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

func init() {
	Auth.AddCmd(AuthSendCode)
	Auth.AddCmd(AuthCheckPhone)
	Auth.AddCmd(AuthRegister)
	Auth.AddCmd(AuthLogin)
	Auth.AddCmd(AuthRecall)
	Auth.AddCmd(AuthLogout)
}
