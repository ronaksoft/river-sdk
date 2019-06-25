package riversdk

import (
	"encoding/json"
	fileCtrl "git.ronaksoftware.com/ronak/riversdk/pkg/ctrl_file"
	"os"
	"strconv"
	"strings"

	"git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"go.uber.org/zap"
)

// GetFileStatus returns file status
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

	fs, err := repo.Files.GetFileStatus(msgID)
	if err == nil && fs != nil {
		// file is inprogress state
		// double check

		if fs.IsCompleted {
			go repo.Files.DeleteFileStatus(fs.MessageID)
		}
		status = domain.RequestStatus(fs.RequestStatus)
		filePath = fs.FilePath
		if fs.TotalParts > 0 {
			partList := domain.MInt64B{}
			json.Unmarshal(fs.PartList, &partList)
			processedParts := fs.TotalParts - int64(len(partList))
			progress = (float64(processedParts) / float64(fs.TotalParts) * float64(100))
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
	m := repo.Messages.GetMessage(msgID)
	if m != nil {

		switch m.MediaType {
		case msg.MediaTypeDocument:
			x := new(msg.MediaDocument)
			err := x.Unmarshal(m.Media)
			if err == nil {
				// check file existence
				filePath := repo.Files.GetFilePath(m.ID, x.Doc.ID)
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
	}
	return ""
}

// FileDownload add download request to filemanager queue
func (r *River) FileDownload(msgID int64) {
	status, progress, filePath := getFileStatus(msgID)
	logs.Debug("SDK::FileDownload() current file progress status",
		zap.String("Status", status.ToString()),
		zap.Float64("Progress", progress),
		zap.String("FilePath", filePath),
	)
	m := repo.Messages.GetMessage(msgID)
	if m == nil {
		logs.Warn("SDK::FileDownload()", zap.Int64("Message does not exist", msgID))
		return
	}

	switch status {
	case domain.RequestStatusNone:
		fileCtrl.Ctx().Download(m)
	case domain.RequestStatusInProgress:
		// already downloading
		// filemanager.Ctx().Download(m)
	case domain.RequestStatusCompleted:
		r.onFileDownloadCompleted(m.ID, filePath, domain.FileStateExistedDownload)
	case domain.RequestStatusPaused:
		fileCtrl.Ctx().Download(m)
	case domain.RequestStatusCanceled:
		fileCtrl.Ctx().Download(m)
	case domain.RequestStatusError:
		fileCtrl.Ctx().Download(m)
	}

	fs, err := repo.Files.GetFileStatus(msgID)
	if err == nil && fs != nil {

	} else {
		m := repo.Messages.GetMessage(msgID)
		fileCtrl.Ctx().Download(m)
	}
}

// PauseDownload pause download
func (r *River) PauseDownload(msgID int64) {
	fs, err := repo.Files.GetFileStatus(msgID)
	if err != nil {
		logs.Warn("SDK::PauseDownload()", zap.Int64("MsgID", msgID), zap.Error(err))
		return
	}
	fileCtrl.Ctx().DeleteFromQueue(fs.MessageID, domain.RequestStatusPaused)
	_ = repo.Files.UpdateFileStatus(msgID, domain.RequestStatusPaused)
}

// CancelDownload cancel download
func (r *River) CancelDownload(msgID int64) {
	fs, err := repo.Files.GetFileStatus(msgID)
	if err != nil {
		logs.Warn("SDK::CancelDownload()", zap.Int64("MsgID", msgID), zap.Error(err))
		return
	}
	fileCtrl.Ctx().DeleteFromQueue(fs.MessageID, domain.RequestStatusCanceled)
	_ = repo.Files.UpdateFileStatus(msgID, domain.RequestStatusCanceled)
}

// PauseUpload pause upload
func (r *River) PauseUpload(msgID int64) {
	fs, err := repo.Files.GetFileStatus(msgID)
	if err != nil {
		logs.Warn("SDK::PauseUpload()", zap.Int64("MsgID", msgID), zap.Error(err))
		return
	}
	fileCtrl.Ctx().DeleteFromQueue(fs.MessageID, domain.RequestStatusPaused)
	_ = repo.Files.UpdateFileStatus(msgID, domain.RequestStatusPaused)
	// repo.MessagesPending.DeletePendingMessage(fs.MessageID)

}

// CancelUpload cancel upload
func (r *River) CancelUpload(msgID int64) {
	fs, err := repo.Files.GetFileStatus(msgID)
	if err != nil {
		logs.Warn("SDK::CancelUpload()", zap.Int64("MsgID", msgID), zap.Error(err))
		return
	}
	fileCtrl.Ctx().DeleteFromQueue(fs.MessageID, domain.RequestStatusCanceled)
	_ = repo.Files.DeleteFileStatus(msgID)
	_ = repo.PendingMessages.DeletePendingMessage(fs.MessageID)

}

// AccountUploadPhoto upload user profile photo
func (r *River) AccountUploadPhoto(filePath string) (msgID int64) {
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

	// // fileName := fileInfo.Name()
	totalSize := fileInfo.Size() // size in Byte
	// if totalSize > domain.FileMaxPhotoSize {
	// 	log.Error("SDK::AccountUploadPhoto()", zap.Error(errors.New("max allowed file size is 1 MB")))
	// 	return 0
	// }

	state := fileCtrl.NewFileStatus(msgID, fileID, 0, totalSize, filePath, domain.FileStateUploadAccountPhoto, 0, 0, 0, r.onFileProgressChanged)

	fileCtrl.Ctx().AddToQueue(state)

	return msgID
}

// AccountGetPhoto_Big download user profile picture
func (r *River) AccountGetPhotoBig(userID int64) string {
	user := repo.Users.Get(userID)
	if user != nil {
		if user.Photo != nil {
			dtoPhoto := repo.Users.GetPhoto(userID, user.Photo.PhotoID)
			if dtoPhoto != nil {
				if dtoPhoto.BigFilePath != "" {
					// check if file exist
					if _, err := os.Stat(dtoPhoto.BigFilePath); os.IsNotExist(err) {
						return downloadAccountPhoto(userID, user.Photo, true)
					}
					// check if fileID is changed re-download
					strFileID := strconv.FormatInt(dtoPhoto.BigFileID, 10)
					if strings.Index(dtoPhoto.BigFilePath, strFileID) < 0 {
						return downloadAccountPhoto(user.ID, user.Photo, true)
					}
					return dtoPhoto.BigFilePath
				}
				return downloadAccountPhoto(userID, user.Photo, true)
			}
			return downloadAccountPhoto(userID, user.Photo, true)

		}
		logs.Warn("SDK::AccountGetPhoto_Big() user photo is null")
		return ""
	}
	return ""
}

// AccountGetPhoto_Small download user profile picture thumbnail
func (r *River) AccountGetPhotoSmall(userID int64) string {
	user := repo.Users.Get(userID)
	if user != nil {
		if user.Photo != nil {
			dtoPhoto := repo.Users.GetPhoto(userID, user.Photo.PhotoID)
			if dtoPhoto != nil {
				if dtoPhoto.SmallFilePath != "" {
					// check if file exist
					if _, err := os.Stat(dtoPhoto.SmallFilePath); os.IsNotExist(err) {
						return downloadAccountPhoto(userID, user.Photo, false)
					}

					// check if fileID is changed re-download
					strFileID := strconv.FormatInt(dtoPhoto.SmallFileID, 10)
					if strings.Index(dtoPhoto.SmallFilePath, strFileID) < 0 {
						return downloadAccountPhoto(user.ID, user.Photo, true)
					}

					return dtoPhoto.SmallFilePath
				}
				return downloadAccountPhoto(userID, user.Photo, false)

			}
			return downloadAccountPhoto(userID, user.Photo, false)
		}
		return ""
	}
	return ""
}

// downloadAccountPhoto this function is sync
func downloadAccountPhoto(userID int64, photo *msg.UserPhoto, isBig bool) string {
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
	filePath, err := fileCtrl.Ctx().DownloadAccountPhoto(userID, photo, isBig)
	if err != nil {
		logs.Error("SDK::downloadAccountPhoto() error", zap.Error(err))
		return ""
	}
	return filePath
}

// GroupUploadPhoto upload group profile photo
func (r *River) GroupUploadPhoto(groupID int64, filePath string) (msgID int64) {

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

	state := fileCtrl.NewFileStatus(msgID, fileID, groupID, totalSize, filePath, domain.FileStateUploadGroupPhoto, 0, 0, 0, r.onFileProgressChanged)

	fileCtrl.Ctx().AddToQueue(state)

	return msgID
}

// GroupGetPhoto_Big download group profile picture
func (r *River) GroupGetPhotoBig(groupID int64) string {

	group, err := repo.Groups.GetGroupDTO(groupID)
	if err == nil && group != nil {
		if group.Photo != nil {
			groupPhoto := new(msg.GroupPhoto)
			err = groupPhoto.Unmarshal(group.Photo)
			if err != nil {
				logs.Error("SDK::GroupGetPhoto_Big() failed to unmarshal GroupPhoto", zap.Error(err))
				return ""
			}
			if group.BigFilePath != "" {
				// check if file exist
				if _, err := os.Stat(group.BigFilePath); os.IsNotExist(err) {
					return downloadGroupPhoto(groupID, groupPhoto, true)
				}
				// check if fileID is changed redownload
				strFileID := strconv.FormatInt(groupPhoto.PhotoBig.FileID, 10)
				if strings.Index(group.BigFilePath, strFileID) < 0 {
					return downloadGroupPhoto(groupID, groupPhoto, true)
				}
				return group.BigFilePath

			}
			return downloadGroupPhoto(groupID, groupPhoto, true)

		}
		logs.Error("SDK::GroupGetPhoto_Big() group photo is null")
		return ""
	}
	return ""
}

// GroupGetPhoto_Small download group profile picture thumbnail
func (r *River) GroupGetPhotoSmall(groupID int64) string {

	group, err := repo.Groups.GetGroupDTO(groupID)
	if err == nil && group != nil {
		if group.Photo != nil {
			groupPhoto := new(msg.GroupPhoto)
			err = groupPhoto.Unmarshal(group.Photo)
			if err != nil {
				logs.Error("SDK::GroupGetPhoto_Small() failed to unmarshal GroupPhoto", zap.Error(err))
				return ""
			}
			if group.SmallFilePath != "" {

				// check if file exist
				if _, err := os.Stat(group.SmallFilePath); os.IsNotExist(err) {
					return downloadGroupPhoto(groupID, groupPhoto, false)
				}

				// check if fileID is changed redownload
				strFileID := strconv.FormatInt(groupPhoto.PhotoSmall.FileID, 10)
				if strings.Index(group.SmallFilePath, strFileID) < 0 {
					return downloadGroupPhoto(groupID, groupPhoto, false)
				}

				return group.SmallFilePath

			}
			return downloadGroupPhoto(groupID, groupPhoto, false)

		}
		logs.Error("SDK::GroupGetPhoto_Small() group photo is null")
		return ""
	}
	return ""
}

// this function is sync
func downloadGroupPhoto(groupID int64, photo *msg.GroupPhoto, isBig bool) string {
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
	filePath, err := fileCtrl.Ctx().DownloadGroupPhoto(groupID, photo, isBig)
	if err != nil {
		logs.Error("SDK::downloadGroupPhoto() error", zap.Error(err))
		return ""
	}
	return filePath
}

// FileDownloadThumbnail download file thumbnail
func (r *River) FileDownloadThumbnail(msgID int64) string {
	// its pending message
	if msgID < 0 {
		pmsg, err := repo.PendingMessages.GetPendingMessageByID(msgID)
		if err != nil {
			logs.Error("SDK::FileDownloadThumbnail()", zap.Int64("PendingMsgID", msgID), zap.Error(err))
			return ""
		}
		switch msg.InputMediaType(pmsg.MediaType) {
		case msg.InputMediaTypeEmpty:
			// NOT IMPLEMENTED
		case msg.InputMediaTypeUploadedPhoto:
			// NOT IMPLEMENTED
		case msg.InputMediaTypePhoto:
			// NOT IMPLEMENTED
		case msg.InputMediaTypeGeoLocation:
			// NOT IMPLEMENTED
		case msg.InputMediaTypeContact:
			// NOT IMPLEMENTED
		case msg.InputMediaTypeDocument:
			// pending message media is a file that already has been uploaded so pending message media is InputMediaDocument
			doc := new(msg.InputMediaDocument)
			err := doc.Unmarshal(pmsg.Media)
			if err != nil {
				logs.Error("SDK::FileDownloadThumbnail() failed to unmarshal to InputMediaDocument", zap.Int64("PendingMsgID", msgID), zap.Error(err))
				return ""
			}
			// Get userMessage ID by DocumentID and extract thumbnail path from it
			existedDocumentFile, err := repo.Files.GetFileByDocumentID(doc.Document.ID)
			if err != nil {
				logs.Error("SDK::FileDownloadThumbnail() failed to fetch GetFileByDocumentID", zap.Int64("PendingMsgID", msgID), zap.Int64("DocID", doc.Document.ID), zap.Error(err))
				return ""
			}
			return r.FileDownloadThumbnail(existedDocumentFile.MessageID)

		case msg.InputMediaTypeUploadedDocument:
			// pending message media is new upload so its type should be ClientSendMessageMedia
			clientMedia := new(msg.ClientSendMessageMedia)
			err := clientMedia.Unmarshal(pmsg.Media)
			if err != nil {
				logs.Error("SDK::FileDownloadThumbnail() failed to unmarshal to ClientSendMessageMedia", zap.Int64("PendingMsgID", msgID), zap.Error(err))
				return ""
			}
			return clientMedia.ThumbFilePath
		case msg.Reserved1:
			// NOT IMPLEMENTED
		case msg.Reserved2:
			// NOT IMPLEMENTED
		case msg.Reserved3:
			// NOT IMPLEMENTED
		case msg.Reserved4:
			// NOT IMPLEMENTED
		case msg.Reserved5:
			// NOT IMPLEMENTED
		}
	}

	m := repo.Messages.GetMessage(msgID)
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
		// TODO:: implement it
	case msg.MediaTypePhoto:
		// // TODO:: implement it
	case msg.MediaTypeDocument:
		x := new(msg.MediaDocument)
		x.Unmarshal(m.Media)
		if x.Doc.Thumbnail != nil {
			docID = x.Doc.Thumbnail.FileID
			clusterID = x.Doc.Thumbnail.ClusterID
			accessHash = x.Doc.Thumbnail.AccessHash
			// version = x.Doc.Thumbnail.Version
		} else {
			logs.Warn("SDK::FileDownloadThumbnail() Message does not have thumbnail", zap.Int64("MsgID", msgID))
			return filePath
		}

	case msg.MediaTypeContact:
		// TODO:: implement it
	default:
		logs.Error("SDK::FileDownloadThumbnail() Invalid SharedMediaType")
	}

	dto, err := repo.Files.GetFile(msgID)
	if err != nil {
		path, err := fileCtrl.Ctx().DownloadThumbnail(m.ID, docID, accessHash, clusterID, version)
		if err != nil {
			logs.Error("SDK::FileDownloadThumbnail()-> DownloadThumbnail()", zap.Error(err))
		}
		return path
	}
	_, err = os.Stat(dto.ThumbFilePath)
	if os.IsNotExist(err) {
		path, err := fileCtrl.Ctx().DownloadThumbnail(m.ID, docID, accessHash, clusterID, version)
		if err != nil {
			logs.Error("SDK::FileDownloadThumbnail()-> DownloadThumbnail()", zap.Error(err))
		}
		return path
	}
	return dto.ThumbFilePath
}
