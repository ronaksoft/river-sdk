package call

import (
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"go.uber.org/zap"
	"sync"
)

func (c *call) toggleVide(enable bool) (err error) {
	c.propagateMediaSettings(MediaSettingsIn{
		Video: &enable,
	})

	return c.modifyMediaStream(enable)
}

func (c *call) toggleAudio(enable bool) (err error) {
	c.propagateMediaSettings(MediaSettingsIn{
		Audio: &enable,
	})
	return
}

func (c *call) tryReconnect(connId int32) (err error) {
	_ = c.checkDisconnection(connId, "disconnected", false)
	return
}

func (c *call) start(peer *msg.InputPeer, participants []*msg.InputUser, callID int64) (id int64, err error) {
	c.peer = peer
	initRes, err := c.apiInit(peer, callID)
	if err != nil {
		logs.Warn("Init", zap.Error(err))
		return
	}

	c.iceServer = initRes.IceServers
	if callID != 0 {
		c.activeCallID = callID
		var joinRes *msg.PhoneParticipants
		joinRes, err = c.apiJoin(peer, c.activeCallID)
		if err != nil {
			return
		}

		c.initParticipants(c.activeCallID, joinRes.Participants, true)
		_, err = c.initConnections(peer, c.activeCallID, false, nil)
		if err != nil {
			return
		}
	} else {
		c.activeCallID = 0
		c.initCallParticipants(TempCallID, participants)
		_, err = c.initConnections(peer, TempCallID, true, nil)
		if err != nil {
			return
		}

		c.swapTempInfo(c.activeCallID)
	}
	id = c.activeCallID
	return
}

func (c *call) join(peer *msg.InputPeer, callID int64) {
	if c.activeCallID == 0 {
		return
	}

	c.activeCallID = callID
	update := msg.CallUpdateCallPreview{
		CallID: callID,
		Peer:   peer,
	}
	updateData, uErr := update.Marshal()
	if uErr == nil {
		c.callUpdate(msg.CallUpdate_CallPreview, updateData)
	}
}

func (c *call) accept(callID int64, video bool) (err error) {
	if c.peer == nil {
		err = ErrInvalidPeerInput
		return
	}

	initRes, err := c.apiInit(c.peer, callID)
	if err != nil {
		return
	}

	c.mu.Lock()
	c.activeCallID = callID
	c.iceServer = initRes.IceServers
	c.mu.Unlock()

	info := c.getCallInfo(callID)
	if info == nil {
		err = ErrInvalidCallRequest
		return
	}

	if len(info.requests) == 0 {
		err = ErrNoCallRequest
		return
	}

	initFn := func() error {
		wg := sync.WaitGroup{}
		for _, request := range info.requests {
			wg.Add(0)
			go func(req *UpdatePhoneCall) {
				defer wg.Done()
				_, innerErr := c.initConnections(c.peer, callID, false, req)
				if innerErr != nil {
					return
				}

				streamState := c.getStreamState()
				c.mediaSettingsInit(streamState)
				c.propagateMediaSettings(MediaSettingsIn{
					Audio:       &streamState.Audio,
					ScreenShare: &streamState.ScreenShare,
					Video:       &streamState.Video,
				})
			}(request)
		}
		wg.Wait()

		return nil
	}

	if !info.dialed {
		err = c.callback.InitStream(true, video)
		if err != nil {
			return
		}

		return initFn()
	}

	return initFn()
}

func (c *call) reject(callID int64, duration int32, reason msg.DiscardReason, targetPeer *msg.InputPeer) (err error) {
	peer := c.peer
	if targetPeer != nil {
		peer = targetPeer
	}

	if peer == nil {
		err = ErrInvalidPeerInput
		return
	}

	_, err = c.apiReject(peer, callID, reason, duration)
	return
}

func (c *call) groupAddParticipant(callID int64, participants []*msg.InputUser) (err error) {
	if c.peer == nil {
		err = ErrInvalidPeerInput
		return
	}

	_, err = c.apiAddParticipant(c.peer, callID, participants)
	return
}

func (c *call) groupRemoveParticipant(callID int64, userIDs []int64, timeout bool) (err error) {
	if c.peer == nil {
		err = ErrInvalidPeerInput
		return
	}

	inputUsers := c.getInputUserByUserIDs(callID, userIDs)
	if len(inputUsers) == 0 {
		err = ErrInvalidCallId
		return
	}

	_, err = c.apiRemoveParticipant(c.peer, callID, inputUsers, timeout)
	if err != nil {
		return
	}

	for _, userID := range userIDs {
		c.removeParticipant(userID, nil)
	}

	update := msg.CallUpdateParticipantRemoved{
		UserIDs: userIDs,
		Timeout: timeout,
	}
	updateData, uErr := update.Marshal()
	if uErr == nil {
		c.callUpdate(msg.CallUpdate_ParticipantRemoved, updateData)
	}
	return
}

func (c *call) groupUpdateAdmin(callID int64, userID int64, admin bool) (err error) {
	if c.peer == nil {
		err = ErrInvalidPeerInput
		return
	}

	inputUsers := c.getInputUserByUserIDs(callID, []int64{userID})
	if len(inputUsers) == 0 {
		err = ErrInvalidCallId
		return
	}

	_, err = c.apiUpdateAdmin(c.peer, callID, inputUsers[0], admin)
	if err != nil {
		return
	}

	c.updateAdmin(userID, admin)

	update := msg.CallUpdateParticipantAdminUpdated{
		UserID: userID,
		Admin:  admin,
	}
	updateData, uErr := update.Marshal()
	if uErr == nil {
		c.callUpdate(msg.CallUpdate_ParticipantAdminUpdated, updateData)
	}

	return
}

func (c *call) groupGetParticipantByUserID(callID int64, userID int64) (participant *Participant) {
	connId, info, valid := c.getConnId(callID, userID)
	if !valid {
		return
	}

	info.mu.RLock()
	participant = info.participants[connId]
	info.mu.RUnlock()
	return
}

func (c *call) groupGetParticipantByConnId(connId int32) (participant *Participant) {
	if c.activeCallID == 0 {
		return
	}

	info := c.getCallInfo(c.activeCallID)
	if info == nil {
		return
	}

	info.mu.RLock()
	participant = info.participants[connId]
	info.mu.RUnlock()
	return
}

func (c *call) groupGetParticipantList(callID int64, excludeCurrent bool) (participants []*Participant) {
	info := c.getCallInfo(callID)
	if info == nil {
		return
	}

	c.mu.RLock()
	for _, participant := range info.participants {
		if excludeCurrent == false || participant.Peer.UserID == c.userID {
			if conn, ok := c.peerConnections[participant.ConnectionId]; ok && conn.StreamID != 0 {
				participant.Started = true
			}
			participants = append(participants, participant)
		}
	}
	c.mu.RUnlock()

	return
}

func (c *call) destroyConnections(callID int64) {
	closeFn := func(conn *Connection) {
		_ = c.callback.CloseConnection(conn.ConnId, true)
		if conn.connectTimout != nil {
			conn.connectTimout.Stop()
		}
		if conn.reconnectTimout != nil {
			conn.reconnectTimout.Stop()
		}
	}
	c.mu.RLock()
	for _, conn := range c.peerConnections {
		closeFn(conn)
	}
	c.mu.RUnlock()
	c.mu.Lock()
	c.peerConnections = make(map[int32]*Connection)
	delete(c.callInfo, callID)
	c.activeCallID = 0
	c.peer = nil
	c.mu.Unlock()
}

func (c *call) areAllAudio() bool {
	streamState := c.getStreamState()
	if streamState.Video {
		return false
	}

	isAllAudio := true
	participants := c.groupGetParticipantList(c.activeCallID, true)
	for _, participant := range participants {
		if participant.MediaSettings.Video {
			isAllAudio = false
			break
		}
	}
	return isAllAudio
}

func (c *call) groupMuteParticipant(userID int64, muted bool) {
	if c.activeCallID == 0 {
		return
	}

	connId, info, valid := c.getConnId(c.activeCallID, userID)
	if !valid {
		return
	}

	info.mu.Lock()
	info.participants[connId].Muted = muted
	info.mu.Unlock()

	update := msg.CallUpdateParticipantMuted{
		ConnectionID: connId,
		Muted:        muted,
		UserID:       userID,
	}
	updateData, uErr := update.Marshal()
	if uErr == nil {
		c.callUpdate(msg.CallUpdate_ParticipantMuted, updateData)
	}
}