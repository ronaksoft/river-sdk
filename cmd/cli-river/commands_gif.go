package main

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"gopkg.in/abiosoft/ishell.v2"
)

var Gif = &ishell.Cmd{
	Name: "Gif",
}

var GifSave = &ishell.Cmd{
	Name: "Save",
	Func: func(c *ishell.Context) {
		req := msg.GifSave{
			Doc: &msg.InputDocument{},
		}
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
		req := msg.GifGetSaved{}
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_GifGetSaved, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var GifDelete = &ishell.Cmd{
	Name: "Delete",
	Func: func(c *ishell.Context) {
		req := msg.GifDelete{
			Doc: &msg.InputDocument{},
		}
		req.Doc.ID = fnGetFileID(c)
		req.Doc.AccessHash = fnGetAccessHash(c)
		req.Doc.ClusterID = 1
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_GifDelete, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

func init() {
	Gif.AddCmd(GifSave)
	Gif.AddCmd(GifGetSaved)
	Gif.AddCmd(GifDelete)
}
