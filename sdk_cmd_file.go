package riversdk

import (
	"encoding/json"
	"git.ronaksoftware.com/ronak/riversdk/msg/ext"
	fileCtrl "git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_file"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	mon "git.ronaksoftware.com/ronak/riversdk/pkg/monitoring"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"go.uber.org/zap"
	"os"
	"strings"
	"time"
)

// GetStatus returns file status
// TODO :: change response to protobuff
func (r *River) GetFileStatus(msgID int64) string {
	status, progress, filePath := getFileStatus(msgID)
	x := struct {
		Status   int32   `json:"status"`
		Progress float64 `json:"progress"`
		Filepath string  `json:"filepath"`
	}{
		Status:   int32(status),
		Progress: progress,
		Filepath: filePath,
	}

	buff, _ := json.Marshal(x)

	return string(buff)
}
func getFileStatus(msgID int64) (status domain.RequestStatus, progress float64, filePath string) {
	fs, err := repo.Files.GetStatus(msgID)
	if err == nil && fs != nil {
		// file is in-progress state
		// double check
		if fs.IsCompleted {
			go repo.Files.DeleteStatus(fs.MessageID)
		}
		status = domain.RequestStatus(fs.RequestStatus)
		filePath = fs.FilePath
		if fs.TotalParts > 0 {
			partList := domain.MInt64B{}
			json.Unmarshal(fs.PartList, &partList)
			processedParts := fs.TotalParts - int64(len(partList))
			progress = float64(processedParts) / float64(fs.TotalParts) * float64(100)
		}
	} else {
		filePath = getFilePath(msgID)
		if filePath != "" {
			// file exists so it means download completed
			status = domain.RequestStatusCompleted
			progress = 100
		} else {
			// file does not exist and its progress state does not exist too
			status = domain.RequestStatusNone
			progress = 0
			filePath = ""
		}
	}

	return
}
func getFilePath(msgID int64) string {
	m := repo.Messages.Get(msgID)
	if m == nil {
		return ""
	}

	switch m.MediaType {
	case msg.MediaTypeDocument:
		x := new(msg.MediaDocument)
		err := x.Unmarshal(m.Media)
		if err == nil {
			// check file existence

			filePath := fileCtrl.GetFilePath(x.Doc.MimeType, x.Doc.ID)
			if _, err = os.Stat(filePath); os.IsNotExist(err) {
				filePath = ""
			}
			return filePath
		}
	default:
		// Probably this is pendingMessage so MediaData is ClientSendMessageMedia
		x := new(msg.ClientSendMessageMedia)
		err := x.Unmarshal(m.Media)
		if err == nil {
			// check file existence
			filePath := x.FilePath
			if _, err = os.Stat(filePath); os.IsNotExist(err) {
				filePath = ""
			}
			return filePath
		}
	}
	return ""
}

// FileDownload add download request to file controller queue
func (r *River) FileDownload(msgID int64) {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("FileDownload", time.Now().Sub(startTime))
	}()
	status, progress, filePath := getFileStatus(msgID)
	logs.Debug("SDK::FileDownload() current file progress status",
		zap.String("Status", status.ToString()),
		zap.Float64("Progress", progress),
		zap.String("FilePath", filePath),
	)
	m := repo.Messages.Get(msgID)
	if m == nil {
		logs.Warn("SDK::FileDownload()", zap.Int64("Message does not exist", msgID))
		return
	}

	switch status {
	case domain.RequestStatusNone:
		r.fileCtrl.Download(m)
	case domain.RequestStatusCompleted:
		r.onFileDownloadCompleted(m.ID, filePath, domain.FileStateExistedDownload)
	case domain.RequestStatusPaused, domain.RequestStatusCanceled, domain.RequestStatusError:
		r.fileCtrl.Download(m)
	case domain.RequestStatusInProgress:
	}

	fs, err := repo.Files.GetStatus(msgID)
	if err == nil && fs != nil {

	} else {
		m := repo.Messages.Get(msgID)
		r.fileCtrl.Download(m)
	}
}

// PauseDownload pause download
func (r *River) PauseDownload(msgID int64) {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("PauseDownload", time.Now().Sub(startTime))
	}()
	fs, err := repo.Files.GetStatus(msgID)
	if err != nil {
		logs.Warn("SDK::PauseDownload()", zap.Int64("MsgID", msgID), zap.Error(err))
		return
	}

	r.fileCtrl.DeleteFromQueue(fs.MessageID, domain.RequestStatusPaused)

	repo.Files.UpdateFileStatus(msgID, domain.RequestStatusPaused)
}

// CancelDownload cancel download
func (r *River) CancelDownload(msgID int64) {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("CancelDownload", time.Now().Sub(startTime))
	}()
	fs, err := repo.Files.GetStatus(msgID)
	if err != nil {
		logs.Warn("SDK::CancelDownload()", zap.Int64("MsgID", msgID), zap.Error(err))
		return
	}

	r.fileCtrl.DeleteFromQueue(fs.MessageID, domain.RequestStatusCanceled)

	repo.Files.UpdateFileStatus(msgID, domain.RequestStatusCanceled)
}

// PauseUpload pause upload
func (r *River) PauseUpload(msgID int64) {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("PauseUpload", time.Now().Sub(startTime))
	}()
	fs, err := repo.Files.GetStatus(msgID)
	if err != nil {
		logs.Warn("SDK::PauseUpload()", zap.Int64("MsgID", msgID), zap.Error(err))
		return
	}

	r.fileCtrl.DeleteFromQueue(fs.MessageID, domain.RequestStatusPaused)

	repo.Files.UpdateFileStatus(msgID, domain.RequestStatusPaused)
	// repo.MessagesPending.Delete(fs.MessageID)

}

// CancelUpload cancel upload
func (r *River) CancelUpload(msgID int64) {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("CancelUpload", time.Now().Sub(startTime))
	}()
	fs, err := repo.Files.GetStatus(msgID)
	if err != nil {
		logs.Warn("SDK::CancelUpload()", zap.Int64("MsgID", msgID), zap.Error(err))
		return
	}
	r.fileCtrl.DeleteFromQueue(fs.MessageID, domain.RequestStatusCanceled)

	repo.Files.DeleteStatus(msgID)
	repo.PendingMessages.Delete(fs.MessageID)

}

// AccountUploadPhoto upload user profile photo
func (r *River) AccountUploadPhoto(filePath string) (msgID int64) {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("AccountUploadPhoto", time.Now().Sub(startTime))
	}()
	// TOF
	msgID = domain.SequentialUniqueID()
	fileID := domain.SequentialUniqueID()

	// support IOS file path
	if strings.HasPrefix(filePath, "file://") {
		filePath = filePath[7:]
	}

	file, err := os.Open(filePath)
	if err != nil {
		logs.Warn("SDK::AccountUploadPhoto()", zap.Error(err))
		return 0
	}
	fileInfo, err := file.Stat()
	if err != nil {
		logs.Warn("SDK::AccountUploadPhoto()", zap.Error(err))
		return 0
	}

	totalSize := fileInfo.Size() // size in Byte
	// if totalSize > domain.FileMaxPhotoSize {
	// 	log.Error("SDK::AccountUploadPhoto()", zap.Error(errors.New("max allowed file size is 1 MB")))
	// 	return 0
	// }

	state := fileCtrl.NewFile(msgID, fileID, 0, totalSize, filePath, domain.FileStateUploadAccountPhoto, 0, 0, 0, r.onFileProgressChanged)

	r.fileCtrl.AddToQueue(state)

	return msgID
}

// AccountGetPhoto_Big download user profile picture
func (r *River) AccountGetPhotoBig(userID int64) string {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("AccountGetPhotoBig", time.Now().Sub(startTime))
	}()
	user := repo.Users.Get(userID)
	if user == nil || user.Photo == nil {
		return ""
	}

	filePath := fileCtrl.GetAccountAvatarPath(user.ID, user.Photo.PhotoBig.FileID)

	// check if file exist
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return r.downloadAccountPhoto(user.ID, user.Photo, true)
	}
	return filePath

}

// AccountGetPhoto_Small download user profile picture thumbnail
func (r *River) AccountGetPhotoSmall(userID int64) string {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("AccountGetPhotoSmall", time.Now().Sub(startTime))
	}()
	user := repo.Users.Get(userID)
	if user == nil || user.Photo == nil {
		return ""
	}

	filePath := fileCtrl.GetAccountAvatarPath(user.ID, user.Photo.PhotoSmall.FileID)

	// check if file exist
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return r.downloadAccountPhoto(user.ID, user.Photo, true)
	}
	return filePath
}

// downloadAccountPhoto this function is sync
func (r *River) downloadAccountPhoto(userID int64, photo *msg.UserPhoto, isBig bool) string {
	logs.Debug("SDK::downloadAccountPhoto",
		zap.Int64("userID", userID),
		zap.Bool("IsBig", isBig),
		zap.Int64("PhotoBig.FileID", photo.PhotoBig.FileID),
		zap.Uint64("PhotoBig.AccessHash", photo.PhotoBig.AccessHash),
		zap.Int32("PhotoBig.ClusterID", photo.PhotoBig.ClusterID),
		zap.Int64("SmallBig.FileID", photo.PhotoSmall.FileID),
		zap.Uint64("SmallBig.AccessHash", photo.PhotoSmall.AccessHash),
		zap.Int32("SmallBig.ClusterID", photo.PhotoSmall.ClusterID),
	)

	// send Download request
	filePath, err := r.fileCtrl.DownloadAccountPhoto(userID, photo, isBig)
	if err != nil {
		logs.Debug("SDK::downloadAccountPhoto() error", zap.Error(err))
		return ""
	}
	return filePath
}

// GroupUploadPhoto upload group profile photo
func (r *River) GroupUploadPhoto(groupID int64, filePath string) (msgID int64) {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("GroupUploadPhoto", time.Now().Sub(startTime))
	}()
	// TOF
	msgID = domain.SequentialUniqueID()
	fileID := domain.SequentialUniqueID()

	// support IOS file path
	if strings.HasPrefix(filePath, "file://") {
		filePath = filePath[7:]
	}

	file, err := os.Open(filePath)
	if err != nil {
		logs.Warn("SDK::GroupUploadPhoto()", zap.Error(err))
		return 0
	}
	fileInfo, err := file.Stat()
	if err != nil {
		logs.Warn("SDK::GroupUploadPhoto()", zap.Error(err))
		return 0
	}

	// // fileName := fileInfo.Name()
	totalSize := fileInfo.Size() // size in Byte
	// if totalSize > domain.FileMaxPhotoSize {
	// 	log.Error("SDK::GroupUploadPhoto()", zap.Error(errors.New("max allowed file size is 1 MB")))
	// 	return 0
	// }

	state := fileCtrl.NewFile(msgID, fileID, groupID, totalSize, filePath, domain.FileStateUploadGroupPhoto, 0, 0, 0, r.onFileProgressChanged)

	r.fileCtrl.AddToQueue(state)

	return msgID
}

// GroupGetPhoto_Big download group profile picture
func (r *River) GroupGetPhotoBig(groupID int64) string {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("GroupGetPhotoBig", time.Now().Sub(startTime))
	}()
	group := repo.Groups.Get(groupID)
	if group == nil || group.Photo == nil {
		return ""
	}

	filePath := fileCtrl.GetGroupAvatarPath(group.ID, group.Photo.PhotoBig.FileID)
	// check if file exist
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return r.downloadGroupPhoto(groupID, group.Photo, true)
	}
	return filePath
}

// GroupGetPhoto_Small download group profile picture thumbnail
func (r *River) GroupGetPhotoSmall(groupID int64) string {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("GroupGetPhotoSmall", time.Now().Sub(startTime))
	}()
	group := repo.Groups.Get(groupID)
	if group == nil {
		return ""
	}
	if group.Photo == nil {
		return ""
	}

	filePath := fileCtrl.GetGroupAvatarPath(group.ID, group.Photo.PhotoSmall.FileID)
	// check if file exist
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return r.downloadGroupPhoto(groupID, group.Photo, true)
	}
	return filePath
}

// this function is sync
func (r *River) downloadGroupPhoto(groupID int64, photo *msg.GroupPhoto, isBig bool) string {
	logs.Debug("SDK::downloadGroupPhoto",
		zap.Int64("userID", groupID),
		zap.Bool("IsBig", isBig),
		zap.Int64("PhotoBig.FileID", photo.PhotoBig.FileID),
		zap.Uint64("PhotoBig.AccessHash", photo.PhotoBig.AccessHash),
		zap.Int32("PhotoBig.ClusterID", photo.PhotoBig.ClusterID),
		zap.Int64("SmallBig.FileID", photo.PhotoSmall.FileID),
		zap.Uint64("SmallBig.AccessHash", photo.PhotoSmall.AccessHash),
		zap.Int32("SmallBig.ClusterID", photo.PhotoSmall.ClusterID),
	)

	// send Download request
	filePath, err := r.fileCtrl.DownloadGroupPhoto(groupID, photo, isBig)
	if err != nil {
		logs.Debug("SDK::downloadGroupPhoto() error", zap.Error(err))
		return ""
	}
	return filePath
}

// FileDownloadThumbnail download file thumbnail
func (r *River) FileDownloadThumbnail(msgID int64) string {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("FileDownloadThumbnail", time.Now().Sub(startTime))
	}()

	// its pending message
	if msgID < 0 {
		return r.downloadPendingThumbnail(msgID)
	}

	m := repo.Messages.Get(msgID)
	if m == nil {
		logs.Error("SDK::FileDownloadThumbnail() message does not exist")
		return ""
	}
	filePath := ""
	docID := int64(0)
	clusterID := int32(0)
	accessHash := uint64(0)
	version := int32(0)
	switch m.MediaType {
	case msg.MediaTypeEmpty:
		return ""
	case msg.MediaTypeDocument:
		x := new(msg.MediaDocument)
		_ = x.Unmarshal(m.Media)
		if x.Doc.Thumbnail == nil {
			return ""
		}
		docID = x.Doc.Thumbnail.FileID
		clusterID = x.Doc.Thumbnail.ClusterID
		accessHash = x.Doc.Thumbnail.AccessHash

		filePath = fileCtrl.GetThumbnailPath(docID, clusterID)
		// check if file exist
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			filePath, err = r.fileCtrl.DownloadThumbnail(docID, accessHash, clusterID, version)
		}
	case msg.MediaTypeContact:
	case msg.MediaTypePhoto:
	default:
		return ""
	}

	return filePath
}
func (r *River) downloadPendingThumbnail(msgID int64) string {
	pmsg := repo.PendingMessages.GetByID(msgID)
	if pmsg == nil {
		return ""
	}
	switch msg.InputMediaType(pmsg.MediaType) {
	case msg.InputMediaTypeDocument:
		// pending message media is a file that already has been uploaded so pending message media is InputMediaDocument
		doc := new(msg.InputMediaDocument)
		err := doc.Unmarshal(pmsg.Media)
		if err != nil {
			logs.Error("SDK::FileDownloadThumbnail() failed to unmarshal to InputMediaDocument", zap.Int64("PendingMsgID", msgID), zap.Error(err))
			return ""
		}

		filePath := fileCtrl.GetThumbnailPath(doc.Document.ID, doc.Document.ClusterID)
		// check if file exist
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			filePath, err = r.fileCtrl.DownloadThumbnail(doc.Document.ID, doc.Document.AccessHash, doc.Document.ClusterID, 0)
			return filePath
		}
		return filePath
	case msg.InputMediaTypeUploadedDocument:
		// pending message media is new upload so its type should be ClientSendMessageMedia
		clientMedia := new(msg.ClientSendMessageMedia)
		err := clientMedia.Unmarshal(pmsg.Media)
		if err != nil {
			logs.Error("SDK::FileDownloadThumbnail() failed to unmarshal to ClientSendMessageMedia", zap.Int64("PendingMsgID", msgID), zap.Error(err))
			return ""
		}
		return clientMedia.ThumbFilePath
	}
	return ""
}

// ClearCache removes files from client device, allMedia means clear all media types
// peerID 0 means all peers
func (r *River) ClearCache(peerID int64, mediaTypes string, allMedia bool) bool {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("ClearCache", time.Now().Sub(startTime))
	}()
	// var messageIDs []int64
	// clearDatabaseStatus := func() {
	// 	for k := range DatabaseStatus {
	// 		delete(DatabaseStatus, k)
	// 	}
	// }
	// defer clearDatabaseStatus()
	// if allMedia {
	// 	// peerID = 0 means all peers
	// 	// all peers and all media types
	// 	if peerID == 0 {
	// 		for _, mediaData := range DatabaseStatus {
	// 			for _, mediaInfo := range mediaData {
	// 				messageIDs = append(messageIDs, mediaInfo.MessageIDs...)
	// 			}
	// 		}
	// 	} else {
	// 		// all media types of a specific peer
	// 		for _, mediaInfo := range DatabaseStatus[peerID] {
	// 			messageIDs = append(messageIDs, mediaInfo.MessageIDs...)
	// 		}
	// 	}
	// } else {
	// 	// all peers with specific media types
	// 	if peerID == 0 {
	// 		for _, mediaData := range DatabaseStatus {
	// 			mediaTypeSlices := strings.Split(mediaTypes, ",")
	// 			for _, mediaType := range mediaTypeSlices {
	// 				castedType, _ := strconv.Atoi(mediaType)
	// 				messageIDs = append(messageIDs, mediaData[msg.DocumentAttributeType(castedType)].MessageIDs...)
	// 			}
	// 		}
	// 	} else {
	// 		// specific peer with specific media type
	// 		mediaInfo := DatabaseStatus[peerID]
	// 		mediaTypeSlices := strings.Split(mediaTypes, ",")
	// 		for _, mediaType := range mediaTypeSlices {
	// 			castedType, _ := strconv.Atoi(mediaType)
	// 			messageIDs = append(messageIDs, mediaInfo[msg.DocumentAttributeType(castedType)].MessageIDs...)
	// 		}
	// 	}
	// }
	//
	// logs.Info("River::ClearCache",
	// 	zap.Any("peerID", peerID),
	// 	zap.String("mediaTypes", mediaTypes),
	// 	zap.Bool("all", allMedia),
	// 	zap.Any("DatabaseStatus Map", DatabaseStatus),
	// )
	//
	// logs.Debug("ClearCache", zap.Int64s("messageIDs", messageIDs))
	//
	// if filePaths, err := repo.Files.ClearMedia(messageIDs); err != nil {
	// 	logs.Debug("River::ClearCache",
	// 		zap.String("clear media error", err.Error()),
	// 	)
	// 	return false
	// } else {
	// 	logs.Debug("ClearCache", zap.Strings("media paths", filePaths))
	// 	err = r.fileCtrl.ClearFiles(filePaths)
	// 	if err != nil {
	// 		logs.Debug("River::ClearCache",
	// 			zap.String("clear files error", err.Error()),
	// 		)
	// 		return false
	// 	}
	// }
	return true
}

// GetGetDBStatus returns message IDs and total size of each media stored in user's database
func (r *River) GetDBStatus(delegate RequestDelegate) {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("GetDBStatus", time.Now().Sub(startTime))
	}()
	delegate.OnTimeout(domain.ErrDoesNotExists)
	// res := msg.DBMediaInfo{}
	// if GetDBStatusIsRunning {
	// 	err := errors.New("GetDBStatus is running")
	// 	if delegate != nil {
	// 		delegate.OnTimeout(err)
	// 	}
	// 	return
	// }
	// GetDBStatusIsRunning = true
	// for k := range DatabaseStatus {
	// 	delete(DatabaseStatus, k)
	// }
	// logs.Debug("DatabaseStatus Must be Empty", zap.Any("", fmt.Sprintf("%+v", DatabaseStatus)))
	// peerMediaSizeMap, err := repo.Files.GetDBStatus()
	// if err != nil {
	// 	GetDBStatusIsRunning = false
	// 	logs.Error(err.Error())
	// 	delegate.OnTimeout(err)
	// 	return
	// }
	// logs.Debug("peerMediaSizeMap", zap.Any("peerMediaSizeMap", peerMediaSizeMap))
	// peerMediaInfo := make([]*msg.PeerMediaInfo, 0)
	// for peerID, mediaInfoMap := range peerMediaSizeMap {
	// 	mediaSize := make([]*msg.MediaSize, 0)
	// 	for mediaType, mediaInfo := range mediaInfoMap {
	// 		mediaSize = append(mediaSize, &msg.MediaSize{MediaType: int32(mediaType), TotalSize: mediaInfo.Size})
	// 	}
	// 	peerMediaInfo = append(peerMediaInfo, &msg.PeerMediaInfo{PeerID: peerID, Media: mediaSize})
	// }
	// res.MediaInfo = peerMediaInfo
	// logs.Debug("MediaInfo", zap.String("", fmt.Sprintf("%+v", res.MediaInfo)))
	// pBytes, _ := res.Marshal()
	// if delegate != nil {
	// 	delegate.OnComplete(pBytes)
	// }
	// GetDBStatusIsRunning = false
	// DatabaseStatus = peerMediaSizeMap
}
