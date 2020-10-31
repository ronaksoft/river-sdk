package fileCtrl

import (
	"fmt"
	"git.ronaksoft.com/river/msg/msg"
	"git.ronaksoft.com/river/sdk/internal/ctrl_file/executor"
	networkCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_network"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"github.com/dgraph-io/badger/v2"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"strings"
	"sync"
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
	DbPath               string
	PostUploadProcessCB  func(req msg.ClientFileRequest) bool
	ProgressChangedCB    func(reqID string, clusterID int32, fileID, accessHash int64, percent int64, peerID int64)
	CompletedCB          func(reqID string, clusterID int32, fileID, accessHash int64, filePath string, peerID int64)
	CancelCB             func(reqID string, clusterID int32, fileID, accessHash int64, hasError bool, peerID int64)
}

type Controller struct {
	network      *networkCtrl.Controller
	mtxDownloads sync.Mutex
	downloader   *executor.Executor
	uploader     *executor.Executor

	// Callbacks
	onProgressChanged func(reqID string, clusterID int32, fileID, accessHash int64, percent int64, peerID int64)
	onCompleted       func(reqID string, clusterID int32, fileID, accessHash int64, filePath string, peerID int64)
	onCancel          func(reqID string, clusterID int32, fileID, accessHash int64, hasError bool, peerID int64)
	postUploadProcess func(req msg.ClientFileRequest) bool
}

func New(config Config) *Controller {
	var (
		err error
	)
	ctrl := &Controller{
		network:           config.Network,
		postUploadProcess: config.PostUploadProcessCB,
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

	ctrl.downloader, err = executor.NewExecutor(config.DbPath, "downloader", func(data []byte) executor.Request {
		r := &DownloadRequest{
			ctrl: ctrl,
		}
		_ = r.Unmarshal(data)
		return r
	}, executor.WithConcurrency(config.MaxInflightDownloads))
	if err != nil {
		logs.Fatal("FileCtrl got error on initializing uploader", zap.Error(err))
	}

	ctrl.uploader, err = executor.NewExecutor(config.DbPath, "uploader", func(data []byte) executor.Request {
		r := &UploadRequest{
			ctrl: ctrl,
		}
		_ = r.Unmarshal(data)
		return r
	}, executor.WithConcurrency(config.MaxInflightUploads))
	if err != nil {
		logs.Fatal("FileCtrl got error on initializing uploader", zap.Error(err))
	}

	return ctrl
}

func (ctrl *Controller) Start() {
	reqs, _ := repo.Files.GetAllFileRequests()
	for _, req := range reqs {
		_ = repo.Files.DeleteFileRequest(getRequestID(req.ClusterID, req.FileID, req.AccessHash))
	}
}

func (ctrl *Controller) Stop() {
	// Nothing
}

func (ctrl *Controller) GetUploadRequest(fileID int64) *msg.ClientFileRequest {
	return ctrl.GetRequest(0, fileID, 0)
}
func (ctrl *Controller) GetDownloadRequest(clusterID int32, fileID int64, accessHash uint64) *msg.ClientFileRequest {
	return ctrl.GetRequest(clusterID, fileID, accessHash)
}
func (ctrl *Controller) GetRequest(clusterID int32, fileID int64, accessHash uint64) *msg.ClientFileRequest {
	req, err := repo.Files.GetFileRequest(getRequestID(clusterID, fileID, accessHash))
	if err != nil {
		return nil
	}
	return req
}
func (ctrl *Controller) CancelUploadRequest(fileID int64) {
	logs.Info("FileCtrl cancels UploadRequest", zap.Int64("FileID", fileID))
	ctrl.CancelRequest(getRequestID(0, fileID, 0))
}
func (ctrl *Controller) CancelDownloadRequest(clusterID int32, fileID int64, accessHash uint64) {
	logs.Info("FileCtrl cancels DownloadRequest",
		zap.Int32("ClusterID", clusterID),
		zap.Int64("FileID", fileID),
	)
	ctrl.CancelRequest(getRequestID(clusterID, fileID, accessHash))
}
func (ctrl *Controller) CancelRequest(reqID string) {
	_ = repo.Files.DeleteFileRequest(reqID)
}

func (ctrl *Controller) DownloadAsync(clusterID int32, fileID int64, accessHash uint64, skipDelegates bool) (reqID string, err error) {
	defer logs.RecoverPanic(
		"FileCtrl::DownloadASync",
		domain.M{
			"OS":        domain.ClientOS,
			"Ver":       domain.ClientVersion,
			"FileID":    fileID,
			"ClusterID": clusterID,
		},
		nil,
	)

	clientFile, err := repo.Files.Get(clusterID, fileID, accessHash)
	if err != nil {
		return "", err
	}

	err = ctrl.download(&DownloadRequest{
		ClientFileRequest: msg.ClientFileRequest{
			MessageID:        clientFile.MessageID,
			ClusterID:        clientFile.ClusterID,
			FileID:           clientFile.FileID,
			AccessHash:       clientFile.AccessHash,
			Version:          clientFile.Version,
			FileSize:         clientFile.FileSize,
			ChunkSize:        defaultChunkSize,
			FilePath:         repo.Files.GetFilePath(clientFile),
			SkipDelegateCall: skipDelegates,
			PeerID:           clientFile.PeerID,
		},
	}, false)
	logs.WarnOnErr("Error On DownloadAsync", err,
		zap.Int32("ClusterID", clusterID),
		zap.Int64("FileID", fileID),
		zap.Uint64("AccessHash", accessHash),
	)

	return getRequestID(clusterID, fileID, accessHash), err
}
func (ctrl *Controller) DownloadSync(clusterID int32, fileID int64, accessHash uint64, skipDelegate bool) (filePath string, err error) {
	defer logs.RecoverPanic(
		"FileCtrl::DownloadSync",
		domain.M{
			"OS":        domain.ClientOS,
			"Ver":       domain.ClientVersion,
			"FileID":    fileID,
			"ClusterID": clusterID,
		},
		nil,
	)

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
		err = ctrl.download(&DownloadRequest{
			ClientFileRequest: msg.ClientFileRequest{
				MessageID:        clientFile.MessageID,
				ClusterID:        clientFile.ClusterID,
				FileID:           clientFile.FileID,
				AccessHash:       clientFile.AccessHash,
				Version:          clientFile.Version,
				FileSize:         clientFile.FileSize,
				ChunkSize:        defaultChunkSize,
				FilePath:         filePath,
				SkipDelegateCall: skipDelegate,
				PeerID:           clientFile.PeerID,
			},
		}, true)
	}

	return
}
func (ctrl *Controller) downloadAccountPhoto(clientFile *msg.ClientFile) (filePath string, err error) {
	err = domain.Try(retryMaxAttempts, retryWaitTime, func() error {
		req := &msg.FileGet{
			Location: &msg.InputFileLocation{
				ClusterID:  clientFile.ClusterID,
				FileID:     clientFile.FileID,
				AccessHash: clientFile.AccessHash,
				Version:    clientFile.Version,
			},
			Offset: 0,
			Limit:  0,
		}

		envelop := &msg.MessageEnvelope{}
		envelop.Constructor = msg.C_FileGet
		envelop.Message, _ = req.Marshal()
		envelop.RequestID = uint64(domain.SequentialUniqueID())

		filePath = repo.Files.GetFilePath(clientFile)
		res, err := ctrl.network.SendHttp(nil, envelop)
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
		req := &msg.FileGet{
			Location: &msg.InputFileLocation{
				ClusterID:  clientFile.ClusterID,
				FileID:     clientFile.FileID,
				AccessHash: clientFile.AccessHash,
				Version:    clientFile.Version,
			},
			Offset: 0,
			Limit:  0,
		}

		envelop := &msg.MessageEnvelope{}
		envelop.Constructor = msg.C_FileGet
		envelop.Message, _ = req.Marshal()
		envelop.RequestID = uint64(domain.SequentialUniqueID())

		filePath = repo.Files.GetFilePath(clientFile)
		res, err := ctrl.network.SendHttp(nil, envelop)
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
		req := &msg.FileGet{
			Location: &msg.InputFileLocation{
				ClusterID:  clientFile.ClusterID,
				FileID:     clientFile.FileID,
				AccessHash: clientFile.AccessHash,
				Version:    clientFile.Version,
			},
			Offset: 0,
			Limit:  0,
		}

		envelop := &msg.MessageEnvelope{}
		envelop.Constructor = msg.C_FileGet
		envelop.Message, _ = req.Marshal()
		envelop.RequestID = uint64(domain.SequentialUniqueID())

		filePath = repo.Files.GetFilePath(clientFile)
		res, err := ctrl.network.SendHttp(nil, envelop)
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
		req := &msg.FileGet{
			Location: &msg.InputFileLocation{
				ClusterID:  clientFile.ClusterID,
				FileID:     clientFile.FileID,
				AccessHash: clientFile.AccessHash,
				Version:    clientFile.Version,
			},
			Offset: 0,
			Limit:  0,
		}

		envelop := &msg.MessageEnvelope{}
		envelop.Constructor = msg.C_FileGet
		envelop.Message, _ = req.Marshal()
		envelop.RequestID = uint64(domain.SequentialUniqueID())

		filePath = repo.Files.GetFilePath(clientFile) // getThumbnailPath(clientFile.FileID, clientFile.ClusterID)
		res, err := ctrl.network.SendHttp(nil, envelop)
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
func (ctrl *Controller) download(req *DownloadRequest, blocking bool) error {
	logs.Info("FileCtrl received download request",
		zap.Bool("Blocking", blocking),
		zap.Int64("FileID", req.FileID),
		zap.Uint64("AccessHash", req.AccessHash),
		zap.Int64("Size", req.FileSize),
	)

	if req.ClusterID == 0 {
		return domain.ErrInvalidData
	}
	_, err := repo.Files.GetFileRequest(getRequestID(req.ClusterID, req.FileID, req.AccessHash))
	if err == nil {
		return domain.ErrAlreadyDownloading
	}

	_ = repo.Files.SaveFileRequest(
		getRequestID(req.ClusterID, req.FileID, req.AccessHash),
		&req.ClientFileRequest,
		false,
	)

	req.TempPath = fmt.Sprintf("%s.tmp", req.FilePath)
	if blocking {
		waitGroup := &sync.WaitGroup{}
		waitGroup.Add(1)
		err = ctrl.downloader.ExecuteAndWait(waitGroup, req)
		if err != nil {
			return err
		}
		waitGroup.Wait()
	} else {
		err = ctrl.downloader.Execute(req)
		if err != nil {
			return err
		}

	}

	return nil
}

func (ctrl *Controller) UploadUserPhoto(filePath string) (reqID string) {
	// support IOS file path
	if strings.HasPrefix(filePath, "file://") {
		filePath = filePath[7:]
	}

	fileID := domain.RandomInt63()
	err := ctrl.upload(msg.ClientFileRequest{
		IsProfilePhoto: true,
		FileID:         fileID,
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
	err := ctrl.upload(msg.ClientFileRequest{
		IsProfilePhoto: true,
		GroupID:        groupID,
		FileID:         fileID,
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
	defer logs.RecoverPanic(
		"FileCtrl::UploadMessageDocument",
		domain.M{
			"OS":       domain.ClientOS,
			"Ver":      domain.ClientVersion,
			"FilePath": filePath,
		},
		nil,
	)

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

	reqFile := msg.ClientFileRequest{
		MessageID:   messageID,
		FileID:      fileID,
		FilePath:    filePath,
		ThumbID:     thumbID,
		ThumbPath:   thumbPath,
		FileSha256:  fileSha256,
		PeerID:      peerID,
		CheckSha256: checkSha256,
	}

	// If there is a thumbnail then set the reqFile as the next
	if thumbID != 0 {
		reqFile = msg.ClientFileRequest{
			Next: &msg.ClientFileRequest{
				MessageID:   messageID,
				FileID:      fileID,
				FilePath:    filePath,
				ThumbID:     thumbID,
				ThumbPath:   thumbPath,
				FileSha256:  fileSha256,
				PeerID:      peerID,
				CheckSha256: checkSha256,
			},
			MessageID:        0,
			FileID:           thumbID,
			FilePath:         thumbPath,
			SkipDelegateCall: true,
		}
	}

	err := ctrl.upload(reqFile)
	if err != nil {
		logs.WarnOnErr("Error On Upload Message Media", err, zap.Int64("FileID", reqFile.FileID))
	}
}
func (ctrl *Controller) upload(req msg.ClientFileRequest) error {
	if req.ClusterID != 0 {
		return domain.ErrInvalidData
	}
	if req.FilePath == "" {
		return domain.ErrNoFilePath
	}

	_, err := repo.Files.GetFileRequest(getRequestID(req.ClusterID, req.FileID, req.AccessHash))
	if err == nil {
		return domain.ErrAlreadyUploading
	}

	_ = repo.Files.SaveFileRequest(
		getRequestID(0, req.FileID, 0),
		&req,
		false,
	)

	err = ctrl.uploader.Execute(&UploadRequest{
		ClientFileRequest: req,
	})
	if err != nil {
		return err
	}

	return nil
}
