package main

import (
	"git.ronaksoftware.com/river/msg/msg"
	"gopkg.in/abiosoft/ishell.v2"
)

var Gif = &ishell.Cmd{
	Name: "Gif",
}

var GifSave = &ishell.Cmd{
	Name: "Save",
	Func: func(c *ishell.Context) {
		req := msg.GifSave{}
		req.Doc.ClusterID = 1
		req.Doc.ID = fnGetFileID(c)
		req.Doc.AccessHash = fnGetAccessHash(c)
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_GifSave, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var GifGetSaved = &ishell.Cmd{
	Name: "GetSaved",
	Func: func(c *ishell.Context) {
		req := msg.ClientGetSavedGifs{}
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_ClientGetSavedGifs, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}


func init() {
	Gif.AddCmd(GifSave)
	Gif.AddCmd(GifGetSaved)
}
