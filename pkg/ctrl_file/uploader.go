package fileCtrl

import (
	"context"
	"fmt"
	"git.ronaksoftware.com/river/msg/msg"
	"git.ronaksoftware.com/ronak/riversdk/internal/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"github.com/gobwas/pool/pbytes"
	"go.uber.org/zap"
	"io"
	"math"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"time"
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
	httpContext context.Context
	cancelFunc  context.CancelFunc

	// IsProfilePhoto indicates that the uploaded file will be used as a profile photo for Group or User
	IsProfilePhoto bool `json:"is_profile_photo"`
	// PeerID will be set if IsProfilePhoto has been set to TRUE and user is going to upload group photo
	GroupID int64 `json:"group_id"`
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
	// SkipDelegateCall identifies to call delegate function on specified states
	SkipDelegateCall bool  `json:"skip_delegate_call"`
	PeerID           int64 `json:"peer_id"`

	// These parts are used to check if the file has been already uploaded
	CheckSha256 bool   `json:"check_sha_256"`
	FileSha256  string `json:"file_sha256"`
	AccessHash  uint64 `json:"access_hash"`
	ClusterID   int32  `json:"cluster_id"`
	DocumentID  int64  `json:"document_id"`
}

func (r UploadRequest) GetID() string {
	return fmt.Sprintf("0.%d.0", r.FileID)
}

type uploadContext struct {
	mtx          sync.Mutex
	rateLimit    chan struct{}
	parts        chan int32
	file         *os.File
	req          UploadRequest
	lastProgress int64
}

func (ctx *uploadContext) prepare() {
	dividend := int32(ctx.req.FileSize / int64(ctx.req.ChunkSize))
	if ctx.req.FileSize%int64(ctx.req.ChunkSize) > 0 {
		ctx.req.TotalParts = dividend + 1
	} else {
		ctx.req.TotalParts = dividend
	}

	ctx.parts = make(chan int32, ctx.req.TotalParts+ctx.req.MaxInFlights)
	ctx.rateLimit = make(chan struct{}, ctx.req.MaxInFlights)
	for partIndex := int32(0); partIndex < ctx.req.TotalParts-1; partIndex++ {
		if ctx.isUploaded(partIndex) {
			continue
		}
		ctx.parts <- partIndex
	}
}

func (ctx *uploadContext) resetUploadedList(ctrl *Controller) {
	ctx.mtx.Lock()
	ctx.req.cancelFunc()
	ctx.req.httpContext, ctx.req.cancelFunc = context.WithCancel(context.Background())
	ctx.req.UploadedParts = ctx.req.UploadedParts[:0]
	ctx.mtx.Unlock()
	ctrl.saveUploads(ctx.req)
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

func (ctx *uploadContext) addToUploaded(ctrl *Controller, partIndex int32) {
	if ctx.isUploaded(partIndex) {
		return
	}
	ctx.mtx.Lock()
	ctx.req.UploadedParts = append(ctx.req.UploadedParts, partIndex)
	progress := int64(float64(len(ctx.req.UploadedParts)) / float64(ctx.req.TotalParts) * 100)
	skipOnProgress := false
	if ctx.lastProgress > progress {
		skipOnProgress = true
	} else {
		ctx.lastProgress = progress
	}
	ctx.mtx.Unlock()
	ctrl.saveUploads(ctx.req)
	if !ctx.req.SkipDelegateCall && !skipOnProgress {
		ctrl.onProgressChanged(ctx.req.GetID(), 0, ctx.req.FileID, 0, progress, ctx.req.PeerID)
	}
}

func (ctx *uploadContext) generateFileSavePart(fileID int64, partID int32, totalParts int32, bytes []byte) *msg.MessageEnvelope {
	envelop := msg.MessageEnvelope{
		RequestID:   uint64(domain.SequentialUniqueID()),
		Constructor: msg.C_FileSavePart,
	}
	req := msg.FileSavePart{
		TotalParts: totalParts,
		Bytes:      bytes,
		FileID:     fileID,
		PartID:     partID,
	}
	envelop.Message, _ = req.Marshal()

	logs.Debug("FileCtrl generates FileSavePart",
		zap.Int64("MsgID", ctx.req.MessageID),
		zap.Int64("FileID", req.FileID),
		zap.Int32("PartID", req.PartID),
		zap.Int32("TotalParts", req.TotalParts),
		zap.Int("Bytes", len(req.Bytes)),
	)
	return &envelop
}

func (ctx *uploadContext) execute(ctrl *Controller) domain.RequestStatus {
	for {
		ctx.prepare()
		logs.Info("FileCtrl executes Upload",
			zap.Int64("FileID", ctx.req.FileID),
			zap.Int32("TotalParts", ctx.req.TotalParts),
			zap.Int32("ChunkSize", ctx.req.ChunkSize),
		)

		maxRetries := int32(math.Min(float64(ctx.req.MaxInFlights), float64(ctx.req.TotalParts)))
		waitGroup := sync.WaitGroup{}
		for maxRetries > 0 {
			select {
			case partIndex := <-ctx.parts:
				if !ctrl.existUploadRequest(ctx.req.GetID()) {
					waitGroup.Wait()
					_ = ctx.file.Close()
					logs.Warn("Upload Canceled (Request Not Exists)",
						zap.Int64("FileID", ctx.req.FileID),
						zap.Int64("Size", ctx.req.FileSize),
						zap.String("Path", ctx.req.FilePath),
					)
					if !ctx.req.SkipDelegateCall {
						ctrl.onCancel(ctx.req.GetID(), 0, ctx.req.FileID, 0, false, ctx.req.PeerID)
					}
					return domain.RequestStatusCanceled
				}
				waitGroup.Add(1)
				ctx.rateLimit <- struct{}{}
				go ctx.uploadJob(ctrl, &maxRetries, &waitGroup, partIndex)
			default:
				switch int32(len(ctx.req.UploadedParts)) {
				case ctx.req.TotalParts - 1:
					logs.Debug("FileCtrl waits for all (n-1) parts uploads to complete")
					waitGroup.Wait()
					ctx.parts <- ctx.req.TotalParts - 1
				case ctx.req.TotalParts:
					logs.Debug("FileCtrl waits for last part to upload")
					waitGroup.Wait()
					if !ctrl.postUploadProcess(ctx.req) {
						ctrl.onCancel(ctx.req.GetID(), 0, ctx.req.FileID, 0, true, ctx.req.PeerID)
						return domain.RequestStatusError
					}
					// We have finished our uploads
					_ = ctx.file.Close()
					if !ctx.req.SkipDelegateCall {
						ctrl.onCompleted(ctx.req.GetID(), 0, ctx.req.FileID, 0, ctx.req.FilePath, ctx.req.PeerID)
					}
					_ = repo.Files.MarkAsUploaded(ctx.req.FileID)
					return domain.RequestStatusCompleted
				default:
					// Keep Uploading
					time.Sleep(time.Millisecond * 250)
				}
			}
		}

		minChunkSize := minChunkSize(ctx.req.FileSize)
		if ctx.req.ChunkSize > minChunkSize {
			logs.Info("FileCtrl retries upload with smaller chunk size",
				zap.Int32("Old", ctx.req.ChunkSize>>10),
				zap.Int32("New", minChunkSize>>10),
			)
			ctx.req.ChunkSize = minChunkSize
			ctx.resetUploadedList(ctrl)
			continue
		}
		_ = ctx.file.Close()
		logs.Warn("Upload Canceled (Max Retries Exceeds)",
			zap.Int64("FileID", ctx.req.FileID),
			zap.Int64("Size", ctx.req.FileSize),
			zap.String("Path", ctx.req.FilePath),
		)
		if !ctx.req.SkipDelegateCall {
			ctrl.onCancel(ctx.req.GetID(), 0, ctx.req.FileID, 0, true, ctx.req.PeerID)
		}
		return domain.RequestStatusError
	}

}
func (ctx *uploadContext) uploadJob(ctrl *Controller, maxRetries *int32, waitGroup *sync.WaitGroup, partIndex int32) {
	defer waitGroup.Done()
	defer func() {
		<-ctx.rateLimit
	}()

	bytes := pbytes.GetLen(int(ctx.req.ChunkSize))
	defer pbytes.Put(bytes)
	offset := partIndex * ctx.req.ChunkSize
	n, err := ctx.file.ReadAt(bytes, int64(offset))
	if err != nil && err != io.EOF {
		logs.Warn("Error in ReadFile", zap.Error(err))
		atomic.StoreInt32(maxRetries, 0)
		ctx.parts <- partIndex
		return
	}
	if n == 0 {
		return
	}
	res, err := ctrl.network.SendHttp(
		ctx.req.httpContext,
		ctx.generateFileSavePart(ctx.req.FileID, partIndex+1, ctx.req.TotalParts, bytes[:n]),
		ctrl.httpRequestTimeout,
	)
	if err != nil {
		logs.Warn("Error On Http Response", zap.Error(err))
		switch e := err.(type) {
		case *url.Error:
			if e.Timeout() {
				atomic.AddInt32(maxRetries, -1)
			}
		default:
		}
		time.Sleep(100 * time.Millisecond)
		ctx.parts <- partIndex
		return
	}
	switch res.Constructor {
	case msg.C_Bool:
		ctx.addToUploaded(ctrl, partIndex)
	case msg.C_Error:
		x := &msg.Error{}
		_ = x.Unmarshal(res.Message)
		logs.Debug("FileCtrl received Error response",
			zap.Int32("PartID", partIndex+1),
			zap.String("Code", x.Code),
			zap.String("Item", x.Items),
		)
		ctx.parts <- partIndex
	default:
		logs.Debug("FileCtrl received unexpected response", zap.String("C", msg.ConstructorNames[res.Constructor]))
		atomic.StoreInt32(maxRetries, 0)
		return
	}
}
