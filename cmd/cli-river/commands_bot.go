package main

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"gopkg.in/abiosoft/ishell.v2"
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
		req.Peer.Type = msg.PeerUser
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

func init() {
	Bot.AddCmd(BotGetInlineQueryResults)
	Bot.AddCmd(BotSendInlineQueryResults)
}
