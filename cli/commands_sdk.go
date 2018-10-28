package main

import (
	"strconv"

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
		c.Println("UserID:", _SDK.ConnInfo.UserID)
		c.Println("AuthID:", _SDK.ConnInfo.AuthID)
		c.Println("Phone:", _SDK.ConnInfo.Phone)
		c.Println("FirstName:", _SDK.ConnInfo.FirstName)
		c.Println("LastName:", _SDK.ConnInfo.LastName)
		c.Println("AuthKey:", _SDK.ConnInfo.AuthKey)
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
