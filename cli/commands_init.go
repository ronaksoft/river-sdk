package main

import (
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
            _Log.Error(err.Error())
        } else {
            _Log.Debug("== OK ==")
        }
    },
}

func init() {
    Init.AddCmd(InitAuth)
}
