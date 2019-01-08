package filemanager

import (
	"os"
	"sync"

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
	MessageID           int64     `json:"MessageID"`
	FileID              int64     `json:"FileID"`
	ClusterID           int32     `json:"ClusterID"`
	AccessHash          uint64    `json:"AccessHash"`
	FilePath            string    `json:"FilePath"`
	Position            int64     `json:"Position"`
	TotalSize           int64     `json:"TotalSize"`
	PartNo              int32     `json:"PartNo"`
	TotalParts          int32     `json:"TotalParts"`
	Type                StateType `json:"StatusType"`
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
	progress domain.OnFileStatusChanged) *FileStatus {
	fs := &FileStatus{
		MessageID:           messageID,
		FileID:              fileID,
		FilePath:            filePath,
		TotalSize:           totalSize,
		ClusterID:           clusterID,
		AccessHash:          accessHash,
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

	buff := make([]byte, 0, domain.FilePayloadSize)
	readCount, err := file.ReadAt(buff, fs.Position)
	file.Close()
	if err != nil {
		return nil, 0, err
	}
	fs.mx.Unlock()

	return buff, readCount, nil
}

// Write writes givin data to current position of file
func (fs *FileStatus) Write(data []byte) error {
	fs.mx.Lock()

	var err error
	var file *os.File

	// create file if its not exist
	if _, err = os.Stat(fs.FilePath); os.IsNotExist(err) {
		file, err = os.Create(fs.FilePath)
		if err != nil {
			return err
		}
		// truncate reserves size of file
		err = file.Truncate(fs.TotalSize)
		if err != nil {
			return err
		}
	}
	// open file if its not open
	if file == nil {
		file, err = os.Open(fs.FilePath)
		if err != nil {
			return err
		}
	}

	// write to file
	count, err := file.WriteAt(data, fs.Position)
	if err != nil {
		return err
	}
	fs.Position += int64(count)

	fs.fileStatusChanged()

	fs.mx.Unlock()

	return nil
}

//ReadCommit apply that last read process result was success and increase counter and progress
func (fs *FileStatus) ReadCommit(count int64) (isCompleted bool) {
	fs.Position += count
	fs.PartNo++
	isCompleted = fs.PartNo == fs.TotalParts
	fs.fileStatusChanged()
	return isCompleted
}

func (fs *FileStatus) fileStatusChanged() {
	// TODO : save file status to DB
	err := repo.Ctx().Files.SaveFileStatus(fs.FileID, fs.FilePath, fs.Position, fs.TotalSize, fs.PartNo, fs.TotalParts)
	if err != nil {
		log.LOG_Debug("fileStatusChanged() failed to save in DB", zap.Error(err))
	}
	if fs.onFileStatusChanged != nil {
		fs.onFileStatusChanged(fs.MessageID, fs.Position, fs.TotalSize)
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
	req.PartID = fs.PartNo
	req.TotalParts = fs.TotalParts

	envelop.Message, err = req.Marshal()
	envelop.RequestID = uint64(domain.SequentialUniqueID())

	return
}
