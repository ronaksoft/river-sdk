package main

import (
	"mime"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/filemanager"

	"go.uber.org/zap"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/log"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/repo"
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

var DownloadMultiConnection = &ishell.Cmd{
	Name: "DownloadMultiConnection",
	Func: func(c *ishell.Context) {
		messageID := fnGetMessageID(c)
		segmentCount := 8
		x := new(msg.MediaDocument)
		m := repo.Ctx().Messages.GetMessage(messageID)
		if m.MediaType == msg.MediaTypeDocument {

			err := x.Unmarshal(m.Media)
			if err != nil {
				log.LOG_Error("Error", zap.Error(err))
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
			log.LOG_Error("Error", zap.Error(err))
		}
		defer f.Close()
		for partIdx, buff := range fileBuff {
			position := partIdx * domain.FilePayloadSize
			_, err := f.WriteAt(buff, int64(position))
			if err != nil {
				log.LOG_Error("Error", zap.Error(err))
			}
		}
		log.LOG_Info("File save Completed :", zap.String("fileName", f.Name()))
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
			_Log.Debug(err.Error())
		} else {
			reqDelegate.RequestID = reqID
		}

	},
}

func init() {
	File.AddCmd(Upload)
	File.AddCmd(Download)
	File.AddCmd(DownloadMultiConnection)
	File.AddCmd(ShareContact)

}

func downloadWorker(workerIdx int, wg *sync.WaitGroup, partQueue chan int, fileBuff map[int][]byte, fileLock *sync.Mutex, x *msg.MediaDocument) {
	defer wg.Done()

	log.LOG_Info("Worker Started :", zap.Int("worker", workerIdx))
	for {
		select {
		case partIdx := <-partQueue:
			ctx := filemanager.Ctx()
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
			for _SDK.GetNetworkStatus() == int32(domain.DISCONNECTED) || _SDK.GetNetworkStatus() == int32(domain.CONNECTING) {
				log.LOG_Warn("network is not connected", zap.Int("worker", workerIdx), zap.Int("PartIdx", partIdx))
				time.Sleep(500 * time.Millisecond)
			}

			log.LOG_Debug("send download request", zap.Int("worker", workerIdx), zap.Int("PartIdx", partIdx))
			res, err := ctx.Send(envelop)

			if err == nil {
				responseID := res.RequestID
				if requestID != responseID {
					log.LOG_Warn("RequestIDs are not equal", zap.Uint64("reqID", requestID), zap.Uint64("resID", responseID))
				} else {
					log.LOG_Debug("RequestIDs are equal", zap.Uint64("reqID", requestID), zap.Uint64("resID", responseID))
				}

				switch res.Constructor {
				case msg.C_Error:
					x := new(msg.Error)
					x.Unmarshal(res.Message)
					log.LOG_Error("received Error response", zap.Int("worker", workerIdx), zap.Int("PartIdx", partIdx), zap.String("Code", x.Code), zap.String("Item", x.Items))
					// on error add to queue again
					partQueue <- partIdx
				case msg.C_File:
					x := new(msg.File)
					err := x.Unmarshal(res.Message)
					if err != nil {
						// on error add to queue again
						partQueue <- partIdx
						log.LOG_Error("failed to unmarshal C_File", zap.Int("worker", workerIdx), zap.Int("PartIdx", partIdx), zap.Error(err))
					} else {
						fileLock.Lock()
						fileBuff[partIdx] = x.Bytes
						fileLock.Unlock()
					}
				default:
					// on error add to queue again
					partQueue <- partIdx
					log.LOG_Error("received unknown response", zap.Int("worker", workerIdx), zap.Int("PartIdx", partIdx), zap.Error(err))
				}
			} else {
				// on error add to queue again
				partQueue <- partIdx
				log.LOG_Error("downloadWorker()", zap.Int("worker", workerIdx), zap.Int("PartIdx", partIdx), zap.Error(err))
			}

		default:
			log.LOG_Info("Worker Exited :", zap.Int("worker", workerIdx))
			return
		}
	}

}
