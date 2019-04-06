package main

import (
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"go.uber.org/zap"
	"gopkg.in/abiosoft/ishell.v2"
)

var Init = &ishell.Cmd{
	Name:    "Init",
	Aliases: []string{"init", "i"},
}

var InitAuth = &ishell.Cmd{
	Name: "Auth",
	Func: func(c *ishell.Context) {
		if err := _SDK.CreateAuthKey(); err != nil {
			logs.Error("CreateAuthKey failed", zap.Error(err))
		} else {
			logs.Message("CreateAuthKey == OK ==")
		}
	},
}

func init() {
	Init.AddCmd(InitAuth)
}
