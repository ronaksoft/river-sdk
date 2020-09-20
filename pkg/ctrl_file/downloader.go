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
	lastProgress int64
	mtx sync.Mutex
	file  *os.File
	ctrl  *Controller
	parts chan int32
}

func (d *DownloadRequest) isDownloaded(partIndex int32) bool {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	for _, index := range d.FinishedParts {
		if partIndex == index {
			return true
		}
	}
	return false
}

func (d *DownloadRequest) addToDownloaded(ctrl *Controller, partIndex int32) {
	d.mtx.Lock()
	d.FinishedParts = append(d.FinishedParts, partIndex)
	progress := int64(float64(len(d.FinishedParts)) / float64(d.TotalParts) * 100)
	skipOnProgress := false
	if d.lastProgress > progress {
		skipOnProgress = true
	} else {
		d.lastProgress = progress
	}
	d.mtx.Unlock()
	// ctrl.saveDownloads(ctx.req)

	if !d.SkipDelegateCall && !skipOnProgress {
		ctrl.onProgressChanged(d.GetID(),d.ClusterID, d.FileID, int64(d.AccessHash), progress, d.PeerID)
	}
}

func NewDownloadRequest(ctrl *Controller, fr msg.ClientFileRequest) DownloadRequest {
	return DownloadRequest{
		ClientFileRequest: fr,
		ctrl:              ctrl,
		file:              nil,
	}
}

func (d *DownloadRequest) GetID() string {
	return getRequestID(d.ClusterID, d.FileID, d.AccessHash)
}

func (d *DownloadRequest) Prepare() error {
	_, err := os.Stat(d.TempPath)
	if err != nil {
		if os.IsNotExist(err) {
			d.file, err = os.Create(d.TempPath)
			if err != nil {
				d.ctrl.onCancel(d.GetID(), d.ClusterID, d.FileID, int64(d.AccessHash), true, d.PeerID)
				return err
			}
		} else {
			d.ctrl.onCancel(d.GetID(), d.ClusterID, d.FileID, int64(d.AccessHash), true, d.PeerID)
			return err
		}
	} else {
		d.file, err = os.OpenFile(d.TempPath, os.O_RDWR, 0666)
		if err != nil {
			d.ctrl.onCancel(d.GetID(), d.ClusterID, d.FileID, int64(d.AccessHash), true, d.PeerID)
			return err
		}
	}

	if d.FileSize > 0 {
		err := os.Truncate(d.TempPath, d.FileSize)
		if err != nil {
			d.ctrl.onCancel(d.GetID(), d.ClusterID, d.FileID, int64(d.AccessHash), true, d.PeerID)
			return err
		}
		dividend := int32(d.FileSize / int64(d.ChunkSize))
		if d.FileSize%int64(d.ChunkSize) > 0 {
			d.TotalParts = dividend + 1
		} else {
			d.TotalParts = dividend
		}
	} else {
		d.TotalParts = 1
		d.ChunkSize = 0
	}

	for partIndex := int32(0); partIndex < d.TotalParts; partIndex++ {
		if d.isDownloaded(partIndex) {
			continue
		}
		d.parts <- partIndex
	}
	return nil
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
