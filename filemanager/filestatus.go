package filemanager

import (
	"encoding/json"
	"os"
	"sync"

	"git.ronaksoftware.com/ronak/riversdk/repo/dto"

	"go.uber.org/zap"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/logs"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/repo"
)

// FileStatus monitors file upload/download state
type FileStatus struct {
	mx         sync.Mutex
	mxPartList sync.Mutex
	MessageID  int64  `json:"MessageID"`
	FileID     int64  `json:"FileID"`
	TargetID   int64  `json:"TargetID"`
	ClusterID  int32  `json:"ClusterID"`
	AccessHash uint64 `json:"AccessHash"`
	Version    int32  `json:"Version"`
	FilePath   string `json:"FilePath"`

	TotalSize       int64                       `json:"TotalSize"`
	PartList        map[int64]bool              `json:"PartNo"`
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
	chPartList       chan int64
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

		TotalParts:          0,
		onFileStatusChanged: progress,
		Type:                stateType,
		RequestStatus:       domain.RequestStateInProgress,
	}

	// create partlist
	fs.TotalParts = CalculatePartsCount(totalSize)

	fs.PartList = make(map[int64]bool, fs.TotalParts)
	fs.chPartList = make(chan int64, fs.TotalParts)
	for i := int64(0); i < fs.TotalParts; i++ {
		fs.PartList[i] = true
		fs.chPartList <- i
	}

	return fs
}

// CalculatePartsCount return total parts
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

	if fs.stop {
		return
	}

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

	fs.deleteFromPartList(partIdx)
	count := fs.partListCount()
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

// ReadCommit apply that last read process result was success and increase counter and progress
func (fs *FileStatus) ReadCommit(count int64, isThumbnail bool, partIdx int64) (isCompleted bool) {
	if isThumbnail {
		fs.ThumbPosition += count
		fs.ThumbPartNo++
		repo.Ctx().Files.SaveFileStatus(fs.GetDTO())
		return
	}
	if fs.stop {
		return
	}

	fs.deleteFromPartList(partIdx)
	partCount := fs.partListCount()
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
		logs.Error("fileStatusChanged() failed to save in DB", zap.Error(err))
	}

	lenParts := int64(fs.partListCount())

	processedParts := fs.TotalParts - lenParts
	if fs.onFileStatusChanged != nil {
		fs.onFileStatusChanged(fs.MessageID, processedParts, fs.TotalParts, fs.Type)
	}

}

// ReadAsFileSavePart read required chunk of data and pack them into FileSavePart
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

// GetDTO map FileStatus to its repo DTO to save in DB
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

	partList, _ := json.Marshal(fs.PartList)
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

// LoadDTO Map related to repo DTO to FileStatus
func (fs *FileStatus) LoadDTO(d dto.FileStatus, progress domain.OnFileStatusChanged) {
	fs.MessageID = d.MessageID
	fs.FileID = d.FileID
	fs.ClusterID = d.ClusterID
	fs.AccessHash = uint64(d.AccessHash)
	fs.Version = d.Version
	fs.FilePath = d.FilePath
	// fs.Position = d.Position
	fs.TotalSize = d.TotalSize

	fs.PartList = make(map[int64]bool)
	json.Unmarshal(d.PartList, &fs.PartList)

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

	fs.chPartList = make(chan int64, len(fs.PartList))
	for p := range fs.PartList {
		fs.chPartList <- p
	}
}

// ReadAsFileGet create related FileGet request
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

	logs.Debug("FileStatus::ReadAsFileGet()",
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

// StartDownload begins download
func (fs *FileStatus) StartDownload(fm *FileManager) {
	fs.mx.Lock()
	if fs.started {
		fs.mx.Unlock()
		return
	}
	fs.started = true
	fs.mx.Unlock()

	fs.stop = false

	partCount := fs.partListCount()
	workersCount := domain.FilePipelineCount
	if partCount < domain.FilePipelineCount {
		workersCount = partCount
	}

	// call onProgresschange to notify ui
	fs.fileStatusChanged()

	for i := 0; i < workersCount; i++ {
		go fs.downloaderJob(fm)
	}
}

// StartUpload begins upload
func (fs *FileStatus) StartUpload(fm *FileManager) {

	fs.mx.Lock()
	if fs.started {
		fs.mx.Unlock()
		return
	}
	fs.started = true
	fs.mx.Unlock()

	fs.stop = false

	partCount := fs.partListCount()
	workersCount := domain.FilePipelineCount
	if partCount < domain.FilePipelineCount {
		workersCount = partCount
	}

	// call onProgresschange to notify ui
	fs.fileStatusChanged()

	// upload thumbnail first
	fs.uploadThumbnail(fm)

	fs.chUploadProgress = make(chan int64, workersCount)
	go fs.monitorUploadProgress(fm)
	// start uploading file
	for i := 0; i < workersCount; i++ {
		go fs.uploaderJob(fm)
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

// Stop set stop flag
func (fs *FileStatus) Stop() {
	fs.stop = true
	fs.started = false
}

func (fs *FileStatus) downloaderJob(fm *FileManager) {
	for {
		if fs.stop {
			return
		}

		select {
		case partIdx := <-fs.chPartList:
			envelop, err := fs.ReadAsFileGet(partIdx)
			if err != nil {
				logs.Error("downloaderJob()", zap.Int64("msgID", fs.MessageID), zap.Int64("PartNo", partIdx), zap.Error(err))
				break
			}
			fm.SendDownloadRequest(envelop, fs, partIdx)
		default:
			return
		}
	}
}
func (fs *FileStatus) uploaderJob(fm *FileManager) {
	for {
		if fs.stop {
			return
		}
		select {
		case partIdx := <-fs.chPartList:
			// keep last part until all other parts upload successfully
			partCount := fs.partListCount()
			if (partIdx+1) == fs.TotalParts && partCount > 1 {
				fs.chPartList <- partIdx
				break
			}
			envelop, readCount, err := fs.ReadAsFileSavePart(false, partIdx)
			if err != nil {
				logs.Error("uploaderJob()", zap.Error(err), zap.String("filePath", fs.FilePath))
				break
			}
			fm.SendUploadRequest(envelop, int64(readCount), fs, partIdx)
		default:
			return
		}
	}

}

func (fs *FileStatus) uploadThumbnail(fm *FileManager) {
	for fs.ThumbPosition < fs.ThumbTotalSize {
		if fs.stop {
			return
		}

		envelop, readCount, err := fs.ReadAsFileSavePart(true, 0)
		if err != nil {
			logs.Error("uploaderJob()", zap.Error(err), zap.String("filePath", fs.FilePath))
			continue
		}
		fm.SendUploadRequest(envelop, int64(readCount), fs, 0)
	}
}

func (fs *FileStatus) partListCount() int {
	fs.mxPartList.Lock()
	count := len(fs.PartList)
	fs.mxPartList.Unlock()
	return count
}
func (fs *FileStatus) deleteFromPartList(partIdx int64) {
	fs.mxPartList.Lock()
	delete(fs.PartList, partIdx)
	fs.mxPartList.Unlock()
}
