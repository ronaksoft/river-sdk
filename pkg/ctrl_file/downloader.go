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
	DownloadedParts []int32 `json:"downloaded_parts"`
	TotalParts      int32   `json:"total_parts"`
}

type downloadStatus struct {
	mtx       sync.Mutex
	rateLimit chan struct{}
	parts     chan int32
	file      *os.File
	ctrl      *Controller
	req       DownloadRequest
}

func (ds *downloadStatus) isDownloaded(partIndex int32) bool {
	ds.mtx.Lock()
	defer ds.mtx.Unlock()
	for _, index := range ds.req.DownloadedParts {
		if partIndex == index {
			return true
		}
	}
	return false
}

func (ds *downloadStatus) addToDownloaded(partIndex int32) {
	ds.mtx.Lock()
	ds.req.DownloadedParts = append(ds.req.DownloadedParts, partIndex)
	progress := int64(float64(len(ds.req.DownloadedParts)) / float64(ds.req.TotalParts) * 100)
	ds.mtx.Unlock()
	ds.ctrl.saveDownloads(ds.req)
	ds.ctrl.onProgressChanged(ds.req.MessageID, progress)
}

func (ds *downloadStatus) generateFileGet(offset, limit int32) *msg.MessageEnvelope {
	req := new(msg.FileGet)
	req.Location = &msg.InputFileLocation{
		ClusterID:  ds.req.ClusterID,
		FileID:     ds.req.FileID,
		AccessHash: ds.req.AccessHash,
		Version:    ds.req.Version,
	}
	req.Offset = offset
	req.Limit = limit

	envelop := new(msg.MessageEnvelope)
	envelop.Constructor = msg.C_FileGet
	envelop.Message, _ = req.Marshal()
	envelop.RequestID = uint64(domain.SequentialUniqueID())
	logs.Debug("FilesStatus::generateFileGet()",
		zap.Int64("MsgID", ds.req.MessageID),
		zap.Int32("Offset", req.Offset),
		zap.Int32("Limit", req.Limit),
		zap.Int64("FileID", req.Location.FileID),
		zap.Uint64("AccessHash", req.Location.AccessHash),
		zap.Int32("ClusterID", req.Location.ClusterID),
		zap.Int32("Version", req.Location.Version),
	)
	return envelop
}

func (ds *downloadStatus) execute() domain.RequestStatus {
	waitGroup := sync.WaitGroup{}

	for ds.req.MaxRetries > 0 {
		select {
		case partIndex := <-ds.parts:
			ds.rateLimit <- struct{}{}
			waitGroup.Add(1)

			go func(partIndex int32) {
				defer waitGroup.Done()
				defer func() {
					<-ds.rateLimit
				}()

				offset := partIndex * ds.req.ChunkSize
				res, err := ds.ctrl.network.SendHttp(ds.generateFileGet(offset, ds.req.ChunkSize))
				if err != nil {
					logs.Warn("Error in SentHTTP", zap.Error(err))
					atomic.AddInt32(&ds.req.MaxRetries, -1)
					ds.parts <- partIndex
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
						atomic.AddInt32(&ds.req.MaxRetries, -1)
						ds.parts <- partIndex
						return
					}
					_, err := ds.file.WriteAt(file.Bytes, int64(offset))
					if err != nil {
						logs.Warn("Error in WriteFile",
							zap.Error(err),
							zap.Int32("Offset", offset),
							zap.Int("Byte", len(file.Bytes)),
						)
						atomic.AddInt32(&ds.req.MaxRetries, -1)
						ds.parts <- partIndex
						return
					}
					ds.addToDownloaded(partIndex)
				default:
					atomic.AddInt32(&ds.req.MaxRetries, -1)
					ds.parts <- partIndex
					return
				}
			}(partIndex)
		default:
			waitGroup.Wait()
			if int32(len(ds.req.DownloadedParts)) == ds.req.TotalParts {
				_ = ds.file.Close()
				ds.ctrl.onCompleted(ds.req.MessageID, ds.req.FilePath)
				return domain.RequestStatusCompleted
			}
		}
	}
	ds.ctrl.onError(ds.req.MessageID, ds.req.FilePath, ronak.StrToByte("max retry exceeded without success"))
	return domain.RequestStatusError
}
