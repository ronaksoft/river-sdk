package main

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"gopkg.in/abiosoft/ishell.v2"
)

var User = &ishell.Cmd{
	Name: "User",
}

var UserGet = &ishell.Cmd{
	Name: "Get",
	Func: func(c *ishell.Context) {
		// for just one user
		req := msg.UsersGet{}
		req.Users = []*msg.InputUser{fnGetUser(c)}
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_UsersGet, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var UserGetFull = &ishell.Cmd{
	Name: "GetFull",
	Func: func(c *ishell.Context) {
		// for just one user
		req := msg.UsersGetFull{}
		req.Users = []*msg.InputUser{fnGetUser(c)}

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
	User.AddCmd(UserGet)
	User.AddCmd(UserGetFull)
	User.AddCmd(SetLangCode)
}
