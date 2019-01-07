package filemanger

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"git.ronaksoftware.com/ronak/riversdk/domain"

	"git.ronaksoftware.com/ronak/riversdk/log"
)

const (
	// FileSizeThresholdForCheckHash for files thatare smaller than  this number we will calculate md5 hash to do not reupload same file twice
	FileSizeThresholdForCheckHash = 10 * 1024 * 1024 // 10MB
)

// FileManager manages files status and cache
type FileManager struct {
	mxDown        sync.Mutex
	mxUp          sync.Mutex
	UploadQueue   map[int64]*FileStatus
	DownloadQueue map[int64]*FileStatus
}

// Upload file to server
func (fm *FileManager) Upload(filePath string, progressCB domain.OnFileStatusChanged) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	// fileName := fileInfo.Name()
	fileSize := fileInfo.Size() // size in Byte

	strMd5, err := fm.CalculateMD5(file)
	if err == nil {
		// TODO : check DB with file md5 hash and meeeeehhhhh :/
		log.LOG_Debug(strMd5)
	}

	fileID := domain.SequentialUniqueID()

	state := NewFileStatus(fileID, fileSize, filePath, UploadType, progressCB)
	fm.AddToQueue(state)
	return nil
}

// Download
func (fm *FileManager) Download() {
	// TODO : implement it
}

// AddToQueue add request to queue
func (fm *FileManager) AddToQueue(status *FileStatus) {

	if status.StatusType == UploadType {
		fm.mxUp.Lock()
		fm.UploadQueue[status.FileID] = status
		fm.mxUp.Unlock()
	} else {
		fm.mxDown.Lock()
		fm.DownloadQueue[status.FileID] = status
		fm.mxDown.Unlock()
	}
}

// func (f *FileManager) fnFileSendPart(file *os.File, offIdx int64, wg *sync.WaitGroup) error {
// 	buff := make([]byte, PayloadSize, PayloadSize)
// 	count, err := file.ReadAt(buff, offIdx*PayloadSize)
// 	if err != nil {
// 		return err
// 	}

// 	req := new(msg.FileSavePart)
// 	req.PartID = int32(offIdx + 1)
// 	req.TotalParts
// ====================================
// count := fileSize / PayloadSize
// totalParts := int32(0)
// if (count * PayloadSize) < fileSize {
// 	//we have extra part to send as final part
// 	totalParts = int32(count + 1)
// } else {
// 	totalParts = int32(count)
// 	count-- // keep last part to send as final part

// }
// fileID := domain.SequentialUniqueID()
// buff := make([]byte, PayloadSize, PayloadSize)

// wg := sync.WaitGroup{}
// for idx := int64(0); idx < count; {
// 	wg.Add(1)
// 	readCount, err := file.ReadAt(buff, idx*PayloadSize)
// 	if err != nil {
// 		return err
// 	}
// 	req := new(msg.FileSavePart)
// 	req.FileID = fileID
// 	req.PartID = int32(idx + 1)
// 	req.TotalParts = totalParts
// 	req.Bytes = make([]byte, readCount, readCount)
// 	copy(req.Bytes, buff)

// 	// TODO : send file part and save status
// }
// wg.Wait()

// // TODO : handle failed parts

// // TODO :  make sure all file has been sent

// // Make sure to send the last last of all parts
// readCount, err := file.ReadAt(buff, count*PayloadSize)
// if err != nil {
// 	return err
// }
// req := new(msg.FileSavePart)
// req.FileID = fileID
// req.PartID = int32(count + 1)
// req.TotalParts = totalParts
// req.Bytes = make([]byte, readCount, readCount)
// copy(req.Bytes, buff)

// // TODO : send file part and save status

// // TODO :  on last part success send MessagesSendMedia request
// }

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
