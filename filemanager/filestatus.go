package filemanager

import (
	"os"
	"sync"

	"git.ronaksoftware.com/ronak/riversdk/repo/dto"

	"go.uber.org/zap"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/log"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/repo"
)

type StateType bool

var (
	StateDownload StateType = true
	StateUpload   StateType = false
)

// FileStatus monitors file state
type FileStatus struct {
	mx                  sync.Mutex
	MessageID           int64                       `json:"MessageID"`
	FileID              int64                       `json:"FileID"`
	ClusterID           int32                       `json:"ClusterID"`
	AccessHash          uint64                      `json:"AccessHash"`
	Version             int32                       `json:"Version"`
	FilePath            string                      `json:"FilePath"`
	Position            int64                       `json:"Position"`
	TotalSize           int64                       `json:"TotalSize"`
	PartNo              int32                       `json:"PartNo"`
	TotalParts          int32                       `json:"TotalParts"`
	Type                StateType                   `json:"StatusType"`
	IsCompleted         bool                        `json:"IsCompleted"`
	RequestStatus       domain.RequestStatus        `json:"RequestStatus"`
	UploadRequest       *msg.ClientSendMessageMedia `json:"UploadRequest"`
	DownloadRequest     *msg.Document               `json:"DownloadRequest"`
	onFileStatusChanged domain.OnFileStatusChanged
}

// NewFileStatus create new instance
func NewFileStatus(messageID int64,
	fileID int64,
	totalSize int64,
	filePath string,
	isDownload StateType,
	clusterID int32,
	accessHash uint64,
	version int32,
	progress domain.OnFileStatusChanged) *FileStatus {
	fs := &FileStatus{
		MessageID:           messageID,
		FileID:              fileID,
		FilePath:            filePath,
		TotalSize:           totalSize,
		ClusterID:           clusterID,
		AccessHash:          accessHash,
		Version:             version,
		Position:            0,
		PartNo:              0,
		TotalParts:          0,
		onFileStatusChanged: progress,
		Type:                isDownload,
	}

	count := totalSize / domain.FilePayloadSize
	if (count * domain.FilePayloadSize) < totalSize {
		fs.TotalParts = int32(count + 1)
	} else {
		fs.TotalParts = int32(count)
	}

	return fs
}

// Read reads next required chunk of data
func (fs *FileStatus) Read() ([]byte, int, error) {
	fs.mx.Lock()

	file, err := os.Open(fs.FilePath)
	if err != nil {
		return nil, 0, err
	}
	var requiredBytes int64 = domain.FilePayloadSize
	if (fs.Position + domain.FilePayloadSize) > fs.TotalSize {
		requiredBytes = fs.TotalSize - fs.Position
	}
	buff := make([]byte, requiredBytes)
	readCount, err := file.ReadAt(buff, fs.Position)
	file.Close()
	if err != nil {
		return nil, 0, err
	}
	fs.mx.Unlock()

	return buff, readCount, nil
}

// Write writes givin data to current position of file
func (fs *FileStatus) Write(data []byte) (isCompleted bool, err error) {
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
		file, err = os.OpenFile(fs.FilePath, os.O_RDWR, os.ModePerm)
		if err != nil {
			return
		}
	}

	// write to file
	count, err := file.WriteAt(data, fs.Position)
	file.Close()
	if err != nil {
		return
	}
	fs.Position += int64(count)
	fs.IsCompleted = fs.Position >= fs.TotalSize
	isCompleted = fs.IsCompleted
	if isCompleted {
		repo.Ctx().Files.SaveDownloadingFile(fs.GetDTO())
	}
	fs.fileStatusChanged()

	return
}

//ReadCommit apply that last read process result was success and increase counter and progress
func (fs *FileStatus) ReadCommit(count int64) (isCompleted bool) {
	fs.Position += count
	fs.PartNo++
	fs.IsCompleted = fs.PartNo == fs.TotalParts
	fs.fileStatusChanged()
	return fs.IsCompleted
}

func (fs *FileStatus) fileStatusChanged() {
	// TODO : save file status to DB
	err := repo.Ctx().Files.SaveFileStatus(fs.GetDTO())
	if err != nil {
		log.LOG_Debug("fileStatusChanged() failed to save in DB", zap.Error(err))
	}
	if fs.onFileStatusChanged != nil {
		fs.onFileStatusChanged(fs.MessageID, fs.Position, fs.TotalSize, fs.Type == StateDownload)
	}

}

func (fs *FileStatus) ReadAsFileSavePart() (envelop *msg.MessageEnvelope, readCount int, err error) {

	var buff []byte
	buff, readCount, err = fs.Read()
	if err != nil {
		return
	}
	req := new(msg.FileSavePart)
	req.Bytes = buff
	req.FileID = fs.FileID
	req.PartID = fs.PartNo + 1
	req.TotalParts = fs.TotalParts

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
	m.Position = fs.Position
	m.TotalSize = fs.TotalSize
	m.PartNo = fs.PartNo
	m.TotalParts = fs.TotalParts
	m.Type = bool(fs.Type)
	m.IsCompleted = fs.IsCompleted
	m.RequestStatus = int32(fs.RequestStatus)
	if fs.UploadRequest != nil {
		m.UploadRequest, _ = fs.UploadRequest.Marshal()
	}
	if fs.DownloadRequest != nil {
		m.DownloadRequest, _ = fs.DownloadRequest.Marshal()
	}

	return m
}

func (fs *FileStatus) LoadDTO(d dto.FileStatus, progress domain.OnFileStatusChanged) {
	fs.MessageID = d.MessageID
	fs.FileID = d.FileID
	fs.ClusterID = d.ClusterID
	fs.AccessHash = uint64(d.AccessHash)
	fs.Version = d.Version
	fs.FilePath = d.FilePath
	fs.Position = d.Position
	fs.TotalSize = d.TotalSize
	fs.PartNo = d.PartNo
	fs.TotalParts = d.TotalParts
	fs.Type = StateType(d.Type)
	fs.IsCompleted = d.IsCompleted
	fs.RequestStatus = domain.RequestStatus(d.RequestStatus)
	fs.UploadRequest = new(msg.ClientSendMessageMedia)
	fs.UploadRequest.Unmarshal(d.UploadRequest)
	fs.DownloadRequest = new(msg.Document)
	fs.DownloadRequest.Unmarshal(d.DownloadRequest)
	fs.onFileStatusChanged = progress
}

func (fs *FileStatus) ReadAsFileGet() (envelop *msg.MessageEnvelope, err error) {

	req := new(msg.FileGet)
	req.Location = new(msg.InputFileLocation)
	req.Location.ClusterID = fs.ClusterID
	req.Location.AccessHash = fs.AccessHash
	req.Location.FileID = fs.FileID
	req.Location.Version = fs.Version
	req.Offset = int32(fs.Position)
	var requiredBytes int32 = domain.FilePayloadSize
	if (fs.Position + domain.FilePayloadSize) > fs.TotalSize {
		requiredBytes = int32(fs.TotalSize - fs.Position)
	}
	req.Limit = requiredBytes

	envelop = new(msg.MessageEnvelope)
	envelop.Constructor = msg.C_FileGet
	envelop.Message, err = req.Marshal()
	envelop.RequestID = uint64(domain.SequentialUniqueID())

	return
}
