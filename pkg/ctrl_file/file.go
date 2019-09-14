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
	OnProgressChanged    func(reqID string, percent int64)
	OnCompleted          func(reqID string, filePath string)
	OnError              func(reqID string, filePath string, err []byte)
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
	onProgressChanged  func(reqID string, percent int64)
	onCompleted        func(reqID string, filePath string)
	onError            func(reqID string, filePath string, err []byte)
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
		ctrl.onCompleted = func(reqID string, filePath string) {}
	} else {
		ctrl.onCompleted = config.OnCompleted
	}
	if config.OnProgressChanged == nil {
		ctrl.onProgressChanged = func(reqID string, percent int64) {}
	} else {
		ctrl.onProgressChanged = config.OnProgressChanged
	}
	if config.OnError == nil {
		ctrl.onError = func(reqID string, filePath string, err []byte) {}
	} else {
		ctrl.onError = config.OnError
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
				_ = ctrl.Download(req)
			}(req)
		}
	}

	// Resume uploads
	dBytes, err = repo.System.LoadBytes("Uploads")
	if err == nil {
		_ = json.Unmarshal(dBytes, &ctrl.uploadRequests)
		for _, req := range ctrl.uploadRequests {
			logs.Info("Unfinished Upload")
			go ctrl.Upload(req)
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
func (ctrl *Controller) saveUploads(req UploadRequest) {
	ctrl.mtxUploads.Lock()
	ctrl.uploadRequests[req.GetID()] = req
	ctrl.mtxUploads.Unlock()
	ctrl.uploadsSaver.EnterWithResult(nil, nil)
}
func (ctrl *Controller) deleteUpdateRequest(reqID string) {
	ctrl.mtxUploads.Lock()
	delete(ctrl.uploadRequests, reqID)
	ctrl.mtxUploads.Unlock()
	ctrl.uploadsSaver.EnterWithResult(nil, nil)
}

func GetDownloadRequestID(clusterID int32, fileID int64, accessHash uint64) string {
	return fmt.Sprintf("%d.%d.%d", clusterID, fileID, accessHash)
}
func (ctrl *Controller) GetDownloadRequest(clusterID int32, fileID int64, accessHash uint64) (DownloadRequest, bool) {
	ctrl.mtxDownloads.Lock()
	req, ok := ctrl.downloadRequests[GetDownloadRequestID(clusterID, fileID, accessHash)]
	ctrl.mtxDownloads.Unlock()
	return req, ok
}
func GetUploadRequestID(fileID int64) string {
	return fmt.Sprintf("%d", fileID)
}
func (ctrl *Controller) GetUploadRequest(fileID int64) (UploadRequest, bool) {
	ctrl.mtxUploads.Lock()
	req, ok := ctrl.uploadRequests[GetUploadRequestID(fileID)]
	ctrl.mtxUploads.Unlock()
	return req, ok
}

// DownloadAccountPhoto downloads the account profile photo (Big/Small) and returns the file path of the stored
// image.
func (ctrl *Controller) DownloadAccountPhoto(userID int64, photo *msg.UserPhoto, isBig bool) (filePath string, err error) {
	err = ronak.Try(retryMaxAttempts, retryWaitTime, func() error {
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
	return
}

// DownloadGroupPhoto downloads the profile photo of the group 'groupID' and returns the file path which
// the file has been stored.
func (ctrl *Controller) DownloadGroupPhoto(groupID int64, photo *msg.GroupPhoto, isBig bool) (filePath string, err error) {
	err = ronak.Try(retryMaxAttempts, retryWaitTime, func() error {
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
	return
}

// DownloadThumbnail downloads the thumbnail file
func (ctrl *Controller) DownloadThumbnail(fileID int64, accessHash uint64, clusterID, version int32) (filePath string, err error) {
	err = ronak.Try(10, 100*time.Millisecond, func() error {
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
	return
}

// DownloadByMessage downloads the attachment of the message, if any.
func (ctrl *Controller) DownloadByMessage(userMessage *msg.UserMessage) (filePath string, err error) {
	switch userMessage.MediaType {
	case msg.MediaTypeEmpty:
	case msg.MediaTypeDocument:
		x := new(msg.MediaDocument)
		err = x.Unmarshal(userMessage.Media)
		if err != nil {
			return
		}

		filePath = GetFilePath(x.Doc.MimeType, x.Doc.ID)
		err = ctrl.Download(DownloadRequest{
			MessageID:    userMessage.ID,
			ClusterID:    x.Doc.ClusterID,
			FileID:       x.Doc.ID,
			AccessHash:   x.Doc.AccessHash,
			Version:      x.Doc.Version,
			FileSize:     int64(x.Doc.FileSize),
			ChunkSize:    downloadChunkSize,
			MaxInFlights: maxDownloadInFlights,
			FilePath:     filePath,
		})
	}
	return
}
func (ctrl *Controller) Download(req DownloadRequest) error {
	req.TempFilePath = fmt.Sprintf("%s.tmp", req.FilePath)
	ctrl.saveDownloads(req)
	ctrl.downloadsRateLimit <- struct{}{}
	defer func() {
		<-ctrl.downloadsRateLimit
	}()

	ds := &downloadContext{
		rateLimit: make(chan struct{}, req.MaxInFlights),
		ctrl:      ctrl,
		req:       req,
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
			ctrl.onError(req.GetID(), req.TempFilePath, ronak.StrToByte(err.Error()))
			return err
		}
		dividend := int32(req.FileSize / int64(req.ChunkSize))
		if req.FileSize%int64(req.ChunkSize) > 0 {
			ds.req.TotalParts = dividend + 1
		} else {
			ds.req.TotalParts = dividend
		}
	} else {
		ds.req.TotalParts = 1
		ds.req.ChunkSize = 0
	}

	if req.MaxRetries <= 0 {
		req.MaxRetries = retryMaxAttempts
	}

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
	ctrl.Upload(UploadRequest{
		IsProfilePhoto: true,
		FileID:         fileID,
		MaxInFlights:   3,
		FilePath:       filePath,
	})
	reqID = GetUploadRequestID(fileID)
	return
}
func (ctrl *Controller) UploadGroupPhoto(groupID int64, filePath string) (reqID string) {
	// support IOS file path
	if strings.HasPrefix(filePath, "file://") {
		filePath = filePath[7:]
	}

	fileID := ronak.RandomInt64(0)
	ctrl.Upload(UploadRequest{
		IsProfilePhoto: true,
		GroupID:        groupID,
		FileID:         fileID,
		MaxInFlights:   3,
		FilePath:       filePath,
	})

	reqID = GetUploadRequestID(fileID)
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
	ctrl.Upload(UploadRequest{
		MessageID:    0,
		FileID:       thumbID,
		MaxInFlights: 3,
		FilePath:     thumbPath,
	})

	// Upload File
	ctrl.Upload(req)
}
func (ctrl *Controller) Upload(req UploadRequest) {
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
		ctrl.onError(req.GetID(), req.FilePath, ronak.StrToByte(err.Error()))
		return
	}
	ds.file, err = os.OpenFile(req.FilePath, os.O_RDONLY, 0666)
	if err != nil {
		ctrl.onError(req.GetID(), req.FilePath, ronak.StrToByte(err.Error()))
		return
	}

	req.FileSize = fileInfo.Size()
	if req.FileSize <= 0 {
		ctrl.onError(req.GetID(), req.FilePath, ronak.StrToByte("file size is not positive"))
		return
	}

	if req.FileSize > domain.FileMaxAllowedSize {
		ctrl.onError(req.GetID(), req.FilePath, ronak.StrToByte("file size is bigger than maximum allowed"))
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
	ctrl.deleteUpdateRequest(req.GetID())
	return
}
