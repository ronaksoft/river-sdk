package riversdk

import (
	"errors"
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo/dto"
	"go.uber.org/zap"
)

var GetDBStatusIsRunning bool
var DatabaseStatus map[int64]map[msg.DocumentAttributeType]dto.MediaInfo

// CancelRequest remove given requestID callbacks&delegates and if its not processed by queue we skip it on queue distributor
func (r *River) CancelRequest(requestID int64) {
	// Remove delegate
	r.delegateMutex.Lock()
	delete(r.delegates, int64(requestID))
	r.delegateMutex.Unlock()

	// Remove Callback
	domain.RemoveRequestCallback(uint64(requestID))

	// Cancel Request
	r.queueCtrl.CancelRequest(requestID)

}

// DeletePendingMessage removes pending message from DB
func (r *River) DeletePendingMessage(id int64) (isSuccess bool) {
	err := repo.PendingMessages.DeletePendingMessage(id)
	isSuccess = err == nil
	return
}

// RetryPendingMessage puts pending message again in command queue to re send it
func (r *River) RetryPendingMessage(id int64) (isSuccess bool) {
	pmsg, err := repo.PendingMessages.GetPendingMessageByID(id)
	if err != nil {
		logs.Error("River::RetryPendingMessage() -> GetPendingMessageByID()", zap.Error(err))
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
	return int32(r.networkCtrl.GetQuality())
}

// GetSyncStatus returns SyncController status
func (r *River) GetSyncStatus() int32 {
	return int32(r.syncCtrl.GetSyncStatus())
}

// Logout drop queue & database , etc ...
func (r *River) Logout(notifyServer bool, reason int) (int64, error) {
	// unregister device if token exist
	if notifyServer && r.DeviceToken != nil {
		reqID := uint64(domain.SequentialUniqueID())
		req := new(msg.AccountUnregisterDevice)
		req.Token = r.DeviceToken.Token
		req.TokenType = int32(r.DeviceToken.TokenType)
		reqBytes, _ := req.Marshal()
		_ = r.queueCtrl.ExecuteRealtimeCommand(
			reqID,
			msg.C_AccountUnregisterDevice,
			reqBytes,
			nil, nil, true, false,
		)
	}

	dataDir, err := r.queueCtrl.DropQueue()
	if err != nil {
		logs.Error("River::Logout() failed to drop queue", zap.Error(err))
	}

	// drop and recreate database
	err = repo.ReInitiateDatabase()
	if err != nil {
		logs.Error("River::Logout() failed to re initiate database", zap.Error(err))
	}

	// open queue
	err = r.queueCtrl.OpenQueue(dataDir)
	if err != nil {
		logs.Error("River::Logout() failed to re open queue", zap.Error(err))
	}

	// send logout request to server
	requestID := domain.RandomInt63()
	timeoutCallback := func() {
		err = domain.ErrRequestTimeout
		r.releaseDelegate(requestID)

		r.clearSystemConfig()
		r.syncCtrl.ClearUpdateID()
		r.networkCtrl.Reconnect()
		r.syncCtrl.ClearUpdateID()
	}
	successCallback := func(envelope *msg.MessageEnvelope) {
		r.releaseDelegate(requestID)

		r.clearSystemConfig()
		r.syncCtrl.ClearUpdateID() // TOF
		r.networkCtrl.Reconnect()
		r.syncCtrl.ClearUpdateID()
	}

	if notifyServer {
		req := new(msg.AuthLogout)
		buff, _ := req.Marshal()
		err = r.queueCtrl.ExecuteRealtimeCommand(uint64(requestID), msg.C_AuthLogout, buff, timeoutCallback, successCallback, true, false)
		if err != nil {
			r.releaseDelegate(requestID)
		}
	} else {
		r.clearSystemConfig()
		r.syncCtrl.ClearUpdateID() // TOF
		r.networkCtrl.Reconnect()
		r.syncCtrl.ClearUpdateID()
	}

	if r.mainDelegate != nil {
		r.mainDelegate.OnSessionClosed(reason)
	}

	return requestID, err
}

// UISettingGet fetch from key/value storage for UI settings
func (r *River) UISettingGet(key string) string {
	val, err := repo.UISettings.Get(key)
	if err != nil {
		logs.Warn("River::UISettingsGet()", zap.Error(err))
	}
	return val
}

// UISettingPut save to key/value storage for UI settings
func (r *River) UISettingPut(key, value string) bool {
	err := repo.UISettings.Put(key, value)
	if err != nil {
		logs.Error("River::UISettingsPut()", zap.Error(err))
	}
	return err == nil
}

// UISettingDelete remove from key/value storage for UI settings
func (r *River) UISettingDelete(key string) bool {
	err := repo.UISettings.Delete(key)
	if err != nil {
		logs.Error("River::UISettingsDelete()", zap.Error(err))
	}
	return err == nil
}

// SearchContacts searches contacts
func (r *River) SearchContacts(requestID int64, searchPhrase string, delegate RequestDelegate) {
	res := new(msg.MessageEnvelope)
	res.Constructor = msg.C_ContactsMany
	res.RequestID = uint64(requestID)

	contacts := new(msg.ContactsMany)
	contacts.Users, contacts.Contacts = repo.Users.SearchContacts(searchPhrase)

	res.Message, _ = contacts.Marshal()

	buff, _ := res.Marshal()
	if delegate != nil {
		delegate.OnComplete(buff)
	}
}

// UpdateContactInfo update contact name
func (r *River) UpdateContactInfo(userID int64, firstName, lastName string) error {
	err := repo.Users.UpdateContactInfo(userID, firstName, lastName)
	if err != nil {
		logs.Error("SDK::UpdateContactInfo() => Users.UpdateContactInfo()", zap.Error(err))
	}
	return err
}

// SearchDialogs search dialog title
func (r *River) SearchDialogs(requestID int64, searchPhrase string, delegate RequestDelegate) {
	res := new(msg.MessageEnvelope)
	res.Constructor = msg.C_MessagesDialogs
	res.RequestID = uint64(requestID)

	dlgs := new(msg.MessagesDialogs)

	users := repo.Users.SearchUsers(searchPhrase)
	groups := repo.Groups.Search(searchPhrase)
	dlgs.Users = users
	dlgs.Groups = groups

	mDialogs := domain.MInt64B{}
	for _, v := range users {
		mDialogs[v.ID] = true
	}
	for _, v := range groups {
		mDialogs[v.ID] = true
	}

	dialogs := repo.Dialogs.GetMany(mDialogs.ToArray())
	dlgs.Dialogs = dialogs

	mMessages := domain.MInt64B{}
	for _, v := range dialogs {
		mMessages[v.TopMessageID] = true
	}
	dlgs.Messages = repo.Messages.GetManyMessages(mMessages.ToArray())

	res.Message, _ = dlgs.Marshal()
	buff, _ := res.Marshal()
	if delegate != nil {
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

	accessHash, err := repo.Users.GetAccessHash(userID)
	if err != nil || accessHash == 0 {
		participant, err := repo.Groups.GetParticipants(groupID)
		if err == nil {
			for _, p := range participant {
				if p.UserID == userID {
					accessHash = p.AccessHash
					break
				}
			}
		} else {
			logs.Error("GetGroupInputUser() -> GetParticipants()", zap.Error(err))
		}
	} else {
		logs.Error("GetGroupInputUser() -> GetAccessHash()", zap.Error(err))
	}

	if accessHash == 0 {
		// get group full and get its access hash from its participants
		req := new(msg.GroupsGetFull)
		req.GroupID = groupID
		reqBytes, _ := req.Marshal()

		out := new(msg.MessageEnvelope)
		// Timeout Callback
		timeoutCB := func() {
			if delegate != nil {
				delegate.OnTimeout(domain.ErrRequestTimeout)
			}
		}

		// Success Callback
		successCB := func(response *msg.MessageEnvelope) {
			if response.Constructor != msg.C_GroupFull {
				msg.ResultError(out, &msg.Error{Code: "00", Items: "response type is not GroupFull"})
				if delegate != nil {
					outBytes, _ := out.Marshal()
					delegate.OnComplete(outBytes)
				}
				return
			}

			groupFull := new(msg.GroupFull)
			err := groupFull.Unmarshal(response.Message)
			if err != nil {
				msg.ResultError(out, &msg.Error{Code: "00", Items: err.Error()})
				if delegate != nil {
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
			if delegate != nil {
				delegate.OnComplete(resBytes)
			}
		}
		// Send GroupsGetFull request to get user AccessHash
		r.queueCtrl.ExecuteRealtimeCommand(uint64(requestID), msg.C_GroupsGetFull, reqBytes, timeoutCB, successCB, true, false)

	} else {
		user.AccessHash = accessHash
		res.Message, _ = user.Marshal()

		buff, _ := res.Marshal()
		if delegate != nil {
			delegate.OnComplete(buff)
		}
	}
}

// GetSharedMedia search in given dialog files
func (r *River) GetSharedMedia(peerID int64, peerType int32, mediaType int32, delegate RequestDelegate) {
	msgs, err := repo.Files.GetSharedMedia(peerID, peerType, mediaType)
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
		} else {
			groupIDs[m.PeerID] = true
		}

		if m.FwdSenderID > 0 {
			userIDs[m.FwdSenderID] = true
		} else {
			groupIDs[m.FwdSenderID] = true
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

func (r *River) GetScrollStatus(peerID int64, peerType int32) int64 {
	status, err := repo.MessagesExtra.GetScrollID(peerID, peerType)
	if err != nil {
		return 0
	} else {
		return status
	}
}

func (r *River) SetScrollStatus(peerID, msgID int64, peerType int32) {
	if err := repo.MessagesExtra.SaveScrollID(peerID, msgID, peerType); err != nil {
		logs.Error("SetScrollStatus::Failed to set scroll ID")
	}
}

// SearchGlobal returns messages, contacts and groups matching given text
// peerID 0 means search is not limited to a specific peerID
func (r *River) SearchGlobal(text string, peerID int64, delegate RequestDelegate) {
	searchResults := new(msg.ClientSearchResult)
	var userContacts []*msg.ContactUser
	var NonContactUsersWithDialogs []*msg.ContactUser
	var msgs []*msg.UserMessage
	var peerIDs []int64
	if peerID != 0 {
		msgs = repo.Messages.SearchTextByPeerID(text, peerID)
	} else {
		msgs = repo.Messages.SearchText(text)
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

	users := repo.Users.GetMany(userIDs.ToArray())
	groups := repo.Groups.GetMany(groupIDs.ToArray())

	// if peerID == 0 then look for group and contact names too
	if peerID == 0 {
		userContacts, _ = repo.Users.SearchContacts(text)
		peerIDs = repo.Dialogs.GetPeerIDs()

		// Get users who have dialog with me, but are not my contact
		NonContactUsersWithDialogs = repo.Users.SearchNonContactsWithIDs(peerIDs, text)

		userContacts = append(userContacts, NonContactUsersWithDialogs...)
	}

	searchResults.Messages = msgs
	searchResults.Users = users
	searchResults.Groups = groups
	searchResults.MatchedUsers = userContacts
	searchResults.MatchedGroups = repo.Groups.SearchByTitle(text)

	outBytes, _ := searchResults.Marshal()

	if delegate != nil {
		delegate.OnComplete(outBytes)
	}
}

// GetGetDBStatus returns message IDs and total size of each media stored in user's database
func (r *River) GetDBStatus(delegate RequestDelegate) {
	res := msg.DBMediaInfo{}
	if GetDBStatusIsRunning {
		err := errors.New("GetDBStatus is running")
		if delegate != nil {
			delegate.OnTimeout(err)
		}
		return
	}
	GetDBStatusIsRunning = true
	for k := range DatabaseStatus {
		delete(DatabaseStatus, k)
	}
	logs.Debug("DatabaseStatus Must be Empty", zap.Any("", fmt.Sprintf("%+v", DatabaseStatus)))
	peerMediaSizeMap, err := repo.Files.GetDBStatus()
	if err != nil {
		GetDBStatusIsRunning = false
		logs.Error(err.Error())
		delegate.OnTimeout(err)
		return
	}
	logs.Debug("peerMediaSizeMap", zap.Any("peerMediaSizeMap", peerMediaSizeMap))
	peerMediaInfo := make([]*msg.PeerMediaInfo, 0)
	for peerID, mediaInfoMap := range peerMediaSizeMap {
		mediaSize := make([]*msg.MediaSize, 0)
		for mediaType, mediaInfo := range mediaInfoMap {
			mediaSize = append(mediaSize, &msg.MediaSize{MediaType: int32(mediaType), TotalSize: mediaInfo.Size})
		}
		peerMediaInfo = append(peerMediaInfo, &msg.PeerMediaInfo{PeerID: peerID, Media: mediaSize})
	}
	res.MediaInfo = peerMediaInfo
	logs.Debug("MediaInfo", zap.String("", fmt.Sprintf("%+v", res.MediaInfo)))
	pBytes, _ := res.Marshal()
	if delegate != nil {
		delegate.OnComplete(pBytes)
	}
	GetDBStatusIsRunning = false
	DatabaseStatus = peerMediaSizeMap
}

func (r *River) GetSDKSalt() int64 {
	return r.networkCtrl.GetServerSalt()
}