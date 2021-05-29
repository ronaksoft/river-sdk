package mini

import (
	"fmt"
	"git.ronaksoft.com/river/msg/go/msg"
	fileCtrl "git.ronaksoft.com/river/sdk/internal/ctrl_file"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/repo"
	"git.ronaksoft.com/river/sdk/internal/request"
	"github.com/ronaksoft/rony"
	"github.com/ronaksoft/rony/tools"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"path"
)

func (r *River) GetFilePath(clusterID int32, fileID int64, accessHash int64) string {
	// We only support thumbnail for now
	return path.Join(repo.DirCache, fmt.Sprintf("%d%d%s", fileID, clusterID, ".jpg"))
}

func (r *River) GetFileStatus(clusterID int32, fileID int64, accessHash int64) []byte {
	fileStatus := &msg.ClientFileStatus{
		Status:   int32(request.StatusNone),
		Progress: 0,
		FilePath: "",
	}
	filePath := r.GetFilePath(clusterID, fileID, accessHash)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fileStatus.FilePath = ""
	} else {
		fileStatus.FilePath = filePath
		fileStatus.Progress = 100
		fileStatus.Status = int32(request.StatusCompleted)
	}

	buf, _ := fileStatus.Marshal()
	return buf
}

func (r *River) FileDownloadThumbnail(clusterID int32, fileID int64, accessHash int64) (err error) {
	filePath := path.Join(repo.DirCache, fmt.Sprintf("%d%d%s", fileID, clusterID, ".jpg"))
	req := &msg.FileGet{
		Location: &msg.InputFileLocation{
			ClusterID:  clusterID,
			FileID:     fileID,
			AccessHash: uint64(accessHash),
			Version:    0,
		},
		Offset: 0,
		Limit:  0,
	}

	return tools.Try(fileCtrl.RetryMaxAttempts, fileCtrl.RetryWaitTime, func() error {
		reqCB := request.NewCallback(
			0, 0, domain.NextRequestID(), msg.C_FileGet, req,
			func() {
				err = domain.ErrRequestTimeout
			},
			func(res *rony.MessageEnvelope) {
				switch res.Constructor {
				case rony.C_Error:
					x := &rony.Error{}
					_ = x.Unmarshal(res.Message)
					err = x
				case msg.C_File:
					x := &msg.File{}
					err = x.Unmarshal(res.Message)
					if err != nil {
						return
					}

					// write to file path
					err = ioutil.WriteFile(filePath, x.Bytes, 0666)
					if err != nil {
						return
					}
				}
			},
			nil, false, 0, domain.HttpRequestTimeout,
		)
		r.network.HttpCommand(nil, reqCB)
		return err
	})
}

// GetDocumentHash returns the md5 hash of the document
func (r *River) GetDocumentHash(clusterID int32, fileID int64, accessHash int64) string {
	file, err := repo.Files.Get(clusterID, fileID, uint64(accessHash))

	if err != nil {
		logger.Warn("Error On GetDocumentHash (Files.Get)",
			zap.Int32("ClusterID", clusterID),
			zap.Int64("FileID", fileID),
			zap.Int64("AccessHash", accessHash),
			zap.Error(err),
		)
		return ""
	}

	if file.MessageID == 0 {
		logger.Warn("Not a message document",
			zap.Int32("ClusterID", clusterID),
			zap.Int64("FileID", fileID),
			zap.Int64("AccessHash", accessHash),
		)
		return ""
	}

	return file.MD5Checksum
}
