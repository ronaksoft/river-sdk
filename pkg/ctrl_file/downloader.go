package fileCtrl

import (
	"fmt"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"go.uber.org/zap"
	"os"
	"sync"
	"sync/atomic"
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
	// MaxRetries defines how many time each request could encounter error before giving up
	MaxRetries int32 `json:"max_retries"`
	// MessageID (Optional) if is set then (ClusterID, FileID, AccessHash, Version) will be read from the message
	// document object, or if message has no document then return error
	MessageID int64 `json:"message_id"`
	// ClusterID, FileID, AccessHash and Version identify the file address which needs to be downloaded
	ClusterID  int32  `json:"cluster_id"`
	FileID     int64  `json:"file_id"`
	AccessHash uint64 `json:"access_hash"`
	Version    int32  `json:"version"`
	// FileSize (Optional) if is set then progress will be calculated
	FileSize int64 `json:"file_size"`
	// ChunkSize identifies how many request we need to send to server to Download a file.
	ChunkSize int32 `json:"chunk_size"`
	// MaxInFlights defines that how many requests could be send concurrently
	MaxInFlights int `json:"max_in_flights"`
	// FilePath defines the path which downloaded file will be stored. It must be a file not a directory.
	// Also it will be overwritten if Overwrite is TRUE
	FilePath        string  `json:"file_path"`
	TempFilePath    string  `json:"temp_file_path"`
	DownloadedParts []int32 `json:"downloaded_parts"`
	TotalParts      int32   `json:"total_parts"`
	Canceled        bool    `json:"canceled"`
	// SkipDelegateCall identifies to call delegate function on specified states
	SkipDelegateCall bool `json:"skip_delegate_call"`
}

func (r DownloadRequest) GetID() string {
	return fmt.Sprintf("%d.%d.%d", r.ClusterID, r.FileID, r.AccessHash)
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
	for _, index := range ctx.req.DownloadedParts {
		if partIndex == index {
			return true
		}
	}
	return false
}

func (ctx *downloadContext) addToDownloaded(ctrl *Controller, partIndex int32) {
	ctx.mtx.Lock()
	ctx.req.DownloadedParts = append(ctx.req.DownloadedParts, partIndex)
	progress := int64(float64(len(ctx.req.DownloadedParts)) / float64(ctx.req.TotalParts) * 100)
	skipOnProgress := false
	if ctx.lastProgress > progress {
		skipOnProgress = true
	} else {
		ctx.lastProgress = progress
	}
	ctx.mtx.Unlock()
	ctrl.saveDownloads(ctx.req)

	if !ctx.req.SkipDelegateCall && !skipOnProgress {
		ctrl.onProgressChanged(ctx.req.GetID(), ctx.req.ClusterID, ctx.req.FileID, int64(ctx.req.AccessHash), progress)
	}
}

func (ctx *downloadContext) generateFileGet(offset, limit int32) *msg.MessageEnvelope {
	req := new(msg.FileGet)
	req.Location = &msg.InputFileLocation{
		ClusterID:  ctx.req.ClusterID,
		FileID:     ctx.req.FileID,
		AccessHash: ctx.req.AccessHash,
		Version:    ctx.req.Version,
	}
	req.Offset = offset
	req.Limit = limit

	envelop := new(msg.MessageEnvelope)
	envelop.Constructor = msg.C_FileGet
	envelop.Message, _ = req.Marshal()
	envelop.RequestID = uint64(domain.SequentialUniqueID())

	return envelop
}

func (ctx *downloadContext) execute(ctrl *Controller) domain.RequestStatus {
	waitGroup := sync.WaitGroup{}
	for ctx.req.MaxRetries > 0 {
		select {
		case partIndex := <-ctx.parts:
			if !ctrl.existDownloadRequest(ctx.req.GetID()) {
				waitGroup.Wait()
				_ = ctx.file.Close()
				_ = os.Remove(ctx.req.TempFilePath)
				if !ctx.req.SkipDelegateCall {
					ctrl.onCancel(ctx.req.GetID(), ctx.req.ClusterID, ctx.req.FileID, int64(ctx.req.AccessHash), false)
				}

				return domain.RequestStatusCanceled
			}
			ctx.rateLimit <- struct{}{}
			waitGroup.Add(1)

			go func(partIndex int32) {
				defer waitGroup.Done()
				defer func() {
					<-ctx.rateLimit
				}()

				offset := partIndex * ctx.req.ChunkSize
				res, err := ctrl.network.SendHttp(ctx.generateFileGet(offset, ctx.req.ChunkSize))
				if err != nil {
					logs.Warn("Error in SentHTTP", zap.Error(err))
					atomic.AddInt32(&ctx.req.MaxRetries, -1)
					ctx.parts <- partIndex
					return
				}
				switch res.Constructor {
				case msg.C_File:
					file := new(msg.File)
					err = file.Unmarshal(res.Message)
					if err != nil {
						logs.Warn("Error in Unmarshal",
							zap.Error(err),
							zap.Int32("Offset", offset),
							zap.Int("Byte", len(file.Bytes)),
						)
						atomic.AddInt32(&ctx.req.MaxRetries, -1)
						ctx.parts <- partIndex
						return
					}
					_, err := ctx.file.WriteAt(file.Bytes, int64(offset))
					if err != nil {
						logs.Error("Error in WriteFile",
							zap.Error(err),
							zap.Int32("Offset", offset),
							zap.Int("Byte", len(file.Bytes)),
						)
						atomic.AddInt32(&ctx.req.MaxRetries, -1)
						ctx.parts <- partIndex
						return
					}
					ctx.addToDownloaded(ctrl, partIndex)
				default:
					atomic.AddInt32(&ctx.req.MaxRetries, -1)
					ctx.parts <- partIndex
					return
				}
			}(partIndex)
		default:
			waitGroup.Wait()
			if int32(len(ctx.req.DownloadedParts)) == ctx.req.TotalParts {
				_ = ctx.file.Close()
				err := os.Rename(ctx.req.TempFilePath, ctx.req.FilePath)
				if err != nil {
					_ = os.Remove(ctx.req.TempFilePath)
					if !ctx.req.SkipDelegateCall {
						ctrl.onCancel(ctx.req.GetID(), ctx.req.ClusterID, ctx.req.FileID, int64(ctx.req.AccessHash), true)
					}
					return domain.RequestStatusError
				}
				if !ctx.req.SkipDelegateCall {
					ctrl.onCompleted(ctx.req.GetID(), ctx.req.ClusterID, ctx.req.FileID, int64(ctx.req.AccessHash), ctx.req.FilePath)
				}

				return domain.RequestStatusCompleted
			}
		}
	}
	_ = ctx.file.Close()
	_ = os.Remove(ctx.req.TempFilePath)
	if !ctx.req.SkipDelegateCall {
		ctrl.onCancel(ctx.req.GetID(), ctx.req.ClusterID, ctx.req.FileID, int64(ctx.req.AccessHash), true)
	}

	return domain.RequestStatusError
}
