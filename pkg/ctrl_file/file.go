package fileCtrl

import (
	"encoding/json"
	"fmt"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	networkCtrl "git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_network"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"go.uber.org/zap"
	"io/ioutil"
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

type Config struct {
	Network              *networkCtrl.Controller
	MaxInflightDownloads int32
	MaxInflightUploads   int32
}
type Controller struct {
	network           *networkCtrl.Controller
	rateLimitDownload chan struct{}
	rateLimitUpload   chan struct{}
	mtxDownloads      sync.Mutex
	downloadRequests  map[int64]DownloadRequest
	saveSnapshot      *ronak.Flusher
}

func New(config Config) *Controller {
	ctrl := new(Controller)
	ctrl.network = config.Network
	ctrl.rateLimitDownload = make(chan struct{}, config.MaxInflightDownloads)
	ctrl.rateLimitUpload = make(chan struct{}, config.MaxInflightUploads)

	dBytes, err := repo.System.LoadBytes("Downloads")
	if err != nil {
		ctrl.downloadRequests = make(map[int64]DownloadRequest)
	} else {
		_ = json.Unmarshal(dBytes, &ctrl.downloadRequests)
		for _, req := range ctrl.downloadRequests {
			_ = ctrl.startDownload(req)
		}
	}

	ctrl.saveSnapshot = ronak.NewFlusher(100, 1, time.Millisecond*100, func(items []ronak.FlusherEntry) {
		if dBytes, err := json.Marshal(ctrl.downloadRequests); err == nil {
			_ = repo.System.SaveBytes("Downloads", dBytes)
		}
		for idx := range items {
			items[idx].Callback(nil)
		}
	})
	return ctrl
}

func (ctrl *Controller) startDownload(req DownloadRequest) (err error) {
	ctrl.rateLimitDownload <- struct{}{}
	defer func() {
		<-ctrl.rateLimitDownload
	}()

	ds := &downloadStatus{
		rateLimit: make(chan struct{}, req.MaxInFlights),
		ctrl:      ctrl,
		Request:   req,
	}

	_, err = os.Stat(req.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			ds.file, err = os.Create(req.FilePath)
			if err != nil {
				return err
			}
			if req.FileSize > 0 {
				err := os.Truncate(req.FilePath, req.FileSize)
				if err != nil {
					return err
				}
			}
		} else {
			return err
		}
	}

	if req.FileSize > 0 {
		dividend := int32(req.FileSize / int64(req.ChunkSize))
		if req.FileSize%int64(req.ChunkSize) > 0 {
			ds.Request.TotalParts = dividend + 1
		} else {
			ds.Request.TotalParts = dividend
		}
	} else {
		ds.Request.TotalParts = 1
		ds.Request.ChunkSize = 0
	}

	ds.parts = make(chan int32, ds.Request.TotalParts)
	for partIndex := int32(0); partIndex < ds.Request.TotalParts; partIndex++ {
		if ds.isDownloaded(partIndex) {
			continue
		}
		ds.parts <- partIndex
	}

	// This is blocking call, until all the parts are downloaded
	ds.run()

	return nil
}

func (ctrl *Controller) GetDownloadRequest(messageID int64) (DownloadRequest, bool) {
	ctrl.mtxDownloads.Lock()
	req, ok := ctrl.downloadRequests[messageID]
	ctrl.mtxDownloads.Unlock()
	return req, ok
}

// Download add download request
func (ctrl *Controller) Download(userMessage *msg.UserMessage) {
	switch userMessage.MediaType {
	case msg.MediaTypeEmpty:
	case msg.MediaTypeDocument:
		x := new(msg.MediaDocument)
		err := x.Unmarshal(userMessage.Media)
		if err != nil {
			logs.Error("Error In Download", zap.Error(err))
			return
		}
		err = ctrl.startDownload(DownloadRequest{
			MaxRetries:   10,
			MessageID:    userMessage.ID,
			ClusterID:    x.Doc.ClusterID,
			FileID:       x.Doc.ID,
			AccessHash:   x.Doc.AccessHash,
			Version:      x.Doc.Version,
			FileSize:     int64(x.Doc.FileSize),
			ChunkSize:    downloadChunkSize,
			MaxInFlights: maxDownloadInFlights,
			FilePath:     GetFilePath(x.Doc.MimeType, x.Doc.ID),
		})
		if err != nil {
			logs.Error("Error In Download", zap.Error(err))
		}
	default:
		return
	}
}

// func (ctrl *Controller) UploadProfilePhoto() {}
//
// // Upload file to server
// func (ctrl *Controller) Upload(fileID int64, req *msg.ClientPendingMessage) error {
// 	x := new(msg.ClientSendMessageMedia)
// 	err := x.Unmarshal(req.Media)
// 	if err != nil {
// 		return err
// 	}
// 	file, err := os.Open(x.FilePath)
// 	if err != nil {
// 		return err
// 	}
// 	fileInfo, err := file.Stat()
// 	if err != nil {
// 		return err
// 	}
// 	if x.FileName == "" {
// 		x.FileName = fileInfo.Name()
// 	}
// 	fileSize := fileInfo.Size() // size in Byte
// 	if fileSize > domain.FileMaxAllowedSize {
// 		return domain.ErrMaxFileSize
// 	}
//
//
// 	return nil
// }

// DownloadAccountPhoto download account photo from server its sync
func (ctrl *Controller) DownloadAccountPhoto(userID int64, photo *msg.UserPhoto, isBig bool) (string, error) {
	var filePath string
	err := ronak.Try(retryMaxAttempts, retryWaitTime, func() error {
		req := new(msg.FileGet)
		req.Location = new(msg.InputFileLocation)
		if isBig {
			req.Location.ClusterID = photo.PhotoBig.ClusterID
			req.Location.AccessHash = photo.PhotoBig.AccessHash
			req.Location.FileID = photo.PhotoBig.FileID
			req.Location.Version = 0
		} else {
			req.Location.ClusterID = photo.PhotoSmall.ClusterID
			req.Location.AccessHash = photo.PhotoSmall.AccessHash
			req.Location.FileID = photo.PhotoSmall.FileID
			req.Location.Version = 0
		}
		// get all bytes
		req.Offset = 0
		req.Limit = 0

		envelop := new(msg.MessageEnvelope)
		envelop.Constructor = msg.C_FileGet
		envelop.Message, _ = req.Marshal()
		envelop.RequestID = uint64(domain.SequentialUniqueID())

		filePath = GetAccountAvatarPath(userID, req.Location.FileID)
		res, err := ctrl.network.SendHttp(envelop)
		if err != nil {
			return err
		}

		switch res.Constructor {
		case msg.C_Error:
			strErr := ""
			x := new(msg.Error)
			if err := x.Unmarshal(res.Message); err == nil {
				strErr = "Code :" + x.Code + ", Items :" + x.Items
			}
			return fmt.Errorf("received error response {userID: %d,  %s }", userID, strErr)
		case msg.C_File:
			x := new(msg.File)
			err := x.Unmarshal(res.Message)
			if err != nil {
				return err
			}

			// write to file path
			err = ioutil.WriteFile(filePath, x.Bytes, 0666)
			if err != nil {
				return err
			}

			// save to DB
			return nil
		default:
			return fmt.Errorf("received unknown response constructor {UserId : %d}", userID)
		}

	})
	if err != nil {
		return "", err
	}

	return filePath, nil
}

// DownloadGroupPhoto download group photo from server its sync
func (ctrl *Controller) DownloadGroupPhoto(groupID int64, photo *msg.GroupPhoto, isBig bool) (string, error) {
	var filePath string
	err := ronak.Try(retryMaxAttempts, retryWaitTime, func() error {
		req := new(msg.FileGet)
		req.Location = new(msg.InputFileLocation)
		if isBig {
			req.Location.ClusterID = photo.PhotoBig.ClusterID
			req.Location.AccessHash = photo.PhotoBig.AccessHash
			req.Location.FileID = photo.PhotoBig.FileID
			req.Location.Version = 0
		} else {
			req.Location.ClusterID = photo.PhotoSmall.ClusterID
			req.Location.AccessHash = photo.PhotoSmall.AccessHash
			req.Location.FileID = photo.PhotoSmall.FileID
			req.Location.Version = 0
		}
		// get all bytes
		req.Offset = 0
		req.Limit = 0

		envelop := new(msg.MessageEnvelope)
		envelop.Constructor = msg.C_FileGet
		envelop.Message, _ = req.Marshal()
		envelop.RequestID = uint64(domain.SequentialUniqueID())

		filePath = GetGroupAvatarPath(groupID, req.Location.FileID)
		res, err := ctrl.network.SendHttp(envelop)
		if err != nil {
			return err
		}
		switch res.Constructor {
		case msg.C_Error:
			strErr := ""
			x := new(msg.Error)
			if err := x.Unmarshal(res.Message); err == nil {
				strErr = "Code :" + x.Code + ", Items :" + x.Items
			}
			return fmt.Errorf("received error response {GroupID: %d,  %s }", groupID, strErr)
		case msg.C_File:
			x := new(msg.File)
			err := x.Unmarshal(res.Message)
			if err != nil {
				return err
			}

			// write to file path
			err = ioutil.WriteFile(filePath, x.Bytes, 0666)
			if err != nil {
				return err
			}

			// save to DB
			return nil

		default:
			return fmt.Errorf("received unknown response constructor {GroupID : %d}", groupID)
		}

	})
	if err != nil {
		return "", err
	}
	return filePath, nil
}

// DownloadThumbnail download thumbnail from server its sync
func (ctrl *Controller) DownloadThumbnail(fileID int64, accessHash uint64, clusterID, version int32) (string, error) {
	filePath := ""
	err := ronak.Try(10, 100*time.Millisecond, func() error {
		req := new(msg.FileGet)
		req.Location = &msg.InputFileLocation{
			AccessHash: accessHash,
			ClusterID:  clusterID,
			FileID:     fileID,
			Version:    version,
		}
		// get all bytes
		req.Offset = 0
		req.Limit = 0

		envelop := new(msg.MessageEnvelope)
		envelop.Constructor = msg.C_FileGet
		envelop.Message, _ = req.Marshal()
		envelop.RequestID = uint64(domain.SequentialUniqueID())

		filePath = GetThumbnailPath(fileID, clusterID)
		res, err := ctrl.network.SendHttp(envelop)
		if err != nil {
			return err
		}

		switch res.Constructor {
		case msg.C_Error:
			strErr := ""
			x := new(msg.Error)
			if err := x.Unmarshal(res.Message); err == nil {
				strErr = "Code :" + x.Code + ", Items :" + x.Items
			}
			return fmt.Errorf("received error response {%s}", strErr)
		case msg.C_File:
			x := new(msg.File)
			err := x.Unmarshal(res.Message)
			if err != nil {
				return err
			}

			// write to file path
			err = ioutil.WriteFile(filePath, x.Bytes, 0666)
			if err != nil {
				return err
			}

			// save to DB
			return nil

		default:
			return nil
		}

	})
	if err != nil {
		return "", err
	}
	return filePath, nil
}
