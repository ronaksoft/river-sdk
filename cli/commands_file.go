package main

import (
	"git.ronaksoftware.com/ronak/riversdk/msg"
	ishell "gopkg.in/abiosoft/ishell.v2"
)

var File = &ishell.Cmd{
	Name: "File",
}
var FileSavePart = &ishell.Cmd{
	Name: "FileSavePart",
	Func: func(c *ishell.Context) {
		// for just one user
		req := msg.FileSavePart{}

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_FileSavePart, reqBytes, reqDelegate, false, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}
var MessagesSendMedia = &ishell.Cmd{
	Name: "MessagesSendMedia",
	Func: func(c *ishell.Context) {
		// for just one user
		req := msg.MessagesSendMedia{}

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesSendMedia, reqBytes, reqDelegate, false, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var FileGet = &ishell.Cmd{
	Name: "FileGet",
	Func: func(c *ishell.Context) {
		// for just one user
		req := msg.FileGet{}

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_FileGet, reqBytes, reqDelegate, false, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

func init() {
	Debug.AddCmd(FileGet)
	File.AddCmd(FileSavePart)
	Debug.AddCmd(MessagesSendMedia)

}
