package filemanager

import (
	"encoding/json"
	"os"
	"sync"

	"git.ronaksoftware.com/ronak/riversdk/repo/dto"

	"go.uber.org/zap"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/log"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/repo"
)

// FileStatus monitors file state
type FileStatus struct {
	mx         sync.Mutex
	MessageID  int64  `json:"MessageID"`
	FileID     int64  `json:"FileID"`
	TargetID   int64  `json:"TargetID"`
	ClusterID  int32  `json:"ClusterID"`
	AccessHash uint64 `json:"AccessHash"`
	Version    int32  `json:"Version"`
	FilePath   string `json:"FilePath"`

	TotalSize       int64                       `json:"TotalSize"`
	PartList        *QueueParts                 `json:"PartNo"`
	TotalParts      int64                       `json:"TotalParts"`
	Type            domain.FileStateType        `json:"StatusType"`
	IsCompleted     bool                        `json:"IsCompleted"`
	RequestStatus   domain.RequestStatus        `json:"RequestStatus"`
	UploadRequest   *msg.ClientSendMessageMedia `json:"UploadRequest"`
	DownloadRequest *msg.Document               `json:"DownloadRequest"`

	ThumbFileID     int64  `json:"ThumbFileID"`
	ThumbFilePath   string `json:"ThumbFilePath"`
	ThumbPosition   int64  `json:"ThumbPosition"`
	ThumbTotalSize  int64  `json:"ThumbTotalSize"`
	ThumbPartNo     int32  `json:"ThumbPartNo"`
	ThumbTotalParts int32  `json:"ThumbTotalParts"`

	onFileStatusChanged domain.OnFileStatusChanged
	// on receive unknown reposne or send error increase this counter to reach its threshold
	retryCounter int

	stop             bool
	started          bool
	chUploadProgress chan int64
}

// NewFileStatus create new instance
func NewFileStatus(messageID int64,
	fileID int64,
	targetID int64,
	totalSize int64,
	filePath string,
	stateType domain.FileStateType,
	clusterID int32,
	accessHash uint64,
	version int32,
	progress domain.OnFileStatusChanged) *FileStatus {

	fs := &FileStatus{
		MessageID:  messageID,
		FileID:     fileID,
		TargetID:   targetID,
		FilePath:   filePath,
		TotalSize:  totalSize,
		ClusterID:  clusterID,
		AccessHash: accessHash,
		Version:    version,

		PartList:            NewQueueParts(),
		TotalParts:          0,
		onFileStatusChanged: progress,
		Type:                stateType,
		RequestStatus:       domain.RequestStateInProgress,
	}

	// create partlist
	fs.TotalParts = CalculatePartsCount(totalSize)

	items := make([]int64, 0, fs.TotalParts)
	for i := int64(0); i < fs.TotalParts; i++ {
		items = append(items, i)
	}
	fs.PartList.PushMany(items)

	return fs
}

func CalculatePartsCount(fileSize int64) int64 {
	count := fileSize / domain.FilePayloadSize
	if (count * domain.FilePayloadSize) < fileSize {
		return count + 1
	}
	return count

}

// Read reads next required chunk of data
func (fs *FileStatus) Read(isThumbnail bool, partIdx int64) ([]byte, int, error) {
	fs.mx.Lock()

	filePath := fs.FilePath
	position := int64(partIdx * domain.FilePayloadSize)
	totalSize := fs.TotalSize

	if isThumbnail {
		filePath = fs.ThumbFilePath
		position = fs.ThumbPosition
		totalSize = fs.ThumbTotalSize
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, 0, err
	}
	var requiredBytes int64 = domain.FilePayloadSize
	if (position + domain.FilePayloadSize) > totalSize {
		requiredBytes = totalSize - position
	}
	buff := make([]byte, requiredBytes)
	readCount, err := file.ReadAt(buff, position)
	file.Close()
	if err != nil {
		return nil, 0, err
	}
	fs.mx.Unlock()

	return buff, readCount, nil
}

// Write writes givin data to current position of file
func (fs *FileStatus) Write(data []byte, partIdx int64) (isCompleted bool, err error) {
	fs.mx.Lock()
	defer fs.mx.Unlock()
	var file *os.File

	// create file if its not exist
	if _, err = os.Stat(fs.FilePath); os.IsNotExist(err) {
		file, err = os.Create(fs.FilePath)
		if err != nil {
			return
		}

		// truncate reserves size of file
		err = file.Truncate(fs.TotalSize)
		if err != nil {
			file.Close()
			return
		}
	}
	// open file if its not open
	if file == nil {
		file, err = os.OpenFile(fs.FilePath, os.O_RDWR, 0666)
		if err != nil {
			return
		}
	}

	// write to file
	position := partIdx * domain.FilePayloadSize
	_, err = file.WriteAt(data, int64(position))
	file.Close()
	if err != nil {
		return
	}
	//fs.Position += int64(count)

	count := fs.PartList.Length()
	fs.IsCompleted = count == 0

	if fs.IsCompleted {
		fs.RequestStatus = domain.RequestStateCompleted
	}
	isCompleted = fs.IsCompleted
	if isCompleted {
		repo.Ctx().Files.SaveDownloadingFile(fs.GetDTO())
	}
	fs.fileStatusChanged()

	return
}

//ReadCommit apply that last read process result was success and increase counter and progress
func (fs *FileStatus) ReadCommit(count int64, isThumbnail bool, partIdx int64) (isCompleted bool) {
	if isThumbnail {
		fs.ThumbPosition += count
		fs.ThumbPartNo++
		repo.Ctx().Files.SaveFileStatus(fs.GetDTO())
		return
	}

	partCount := fs.PartList.Length()
	fs.IsCompleted = partCount == 0

	if fs.IsCompleted {
		fs.RequestStatus = domain.RequestStateCompleted
	}
	fs.fileStatusChanged()
	return fs.IsCompleted
}

func (fs *FileStatus) fileStatusChanged() {
	// TODO : save file status to DB
	err := repo.Ctx().Files.SaveFileStatus(fs.GetDTO())
	if err != nil {
		log.LOG_Debug("fileStatusChanged() failed to save in DB", zap.Error(err))
	}

	lenParts := int64(fs.PartList.Length())

	processedParts := fs.TotalParts - lenParts
	if fs.onFileStatusChanged != nil {
		fs.onFileStatusChanged(fs.MessageID, processedParts, fs.TotalParts, fs.Type)
	}

}

func (fs *FileStatus) ReadAsFileSavePart(isThumbnail bool, partIdx int64) (envelop *msg.MessageEnvelope, readCount int, err error) {

	var buff []byte
	buff, readCount, err = fs.Read(isThumbnail, partIdx)
	if err != nil {
		return
	}
	req := new(msg.FileSavePart)
	req.Bytes = buff

	if isThumbnail {
		req.FileID = fs.ThumbFileID
		req.PartID = fs.ThumbPartNo + 1
		req.TotalParts = fs.ThumbTotalParts
	} else {
		req.FileID = fs.FileID
		req.PartID = int32(partIdx + 1)
		req.TotalParts = int32(fs.TotalParts)
	}

	envelop = new(msg.MessageEnvelope)
	envelop.Constructor = msg.C_FileSavePart
	envelop.Message, err = req.Marshal()
	envelop.RequestID = uint64(domain.SequentialUniqueID())

	return
}

func (fs *FileStatus) GetDTO() *dto.FileStatus {
	m := new(dto.FileStatus)

	m.MessageID = fs.MessageID
	m.FileID = fs.FileID
	m.ClusterID = fs.ClusterID
	m.AccessHash = int64(fs.AccessHash)
	m.Version = fs.Version
	m.FilePath = fs.FilePath
	//m.Position = fs.Position
	m.TotalSize = fs.TotalSize

	partList, _ := json.Marshal(fs.PartList.GetRawItems())
	m.PartList = partList

	m.TotalParts = fs.TotalParts
	m.Type = int32(fs.Type)
	m.IsCompleted = fs.IsCompleted
	m.RequestStatus = int32(fs.RequestStatus)
	if fs.UploadRequest != nil {
		m.UploadRequest, _ = fs.UploadRequest.Marshal()
	}
	if fs.DownloadRequest != nil {
		m.DownloadRequest, _ = fs.DownloadRequest.Marshal()
	}
	m.ThumbFileID = fs.ThumbFileID
	m.ThumbFilePath = fs.ThumbFilePath
	m.ThumbPosition = fs.ThumbPosition
	m.ThumbTotalSize = fs.ThumbTotalSize
	m.ThumbPartNo = fs.ThumbPartNo
	m.ThumbTotalParts = fs.ThumbTotalParts

	return m
}

func (fs *FileStatus) LoadDTO(d dto.FileStatus, progress domain.OnFileStatusChanged) {
	fs.MessageID = d.MessageID
	fs.FileID = d.FileID
	fs.ClusterID = d.ClusterID
	fs.AccessHash = uint64(d.AccessHash)
	fs.Version = d.Version
	fs.FilePath = d.FilePath
	// fs.Position = d.Position
	fs.TotalSize = d.TotalSize

	items := make([]int64, 0)
	json.Unmarshal(d.PartList, &items)
	fs.PartList = NewQueueParts()
	fs.PartList.PushMany(items)

	fs.TotalParts = d.TotalParts
	fs.Type = domain.FileStateType(d.Type)
	fs.IsCompleted = d.IsCompleted
	fs.RequestStatus = domain.RequestStatus(d.RequestStatus)
	fs.UploadRequest = new(msg.ClientSendMessageMedia)
	fs.UploadRequest.Unmarshal(d.UploadRequest)
	fs.DownloadRequest = new(msg.Document)
	fs.DownloadRequest.Unmarshal(d.DownloadRequest)
	fs.onFileStatusChanged = progress

	fs.ThumbFileID = d.ThumbFileID
	fs.ThumbFilePath = d.ThumbFilePath
	fs.ThumbPosition = d.ThumbPosition
	fs.ThumbTotalSize = d.ThumbTotalSize
	fs.ThumbPartNo = d.ThumbPartNo
	fs.ThumbTotalParts = d.ThumbTotalParts

}

func (fs *FileStatus) ReadAsFileGet(partNo int64) (envelop *msg.MessageEnvelope, err error) {

	req := new(msg.FileGet)
	req.Location = new(msg.InputFileLocation)
	req.Location.ClusterID = fs.ClusterID
	req.Location.AccessHash = fs.AccessHash
	req.Location.FileID = fs.FileID
	req.Location.Version = fs.Version

	req.Offset = int32(partNo * domain.FilePayloadSize)
	position := int64(req.Offset)
	var requiredBytes int32 = domain.FilePayloadSize
	if (position + domain.FilePayloadSize) > fs.TotalSize {
		requiredBytes = int32(fs.TotalSize - position)
	}
	req.Limit = requiredBytes

	envelop = new(msg.MessageEnvelope)
	envelop.Constructor = msg.C_FileGet
	envelop.Message, err = req.Marshal()
	envelop.RequestID = uint64(domain.SequentialUniqueID())

	log.LOG_Debug("FileStatus::ReadAsFileGet()",
		zap.Int64("MsgID", fs.MessageID),
		zap.Int32("Offset", req.Offset),
		zap.Int32("Limit", req.Limit),
		zap.Int64("FileID", req.Location.FileID),
		zap.Uint64("AccessHash", req.Location.AccessHash),
		zap.Int32("ClusterID", req.Location.ClusterID),
		zap.Int32("Version", req.Location.Version),
	)

	return
}

func (fs *FileStatus) StartDownload(fm *FileManager) {
	fs.mx.Lock()
	if fs.started {
		fs.mx.Unlock()
		return
	}
	fs.started = true
	fs.mx.Unlock()

	fs.stop = false

	partCount := fs.PartList.Length()
	workersCount := domain.FilePipelineCount
	if partCount < domain.FilePipelineCount {
		workersCount = partCount
	}

	for i := 0; i < workersCount; i++ {
		go fs.downloader_job(fm)
	}
}

func (fs *FileStatus) StartUpload(fm *FileManager) {

	fs.mx.Lock()
	if fs.started {
		fs.mx.Unlock()
		return
	}
	fs.started = true
	fs.mx.Unlock()

	fs.stop = false

	partCount := fs.PartList.Length()
	workersCount := domain.FilePipelineCount
	if partCount < domain.FilePipelineCount {
		workersCount = partCount
	}
	// upload thumbnail first
	fs.upload_thumbnail(fm)

	fs.chUploadProgress = make(chan int64, workersCount)
	go fs.monitorUploadProgress(fm)
	// start uploading file
	for i := 0; i < workersCount; i++ {
		go fs.uploader_job(fm)
	}
}

func (fs *FileStatus) monitorUploadProgress(fm *FileManager) {
	for {
		select {
		case partIdx := <-fs.chUploadProgress:
			isCompleted := fs.ReadCommit(0, false, partIdx)
			if isCompleted {
				// call completed delegate
				fm.uploadCompleted(fs.MessageID, fs.FileID, fs.TargetID, fs.ClusterID, fs.TotalParts, fs.Type, fs.FilePath, fs.UploadRequest, fs.ThumbFileID, fs.ThumbTotalParts)
				return
			}
		}
	}
}
func (fs *FileStatus) Stop() {
	fs.stop = true
	fs.started = false
}

func (fs *FileStatus) downloader_job(fm *FileManager) {
	for {
		partIdx, err := fs.PartList.Pop()
		if err == nil {
			envelop, err := fs.ReadAsFileGet(partIdx)
			if err != nil {
				log.LOG_Error("downloader_job() -> ReadAsFileGet()", zap.Int64("msgID", fs.MessageID), zap.Int64("PartNo", partIdx))
				continue
			}
			fm.SendDownloadRequest(envelop, fs, partIdx)
		} else {
			return
		}
	}
}
func (fs *FileStatus) uploader_job(fm *FileManager) {
	for {
		if fs.stop {
			return
		}
		partIdx, err := fs.PartList.Pop()
		if err == nil {
			// keep last part until all other parts upload successfully
			partCount := fs.PartList.Length()
			if (partIdx+1) == fs.TotalParts && partCount > 0 {
				fs.PartList.Push(partIdx)
				continue
			}
			envelop, readCount, err := fs.ReadAsFileSavePart(false, partIdx)
			if err != nil {
				log.LOG_Error("FileManager::startUploadQueue()", zap.Error(err), zap.String("filePath", fs.FilePath))
				continue
			}
			fm.SendUploadRequest(envelop, int64(readCount), fs, partIdx)

		} else {
			return
		}

	}

}

func (fs *FileStatus) upload_thumbnail(fm *FileManager) {
	for fs.ThumbPosition < fs.ThumbTotalSize {
		envelop, readCount, err := fs.ReadAsFileSavePart(true, 0)
		if err != nil {
			log.LOG_Error("FileManager::startUploadQueue()", zap.Error(err), zap.String("filePath", fs.FilePath))
			continue
		}
		fm.SendUploadRequest(envelop, int64(readCount), fs, 0)
	}
}
