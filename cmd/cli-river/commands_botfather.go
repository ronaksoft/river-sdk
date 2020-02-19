package main

import (
	msg "git.ronaksoftware.com/ronak/riversdk/msg/chat"
	"git.ronaksoftware.com/ronak/toolbox"
	"go.uber.org/zap"
	"gopkg.in/abiosoft/ishell.v2"
)

var Botfather = &ishell.Cmd{
	Name: "Botfather",
}

var BotStart = &ishell.Cmd{
	Name: "BotStart",
	Func: func(c *ishell.Context) {
		req := msg.BotStart{}
		req.Bot = &msg.InputPeer{}
		req.Bot.Type = fnGetPeerType(c)
		req.Bot.ID = fnGetBotID(c)
		req.RandomID = ronak.RandomInt64(0)
		req.StartParam = "startparam"
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_BotStart, reqBytes, reqDelegate); err != nil {
			_Log.Error("EnqueueCommand failed", zap.Error(err))
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var BotSetInfo = &ishell.Cmd{
	Name: "BotSetInfo",
	Func: func(c *ishell.Context) {
		req := msg.BotSetInfo{}
		req.BotID = fnGetBotID(c)
		req.Owner = _SDK.ConnInfo.UserID
		req.RandomID = ronak.RandomInt64(0)
		req.BotCommands = FnGetCommands(c)
		req.Description = FnGetDescription(c)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_BotSetInfo, reqBytes, reqDelegate); err != nil {
			_Log.Error("EnqueueCommand failed", zap.Error(err))
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var BotsGetInfo = &ishell.Cmd{
	Name: "BotsGetInfo",
	Func: func(c *ishell.Context) {
		req := msg.BotGet{}
		req.Limit = fnGetLimit(c)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_BotGet, reqBytes, reqDelegate); err != nil {
			_Log.Error("EnqueueCommand failed", zap.Error(err))
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

func init() {
	Botfather.AddCmd(BotStart)
	Botfather.AddCmd(BotSetInfo)
	Botfather.AddCmd(BotsGetInfo)
}
