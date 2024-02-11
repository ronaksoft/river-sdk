package main

import (
    "github.com/abiosoft/ishell/v2"
    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/river-sdk/internal/domain"
)

/*
   Creation Time: 2020 - May - 15
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

var Bot = &ishell.Cmd{
    Name: "Bot",
}

var BotGetInlineQueryResults = &ishell.Cmd{
    Name: "GetInlineQueryResults",
    Func: func(c *ishell.Context) {
        req := msg.BotGetInlineResults{}
        c.Println("Enter Bot:")
        req.Bot = fnGetBot(c)
        c.Println("Enter Peer:")
        req.Peer = fnGetPeer(c)
        req.Peer.Type = msg.PeerType_PeerUser
        req.Query = fnGetQuery(c)
        req.Offset = ""
        reqBytes, _ := req.Marshal()
        reqDelegate := new(RequestDelegate)
        if reqID, err := _SDK.ExecuteCommand(msg.C_BotGetInlineResults, reqBytes, reqDelegate); err != nil {
            c.Println("Command Failed:", err)
        } else {
            reqDelegate.RequestID = reqID
        }

    },
}

var BotSendInlineQueryResults = &ishell.Cmd{
    Name: "SendInlineQueryResults",
    Func: func(c *ishell.Context) {
        req := msg.BotSendInlineResults{}
        c.Println("Enter Bot:")
        req.QueryID = fnGetQueryID(c)
        req.ResultID = fnGetResultID(c)
        req.Peer = fnGetPeer(c)
        req.RandomID = domain.RandomInt64(0)
        reqBytes, _ := req.Marshal()
        reqDelegate := new(RequestDelegate)
        if reqID, err := _SDK.ExecuteCommand(msg.C_BotSendInlineResults, reqBytes, reqDelegate); err != nil {
            c.Println("Command Failed:", err)
        } else {
            reqDelegate.RequestID = reqID
        }

    },
}

var BotGetCommands = &ishell.Cmd{
    Name: "BotGetCommands",
    Func: func(c *ishell.Context) {
        req := msg.BotGetCommands{}
        req.Peer = &msg.InputPeer{}
        req.Peer.Type = fnGetPeerType(c)
        req.Peer.ID = fnGetBotID(c)
        reqBytes, _ := req.Marshal()
        reqDelegate := new(RequestDelegate)
        if reqID, err := _SDK.ExecuteCommand(msg.C_BotGetCommands, reqBytes, reqDelegate); err != nil {
            c.Println("Command Failed:", err)
        } else {
            reqDelegate.RequestID = reqID
        }
    },
}

var BotStart = &ishell.Cmd{
    Name: "BotStart",
    Func: func(c *ishell.Context) {
        req := msg.BotStart{}
        req.Bot = &msg.InputPeer{}
        req.Bot.Type = fnGetPeerType(c)
        req.Bot.ID = fnGetBotID(c)
        req.RandomID = domain.RandomInt63()
        req.StartParam = "startparam"
        reqBytes, _ := req.Marshal()
        reqDelegate := new(RequestDelegate)
        if reqID, err := _SDK.ExecuteCommand(msg.C_BotStart, reqBytes, reqDelegate); err != nil {
            c.Println("Command Failed:", err)
        } else {
            reqDelegate.RequestID = reqID
        }

    },
}

func init() {
    Bot.AddCmd(BotGetInlineQueryResults)
    Bot.AddCmd(BotSendInlineQueryResults)
    Bot.AddCmd(BotStart)
    Bot.AddCmd(BotGetCommands)
}
