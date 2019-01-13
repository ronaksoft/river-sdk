package main

import (
	"mime"
	"os"
	"path"

	"git.ronaksoftware.com/ronak/riversdk/msg"
	ishell "gopkg.in/abiosoft/ishell.v2"
)

var File = &ishell.Cmd{
	Name: "File",
}

var Upload = &ishell.Cmd{
	Name: "Upload",
	Func: func(c *ishell.Context) {

		req := msg.ClientSendMessageMedia{}
		req.Attributes = make([]*msg.DocumentAttribute, 0)

		req.Peer = new(msg.InputPeer)
		req.Peer.Type = fnGetPeerType(c)
		req.Peer.ID = fnGetPeerID(c)
		req.Peer.AccessHash = fnGetAccessHash(c)
		req.FilePath = fnGetFilePath(c)
		req.ReplyTo = fnGetReplyTo(c)

		f, _ := os.Open(req.FilePath)
		filename := f.Name()
		f.Close()

		req.Caption = filename
		req.ClearDraft = true
		req.FileMIME = mime.TypeByExtension(path.Ext(filename))
		req.FileName = filename

		req.ThumbFilePath = ""
		req.ThumbMIME = ""

		req.MediaType = fnGetInputMediaType(c)
		req.Attributes = fnGetAttributes(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_ClientSendMessageMedia, reqBytes, reqDelegate, false, false); err != nil {
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var Download = &ishell.Cmd{
	Name: "Download",
	Func: func(c *ishell.Context) {
		messageID := fnGetMessageID(c)
		_SDK.FileDownload(messageID)
	},
}

func init() {
	File.AddCmd(Upload)
	File.AddCmd(Download)

}
