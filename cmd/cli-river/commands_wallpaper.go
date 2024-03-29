package main

import (
    "github.com/abiosoft/ishell/v2"
    "github.com/ronaksoft/river-msg/go/msg"
)

var WallPaper = &ishell.Cmd{
    Name: "WallPaper",
}

var WallPaperGet = &ishell.Cmd{
    Name: "Get",
    Func: func(c *ishell.Context) {
        req := msg.WallPaperGet{}
        reqBytes, _ := req.Marshal()
        reqDelegate := new(RequestDelegate)
        if reqID, err := _SDK.ExecuteCommand(msg.C_WallPaperGet, reqBytes, reqDelegate); err != nil {
            c.Println("Command Failed:", err)
        } else {
            reqDelegate.RequestID = reqID
        }
    },
}

func init() {
    WallPaper.AddCmd(WallPaperGet)
}
