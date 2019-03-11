package main

import (
	"git.ronaksoftware.com/ronak/riversdk/logs"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"go.uber.org/zap"
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
			logs.Error("ExecuteCommand failed", zap.Error(err))
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
			logs.Error("ExecuteCommand failed", zap.Error(err))
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
			logs.Error("ExecuteCommand failed", zap.Error(err))
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
			logs.Error("ExecuteCommand failed", zap.Error(err))
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var AuthLogout = &ishell.Cmd{
	Name: "Logout",
	Func: func(c *ishell.Context) {
		if _, err := _SDK.Logout(true, 0); err != nil {
			logs.Error("ExecuteCommand failed", zap.Error(err))
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
			logs.Error("ExecuteCommand failed", zap.Error(err))
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var AuthLoginByToken = &ishell.Cmd{
	Name: "LoginByToken",
	Func: func(c *ishell.Context) {
		req := msg.AuthLoginByToken{}
		req.Provider = fnGetProvider(c)
		req.Token = fnGetToken(c)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_AuthLoginByToken, reqBytes, reqDelegate, false, false); err != nil {
			logs.Error("ExecuteCommand failed", zap.Error(err))
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
	Auth.AddCmd(AuthLoginByToken)
}
