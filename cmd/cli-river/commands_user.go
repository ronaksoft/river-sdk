package main

import (
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
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
			logs.Error("ExecuteCommand failed", zap.Error(err))
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
			logs.Error("ExecuteCommand failed", zap.Error(err))
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

func init() {
	User.AddCmd(UsersGet)
	User.AddCmd(UsersGetFull)
}
