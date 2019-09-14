package fileCtrl

import (
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	ronak "git.ronaksoftware.com/ronak/toolbox"
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
}

type downloadContext struct {
	mtx       sync.Mutex
	rateLimit chan struct{}
	parts     chan int32
	file      *os.File
	ctrl      *Controller
	req       DownloadRequest
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

func (ctx *downloadContext) addToDownloaded(partIndex int32) {
	ctx.mtx.Lock()
	ctx.req.DownloadedParts = append(ctx.req.DownloadedParts, partIndex)
	progress := int64(float64(len(ctx.req.DownloadedParts)) / float64(ctx.req.TotalParts) * 100)
	ctx.mtx.Unlock()
	ctx.ctrl.saveDownloads(ctx.req)
	ctx.ctrl.onProgressChanged(ctx.req.MessageID, progress)
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
	logs.Debug("FilesStatus::generateFileGet()",
		zap.Int64("MsgID", ctx.req.MessageID),
		zap.Int32("Offset", req.Offset),
		zap.Int32("Limit", req.Limit),
		zap.Int64("FileID", req.Location.FileID),
		zap.Uint64("AccessHash", req.Location.AccessHash),
		zap.Int32("ClusterID", req.Location.ClusterID),
		zap.Int32("Version", req.Location.Version),
	)
	return envelop
}

func (ctx *downloadContext) execute() domain.RequestStatus {
	waitGroup := sync.WaitGroup{}

	for ctx.req.MaxRetries > 0 {
		select {
		case partIndex := <-ctx.parts:
			ctx.rateLimit <- struct{}{}
			waitGroup.Add(1)

			go func(partIndex int32) {
				defer waitGroup.Done()
				defer func() {
					<-ctx.rateLimit
				}()

				offset := partIndex * ctx.req.ChunkSize
				res, err := ctx.ctrl.network.SendHttp(ctx.generateFileGet(offset, ctx.req.ChunkSize))
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
						logs.Warn("Error in WriteFile",
							zap.Error(err),
							zap.Int32("Offset", offset),
							zap.Int("Byte", len(file.Bytes)),
						)
						atomic.AddInt32(&ctx.req.MaxRetries, -1)
						ctx.parts <- partIndex
						return
					}
					ctx.addToDownloaded(partIndex)
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
					ctx.ctrl.onError(ctx.req.MessageID, ctx.req.TempFilePath, ronak.StrToByte(err.Error()))
					return domain.RequestStatusError
				}
				ctx.ctrl.onCompleted(ctx.req.MessageID, ctx.req.FilePath)
				return domain.RequestStatusCompleted
			}
		}
	}
	_ = ctx.file.Close()
	_ = os.Remove(ctx.req.TempFilePath)
	ctx.ctrl.onError(ctx.req.MessageID, ctx.req.TempFilePath, ronak.StrToByte("max retry exceeded without success"))
	return domain.RequestStatusError
}
