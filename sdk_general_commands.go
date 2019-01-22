package riversdk

import (
	"errors"
	"os"

	"git.ronaksoftware.com/ronak/riversdk/domain"
	"git.ronaksoftware.com/ronak/riversdk/filemanager"
	"git.ronaksoftware.com/ronak/riversdk/log"
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
		log.LOG_Debug("River::RetryPendingMessage()",
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
	log.LOG_Debug("River::RetryPendingMessage() Request enqueued")

	return
}

// GetNetworkStatus returns NetworkController status
func (r *River) GetNetworkStatus() int32 {
	return int32(r.networkCtrl.Quality())
}

// GetSyncStatus returns SyncController status
func (r *River) GetSyncStatus() int32 {

	log.LOG_Debug("River::GetSyncStatus()",
		zap.String("syncStatus", domain.SyncStatusName[r.syncCtrl.Status()]),
	)
	return int32(r.syncCtrl.Status())
}

// Logout drop queue & database , etc ...
func (r *River) Logout() (int64, error) {

	dataDir, err := r.queueCtrl.DropQueue()

	if err != nil {
		log.LOG_Debug("River::Logout() failed to drop queue",
			zap.Error(err),
		)
	}

	// drop and recreate database
	err = repo.Ctx().ReinitiateDatabase()
	if err != nil {
		log.LOG_Debug("River::Logout() failed to re initiate database",
			zap.Error(err),
		)
	}

	// open queue
	err = r.queueCtrl.OpenQueue(dataDir)
	if err != nil {
		log.LOG_Debug("River::Logout() failed to re open queue",
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
		log.LOG_Info("River::UISettingsGet()",
			zap.String("Error", err.Error()),
		)
	}
	return val
}

// UISettingPut save to key/value storage for UI settings
func (r *River) UISettingPut(key, value string) bool {
	err := repo.Ctx().UISettings.Put(key, value)
	if err != nil {
		log.LOG_Info("River::UISettingsPut()",
			zap.String("Error", err.Error()),
		)
	}
	return err == nil
}

// UISettingDelete remove from key/value storage for UI settings
func (r *River) UISettingDelete(key string) bool {
	err := repo.Ctx().UISettings.Delete(key)
	if err != nil {
		log.LOG_Info("River::UISettingsDelete()",
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
		log.LOG_Debug("SDK::GetRealTopMessageID() => Messages.GetTopMessageID()", zap.String("Error", err.Error()))
		return -1
	}
	return topMsgID
}

// UpdateContactinfo update contact name
func (r *River) UpdateContactinfo(userID int64, firstName, lastName string) error {
	err := repo.Ctx().Users.UpdateContactinfo(userID, firstName, lastName)
	if err != nil {
		log.LOG_Debug("SDK::UpdateContactinfo() => Users.UpdateContactInfo()", zap.String("Error", err.Error()))
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

func (r *River) GetFileProgess(msgID int64) float64 {
	fs, err := repo.Ctx().Files.GetFileStatus(msgID)
	if err != nil {
		return -1
	}
	if fs.TotalParts <= 0 {
		return -1
	}
	if fs.IsCompleted || fs.Position == fs.TotalSize {
		go repo.Ctx().Files.DeleteFileStatus(fs.MessageID)
		return -1
	}
	return (float64(fs.Position) / float64(fs.TotalSize) * float64(100))
}
func (r *River) GetFilePath(msgID int64) string {
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
		}
	}
	return ""
}

func (r *River) FileDownload(msgID int64) {
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
					filemanager.Ctx().Download(m)
				} else {
					r.onFileDownloadCompleted(m.ID, filePath, domain.FileStateExistedDownload)
				}

			} else {
				log.LOG_Error("SDK::FileDownload()", zap.Error(err))
			}

		default:
			log.LOG_Error("SDK::FileDownload() MediaType is invalid", zap.Int32("MediaType", int32(m.MediaType)))
		}
	} else {
		log.LOG_Error("SDK::FileDownload()", zap.Int64("Message does not exist", msgID))
	}
}

func (r *River) CancelDownload(msgID int64) {
	fs, err := repo.Ctx().Files.GetFileStatus(msgID)
	if err == nil {
		filemanager.Ctx().DeleteFromQueue(fs.FileID)
		repo.Ctx().Files.UpdateFileStatus(msgID, domain.RequestStatePused)
	}
}

func (r *River) CancelUpload(msgID int64) {
	fs, err := repo.Ctx().Files.GetFileStatus(msgID)
	if err == nil {
		filemanager.Ctx().DeleteFromQueue(fs.FileID)
		repo.Ctx().PendingMessages.DeletePendingMessage(fs.MessageID)
	} else {
		log.LOG_Error("SDK::CancelUpload()", zap.Int64("MsgID", msgID), zap.Error(err))
	}
}

func (r *River) AccountUploadPhoto(filePath string) (msgID int64) {

	//TOF
	msgID = domain.SequentialUniqueID()
	fileID := domain.SequentialUniqueID()

	file, err := os.Open(filePath)
	if err != nil {
		log.LOG_Error("SDK::AccountUploadPhoto()", zap.Error(err))
		return 0
	}
	fileInfo, err := file.Stat()
	if err != nil {
		log.LOG_Error("SDK::AccountUploadPhoto()", zap.Error(err))
		return 0
	}

	// fileName := fileInfo.Name()
	totalSize := fileInfo.Size() // size in Byte
	if totalSize > domain.FileMaxPhotoSize {
		log.LOG_Error("SDK::AccountUploadPhoto()", zap.Error(errors.New("max allowed file size is 1 MB")))
		return 0
	}

	state := filemanager.NewFileStatus(msgID, fileID, totalSize, filePath, domain.FileStateUploadAccountPhoto, 0, 0, 0, r.onFileProgressChanged)

	filemanager.Ctx().AddToQueue(state)

	return msgID
}
