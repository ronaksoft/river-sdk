package main

import (
    "fmt"
    "time"

    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/river-sdk/internal/domain"
    "github.com/ronaksoft/river-sdk/internal/logs"
    mon "github.com/ronaksoft/river-sdk/internal/monitoring"
    riversdk "github.com/ronaksoft/river-sdk/sdk/prime"
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
        c.Println("Username", _SDK.ConnInfo.Username)
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
        _SDK.AppForeground(true)
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
        c.Println(domain.GetCurrTeamID(), domain.GetCurrTeamAccess())
    },
}

var SdkDeletePending = &ishell.Cmd{
    Name: "DeletePending",
    Func: func(c *ishell.Context) {
        _SDK.DeletePendingMessage(fnGetMessageID(c))

    },
}

var SdkCancelFileRequest = &ishell.Cmd{
    Name: "CancelFileRequest",
    Func: func(c *ishell.Context) {
        _SDK.CancelFileRequest(fnGetString(c, "FileRequestID"))
    },
}

var SdkDeleteAllPendingMessages = &ishell.Cmd{
    Name: "DeleteAllPendingMessages",
    Func: func(c *ishell.Context) {
        _SDK.DeleteAllPendingMessages()
    },
}

var SdkRemoteLog = &ishell.Cmd{
    Name: "LiveLog",
    Func: func(c *ishell.Context) {
        logs.SetRemoteLog("https://livelog.ronaksoftware.com/cli")
    },
}

func init() {
    SDK.AddCmd(SdkDeleteAllPendingMessages)
    SDK.AddCmd(SdkCancelFileRequest)
    SDK.AddCmd(SdkDeletePending)
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
    SDK.AddCmd(SdkRemoteLog)
}
