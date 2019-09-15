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
	"strings"
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
	PostUploadProcess    func(req UploadRequest)
	OnProgressChanged    func(reqID string, clusterID int32, fileID, accessHash int64, percent int64)
	OnCompleted          func(reqID string, clusterID int32, fileID, accessHash int64, filePath string)
	OnCancel             func(reqID string, clusterID int32, fileID, accessHash int64, hasError bool)
}
type Controller struct {
	network            *networkCtrl.Controller
	mtxDownloads       sync.Mutex
	downloadRequests   map[string]DownloadRequest
	downloadsSaver     *ronak.Flusher
	downloadsRateLimit chan struct{}
	mtxUploads         sync.Mutex
	uploadRequests     map[string]UploadRequest
	uploadsSaver       *ronak.Flusher
	uploadsRateLimit   chan struct{}
	onProgressChanged  func(reqID string, clusterID int32, fileID, accessHash int64, percent int64)
	onCompleted        func(reqID string, clusterID int32, fileID, accessHash int64, filePath string)
	onCancel           func(reqID string, clusterID int32, fileID, accessHash int64, hasError bool)
	postUploadProcess  func(req UploadRequest)
}

func New(config Config) *Controller {
	ctrl := new(Controller)
	ctrl.network = config.Network
	ctrl.downloadsRateLimit = make(chan struct{}, config.MaxInflightDownloads)
	ctrl.uploadsRateLimit = make(chan struct{}, config.MaxInflightUploads)
	ctrl.downloadRequests = make(map[string]DownloadRequest)
	ctrl.uploadRequests = make(map[string]UploadRequest)
	ctrl.postUploadProcess = config.PostUploadProcess
	if config.OnCompleted == nil {
		ctrl.onCompleted = func(reqID string, clusterID int32, fileID, accessHash int64, filePath string) {}
	} else {
		ctrl.onCompleted = config.OnCompleted
	}
	if config.OnProgressChanged == nil {
		ctrl.onProgressChanged = func(reqID string, clusterID int32, fileID, accessHash int64, percent int64) {}
	} else {
		ctrl.onProgressChanged = config.OnProgressChanged
	}
	if config.OnCancel == nil {
		ctrl.onCancel = func(reqID string, clusterID int32, fileID, accessHash int64, hasError bool) {}
	} else {
		ctrl.onCancel = config.OnCancel
	}

	ctrl.downloadsSaver = ronak.NewFlusher(100, 1, time.Millisecond*100, func(items []ronak.FlusherEntry) {
		if dBytes, err := json.Marshal(ctrl.downloadRequests); err == nil {
			_ = repo.System.SaveBytes("Downloads", dBytes)
		}
		for idx := range items {
			items[idx].Callback(nil)
		}
	})
	ctrl.uploadsSaver = ronak.NewFlusher(100, 1, time.Millisecond*100, func(items []ronak.FlusherEntry) {
		if dBytes, err := json.Marshal(ctrl.uploadRequests); err == nil {
			_ = repo.System.SaveBytes("Uploads", dBytes)
		}
		for idx := range items {
			items[idx].Callback(nil)
		}
	})

	return ctrl
}

func (ctrl *Controller) Start() {
	// Resume downloads
	dBytes, err := repo.System.LoadBytes("Downloads")
	if err == nil {
		_ = json.Unmarshal(dBytes, &ctrl.downloadRequests)
		for _, req := range ctrl.downloadRequests {
			go func(req DownloadRequest) {
				_ = ctrl.download(req)
			}(req)
		}
	}

	// Resume uploads
	dBytes, err = repo.System.LoadBytes("Uploads")
	if err == nil {
		_ = json.Unmarshal(dBytes, &ctrl.uploadRequests)
		for _, req := range ctrl.uploadRequests {
			logs.Info("Unfinished Upload")
			go ctrl.upload(req)
		}
	}
}

func (ctrl *Controller) saveDownloads(req DownloadRequest) {
	ctrl.mtxDownloads.Lock()
	ctrl.downloadRequests[req.GetID()] = req
	ctrl.mtxDownloads.Unlock()
	ctrl.downloadsSaver.EnterWithResult(nil, nil)
}
func (ctrl *Controller) deleteDownloadRequest(reqID string) {
	ctrl.mtxDownloads.Lock()
	delete(ctrl.downloadRequests, reqID)
	ctrl.mtxDownloads.Unlock()
	ctrl.downloadsSaver.EnterWithResult(nil, nil)
}
func (ctrl *Controller) existDownloadRequest(reqID string) bool {
	ctrl.mtxDownloads.Lock()
	_, ok := ctrl.downloadRequests[reqID]
	ctrl.mtxDownloads.Unlock()
	return ok
}
func (ctrl *Controller) saveUploads(req UploadRequest) {
	ctrl.mtxUploads.Lock()
	ctrl.uploadRequests[req.GetID()] = req
	ctrl.mtxUploads.Unlock()
	ctrl.uploadsSaver.EnterWithResult(nil, nil)
}
func (ctrl *Controller) deleteUploadRequest(reqID string) {
	ctrl.mtxUploads.Lock()
	delete(ctrl.uploadRequests, reqID)
	ctrl.mtxUploads.Unlock()
	ctrl.uploadsSaver.EnterWithResult(nil, nil)
}
func (ctrl *Controller) existUploadRequest(reqID string) bool {
	ctrl.mtxUploads.Lock()
	_, ok := ctrl.uploadRequests[reqID]
	ctrl.mtxUploads.Unlock()
	return ok
}
func GetRequestID(clusterID int32, fileID int64, accessHash uint64) string {
	return fmt.Sprintf("%d.%d.%d", clusterID, fileID, accessHash)
}
func (ctrl *Controller) GetDownloadRequest(clusterID int32, fileID int64, accessHash uint64) (DownloadRequest, bool) {
	ctrl.mtxDownloads.Lock()
	req, ok := ctrl.downloadRequests[GetRequestID(clusterID, fileID, accessHash)]
	ctrl.mtxDownloads.Unlock()
	return req, ok
}
func (ctrl *Controller) CancelDownloadRequest(reqID string) {
	ctrl.deleteDownloadRequest(reqID)
}
func (ctrl *Controller) GetUploadRequest(fileID int64) (UploadRequest, bool) {
	ctrl.mtxUploads.Lock()
	req, ok := ctrl.uploadRequests[GetRequestID(0, fileID, 0)]
	ctrl.mtxUploads.Unlock()
	return req, ok
}
func (ctrl *Controller) CancelUploadRequest(reqID string) {
	ctrl.deleteUploadRequest(reqID)
}

func (ctrl *Controller) DownloadAsync(clusterID int32, fileID int64, accessHash uint64) (reqID string, err error) {
	clientFile, err := repo.Files.Get(clusterID, fileID, accessHash)
	if err != nil {
		return "", err
	}
	go func() {
		err = ctrl.download(DownloadRequest{
			MessageID:    clientFile.MessageID,
			ClusterID:    clientFile.ClusterID,
			FileID:       clientFile.FileID,
			AccessHash:   clientFile.AccessHash,
			Version:      clientFile.Version,
			FileSize:     clientFile.FileSize,
			ChunkSize:    downloadChunkSize,
			MaxInFlights: maxDownloadInFlights,
			FilePath:     GetFilePath(clientFile),
		})
		logs.WarnOnErr("Error On DownloadAsync", err,
			zap.Int32("ClusterID", clusterID),
			zap.Int64("FileID", fileID),
			zap.Uint64("AccessHash", accessHash),
		)
	}()
	return GetRequestID(clusterID, fileID, accessHash), nil

}
func (ctrl *Controller) DownloadSync(clusterID int32, fileID int64, accessHash uint64) (filePath string, err error) {
	clientFile, err := repo.Files.Get(clusterID, fileID, accessHash)
	if err != nil {
		return "", err
	}
	filePath = GetFilePath(clientFile)
	switch clientFile.Type {
	case msg.ClientFileType_GroupProfilePhoto:
		return ctrl.downloadGroupPhoto(clientFile)
	case msg.ClientFileType_AccountProfilePhoto:
		return ctrl.downloadAccountPhoto(clientFile)
	case msg.ClientFileType_Thumbnail:
		return ctrl.downloadThumbnail(clientFile)
	default:
		err = ctrl.download(DownloadRequest{
			MessageID:    clientFile.MessageID,
			ClusterID:    clientFile.ClusterID,
			FileID:       clientFile.FileID,
			AccessHash:   clientFile.AccessHash,
			Version:      clientFile.Version,
			FileSize:     clientFile.FileSize,
			ChunkSize:    downloadChunkSize,
			MaxInFlights: maxDownloadInFlights,
			FilePath:     filePath,
		})
	}

	return
}
func (ctrl *Controller) downloadAccountPhoto(clientFile *msg.ClientFile) (filePath string, err error) {
	err = ronak.Try(retryMaxAttempts, retryWaitTime, func() error {
		req := new(msg.FileGet)
		req.Location = new(msg.InputFileLocation)
		req.Location.ClusterID = clientFile.ClusterID
		req.Location.FileID = clientFile.FileID
		req.Location.AccessHash = clientFile.AccessHash
		req.Location.Version = clientFile.Version

		// get all bytes
		req.Offset = 0
		req.Limit = 0

		envelop := new(msg.MessageEnvelope)
		envelop.Constructor = msg.C_FileGet
		envelop.Message, _ = req.Marshal()
		envelop.RequestID = uint64(domain.SequentialUniqueID())

		filePath = getAccountProfilePath(clientFile.UserID, req.Location.FileID)
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
			return fmt.Errorf("received error response {userID: %d,  %s }", clientFile.UserID, strErr)
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
			_ = repo.Files.Save(clientFile)
			return nil
		default:
			return fmt.Errorf("received unknown response constructor {UserId : %d}", clientFile.UserID)
		}

	})
	return
}
func (ctrl *Controller) downloadGroupPhoto(clientFile *msg.ClientFile) (filePath string, err error) {
	err = ronak.Try(retryMaxAttempts, retryWaitTime, func() error {
		req := new(msg.FileGet)
		req.Location = new(msg.InputFileLocation)
		req.Location.ClusterID = clientFile.ClusterID
		req.Location.FileID = clientFile.FileID
		req.Location.AccessHash = clientFile.AccessHash
		req.Location.Version = clientFile.Version

		// get all bytes
		req.Offset = 0
		req.Limit = 0

		envelop := new(msg.MessageEnvelope)
		envelop.Constructor = msg.C_FileGet
		envelop.Message, _ = req.Marshal()
		envelop.RequestID = uint64(domain.SequentialUniqueID())

		filePath = getGroupProfilePath(clientFile.GroupID, req.Location.FileID)
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
			return fmt.Errorf("received error response {GroupID: %d,  %s }", clientFile.GroupID, strErr)
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
			_ = repo.Files.Save(clientFile)
			return nil

		default:
			return fmt.Errorf("received unknown response constructor {GroupID : %d}", clientFile.GroupID)
		}

	})
	return
}
func (ctrl *Controller) downloadThumbnail(clientFile *msg.ClientFile) (filePath string, err error) {
	err = ronak.Try(10, 100*time.Millisecond, func() error {
		req := new(msg.FileGet)
		req.Location = &msg.InputFileLocation{
			AccessHash: clientFile.AccessHash,
			ClusterID:  clientFile.ClusterID,
			FileID:     clientFile.FileID,
			Version:    clientFile.Version,
		}
		// get all bytes
		req.Offset = 0
		req.Limit = 0

		envelop := new(msg.MessageEnvelope)
		envelop.Constructor = msg.C_FileGet
		envelop.Message, _ = req.Marshal()
		envelop.RequestID = uint64(domain.SequentialUniqueID())

		filePath = getThumbnailPath(clientFile.FileID, clientFile.ClusterID)
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
			_ = repo.Files.Save(clientFile)

			return nil

		default:
			return nil
		}

	})
	return
}
func (ctrl *Controller) download(req DownloadRequest) error {
	req.TempFilePath = fmt.Sprintf("%s.tmp", req.FilePath)
	ctrl.saveDownloads(req)
	ctrl.downloadsRateLimit <- struct{}{}
	defer func() {
		<-ctrl.downloadsRateLimit
	}()

	ds := &downloadContext{
		rateLimit: make(chan struct{}, req.MaxInFlights),
		ctrl:      ctrl,
	}

	_, err := os.Stat(req.TempFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			ds.file, err = os.Create(req.TempFilePath)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		ds.file, err = os.OpenFile(req.TempFilePath, os.O_RDWR, 0666)
		if err != nil {
			return err
		}
	}

	if req.FileSize > 0 {
		err := os.Truncate(req.TempFilePath, req.FileSize)
		if err != nil {
			ctrl.onCancel(req.GetID(), req.ClusterID, req.FileID, int64(req.AccessHash), true)
			return err
		}
		dividend := int32(req.FileSize / int64(req.ChunkSize))
		if req.FileSize%int64(req.ChunkSize) > 0 {
			req.TotalParts = dividend + 1
		} else {
			req.TotalParts = dividend
		}
	} else {
		req.TotalParts = 1
		req.ChunkSize = 0
	}

	if req.MaxRetries <= 0 {
		req.MaxRetries = retryMaxAttempts
	}

	ds.req = req
	ds.parts = make(chan int32, ds.req.TotalParts)
	for partIndex := int32(0); partIndex < ds.req.TotalParts; partIndex++ {
		if ds.isDownloaded(partIndex) {
			continue
		}
		ds.parts <- partIndex
	}

	// This is blocking call, until all the parts are downloaded
	ds.execute()

	// Remove the Download request from the list
	ctrl.deleteDownloadRequest(req.GetID())
	return nil
}

func (ctrl *Controller) UploadUserPhoto(filePath string) (reqID string) {
	// support IOS file path
	if strings.HasPrefix(filePath, "file://") {
		filePath = filePath[7:]
	}

	fileID := ronak.RandomInt64(0)
	ctrl.upload(UploadRequest{
		IsProfilePhoto: true,
		FileID:         fileID,
		MaxInFlights:   3,
		FilePath:       filePath,
	})
	reqID = GetRequestID(0, fileID, 0)
	return
}
func (ctrl *Controller) UploadGroupPhoto(groupID int64, filePath string) (reqID string) {
	// support IOS file path
	if strings.HasPrefix(filePath, "file://") {
		filePath = filePath[7:]
	}

	fileID := ronak.RandomInt64(0)
	ctrl.upload(UploadRequest{
		IsProfilePhoto: true,
		GroupID:        groupID,
		FileID:         fileID,
		MaxInFlights:   3,
		FilePath:       filePath,
	})

	reqID = GetRequestID(0, fileID, 0)
	return
}
func (ctrl *Controller) UploadMessageDocument(messageID int64, filePath, thumbPath string, fileID, thumbID int64) {
	// support IOS file path
	if strings.HasPrefix(filePath, "file://") {
		filePath = filePath[7:]

	}
	// support IOS file path
	if strings.HasPrefix(thumbPath, "file://") {
		thumbPath = thumbPath[7:]
	}

	// We prepare upload request for the actual file before uploading the thumbnail to save it
	// in case of execution stopped, then we are assured that we will continue the upload process
	req := UploadRequest{
		MessageID:    messageID,
		FileID:       fileID,
		FilePath:     filePath,
		ThumbID:      thumbID,
		ThumbPath:    thumbPath,
		MaxInFlights: 3,
	}
	ctrl.saveUploads(req)

	// Upload Thumbnail
	ctrl.upload(UploadRequest{
		MessageID:    0,
		FileID:       thumbID,
		MaxInFlights: 3,
		FilePath:     thumbPath,
	})

	// Upload File
	ctrl.upload(req)
}
func (ctrl *Controller) upload(req UploadRequest) {
	ctrl.saveUploads(req)
	ctrl.uploadsRateLimit <- struct{}{}
	defer func() {
		<-ctrl.uploadsRateLimit
	}()

	ds := &uploadContext{
		rateLimit: make(chan struct{}, req.MaxInFlights),
		ctrl:      ctrl,
	}

	fileInfo, err := os.Stat(req.FilePath)
	if err != nil {
		ctrl.onCancel(req.GetID(), 0, req.FileID, 0, true)
		return
	}
	ds.file, err = os.OpenFile(req.FilePath, os.O_RDONLY, 0666)
	if err != nil {
		ctrl.onCancel(req.GetID(), 0, req.FileID, 0, true)
		return
	}

	req.FileSize = fileInfo.Size()
	if req.FileSize <= 0 {
		ctrl.onCancel(req.GetID(), 0, req.FileID, 0, true)
		return
	}

	if req.FileSize > domain.FileMaxAllowedSize {
		ctrl.onCancel(req.GetID(), 0, req.FileID, 0, true)
		return
	}

	if req.ChunkSize == 0 {
		req.ChunkSize = downloadChunkSize
	}

	if req.MaxRetries <= 0 {
		req.MaxRetries = retryMaxAttempts
	}

	dividend := int32(req.FileSize / int64(req.ChunkSize))
	if req.FileSize%int64(req.ChunkSize) > 0 {
		req.TotalParts = dividend + 1
	} else {
		req.TotalParts = dividend
	}

	ds.parts = make(chan int32, req.TotalParts+req.MaxInFlights)
	ds.req = req
	for partIndex := int32(0); partIndex < req.TotalParts-1; partIndex++ {
		if ds.isUploaded(partIndex) {
			continue
		}
		ds.parts <- partIndex
	}

	// This is blocking call, until all the parts are downloaded
	ds.execute()

	// Remove the Download request from the list
	ctrl.deleteUploadRequest(req.GetID())
	return
}
