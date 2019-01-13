package filemanager

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
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
	_Clusters  map[int32]*msg.Cluster
	_DirAudio  string
	_DirFile   string
	_DirPhoto  string
	_DirVideo  string
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
}

func Ctx() *FileManager {
	if ctx == nil {
		panic("FileManager::Ctx() file manager not initialized !")
	}
	return ctx
}

func InitFileManager(onUploadCompleted domain.OnFileUploadCompleted, progressCallback domain.OnFileStatusChanged, onDownloadCompleted domain.OnFileDownloadCompleted) {
	if _Clusters == nil {
		_Clusters = make(map[int32]*msg.Cluster)
	}

	if ctx == nil {
		singletone.Lock()
		defer singletone.Unlock()
		if ctx == nil {
			ctx = &FileManager{
				UploadQueue:         make(map[int64]*FileStatus, 0),
				DownloadQueue:       make(map[int64]*FileStatus, 0),
				chStopUploader:      make(chan bool),
				chStopDownloader:    make(chan bool),
				onUploadCompleted:   onUploadCompleted,
				progressCallback:    progressCallback,
				onDownloadCompleted: onDownloadCompleted,
			}

		}
	}

	go ctx.startDownloadQueue()
	go ctx.startUploadQueue()

}
func SetRootFolders(audioDir, fileDir, photoDir, videoDir string) {
	_DirAudio = audioDir
	_DirFile = fileDir
	_DirPhoto = photoDir
	_DirVideo = videoDir
}

func SetAvailableClusters(clusters []*msg.Cluster) {
	// double check
	if _Clusters == nil {
		_Clusters = make(map[int32]*msg.Cluster)
	}
	for _, c := range clusters {
		_Clusters[c.ID] = c
	}
}

func GetAvailableClusters() map[int32]*msg.Cluster {
	return _Clusters
}

func GetBestCluster() *msg.Cluster {

	for {
		if len(_Clusters) > 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	// TODO : fix this later
	// get first cluster
	for _, c := range _Clusters {
		return c
	}

	return nil
}

func GetFilePath(mime string, docID int64, fileName string) string {
	lower := strings.ToLower(mime)
	strDocID := strconv.FormatInt(docID, 10)
	ext := path.Ext(fileName)
	if strings.HasPrefix(lower, "video/") {
		return path.Join(_DirVideo, fmt.Sprintf("%s%s", strDocID, ext))
	}
	if strings.HasPrefix(lower, "audio/") {
		return path.Join(_DirVideo, fmt.Sprintf("%s%s", strDocID, ext))
	}
	if strings.HasPrefix(lower, "image/") {
		return path.Join(_DirVideo, fmt.Sprintf("%s%s", strDocID, ext))
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

	cluster := GetBestCluster()
	state := NewFileStatus(req.ID, fileID, fileSize, x.FilePath, StateUpload, cluster.ID, 0, 0, fm.progressCallback)
	state.UploadRequest = x
	fm.AddToQueue(state)
	repo.Ctx().Files.SaveFileStatus(state.GetDTO())
	return nil
}

// Download add download request
func (fm *FileManager) Download(req *msg.UserMessage) {
	var docID int64
	var clusterID int32
	var accessHash uint64
	var version int32
	var fileSize int32
	var filePath string
	var state *FileStatus
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
		docID = x.Doc.ID
		clusterID = x.Doc.ClusterID
		accessHash = x.Doc.AccessHash
		version = x.Doc.Version
		fileSize = x.Doc.FileSize
		filePath = GetFilePath(x.Doc.MimeType, x.Doc.ID, "")
		state = NewFileStatus(req.ID, docID, int64(fileSize), filePath, StateDownload, clusterID, accessHash, version, fm.progressCallback)
		state.DownloadRequest = x.Doc

	case msg.MediaTypeContact:
		// TODO:: implement it
	default:
		log.LOG_Error("FileManager::Download() Invalid MediaType")
	}
	if state != nil {
		fm.AddToQueue(state)
		repo.Ctx().Files.SaveFileStatus(state.GetDTO())
		repo.Ctx().Files.SaveDownloadingFile(state.GetDTO())
	}
}

// AddToQueue add request to queue
func (fm *FileManager) AddToQueue(status *FileStatus) {
	if status.Type == StateUpload {
		fm.mxUp.Lock()
		fm.UploadQueue[status.FileID] = status
		fm.mxUp.Unlock()
	} else {
		fm.mxDown.Lock()
		fm.DownloadQueue[status.FileID] = status
		fm.mxDown.Unlock()
	}
}

func (fm *FileManager) DeleteFromQueue(messageID int64) {
	fm.mxUp.Lock()
	delete(fm.UploadQueue, messageID)
	fm.mxUp.Unlock()

	fm.mxDown.Lock()
	delete(fm.DownloadQueue, messageID)
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
	if GetBestCluster() == nil {
		time.Sleep(100 * time.Millisecond)
	}
	for {
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
				completedDownloads := domain.MInt64B{}
				for _, fs := range fm.DownloadQueue {
					if fs.IsCompleted {
						completedDownloads[fs.FileID] = true
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
				//remove completed downloads from queue
				if len(completedDownloads) > 0 {
					for key := range completedDownloads {
						delete(fm.DownloadQueue, key)
					}
					go repo.Ctx().Files.DeleteManyFileStatus(completedDownloads.ToArray())
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
	if GetBestCluster() == nil {
		time.Sleep(100 * time.Millisecond)
	}
	for {
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
				for _, fs := range fm.UploadQueue {
					if fs.IsCompleted {
						continue
					}

					envelop, readCount, err := fs.ReadAsFileSavePart()
					if err != nil {
						log.LOG_Error("FileManager::startUploadQueue()", zap.Error(err), zap.String("filePath", fs.FilePath))
						continue
					}
					wg.Add(1)
					go fm.sendUploadRequest(envelop, int64(readCount), fs, wg)
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
	res, err := fm.Send(req, _Clusters[fs.ClusterID])
	if err == nil {
		switch res.Constructor {
		case msg.C_Error:
			x := new(msg.Error)
			x.Unmarshal(res.Message)
			log.LOG_Error("sendUploadRequest() received Error response", zap.String("Code", x.Code), zap.String("Item", x.Items))
		case msg.C_Bool:
			x := new(msg.Bool)
			err := x.Unmarshal(res.Message)
			if err != nil {
				log.LOG_Error("sendUploadRequest() failed to unmarshal C_Bool", zap.Error(err))
			}
			if x.Result {
				isCompleted := fs.ReadCommit(count)
				if isCompleted {
					// TODO : call completed delegate to execute SendMessageMedia
					if fm.onUploadCompleted != nil {
						go fm.onUploadCompleted(fs.MessageID, fs.FileID, fs.ClusterID, fs.TotalParts, fs.UploadRequest)
					}
				}
			}
		default:
			log.LOG_Error("sendUploadRequest() received unknown response", zap.Error(err))
		}
	} else {
		log.LOG_Error("sendUploadRequest()", zap.Error(err))
	}
	wg.Done()
}

func (fm *FileManager) sendDownloadRequest(req *msg.MessageEnvelope, fs *FileStatus, wg *sync.WaitGroup) {
	// time out has been set in Send()
	res, err := fm.Send(req, _Clusters[fs.ClusterID])
	if err == nil {
		switch res.Constructor {
		case msg.C_Error:
			x := new(msg.Error)
			x.Unmarshal(res.Message)
			log.LOG_Error("sendDownloadRequest() received Error response", zap.String("Code", x.Code), zap.String("Item", x.Items))
		case msg.C_File:
			x := new(msg.File)
			err := x.Unmarshal(res.Message)
			if err != nil {
				log.LOG_Error("sendDownloadRequest() failed to unmarshal C_File", zap.Error(err))
			}
			isCompleted, err := fs.Write(x.Bytes)
			if err != nil {
				log.LOG_Error("sendDownloadRequest() failed write to file", zap.Error(err))
			} else if isCompleted {
				//call completed delegate
				if fm.onDownloadCompleted != nil {
					go fm.onDownloadCompleted(fs.MessageID, fs.FilePath)
				}
			}
		default:
			log.LOG_Error("sendDownloadRequest() received unknown response", zap.Error(err))
		}
	} else {
		log.LOG_Error("sendUploadRequest()", zap.Error(err))
	}
	wg.Done()
}

func (fm *FileManager) LoadFileStatusQueue() {
	//TODO : load pended file status
	dtos := repo.Ctx().Files.GetAllFileStatus()
	for _, d := range dtos {
		fs := new(FileStatus)
		fs.LoadDTO(d, fm.progressCallback)
		fm.AddToQueue(fs)
	}
}
