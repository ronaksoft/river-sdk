package riversdk

import (
	"git.ronaksoftware.com/ronak/riversdk/domain"
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
func (r *River) SearchContacts(requestID uint64, searchPhrase string, delegate RequestDelegate) {
	res := new(msg.MessageEnvelope)
	res.Constructor = msg.C_ContactsMany
	res.RequestID = requestID

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
