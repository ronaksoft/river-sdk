package main

import (
	"git.ronaksoftware.com/ronak/riversdk/logs"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"go.uber.org/zap"
	"gopkg.in/abiosoft/ishell.v2"
)

var SDK = &ishell.Cmd{
	Name: "SDK",
}

var SdkConnInfo = &ishell.Cmd{
	Name: "ConnInfo",
	Func: func(c *ishell.Context) {
		c.Println("UserID:", _SDK.ConnInfo.UserID)
		c.Println("AuthID:", _SDK.ConnInfo.PickupAuthID())
		c.Println("Phone:", _SDK.ConnInfo.Phone)
		c.Println("FirstName:", _SDK.ConnInfo.FirstName)
		c.Println("LastName:", _SDK.ConnInfo.LastName)
		c.Println("AuthKey:", _SDK.ConnInfo.PickupAuthKey())
	},
}

var SdkSetLogLevel = &ishell.Cmd{
	Name: "SetLogLevel",
	Func: func(c *ishell.Context) {
		choiceIndex := c.MultiChoice([]string{
			"Debug", "Info", "Warn", "Error",
		}, "Level")
		logs.SetLogLevel(choiceIndex - 1)
	},
}

var SdkGetDiffrence = &ishell.Cmd{
	Name: "GetDiffrence",
	Func: func(c *ishell.Context) {
		req := msg.UpdateGetDifference{}
		req.Limit = fnGetLimit(c)
		req.From = fnGetFromUpdateID(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_UpdateGetDifference, reqBytes, reqDelegate, false, false); err != nil {
			logs.Error("ExecuteCommand failed", zap.Error(err))
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var SdkGetServerTime = &ishell.Cmd{
	Name: "GetServerTime",
	Func: func(c *ishell.Context) {
		req := msg.SystemGetServerTime{}
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_SystemGetServerTime, reqBytes, reqDelegate, false, false); err != nil {
			logs.Error("ExecuteCommand failed", zap.Error(err))
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

func init() {
	SDK.AddCmd(SdkConnInfo)
	SDK.AddCmd(SdkSetLogLevel)
	SDK.AddCmd(SdkGetDiffrence)
	SDK.AddCmd(SdkGetServerTime)
}
