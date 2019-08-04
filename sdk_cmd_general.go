package riversdk

import (
	"git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"git.ronaksoftware.com/ronak/riversdk/pkg/logs"
	mon "git.ronaksoftware.com/ronak/riversdk/pkg/monitoring"
	"git.ronaksoftware.com/ronak/riversdk/pkg/repo"
	"git.ronaksoftware.com/ronak/riversdk/pkg/salt"
	"git.ronaksoftware.com/ronak/riversdk/pkg/uiexec"
	"go.uber.org/zap"
	"os"
	"time"
)

// var DatabaseStatus map[int64]map[msg.DocumentAttributeType]dto.MediaInfo

// CancelRequest remove given requestID callbacks&delegates and if its not processed by queue we skip it on queue distributor
func (r *River) CancelRequest(requestID int64) {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("CancelRequest", time.Now().Sub(startTime))
	}()
	// Remove delegate
	r.delegateMutex.Lock()
	delete(r.delegates, int64(requestID))
	r.delegateMutex.Unlock()

	// Remove Callback
	domain.RemoveRequestCallback(uint64(requestID))

	// Cancel Request
	r.queueCtrl.CancelRequest(requestID)

}

// Delete removes pending message from DB
func (r *River) DeletePendingMessage(id int64) (isSuccess bool) {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("Delete", time.Now().Sub(startTime))
	}()
	repo.PendingMessages.Delete(id)
	isSuccess = true
	return
}

// RetryPendingMessage puts pending message again in command queue to re send it
func (r *River) RetryPendingMessage(id int64) (isSuccess bool) {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("RetryPendingMessage", time.Now().Sub(startTime))
	}()
	pmsg := repo.PendingMessages.GetByID(id)
	if pmsg == nil {
		isSuccess = false
		return
	}
	req := new(msg.MessagesSend)
	req.Body = pmsg.Body
	req.Peer = new(msg.InputPeer)
	req.Peer.AccessHash = pmsg.AccessHash
	req.Peer.ID = pmsg.PeerID
	req.Peer.Type = msg.PeerType(pmsg.PeerType)
	req.RandomID = pmsg.RequestID
	req.ReplyTo = pmsg.ReplyTo
	req.ClearDraft = pmsg.ClearDraft
	req.Entities = pmsg.Entities
	buff, _ := req.Marshal()
	r.queueCtrl.ExecuteCommand(uint64(req.RandomID), msg.C_MessagesSend, buff, nil, nil, true)
	isSuccess = true
	logs.Debug("River::RetryPendingMessage() Request enqueued")

	return
}

// GetSyncStatus returns SyncController status
func (r *River) GetSyncStatus() int32 {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("GetSyncStatus", time.Now().Sub(startTime))
	}()
	return int32(r.syncCtrl.GetSyncStatus())
}

// Logout drop queue & database , etc ...
func (r *River) Logout(notifyServer bool, reason int) error {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("Logout", time.Now().Sub(startTime))
	}()
	// unregister device if token exist
	if notifyServer {
		if r.DeviceToken != nil {
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
		// send logout request to server
		requestID := domain.RandomInt63()
		timeoutCallback := func() {
			r.releaseDelegate(requestID)
			r.networkCtrl.Disconnect()
			r.clearSystemConfig()
		}
		successCallback := func(envelope *msg.MessageEnvelope) {
			r.releaseDelegate(requestID)
			r.networkCtrl.Disconnect()
			r.clearSystemConfig()
		}
		req := new(msg.AuthLogout)
		buff, _ := req.Marshal()
		err := r.queueCtrl.ExecuteRealtimeCommand(uint64(requestID), msg.C_AuthLogout, buff, timeoutCallback, successCallback, true, false)
		if err != nil {
			r.releaseDelegate(requestID)
		}
	}
	if r.mainDelegate != nil {
		r.mainDelegate.OnSessionClosed(reason)
	}


	// Stop all the controllers and repo
	r.logout()

	for os.RemoveAll(r.dbPath) != nil {
		time.Sleep(time.Second)
	}
	r.queueCtrl.DropQueue()

	err := r.Start()
	if err != nil {
		return err
	}

	r.clearSystemConfig()

	return err
}
func (r *River) logout() {
	logs.Debug("StopServices-River::Stop() -> Called")

	// Disconnect from Server
	r.networkCtrl.Disconnect()

	// Stop Controllers
	r.syncCtrl.Stop()
	r.queueCtrl.Stop()
	r.networkCtrl.Stop()
	r.fileCtrl.Stop()
	uiexec.Ctx().Stop()

	// Close database connection
	err := repo.Close()
	if err != nil {
		logs.Debug("River::Stop() failed to close DB context",
			zap.String("Error", err.Error()),
		)
	}
}


// UpdateContactInfo update contact name
func (r *River) UpdateContactInfo(userID int64, firstName, lastName string) error {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("UpdateContactInfo", time.Now().Sub(startTime))
	}()
	repo.Users.UpdateContactInfo(userID, firstName, lastName)
	return nil
}

// GetGroupInputUser get group participant user
func (r *River) GetGroupInputUser(requestID int64, groupID int64, userID int64, delegate RequestDelegate) {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("GetGroupInputUser", time.Now().Sub(startTime))
	}()
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
		// SendWebsocket GroupsGetFull request to get user AccessHash
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

func (r *River) GetScrollStatus(peerID int64, peerType int32) int64 {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("GetScrollStatus", time.Now().Sub(startTime))
	}()
	return repo.MessagesExtra.GetScrollID(peerID, peerType)
}

func (r *River) SetScrollStatus(peerID, msgID int64, peerType int32) {
	startTime := time.Now()
	defer func() {
		mon.FunctionResponseTime("SetScrollStatus", time.Now().Sub(startTime))
	}()
	repo.MessagesExtra.SaveScrollID(peerID, peerType, msgID)

}

func (r *River) GetSDKSalt() int64 {
	return salt.Get()
}
