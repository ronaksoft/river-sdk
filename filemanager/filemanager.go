package filemanager

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

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

	chStopUploader    chan bool
	chStopDownloader  chan bool
	onUploadCompleted domain.OnFileUploadCompleted
	progressCallback  domain.OnFileStatusChanged
}

func Ctx() *FileManager {
	if ctx == nil {
		panic("FileManager::Ctx() file manager not initialized !")
	}
	return ctx
}

func InitFileManager(onUploadCompleted domain.OnFileUploadCompleted, progressCallback domain.OnFileStatusChanged) {
	if _Clusters == nil {
		_Clusters = make(map[int32]*msg.Cluster)
	}

	if ctx == nil {
		singletone.Lock()
		defer singletone.Unlock()
		if ctx == nil {
			ctx = &FileManager{
				UploadQueue:       make(map[int64]*FileStatus, 0),
				DownloadQueue:     make(map[int64]*FileStatus, 0),
				chStopUploader:    make(chan bool),
				chStopDownloader:  make(chan bool),
				onUploadCompleted: onUploadCompleted,
				progressCallback:  progressCallback,
			}

		}
	}

	go ctx.startDownloadQueue()
	go ctx.startUploadQueue()

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

func GetBestCluster() *msg.Cluster {
	// TODO : fix this later
	// get first cluster
	for _, c := range _Clusters {
		return c
	}

	return nil
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
func (fm *FileManager) Upload(req *msg.ClientPendingMessage) error {
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
	strMD5, err := fm.CalculateMD5(file)
	if err == nil {
		// TODO : check DB with file md5 hash and meeeeehhhhh :/
		log.LOG_Debug(strMD5)
	}

	fileID := req.RequestID
	cluster := GetBestCluster()
	state := NewFileStatus(req.ID, fileID, fileSize, x.FilePath, StateUpload, cluster.ID, 0, fm.progressCallback)

	fm.AddToQueue(state)
	return nil
}

// Download add download request
func (fm *FileManager) Download(filePath string, req *msg.UserMessage) {
	var docID int64
	var clusterID int32
	var accessHash uint64
	var fileSize int32

	switch req.MediaType {
	case msg.MediaTypeEmpty:
		// TODO:: implement it
	case msg.MediaTypePhoto:
		// TODO:: implement it
	case msg.MediaTypeDocument:
		x := new(msg.Document)
		x.Unmarshal(req.Media)
		docID = x.ID
		clusterID = x.ClusterID
		accessHash = x.AccessHash
		fileSize = x.FileSize
	case msg.MediaTypeContact:
		// TODO:: implement it
	default:
	}

	state := NewFileStatus(req.ID, docID, int64(fileSize), filePath, StateDownload, clusterID, accessHash, fm.progressCallback)
	fm.AddToQueue(state)
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
	// for {
	// 	select {
	// 	case <-fm.chStopDownloader:
	// 		return
	// 	default:
	// 	}
	// }
}

func (fm *FileManager) startUploadQueue() {
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
			log.LOG_Error("sendUploadRequest() received Error response", zap.String("Code", x.Code), zap.String("Item", x.Code))
		case msg.C_Bool:
			x := new(msg.Bool)
			x.Unmarshal(res.Message)
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

func (fm *FileManager) LoadFileStatusQueue() {
	//TODO : load pended file status
	dtos := repo.Ctx().Files.GetAllFileStatus()
	for _, d := range dtos {
		fs := new(FileStatus)
		fs.LoadDTO(d, fm.progressCallback)
		fm.AddToQueue(fs)
	}
}
