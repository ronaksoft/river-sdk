package riversdk

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/filemanager"
	"git.ronaksoftware.com/ronak/riversdk/logs"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"git.ronaksoftware.com/ronak/riversdk/repo"
	"go.uber.org/zap"
)

// CancelRequest remove given requestID callbacks&delegates and if its not processed by queue we skip it on queue distributor
func (r *River) CancelRequest(requestID int64) {

	// Remove delegate
	r.delegateMutex.Lock()
	delete(r.delegates, int64(requestID))
	r.delegateMutex.Unlock()

	// Remove callback
	domain.RemoveRequestCallback(uint64(requestID))

	// Remove from goque levelDB
	// the goque pkg does not support this
	r.queueCtrl.CancelRequest(requestID)

}

// DeletePendingMessage removes pending message from DB
func (r *River) DeletePendingMessage(id int64) (isSuccess bool) {
	err := repo.Ctx().PendingMessages.DeletePendingMessage(id)
	isSuccess = err == nil
	return
}

// RetryPendingMessage puts pending message again in command queue to re send it
func (r *River) RetryPendingMessage(id int64) (isSuccess bool) {
	pmsg, err := repo.Ctx().PendingMessages.GetPendingMessageByID(id)
	if err != nil {
		logs.Debug("River::RetryPendingMessage()",
			zap.String("GetPendingMessageByID", err.Error()),
		)
		isSuccess = false
		return
	}
	req := new(msg.MessagesSend)
	pmsg.MapToMessageSend(req)

	buff, _ := req.Marshal()
	r.queueCtrl.ExecuteCommand(uint64(req.RandomID), msg.C_MessagesSend, buff, nil, nil, true)
	isSuccess = true
	logs.Debug("River::RetryPendingMessage() Request enqueued")

	return
}

// GetNetworkStatus returns NetworkController status
func (r *River) GetNetworkStatus() int32 {
	return int32(r.networkCtrl.Quality())
}

// GetSyncStatus returns SyncController status
func (r *River) GetSyncStatus() int32 {

	logs.Debug("River::GetSyncStatus()",
		zap.String("syncStatus", domain.SyncStatusName[r.syncCtrl.Status()]),
	)
	return int32(r.syncCtrl.Status())
}

// Logout drop queue & database , etc ...
func (r *River) Logout() (int64, error) {

	dataDir, err := r.queueCtrl.DropQueue()

	if err != nil {
		logs.Debug("River::Logout() failed to drop queue",
			zap.Error(err),
		)
	}

	// drop and recreate database
	err = repo.Ctx().ReinitiateDatabase()
	if err != nil {
		logs.Debug("River::Logout() failed to re initiate database",
			zap.Error(err),
		)
	}

	// open queue
	err = r.queueCtrl.OpenQueue(dataDir)
	if err != nil {
		logs.Debug("River::Logout() failed to re open queue",
			zap.Error(err),
		)
	}

	// TODO : send logout request to server
	requestID := domain.RandomInt63()
	timeoutCallback := func() {
		err = domain.ErrRequestTimeout
		r.releaseDelegate(requestID)

		r.clearSystemConfig()
		r.networkCtrl.Reconnect()
		r.syncCtrl.ClearUpdateID()
	}
	successCallback := func(envelope *msg.MessageEnvelope) {
		r.releaseDelegate(requestID)

		r.clearSystemConfig()
		r.networkCtrl.Reconnect()
		r.syncCtrl.ClearUpdateID()
	}

	req := new(msg.AuthLogout)
	buff, _ := req.Marshal()
	err = r.queueCtrl.ExecuteRealtimeCommand(uint64(requestID), msg.C_AuthLogout, buff, timeoutCallback, successCallback, true, false)
	if err != nil {
		r.releaseDelegate(requestID)
	}

	if r.mainDelegate != nil && r.mainDelegate.OnSessionClosed != nil {
		r.mainDelegate.OnSessionClosed(0)
	}

	return requestID, err
}

// UISettingGet fetch from key/value storage for UI settings
func (r *River) UISettingGet(key string) string {
	val, err := repo.Ctx().UISettings.Get(key)
	if err != nil {
		logs.Info("River::UISettingsGet()",
			zap.String("Error", err.Error()),
		)
	}
	return val
}

// UISettingPut save to key/value storage for UI settings
func (r *River) UISettingPut(key, value string) bool {
	err := repo.Ctx().UISettings.Put(key, value)
	if err != nil {
		logs.Info("River::UISettingsPut()",
			zap.String("Error", err.Error()),
		)
	}
	return err == nil
}

// UISettingDelete remove from key/value storage for UI settings
func (r *River) UISettingDelete(key string) bool {
	err := repo.Ctx().UISettings.Delete(key)
	if err != nil {
		logs.Info("River::UISettingsDelete()",
			zap.String("Error", err.Error()),
		)
	}
	return err == nil
}

// SearchContacts searchs contacts
func (r *River) SearchContacts(requestID int64, searchPhrase string, delegate RequestDelegate) {
	res := new(msg.MessageEnvelope)
	res.Constructor = msg.C_ContactsMany
	res.RequestID = uint64(requestID)

	contacts := new(msg.ContactsMany)
	contacts.Users, contacts.Contacts = repo.Ctx().Users.SearchContacts(searchPhrase)

	res.Message, _ = contacts.Marshal()

	buff, _ := res.Marshal()
	if delegate.OnComplete != nil {
		delegate.OnComplete(buff)
	}
}

// GetRealTopMessageID returns max message id
func (r *River) GetRealTopMessageID(peerID int64, peerType int32) int64 {

	topMsgID, err := repo.Ctx().Messages.GetTopMessageID(peerID, peerType)
	if err != nil {
		logs.Debug("SDK::GetRealTopMessageID() => Messages.GetTopMessageID()", zap.String("Error", err.Error()))
		return -1
	}
	return topMsgID
}

// UpdateContactinfo update contact name
func (r *River) UpdateContactinfo(userID int64, firstName, lastName string) error {
	err := repo.Ctx().Users.UpdateContactinfo(userID, firstName, lastName)
	if err != nil {
		logs.Debug("SDK::UpdateContactinfo() => Users.UpdateContactInfo()", zap.String("Error", err.Error()))
	}
	return err
}

// SearchInDialogs search dialog title
func (r *River) SearchInDialogs(requestID int64, searchPhrase string, delegate RequestDelegate) {
	res := new(msg.MessageEnvelope)
	res.Constructor = msg.C_MessagesDialogs
	res.RequestID = uint64(requestID)

	dlgs := new(msg.MessagesDialogs)

	users := repo.Ctx().Users.SearchUsers(searchPhrase)
	groups := repo.Ctx().Groups.SearchGroups(searchPhrase)
	dlgs.Users = users
	dlgs.Groups = groups

	mDialogs := domain.MInt64B{}
	for _, v := range users {
		mDialogs[v.ID] = true
	}
	for _, v := range groups {
		mDialogs[v.ID] = true
	}

	dialogs := repo.Ctx().Dialogs.GetManyDialog(mDialogs.ToArray())
	dlgs.Dialogs = dialogs

	mMessages := domain.MInt64B{}
	for _, v := range dialogs {
		mMessages[v.TopMessageID] = true
	}
	dlgs.Messages = repo.Ctx().Messages.GetManyMessages(mMessages.ToArray())

	res.Message, _ = dlgs.Marshal()
	buff, _ := res.Marshal()
	if delegate.OnComplete != nil {
		delegate.OnComplete(buff)
	}
}

// GetGroupInputUser get group participant user
func (r *River) GetGroupInputUser(requestID int64, groupID int64, userID int64, delegate RequestDelegate) {
	res := new(msg.MessageEnvelope)
	res.Constructor = msg.C_InputUser
	res.RequestID = uint64(requestID)

	user := new(msg.InputUser)
	user.UserID = userID

	accessHash, err := repo.Ctx().Users.GetAccessHash(userID)
	if err != nil || accessHash == 0 {
		participant, err := repo.Ctx().Groups.GetParticipants(groupID)
		if err == nil {
			for _, p := range participant {
				if p.UserID == userID {
					accessHash = p.AccessHash
					break
				}
			}
		}
	}

	if accessHash == 0 {
		// get group full and get its access hash from its participants
		req := new(msg.GroupsGetFull)
		req.GroupID = groupID
		reqBytes, _ := req.Marshal()

		out := new(msg.MessageEnvelope)
		// Timeout Callback
		timeoutCB := func() {
			if delegate.OnTimeout != nil {
				delegate.OnTimeout(domain.ErrRequestTimeout)
			}
		}

		// Success Callback
		successCB := func(response *msg.MessageEnvelope) {
			if response.Constructor != msg.C_GroupFull {
				msg.ResultError(out, &msg.Error{Code: "00", Items: "response type is not GroupFull"})
				if delegate.OnComplete != nil {
					outBytes, _ := out.Marshal()
					delegate.OnComplete(outBytes)
				}
				return
			}

			groupFull := new(msg.GroupFull)
			err := groupFull.Unmarshal(response.Message)
			if err != nil {
				msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
				if delegate.OnComplete != nil {
					outBytes, _ := out.Marshal()
					delegate.OnComplete(outBytes)
				}
				return
			}

			for _, p := range groupFull.Participants {
				if p.UserID == userID {
					user.AccessHash = p.AccessHash
					break
				}
			}

			res.Message, _ = user.Marshal()
			resBytes, _ := res.Marshal()
			if delegate.OnComplete != nil {
				delegate.OnComplete(resBytes)
			}
		}
		// Send GroupsGetFull request to get user AccessHash
		r.queueCtrl.ExecuteRealtimeCommand(uint64(requestID), msg.C_GroupsGetFull, reqBytes, timeoutCB, successCB, true, false)

	} else {
		user.AccessHash = accessHash
		res.Message, _ = user.Marshal()

		buff, _ := res.Marshal()
		if delegate.OnComplete != nil {
			delegate.OnComplete(buff)
		}
	}
}

// GetFileStatus returns file status
// TODO :: change response to protobuff
func (r *River) GetFileStatus(msgID int64) string {

	status, progress, filePath := geFileStatus(msgID)
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

func getFilePath(msgID int64) string {
	m := repo.Ctx().Messages.GetMessage(msgID)
	if m != nil {

		switch m.MediaType {
		case msg.MediaTypeDocument:
			x := new(msg.MediaDocument)
			err := x.Unmarshal(m.Media)
			if err == nil {
				// check file existance
				filePath := repo.Ctx().Files.GetFilePath(m.ID, x.Doc.ID)
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
				// check file existance
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

	status, progress, filePath := geFileStatus(msgID)
	logs.Debug("SDK::FileDownload() current file progress status",
		zap.String("Status", domain.RequestStatusNames[status]),
		zap.Float64("Progress", progress),
		zap.String("FilePath", filePath),
	)
	m := repo.Ctx().Messages.GetMessage(msgID)
	if m == nil {
		logs.Error("SDK::FileDownload()", zap.Int64("Message does not exist", msgID))
		return
	}

	switch status {
	case domain.RequestStateNone:
		filemanager.Ctx().Download(m)
	case domain.RequestStateInProgress:
		// already downloading
		// filemanager.Ctx().Download(m)
	case domain.RequestStateCompleted:
		r.onFileDownloadCompleted(m.ID, filePath, domain.FileStateExistedDownload)
	case domain.RequestStatePaused:
		filemanager.Ctx().Download(m)
	case domain.RequestStateCanceled:
		filemanager.Ctx().Download(m)
	case domain.RequestStateError:
		filemanager.Ctx().Download(m)
	}

	fs, err := repo.Ctx().Files.GetFileStatus(msgID)
	if err == nil && fs != nil {

	} else {
		m := repo.Ctx().Messages.GetMessage(msgID)
		filemanager.Ctx().Download(m)
	}
}

// PauseDownload pause download
func (r *River) PauseDownload(msgID int64) {
	fs, err := repo.Ctx().Files.GetFileStatus(msgID)
	if err == nil {
		filemanager.Ctx().DeleteFromQueue(fs.FileID)
		repo.Ctx().Files.UpdateFileStatus(msgID, domain.RequestStatePaused)
	} else {
		logs.Error("SDK::PauseDownload()", zap.Int64("MsgID", msgID), zap.Error(err))
	}
}

// CancelDownload cancel download
func (r *River) CancelDownload(msgID int64) {
	fs, err := repo.Ctx().Files.GetFileStatus(msgID)
	if err == nil {
		filemanager.Ctx().DeleteFromQueue(fs.FileID)
		repo.Ctx().Files.UpdateFileStatus(msgID, domain.RequestStateCanceled)
	} else {
		logs.Error("SDK::CancelDownload()", zap.Int64("MsgID", msgID), zap.Error(err))
	}
}

// PauseUpload pause upload
func (r *River) PauseUpload(msgID int64) {
	fs, err := repo.Ctx().Files.GetFileStatus(msgID)
	if err == nil {
		filemanager.Ctx().DeleteFromQueue(fs.FileID)
		repo.Ctx().Files.UpdateFileStatus(msgID, domain.RequestStatePaused)
		// repo.Ctx().PendingMessages.DeletePendingMessage(fs.MessageID)
	} else {
		logs.Error("SDK::PauseUpload()", zap.Int64("MsgID", msgID), zap.Error(err))
	}
}

// CancelUpload cancel upload
func (r *River) CancelUpload(msgID int64) {
	fs, err := repo.Ctx().Files.GetFileStatus(msgID)
	if err == nil {
		filemanager.Ctx().DeleteFromQueue(fs.FileID)
		repo.Ctx().Files.DeleteFileStatus(msgID)
		repo.Ctx().PendingMessages.DeletePendingMessage(fs.MessageID)
	} else {
		logs.Error("SDK::CancelUpload()", zap.Int64("MsgID", msgID), zap.Error(err))
	}
}

// AccountUploadPhoto upload user profile photo
func (r *River) AccountUploadPhoto(filePath string) (msgID int64) {

	//TOF
	msgID = domain.SequentialUniqueID()
	fileID := domain.SequentialUniqueID()

	// support IOS file path
	if strings.HasPrefix(filePath, "file://") {
		filePath = filePath[7:]
	}

	file, err := os.Open(filePath)
	if err != nil {
		logs.Error("SDK::AccountUploadPhoto()", zap.Error(err))
		return 0
	}
	fileInfo, err := file.Stat()
	if err != nil {
		logs.Error("SDK::AccountUploadPhoto()", zap.Error(err))
		return 0
	}

	// // fileName := fileInfo.Name()
	totalSize := fileInfo.Size() // size in Byte
	// if totalSize > domain.FileMaxPhotoSize {
	// 	log.Error("SDK::AccountUploadPhoto()", zap.Error(errors.New("max allowed file size is 1 MB")))
	// 	return 0
	// }

	state := filemanager.NewFileStatus(msgID, fileID, 0, totalSize, filePath, domain.FileStateUploadAccountPhoto, 0, 0, 0, r.onFileProgressChanged)

	filemanager.Ctx().AddToQueue(state)

	return msgID
}

// AccountGetPhoto_Big download user profile picture
func (r *River) AccountGetPhoto_Big(userID int64) string {

	user := repo.Ctx().Users.GetUser(userID)
	if user != nil {
		if user.Photo != nil {
			dtoPhoto := repo.Ctx().Users.GetUserPhoto(userID, user.Photo.PhotoID)
			if dtoPhoto != nil {
				if dtoPhoto.Big_FilePath != "" {
					// check if file exist
					if _, err := os.Stat(dtoPhoto.Big_FilePath); os.IsNotExist(err) {
						return downloadAccountPhoto(userID, user.Photo, true)

					}
					// check if fileID is changed redownload
					strFileID := strconv.FormatInt(dtoPhoto.Big_FileID, 10)
					if strings.Index(dtoPhoto.Big_FilePath, strFileID) < 0 {
						return downloadAccountPhoto(user.ID, user.Photo, true)
					}

					return dtoPhoto.Big_FilePath

				}
				return downloadAccountPhoto(userID, user.Photo, true)

			}
			return downloadAccountPhoto(userID, user.Photo, true)

		}
		logs.Error("SDK::AccountGetPhoto_Big() user photo is null")
		return ""
	}
	logs.Error("SDK::AccountGetPhoto_Big() user does not exist")
	return ""
}

// AccountGetPhoto_Small download user profile picture thumbnail
func (r *River) AccountGetPhoto_Small(userID int64) string {

	user := repo.Ctx().Users.GetUser(userID)
	if user != nil {
		if user.Photo != nil {
			dtoPhoto := repo.Ctx().Users.GetUserPhoto(userID, user.Photo.PhotoID)
			if dtoPhoto != nil {

				if dtoPhoto.Small_FilePath != "" {
					// check if file exist
					if _, err := os.Stat(dtoPhoto.Small_FilePath); os.IsNotExist(err) {
						return downloadAccountPhoto(userID, user.Photo, false)
					}

					// check if fileID is changed redownload
					strFileID := strconv.FormatInt(dtoPhoto.Small_FileID, 10)
					if strings.Index(dtoPhoto.Small_FilePath, strFileID) < 0 {
						return downloadAccountPhoto(user.ID, user.Photo, true)
					}

					return dtoPhoto.Small_FilePath
				}
				return downloadAccountPhoto(userID, user.Photo, false)

			}
			return downloadAccountPhoto(userID, user.Photo, false)
		}
		logs.Error("SDK::AccountGetPhoto_Small() user photo is null")
		return ""
	}
	logs.Error("SDK::AccountGetPhoto_Small() user does not exist")
	return ""
}

// downloadAccountPhoto this function is sync
func downloadAccountPhoto(userID int64, photo *msg.UserPhoto, isBig bool) string {

	logs.Debug("SDK::downloadAccountPhoto",
		zap.Int64("UserID", userID),
		zap.Bool("IsBig", isBig),
		zap.Int64("PhotoBig.FileID", photo.PhotoBig.FileID),
		zap.Uint64("PhotoBig.AccessHash", photo.PhotoBig.AccessHash),
		zap.Int32("PhotoBig.ClusterID", photo.PhotoBig.ClusterID),
		zap.Int64("SmallBig.FileID", photo.PhotoSmall.FileID),
		zap.Uint64("SmallBig.AccessHash", photo.PhotoSmall.AccessHash),
		zap.Int32("SmallBig.ClusterID", photo.PhotoSmall.ClusterID),
	)

	// send Download request
	filePath, err := filemanager.Ctx().DownloadAccountPhoto(userID, photo, isBig)
	if err != nil {
		logs.Error("SDK::downloadAccountPhoto() error", zap.Error(err))
		return ""
	}
	return filePath
}

// geFileStatus
func geFileStatus(msgID int64) (status domain.RequestStatus, progress float64, filePath string) {

	fs, err := repo.Ctx().Files.GetFileStatus(msgID)
	if err == nil && fs != nil {
		// file is inprogress state
		// double check

		if fs.IsCompleted {
			go repo.Ctx().Files.DeleteFileStatus(fs.MessageID)
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
			status = domain.RequestStateCompleted
			progress = 100
		} else {
			// file does not exist and its progress state does not exist too
			status = domain.RequestStateNone
			progress = 0
			filePath = ""
		}
	}

	return
}

// GroupUploadPhoto upload group profile photo
func (r *River) GroupUploadPhoto(groupID int64, filePath string) (msgID int64) {

	//TOF
	msgID = domain.SequentialUniqueID()
	fileID := domain.SequentialUniqueID()

	// support IOS file path
	if strings.HasPrefix(filePath, "file://") {
		filePath = filePath[7:]
	}

	file, err := os.Open(filePath)
	if err != nil {
		logs.Error("SDK::GroupUploadPhoto()", zap.Error(err))
		return 0
	}
	fileInfo, err := file.Stat()
	if err != nil {
		logs.Error("SDK::GroupUploadPhoto()", zap.Error(err))
		return 0
	}

	// // fileName := fileInfo.Name()
	totalSize := fileInfo.Size() // size in Byte
	// if totalSize > domain.FileMaxPhotoSize {
	// 	log.Error("SDK::GroupUploadPhoto()", zap.Error(errors.New("max allowed file size is 1 MB")))
	// 	return 0
	// }

	state := filemanager.NewFileStatus(msgID, fileID, groupID, totalSize, filePath, domain.FileStateUploadGroupPhoto, 0, 0, 0, r.onFileProgressChanged)

	filemanager.Ctx().AddToQueue(state)

	return msgID
}

// GroupGetPhoto_Big download group profile picture
func (r *River) GroupGetPhoto_Big(groupID int64) string {

	group, err := repo.Ctx().Groups.GetGroupDTO(groupID)
	if err == nil && group != nil {
		if group.Photo != nil {
			groupPhoto := new(msg.GroupPhoto)
			err = groupPhoto.Unmarshal(group.Photo)
			if err != nil {
				logs.Error("SDK::GroupGetPhoto_Big() failed to unmarshal GroupPhoto", zap.Error(err))
				return ""
			}
			if group.Big_FilePath != "" {
				// check if file exist
				if _, err := os.Stat(group.Big_FilePath); os.IsNotExist(err) {
					return downloadGroupPhoto(groupID, groupPhoto, true)
				}
				// check if fileID is changed redownload
				strFileID := strconv.FormatInt(groupPhoto.PhotoBig.FileID, 10)
				if strings.Index(group.Big_FilePath, strFileID) < 0 {
					return downloadGroupPhoto(groupID, groupPhoto, true)
				}
				return group.Big_FilePath

			}
			return downloadGroupPhoto(groupID, groupPhoto, true)

		}
		logs.Error("SDK::GroupGetPhoto_Big() group photo is null")
		return ""
	}
	logs.Error("SDK::GroupGetPhoto_Big() group does not exist")
	return ""
}

// GroupGetPhoto_Small download group profile picture thumbnail
func (r *River) GroupGetPhoto_Small(groupID int64) string {

	group, err := repo.Ctx().Groups.GetGroupDTO(groupID)
	if err == nil && group != nil {
		if group.Photo != nil {
			groupPhoto := new(msg.GroupPhoto)
			err = groupPhoto.Unmarshal(group.Photo)
			if err != nil {
				logs.Error("SDK::GroupGetPhoto_Small() failed to unmarshal GroupPhoto", zap.Error(err))
				return ""
			}
			if group.Small_FilePath != "" {

				// check if file exist
				if _, err := os.Stat(group.Small_FilePath); os.IsNotExist(err) {
					return downloadGroupPhoto(groupID, groupPhoto, false)
				}

				// check if fileID is changed redownload
				strFileID := strconv.FormatInt(groupPhoto.PhotoSmall.FileID, 10)
				if strings.Index(group.Small_FilePath, strFileID) < 0 {
					return downloadGroupPhoto(groupID, groupPhoto, false)
				}

				return group.Small_FilePath

			}
			return downloadGroupPhoto(groupID, groupPhoto, false)

		}
		logs.Error("SDK::GroupGetPhoto_Small() group photo is null")
		return ""
	}
	logs.Error("SDK::GroupGetPhoto_Small() group does not exist")
	return ""
}

// this function is sync
func downloadGroupPhoto(groupID int64, photo *msg.GroupPhoto, isBig bool) string {

	logs.Debug("SDK::downloadGroupPhoto",
		zap.Int64("UserID", groupID),
		zap.Bool("IsBig", isBig),
		zap.Int64("PhotoBig.FileID", photo.PhotoBig.FileID),
		zap.Uint64("PhotoBig.AccessHash", photo.PhotoBig.AccessHash),
		zap.Int32("PhotoBig.ClusterID", photo.PhotoBig.ClusterID),
		zap.Int64("SmallBig.FileID", photo.PhotoSmall.FileID),
		zap.Uint64("SmallBig.AccessHash", photo.PhotoSmall.AccessHash),
		zap.Int32("SmallBig.ClusterID", photo.PhotoSmall.ClusterID),
	)

	// send Download request
	filePath, err := filemanager.Ctx().DownloadGroupPhoto(groupID, photo, isBig)
	if err != nil {
		logs.Error("SDK::downloadGroupPhoto() error", zap.Error(err))
		return ""
	}
	return filePath
}

// FileDownloadThumbnail download file thumbnail
func (r *River) FileDownloadThumbnail(msgID int64) string {

	m := repo.Ctx().Messages.GetMessage(msgID)
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

	dto, err := repo.Ctx().Files.GetFile(msgID)

	if err == nil {
		if _, err = os.Stat(dto.ThumbFilePath); os.IsNotExist(err) {
			path, err := filemanager.Ctx().DownloadThumbnail(m.ID, docID, accessHash, clusterID, version)
			if err != nil {
				logs.Error("SDK::FileDownloadThumbnail()-> filemanager.DownloadThumbnail()", zap.Error(err))
			}
			filePath = path
		} else {
			filePath = dto.ThumbFilePath
		}
	} else {
		path, _ := filemanager.Ctx().DownloadThumbnail(m.ID, docID, accessHash, clusterID, version)
		if err != nil {
			logs.Error("SDK::FileDownloadThumbnail()-> filemanager.DownloadThumbnail()", zap.Error(err))
		}
		filePath = path
	}

	return filePath
}

// GetSharedMedia search in given dialog files
func (r *River) GetSharedMedia(peerID int64, peerType int32, mediaType int32, delegate RequestDelegate) {
	msgs, err := repo.Ctx().Files.GetSharedMedia(peerID, peerType, mediaType)
	if err != nil {
		out := new(msg.MessageEnvelope)
		res := new(msg.Error)
		res.Code = "00"
		res.Items = err.Error()
		msg.ResultError(out, res)
		outBytes, _ := out.Marshal()
		if delegate.OnComplete != nil {
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
		} else {
			groupIDs[m.PeerID] = true
		}

		if m.FwdSenderID > 0 {
			userIDs[m.FwdSenderID] = true
		} else {
			groupIDs[m.FwdSenderID] = true
		}
	}

	users := repo.Ctx().Users.GetAnyUsers(userIDs.ToArray())
	groups := repo.Ctx().Groups.GetManyGroups(groupIDs.ToArray())

	msgMany := new(msg.MessagesMany)
	msgMany.Messages = msgs
	msgMany.Users = users
	msgMany.Groups = groups

	out := new(msg.MessageEnvelope)
	out.Constructor = msg.C_MessagesMany
	out.Message, _ = msgMany.Marshal()
	outBytes, _ := out.Marshal()
	if delegate.OnComplete != nil {
		delegate.OnComplete(outBytes)
	}
}
