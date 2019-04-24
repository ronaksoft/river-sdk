package main

import (
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
			_Log.Error("CreateAuthKey failed", zap.Error(err))
		} else {
			_Log.Info("CreateAuthKey == OK ==")
		}
	},
}

func init() {
	Init.AddCmd(InitAuth)
}
