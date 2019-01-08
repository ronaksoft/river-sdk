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

		// xxx := new(msg.ClientSendMessageMedia)

		// // for just one user
		// req := msg.MessagesSendMedia{}

		// req.ClearDraft = true
		// req.MediaType = msg.MediaTypeDocument

		// doc := new(msg.InputMediaUploadedDocument)

		// photo := new(msg.InputMediaUploadedPhoto)

		// doc.MimeType = ""
		// doc.Caption = "Test File"

		// // set attributes
		// doc.Attributes = make([]*msg.DocumentAttribute)

		// docAttrib := new(msg.DocumentAttribute)
		// docAttrib.Type = msg.DocumentAttributeFile

		// attribFile := new(msg.DocumentAttributeFile)

		// xx = new(msg.DocumentAttributeAudio)

		// attribFile.Filename = fnGetFileName(c)
		// // marshal attribs
		// docAttrib.Data, _ = attribFile.Marshal()

		// // file
		// file := new(msg.InputFile)
		// file.
		// 	doc.File = file

		// // marshal doc data
		// req.MediaType = msg.InputMediaTypeUploadedDocument
		// req.MediaData, _ = doc.Marshal()

		// req.Peer = msg.InputPeer{}
		// req.Peer.Type = fnGetPeerType(c)
		// req.Peer.ID = fnGetPeerID(c)
		// req.Peer.AccessHash = fnGetAccessHash(c)

		// req.RandomID = domain.SequentialUniqueID()

		// reqBytes, _ := req.Marshal()
		// reqDelegate := new(RequestDelegate)
		// if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesSendMedia, reqBytes, reqDelegate, false, false); err != nil {
		// 	_Log.Debug(err.Error())
		// } else {
		// 	reqDelegate.RequestID = reqID
		// }

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
