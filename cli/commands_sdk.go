package main

import (
	"strconv"

	"git.ronaksoftware.com/ronak/riversdk/configs"
	"git.ronaksoftware.com/ronak/riversdk/log"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"gopkg.in/abiosoft/ishell.v2"
)

var SDK = &ishell.Cmd{
	Name: "SDK",
}

var SdkConnInfo = &ishell.Cmd{
	Name: "ConnInfo",
	Func: func(c *ishell.Context) {
		c.Println("UserID:", configs.Get().UserID)
		c.Println("AuthID:", configs.Get().AuthID)
		c.Println("Phone:", configs.Get().Phone)
		c.Println("FirstName:", configs.Get().FirstName)
		c.Println("LastName:", configs.Get().LastName)
		c.Println("AuthKey:", configs.Get().AuthKey)
	},
}

var SdkSetLogLevel = &ishell.Cmd{
	Name: "SetLogLevel",
	Func: func(c *ishell.Context) {
		choiceIndex := c.MultiChoice([]string{
			"Debug", "Info", "Warn", "Error",
		}, "Level")
		log.SetLogLevel(choiceIndex - 1)
	},
}

var SdkGetDiffrence = &ishell.Cmd{
	Name: "GetDiffrence",
	Func: func(c *ishell.Context) {
		req := msg.UpdateGetDifference{}
		for {
			c.Print("Limit: ")
			limit, err := strconv.ParseInt(c.ReadLine(), 10, 32)
			req.Limit = int32(limit)
			if err == nil {
				break
			} else {
				c.Println(err.Error())
			}
		}
		for {
			c.Print("From UpdateID: ")
			fromUpdateID, err := strconv.ParseInt(c.ReadLine(), 10, 64)
			req.From = fromUpdateID
			if err == nil {
				break
			} else {
				c.Println(err.Error())
			}
		}

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(int64(msg.C_UpdateGetDifference), reqBytes, reqDelegate, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

func init() {
	SDK.AddCmd(SdkConnInfo)
	SDK.AddCmd(SdkSetLogLevel)
	SDK.AddCmd(SdkGetDiffrence)
}
