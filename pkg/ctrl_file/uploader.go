package fileCtrl

import (
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"github.com/gobwas/pool/pbytes"
	"go.uber.org/zap"
	"os"
	"sync"
	"sync/atomic"
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
	// MaxRetries defines how many time each request could encounter error before giving up
	MaxRetries int32 `json:"max_retries"`
	// MessageID (Optional) if is set then (ClusterID, FileID, AccessHash, Version) will be read from the message
	// document object, or if message has no document then return error
	MessageID int64 `json:"message_id"`
	FileID    int64 `json:"file_id"`
	// FileSize (Optional) if is set then progress will be calculated
	FileSize int64 `json:"file_size"`
	// ChunkSize identifies how many request we need to send to server to Download a file.
	ChunkSize int32 `json:"chunk_size"`
	// MaxInFlights defines that how many requests could be send concurrently
	MaxInFlights int `json:"max_in_flights"`
	// FilePath defines the path which downloaded file will be stored. It must be a file not a directory.
	// Also it will be overwritten if Overwrite is TRUE
	FilePath      string               `json:"file_path"`
	UploadedParts []int32              `json:"downloaded_parts"`
	TotalParts    int32                `json:"total_parts"`
	Status        domain.RequestStatus `json:"status"`
}

type uploadStatus struct {
	mtx       sync.Mutex
	rateLimit chan struct{}
	parts     chan int32
	file      *os.File
	ctrl      *Controller
	Request   UploadRequest
}

func (us *uploadStatus) isUploaded(partIndex int32) bool {
	us.mtx.Lock()
	defer us.mtx.Unlock()
	for _, index := range us.Request.UploadedParts {
		if partIndex == index {
			return true
		}
	}
	return false
}

func (us *uploadStatus) addToUploaded(partIndex int32) {
	us.mtx.Lock()
	us.Request.UploadedParts = append(us.Request.UploadedParts, partIndex)
	us.mtx.Unlock()
	us.ctrl.saveUploads(us.Request)
}

func (us *uploadStatus) generateFileSavePart(fileID int64, partID int32, totalParts int32, bytes []byte) *msg.MessageEnvelope {
	req := new(msg.FileSavePart)
	req.TotalParts = totalParts
	req.Bytes = bytes
	req.FileID = fileID
	req.PartID = partID

	envelop := new(msg.MessageEnvelope)
	envelop.Constructor = msg.C_FileSavePart
	envelop.Message, _ = req.Marshal()
	envelop.RequestID = uint64(domain.SequentialUniqueID())
	logs.Debug("FilesStatus::generateFileSavePart()",
		zap.Int64("MsgID", us.Request.MessageID),
		zap.Int64("FileID", req.FileID),
		zap.Int32("PartID", req.PartID),
		zap.Int32("TotalParts", req.TotalParts),
		zap.Int("Bytes", len(req.Bytes)),
	)
	return envelop
}

func (us *uploadStatus) execute() {
	us.Request.Status = domain.RequestStatusInProgress
	waitGroup := sync.WaitGroup{}
	for us.Request.MaxRetries > 0 {
		select {
		case partIndex := <-us.parts:
			us.rateLimit <- struct{}{}
			waitGroup.Add(1)

			go func(partIndex int32) {
				defer waitGroup.Done()
				defer func() {
					<-us.rateLimit
				}()

				bytes := pbytes.GetLen(int(us.Request.ChunkSize))
				defer pbytes.Put(bytes)
				offset := partIndex * us.Request.ChunkSize
				_, err := us.file.ReadAt(bytes, int64(offset))
				if err != nil {
					logs.Warn("Error in ReadFile", zap.Error(err))
					atomic.AddInt32(&us.Request.MaxRetries, -1)
					us.parts <- partIndex
					return
				}
				res, err := us.ctrl.network.SendHttp(us.generateFileSavePart(us.Request.FileID, partIndex+1, us.Request.TotalParts, bytes))
				if err != nil {
					logs.Warn("Error in SendHttp", zap.Error(err))
					atomic.AddInt32(&us.Request.MaxRetries, -1)
					us.parts <- partIndex
					return
				}
				switch res.Constructor {
				case msg.C_Bool:
					us.addToUploaded(partIndex)
				default:
					atomic.AddInt32(&us.Request.MaxRetries, -1)
					us.parts <- partIndex
					return
				}
			}(partIndex)
		default:
			waitGroup.Wait()
			switch int32(len(us.Request.UploadedParts)){
			case us.Request.TotalParts -1:
				// If we finished uploading n-1 parts then run the last loop with the last part
				us.parts <- us.Request.TotalParts - 1
			case us.Request.TotalParts:
				// We have finished our uploads
				_ = us.file.Close()
				us.Request.Status = domain.RequestStatusCompleted
				return
			}
		}
	}
}
