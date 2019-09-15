package fileCtrl

import (
	"fmt"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"github.com/gobwas/pool/pbytes"
	"go.uber.org/zap"
	"io"
	"os"
	"sync"
	"sync/atomic"
)

/*
   Creation Time: 2019 - Sep - 07
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

type UploadRequest struct {
	// IsProfilePhoto indicates that the uploaded file will be used as a profile photo for Group or User
	IsProfilePhoto bool `json:"is_profile_photo"`
	// PeerID will be set if IsProfilePhoto has been set to TRUE and user is going to upload group photo
	GroupID int64 `json:"group_id"`
	// MaxRetries defines how many time each request could encounter error before giving up
	MaxRetries int32 `json:"max_retries"`
	// MessageID (Optional) if is set then (ClusterID, FileID, AccessHash, Version) will be read from the message
	// document object, or if message has no document then return error
	MessageID int64 `json:"message_id"`
	// FilePath defines the path which downloaded file will be stored. It must be a file not a directory.
	// Also it will be overwritten if Overwrite is TRUE
	FilePath string `json:"file_path"`
	// FileID is a random id which must be unique for the client, it will be used to reference the uploaded document
	// in the server
	FileID int64 `json:"file_id"`
	// FileSize (Optional) if is set then progress will be calculated
	FileSize int64 `json:"file_size"`
	// ThumbPath (Optional) defines the path of thumbnail of the file will be stored. It must be a file not a directory.
	// Also it will be overwritten if Overwrite is TRUE
	ThumbPath string `json:"thumb_path"`
	// ThumbID is a random id which must be unique for the client, it will be used to reference the uploaded document
	// thumbnail in the server
	ThumbID int64
	// ChunkSize identifies how many request we need to send to server to Download a file.
	ChunkSize int32 `json:"chunk_size"`
	// MaxInFlights defines that how many requests could be send concurrently
	MaxInFlights  int32   `json:"max_in_flights"`
	UploadedParts []int32 `json:"downloaded_parts"`
	TotalParts    int32   `json:"total_parts"`
	Canceled      bool    `json:"canceled"`
}

func (r UploadRequest) GetID() string {
	return fmt.Sprintf("0.%d.0", r.FileID)
}

type uploadContext struct {
	mtx       sync.Mutex
	rateLimit chan struct{}
	parts     chan int32
	file      *os.File
	ctrl      *Controller
	req       UploadRequest
}

func (ctx *uploadContext) isUploaded(partIndex int32) bool {
	ctx.mtx.Lock()
	defer ctx.mtx.Unlock()
	for _, index := range ctx.req.UploadedParts {
		if partIndex == index {
			return true
		}
	}
	return false
}

func (ctx *uploadContext) addToUploaded(partIndex int32) {
	ctx.mtx.Lock()
	ctx.req.UploadedParts = append(ctx.req.UploadedParts, partIndex)
	progress := int64(float64(len(ctx.req.UploadedParts)) / float64(ctx.req.TotalParts) * 100)
	ctx.mtx.Unlock()
	ctx.ctrl.saveUploads(ctx.req)
	ctx.ctrl.onProgressChanged(ctx.req.GetID(), 0, ctx.req.FileID, 0, progress)
}

func (ctx *uploadContext) generateFileSavePart(fileID int64, partID int32, totalParts int32, bytes []byte) *msg.MessageEnvelope {
	req := new(msg.FileSavePart)
	req.TotalParts = totalParts
	req.Bytes = bytes
	req.FileID = fileID
	req.PartID = partID

	envelop := new(msg.MessageEnvelope)
	envelop.Constructor = msg.C_FileSavePart
	envelop.Message, _ = req.Marshal()
	envelop.RequestID = uint64(domain.SequentialUniqueID())
	logs.Debug("FilesStatus::generateFileSavePart()",
		zap.Int64("MsgID", ctx.req.MessageID),
		zap.Int64("FileID", req.FileID),
		zap.Int32("PartID", req.PartID),
		zap.Int32("TotalParts", req.TotalParts),
		zap.Int("Bytes", len(req.Bytes)),
	)
	return envelop
}

func (ctx *uploadContext) execute() domain.RequestStatus {
	waitGroup := sync.WaitGroup{}
	for ctx.req.MaxRetries > 0 {
		select {
		case partIndex := <-ctx.parts:
			if !ctx.ctrl.existUploadRequest(ctx.req.GetID()) {
				waitGroup.Wait()
				_ = ctx.file.Close()
				ctx.ctrl.onCancel(ctx.req.GetID(), 0, ctx.req.FileID, 0, false)
				return domain.RequestStatusCanceled
			}
			ctx.rateLimit <- struct{}{}
			waitGroup.Add(1)

			go func(partIndex int32) {
				defer waitGroup.Done()
				defer func() {
					<-ctx.rateLimit
				}()

				bytes := pbytes.GetLen(int(ctx.req.ChunkSize))
				defer pbytes.Put(bytes)
				offset := partIndex * ctx.req.ChunkSize
				_, err := ctx.file.ReadAt(bytes, int64(offset))
				if err != nil && err != io.EOF {
					logs.Warn("Error in ReadFile", zap.Error(err))
					atomic.AddInt32(&ctx.req.MaxRetries, -1)
					ctx.parts <- partIndex
					return
				}
				res, err := ctx.ctrl.network.SendHttp(ctx.generateFileSavePart(ctx.req.FileID, partIndex+1, ctx.req.TotalParts, bytes))
				if err != nil {
					logs.Warn("Error in SendHttp", zap.Error(err))
					atomic.AddInt32(&ctx.req.MaxRetries, -1)
					ctx.parts <- partIndex
					return
				}
				switch res.Constructor {
				case msg.C_Bool:
					ctx.addToUploaded(partIndex)
				default:
					atomic.AddInt32(&ctx.req.MaxRetries, -1)
					ctx.parts <- partIndex
				}
			}(partIndex)
		default:
			waitGroup.Wait()
			switch int32(len(ctx.req.UploadedParts)) {
			case ctx.req.TotalParts - 1:
				// If we finished uploading n-1 parts then run the last loop with the last part
				ctx.parts <- ctx.req.TotalParts - 1
			case ctx.req.TotalParts:
				// We have finished our uploads
				_ = ctx.file.Close()
				ctx.ctrl.onCompleted(ctx.req.GetID(), 0, ctx.req.FileID, 0, ctx.req.FilePath)
				if ctx.ctrl.postUploadProcess != nil {
					ctx.ctrl.postUploadProcess(ctx.req)
				}
				return domain.RequestStatusCompleted
			}
		}
	}

	_ = ctx.file.Close()
	ctx.ctrl.onCancel(ctx.req.GetID(), 0, ctx.req.FileID, 0, true)
	return domain.RequestStatusError
}
