package fileCtrl

import (
    "fmt"
    "io/ioutil"
    "os"
    "strings"
    "sync"

    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/river-sdk/internal/ctrl_file/executor"
    networkCtrl "github.com/ronaksoft/river-sdk/internal/ctrl_network"
    "github.com/ronaksoft/river-sdk/internal/domain"
    "github.com/ronaksoft/river-sdk/internal/logs"
    "github.com/ronaksoft/river-sdk/internal/repo"
    "github.com/ronaksoft/river-sdk/internal/request"
    "github.com/ronaksoft/rony"
    "github.com/ronaksoft/rony/tools"
    "go.uber.org/zap"
)

/*
   Creation Time: 2019 - Aug - 18
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/
var (
    logger *logs.Logger
)

func init() {
    logger = logs.With("FileCtrl")
}

type Config struct {
    Network              *networkCtrl.Controller
    MaxInflightDownloads int32
    MaxInflightUploads   int32
    DbPath               string
    PostUploadProcessCB  func(req *msg.ClientFileRequest) bool
    ProgressChangedCB    func(reqID string, clusterID int32, fileID, accessHash int64, percent int64, peerID int64)
    CompletedCB          func(reqID string, clusterID int32, fileID, accessHash int64, filePath string, peerID int64)
    CancelCB             func(reqID string, clusterID int32, fileID, accessHash int64, hasError bool, peerID int64)
}

type Controller struct {
    network    *networkCtrl.Controller
    downloader *executor.Executor
    uploader   *executor.Executor

    // Callbacks
    onProgressChanged func(reqID string, clusterID int32, fileID, accessHash int64, percent int64, peerID int64)
    onCompleted       func(reqID string, clusterID int32, fileID, accessHash int64, filePath string, peerID int64)
    onCancel          func(reqID string, clusterID int32, fileID, accessHash int64, hasError bool, peerID int64)
    postUploadProcess func(req *msg.ClientFileRequest) bool
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
        logger.Fatal("got error on initializing uploader", zap.Error(err))
    }

    ctrl.uploader, err = executor.NewExecutor(config.DbPath, "uploader", func(data []byte) executor.Request {
        r := &UploadRequest{
            cfr:       &msg.ClientFileRequest{},
            ctrl:      ctrl,
            startTime: domain.Now(),
        }
        _ = r.cfr.Unmarshal(data)
        return r
    }, executor.WithConcurrency(config.MaxInflightUploads))
    if err != nil {
        logger.Fatal("got error on initializing uploader", zap.Error(err))
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
    logger.Info("cancels UploadRequest", zap.Int64("FileID", fileID))
    ctrl.CancelRequest(getRequestID(0, fileID, 0))
}
func (ctrl *Controller) CancelDownloadRequest(clusterID int32, fileID int64, accessHash uint64) {
    logger.Info("cancels DownloadRequest",
        zap.Int32("ClusterID", clusterID),
        zap.Int64("FileID", fileID),
    )
    ctrl.CancelRequest(getRequestID(clusterID, fileID, accessHash))
}
func (ctrl *Controller) CancelRequest(reqID string) {
    _ = repo.Files.DeleteFileRequest(reqID)
}

func (ctrl *Controller) DownloadAsync(clusterID int32, fileID int64, accessHash uint64, skipDelegates bool) (reqID string, err error) {
    defer logger.RecoverPanic(
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
            ChunkSize:        DefaultChunkSize,
            FilePath:         repo.Files.GetFilePath(clientFile),
            SkipDelegateCall: skipDelegates,
            PeerID:           clientFile.PeerID,
        },
    }, false)
    logger.WarnOnErr("Error On DownloadAsync", err,
        zap.Int32("ClusterID", clusterID),
        zap.Int64("FileID", fileID),
        zap.Uint64("AccessHash", accessHash),
    )

    return getRequestID(clusterID, fileID, accessHash), err
}
func (ctrl *Controller) DownloadSync(clusterID int32, fileID int64, accessHash uint64, skipDelegate bool) (filePath string, err error) {
    defer logger.RecoverPanic(
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
        return "", err
    }
    filePath = repo.Files.GetFilePath(clientFile)
    switch clientFile.Type {
    case msg.ClientFileType_GroupProfilePhoto, msg.ClientFileType_AccountProfilePhoto,
        msg.ClientFileType_Thumbnail, msg.ClientFileType_Wallpaper:
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
        err = tools.Try(RetryMaxAttempts, RetryWaitTime, func() error {
            reqCB := request.NewCallback(
                0, 0, domain.NextRequestID(), msg.C_FileGet, req,
                func() {
                    err = domain.ErrRequestTimeout
                },
                func(res *rony.MessageEnvelope) {
                    switch res.Constructor {
                    case rony.C_Error:
                        x := &rony.Error{}
                        _ = x.Unmarshal(res.Message)
                        err = x
                    case msg.C_File:
                        x := new(msg.File)
                        err = x.Unmarshal(res.Message)
                        if err != nil {
                            return
                        }

                        // write to file path
                        err = ioutil.WriteFile(filePath, x.Bytes, 0666)
                        if err != nil {
                            return
                        }

                        // save to DB
                        _ = repo.Files.Save(clientFile)
                        return
                    default:
                        err = domain.ErrServer
                        return
                    }

                },
                nil, false, 0, domain.HttpRequestTimeout,
            )

            ctrl.network.HttpCommand(nil, reqCB)
            return err
        })
        return filePath, err
    default:
        err = ctrl.download(&DownloadRequest{
            ClientFileRequest: msg.ClientFileRequest{
                MessageID:        clientFile.MessageID,
                ClusterID:        clientFile.ClusterID,
                FileID:           clientFile.FileID,
                AccessHash:       clientFile.AccessHash,
                Version:          clientFile.Version,
                FileSize:         clientFile.FileSize,
                ChunkSize:        DefaultChunkSize,
                FilePath:         filePath,
                SkipDelegateCall: skipDelegate,
                PeerID:           clientFile.PeerID,
            },
        }, true)
    }

    return
}
func (ctrl *Controller) download(req *DownloadRequest, blocking bool) error {
    logger.Info("received download request",
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

    _, _ = repo.Files.SaveFileRequest(
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
    filePath = strings.TrimPrefix(filePath, "file://")
    fileID := domain.RandomInt63()
    err := ctrl.upload(&msg.ClientFileRequest{
        IsProfilePhoto: true,
        FileID:         fileID,
        FilePath:       filePath,
        PeerID:         0,
    })
    logger.WarnOnErr("Error On UploadUserPhoto", err)
    reqID = getRequestID(0, fileID, 0)
    return
}
func (ctrl *Controller) UploadGroupPhoto(groupID int64, filePath string) (reqID string) {
    // support IOS file path
    filePath = strings.TrimPrefix(filePath, "file://")

    fileID := domain.RandomInt63()
    err := ctrl.upload(&msg.ClientFileRequest{
        IsProfilePhoto: true,
        GroupID:        groupID,
        FileID:         fileID,
        FilePath:       filePath,
        PeerID:         groupID,
    })
    logger.WarnOnErr("Error On UploadGroupPhoto", err)
    reqID = getRequestID(0, fileID, 0)
    return
}
func (ctrl *Controller) UploadMessageDocument(
        messageID int64, filePath, thumbPath string, fileID, thumbID int64, fileSha256 []byte, peerID int64, checkSha256 bool,
) {
    defer logger.RecoverPanic(
        "FileCtrl::UploadMessageDocument",
        domain.M{
            "OS":       domain.ClientOS,
            "Ver":      domain.ClientVersion,
            "FilePath": filePath,
        },
        nil,
    )

    if _, err := os.Stat(filePath); err != nil {
        logger.Warn("got error on upload message document (thumbnail)", zap.Error(err))
        return
    }

    if thumbPath != "" {
        if _, err := os.Stat(thumbPath); err != nil {
            logger.Warn("got error on upload message document (thumbnail)", zap.Error(err))
            return
        }
    }

    reqFile := &msg.ClientFileRequest{
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
        reqFile = &msg.ClientFileRequest{
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
        logger.WarnOnErr("Error On Upload Message Media", err, zap.Int64("FileID", reqFile.FileID))
    }
}
func (ctrl *Controller) upload(req *msg.ClientFileRequest) error {
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

    _, _ = repo.Files.SaveFileRequest(
        getRequestID(0, req.FileID, 0),
        req,
        false,
    )

    err = ctrl.uploader.Execute(&UploadRequest{
        cfr:       req,
        startTime: domain.Now(),
    })
    if err != nil {
        return err
    }

    return nil
}
