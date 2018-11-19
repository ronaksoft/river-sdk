package main

import (
	"git.ronaksoftware.com/ronak/riversdk/msg"
	ishell "gopkg.in/abiosoft/ishell.v2"
)

var Group = &ishell.Cmd{
	Name: "Group",
}

var Create = &ishell.Cmd{
	Name: "Create",
	Func: func(c *ishell.Context) {
		req := msg.MessagesCreateGroup{}
		req.Title = fnGetTitle(c)
		req.Users = fnGetInputUser(c)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesCreateGroup, reqBytes, reqDelegate, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

func init() {
	Group.AddCmd(Create)
}
