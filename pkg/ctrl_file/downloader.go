package fileCtrl

import (
	"context"
	"git.ronaksoft.com/river/msg/msg"
	"git.ronaksoft.com/river/sdk/pkg/ctrl_file/executor"
	"git.ronaksoft.com/river/sdk/pkg/domain"
	"os"
	"sync"
)

/*
   Creation Time: 2019 - Aug - 18
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

type DownloadRequest struct {
	msg.ClientFileRequest
}

func (d *DownloadRequest) GetID() string {
	return getRequestID(d.ClusterID, d.FileID, d.AccessHash)
}

func (d *DownloadRequest) Prepare() {

}

func (d *DownloadRequest) NextAction() executor.Action {
	panic("implement me")
}

func (d *DownloadRequest) ActionDone(id int32) {
	panic("implement me")
}

func (d *DownloadRequest) Serialize() []byte {
	b, err := d.Marshal()
	if err != nil {
		panic(err)
	}
	return b
}

func (d *DownloadRequest) Deserialize(b []byte) {
	_ = d.Unmarshal(b)
}

type DownloadAction struct {
	id  int32
	req *DownloadRequest
}

func (a *DownloadAction) ID() int32 {
	return a.id
}

func (a *DownloadAction) Do(ctx context.Context) {

}

func (a *DownloadAction) generateFileGet(offset, limit int32) *msg.MessageEnvelope {
	req := new(msg.FileGet)
	req.Location = &msg.InputFileLocation{
		ClusterID:  a.req.ClusterID,
		FileID:     a.req.FileID,
		AccessHash: a.req.AccessHash,
		Version:    a.req.Version,
	}
	req.Offset = offset
	req.Limit = limit

	envelop := new(msg.MessageEnvelope)
	envelop.Constructor = msg.C_FileGet
	envelop.Message, _ = req.Marshal()
	envelop.RequestID = uint64(domain.SequentialUniqueID())

	return envelop
}

type downloadContext struct {
	mtx          sync.Mutex
	rateLimit    chan struct{}
	parts        chan int32
	file         *os.File
	req          DownloadRequest
	lastProgress int64
}

func (ctx *downloadContext) isDownloaded(partIndex int32) bool {
	ctx.mtx.Lock()
	defer ctx.mtx.Unlock()
	for _, index := range ctx.req.FinishedParts {
		if partIndex == index {
			return true
		}
	}
	return false
}

func (ctx *downloadContext) addToDownloaded(ctrl *Controller, partIndex int32) {
	ctx.mtx.Lock()
	ctx.req.FinishedParts = append(ctx.req.FinishedParts, partIndex)
	progress := int64(float64(len(ctx.req.FinishedParts)) / float64(ctx.req.TotalParts) * 100)
	skipOnProgress := false
	if ctx.lastProgress > progress {
		skipOnProgress = true
	} else {
		ctx.lastProgress = progress
	}
	ctx.mtx.Unlock()
	// ctrl.saveDownloads(ctx.req)

	if !ctx.req.SkipDelegateCall && !skipOnProgress {
		ctrl.onProgressChanged(ctx.req.GetID(), ctx.req.ClusterID, ctx.req.FileID, int64(ctx.req.AccessHash), progress, ctx.req.PeerID)
	}
}

func (ctx *downloadContext) execute(ctrl *Controller) domain.RequestStatus {
	// waitGroup := sync.WaitGroup{}
	// for ctx.req.MaxRetries > 0 {
	// 	select {
	// 	case partIndex := <-ctx.parts:
	// 		if !ctrl.existDownloadRequest(ctx.req.GetID()) {
	// 			waitGroup.Wait()
	// 			_ = ctx.file.Close()
	// 			_ = os.Remove(ctx.req.TempFilePath)
	// 			if !ctx.req.SkipDelegateCall {
	// 				ctrl.onCancel(ctx.req.GetID(), ctx.req.ClusterID, ctx.req.FileID, int64(ctx.req.AccessHash), false, ctx.req.PeerID)
	// 			}
	// 			return domain.RequestStatusCanceled
	// 		}
	// 		ctx.rateLimit <- struct{}{}
	// 		waitGroup.Add(1)
	// 		go func(partIndex int32) {
	// 			defer waitGroup.Done()
	// 			defer func() {
	// 				<-ctx.rateLimit
	// 			}()
	//
	// 			offset := partIndex * ctx.req.ChunkSize
	// 			res, err := ctrl.network.SendHttp(ctx.req.httpContext, ctx.generateFileGet(offset, ctx.req.ChunkSize))
	// 			if err != nil {
	// 				logs.Warn("Downloader got error from NetworkController SendHTTP", zap.Error(err))
	// 				atomic.AddInt32(&ctx.req.MaxRetries, -1)
	// 				ctx.parts <- partIndex
	// 				return
	// 			}
	// 			switch res.Constructor {
	// 			case msg.C_File:
	// 				file := new(msg.File)
	// 				err = file.Unmarshal(res.Message)
	// 				if err != nil {
	// 					logs.Warn("Downloader couldn't unmarshal server response FileGet (File)",
	// 						zap.Error(err),
	// 						zap.Int32("Offset", offset),
	// 						zap.Int("Byte", len(file.Bytes)),
	// 					)
	// 					atomic.AddInt32(&ctx.req.MaxRetries, -1)
	// 					ctx.parts <- partIndex
	// 					return
	// 				}
	// 				_, err := ctx.file.WriteAt(file.Bytes, int64(offset))
	// 				if err != nil {
	// 					logs.Error("Downloader couldn't write to file, will retry...",
	// 						zap.Error(err),
	// 						zap.Int32("Offset", offset),
	// 						zap.Int("Byte", len(file.Bytes)),
	// 					)
	// 					atomic.AddInt32(&ctx.req.MaxRetries, -1)
	// 					ctx.parts <- partIndex
	// 					return
	// 				}
	// 				ctx.addToDownloaded(ctrl, partIndex)
	// 			default:
	// 				atomic.AddInt32(&ctx.req.MaxRetries, -1)
	// 				ctx.parts <- partIndex
	// 				return
	// 			}
	// 		}(partIndex)
	// 	default:
	// 		waitGroup.Wait()
	// 		totalDownloadedParts := int32(len(ctx.req.DownloadedParts))
	// 		switch {
	// 		case totalDownloadedParts > ctx.req.TotalParts:
	// 			ctx.req.DownloadedParts = unique(ctx.req.DownloadedParts)
	// 			totalDownloadedParts = int32(len(ctx.req.DownloadedParts))
	// 			if totalDownloadedParts < ctx.req.TotalParts {
	// 				break
	// 			}
	// 			fallthrough
	// 		case totalDownloadedParts == ctx.req.TotalParts:
	// 			_ = ctx.file.Close()
	// 			err := os.Rename(ctx.req.TempFilePath, ctx.req.FilePath)
	// 			if err != nil {
	// 				_ = os.Remove(ctx.req.TempFilePath)
	// 				if !ctx.req.SkipDelegateCall {
	// 					ctrl.onCancel(ctx.req.GetID(), ctx.req.ClusterID, ctx.req.FileID, int64(ctx.req.AccessHash), true, ctx.req.PeerID)
	// 				}
	// 				return domain.RequestStatusError
	// 			}
	// 			if !ctx.req.SkipDelegateCall {
	// 				ctrl.onCompleted(ctx.req.GetID(), ctx.req.ClusterID, ctx.req.FileID, int64(ctx.req.AccessHash), ctx.req.FilePath, ctx.req.PeerID)
	// 			}
	//
	// 			return domain.RequestStatusCompleted
	// 		default:
	// 			time.Sleep(time.Millisecond * 250)
	// 			// Keep downloading
	// 		}
	// 	}
	// }
	// _ = ctx.file.Close()
	// _ = os.Remove(ctx.req.TempPath)
	// if !ctx.req.SkipDelegateCall {
	// 	ctrl.onCancel(ctx.req.GetID(), ctx.req.ClusterID, ctx.req.FileID, int64(ctx.req.AccessHash), true, ctx.req.PeerID)
	// }
	//
	return domain.RequestStatusError
}
