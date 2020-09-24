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
	msg.ClientFileRequest
	ctrl          *Controller
	mtx           sync.Mutex
	file          *os.File
	parts         chan int32
	done          chan struct{}
	lastPartSent  bool
	progress      int64
	failedActions int32
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

func (u *UploadRequest) resetUploadedList() {
	u.mtx.Lock()
	u.FinishedParts = u.FinishedParts[:0]
	u.mtx.Unlock()

	_ = repo.Files.SaveFileRequest(u.GetID(), &u.ClientFileRequest, false)
	if !u.SkipDelegateCall {
		u.ctrl.onProgressChanged(u.GetID(), 0, u.FileID, 0, 0, u.PeerID)
	}
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
	if u.progress > progress {
		skipOnProgress = true
	} else {
		u.progress = progress
	}
	u.mtx.Unlock()

	_ = repo.Files.SaveFileRequest(u.GetID(), &u.ClientFileRequest, true)

	if !u.SkipDelegateCall && !skipOnProgress {
		u.ctrl.onProgressChanged(u.GetID(), 0, u.FileID, 0, progress, u.PeerID)
	}
}

func (u *UploadRequest) reset() {
	// Reset failed counter
	atomic.StoreInt32(&u.failedActions, 0)

	// Reset the uploaded list
	u.resetUploadedList()
	u.progress = 0

	if u.file != nil {
		_ = u.file.Close()
	}
}

func (u *UploadRequest) cancel(err error) {
	_ = repo.Files.DeleteFileRequest(u.GetID())
	if !u.SkipDelegateCall {
		u.ctrl.onCancel(u.GetID(), 0, u.FileID, 0, err != nil, u.PeerID)
	}
}

func (u *UploadRequest) complete() {
	_ = repo.Files.DeleteFileRequest(u.GetID())
	if !u.SkipDelegateCall {
		u.ctrl.onCompleted(u.GetID(), 0, u.FileID, 0, u.FilePath, u.PeerID)
	}

}

func (u *UploadRequest) GetID() string {
	return getRequestID(u.ClusterID, u.FileID, u.AccessHash)
}

func (u *UploadRequest) Prepare() error {
	logs.Info("FileCtrl prepare UploadRequest", zap.String("ReqID", u.GetID()))
	u.reset()

	// Check File stats and return error if any problem exists
	fileInfo, err := os.Stat(u.FilePath)
	if err != nil {
		u.cancel(err)
		return err
	} else {
		u.FileSize = fileInfo.Size()
		if u.FileSize <= 0 {
			err = domain.ErrInvalidData
			u.cancel(err)
			return err
		} else if u.FileSize > maxFileSizeAllowedSize {
			err = domain.ErrFileTooLarge
			u.cancel(err)
			return err
		}
	}

	// If Sha256 exists in the request then we check server if this file has been already uploaded, if true, then
	// we do not upload it again and we call postUploadProcess with the updated details
	if u.CheckSha256 && len(u.FileSha256) != 0 {
		oldReqID := u.GetID()
		err = u.checkSha256()
		if err == nil {
			logs.Info("FileCtrl detects the file already exists in the server",
				zap.Int64("FileID", u.FileID),
				zap.Int32("ClusterID", u.ClusterID),
			)
			_ = repo.Files.DeleteFileRequest(oldReqID)
			err = domain.ErrAlreadyUploaded
			if !u.ctrl.postUploadProcess(u.ClientFileRequest) {
				err = domain.ErrInvalidData
				u.cancel(err)
			}
			return err
		}
	}

	// Open the file for read
	u.file, err = os.OpenFile(u.FilePath, os.O_RDONLY, 0666)
	if err != nil {
		u.cancel(err)
		return err
	}

	// If chunk size is not set recalculate it
	if u.ChunkSize <= 0 {
		u.ChunkSize = bestChunkSize(u.FileSize)
	}

	// Calculate number of parts based on our chunk size
	dividend := int32(u.FileSize / int64(u.ChunkSize))
	if u.FileSize%int64(u.ChunkSize) > 0 {
		u.TotalParts = dividend + 1
	} else {
		u.TotalParts = dividend
	}

	// Reset FinishedParts if all parts are finished. Probably something went wrong, it is better to retry
	if int32(len(u.FinishedParts)) == u.TotalParts {
		u.FinishedParts = u.FinishedParts[:0]
	}

	// Prepare Channels to active the system dynamics
	u.parts = make(chan int32, u.TotalParts)
	u.done = make(chan struct{}, 1)
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

	logs.Debug("Upload Prepared",
		zap.String("ID", u.GetID()),
		zap.Int32("TotalParts", u.TotalParts),
		zap.Int32s("Finished", u.FinishedParts),
	)
	return nil
}

func (u *UploadRequest) NextAction() executor.Action {
	// If request is canceled then return nil
	if _, err := repo.Files.GetFileRequest(u.GetID()); err != nil {
		logs.Warn("FileCtrl did not find UploadRequest, we cancel it", zap.Error(err))
		return nil
	}

	// Wait for next part, or return nil if we finished
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
	logs.Info("FileCtrl finished upload part",
		zap.String("ReqID", u.GetID()),
		zap.Int32("PartID", id),
		zap.Int("FinishedParts", len(u.FinishedParts)),
		zap.Int32("TotalParts", u.TotalParts),
	)
	// If we have failed too many times, and we can decrease the chunk size the we do it again.
	if atomic.LoadInt32(&u.failedActions) > retryMaxAttempts {
		atomic.StoreInt32(&u.failedActions, 0)
		logs.Debug("Max Attempts",
			zap.Int32("ChunkSize", u.ChunkSize),
		)
	}

	// For single part uploads we are done
	// For n-part uploads if we have done n-1 part then we add the last part
	switch u.TotalParts {
	case 1:
		if int32(len(u.FinishedParts)) != u.TotalParts {
			logs.Fatal("FileCtrl got serious error total parts != finished parts")
		}
	default:
		finishedParts := int32(len(u.FinishedParts))
		switch {
		case finishedParts < u.TotalParts-1:
			return
		case finishedParts == u.TotalParts-1:
			if u.lastPartSent {
				return
			}
			u.lastPartSent = true
			u.parts <- u.TotalParts - 1
			return
		}
	}

	// This is last part so we make the executor free to run the next job if exist
	u.done <- struct{}{}
	_ = u.file.Close()

	// Run the post process
	if !u.ctrl.postUploadProcess(u.ClientFileRequest) {
		u.cancel(domain.ErrNoPostProcess)
		return
	}

	// Clean up
	u.complete()

	return
}

func (u *UploadRequest) Serialize() []byte {
	b, err := u.Marshal()
	if err != nil {
		panic(err)
	}
	return b
}

func (u *UploadRequest) Next() executor.Request {
	// If the request has a chained request then we swap them and reset the progress
	if u.ClientFileRequest.Next == nil {
		return nil
	}

	// swap the child content
	u.ClientFileRequest = *u.ClientFileRequest.Next

	return u
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

	// Calculate offset based on chunk id and the chunk size
	offset := a.id * a.req.ChunkSize

	// We try to read the chunk, if it failed we try one more time
	n, err := a.req.file.ReadAt(bytes, int64(offset))
	if err != nil && err != io.EOF {
		logs.Warn("FileCtrl got error in ReadFile (Upload)", zap.Error(err))
		a.req.parts <- a.id
		return
	}

	// If we read 0 bytes then something is wrong
	if n == 0 {
		logs.Fatal("FileCtrl read zero bytes from file",
			zap.String("FilePath", a.req.FilePath),
			zap.Int32("TotalParts", a.req.TotalParts),
			zap.Int32("ChunkSize", a.req.ChunkSize),
		)
	}

	// Send the http request to server
	res, err := a.req.ctrl.network.SendHttp(
		ctx,
		a.req.generateFileSavePart(a.req.FileID, a.id+1, a.req.TotalParts, bytes[:n]),
	)
	if err != nil {
		logs.Warn("FileCtrl got error On SendHttp (Upload)", zap.Error(err))
		atomic.AddInt32(&a.req.failedActions, 1)
		a.req.parts <- a.id
		return
	}
	switch res.Constructor {
	case msg.C_Bool:
		a.req.addToUploaded(a.id)
	case msg.C_Error:
		x := &msg.Error{}
		_ = x.Unmarshal(res.Message)
		logs.Warn("FileCtrl received Error response (Upload)",
			zap.Int32("PartID", a.id+1),
			zap.String("Code", x.Code),
			zap.String("Item", x.Items),
		)
		atomic.AddInt32(&a.req.failedActions, 1)
		a.req.parts <- a.id
	default:
		logs.Fatal("FileCtrl received unexpected response (Upload)", zap.String("C", msg.ConstructorNames[res.Constructor]))
		return
	}
}
