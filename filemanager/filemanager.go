package filemanager

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/repo"

	"go.uber.org/zap"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/log"
	"git.ronaksoftware.com/ronak/riversdk/msg"
)

var (
	singletone sync.Mutex
	ctx        *FileManager
	_DirAudio  string
	_DirFile   string
	_DirPhoto  string
	_DirVideo  string
	_DirCache  string
)

const (
	// FileSizeThresholdForCheckHash for files thatare smaller than  this number we will calculate md5 hash to do not reupload same file twice
	FileSizeThresholdForCheckHash = 10 * 1024 * 1024 // 10MB

)

// FileManager manages files status and cache
type FileManager struct {
	authKey    []byte
	authID     int64
	messageSeq int64

	ServerAddress string
	NetworkStatus domain.NetworkStatus

	mxDown               sync.Mutex
	mxUp                 sync.Mutex
	UploadQueue          map[int64]*FileStatus
	DownloadQueue        map[int64]*FileStatus
	UploadQueueStarted   bool
	DownloadQueueStarted bool

	chStopUploader      chan bool
	chStopDownloader    chan bool
	onUploadCompleted   domain.OnFileUploadCompleted
	onDownloadCompleted domain.OnFileDownloadCompleted
	progressCallback    domain.OnFileStatusChanged
	onUploadError       domain.OnFileUploadError
	onDownloadError     domain.OnFileDownloadError
}

func Ctx() *FileManager {
	if ctx == nil {
		panic("FileManager::Ctx() file manager not initialized !")
	}
	return ctx
}

func InitFileManager(serverAddress string,
	onUploadCompleted domain.OnFileUploadCompleted,
	progressCallback domain.OnFileStatusChanged,
	onDownloadCompleted domain.OnFileDownloadCompleted,
	onFileUploadError domain.OnFileUploadError,
	onFileDownloadError domain.OnFileDownloadError,
) {

	if ctx == nil {
		singletone.Lock()
		defer singletone.Unlock()
		if ctx == nil {
			ctx = &FileManager{
				ServerAddress:       serverAddress,
				UploadQueue:         make(map[int64]*FileStatus, 0),
				DownloadQueue:       make(map[int64]*FileStatus, 0),
				chStopUploader:      make(chan bool),
				chStopDownloader:    make(chan bool),
				onUploadCompleted:   onUploadCompleted,
				progressCallback:    progressCallback,
				onDownloadCompleted: onDownloadCompleted,
				onUploadError:       onFileUploadError,
				onDownloadError:     onFileDownloadError,
			}

		}
	}

	go ctx.startDownloadQueue()
	go ctx.startUploadQueue()

}
func SetRootFolders(audioDir, fileDir, photoDir, videoDir, cacheDir string) {
	_DirAudio = audioDir
	_DirFile = fileDir
	_DirPhoto = photoDir
	_DirVideo = videoDir
	_DirCache = cacheDir
}

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
		return path.Join(_DirCache, fmt.Sprintf("%s%s", strDocID, ext))
	}

	if strings.HasPrefix(lower, "video/") {
		return path.Join(_DirVideo, fmt.Sprintf("%s%s", strDocID, ext))
	}
	if strings.HasPrefix(lower, "audio/") {
		return path.Join(_DirAudio, fmt.Sprintf("%s%s", strDocID, ext))
	}
	if strings.HasPrefix(lower, "image/") {
		return path.Join(_DirPhoto, fmt.Sprintf("%s%s", strDocID, ext))
	}

	return path.Join(_DirFile, fmt.Sprintf("%s%s", strDocID, ext))
}

func (fm *FileManager) Stop() {
	if fm.UploadQueueStarted {
		fm.chStopUploader <- true
	}
	if fm.DownloadQueueStarted {
		fm.chStopDownloader <- true
	}
}

// Upload file to server
func (fm *FileManager) Upload(fileID int64, req *msg.ClientPendingMessage) error {
	x := new(msg.ClientSendMessageMedia)
	x.Unmarshal(req.Media)

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
	// strMD5, err := fm.CalculateMD5(file)
	// if err == nil {
	// 	// TODO : check DB with file md5 hash and meeeeehhhhh :/
	// 	log.LOG_Debug(strMD5)
	// }

	state := NewFileStatus(req.ID, fileID, 0, fileSize, x.FilePath, domain.FileStateUpload, 0, 0, 0, fm.progressCallback)
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
			state.ThumbTotalParts = CalculatePartsCount(thumbInfo.Size())
		}
	}

	fm.AddToQueue(state)
	repo.Ctx().Files.SaveFileStatus(state.GetDTO())
	return nil
}

// Download add download request
func (fm *FileManager) Download(req *msg.UserMessage) {

	var state *FileStatus
	dtoState, err := repo.Ctx().Files.GetFileStatus(req.ID)

	if err == nil && dtoState != nil {
		if dtoState.IsCompleted {
			fm.downloadCompleted(dtoState.MessageID, dtoState.FilePath, domain.FileStateType(dtoState.Type))
			return
		}
		state = new(FileStatus)
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
			// // TODO:: implement it
			// x := new(msg.MediaPhoto)
			// x.Unmarshal(req.Media)
			// docID = x.Doc.ID
			// clusterID = x.Doc.ClusterID
			// accessHash = x.Doc.AccessHash
			// fileSize = x.Doc.FileSize
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
			state = NewFileStatus(req.ID, docID, 0, int64(fileSize), filePath, domain.FileStateDownload, clusterID, accessHash, version, fm.progressCallback)
			state.DownloadRequest = x.Doc

		case msg.MediaTypeContact:
			// TODO:: implement it
		default:
			log.LOG_Error("FileManager::Download() Invalid MediaType")
		}
	}

	if state != nil {
		fm.AddToQueue(state)
		repo.Ctx().Files.SaveFileStatus(state.GetDTO())
		repo.Ctx().Files.SaveDownloadingFile(state.GetDTO())
		repo.Ctx().Files.UpdateFileStatus(state.MessageID, domain.RequestStateInProgress)
	}
}

// AddToQueue add request to queue
func (fm *FileManager) AddToQueue(status *FileStatus) {
	if status.Type == domain.FileStateUpload || status.Type == domain.FileStateUploadAccountPhoto || status.Type == domain.FileStateUploadGroupPhoto {
		fm.mxUp.Lock()
		fm.UploadQueue[status.MessageID] = status
		fm.mxUp.Unlock()
	} else if status.Type == domain.FileStateDownload {
		fm.mxDown.Lock()
		fm.DownloadQueue[status.MessageID] = status
		fm.mxDown.Unlock()
	}
}

func (fm *FileManager) DeleteFromQueue(msgID int64) {
	fm.mxUp.Lock()
	delete(fm.UploadQueue, msgID)
	fm.mxUp.Unlock()

	fm.mxDown.Lock()
	delete(fm.DownloadQueue, msgID)
	fm.mxDown.Unlock()
}

// CalculateMD5 this will calculate md5 hash for files that are smaller than threshold
func (fm *FileManager) CalculateMD5(file *os.File) (string, error) {
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

func (fm *FileManager) SetAuthorization(authID int64, authKey []byte) {
	fm.authKey = make([]byte, len(authKey))
	fm.authID = authID
	copy(fm.authKey, authKey)
}

func (fm *FileManager) startDownloadQueue() {
	for {
		if fm.NetworkStatus == domain.DISCONNECTED || fm.NetworkStatus == domain.CONNECTING {
			time.Sleep(100 * time.Millisecond)
		}

		fm.DownloadQueueStarted = true
		wg := &sync.WaitGroup{}
		select {
		case <-fm.chStopDownloader:
			// wait for running request gracefuly end
			wg.Wait()
			fm.DownloadQueueStarted = false
			return
		default:
			// round robin and wait till all items in queus moves on step forward
			fm.mxDown.Lock()
			if len(fm.DownloadQueue) > 0 {
				completedItems := domain.MInt64B{}
				for _, fs := range fm.DownloadQueue {
					if fs.IsCompleted {
						// panic("completed DOWNLOAD exist in download queue")
						completedItems[fs.MessageID] = true
						continue
					}

					envelop, err := fs.ReadAsFileGet()
					if err != nil {
						log.LOG_Error("FileManager::startDownloadQueue()", zap.Error(err), zap.String("filePath", fs.FilePath))
						continue
					}

					wg.Add(1)
					go fm.sendDownloadRequest(envelop, fs, wg)
				}

				//remove completed downloads from queue, other states will be handled in place when its accure
				if len(completedItems) > 0 {
					for key := range completedItems {
						delete(fm.DownloadQueue, key)
					}
					go repo.Ctx().Files.DeleteManyFileStatus(completedItems.ToArray())
				}

				fm.mxDown.Unlock()
			} else {
				fm.mxDown.Unlock()
				time.Sleep(300 * time.Millisecond)
			}

			wg.Wait()

		}
	}
}

func (fm *FileManager) startUploadQueue() {
	for {
		if fm.NetworkStatus == domain.DISCONNECTED || fm.NetworkStatus == domain.CONNECTING {
			time.Sleep(100 * time.Millisecond)
		}

		fm.UploadQueueStarted = true
		wg := &sync.WaitGroup{}
		select {
		case <-fm.chStopUploader:
			// wait for running request gracefuly end
			wg.Wait()
			fm.UploadQueueStarted = false
			return
		default:
			// round robin and wait till all items in queus moves on step forward
			fm.mxUp.Lock()
			if len(fm.UploadQueue) > 0 {
				completedItems := domain.MInt64B{}
				for _, fs := range fm.UploadQueue {
					if fs.IsCompleted {
						// panic("completed UPLOAD exist in upload queue")
						completedItems[fs.MessageID] = true
						continue
					}
					isThumbnail := fs.ThumbPosition < fs.ThumbTotalSize
					envelop, readCount, err := fs.ReadAsFileSavePart(isThumbnail)
					if err != nil {
						log.LOG_Error("FileManager::startUploadQueue()", zap.Error(err), zap.String("filePath", fs.FilePath))
						continue
					}
					wg.Add(1)
					go fm.sendUploadRequest(envelop, int64(readCount), fs, wg)
				}

				//remove completed items from queue, other states will be handled in place when its accure
				if len(completedItems) > 0 {
					for key := range completedItems {
						delete(fm.UploadQueue, key)
					}
					go repo.Ctx().Files.DeleteManyFileStatus(completedItems.ToArray())
				}

				fm.mxUp.Unlock()
			} else {
				fm.mxUp.Unlock()
				time.Sleep(300 * time.Millisecond)
			}

			wg.Wait()

		}
	}
}

func (fm *FileManager) sendUploadRequest(req *msg.MessageEnvelope, count int64, fs *FileStatus, wg *sync.WaitGroup) {
	// time out has been set in Send()
	res, err := fm.Send(req)
	if err == nil {
		switch res.Constructor {
		case msg.C_Error:
			// remove upload from
			fm.DeleteFromQueue(fs.MessageID)
			log.LOG_Error("sendUploadRequest() received error response and removed item from queue", zap.Int64("MsgID", fs.MessageID))
			fs.RequestStatus = domain.RequestStateError
			repo.Ctx().Files.UpdateFileStatus(fs.MessageID, fs.RequestStatus)
			if fm.onUploadError != nil {
				fm.onUploadError(fs.MessageID, int64(req.RequestID), fs.FilePath, res.Message)
			}
		case msg.C_Bool:
			x := new(msg.Bool)
			err := x.Unmarshal(res.Message)
			if err != nil {
				log.LOG_Error("sendUploadRequest() failed to unmarshal C_Bool", zap.Error(err))
			}
			// reset counter
			fs.retryCounter = 0

			if x.Result {
				isThumbnail := fs.ThumbPosition < fs.ThumbTotalSize
				isCompleted := fs.ReadCommit(count, isThumbnail)
				if isCompleted && !isThumbnail {
					//call completed delegate
					fm.uploadCompleted(fs.MessageID, fs.FileID, fs.TargetID, fs.ClusterID, fs.TotalParts, fs.Type, fs.FilePath, fs.UploadRequest, fs.ThumbFileID, fs.ThumbTotalParts)
				}
			}
		default:
			// increase counter
			fs.retryCounter++
			log.LOG_Error("sendUploadRequest() received unknown response", zap.Error(err))

		}
	} else {
		// increase counter
		fs.retryCounter++
		log.LOG_Error("sendUploadRequest()", zap.Error(err))
	}

	if fs.retryCounter > domain.FileRetryThreshold {

		// remove upload from queue
		fm.DeleteFromQueue(fs.MessageID)
		log.LOG_Error("sendUploadRequest() upload request errors passed retry threshold", zap.Int64("MsgID", fs.MessageID))
		fs.RequestStatus = domain.RequestStateError
		repo.Ctx().Files.UpdateFileStatus(fs.MessageID, fs.RequestStatus)
		if fm.onUploadError != nil {
			x := new(msg.Error)
			x.Code = "00"
			x.Items = "upload request errors passed retry threshold"
			xbuff, _ := x.Marshal()
			fm.onUploadError(fs.MessageID, int64(req.RequestID), fs.FilePath, xbuff)
		}
	}

	wg.Done()
}

func (fm *FileManager) sendDownloadRequest(req *msg.MessageEnvelope, fs *FileStatus, wg *sync.WaitGroup) {
	// time out has been set in Send()
	res, err := fm.Send(req)
	if err == nil {
		switch res.Constructor {
		case msg.C_Error:
			// remove download from queue
			fm.DeleteFromQueue(fs.MessageID)
			log.LOG_Error("sendDownloadRequest() received error response and removed item from queue", zap.Int64("MsgID", fs.MessageID))
			fs.RequestStatus = domain.RequestStateError
			repo.Ctx().Files.UpdateFileStatus(fs.MessageID, fs.RequestStatus)

			if fm.onDownloadError != nil {
				fm.onDownloadError(fs.MessageID, int64(req.RequestID), fs.FilePath, res.Message)
			}
		case msg.C_File:
			x := new(msg.File)
			err := x.Unmarshal(res.Message)
			if err != nil {
				log.LOG_Error("sendDownloadRequest() failed to unmarshal C_File", zap.Error(err))
				fs.retryCounter++
				break
			}

			if len(x.Bytes) == 0 {
				log.LOG_Error("sendDownloadRequest() Received 0 bytes from server ",
					zap.Int64("MsgID", fs.MessageID),
					zap.Int64("Position", fs.Position),
					zap.Int64("TotalSize", fs.TotalSize),
				)
				fs.retryCounter++
				break

			} else {
				// reset counter
				fs.retryCounter = 0
			}

			isCompleted, err := fs.Write(x.Bytes)
			if err != nil {
				log.LOG_Error("sendDownloadRequest() failed write to file", zap.Error(err))
			} else if isCompleted {
				//call completed delegate
				fm.downloadCompleted(fs.MessageID, fs.FilePath, fs.Type)

			}
		default:
			// increase counter
			fs.retryCounter++
			log.LOG_Error("sendDownloadRequest() received unknown response", zap.Error(err))
		}
	} else {
		// increase counter
		fs.retryCounter++
		log.LOG_Error("sendDownloadRequest()", zap.Error(err))
	}
	if fs.retryCounter > domain.FileRetryThreshold {
		// remove download from queue
		fm.DeleteFromQueue(fs.MessageID)
		log.LOG_Error("sendDownloadRequest() download request errors passed retry threshold", zap.Int64("MsgID", fs.MessageID))
		fs.RequestStatus = domain.RequestStateError
		repo.Ctx().Files.UpdateFileStatus(fs.MessageID, fs.RequestStatus)
		if fm.onDownloadError != nil {
			x := new(msg.Error)
			x.Code = "00"
			x.Items = "download request errors passed retry threshold"
			xbuff, _ := x.Marshal()
			fm.onDownloadError(fs.MessageID, int64(req.RequestID), fs.FilePath, xbuff)
		}
	}
	wg.Done()
}

func (fm *FileManager) LoadFileStatusQueue() {
	// Load pended file status
	dtos := repo.Ctx().Files.GetAllFileStatus()
	for _, d := range dtos {
		fs := new(FileStatus)
		fs.LoadDTO(d, fm.progressCallback)
		if fs.RequestStatus == domain.RequestStatePused ||
			fs.RequestStatus == domain.RequestStateCanceled ||
			fs.RequestStatus == domain.RequestStateCompleted ||
			fs.RequestStatus == domain.RequestStateError {
			continue
		}

		fm.AddToQueue(fs)
	}
}

// SetNetworkStatus called on network controller state changes to inform filemanager
func (fm *FileManager) SetNetworkStatus(state domain.NetworkStatus) {
	fm.NetworkStatus = state
}

func (fm *FileManager) downloadCompleted(msgID int64, filePath string, stateType domain.FileStateType) {
	// delete file status
	fm.DeleteFromQueue(msgID)
	repo.Ctx().Files.DeleteFileStatus(msgID)
	if fm.onDownloadCompleted != nil {
		fm.onDownloadCompleted(msgID, filePath, stateType)
	}
}

func (fm *FileManager) uploadCompleted(msgID, fileID, targetID int64,
	clusterID, totalParts int32,
	stateType domain.FileStateType,
	filePath string,
	uploadRequest *msg.ClientSendMessageMedia,
	thumbFileID int64,
	thumbTotalParts int32,

) {
	// delete file status
	fm.DeleteFromQueue(msgID)
	repo.Ctx().Files.DeleteFileStatus(msgID)
	if fm.onUploadCompleted != nil {
		fm.onUploadCompleted(msgID, fileID, targetID, clusterID, totalParts, stateType, filePath, uploadRequest, thumbFileID, thumbTotalParts)
	}
}

func (fm *FileManager) DownloadAccountPhoto(userID int64, photo *msg.UserPhoto, isBig bool) (string, error) {
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

	req.Offset = 0
	req.Limit = 0

	envelop := new(msg.MessageEnvelope)
	envelop.Constructor = msg.C_FileGet
	envelop.Message, _ = req.Marshal()
	envelop.RequestID = uint64(domain.SequentialUniqueID())

	filePath := GetFilePath("image/jpeg", req.Location.FileID, "avatar.jpg")
	res, err := fm.Send(envelop)
	if err == nil {
		switch res.Constructor {
		case msg.C_Error:
			strErr := ""
			x := new(msg.Error)
			if err := x.Unmarshal(res.Message); err == nil {
				strErr = "Code :" + x.Code + ", Items :" + x.Items
			}
			return "", fmt.Errorf("received error response {UserID: %d,  %s }", userID, strErr)
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
			return filePath, repo.Ctx().Users.UpdateAccountPhotoPath(userID, photo.PhotoID, isBig, filePath)

		default:
			return "", fmt.Errorf("received unknown response constructor {UserId : %d}", userID)
		}
	}
	return "", err
}

func (fm *FileManager) DownloadGroupPhoto(groupID int64, photo *msg.GroupPhoto, isBig bool) (string, error) {
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

	req.Offset = 0
	req.Limit = 0

	envelop := new(msg.MessageEnvelope)
	envelop.Constructor = msg.C_FileGet
	envelop.Message, _ = req.Marshal()
	envelop.RequestID = uint64(domain.SequentialUniqueID())

	filePath := GetFilePath("image/jpeg", req.Location.FileID, "group_avatar.jpg")
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
			return filePath, repo.Ctx().Groups.UpdateGroupPhotoPath(groupID, isBig, filePath)

		default:
			return "", fmt.Errorf("received unknown response constructor {GroupID : %d}", groupID)
		}
	}
	return "", err
}

func (fm *FileManager) DownloadThumbnail(msgID int64, fileID int64, accessHash uint64, clusterID, version int32) (string, error) {

	req := new(msg.FileGet)
	req.Location = &msg.InputFileLocation{
		AccessHash: accessHash,
		ClusterID:  clusterID,
		FileID:     fileID,
		Version:    version,
	}
	req.Offset = 0
	req.Limit = 0

	envelop := new(msg.MessageEnvelope)
	envelop.Constructor = msg.C_FileGet
	envelop.Message, _ = req.Marshal()
	envelop.RequestID = uint64(domain.SequentialUniqueID())

	filePath := path.Join(_DirCache, fmt.Sprintf("%d%s", fileID, ".jpg"))
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
			return filePath, repo.Ctx().Files.UpdateThumbnailPath(msgID, filePath)

		default:
			return "", fmt.Errorf("DownloadThumbnail() received unknown response constructor {GroupID : %d}", msgID)
		}
	}
	return "", err
}
