package main

import (
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"go.uber.org/zap"
	"gopkg.in/abiosoft/ishell.v2"
)

var SDK = &ishell.Cmd{
	Name: "SDK",
}

var SdkConnInfo = &ishell.Cmd{
	Name: "ConnInfo",
	Func: func(c *ishell.Context) {
		c.Println("userID:", _SDK.ConnInfo.UserID)
		c.Println("AuthID:", _SDK.ConnInfo.AuthID)
		c.Println("Phone:", _SDK.ConnInfo.Phone)
		c.Println("FirstName:", _SDK.ConnInfo.FirstName)
		c.Println("LastName:", _SDK.ConnInfo.LastName)
		c.Println("AuthKey:", _SDK.ConnInfo.AuthKey)
	},
}

var GetAuthKey = &ishell.Cmd{
	Name: "GetAuthKey",
	Func: func(c *ishell.Context) {
		authKey := _SDK.ConnInfo.GetAuthKey()
		fmt.Println("authKey", authKey)
	},
}

var SdkSetLogLevel = &ishell.Cmd{
	Name: "SetLogLevel",
	Func: func(c *ishell.Context) {
		choiceIndex := c.MultiChoice([]string{
			"Debug", "Info", "Warn", "Error",
		}, "Level")
		riversdk.SetLogLevel(choiceIndex - 1)
	},
}

var SdkGetDifference = &ishell.Cmd{
	Name: "GetDifference",
	Func: func(c *ishell.Context) {
		req := msg.UpdateGetDifference{}
		req.Limit = fnGetLimit(c)
		req.From = fnGetFromUpdateID(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_UpdateGetDifference, reqBytes, reqDelegate, false, false); err != nil {
			_Log.Error("ExecuteCommand failed", zap.Error(err))
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
			_Log.Error("ExecuteCommand failed", zap.Error(err))
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var SdkUpdateGetState = &ishell.Cmd{
	Name: "UpdateGetState",
	Func: func(c *ishell.Context) {
		req := msg.UpdateGetState{}
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_UpdateGetState, reqBytes, reqDelegate, false, false); err != nil {
			_Log.Error("ExecuteCommand failed", zap.Error(err))
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

func init() {
	SDK.AddCmd(SdkConnInfo)
	SDK.AddCmd(SdkSetLogLevel)
	SDK.AddCmd(SdkGetDifference)
	SDK.AddCmd(SdkGetServerTime)
	SDK.AddCmd(SdkUpdateGetState)
	SDK.AddCmd(GetAuthKey)
}
