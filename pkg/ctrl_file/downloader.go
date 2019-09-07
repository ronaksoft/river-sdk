package fileCtrl

import (
	"context"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	networkCtrl "git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_network"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"go.uber.org/zap"
	"os"
	"sync"
	"time"
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
	MaxRetries int `json:"max_retries"`
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
	// ChunkSize identifies how many request we need to send to server to download a file.
	ChunkSize int32 `json:"chunk_size"`
	// MaxInFlights defines that how many requests could be send concurrently
	MaxInFlights int `json:"max_in_flights"`
	// FilePath defines the path which downloaded file will be stored. It must be a file not a directory.
	// Also it will be overwritten if Overwrite is TRUE
	FilePath        string  `json:"file_path"`
	DownloadedParts []int32 `json:"downloaded_parts"`
}

type downloadStatus struct {
	mtx         sync.Mutex              `json:"-"`
	ctx         context.Context         `json:"-"`
	rateLimit   chan struct{}           `json:"-"`
	parts       chan int32              `json:"-"`
	file        *os.File                `json:"-"`
	networkCtrl *networkCtrl.Controller `json:"-"`

	Request         DownloadRequest      `json:"request"`
	Status          domain.RequestStatus `json:"status"`
	StartTime       time.Time            `json:"start_time"`
	TotalParts      int32                `json:"total_parts"`
}

func (ds *downloadStatus) isDownloaded(partIndex int32) bool {
	ds.mtx.Lock()
	defer ds.mtx.Unlock()
	for _, index := range ds.Request.DownloadedParts {
		if partIndex == index {
			return true
		}
	}
	return false
}

func (ds *downloadStatus) addToDownloaded(partIndex int32) {
	ds.mtx.Lock()
	ds.Request.DownloadedParts = append(ds.Request.DownloadedParts, partIndex)
	downloads[ds.Request.MessageID] = ds.Request
	saveSnapshot.EnterWithResult(nil, nil)
	ds.mtx.Unlock()
}

func (ds *downloadStatus) generateFileGet(offset, limit int32) *msg.MessageEnvelope {
	req := new(msg.FileGet)
	req.Location = &msg.InputFileLocation{
		ClusterID:  ds.Request.ClusterID,
		FileID:     ds.Request.FileID,
		AccessHash: ds.Request.AccessHash,
		Version:    ds.Request.Version,
	}
	req.Offset = offset
	req.Limit = limit

	envelop := new(msg.MessageEnvelope)
	envelop.Constructor = msg.C_FileGet
	envelop.Message, _ = req.Marshal()
	envelop.RequestID = uint64(domain.SequentialUniqueID())
	logs.Debug("FilesStatus::generateFileGet()",
		zap.Int64("MsgID", ds.Request.MessageID),
		zap.Int32("Offset", req.Offset),
		zap.Int32("Limit", req.Limit),
		zap.Int64("FileID", req.Location.FileID),
		zap.Uint64("AccessHash", req.Location.AccessHash),
		zap.Int32("ClusterID", req.Location.ClusterID),
		zap.Int32("Version", req.Location.Version),
	)
	return envelop
}

func (ds *downloadStatus) run() {
	ds.Status = domain.RequestStatusInProgress
	waitGroup := sync.WaitGroup{}
	for partIndex := range ds.parts {
		ds.rateLimit <- struct{}{}
		waitGroup.Add(1)

		go func(partIndex int32) {
			defer waitGroup.Done()
			defer func() {
				<-ds.rateLimit
			}()

			offset := partIndex * ds.Request.ChunkSize
			res, err := ds.networkCtrl.SendHttp(ds.generateFileGet(offset, ds.Request.ChunkSize))
			if err != nil {
				ds.parts <- partIndex
				return
			}
			switch res.Constructor {
			case msg.C_File:
				file := new(msg.File)
				_ = file.Unmarshal(res.Message)
				_, err := ds.file.WriteAt(file.Bytes, int64(offset))
				if err != nil {
					ds.parts <- partIndex
					return
				}
				ds.addToDownloaded(partIndex)
			default:
				ds.parts <- partIndex
				return
			}
		}(partIndex)

	}
	waitGroup.Wait()
	ds.Status = domain.RequestStatusCompleted
}
