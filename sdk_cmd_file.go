package riversdk

import (
	"git.ronaksoftware.com/ronak/riversdk/msg/ext"
	fileCtrl "git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_file"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	mon "git.ronaksoftware.com/ronak/riversdk/pkg/monitoring"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"os"
	"time"
)

func (r *River) GetFileStatus(clusterID int32, fileID int64, accessHash int64) []byte {
	fileStatus := new(msg.ClientFileStatus)
	downloadRequest, ok := r.fileCtrl.GetDownloadRequest(clusterID, fileID, uint64(accessHash))
	if ok {
		fileStatus.FilePath = downloadRequest.FilePath
		fileStatus.Status = int32(domain.RequestStatusInProgress)
		fileStatus.Progress = int64(float64(len(downloadRequest.DownloadedParts)) / float64(downloadRequest.TotalParts) * 100)
	} else {
		clientFile, err := repo.Files.Get(clusterID, fileID, uint64(accessHash))
		if err == nil {
			filePath := fileCtrl.GetFilePath(clientFile)
			if _, err = os.Stat(filePath); os.IsNotExist(err) {
				fileStatus.FilePath = ""
			} else {
				fileStatus.FilePath = filePath
				fileStatus.Progress = 100
				fileStatus.Status = int32(domain.RequestStatusCompleted)
			}
		}
	}

	buf, _ := fileStatus.Marshal()
	return buf
}

func (r *River) GetFilePath(clusterID int32, fileID int64, accessHash int64) string {
	clientFile, err := repo.Files.Get(clusterID, fileID, uint64(accessHash))
	if err == nil {
		filePath := fileCtrl.GetFilePath(clientFile)
		return filePath
	}
	return ""
}

func (r *River) FileDownloadAsync(clusterID int32, fileID int64, accessHash int64, skipDelegate bool) (reqID string) {
	reqID, _ = r.fileCtrl.DownloadAsync(clusterID, fileID, uint64(accessHash), skipDelegate)
	return
}

func (r *River) FileDownloadSync(clusterID int32, fileID int64, accessHash int64, skipDelegate bool) error {
	_, err := r.fileCtrl.DownloadSync(clusterID, fileID, uint64(accessHash), skipDelegate)
	return err
}

// CancelDownload cancel download
func (r *River) CancelDownload(clusterID int32, fileID int64, accessHash int64) {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("CancelDownload", time.Now().Sub(startTime))
	}()

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

// CancelUpload cancel upload
func (r *River) CancelUpload(clusterID int32, fileID int64, accessHash int64) {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("CancelUpload", time.Now().Sub(startTime))
	}()

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

// ResumeDownload
func (r *River) ResumeUpload(pendingMessageID int64) {
	pendingMessage := repo.PendingMessages.GetByID(pendingMessageID)
	if pendingMessage == nil {
		return
	}
	req := new(msg.ClientSendMessageMedia)
	_ = req.Unmarshal(pendingMessage.Media)

	if _, ok := r.fileCtrl.GetUploadRequest(pendingMessage.FileID); !ok {
		go r.fileCtrl.UploadMessageDocument(pendingMessageID, req.FilePath, req.ThumbFilePath, pendingMessage.FileID, pendingMessage.ThumbID)
	}
}

// AccountUploadPhoto upload user profile photo
func (r *River) AccountUploadPhoto(filePath string) (reqID string) {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("AccountUploadPhoto", time.Now().Sub(startTime))
	}()

	reqID = r.fileCtrl.UploadUserPhoto(filePath)
	return
}

// GroupUploadPhoto upload group profile photo
func (r *River) GroupUploadPhoto(groupID int64, filePath string) (reqID string) {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("GroupUploadPhoto", time.Now().Sub(startTime))
	}()

	reqID = r.fileCtrl.UploadGroupPhoto(groupID, filePath)
	return
}

// GetSharedMedia search in given dialog files
func (r *River) GetSharedMedia(peerID int64, peerType int32, mediaType int32, delegate RequestDelegate) {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("GetSharedMedia", time.Now().Sub(startTime))
	}()

	msgs, err := repo.Messages.GetSharedMedia(peerID, peerType, domain.SharedMediaType(mediaType))
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

// GetGetDBStatus returns message IDs and total size of each media stored in user's database
func (r *River) GetDBStatus(delegate RequestDelegate) {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("GetDBStatus", time.Now().Sub(startTime))
	}()
	delegate.OnTimeout(domain.ErrDoesNotExists)
	res := msg.DBMediaInfo{}

	// TODO:: do stuff here

	bytes, _ := res.Marshal()
	if delegate != nil {
		delegate.OnComplete(bytes)
	}

}

// ClearCache removes files from client device, allMedia means clear all media types
// peerID 0 means all peers
func (r *River) ClearCache(peerID int64, mediaTypes string, allMedia bool) bool {
	return false
}
