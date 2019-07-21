package fileCtrl

import (
	"crypto/md5"
	"errors"
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"io"
	"io/ioutil"
	"mime"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
)

var (
	ctx      *Controller
	dirAudio string
	dirFile  string
	dirPhoto string
	dirVideo string
	dirCache string
)

const (
	// FileSizeThresholdForCheckHash for files that are smaller than this number we will calculate md5 hash to do not reupload same file twice
	FileSizeThresholdForCheckHash = 10 * 1024 * 1024 // 10MB
)

// Config network controller config
type Config struct {
	ServerAddress       string
	OnUploadCompleted   domain.OnFileUploadCompleted
	ProgressCallback    domain.OnFileStatusChanged
	OnDownloadCompleted domain.OnFileDownloadCompleted
	OnFileUploadError   domain.OnFileUploadError
	OnFileDownloadError domain.OnFileDownloadError
}

// Controller manages files download/upload/status and cache
type Controller struct {
	authKey    []byte
	authID     int64
	messageSeq int64

	ServerEndpoint string
	NetworkStatus  domain.NetworkStatus

	mxDown               sync.Mutex
	mxUp                 sync.Mutex
	UploadQueue          map[int64]*File
	DownloadQueue        map[int64]*File
	UploadQueueStarted   bool
	DownloadQueueStarted bool

	chStopUploader    chan bool
	chStopDownloader  chan bool
	chNewDownloadItem chan bool
	chNewUploadItem   chan bool

	onUploadCompleted   domain.OnFileUploadCompleted
	onDownloadCompleted domain.OnFileDownloadCompleted
	progressCallback    domain.OnFileStatusChanged
	onUploadError       domain.OnFileUploadError
	onDownloadError     domain.OnFileDownloadError
}

func New(config Config) *Controller {
	// maxUploadConcurrency := 3
	// maxDownloadConcurrency := 3
	ctx = &Controller{
		ServerEndpoint:      config.ServerAddress,
		UploadQueue:         make(map[int64]*File, 0),
		DownloadQueue:       make(map[int64]*File, 0),
		chStopUploader:      make(chan bool, 1),
		chStopDownloader:    make(chan bool, 1),
		chNewDownloadItem:   make(chan bool, 1),
		chNewUploadItem:     make(chan bool, 1),
		onUploadCompleted:   config.OnUploadCompleted,
		progressCallback:    config.ProgressCallback,
		onDownloadCompleted: config.OnDownloadCompleted,
		onUploadError:       config.OnFileUploadError,
		onDownloadError:     config.OnFileDownloadError,
	}

	go ctx.startDownloadQueue()
	go ctx.startUploadQueue()

	return ctx
}

func (ctrl *Controller) startDownloadQueue() {
	for {
		if ctrl.NetworkStatus == domain.NetworkDisconnected || ctrl.NetworkStatus == domain.NetworkConnecting {
			time.Sleep(100 * time.Millisecond)
		}
		logs.Info("StartDownloadQueue")
		ctrl.DownloadQueueStarted = true
		select {
		case <-ctrl.chStopDownloader:
			ctrl.mxDown.Lock()
			for _, v := range ctrl.DownloadQueue {
				v.Stop()
			}
			ctrl.mxDown.Unlock()
			ctrl.DownloadQueueStarted = false
			return
		case <-ctrl.chNewDownloadItem:
			ctrl.mxDown.Lock()
			for _, theFile := range ctrl.DownloadQueue {
				logs.Info("NewDownload",
					zap.Any("MessageID", theFile.MessageID),
					zap.Any("TotalSize", theFile.TotalSize),
					zap.Any("PartList", theFile.PartList),
					zap.Any("FilePath", theFile.FilePath),
				)
				theFile.StartDownload(ctrl)
			}
			ctrl.mxDown.Unlock()
		}
	}
}

// downloadRequest send request to server
func (ctrl *Controller) downloadRequest(req *msg.MessageEnvelope, fs *File, partIdx int64) {
	// time out has been set in Send()
	res, err := ctrl.Send(req)
	if err == nil {
		switch res.Constructor {
		case msg.C_Error:
			// remove download from queue
			ctrl.DeleteFromQueue(fs.MessageID, domain.RequestStatusError)
			logs.Warn("downloadRequest() received error response and removed item from queue", zap.Int64("MsgID", fs.MessageID))
			fs.RequestStatus = domain.RequestStatusError
			repo.Files.UpdateFileStatus(fs.MessageID, fs.RequestStatus)

			if ctrl.onDownloadError != nil {
				ctrl.onDownloadError(fs.MessageID, int64(req.RequestID), fs.FilePath, res.Message)
			}
		case msg.C_File:
			x := new(msg.File)
			err := x.Unmarshal(res.Message)
			if err != nil {
				logs.Error("downloadRequest() failed to unmarshal C_File", zap.Error(err))
				fs.retryCounter++
				break
			}

			if len(x.Bytes) == 0 {
				logs.Error("downloadRequest() Received 0 bytes from server ",
					zap.Int64("MsgID", fs.MessageID),
					zap.Int64("PartNo", partIdx),
					zap.Int64("TotalSize", fs.TotalSize),
				)
				fs.retryCounter++
				break

			} else {
				// reset counter
				fs.retryCounter = 0
			}

			isCompleted, err := fs.Write(x.Bytes, partIdx)
			if err != nil {
				logs.Error("downloadRequest() failed write to file", zap.Error(err))
			} else if isCompleted {
				// call completed delegate
				ctrl.downloadCompleted(fs.MessageID, fs.FilePath, fs.Type)
			}
		default:
			// increase counter
			fs.retryCounter++
			logs.Error("downloadRequest() received unknown response", zap.Error(err))
		}
	} else {
		// increase counter
		fs.chPartList <- partIdx
		fs.retryCounter++
		logs.Error("downloadRequest()", zap.Error(err))
	}
	if fs.retryCounter > domain.FileMaxRetry {
		// remove download from queue
		ctrl.DeleteFromQueue(fs.MessageID, domain.RequestStatusError)
		logs.Warn("downloadRequest() download request errors passed retry threshold", zap.Int64("MsgID", fs.MessageID))
		fs.RequestStatus = domain.RequestStatusError
		repo.Files.UpdateFileStatus(fs.MessageID, fs.RequestStatus)
		if ctrl.onDownloadError != nil {
			x := new(msg.Error)
			x.Code = "00"
			x.Items = "download request errors passed retry threshold"
			xbuff, _ := x.Marshal()
			ctrl.onDownloadError(fs.MessageID, int64(req.RequestID), fs.FilePath, xbuff)
		}
	}
}

func (ctrl *Controller) downloadCompleted(msgID int64, filePath string, stateType domain.FileStateType) {
	// delete file status
	ctrl.DeleteFromQueue(msgID, domain.RequestStatusCompleted)
	repo.Files.DeleteStatus(msgID)
	if ctrl.onDownloadCompleted != nil {
		ctrl.onDownloadCompleted(msgID, filePath, stateType)
	}
}

func (ctrl *Controller) startUploadQueue() {
	for {
		if ctrl.NetworkStatus == domain.NetworkDisconnected || ctrl.NetworkStatus == domain.NetworkConnecting {
			time.Sleep(1000 * time.Millisecond)
		}

		ctrl.UploadQueueStarted = true
		select {
		case <-ctrl.chStopUploader:
			ctrl.mxUp.Lock()
			for _, v := range ctrl.UploadQueue {
				v.Stop()
			}
			ctrl.mxUp.Unlock()
			ctrl.UploadQueueStarted = false
			return
		case <-ctrl.chNewUploadItem:
			ctrl.mxUp.Lock()
			for _, v := range ctrl.UploadQueue {
				go v.StartUpload(ctrl)
			}
			ctrl.mxUp.Unlock()

		}
	}
}

// uploadRequest send request to server
func (ctrl *Controller) uploadRequest(req *msg.MessageEnvelope, count int64, theFile *File, partIdx int64) {
	// time out has been set in Send()
	res, err := ctrl.Send(req)
	if err == nil {
		switch res.Constructor {
		case msg.C_Error:
			// remove upload from
			ctrl.DeleteFromQueue(theFile.MessageID, domain.RequestStatusError)
			logs.Warn("uploadRequest() received error response and removed item from queue", zap.Int64("MsgID", theFile.MessageID))
			theFile.RequestStatus = domain.RequestStatusError
			repo.Files.UpdateFileStatus(theFile.MessageID, theFile.RequestStatus)
			if ctrl.onUploadError != nil {
				ctrl.onUploadError(theFile.MessageID, int64(req.RequestID), theFile.FilePath, res.Message)
			}
		case msg.C_Bool:
			x := new(msg.Bool)
			_ = x.Unmarshal(res.Message)
			// reset counter
			theFile.retryCounter = 0
			if x.Result {
				isThumbnail := theFile.ThumbPosition < theFile.ThumbTotalSize
				if isThumbnail {
					theFile.ReadCommit(count, true, 0)
				} else {
					select {
					case theFile.chUploadProgress <- partIdx:
					default:
						// progress monitor is exited already
					}
				}
			}
		default:
			// increase counter
			theFile.retryCounter++
			logs.Warn("uploadRequest() received unknown response", zap.Error(err))

		}
	} else {
		// increase counter
		theFile.retryCounter++
		logs.Warn("uploadRequest()", zap.Error(err))
	}

	if theFile.retryCounter > domain.FileMaxRetry {
		// remove upload from queue
		ctrl.DeleteFromQueue(theFile.MessageID, domain.RequestStatusError)
		logs.Error("uploadRequest() upload request errors passed retry threshold", zap.Int64("MsgID", theFile.MessageID))
		theFile.RequestStatus = domain.RequestStatusError
		repo.Files.UpdateFileStatus(theFile.MessageID, theFile.RequestStatus)
		if ctrl.onUploadError != nil {
			x := new(msg.Error)
			x.Code = "00"
			x.Items = "upload request errors passed retry threshold"
			xbuff, _ := x.Marshal()
			ctrl.onUploadError(theFile.MessageID, int64(req.RequestID), theFile.FilePath, xbuff)
		}
	}
}

func (ctrl *Controller) uploadCompleted(msgID, fileID, targetID int64,
	clusterID int32, totalParts int64,
	stateType domain.FileStateType,
	filePath string,
	uploadRequest *msg.ClientSendMessageMedia,
	thumbFileID int64,
	thumbTotalParts int32,
) {
	// delete file status
	ctrl.DeleteFromQueue(msgID, domain.RequestStatusCompleted)
	repo.Files.DeleteStatus(msgID)
	if ctrl.onUploadCompleted != nil {
		ctrl.onUploadCompleted(msgID, fileID, targetID, clusterID, totalParts, stateType, filePath, uploadRequest, thumbFileID, thumbTotalParts)
	}
}

// Stop set stop flag
func (ctrl *Controller) Stop() {
	logs.Debug("FileController Stopping")

	if ctrl.UploadQueueStarted {
		ctrl.chStopUploader <- true
	}
	if ctrl.DownloadQueueStarted {
		ctrl.chStopDownloader <- true
	}

	if ctx != nil {
		ctx = nil
	}
	logs.Debug("FileController Stopped")
}

// Upload file to server
func (ctrl *Controller) Upload(fileID int64, req *msg.ClientPendingMessage) error {
	x := new(msg.ClientSendMessageMedia)
	err := x.Unmarshal(req.Media)
	if err != nil {
		return err
	}

	file, err := os.Open(x.FilePath)
	if err != nil {
		return err
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}
	if x.FileName == "" {
		x.FileName = fileInfo.Name()
	}
	fileSize := fileInfo.Size() // size in Byte
	if fileSize > domain.FileMaxAllowedSize {
		return errors.New("max allowed file size is 750 MB")
	}

	state := NewFile(req.ID, fileID, 0, fileSize, x.FilePath, domain.FileStateUpload, 0, 0, 0, ctrl.progressCallback)
	state.UploadRequest = x

	thumbFile, err := os.Open(x.ThumbFilePath)
	if err == nil {
		thumbInfo, err := thumbFile.Stat()
		if err == nil {
			state.ThumbFileID = domain.SequentialUniqueID()
			state.ThumbFilePath = x.ThumbFilePath
			state.ThumbPosition = 0
			state.ThumbTotalSize = thumbInfo.Size()
			state.ThumbPartNo = 0
			state.ThumbTotalParts = int32(CalculatePartsCount(thumbInfo.Size()))
		}
	}

	ctrl.AddToQueue(state)
	repo.Files.SaveStatus(state.GetDTO())
	return nil
}

// Download add download request
func (ctrl *Controller) Download(userMessage *msg.UserMessage) {
	var theFile *File
	filesStatus, _ := repo.Files.GetStatus(userMessage.ID)
	if filesStatus != nil {
		if filesStatus.IsCompleted {
			ctrl.downloadCompleted(filesStatus.MessageID, filesStatus.FilePath, domain.FileStateType(filesStatus.Type))
			return
		}
		theFile = new(File)
		theFile.LoadDTO(*filesStatus, ctrl.progressCallback)
	} else {
		var docID int64
		var clusterID int32
		var accessHash uint64
		var version int32
		var fileSize int32
		var filePath string
		switch userMessage.MediaType {
		case msg.MediaTypeEmpty:
		case msg.MediaTypeDocument:
			x := new(msg.MediaDocument)
			_ = x.Unmarshal(userMessage.Media)
			docID = x.Doc.ID
			clusterID = x.Doc.ClusterID
			accessHash = x.Doc.AccessHash
			version = x.Doc.Version
			fileSize = x.Doc.FileSize
			filePath = GetFilePath(x.Doc.MimeType, x.Doc.ID)

			theFile = NewFile(userMessage.ID, docID, 0, int64(fileSize), filePath, domain.FileStateDownload, clusterID, accessHash, version, ctrl.progressCallback)
			theFile.DownloadRequest = x.Doc
		default:
			return
		}
	}

	if theFile != nil {
		theFile.RequestStatus = domain.RequestStatusInProgress
		ctrl.AddToQueue(theFile)
		repo.Files.SaveStatus(theFile.GetDTO())
		repo.Files.UpdateFileStatus(theFile.MessageID, theFile.RequestStatus)
	}
}

// AddToQueue add request to queue
func (ctrl *Controller) AddToQueue(theFile *File) {
	switch theFile.Type {
	case domain.FileStateUpload, domain.FileStateUploadAccountPhoto, domain.FileStateUploadGroupPhoto:
		ctrl.mxUp.Lock()
		_, ok := ctrl.UploadQueue[theFile.MessageID]
		if !ok {
			ctrl.UploadQueue[theFile.MessageID] = theFile
		}
		ctrl.mxUp.Unlock()
		if !ok {
			ctrl.chNewUploadItem <- true
		}
	case domain.FileStateDownload:
		ctrl.mxDown.Lock()
		_, ok := ctrl.DownloadQueue[theFile.MessageID]
		if !ok {
			ctrl.DownloadQueue[theFile.MessageID] = theFile
		}
		ctrl.mxDown.Unlock()
		if !ok {
			ctrl.chNewDownloadItem <- true
		}
	}
}

// DeleteFromQueue remove items from download/upload queue and stop them
func (ctrl *Controller) DeleteFromQueue(msgID int64, status domain.RequestStatus) {
	ctrl.mxUp.Lock()
	up, uok := ctrl.UploadQueue[msgID]
	if uok {
		up.RequestStatus = status
		delete(ctrl.UploadQueue, msgID)
		up.Stop()
	}
	ctrl.mxUp.Unlock()

	ctrl.mxDown.Lock()
	down, dok := ctrl.DownloadQueue[msgID]
	if dok {
		down.RequestStatus = status
		delete(ctrl.DownloadQueue, msgID)
		down.Stop()
	}
	ctrl.mxDown.Unlock()
}

// CalculateMD5 this will calculate md5 hash for files that are smaller than threshold
func (ctrl *Controller) CalculateMD5(file *os.File) (string, error) {
	// get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return "", err
	}
	// check file size
	if fileInfo.Size() < FileSizeThresholdForCheckHash {
		h := md5.New()
		if _, err := io.Copy(h, file); err != nil {
			return "", err
		}
		strMD5 := fmt.Sprintf("%x", h.Sum(nil))
		return strMD5, nil
	}
	return "", errors.New("file size is grater than threshold")
}

// SetAuthorization set client AuthID and AuthKey to encrypt&decrypt network requests
func (ctrl *Controller) SetAuthorization(authID int64, authKey []byte) {
	ctrl.authKey = make([]byte, len(authKey))
	ctrl.authID = authID
	copy(ctrl.authKey, authKey)
}

// LoadQueueFromDB load in progress request from database
func (ctrl *Controller) LoadQueueFromDB() {
	// Load pended file status
	filesStatuses := repo.Files.GetAllStatuses()
	for _, filesStatus := range filesStatuses {
		fs := new(File)
		fs.LoadDTO(filesStatus, ctrl.progressCallback)
		if fs.RequestStatus == domain.RequestStatusPaused ||
			fs.RequestStatus == domain.RequestStatusCanceled ||
			fs.RequestStatus == domain.RequestStatusCompleted {
			continue
		}
		ctrl.AddToQueue(fs)
	}
}

// SetNetworkStatus called on network controller state changes to inform file controller
func (ctrl *Controller) SetNetworkStatus(state domain.NetworkStatus) {
	ctrl.NetworkStatus = state
	if state == domain.NetworkWeak || state == domain.NetworkSlow || state == domain.NetworkFast {
		ctrl.LoadQueueFromDB()
	}
}

// DownloadAccountPhoto download account photo from server its sync
func (ctrl *Controller) DownloadAccountPhoto(userID int64, photo *msg.UserPhoto, isBig bool) (string, error) {
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

	filePath := GetAccountAvatarPath(userID, req.Location.FileID)
	res, err := ctrl.Send(envelop)
	if err == nil {
		switch res.Constructor {
		case msg.C_Error:
			strErr := ""
			x := new(msg.Error)
			if err := x.Unmarshal(res.Message); err == nil {
				strErr = "Code :" + x.Code + ", Items :" + x.Items
			}
			return "", fmt.Errorf("received error response {userID: %d,  %s }", userID, strErr)
		case msg.C_File:
			x := new(msg.File)
			err := x.Unmarshal(res.Message)
			if err != nil {
				return "", err
			}

			// write to file path
			err = ioutil.WriteFile(filePath, x.Bytes, 0666)
			if err != nil {
				return "", err
			}

			// save to DB
			return filePath, nil

		default:
			return "", fmt.Errorf("received unknown response constructor {UserId : %d}", userID)
		}
	}
	return "", err
}

// DownloadGroupPhoto download group photo from server its sync
func (ctrl *Controller) DownloadGroupPhoto(groupID int64, photo *msg.GroupPhoto, isBig bool) (string, error) {
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

	filePath := GetGroupAvatarPath(groupID, req.Location.FileID)
	res, err := ctrl.Send(envelop)
	if err == nil {
		switch res.Constructor {
		case msg.C_Error:
			strErr := ""
			x := new(msg.Error)
			if err := x.Unmarshal(res.Message); err == nil {
				strErr = "Code :" + x.Code + ", Items :" + x.Items
			}
			return "", fmt.Errorf("received error response {GroupID: %d,  %s }", groupID, strErr)
		case msg.C_File:
			x := new(msg.File)
			err := x.Unmarshal(res.Message)
			if err != nil {
				return "", err
			}

			// write to file path
			err = ioutil.WriteFile(filePath, x.Bytes, 0666)
			if err != nil {
				return "", err
			}

			// save to DB
			return filePath, nil

		default:
			return "", fmt.Errorf("received unknown response constructor {GroupID : %d}", groupID)
		}
	}
	return "", err
}

// DownloadThumbnail download thumbnail from server its sync
func (ctrl *Controller) DownloadThumbnail(fileID int64, accessHash uint64, clusterID, version int32) (string, error) {
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

	filePath := GetThumbnailPath(fileID, clusterID)
	res, err := ctrl.Send(envelop)
	if err == nil {
		switch res.Constructor {
		case msg.C_Error:
			strErr := ""
			x := new(msg.Error)
			if err := x.Unmarshal(res.Message); err == nil {
				strErr = "Code :" + x.Code + ", Items :" + x.Items
			}
			return "", fmt.Errorf("DownloadThumbnail() received error response { %s }", strErr)
		case msg.C_File:
			x := new(msg.File)
			err := x.Unmarshal(res.Message)
			if err != nil {
				return "", err
			}

			// write to file path
			err = ioutil.WriteFile(filePath, x.Bytes, 0666)
			if err != nil {
				return "", err
			}

			// save to DB
			return filePath, nil

		default:
			return "", fmt.Errorf("DownloadThumbnail() received unknown response constructor")
		}
	}
	return "", err
}

func (ctrl *Controller) ClearFiles(filePaths []string) error {
	for _, filePath := range filePaths {
		if err := os.Remove(filePath); err != nil {
			logs.Warn("ClearFiles::Error removing files", zap.String(fmt.Sprintf(" file path: %s", filePath), err.Error()))
			return err
		}
	}
	return nil
}

// SetRootFolders directory paths to download files
func SetRootFolders(audioDir, fileDir, photoDir, videoDir, cacheDir string) {
	dirAudio = audioDir
	dirFile = fileDir
	dirPhoto = photoDir
	dirVideo = videoDir
	dirCache = cacheDir
}

func GetFilePath(mimeType string, docID int64) string {
	mimeType = strings.ToLower(mimeType)
	var ext string
	if ext == "" {
		exts, _ := mime.ExtensionsByType(mimeType)
		if len(exts) > 0 {
			ext = exts[len(exts)-1]
		}
	}

	// if the file is opus type,
	// means its voice file so it should be saved in cache folder
	// so user could not access to it by file manager
	switch {
	case mimeType == "audio/ogg":
		ext = ".ogg"
		return path.Join(dirCache, fmt.Sprintf("%d%s", docID, ext))
	case strings.HasPrefix(mimeType, "video/"):
		return path.Join(dirVideo, fmt.Sprintf("%d%s", docID, ext))
	case strings.HasPrefix(mimeType, "audio/"):
		return path.Join(dirAudio, fmt.Sprintf("%d%s", docID, ext))
	case strings.HasPrefix(mimeType, "image/"):
		return path.Join(dirPhoto, fmt.Sprintf("%d%s", docID, ext))
	default:
		return path.Join(dirFile, fmt.Sprintf("%d%s", docID, ext))
	}
}

func GetThumbnailPath(fileID int64, clusterID int32) string {
	return path.Join(dirCache, fmt.Sprintf("%d%d%s", fileID, clusterID, ".jpg"))
}

func GetAccountAvatarPath(userID int64, fileID int64) string {
	return path.Join(dirCache, fmt.Sprintf("u%d_%d%s", userID, fileID, ".jpg"))
}

func GetGroupAvatarPath(groupID int64, fileID int64) string {
	return path.Join(dirCache, fmt.Sprintf("g%d_%d%s", groupID, fileID, ".jpg"))
}
