package main

import (
	"git.ronaksoft.com/river/msg/msg"
	"gopkg.in/abiosoft/ishell.v2"
)

var System = &ishell.Cmd{
	Name: "System",
}

var SystemGetConfig = &ishell.Cmd{
	Name: "GetConfig",
	Func: func(c *ishell.Context) {
		req := msg.SystemGetConfig{}
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_SystemGetConfig, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

func init() {
	System.AddCmd(SystemGetConfig)
}
