package main

import (
	msg "git.ronaksoftware.com/river/msg/chat"
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
		if reqID, err := _SDK.ExecuteCommand(msg.C_UsersGet, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
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
		if reqID, err := _SDK.ExecuteCommand(msg.C_UsersGetFull, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
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

		if code != "en" && code != "fa" {
			c.Println("Invalid lang code. Using en as default:", code)
			code = "en"
		}
		c.Println("LangCode:", code)
		req.LangCode = code

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_AccountSetLang, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
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
