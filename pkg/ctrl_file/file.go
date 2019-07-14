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
	"strconv"
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
	ctx = &Controller{
		ServerEndpoint:      config.ServerAddress,
		UploadQueue:         make(map[int64]*File, 0),
		DownloadQueue:       make(map[int64]*File, 0),
		chStopUploader:      make(chan bool),
		chStopDownloader:    make(chan bool),
		chNewDownloadItem:   make(chan bool),
		chNewUploadItem:     make(chan bool),
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

func (fm *Controller) startDownloadQueue() {
	for {
		if fm.NetworkStatus == domain.NetworkDisconnected || fm.NetworkStatus == domain.NetworkConnecting {
			time.Sleep(100 * time.Millisecond)
		}
		fm.DownloadQueueStarted = true
		select {
		case <-fm.chStopDownloader:
			fm.mxDown.Lock()
			for _, v := range fm.DownloadQueue {
				v.Stop()
			}
			fm.mxDown.Unlock()
			fm.DownloadQueueStarted = false
			return
		case <-fm.chNewDownloadItem:
			fm.mxDown.Lock()
			for _, v := range fm.DownloadQueue {
				go v.StartDownload(fm)
			}
			fm.mxDown.Unlock()
		}
	}
}

func (fm *Controller) startUploadQueue() {
	for {
		if fm.NetworkStatus == domain.NetworkDisconnected || fm.NetworkStatus == domain.NetworkConnecting {
			time.Sleep(100 * time.Millisecond)
		}

		fm.UploadQueueStarted = true
		select {
		case <-fm.chStopUploader:
			fm.mxUp.Lock()
			for _, v := range fm.UploadQueue {
				v.Stop()
			}
			fm.mxUp.Unlock()
			fm.UploadQueueStarted = false
			return
		case <-fm.chNewUploadItem:
			fm.mxUp.Lock()
			for _, v := range fm.UploadQueue {
				go v.StartUpload(fm)
			}
			fm.mxUp.Unlock()

		}
	}
}

// downloadRequest send request to server
func (fm *Controller) downloadRequest(req *msg.MessageEnvelope, fs *File, partIdx int64) {
	// time out has been set in Send()
	res, err := fm.Send(req)
	if err == nil {
		switch res.Constructor {
		case msg.C_Error:
			// remove download from queue
			fm.DeleteFromQueue(fs.MessageID, domain.RequestStatusError)
			logs.Warn("downloadRequest() received error response and removed item from queue", zap.Int64("MsgID", fs.MessageID))
			fs.RequestStatus = domain.RequestStatusError
			repo.Files.UpdateFileStatus(fs.MessageID, fs.RequestStatus)

			if fm.onDownloadError != nil {
				fm.onDownloadError(fs.MessageID, int64(req.RequestID), fs.FilePath, res.Message)
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
				fm.downloadCompleted(fs.MessageID, fs.FilePath, fs.Type)

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
		fm.DeleteFromQueue(fs.MessageID, domain.RequestStatusError)
		logs.Warn("downloadRequest() download request errors passed retry threshold", zap.Int64("MsgID", fs.MessageID))
		fs.RequestStatus = domain.RequestStatusError
		_ = repo.Files.UpdateFileStatus(fs.MessageID, fs.RequestStatus)
		if fm.onDownloadError != nil {
			x := new(msg.Error)
			x.Code = "00"
			x.Items = "download request errors passed retry threshold"
			xbuff, _ := x.Marshal()
			fm.onDownloadError(fs.MessageID, int64(req.RequestID), fs.FilePath, xbuff)
		}
	}
}

func (fm *Controller) downloadCompleted(msgID int64, filePath string, stateType domain.FileStateType) {
	// delete file status
	fm.DeleteFromQueue(msgID, domain.RequestStatusCompleted)
	repo.Files.DeleteFileStatus(msgID)
	if fm.onDownloadCompleted != nil {
		fm.onDownloadCompleted(msgID, filePath, stateType)
	}
}

// uploadRequest send request to server
func (fm *Controller) uploadRequest(req *msg.MessageEnvelope, count int64, fs *File, partIdx int64) {
	// time out has been set in Send()
	res, err := fm.Send(req)
	if err == nil {
		switch res.Constructor {
		case msg.C_Error:
			// remove upload from
			fm.DeleteFromQueue(fs.MessageID, domain.RequestStatusError)
			logs.Warn("uploadRequest() received error response and removed item from queue", zap.Int64("MsgID", fs.MessageID))
			fs.RequestStatus = domain.RequestStatusError
			repo.Files.UpdateFileStatus(fs.MessageID, fs.RequestStatus)
			if fm.onUploadError != nil {
				fm.onUploadError(fs.MessageID, int64(req.RequestID), fs.FilePath, res.Message)
			}
		case msg.C_Bool:
			x := new(msg.Bool)
			err := x.Unmarshal(res.Message)
			if err != nil {
				logs.Error("uploadRequest()->Unmarshal()", zap.Error(err))
			}
			// reset counter
			fs.retryCounter = 0
			if x.Result {
				isThumbnail := fs.ThumbPosition < fs.ThumbTotalSize
				if isThumbnail {
					fs.ReadCommit(count, true, 0)
				} else {
					select {
					case fs.chUploadProgress <- partIdx:
					default:
						// progress monitor is exited already
					}
				}
			}
		default:
			// increase counter
			fs.retryCounter++
			logs.Warn("uploadRequest() received unknown response", zap.Error(err))

		}
	} else {
		// increase counter
		fs.retryCounter++
		logs.Warn("uploadRequest()", zap.Error(err))
	}

	if fs.retryCounter > domain.FileMaxRetry {

		// remove upload from queue
		fm.DeleteFromQueue(fs.MessageID, domain.RequestStatusError)
		logs.Error("uploadRequest() upload request errors passed retry threshold", zap.Int64("MsgID", fs.MessageID))
		fs.RequestStatus = domain.RequestStatusError
		_ = repo.Files.UpdateFileStatus(fs.MessageID, fs.RequestStatus)
		if fm.onUploadError != nil {
			x := new(msg.Error)
			x.Code = "00"
			x.Items = "upload request errors passed retry threshold"
			xbuff, _ := x.Marshal()
			fm.onUploadError(fs.MessageID, int64(req.RequestID), fs.FilePath, xbuff)
		}
	}
}

func (fm *Controller) uploadCompleted(msgID, fileID, targetID int64,
	clusterID int32, totalParts int64,
	stateType domain.FileStateType,
	filePath string,
	uploadRequest *msg.ClientSendMessageMedia,
	thumbFileID int64,
	thumbTotalParts int32,
) {
	// delete file status
	fm.DeleteFromQueue(msgID, domain.RequestStatusCompleted)
	repo.Files.DeleteFileStatus(msgID)
	if fm.onUploadCompleted != nil {
		fm.onUploadCompleted(msgID, fileID, targetID, clusterID, totalParts, stateType, filePath, uploadRequest, thumbFileID, thumbTotalParts)
	}
}

// Stop set stop flag
func (fm *Controller) Stop() {
	logs.Debug("FileController Stopping")

	if fm.UploadQueueStarted {
		fm.chStopUploader <- true
	}
	if fm.DownloadQueueStarted {
		fm.chStopDownloader <- true
	}

	if ctx != nil {
		ctx = nil
	}
	logs.Debug("FileController Stopped")
}

// Upload file to server
func (fm *Controller) Upload(fileID int64, req *msg.ClientPendingMessage) error {
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

	state := NewFile(req.ID, fileID, 0, fileSize, x.FilePath, domain.FileStateUpload, 0, 0, 0, fm.progressCallback)
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

	fm.AddToQueue(state)
	err = repo.Files.SaveFileStatus(state.GetDTO())
	return err
}

// Download add download request
func (fm *Controller) Download(req *msg.UserMessage) {

	var state *File
	dtoState, err := repo.Files.GetFileStatus(req.ID)

	if err == nil && dtoState != nil {
		if dtoState.IsCompleted {
			fm.downloadCompleted(dtoState.MessageID, dtoState.FilePath, domain.FileStateType(dtoState.Type))
			return
		}
		state = new(File)
		state.LoadDTO(*dtoState, fm.progressCallback)

	} else {
		var docID int64
		var clusterID int32
		var accessHash uint64
		var version int32
		var fileSize int32
		var filePath string
		switch req.MediaType {
		case msg.MediaTypeEmpty:
			// TODO:: implement it
		case msg.MediaTypePhoto:
			// TODO:: implement it
		case msg.MediaTypeDocument:
			x := new(msg.MediaDocument)
			x.Unmarshal(req.Media)

			fileName := ""
			for _, attr := range x.Doc.Attributes {
				if attr.Type == msg.AttributeTypeFile {
					attrFile := new(msg.DocumentAttributeFile)
					err := attrFile.Unmarshal(attr.Data)
					if err == nil {
						fileName = attrFile.Filename
					}
				}
			}

			docID = x.Doc.ID
			clusterID = x.Doc.ClusterID
			accessHash = x.Doc.AccessHash
			version = x.Doc.Version
			fileSize = x.Doc.FileSize
			filePath = GetFilePath(x.Doc.MimeType, x.Doc.ID, fileName)
			state = NewFile(req.ID, docID, 0, int64(fileSize), filePath, domain.FileStateDownload, clusterID, accessHash, version, fm.progressCallback)
			state.DownloadRequest = x.Doc

		case msg.MediaTypeContact:
			// TODO:: implement it
		default:
			logs.Error("Download() Invalid SharedMediaType")
		}
	}

	if state != nil {
		state.RequestStatus = domain.RequestStatusInProgress
		fm.AddToQueue(state)
		repo.Files.SaveFileStatus(state.GetDTO())
		repo.Files.SaveDownloadingFile(state.GetDTO())
		repo.Files.UpdateFileStatus(state.MessageID, domain.RequestStatusInProgress)
	}
}

// AddToQueue add request to queue
func (fm *Controller) AddToQueue(status *File) {
	switch status.Type {
	case domain.FileStateUpload, domain.FileStateUploadAccountPhoto, domain.FileStateUploadGroupPhoto:
		fm.mxUp.Lock()
		_, ok := fm.UploadQueue[status.MessageID]
		if !ok {
			fm.UploadQueue[status.MessageID] = status
		}
		fm.mxUp.Unlock()
		if !ok {
			fm.chNewUploadItem <- true
		}
	case domain.FileStateDownload:
		fm.mxDown.Lock()
		_, ok := fm.DownloadQueue[status.MessageID]
		if !ok {
			fm.DownloadQueue[status.MessageID] = status
		}
		fm.mxDown.Unlock()
		if !ok {
			fm.chNewDownloadItem <- true
		}
	}
}

// DeleteFromQueue remove items from download/upload queue and stop them
func (fm *Controller) DeleteFromQueue(msgID int64, status domain.RequestStatus) {
	fm.mxUp.Lock()
	up, uok := fm.UploadQueue[msgID]
	if uok {
		up.RequestStatus = status
		delete(fm.UploadQueue, msgID)
		up.Stop()
	}
	fm.mxUp.Unlock()

	fm.mxDown.Lock()
	down, dok := fm.DownloadQueue[msgID]
	if dok {
		down.RequestStatus = status
		delete(fm.DownloadQueue, msgID)
		down.Stop()
	}
	fm.mxDown.Unlock()
}

// CalculateMD5 this will calculate md5 hash for files that are smaller than threshold
func (fm *Controller) CalculateMD5(file *os.File) (string, error) {
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
func (fm *Controller) SetAuthorization(authID int64, authKey []byte) {
	fm.authKey = make([]byte, len(authKey))
	fm.authID = authID
	copy(fm.authKey, authKey)
}

// LoadQueueFromDB load in progress request from database
func (fm *Controller) LoadQueueFromDB() {
	// Load pended file status
	dtos := repo.Files.GetAllFileStatus()
	for _, d := range dtos {
		fs := new(File)
		fs.LoadDTO(d, fm.progressCallback)
		if fs.RequestStatus == domain.RequestStatusPaused ||
			fs.RequestStatus == domain.RequestStatusCanceled ||
			fs.RequestStatus == domain.RequestStatusCompleted {
			continue
		}
		fm.AddToQueue(fs)
	}
}

// SetNetworkStatus called on network controller state changes to inform file controller
func (fm *Controller) SetNetworkStatus(state domain.NetworkStatus) {
	fm.NetworkStatus = state
	if state == domain.NetworkWeak || state == domain.NetworkSlow || state == domain.NetworkFast {
		fm.LoadQueueFromDB()
	}
}

// DownloadAccountPhoto download account photo from server its sync
func (fm *Controller) DownloadAccountPhoto(userID int64, photo *msg.UserPhoto, isBig bool) (string, error) {
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
	res, err := fm.Send(envelop)
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
			return filePath, repo.Users.UpdateAccountPhotoPath(userID, photo.PhotoID, isBig, filePath)

		default:
			return "", fmt.Errorf("received unknown response constructor {UserId : %d}", userID)
		}
	}
	return "", err
}

// DownloadGroupPhoto download group photo from server its sync
func (fm *Controller) DownloadGroupPhoto(groupID int64, photo *msg.GroupPhoto, isBig bool) (string, error) {
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
	res, err := fm.Send(envelop)
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
			return filePath, repo.Groups.UpdatePhotoPath(groupID, isBig, filePath)

		default:
			return "", fmt.Errorf("received unknown response constructor {GroupID : %d}", groupID)
		}
	}
	return "", err
}

// DownloadThumbnail download thumbnail from server its sync
func (fm *Controller) DownloadThumbnail(msgID int64, fileID int64, accessHash uint64, clusterID, version int32) (string, error) {
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

	filePath := path.Join(dirCache, fmt.Sprintf("%d%s", fileID, ".jpg"))
	res, err := fm.Send(envelop)
	if err == nil {
		switch res.Constructor {
		case msg.C_Error:
			strErr := ""
			x := new(msg.Error)
			if err := x.Unmarshal(res.Message); err == nil {
				strErr = "Code :" + x.Code + ", Items :" + x.Items
			}
			return "", fmt.Errorf("DownloadThumbnail() received error response {MsgID: %d,  %s }", msgID, strErr)
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
			return filePath, repo.Files.UpdateThumbnailPath(msgID, filePath)

		default:
			return "", fmt.Errorf("DownloadThumbnail() received unknown response constructor {GroupID : %d}", msgID)
		}
	}
	return "", err
}

func (fm *Controller) ClearFiles(filePaths []string) error {
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

// GetFilePath generate related file path by its mime type
func GetFilePath(mimeType string, docID int64, fileName string) string {
	lower := strings.ToLower(mimeType)
	strDocID := strconv.FormatInt(docID, 10)
	ext := path.Ext(fileName)
	if ext == "" {
		exts, err := mime.ExtensionsByType(mimeType)
		if err == nil {
			for _, val := range exts {
				ext = val
			}
		}
	}

	// if the file is opus type,
	// means its voice file so it should be saved in cache folder
	// so user could not access to it by file manager
	if lower == "audio/ogg" {
		ext = ".ogg"
		return path.Join(dirCache, fmt.Sprintf("%s%s", strDocID, ext))
	}

	if strings.HasPrefix(lower, "video/") {
		return path.Join(dirVideo, fmt.Sprintf("%s%s", strDocID, ext))
	}
	if strings.HasPrefix(lower, "audio/") {
		return path.Join(dirAudio, fmt.Sprintf("%s%s", strDocID, ext))
	}
	if strings.HasPrefix(lower, "image/") {
		return path.Join(dirPhoto, fmt.Sprintf("%s%s", strDocID, ext))
	}

	return path.Join(dirFile, fmt.Sprintf("%s%s", strDocID, ext))
}

func GetAccountAvatarPath(userID int64, fileID int64) string {
	return path.Join(dirCache, fmt.Sprintf("u%d_%d%s", userID, fileID, ".jpg"))
}

func GetGroupAvatarPath(groupID int64, fileID int64) string {
	return path.Join(dirCache, fmt.Sprintf("g%d_%d%s", groupID, fileID, ".jpg"))
}