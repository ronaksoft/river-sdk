package fileCtrl

import (
	"context"
	"encoding/json"
	"fmt"
	"git.ronaksoft.com/river/msg/msg"
	"git.ronaksoft.com/ronak/riversdk/internal/logs"
	networkCtrl "git.ronaksoft.com/ronak/riversdk/pkg/ctrl_network"
	"git.ronaksoft.com/ronak/riversdk/pkg/domain"
	"git.ronaksoft.com/ronak/riversdk/pkg/repo"
	"github.com/dgraph-io/badger/v2"
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
	HttpRequestTimeout   time.Duration
	PostUploadProcessCB  func(req UploadRequest) bool
	ProgressChangedCB    func(reqID string, clusterID int32, fileID, accessHash int64, percent int64, peerID int64)
	CompletedCB          func(reqID string, clusterID int32, fileID, accessHash int64, filePath string, peerID int64)
	CancelCB             func(reqID string, clusterID int32, fileID, accessHash int64, hasError bool, peerID int64)
}

type Controller struct {
	network            *networkCtrl.Controller
	mtxDownloads       sync.Mutex
	downloadRequests   map[string]DownloadRequest
	downloadsSaver     *domain.Flusher
	downloadsRateLimit chan struct{}
	mtxUploads         sync.Mutex
	uploadRequests     map[string]UploadRequest
	uploadsSaver       *domain.Flusher
	uploadsRateLimit   chan struct{}
	httpRequestTimeout time.Duration

	// Callbacks
	onProgressChanged func(reqID string, clusterID int32, fileID, accessHash int64, percent int64, peerID int64)
	onCompleted       func(reqID string, clusterID int32, fileID, accessHash int64, filePath string, peerID int64)
	onCancel          func(reqID string, clusterID int32, fileID, accessHash int64, hasError bool, peerID int64)
	postUploadProcess func(req UploadRequest) bool
}

func New(config Config) *Controller {
	ctrl := &Controller{
		network:            config.Network,
		downloadsRateLimit: make(chan struct{}, config.MaxInflightDownloads),
		uploadsRateLimit:   make(chan struct{}, config.MaxInflightUploads),
		downloadRequests:   make(map[string]DownloadRequest),
		uploadRequests:     make(map[string]UploadRequest),
		postUploadProcess:  config.PostUploadProcessCB,
	}
	ctrl.httpRequestTimeout = domain.HttpRequestTime
	if config.HttpRequestTimeout > 0 {
		ctrl.httpRequestTimeout = config.HttpRequestTimeout
	}
	if config.CompletedCB == nil {
		ctrl.onCompleted = func(reqID string, clusterID int32, fileID, accessHash int64, filePath string, peerID int64) {}
	} else {
		ctrl.onCompleted = config.CompletedCB
	}
	if config.ProgressChangedCB == nil {
		ctrl.onProgressChanged = func(reqID string, clusterID int32, fileID, accessHash int64, percent int64, peerID int64) {}
	} else {
		ctrl.onProgressChanged = config.ProgressChangedCB
	}
	if config.CancelCB == nil {
		ctrl.onCancel = func(reqID string, clusterID int32, fileID, accessHash int64, hasError bool, peerID int64) {}
	} else {
		ctrl.onCancel = config.CancelCB
	}

	ctrl.downloadsSaver = domain.NewFlusher(100, 1, time.Millisecond*100, func(items []domain.FlusherEntry) {
		ctrl.mtxDownloads.Lock()
		if dBytes, err := json.Marshal(ctrl.downloadRequests); err == nil {
			_ = repo.System.SaveBytes("Downloads", dBytes)
		}
		ctrl.mtxDownloads.Unlock()
		for idx := range items {
			items[idx].Callback(nil)
		}
	})
	ctrl.uploadsSaver = domain.NewFlusher(100, 1, time.Millisecond*100, func(items []domain.FlusherEntry) {
		ctrl.mtxUploads.Lock()
		if dBytes, err := json.Marshal(ctrl.uploadRequests); err == nil {
			_ = repo.System.SaveBytes("Uploads", dBytes)
		}
		ctrl.mtxUploads.Unlock()
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
		downloadRequests := make(map[string]DownloadRequest)
		_ = json.Unmarshal(dBytes, &downloadRequests)
		for _, req := range downloadRequests {
			go func(req DownloadRequest) {
				err := ctrl.download(req)
				logs.WarnOnErr("Error On Download Start", err)
			}(req)
		}
	}

	// Resume uploads
	dBytes, err = repo.System.LoadBytes("Uploads")
	if err == nil {
		uploadRequests := make(map[string]UploadRequest)
		_ = json.Unmarshal(dBytes, &uploadRequests)
		for _, req := range uploadRequests {
			go func(req UploadRequest) {
				err := ctrl.upload(req)
				logs.WarnOnErr("Error On Upload Start", err)
			}(req)
		}
	}
}

// download helper functions
func (ctrl *Controller) saveDownloads(req DownloadRequest) {
	ctrl.mtxDownloads.Lock()
	ctrl.downloadRequests[req.GetID()] = req
	ctrl.mtxDownloads.Unlock()
	ctrl.downloadsSaver.EnterWithResult(nil, nil)
}
func (ctrl *Controller) getDownloadRequest(reqID string) (DownloadRequest, bool) {
	ctrl.mtxDownloads.Lock()
	req, ok := ctrl.downloadRequests[reqID]
	ctrl.mtxDownloads.Unlock()
	return req, ok
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

// upload helper functions
func (ctrl *Controller) saveUploads(req UploadRequest) {
	ctrl.mtxUploads.Lock()
	ctrl.uploadRequests[req.GetID()] = req
	ctrl.mtxUploads.Unlock()
	ctrl.uploadsSaver.EnterWithResult(nil, nil)
}
func (ctrl *Controller) getUploadRequest(reqID string) (UploadRequest, bool) {
	ctrl.mtxUploads.Lock()
	req, ok := ctrl.uploadRequests[reqID]
	ctrl.mtxUploads.Unlock()
	return req, ok
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

func (ctrl *Controller) GetDownloadRequest(clusterID int32, fileID int64, accessHash uint64) (DownloadRequest, bool) {
	return ctrl.getDownloadRequest(getRequestID(clusterID, fileID, accessHash))
}
func (ctrl *Controller) CancelDownloadRequest(reqID string) {
	req, ok := ctrl.getDownloadRequest(reqID)
	ctrl.deleteDownloadRequest(reqID)
	if ok && req.cancelFunc != nil {
		req.cancelFunc()
	}
}
func (ctrl *Controller) Stop() {
	ctrl.mtxDownloads.Lock()
	for reqID := range ctrl.downloadRequests {
		req, ok := ctrl.downloadRequests[reqID]
		if ok && req.cancelFunc != nil {
			req.cancelFunc()
		}
		delete(ctrl.downloadRequests, reqID)
	}
	ctrl.mtxDownloads.Unlock()
	ctrl.downloadsSaver.EnterWithResult(nil, nil)

	// delete and cancel all
	ctrl.mtxUploads.Lock()
	for reqID := range ctrl.uploadRequests {
		req, ok := ctrl.uploadRequests[reqID]
		if ok && req.cancelFunc != nil {
			req.cancelFunc()
		}
		delete(ctrl.uploadRequests, reqID)
	}
	ctrl.mtxUploads.Unlock()
	ctrl.uploadsSaver.EnterWithResult(nil, nil)

}
func (ctrl *Controller) GetUploadRequest(fileID int64) (UploadRequest, bool) {
	return ctrl.getUploadRequest(getRequestID(0, fileID, 0))
}
func (ctrl *Controller) CancelUploadRequest(reqID string) {
	req, ok := ctrl.getUploadRequest(reqID)
	ctrl.deleteUploadRequest(reqID)
	if ok && req.cancelFunc != nil {
		req.cancelFunc()
	}
}

func (ctrl *Controller) DownloadAsync(clusterID int32, fileID int64, accessHash uint64, skipDelegates bool) (reqID string, err error) {
	clientFile, err := repo.Files.Get(clusterID, fileID, accessHash)
	if err != nil {
		return "", err
	}
	go func() {
		err = ctrl.download(DownloadRequest{
			MessageID:        clientFile.MessageID,
			ClusterID:        clientFile.ClusterID,
			FileID:           clientFile.FileID,
			AccessHash:       clientFile.AccessHash,
			Version:          clientFile.Version,
			FileSize:         clientFile.FileSize,
			ChunkSize:        defaultChunkSize,
			MaxInFlights:     maxDownloadInFlights,
			FilePath:         repo.Files.GetFilePath(clientFile),
			SkipDelegateCall: skipDelegates,
			PeerID:           clientFile.PeerID,
		})
		logs.WarnOnErr("Error On DownloadAsync", err,
			zap.Int32("ClusterID", clusterID),
			zap.Int64("FileID", fileID),
			zap.Uint64("AccessHash", accessHash),
		)
	}()
	return getRequestID(clusterID, fileID, accessHash), nil
}
func (ctrl *Controller) DownloadSync(clusterID int32, fileID int64, accessHash uint64, skipDelegate bool) (filePath string, err error) {
	clientFile, err := repo.Files.Get(clusterID, fileID, accessHash)
	if err != nil {
		switch err {
		case badger.ErrKeyNotFound:
			logs.Warn("Error On GetFile (Key not found)",
				zap.Int32("ClusterID", clusterID),
				zap.Int64("FileID", fileID),
				zap.Int64("AccessHash", int64(accessHash)),
			)
		default:
			logs.Warn("Error On GetFile",
				zap.Int32("ClusterID", clusterID),
				zap.Int64("FileID", fileID),
				zap.Int64("AccessHash", int64(accessHash)),
				zap.Error(err),
			)
		}

		return "", err
	}
	filePath = repo.Files.GetFilePath(clientFile)
	switch clientFile.Type {
	case msg.GroupProfilePhoto:
		return ctrl.downloadGroupPhoto(clientFile)
	case msg.AccountProfilePhoto:
		return ctrl.downloadAccountPhoto(clientFile)
	case msg.Thumbnail:
		return ctrl.downloadThumbnail(clientFile)
	case msg.Wallpaper:
		return ctrl.downloadWallpaper(clientFile)
	default:
		err = ctrl.download(DownloadRequest{
			MessageID:        clientFile.MessageID,
			ClusterID:        clientFile.ClusterID,
			FileID:           clientFile.FileID,
			AccessHash:       clientFile.AccessHash,
			Version:          clientFile.Version,
			FileSize:         clientFile.FileSize,
			ChunkSize:        defaultChunkSize,
			MaxInFlights:     maxDownloadInFlights,
			FilePath:         filePath,
			SkipDelegateCall: skipDelegate,
			PeerID:           clientFile.PeerID,
		})
	}

	return
}
func (ctrl *Controller) downloadAccountPhoto(clientFile *msg.ClientFile) (filePath string, err error) {
	err = domain.Try(retryMaxAttempts, retryWaitTime, func() error {
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

		filePath = repo.Files.GetFilePath(clientFile)
		res, err := ctrl.network.SendHttp(nil, envelop, ctrl.httpRequestTimeout)
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
	err = domain.Try(retryMaxAttempts, retryWaitTime, func() error {
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

		filePath = repo.Files.GetFilePath(clientFile)
		res, err := ctrl.network.SendHttp(nil, envelop, ctrl.httpRequestTimeout)
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
func (ctrl *Controller) downloadWallpaper(clientFile *msg.ClientFile) (filePath string, err error) {
	err = domain.Try(retryMaxAttempts, retryWaitTime, func() error {
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

		filePath = repo.Files.GetFilePath(clientFile)
		res, err := ctrl.network.SendHttp(nil, envelop, ctrl.httpRequestTimeout)
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
func (ctrl *Controller) downloadThumbnail(clientFile *msg.ClientFile) (filePath string, err error) {
	err = domain.Try(retryMaxAttempts, retryWaitTime, func() error {
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

		filePath = repo.Files.GetFilePath(clientFile) // getThumbnailPath(clientFile.FileID, clientFile.ClusterID)
		res, err := ctrl.network.SendHttp(nil, envelop, ctrl.httpRequestTimeout)
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
	if ctrl.existDownloadRequest(req.GetID()) {
		return domain.ErrAlreadyDownloading
	}
	req.httpContext, req.cancelFunc = context.WithCancel(context.Background())
	req.TempFilePath = fmt.Sprintf("%s.tmp", req.FilePath)
	ctrl.saveDownloads(req)
	ctrl.downloadsRateLimit <- struct{}{}
	defer func() {
		// Remove the Download request from the list
		ctrl.deleteDownloadRequest(req.GetID())

		<-ctrl.downloadsRateLimit
	}()

	ds := &downloadContext{
		rateLimit: make(chan struct{}, req.MaxInFlights),
	}
	_, err := os.Stat(req.TempFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			ds.file, err = os.Create(req.TempFilePath)
			if err != nil {
				ctrl.onCancel(req.GetID(), req.ClusterID, req.FileID, int64(req.AccessHash), true, req.PeerID)
				return err
			}
		} else {
			ctrl.onCancel(req.GetID(), req.ClusterID, req.FileID, int64(req.AccessHash), true, req.PeerID)
			return err
		}
	} else {
		ds.file, err = os.OpenFile(req.TempFilePath, os.O_RDWR, 0666)
		if err != nil {
			ctrl.onCancel(req.GetID(), req.ClusterID, req.FileID, int64(req.AccessHash), true, req.PeerID)
			return err
		}
	}

	if req.FileSize > 0 {
		err := os.Truncate(req.TempFilePath, req.FileSize)
		if err != nil {
			ctrl.onCancel(req.GetID(), req.ClusterID, req.FileID, int64(req.AccessHash), true, req.PeerID)
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
	ds.execute(ctrl)

	return nil
}

func (ctrl *Controller) UploadUserPhoto(filePath string) (reqID string) {
	// support IOS file path
	if strings.HasPrefix(filePath, "file://") {
		filePath = filePath[7:]
	}

	fileID := domain.RandomInt63()
	err := ctrl.upload(UploadRequest{
		IsProfilePhoto: true,
		FileID:         fileID,
		MaxInFlights:   maxUploadInFlights,
		FilePath:       filePath,
		PeerID:         0,
	})
	logs.WarnOnErr("Error On UploadUserPhoto", err)
	reqID = getRequestID(0, fileID, 0)
	return
}
func (ctrl *Controller) UploadGroupPhoto(groupID int64, filePath string) (reqID string) {
	// support IOS file path
	if strings.HasPrefix(filePath, "file://") {
		filePath = filePath[7:]
	}

	fileID := domain.RandomInt63()
	err := ctrl.upload(UploadRequest{
		IsProfilePhoto: true,
		GroupID:        groupID,
		FileID:         fileID,
		MaxInFlights:   maxUploadInFlights,
		FilePath:       filePath,
		PeerID:         groupID,
	})
	logs.WarnOnErr("Error On UploadGroupPhoto", err)
	reqID = getRequestID(0, fileID, 0)
	return
}
func (ctrl *Controller) UploadMessageDocument(
	messageID int64, filePath, thumbPath string, fileID, thumbID int64, fileSha256 []byte, peerID int64, checkSha256 bool,
) {
	if _, err := os.Stat(filePath); err != nil {
		logs.Warn("FileCtrl got error on upload message document (thumbnail)", zap.Error(err))
		return
	}

	if thumbPath != "" {
		if _, err := os.Stat(thumbPath); err != nil {
			logs.Warn("FileCtrl got error on upload message document (thumbnail)", zap.Error(err))
			return
		}
	}
	// We prepare upload request for the actual file before uploading the thumbnail to save it
	// in case of execution stopped, then we are assured that we will continue the upload process
	reqFile := UploadRequest{
		MessageID:    messageID,
		FileID:       fileID,
		FilePath:     filePath,
		ThumbID:      thumbID,
		ThumbPath:    thumbPath,
		MaxInFlights: maxUploadInFlights,
		FileSha256:   string(fileSha256),
		PeerID:       peerID,
		CheckSha256:  checkSha256,
	}

	reqThumb := UploadRequest{
		MessageID:        0,
		FileID:           thumbID,
		FilePath:         thumbPath,
		MaxInFlights:     maxUploadInFlights,
		SkipDelegateCall: false,
		CheckSha256:      checkSha256,
	}

	if thumbID != 0 {
		ctrl.saveUploads(reqThumb)
	}
	ctrl.saveUploads(reqFile)

	go func() {
		if thumbID != 0 {
			// Upload Thumbnail
			err := ctrl.upload(reqThumb)
			logs.WarnOnErr("Error On Upload Thumbnail", err, zap.Int64("FileID", reqThumb.ThumbID))
		}
		// Upload File
		err := ctrl.upload(reqFile)
		logs.WarnOnErr("Error On Upload Message Media", err, zap.Int64("FileID", reqFile.FileID))
	}()

}
func (ctrl *Controller) upload(req UploadRequest) error {
	if req.FilePath == "" {
		ctrl.deleteUploadRequest(req.GetID())
		return domain.ErrNoFilePath
	}
	req.httpContext, req.cancelFunc = context.WithCancel(context.Background())
	ctrl.saveUploads(req)

	ctrl.uploadsRateLimit <- struct{}{}
	defer func() {
		// Remove the Download request from the list
		ctrl.deleteUploadRequest(req.GetID())
		<-ctrl.uploadsRateLimit
	}()

	uploadCtx := &uploadContext{
		rateLimit: make(chan struct{}, req.MaxInFlights),
	}

	fileInfo, err := os.Stat(req.FilePath)
	if err != nil {
		ctrl.onCancel(req.GetID(), 0, req.FileID, 0, true, req.PeerID)
		return err
	} else {
		req.FileSize = fileInfo.Size()
		if req.FileSize <= 0 {
			ctrl.onCancel(req.GetID(), 0, req.FileID, 0, true, req.PeerID)
			return err
		} else if req.FileSize > maxFileSizeAllowedSize {
			ctrl.onCancel(req.GetID(), 0, req.FileID, 0, true, req.PeerID)
			return err
		}
	}

	// If Sha256 exists in the request then we check server if this file has been already uploaded, if true, then
	// we do not upload it again and we call postUploadProcess with the updated details
	if req.CheckSha256 && req.FileSha256 != "" {
		err = ctrl.checkSha256(&req)
		if err == nil {
			logs.Info("File already exists in the server",
				zap.Int64("DocumentID", req.DocumentID),
				zap.Int32("ClusterID", req.ClusterID),
			)
			if !ctrl.postUploadProcess(req) {
				ctrl.onCancel(req.GetID(), 0, req.FileID, 0, true, req.PeerID)
			}
			return nil
		}
	}

	uploadCtx.file, err = os.OpenFile(req.FilePath, os.O_RDONLY, 0666)
	if err != nil {
		ctrl.onCancel(req.GetID(), 0, req.FileID, 0, true, req.PeerID)
		return err
	}

	if req.ChunkSize <= 0 {
		req.ChunkSize = bestChunkSize(req.FileSize)
	}

	uploadCtx.req = req

	// This is blocking call, until all the parts are downloaded
	uploadCtx.execute(ctrl)
	return nil
}
func (ctrl *Controller) checkSha256(uploadRequest *UploadRequest) error {
	envelop := &msg.MessageEnvelope{
		RequestID:   uint64(domain.SequentialUniqueID()),
		Constructor: msg.C_FileGetBySha256,
	}
	req := &msg.FileGetBySha256{
		Sha256:   domain.StrToByte(uploadRequest.FileSha256),
		FileSize: int32(uploadRequest.FileSize),
	}
	envelop.Message, _ = req.Marshal()

	err := domain.Try(3, time.Millisecond*500, func() error {
		res, err := ctrl.network.SendHttp(uploadRequest.httpContext, envelop, time.Second*10)
		if err != nil {
			return err
		}
		switch res.Constructor {
		case msg.C_FileLocation:
			x := &msg.FileLocation{}
			_ = x.Unmarshal(res.Message)
			uploadRequest.ClusterID = x.ClusterID
			uploadRequest.AccessHash = x.AccessHash
			uploadRequest.DocumentID = x.FileID
			uploadRequest.TotalParts = -1 // dirty hack, which queue.Start() knows the upload request is completed
			return nil
		case msg.C_Error:
			x := &msg.Error{}
			_ = x.Unmarshal(res.Message)
		}
		return domain.ErrServer
	})
	return err
}
