package main

import (
	"git.ronaksoftware.com/ronak/riversdk/msg"
	ishell "gopkg.in/abiosoft/ishell.v2"
)

var Account = &ishell.Cmd{
	Name: "Account",
}

var RegisterDevice = &ishell.Cmd{
	Name: "RegisterDevice",
	Func: func(c *ishell.Context) {
		req := msg.AccountRegisterDevice{}
		req.TokenType = 0
		req.Token = "token"
		req.DeviceModel = "River CLI"
		req.SystemVersion = "v0.0.013"
		req.AppVersion = "v0.2.3"
		req.LangCode = "en-us"

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)

		if reqID, err := _SDK.ExecuteCommand(int64(msg.C_AccountRegisterDevice), reqBytes, reqDelegate, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var UpdateUsername = &ishell.Cmd{
	Name: "UpdateUsername",
	Func: func(c *ishell.Context) {
		req := msg.AccountUpdateUsername{}
		req.Username = fnGetUsername(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)

		if reqID, err := _SDK.ExecuteCommand(int64(msg.C_AccountUpdateUsername), reqBytes, reqDelegate, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var CheckUsername = &ishell.Cmd{
	Name: "CheckUsername",
	Func: func(c *ishell.Context) {
		req := msg.AccountCheckUsername{}
		req.Username = fnGetUsername(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)

		if reqID, err := _SDK.ExecuteCommand(int64(msg.C_AccountCheckUsername), reqBytes, reqDelegate, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var UnregisterDevice = &ishell.Cmd{
	Name: "UnregisterDevice",
	Func: func(c *ishell.Context) {
		req := msg.AccountUnregisterDevice{}
		req.TokenType = 1
		req.Token = "token"

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)

		if reqID, err := _SDK.ExecuteCommand(int64(msg.C_AccountUnregisterDevice), reqBytes, reqDelegate, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var UpdateProfile = &ishell.Cmd{
	Name: "UpdateProfile",
	Func: func(c *ishell.Context) {
		req := msg.AccountUpdateProfile{}
		req.FirstName = fnGetFirstName(c)
		req.LastName = fnGetLastName(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)

		if reqID, err := _SDK.ExecuteCommand(int64(msg.C_AccountUpdateProfile), reqBytes, reqDelegate, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

func init() {
	Account.AddCmd(RegisterDevice)
	Account.AddCmd(UpdateUsername)
	Account.AddCmd(CheckUsername)
	Account.AddCmd(UnregisterDevice)
	Account.AddCmd(UpdateProfile)
}
