package fileCtrl

import (
	"context"
	"git.ronaksoft.com/river/msg/msg"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/pkg/ctrl_file/executor"
	"git.ronaksoft.com/river/sdk/pkg/domain"
	"git.ronaksoft.com/river/sdk/pkg/repo"
	"github.com/gobwas/pool/pbytes"
	"go.uber.org/zap"
	"io"
	"os"
	"sync"
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
	msg.ClientFileRequest
	ctrl         *Controller
	mtx          sync.Mutex
	file         *os.File
	parts        chan int32
	done         chan struct{}
	lastProgress int64
}

func (u *UploadRequest) checkSha256() error {
	envelop := &msg.MessageEnvelope{
		RequestID:   uint64(domain.SequentialUniqueID()),
		Constructor: msg.C_FileGetBySha256,
	}
	req := &msg.FileGetBySha256{
		Sha256:   u.FileSha256,
		FileSize: int32(u.FileSize),
	}
	envelop.Message, _ = req.Marshal()

	err := domain.Try(3, time.Millisecond*500, func() error {
		ctx, cancelFunc := context.WithTimeout(context.Background(), domain.HttpRequestTimeout)
		defer cancelFunc()
		res, err := u.ctrl.network.SendHttp(ctx, envelop)
		if err != nil {
			return err
		}
		switch res.Constructor {
		case msg.C_FileLocation:
			x := &msg.FileLocation{}
			_ = x.Unmarshal(res.Message)
			u.ClusterID = x.ClusterID
			u.AccessHash = x.AccessHash
			u.FileID = x.FileID
			u.TotalParts = -1 // dirty hack, which queue.Start() knows the upload request is completed
			return nil
		case msg.C_Error:
			x := &msg.Error{}
			_ = x.Unmarshal(res.Message)
		}
		return domain.ErrServer
	})
	return err
}

func (u *UploadRequest) generateFileSavePart(fileID int64, partID int32, totalParts int32, bytes []byte) *msg.MessageEnvelope {
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
		zap.Int64("MsgID", u.MessageID),
		zap.Int64("FileID", req.FileID),
		zap.Int32("PartID", req.PartID),
		zap.Int32("TotalParts", req.TotalParts),
		zap.Int("Bytes", len(req.Bytes)),
	)
	return &envelop
}

func (u *UploadRequest) isUploaded(partIndex int32) bool {
	u.mtx.Lock()
	defer u.mtx.Unlock()
	for _, index := range u.FinishedParts {
		if partIndex == index {
			return true
		}
	}
	return false
}

func (u *UploadRequest) addToUploaded(partIndex int32) {
	if u.isUploaded(partIndex) {
		return
	}
	u.mtx.Lock()
	u.FinishedParts = append(u.FinishedParts, partIndex)
	progress := int64(float64(len(u.FinishedParts)) / float64(u.TotalParts) * 100)
	skipOnProgress := false
	if u.lastProgress > progress {
		skipOnProgress = true
	} else {
		u.lastProgress = progress
	}
	u.mtx.Unlock()

	_ = repo.Files.SaveFileRequest(u.GetID(), &u.ClientFileRequest)

	if !u.SkipDelegateCall && !skipOnProgress {
		u.ctrl.onProgressChanged(u.GetID(), 0, u.FileID, 0, progress, u.PeerID)
	}
}

func (u *UploadRequest) GetID() string {
	return getRequestID(u.ClusterID, u.FileID, u.AccessHash)
}

func (u *UploadRequest) Prepare() error {
	// Check File stats and return error if any problem exists
	fileInfo, err := os.Stat(u.FilePath)
	if err != nil {
		u.ctrl.onCancel(u.GetID(), 0, u.FileID, 0, true, u.PeerID)
		return err
	} else {
		u.FileSize = fileInfo.Size()
		if u.FileSize <= 0 {
			u.ctrl.onCancel(u.GetID(), 0, u.FileID, 0, true, u.PeerID)
			return err
		} else if u.FileSize > maxFileSizeAllowedSize {
			u.ctrl.onCancel(u.GetID(), 0, u.FileID, 0, true, u.PeerID)
			return err
		}
	}

	// If Sha256 exists in the request then we check server if this file has been already uploaded, if true, then
	// we do not upload it again and we call postUploadProcess with the updated details
	if u.CheckSha256 && len(u.FileSha256) != 0 {
		err = u.checkSha256()
		if err == nil {
			logs.Info("File already exists in the server",
				zap.Int64("FileID", u.FileID),
				zap.Int32("ClusterID", u.ClusterID),
			)
			if !u.ctrl.postUploadProcess(u.ClientFileRequest) {
				u.ctrl.onCancel(u.GetID(), 0, u.FileID, 0, true, u.PeerID)
			}
			return domain.ErrAlreadyUploaded
		}
	}

	// Open the file for read
	u.file, err = os.OpenFile(u.FilePath, os.O_RDONLY, 0666)
	if err != nil {
		u.ctrl.onCancel(u.GetID(), 0, u.FileID, 0, true, u.PeerID)
		return err
	}

	// If chunk size is not set recalculate it
	if u.ChunkSize <= 0 {
		u.ChunkSize = bestChunkSize(u.FileSize)
	}

	dividend := int32(u.FileSize / int64(u.ChunkSize))
	if u.FileSize%int64(u.ChunkSize) > 0 {
		u.TotalParts = dividend + 1
	} else {
		u.TotalParts = dividend
	}

	u.parts = make(chan int32, u.TotalParts)
	u.done = make(chan struct{})
	maxPartIndex := u.TotalParts - 1
	if u.TotalParts == 1 {
		maxPartIndex = u.TotalParts
	}
	for partIndex := int32(0); partIndex < maxPartIndex; partIndex++ {
		if u.isUploaded(partIndex) {
			continue
		}
		u.parts <- partIndex
	}

	return nil
}

func (u *UploadRequest) NextAction() executor.Action {
	select {
	case partID := <-u.parts:
		return &UploadAction{
			id:  partID,
			req: u,
		}
	case <-u.done:
		return nil
	}
}

func (u *UploadRequest) ActionDone(id int32) {
	switch u.TotalParts {
	case 1:
		if int32(len(u.FinishedParts)) != u.TotalParts {
			panic("BUG!! total parts != finished parts")
		}
	default:
		finishedParts := int32(len(u.FinishedParts))
		switch {
		case finishedParts < u.TotalParts-1:
			return
		case finishedParts == u.TotalParts-1:
			u.parts <- u.TotalParts - 1
			return
		}
	}
	// This is last part
	u.done <- struct{}{}
	if !u.ctrl.postUploadProcess(u.ClientFileRequest) {
		u.ctrl.onCancel(u.GetID(), 0, u.FileID, 0, true, u.PeerID)
		return
	}
	// We have finished our uploads
	_ = u.file.Close()
	if !u.SkipDelegateCall {
		u.ctrl.onCompleted(u.GetID(), 0, u.FileID, 0, u.FilePath, u.PeerID)
	}
	_ = repo.Files.DeleteFileRequest(u.GetID())
	return
}

func (u *UploadRequest) Serialize() []byte {
	b, err := u.Marshal()
	if err != nil {
		panic(err)
	}
	return b
}

type UploadAction struct {
	id  int32
	req *UploadRequest
}

func (a *UploadAction) ID() int32 {
	return a.id
}

func (a *UploadAction) Do(ctx context.Context) {
	bytes := pbytes.GetLen(int(a.req.ChunkSize))
	defer pbytes.Put(bytes)
	offset := a.id * a.req.ChunkSize
	n, err := a.req.file.ReadAt(bytes, int64(offset))
	if err != nil && err != io.EOF {
		logs.Warn("Error in ReadFile", zap.Error(err))
		a.req.parts <- a.id
		return
	}
	if n == 0 {
		return
	}
	res, err := a.req.ctrl.network.SendHttp(
		ctx,
		a.req.generateFileSavePart(a.req.FileID, a.id+1, a.req.TotalParts, bytes[:n]),
	)
	if err != nil {
		logs.Warn("Error On Http Response", zap.Error(err))
		time.Sleep(100 * time.Millisecond)
		a.req.parts <- a.id
		return
	}
	switch res.Constructor {
	case msg.C_Bool:
		a.req.addToUploaded(a.id)
	case msg.C_Error:
		x := &msg.Error{}
		_ = x.Unmarshal(res.Message)
		logs.Debug("FileCtrl received Error response",
			zap.Int32("PartID", a.id+1),
			zap.String("Code", x.Code),
			zap.String("Item", x.Items),
		)
		a.req.parts <- a.id
	default:
		logs.Debug("FileCtrl received unexpected response", zap.String("C", msg.ConstructorNames[res.Constructor]))
		return
	}
}

//
// func (ctx *uploadContext) resetUploadedList(ctrl *Controller) {
// 	ctx.mtx.Lock()
// 	ctx.req.cancelFunc()
// 	ctx.req.httpContext, ctx.req.cancelFunc = context.WithCancel(context.Background())
// 	ctx.req.UploadedParts = ctx.req.UploadedParts[:0]
// 	ctx.mtx.Unlock()
// 	ctrl.saveUploads(ctx.req)
// }
//
//
// 		minChunkSize := minChunkSize(ctx.req.FileSize)
// 		if ctx.req.ChunkSize > minChunkSize {
// 			logs.Info("FileCtrl retries upload with smaller chunk size",
// 				zap.Int32("Old", ctx.req.ChunkSize>>10),
// 				zap.Int32("New", minChunkSize>>10),
// 			)
// 			ctx.req.ChunkSize = minChunkSize
// 			ctx.resetUploadedList(ctrl)
// 			continue
// 		}
// 		_ = ctx.file.Close()
// 		logs.Warn("Upload Canceled (Max Retries Exceeds)",
// 			zap.Int64("FileID", ctx.req.FileID),
// 			zap.Int64("Size", ctx.req.FileSize),
// 			zap.String("Path", ctx.req.FilePath),
// 		)
// 		if !ctx.req.SkipDelegateCall {
// 			ctrl.onCancel(ctx.req.GetID(), 0, ctx.req.FileID, 0, true, ctx.req.PeerID)
// 		}
// 		return domain.RequestStatusError
// 	}
//
// }
//
