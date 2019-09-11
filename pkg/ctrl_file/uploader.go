package fileCtrl

import (
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/gobwas/pool/pbytes"
	"go.uber.org/zap"
	"io"
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
	// IsProfilePhoto indicates that the uploaded file will be used as a profile photo for Group or User
	IsProfilePhoto bool `json:"is_profile_photo"`
	// PeerID will be set if IsProfilePhoto has been set to TRUE and user is going to upload group photo
	GroupID int64 `json:"group_id"`
	// MaxRetries defines how many time each request could encounter error before giving up
	MaxRetries int32 `json:"max_retries"`
	// MessageID (Optional) if is set then (ClusterID, FileID, AccessHash, Version) will be read from the message
	// document object, or if message has no document then return error
	MessageID int64 `json:"message_id"`
	// FilePath defines the path which downloaded file will be stored. It must be a file not a directory.
	// Also it will be overwritten if Overwrite is TRUE
	FilePath string `json:"file_path"`
	// FileID is a random id which must be unique for the client, it will be used to reference the uploaded document
	// in the server
	FileID int64 `json:"file_id"`
	// FileSize (Optional) if is set then progress will be calculated
	FileSize int64 `json:"file_size"`
	// ThumbPath (Optional) defines the path of thumbnail of the file will be stored. It must be a file not a directory.
	// Also it will be overwritten if Overwrite is TRUE
	ThumbPath string `json:"thumb_path"`
	// ThumbID is a random id which must be unique for the client, it will be used to reference the uploaded document
	// thumbnail in the server
	ThumbID int64
	// ChunkSize identifies how many request we need to send to server to Download a file.
	ChunkSize int32 `json:"chunk_size"`
	// MaxInFlights defines that how many requests could be send concurrently
	MaxInFlights    int32     `json:"max_in_flights"`
	UploadedParts   []int32 `json:"downloaded_parts"`
	TotalParts      int32   `json:"total_parts"`
}

type uploadStatus struct {
	mtx       sync.Mutex
	rateLimit chan struct{}
	parts     chan int32
	file      *os.File
	ctrl      *Controller
	req       UploadRequest
}

func (us *uploadStatus) isUploaded(partIndex int32) bool {
	us.mtx.Lock()
	defer us.mtx.Unlock()
	for _, index := range us.req.UploadedParts {
		if partIndex == index {
			return true
		}
	}
	return false
}

func (us *uploadStatus) addToUploaded(partIndex int32) {
	us.mtx.Lock()
	us.req.UploadedParts = append(us.req.UploadedParts, partIndex)
	progress := int64(float64(len(us.req.UploadedParts)) / float64(us.req.TotalParts) * 100)
	us.mtx.Unlock()
	us.ctrl.saveUploads(us.req)
	us.ctrl.onProgressChanged(us.req.MessageID, progress)
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
		zap.Int64("MsgID", us.req.MessageID),
		zap.Int64("FileID", req.FileID),
		zap.Int32("PartID", req.PartID),
		zap.Int32("TotalParts", req.TotalParts),
		zap.Int("Bytes", len(req.Bytes)),
	)
	return envelop
}

func (us *uploadStatus) execute() domain.RequestStatus {
	waitGroup := sync.WaitGroup{}
	for us.req.MaxRetries > 0 {
		select {
		case partIndex := <-us.parts:
			us.rateLimit <- struct{}{}
			waitGroup.Add(1)

			go func(partIndex int32) {
				defer waitGroup.Done()
				defer func() {
					<-us.rateLimit
				}()

				bytes := pbytes.GetLen(int(us.req.ChunkSize))
				defer pbytes.Put(bytes)
				offset := partIndex * us.req.ChunkSize
				_, err := us.file.ReadAt(bytes, int64(offset))
				if err != nil && err != io.EOF{
					logs.Warn("Error in ReadFile", zap.Error(err))
					atomic.AddInt32(&us.req.MaxRetries, -1)
					us.parts <- partIndex
					return
				}
				res, err := us.ctrl.network.SendHttp(us.generateFileSavePart(us.req.FileID, partIndex+1, us.req.TotalParts, bytes))
				if err != nil {
					logs.Warn("Error in SendHttp", zap.Error(err))
					atomic.AddInt32(&us.req.MaxRetries, -1)
					us.parts <- partIndex
					return
				}
				switch res.Constructor {
				case msg.C_Bool:
					us.addToUploaded(partIndex)
				default:
					atomic.AddInt32(&us.req.MaxRetries, -1)
					us.parts <- partIndex
				}
			}(partIndex)
		default:
			waitGroup.Wait()
			switch int32(len(us.req.UploadedParts)) {
			case us.req.TotalParts - 1:
				// If we finished uploading n-1 parts then run the last loop with the last part
				us.parts <- us.req.TotalParts - 1
			case us.req.TotalParts:
				// We have finished our uploads
				_ = us.file.Close()
				us.ctrl.onCompleted(us.req.MessageID, us.req.FilePath)
				if us.ctrl.postUploadProcess != nil {
					us.ctrl.postUploadProcess(us.req)
				}
				return domain.RequestStatusCompleted
			}
		}
	}
	us.ctrl.onError(us.req.MessageID, us.req.FilePath, ronak.StrToByte("max retry exceeded without success"))
	return domain.RequestStatusError
}
