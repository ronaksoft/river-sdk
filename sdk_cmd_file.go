package riversdk

import (
	"git.ronaksoftware.com/ronak/riversdk/msg/ext"
	fileCtrl "git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_file"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	mon "git.ronaksoftware.com/ronak/riversdk/pkg/monitoring"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"os"
	"time"
)

func (r *River) GetFileStatus(clusterID int32, fileID int64, accessHash uint64) string {
	fileStatus := new(msg.ClientFileStatus)
	downloadRequest, ok := r.fileCtrl.GetDownloadRequest(clusterID, fileID, accessHash)
	if ok {
		fileStatus.FilePath = downloadRequest.FilePath
		fileStatus.Status = int32(domain.RequestStatusInProgress)
		fileStatus.Progress = int64(float64(len(downloadRequest.DownloadedParts)) / float64(downloadRequest.TotalParts) * 100)
	} else {
		clientFile, err := repo.Files.Get(clusterID, fileID, accessHash)
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
	return ronak.ByteToStr(buf)
}

func (r *River) FileDownload(clusterID int32, fileID int64, accessHash uint64) error {
	_, err := r.fileCtrl.DownloadFile(clusterID, fileID, accessHash)
	return err
}

// CancelDownload cancel download
func (r *River) CancelDownload(clusterID int32, fileID int64, accessHash uint64) {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("CancelDownload", time.Now().Sub(startTime))
	}()
	// fs, err := repo.Files.GetStatus(msgID)
	// if err != nil {
	// 	logs.Warn("SDK::CancelDownload()", zap.Int64("MsgID", msgID), zap.Error(err))
	// 	return
	// }
	//
	// r.fileCtrl.DeleteFromQueue(fs.MessageID, domain.RequestStatusCanceled)
	//
	// repo.Files.UpdateFileStatus(msgID, domain.RequestStatusCanceled)
}

// CancelUpload cancel upload
func (r *River) CancelUpload(clusterID int32, fileID int64, accessHash uint64) {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("CancelUpload", time.Now().Sub(startTime))
	}()

	// TODO:: implement it
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
