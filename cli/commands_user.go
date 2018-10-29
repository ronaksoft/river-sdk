package main

import (
	"strconv"

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
		req.Users = make([]*msg.InputUser, 0)
		for {
			c.Print("Enter none numeric character to break\r\n")

			c.Print(len(req.Users), "User ID: ")
			userID, err := strconv.ParseInt(c.ReadLine(), 10, 64)
			if err != nil {
				break
			}

			c.Print(len(req.Users), "Access Hash: ")
			accessHash, err := strconv.ParseUint(c.ReadLine(), 10, 64)
			if err != nil {
				break
			}

			u := new(msg.InputUser)
			u.UserID = userID
			u.AccessHash = accessHash
			req.Users = append(req.Users, u)
		}

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(int64(msg.C_UsersGet), reqBytes, reqDelegate, false, true); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

func init() {
	User.AddCmd(UsersGet)
}
