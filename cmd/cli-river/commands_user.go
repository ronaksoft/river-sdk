package main

import (
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"go.uber.org/zap"
	"gopkg.in/abiosoft/ishell.v2"
)

var User = &ishell.Cmd{
	Name: "User",
}

var UsersGet = &ishell.Cmd{
	Name: "UsersGet",
	Func: func(c *ishell.Context) {
		// for just one user
		req := msg.UsersGet{}
		req.Users = fnGetInputUser(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_UsersGet, reqBytes, reqDelegate, false, false); err != nil {
			_Log.Error("ExecuteCommand failed", zap.Error(err))
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var UsersGetFull = &ishell.Cmd{
	Name: "UsersGetFull",
	Func: func(c *ishell.Context) {
		// for just one user
		req := msg.UsersGetFull{}
		req.Users = fnGetInputUser(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_UsersGetFull, reqBytes, reqDelegate, false, false); err != nil {
			_Log.Error("ExecuteCommand failed", zap.Error(err))
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var SetLangCode = &ishell.Cmd{
	Name: "SetLangCode",
	Func: func(c *ishell.Context) {
		// for just one user
		req := msg.AccountSetLang{}

		code := fnGetLangCode(c)

		switch code {
		case "en":
			req.LangCode = msg.LangCode_LangCodeEn
		case "fa":
			req.LangCode = msg.LangCode_LangCodeFa
		default:
			req.LangCode = msg.LangCode_LangCodeUnknown
		}

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_AccountSetLang, reqBytes, reqDelegate, false, false); err != nil {
			_Log.Error("ExecuteCommand failed", zap.Error(err))
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

func init() {
	User.AddCmd(UsersGet)
	User.AddCmd(UsersGetFull)
	User.AddCmd(SetLangCode)
}
