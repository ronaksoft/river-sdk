package main

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"github.com/ronaksoft/rony"
	"gopkg.in/abiosoft/ishell.v2"
)

/*
   Creation Time: 2021 - May - 01
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func init() {
	Mini.AddCmd(MiniMessageGetDialogs)
	Mini.AddCmd(MiniAccountGetTeams)
	Mini.AddCmd(MiniMessageSendMediaToSelf)
}

var Mini = &ishell.Cmd{
	Name: "Mini",
}

var MiniMessageGetDialogs = &ishell.Cmd{
	Name: "GetDialogs",
	Func: func(c *ishell.Context) {
		req := msg.MessagesGetDialogs{}
		req.Limit = int32(100)
		req.Offset = fnGetOffset(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _MiniSDK.ExecuteCommand(msg.C_MessagesGetDialogs, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var MiniAccountGetTeams = &ishell.Cmd{
	Name: "GetTeams",
	Func: func(c *ishell.Context) {
		req := msg.AccountGetTeams{}
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)

		if reqID, err := _MiniSDK.ExecuteCommand(msg.C_AccountGetTeams, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}
	},
}

var MiniMessageSendMediaToSelf = &ishell.Cmd{
	Name: "SendMediaToMe",
	Func: func(c *ishell.Context) {
		attrFile := &msg.DocumentAttributeFile{
			Filename: "File.jpg",
		}
		attrFileBytes, _ := attrFile.Marshal()
		attrPhoto := &msg.DocumentAttributePhoto{
			Width:  720,
			Height: 720,
		}
		attrPhotoBytes, _ := attrPhoto.Marshal()
		req := msg.ClientSendMessageMedia{
			Peer: &msg.InputPeer{
				ID:         _SDK.ConnInfo.UserID,
				Type:       msg.PeerType_PeerUser,
				AccessHash: 0,
			},
			MediaType:  msg.InputMediaType_InputMediaTypeUploadedDocument,
			Caption:    "Some Random Caption",
			FileName:   "Uploaded File",
			FileMIME:   "image/jpeg",
			ThumbMIME:  "",
			ReplyTo:    0,
			ClearDraft: false,
			Attributes: []*msg.DocumentAttribute{
				{
					Type: msg.DocumentAttributeType_AttributeTypePhoto,
					Data: attrPhotoBytes,
				},
				{
					Type: msg.DocumentAttributeType_AttributeTypeFile,
					Data: attrFileBytes,
				},
			},
		}
		// _ = exec.Command("cp", "./_testdata/T.jpg", "./_testdata/F.jpg").Run()
		// req.FilePath = "./_testdata/F.jpg"
		// req.ThumbFilePath = "./_testdata/T.jpg"
		req.FilePath = fnGetFilePath(c)
		req.ThumbFilePath = fnGetThumbFilePath(c)
		req.Entities = nil
		reqBytes, _ := req.Marshal()
		reqDelegate := NewCustomDelegate()
		reqDelegate.OnCompleteFunc = func(b []byte) {
			x := &rony.MessageEnvelope{}
			_ = x.Unmarshal(b)
			MessagePrinter(x)
		}
		if reqID, err := _MiniSDK.ExecuteCommand(msg.C_ClientSendMessageMedia, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}
