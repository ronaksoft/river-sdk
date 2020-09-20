package fileCtrl

import (
	"git.ronaksoft.com/river/msg/msg"
	"git.ronaksoft.com/river/sdk/pkg/ctrl_file/executor"
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
}

func (d *UploadRequest) GetID() string {
	return getRequestID(d.ClusterID, d.FileID, d.AccessHash)
}

func (u *UploadRequest) Prepare() {}

func (u *UploadRequest) NextAction() executor.Action {
	panic("implement me")
}

func (u *UploadRequest) ActionDone(id int32) {
	panic("implement me")
}

func (u *UploadRequest) Serialize() []byte {
	panic("implement me")
}

func (u *UploadRequest) Deserialize(bytes []byte) {
	panic("implement me")
}

// type uploadContext struct {
// 	mtx          sync.Mutex
// 	rateLimit    chan struct{}
// 	parts        chan int32
// 	file         *os.File
// 	req          UploadRequest
// 	lastProgress int64
// }
//
// func (ctx *uploadContext) prepare() {
// 	dividend := int32(ctx.req.FileSize / int64(ctx.req.ChunkSize))
// 	if ctx.req.FileSize%int64(ctx.req.ChunkSize) > 0 {
// 		ctx.req.TotalParts = dividend + 1
// 	} else {
// 		ctx.req.TotalParts = dividend
// 	}
//
// 	ctx.parts = make(chan int32, ctx.req.TotalParts+ctx.req.MaxInFlights)
// 	ctx.rateLimit = make(chan struct{}, ctx.req.MaxInFlights)
// 	for partIndex := int32(0); partIndex < ctx.req.TotalParts-1; partIndex++ {
// 		if ctx.isUploaded(partIndex) {
// 			continue
// 		}
// 		ctx.parts <- partIndex
// 	}
// }
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
// func (ctx *uploadContext) isUploaded(partIndex int32) bool {
// 	ctx.mtx.Lock()
// 	defer ctx.mtx.Unlock()
// 	for _, index := range ctx.req.UploadedParts {
// 		if partIndex == index {
// 			return true
// 		}
// 	}
// 	return false
// }
//
// func (ctx *uploadContext) addToUploaded(ctrl *Controller, partIndex int32) {
// 	if ctx.isUploaded(partIndex) {
// 		return
// 	}
// 	ctx.mtx.Lock()
// 	ctx.req.UploadedParts = append(ctx.req.UploadedParts, partIndex)
// 	progress := int64(float64(len(ctx.req.UploadedParts)) / float64(ctx.req.TotalParts) * 100)
// 	skipOnProgress := false
// 	if ctx.lastProgress > progress {
// 		skipOnProgress = true
// 	} else {
// 		ctx.lastProgress = progress
// 	}
// 	ctx.mtx.Unlock()
// 	ctrl.saveUploads(ctx.req)
// 	if !ctx.req.SkipDelegateCall && !skipOnProgress {
// 		ctrl.onProgressChanged(ctx.req.GetID(), 0, ctx.req.FileID, 0, progress, ctx.req.PeerID)
// 	}
// }
//
// func (ctx *uploadContext) generateFileSavePart(fileID int64, partID int32, totalParts int32, bytes []byte) *msg.MessageEnvelope {
// 	envelop := msg.MessageEnvelope{
// 		RequestID:   uint64(domain.SequentialUniqueID()),
// 		Constructor: msg.C_FileSavePart,
// 	}
// 	req := msg.FileSavePart{
// 		TotalParts: totalParts,
// 		Bytes:      bytes,
// 		FileID:     fileID,
// 		PartID:     partID,
// 	}
// 	envelop.Message, _ = req.Marshal()
//
// 	logs.Debug("FileCtrl generates FileSavePart",
// 		zap.Int64("MsgID", ctx.req.MessageID),
// 		zap.Int64("FileID", req.FileID),
// 		zap.Int32("PartID", req.PartID),
// 		zap.Int32("TotalParts", req.TotalParts),
// 		zap.Int("Bytes", len(req.Bytes)),
// 	)
// 	return &envelop
// }
//
// func (ctx *uploadContext) execute(ctrl *Controller) domain.RequestStatus {
// 	for {
// 		ctx.prepare()
// 		logs.Info("FileCtrl executes Upload",
// 			zap.Int64("FileID", ctx.req.FileID),
// 			zap.Int32("TotalParts", ctx.req.TotalParts),
// 			zap.Int32("ChunkSize", ctx.req.ChunkSize),
// 		)
//
// 		maxRetries := int32(math.Min(float64(ctx.req.MaxInFlights), float64(ctx.req.TotalParts)))
// 		waitGroup := sync.WaitGroup{}
// 		for maxRetries > 0 {
// 			select {
// 			case partIndex := <-ctx.parts:
// 				if !ctrl.existUploadRequest(ctx.req.GetID()) {
// 					waitGroup.Wait()
// 					_ = ctx.file.Close()
// 					logs.Warn("Upload Canceled (Request Not Exists)",
// 						zap.Int64("FileID", ctx.req.FileID),
// 						zap.Int64("Size", ctx.req.FileSize),
// 						zap.String("Path", ctx.req.FilePath),
// 					)
// 					if !ctx.req.SkipDelegateCall {
// 						ctrl.onCancel(ctx.req.GetID(), 0, ctx.req.FileID, 0, false, ctx.req.PeerID)
// 					}
// 					return domain.RequestStatusCanceled
// 				}
// 				waitGroup.Add(1)
// 				ctx.rateLimit <- struct{}{}
// 				go ctx.uploadJob(ctrl, &maxRetries, &waitGroup, partIndex)
// 			default:
// 				switch int32(len(ctx.req.UploadedParts)) {
// 				case ctx.req.TotalParts - 1:
// 					logs.Debug("FileCtrl waits for all (n-1) parts uploads to complete")
// 					waitGroup.Wait()
// 					ctx.parts <- ctx.req.TotalParts - 1
// 				case ctx.req.TotalParts:
// 					logs.Debug("FileCtrl waits for last part to upload")
// 					waitGroup.Wait()
// 					if !ctrl.postUploadProcess(ctx.req) {
// 						ctrl.onCancel(ctx.req.GetID(), 0, ctx.req.FileID, 0, true, ctx.req.PeerID)
// 						return domain.RequestStatusError
// 					}
// 					// We have finished our uploads
// 					_ = ctx.file.Close()
// 					if !ctx.req.SkipDelegateCall {
// 						ctrl.onCompleted(ctx.req.GetID(), 0, ctx.req.FileID, 0, ctx.req.FilePath, ctx.req.PeerID)
// 					}
// 					_ = repo.Files.MarkAsUploaded(ctx.req.FileID)
// 					return domain.RequestStatusCompleted
// 				default:
// 					// Keep Uploading
// 					time.Sleep(time.Millisecond * 500)
// 				}
// 			}
// 		}
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
// func (ctx *uploadContext) uploadJob(ctrl *Controller, maxRetries *int32, waitGroup *sync.WaitGroup, partIndex int32) {
// 	defer waitGroup.Done()
// 	defer func() {
// 		<-ctx.rateLimit
// 	}()
//
// 	bytes := pbytes.GetLen(int(ctx.req.ChunkSize))
// 	defer pbytes.Put(bytes)
// 	offset := partIndex * ctx.req.ChunkSize
// 	n, err := ctx.file.ReadAt(bytes, int64(offset))
// 	if err != nil && err != io.EOF {
// 		logs.Warn("Error in ReadFile", zap.Error(err))
// 		atomic.StoreInt32(maxRetries, 0)
// 		ctx.parts <- partIndex
// 		return
// 	}
// 	if n == 0 {
// 		return
// 	}
// 	res, err := ctrl.network.SendHttp(
// 		ctx.req.httpContext,
// 		ctx.generateFileSavePart(ctx.req.FileID, partIndex+1, ctx.req.TotalParts, bytes[:n]),
// 	)
// 	if err != nil {
// 		logs.Warn("Error On Http Response", zap.Error(err))
// 		switch e := err.(type) {
// 		case *url.Error:
// 			if e.Timeout() {
// 				atomic.AddInt32(maxRetries, -1)
// 			}
// 		default:
// 		}
// 		time.Sleep(100 * time.Millisecond)
// 		ctx.parts <- partIndex
// 		return
// 	}
// 	switch res.Constructor {
// 	case msg.C_Bool:
// 		ctx.addToUploaded(ctrl, partIndex)
// 	case msg.C_Error:
// 		x := &msg.Error{}
// 		_ = x.Unmarshal(res.Message)
// 		logs.Debug("FileCtrl received Error response",
// 			zap.Int32("PartID", partIndex+1),
// 			zap.String("Code", x.Code),
// 			zap.String("Item", x.Items),
// 		)
// 		ctx.parts <- partIndex
// 	default:
// 		logs.Debug("FileCtrl received unexpected response", zap.String("C", msg.ConstructorNames[res.Constructor]))
// 		atomic.StoreInt32(maxRetries, 0)
// 		return
// 	}
// }
