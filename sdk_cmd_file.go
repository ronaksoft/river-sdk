package riversdk

import (
	"git.ronaksoftware.com/river/msg/chat"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"github.com/dgraph-io/badger"
	"go.uber.org/zap"
	"os"
)

func (r *River) GetFileStatus(clusterID int32, fileID int64, accessHash int64) []byte {
	fileStatus := new(msg.ClientFileStatus)
	if clusterID == 0 && accessHash == 0 {
		// It it Upload
		uploadRequest, ok := r.fileCtrl.GetUploadRequest(fileID)
		if ok {
			fileStatus.FilePath = uploadRequest.FilePath
			if uploadRequest.TotalParts > 0 {
				fileStatus.Progress = int64(float64(len(uploadRequest.UploadedParts)) / float64(uploadRequest.TotalParts) * 100)
			}
			if repo.Files.IsMarkedAsUploaded(fileID) {
				fileStatus.Status = int32(domain.RequestStatusCompleted)
			} else {
				fileStatus.Status = int32(domain.RequestStatusInProgress)
			}
		} else {
			fileStatus.FilePath = uploadRequest.FilePath
			fileStatus.Status = int32(domain.RequestStatusNone)
			fileStatus.Progress = 0
		}
	} else {
		downloadRequest, ok := r.fileCtrl.GetDownloadRequest(clusterID, fileID, uint64(accessHash))
		if ok {
			fileStatus.FilePath = downloadRequest.FilePath
			fileStatus.Status = int32(domain.RequestStatusInProgress)
			if downloadRequest.TotalParts > 0 {
				fileStatus.Progress = int64(float64(len(downloadRequest.DownloadedParts)) / float64(downloadRequest.TotalParts) * 100)
			}
		} else {
			clientFile, err := repo.Files.Get(clusterID, fileID, uint64(accessHash))
			if err == nil {
				filePath := repo.Files.GetFilePath(clientFile)
				if _, err = os.Stat(filePath); os.IsNotExist(err) {
					fileStatus.FilePath = ""
				} else {
					fileStatus.FilePath = filePath
					fileStatus.Progress = 100
					fileStatus.Status = int32(domain.RequestStatusCompleted)
				}
			}
		}
	}

	buf, _ := fileStatus.Marshal()
	return buf
}

func (r *River) GetFilePath(clusterID int32, fileID int64, accessHash int64) string {
	clientFile, err := repo.Files.Get(clusterID, fileID, uint64(accessHash))
	if err == nil {
		filePath := repo.Files.GetFilePath(clientFile)
		return filePath
	}
	return ""
}

func (r *River) FileDownloadAsync(clusterID int32, fileID int64, accessHash int64, skipDelegate bool) (reqID string) {
	var err error
	reqID, err = r.fileCtrl.DownloadAsync(clusterID, fileID, uint64(accessHash), skipDelegate)
	switch err {
	case badger.ErrKeyNotFound:
		logs.Warn("Error On GetFile (Key not found)",
			zap.Int32("ClusterID", clusterID),
			zap.Int64("FileID", fileID),
			zap.Int64("AccessHash", int64(accessHash)),
		)
	default:
		logs.Warn("Error On GetFile",
			zap.Int32("ClusterID", clusterID),
			zap.Int64("FileID", fileID),
			zap.Int64("AccessHash", int64(accessHash)),
			zap.Error(err),
		)
	}
	return
}

func (r *River) FileDownloadSync(clusterID int32, fileID int64, accessHash int64, skipDelegate bool) error {
	_, err := r.fileCtrl.DownloadSync(clusterID, fileID, uint64(accessHash), skipDelegate)
	return err
}

// CancelDownload cancel download
func (r *River) CancelDownload(clusterID int32, fileID int64, accessHash int64) {
	clientFile, err := repo.Files.Get(clusterID, fileID, uint64(accessHash))
	if err != nil {
		return
	}
	if clientFile.MessageID == 0 {
		return
	}

	downloadRequest, ok := r.fileCtrl.GetDownloadRequest(clusterID, fileID, uint64(accessHash))
	if !ok {
		return
	}
	r.fileCtrl.CancelDownloadRequest(downloadRequest.GetID())
}

// CancelUpload cancels the upload and deletes the pending message associated with that media.
// This function MAY NOT stop the upload instantly but for sure it will be canceled in the first
// possible time in future.
func (r *River) CancelUpload(clusterID int32, fileID int64, accessHash int64) {
	clientFile, err := repo.Files.Get(clusterID, fileID, uint64(accessHash))
	if err != nil {
		return
	}
	if clientFile.MessageID == 0 {
		return
	}
	pendingMessage := repo.PendingMessages.GetByID(clientFile.MessageID)
	if pendingMessage == nil {
		return
	}
	_ = repo.PendingMessages.Delete(pendingMessage.ID)

	uploadRequest, ok := r.fileCtrl.GetUploadRequest(fileID)
	if !ok {
		return
	}
	r.fileCtrl.CancelUploadRequest(uploadRequest.GetID())
}

// ResumeDownload must be called if for any reason the upload of a ClientSendMediaMessage failed,
// then client should call this function by providing the pending message id, or if delete the pending
// message.
func (r *River) ResumeUpload(pendingMessageID int64) {
	pendingMessage := repo.PendingMessages.GetByID(pendingMessageID)
	if pendingMessage == nil {
		return
	}
	req := new(msg.ClientSendMessageMedia)
	_ = req.Unmarshal(pendingMessage.Media)

	logs.Info("River resumes upload", zap.Int64("MsgID", pendingMessageID))
	if _, ok := r.fileCtrl.GetUploadRequest(pendingMessage.FileID); !ok {
		r.fileCtrl.UploadMessageDocument(pendingMessageID, req.FilePath, req.ThumbFilePath, pendingMessage.FileID, pendingMessage.ThumbID)
	}
}

// AccountUploadPhoto upload user profile photo
func (r *River) AccountUploadPhoto(filePath string) (reqID string) {
	reqID = r.fileCtrl.UploadUserPhoto(filePath)
	return
}

// GroupUploadPhoto upload group profile photo
func (r *River) GroupUploadPhoto(groupID int64, filePath string) (reqID string) {
	reqID = r.fileCtrl.UploadGroupPhoto(groupID, filePath)
	return
}

// GetSharedMedia search in given dialog files
func (r *River) GetSharedMedia(peerID int64, peerType int32, mediaType int32, delegate RequestDelegate) {
	msgs, err := repo.Messages.GetSharedMedia(peerID, peerType, msg.ClientMediaType(mediaType))
	if err != nil {
		out := new(msg.MessageEnvelope)
		res := new(msg.Error)
		res.Code = "00"
		res.Items = err.Error()
		msg.ResultError(out, res)
		outBytes, _ := out.Marshal()
		if delegate != nil {
			delegate.OnComplete(outBytes)
		}
		return
	}

	// get users && group IDs
	userIDs := domain.MInt64B{}
	groupIDs := domain.MInt64B{}
	for _, m := range msgs {
		if m.PeerType == int32(msg.PeerSelf) || m.PeerType == int32(msg.PeerUser) {
			userIDs[m.PeerID] = true
		}
		if m.PeerType == int32(msg.PeerGroup) {
			groupIDs[m.PeerID] = true
		}
		if m.SenderID > 0 {
			userIDs[m.SenderID] = true
		}
		if m.FwdSenderID > 0 {
			userIDs[m.FwdSenderID] = true
		}
	}

	users := repo.Users.GetMany(userIDs.ToArray())
	groups := repo.Groups.GetMany(groupIDs.ToArray())

	msgMany := new(msg.MessagesMany)
	msgMany.Messages = msgs
	msgMany.Users = users
	msgMany.Groups = groups

	out := new(msg.MessageEnvelope)
	out.Constructor = msg.C_MessagesMany
	out.Message, _ = msgMany.Marshal()
	outBytes, _ := out.Marshal()
	if delegate != nil {
		delegate.OnComplete(outBytes)
	}
}

// GetDocumentHash
func (r *River) GetDocumentHash(clusterID int32, fileID int64, accessHash int64) string {
	file , err := repo.Files.Get(clusterID,fileID,uint64(accessHash))

	if err != nil {
		logs.Warn("Error On GetDocumentHash (Files.Get)",
			zap.Int32("ClusterID", clusterID),
			zap.Int64("FileID", fileID),
			zap.Int64("AccessHash", int64(accessHash)),
			zap.Error(err),
		)
		return ""
	}

	if file.MessageID == 0 {
		logs.Warn("Not a message document",
			zap.Int32("ClusterID", clusterID),
			zap.Int64("FileID", fileID),
			zap.Int64("AccessHash", int64(accessHash)),
		)
		return ""
	}

	return file.MD5Checksum
}
