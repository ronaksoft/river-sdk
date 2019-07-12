package main

import (
	fileCtrl "git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_file"
	"mime"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
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
		req.ThumbFilePath = fnGetThumbFilePath(c)
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
			_Log.Error("ExecuteCommand failed", zap.Error(err))
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

var Status = &ishell.Cmd{
	Name: "Status",
	Func: func(c *ishell.Context) {
		messageID := fnGetMessageID(c)
		str := _SDK.GetFileStatus(messageID)
		c.Println(str)
	},
}

var DownloadMultiConnection = &ishell.Cmd{
	Name: "DownloadMultiConnection",
	Func: func(c *ishell.Context) {
		messageID := fnGetMessageID(c)
		segmentCount := 8
		x := new(msg.MediaDocument)
		m := repo.Messages.GetMessage(messageID)
		if m.MediaType == msg.MediaTypeDocument {

			err := x.Unmarshal(m.Media)
			if err != nil {
				_Log.Error("Error", zap.Error(err))
				return
			}

		}
		totalParts := 0
		count := int(x.Doc.FileSize / domain.FilePayloadSize)
		if (count * domain.FilePayloadSize) < int(x.Doc.FileSize) {
			totalParts = count + 1
		} else {
			totalParts = count
		}

		partsQueue := make(chan int, totalParts)
		// add all parts to queue
		for i := 0; i < totalParts; i++ {
			partsQueue <- i
		}
		fileBuff := make(map[int][]byte)
		fileLock := &sync.Mutex{}
		wg := &sync.WaitGroup{}
		wg.Add(segmentCount)
		for i := 0; i < segmentCount; i++ {
			go downloadWorker(i, wg, partsQueue, fileBuff, fileLock, x)
		}
		wg.Wait()

		// save file

		strName := strconv.FormatInt(domain.SequentialUniqueID(), 10) + ".tmp"
		f, err := os.Create(strName)
		if err != nil {
			_Log.Error("Error", zap.Error(err))
		}
		defer f.Close()
		for partIdx, buff := range fileBuff {
			position := partIdx * domain.FilePayloadSize
			_, err := f.WriteAt(buff, int64(position))
			if err != nil {
				_Log.Error("Error", zap.Error(err))
			}
		}
		_Log.Info("File save Completed :", zap.String("fileName", f.Name()))
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
		if reqID, err := _SDK.ExecuteCommand(msg.C_MessagesSendMedia, reqBytes, reqDelegate, false, false); err != nil {
			_Log.Error("ExecuteCommand failed", zap.Error(err))
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

var DownloadThumbnail = &ishell.Cmd{
	Name: "DownloadThumbnail",
	Func: func(c *ishell.Context) {
		messageID := fnGetMessageID(c)
		strFilePath := _SDK.FileDownloadThumbnail(messageID)
		_Log.Info("File Download Complete", zap.String("path", strFilePath))
	},
}

var GetSharedMedia = &ishell.Cmd{
	Name: "GetSharedMedia",
	Func: func(c *ishell.Context) {
		peerType := fnGetPeerType(c)
		peerID := fnGetPeerID(c)
		mediaType := fnGetMediaType(c)
		reqDelegate := &RequestDelegate{}
		_SDK.GetSharedMedia(peerID, int32(peerType), int32(mediaType), reqDelegate)
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

		_SDK.ExecuteCommand(msg.C_ClientSendMessageMedia, reqBuff1, req1, false, false)
		_SDK.ExecuteCommand(msg.C_ClientSendMessageMedia, reqBuff2, req2, false, false)
		_SDK.ExecuteCommand(msg.C_ClientSendMessageMedia, reqBuff3, req3, false, false)
		_SDK.ExecuteCommand(msg.C_ClientSendMessageMedia, reqBuff4, req4, false, false)
		_SDK.ExecuteCommand(msg.C_ClientSendMessageMedia, reqBuff5, req5, false, false)

	},
}

func init() {
	File.AddCmd(Upload)
	File.AddCmd(Download)
	File.AddCmd(DownloadMultiConnection)
	File.AddCmd(ShareContact)
	File.AddCmd(Status)
	File.AddCmd(DownloadThumbnail)
	File.AddCmd(GetSharedMedia)

	File.AddCmd(TestUpload)

}

func downloadWorker(workerIdx int, wg *sync.WaitGroup, partQueue chan int, fileBuff map[int][]byte, fileLock *sync.Mutex, x *msg.MediaDocument) {
	defer wg.Done()

	_Log.Info("Worker Started :", zap.Int("worker", workerIdx))
	for {
		select {
		case partIdx := <-partQueue:
			ctx := fileCtrl.New(fileCtrl.Config{})
			req := new(msg.FileGet)
			req.Location = &msg.InputFileLocation{
				AccessHash: x.Doc.AccessHash,
				ClusterID:  x.Doc.ClusterID,
				FileID:     x.Doc.ID,
				Version:    x.Doc.Version,
			}
			req.Offset = int32(partIdx * domain.FilePayloadSize)
			req.Limit = domain.FilePayloadSize
			if req.Offset+domain.FilePayloadSize > x.Doc.FileSize {
				req.Limit = x.Doc.FileSize - (req.Offset)
			}

			requestID := uint64(domain.SequentialUniqueID())

			reqBuff, _ := req.Marshal()

			envelop := new(msg.MessageEnvelope)
			envelop.Constructor = msg.C_FileGet
			envelop.Message = reqBuff
			envelop.RequestID = requestID

			// Send
			for _SDK.GetNetworkStatus() == int32(domain.NetworkDisconnected) || _SDK.GetNetworkStatus() == int32(domain.NetworkConnecting) {
				_Log.Warn("network is not connected", zap.Int("worker", workerIdx), zap.Int("PartIdx", partIdx))
				time.Sleep(500 * time.Millisecond)
			}

			_Log.Debug("send download request", zap.Int("worker", workerIdx), zap.Int("PartIdx", partIdx))
			res, err := ctx.Send(envelop)

			if err == nil {
				responseID := res.RequestID
				if requestID != responseID {
					_Log.Warn("RequestIDs are not equal", zap.Uint64("reqID", requestID), zap.Uint64("resID", responseID))
				} else {
					_Log.Debug("RequestIDs are equal", zap.Uint64("reqID", requestID), zap.Uint64("resID", responseID))
				}

				switch res.Constructor {
				case msg.C_Error:
					x := new(msg.Error)
					x.Unmarshal(res.Message)
					_Log.Error("received Error response", zap.Int("worker", workerIdx), zap.Int("PartIdx", partIdx), zap.String("Code", x.Code), zap.String("Item", x.Items))
					// on error add to queue again
					partQueue <- partIdx
				case msg.C_File:
					x := new(msg.File)
					err := x.Unmarshal(res.Message)
					if err != nil {
						// on error add to queue again
						partQueue <- partIdx
						_Log.Error("failed to unmarshal C_File", zap.Int("worker", workerIdx), zap.Int("PartIdx", partIdx), zap.Error(err))
					} else {
						fileLock.Lock()
						fileBuff[partIdx] = x.Bytes
						fileLock.Unlock()
					}
				default:
					// on error add to queue again
					partQueue <- partIdx
					_Log.Error("received unknown response", zap.Int("worker", workerIdx), zap.Int("PartIdx", partIdx), zap.Error(err))
				}
			} else {
				// on error add to queue again
				partQueue <- partIdx
				_Log.Error("downloadWorker()", zap.Int("worker", workerIdx), zap.Int("PartIdx", partIdx), zap.Error(err))
			}

		default:
			_Log.Info("Worker Exited :", zap.Int("worker", workerIdx))
			return
		}
	}

}
