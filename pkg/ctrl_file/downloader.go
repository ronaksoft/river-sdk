package fileCtrl

import (
	"context"
	networkCtrl "git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_network"
	"os"
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

type downloadStateFunc func(status *downloadStatus) downloadStateFunc

type DownloadRequest struct {
	Ctx         context.Context
	NetworkCtrl *networkCtrl.Controller
	// MaxRetries defines how many time each request could encounter error before giving up
	MaxRetries int
	// MessageID (Optional) if is set then (ClusterID, FileID, AccessHash, Version) will be read from the message
	// document object, or if message has no document then return error
	MessageID int64
	// ClusterID, FileID, AccessHash and Version identify the file address which needs to be downloaded
	ClusterID  int32
	FileID     int64
	AccessHash int64
	Version    int32
	// FileSize (Optional) if is set then progress will be calculated
	FileSize int64
	// ChunkSize identifies how many request we need to send to server to download a file.
	ChunkSize int32
	// MaxInFlights defines that how many requests could be send concurrently
	MaxInFlights int
	// FilePath defines the path which downloaded file will be stored. It must be a file not a directory.
	// Also it will be overwritten if Overwrite is TRUE
	FilePath string
}

type DownloadTracker struct {
	Request  DownloadRequest

}

type downloadStatus struct {
	ctx             context.Context
	cancel          context.CancelFunc
	rateLimit       chan struct{}
	file            *os.File
	Request         DownloadRequest
	StartTime       time.Time
}

func (ctrl *Controller) download(req DownloadRequest) {
	ctx, cancel := context.WithCancel(req.Ctx)
	ds := &downloadStatus{
		ctx:        ctx,
		cancel:     cancel,
		rateLimit:  make(chan struct{}, req.MaxInFlights),
		Request:    req,
		StartTime:  time.Now(),
	}


	if req.FileSize > 0 {
		err := os.Truncate(req.FilePath, req.FileSize)
		if err != nil {
			if req.MessageID != 0 {
				ctrl.delegate.OnDownloadError(req.MessageID, 0, req.FilePath, nil)
			}
			return
		}
	} else {

	}

	// Initial State
	stateFunc := getNextPart
	for {
		select {
		case <-ds.ctx.Done():
			return
		default:

		}
		if stateFunc = stateFunc(ds); stateFunc == nil {
			return
		}
	}

}

func computeParts(ds *downloadStatus) downloadStateFunc {
	panic("implement it")
}

func getFilePath(ds *downloadStatus) downloadStateFunc {
	panic("implement it")
}

func getNextPart(ds *downloadStatus) downloadStateFunc {
	panic("implement it")
}

func wait(ds *downloadStatus) downloadStateFunc {
	panic("implement it")
}

// func generateFileGet(job *downloadJob) *msg.MessageEnvelope {
// 	req := new(msg.FileGet)
// 	req.Location = &msg.InputFileLocation{
// 		ClusterID:  job.ClusterID,
// 		FileID:     job.FileID,
// 		AccessHash: job.AccessHash,
// 		Version:    job.Version,
// 	}
// 	req.Offset = job.Offset
// 	req.Limit = job.Limit
//
// 	envelop := new(msg.MessageEnvelope)
// 	envelop.Constructor = msg.C_FileGet
// 	envelop.Message, _ = req.Marshal()
// 	envelop.RequestID = uint64(domain.SequentialUniqueID())
// 	logs.Debug("FilesStatus::generateFileGet()",
// 		zap.Int64("MsgID", job.MessageID),
// 		zap.Int32("Offset", req.Offset),
// 		zap.Int32("Limit", req.Limit),
// 		zap.Int64("FileID", req.Location.FileID),
// 		zap.Uint64("AccessHash", req.Location.AccessHash),
// 		zap.Int32("ClusterID", req.Location.ClusterID),
// 		zap.Int32("Version", req.Location.Version),
// 	)
// 	return envelop
// }
