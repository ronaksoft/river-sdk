package main

import (
	"fmt"
	"git.ronaksoft.com/river/msg/msg"
	"git.ronaksoft.com/ronak/riversdk"
	mon "git.ronaksoft.com/ronak/riversdk/internal/monitoring"
	"gopkg.in/abiosoft/ishell.v2"
	"time"
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
		if reqID, err := _SDK.ExecuteCommand(msg.C_UpdateGetDifference, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
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
		if reqID, err := _SDK.ExecuteCommand(msg.C_SystemGetServerTime, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
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
		if reqID, err := _SDK.ExecuteCommand(msg.C_UpdateGetState, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var SdkAppForeground = &ishell.Cmd{
	Name: "AppForeground",
	Func: func(c *ishell.Context) {
		_SDK.AppForeground()
	},
}

var SdkAppBackground = &ishell.Cmd{
	Name: "AppBackground",
	Func: func(c *ishell.Context) {
		_SDK.AppBackground()
	},
}

var SdkPrintMonitor = &ishell.Cmd{
	Name: "Monitor",
	Func: func(c *ishell.Context) {
		c.Println("ForegroundTime:", mon.Stats.ForegroundTime)
		c.Println((time.Duration(mon.Stats.ForegroundTime) * time.Second).String())
	},
}

var SdkResetUsage = &ishell.Cmd{
	Name: "ResetUsage",
	Func: func(c *ishell.Context) {
		mon.ResetUsage()
	},
}

var SdkSetTeam = &ishell.Cmd{
	Name: "SetTeam",
	Func: func(c *ishell.Context) {
		teamID := fnGetTeamID(c)
		accessHash := fnGetAccessHash(c)
		_SDK.SetTeam(teamID, int64(accessHash), false)
	},
}

var SdkGetTeam = &ishell.Cmd{
	Name: "GetTeam",
	Func: func(c *ishell.Context) {
		c.Println(riversdk.GetCurrTeam())
	},
}

func init() {
	SDK.AddCmd(SdkConnInfo)
	SDK.AddCmd(SdkSetLogLevel)
	SDK.AddCmd(SdkGetDifference)
	SDK.AddCmd(SdkGetServerTime)
	SDK.AddCmd(SdkUpdateGetState)
	SDK.AddCmd(GetAuthKey)
	SDK.AddCmd(SdkAppForeground)
	SDK.AddCmd(SdkAppBackground)
	SDK.AddCmd(SdkPrintMonitor)
	SDK.AddCmd(SdkResetUsage)
	SDK.AddCmd(SdkSetTeam)
	SDK.AddCmd(SdkGetTeam)
}
