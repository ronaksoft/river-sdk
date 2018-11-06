package main

import (
	"git.ronaksoftware.com/ronak/riversdk/msg"
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
		if reqID, err := _SDK.ExecuteCommand(int64(msg.C_UsersGet), reqBytes, reqDelegate, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

func init() {
	User.AddCmd(UsersGet)
}
