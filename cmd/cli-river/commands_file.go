package main

import (
	"git.ronaksoft.com/river/msg/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"gopkg.in/abiosoft/ishell.v2"
	"mime"
	"os"
	"path"
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
		req.Peer.Type = msg.PeerUser
		req.Peer.ID = _SDK.ConnInfo.UserID
		// req.Peer.Type = fnGetPeerType(c)
		// req.Peer.ID = fnGetPeerID(c)
		// req.Peer.AccessHash = fnGetAccessHash(c)
		req.FilePath = "./_testdata/TEST1.png"
		req.ThumbFilePath = "./_testdata/FileThumb.png"
		// req.FilePath = fnGetFilePath(c)
		// req.ThumbFilePath = fnGetThumbFilePath(c)
		// req.ReplyTo = fnGetReplyTo(c)

		f, _ := os.Open(req.FilePath)
		filename := f.Name()
		f.Close()

		req.Caption = filename
		req.ClearDraft = true
		req.FileMIME = mime.TypeByExtension(path.Ext(filename))
		req.FileName = filename

		req.MediaType = msg.InputMediaTypeUploadedDocument

		// req.MediaType = fnGetInputMediaType(c)
		// req.Attributes = fnGetAttributes(c)

		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_ClientSendMessageMedia, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var ShareContact = &ishell.Cmd{
	Name: "ShareContact",
	Func: func(c *ishell.Context) {

		req := new(msg.MessagesSendMedia)

		req.Peer = &msg.InputPeer{}
		req.Peer.Type = fnGetPeerType(c)
		req.Peer.ID = fnGetPeerID(c)
		req.Peer.AccessHash = fnGetAccessHash(c)
		req.ReplyTo = fnGetReplyTo(c)

		// get Media Contact
		req.MediaType = msg.InputMediaTypeContact
		contact := new(msg.InputMediaContact)
		contact.FirstName = fnGetFirstName(c)
		contact.LastName = fnGetLastName(c)
		contact.Phone = fnGetPhone(c)
		// marshal contact
		req.MediaData, _ = contact.Marshal()

		req.ClearDraft = true
		req.RandomID = domain.SequentialUniqueID()

		// send request to server
		reqBytes, _ := req.Marshal()
		reqDelegate := new(RequestDelegate)
		if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesSendMedia, reqBytes, reqDelegate); err != nil {
			c.Println("Command Failed:", err)
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var GetSharedMedia = &ishell.Cmd{
	Name: "GetSharedMedia",
	Func: func(c *ishell.Context) {
		peerType := fnGetPeerType(c)
		peerID := fnGetPeerID(c)
		mediaType := fnGetMediaType(c)
		reqDelegate := &RequestDelegate{}
		_SDK.GetSharedMedia(domain.GetCurrTeamID(), peerID, int32(peerType), int32(mediaType), reqDelegate)
	},
}

var TestUpload = &ishell.Cmd{
	Name: "TestUpload",
	Func: func(c *ishell.Context) {

		peer := &msg.InputPeer{
			AccessHash: 0,
			ID:         -65984812425083,
			Type:       msg.PeerGroup,
		}
		f1 := new(msg.ClientSendMessageMedia)
		f2 := new(msg.ClientSendMessageMedia)
		f3 := new(msg.ClientSendMessageMedia)
		f4 := new(msg.ClientSendMessageMedia)
		f5 := new(msg.ClientSendMessageMedia)

		f1.Peer = peer
		f2.Peer = peer
		f3.Peer = peer
		f4.Peer = peer
		f5.Peer = peer

		f1.MediaType = msg.InputMediaTypeUploadedDocument
		f2.MediaType = msg.InputMediaTypeUploadedDocument
		f3.MediaType = msg.InputMediaTypeUploadedDocument
		f4.MediaType = msg.InputMediaTypeUploadedDocument
		f5.MediaType = msg.InputMediaTypeUploadedDocument

		f1.Caption = "AAAA-1"
		f2.Caption = "AAAA-2"
		f3.Caption = "AAAA-3"
		f4.Caption = "AAAA-4"
		f5.Caption = "AAAA-5"

		f1.FileName = "AAAA-1.jpg"
		f2.FileName = "AAAA-2.jpg"
		f3.FileName = "AAAA-3.jpg"
		f4.FileName = "AAAA-4.jpg"
		f5.FileName = "AAAA-5.jpg"

		f1.FilePath = "/tmpfs/1.jpg"
		f2.FilePath = "/tmpfs/2.jpg"
		f3.FilePath = "/tmpfs/3.jpg"
		f4.FilePath = "/tmpfs/4.jpg"
		f5.FilePath = "/tmpfs/5.jpg"

		f1.ThumbFilePath = "/tmpfs/t1.jpg"
		f2.ThumbFilePath = "/tmpfs/t2.jpg"
		f3.ThumbFilePath = "/tmpfs/t3.jpg"
		f4.ThumbFilePath = "/tmpfs/t4.jpg"
		f5.ThumbFilePath = "/tmpfs/t5.jpg"

		f1.FileMIME = "image/jpeg"
		f2.FileMIME = "image/jpeg"
		f3.FileMIME = "image/jpeg"
		f4.FileMIME = "image/jpeg"
		f5.FileMIME = "image/jpeg"

		f1.ThumbMIME = "image/jpeg"
		f2.ThumbMIME = "image/jpeg"
		f3.ThumbMIME = "image/jpeg"
		f4.ThumbMIME = "image/jpeg"
		f5.ThumbMIME = "image/jpeg"

		f1.ReplyTo = 0
		f2.ReplyTo = 0
		f3.ReplyTo = 0
		f4.ReplyTo = 0
		f5.ReplyTo = 0

		f1.ClearDraft = true
		f2.ClearDraft = true
		f3.ClearDraft = true
		f4.ClearDraft = true
		f5.ClearDraft = true

		attr1 := msg.DocumentAttributeFile{Filename: "AAAA-1.jpg"}
		attr2 := msg.DocumentAttributeFile{Filename: "AAAA-2.jpg"}
		attr3 := msg.DocumentAttributeFile{Filename: "AAAA-3.jpg"}
		attr4 := msg.DocumentAttributeFile{Filename: "AAAA-4.jpg"}
		attr5 := msg.DocumentAttributeFile{Filename: "AAAA-5.jpg"}

		attrBuff1, _ := attr1.Marshal()
		attrBuff2, _ := attr2.Marshal()
		attrBuff3, _ := attr3.Marshal()
		attrBuff4, _ := attr4.Marshal()
		attrBuff5, _ := attr5.Marshal()

		f1.Attributes = []*msg.DocumentAttribute{&msg.DocumentAttribute{Type: msg.AttributeTypeFile, Data: attrBuff1}}
		f2.Attributes = []*msg.DocumentAttribute{&msg.DocumentAttribute{Type: msg.AttributeTypeFile, Data: attrBuff2}}
		f3.Attributes = []*msg.DocumentAttribute{&msg.DocumentAttribute{Type: msg.AttributeTypeFile, Data: attrBuff3}}
		f4.Attributes = []*msg.DocumentAttribute{&msg.DocumentAttribute{Type: msg.AttributeTypeFile, Data: attrBuff4}}
		f5.Attributes = []*msg.DocumentAttribute{&msg.DocumentAttribute{Type: msg.AttributeTypeFile, Data: attrBuff5}}

		reqBuff1, _ := f1.Marshal()
		reqBuff2, _ := f2.Marshal()
		reqBuff3, _ := f3.Marshal()
		reqBuff4, _ := f4.Marshal()
		reqBuff5, _ := f5.Marshal()

		req1 := &RequestDelegate{}
		req2 := &RequestDelegate{}
		req3 := &RequestDelegate{}
		req4 := &RequestDelegate{}
		req5 := &RequestDelegate{}

		_SDK.ExecuteCommand(msg.C_ClientSendMessageMedia, reqBuff1, req1)
		_SDK.ExecuteCommand(msg.C_ClientSendMessageMedia, reqBuff2, req2)
		_SDK.ExecuteCommand(msg.C_ClientSendMessageMedia, reqBuff3, req3)
		_SDK.ExecuteCommand(msg.C_ClientSendMessageMedia, reqBuff4, req4)
		_SDK.ExecuteCommand(msg.C_ClientSendMessageMedia, reqBuff5, req5)

	},
}

var Download = &ishell.Cmd{
	Name: "Download",
	Func: func(c *ishell.Context) {
		clusterID := int32(1)
		docID := int64(687843935677241931)
		accessHash := uint64(4502154781611237)
		// clusterID := fnGetClusterID(c)
		// docID := fnGetFileID(c)
		// accessHash := fnGetAccessHash(c)
		err := _SDK.FileDownloadSync(clusterID, docID, int64(accessHash), true)
		if err != nil {
			c.Println(err)
			return
		}
	},
}

func init() {
	File.AddCmd(Upload)
	File.AddCmd(Download)
	File.AddCmd(ShareContact)
	File.AddCmd(GetSharedMedia)
	File.AddCmd(TestUpload)
}
