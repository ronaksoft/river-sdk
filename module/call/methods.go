package call

import (
	"fmt"
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"go.uber.org/zap"
	"sync"
	"time"
)

func (c *call) toggleVideo(enable bool) (err error) {
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

func (c *call) destroy(callID int64) {
	closeFn := func(conn *Connection) {
		_ = c.callback.CloseConnection(conn.ConnId, true)
		if conn.connectTicker != nil {
			conn.connectTicker.Stop()
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

func (c *call) areAllAudio() (ok bool, err error) {
	streamState, err := c.getMediaSettings()
	if err != nil {
		return
	}

	if streamState.Video {
		ok = false
		return
	}

	ok = true
	participants, err := c.getParticipantList(c.activeCallID, true)
	if err != nil {
		return
	}

	for _, participant := range participants {
		if participant.MediaSettings.Video {
			ok = false
			break
		}
	}

	return
}

func (c *call) iceCandidate(connId int32, candidate *msg.CallRTCIceCandidate) (err error) {
	if c.activeCallID == 0 {
		err = ErrNoActiveCall
		return
	}

	err = c.sendIceCandidate(c.activeCallID, connId, candidate)
	return
}

func (c *call) iceConnectionStatusChange(connId int32, state string, hasIceError bool) (err error) {
	if c.activeCallID == 0 {
		err = ErrNoActiveCall
		return
	}

	conn, hasConn := c.peerConnections[connId]
	if !hasConn {
		err = ErrInvalidConnId
		return
	}

	conn.mu.Lock()
	conn.IceConnectionState = state
	conn.mu.Unlock()

	if !hasIceError {
		update := msg.CallUpdateConnectionStatusChanged{
			ConnectionID: connId,
			State:        state,
		}
		updateData, uErr := update.Marshal()
		if uErr == nil {
			c.callUpdate(msg.CallUpdate_ConnectionStatusChanged, updateData)
		}
		c.checkAllConnected()
	}
	err = c.checkDisconnection(connId, state, hasIceError)
	return
}

func (c *call) mediaSettingsChange(mediaSettings *msg.CallMediaSettings) (err error) {
	if c.activeCallID == 0 {
		err = ErrNoActiveCall
		return
	}

	info := c.getCallInfo(c.activeCallID)
	if info == nil {
		err = ErrInvalidCallId
		return
	}

	info.mu.Lock()
	info.mediaSettings = mediaSettings
	info.mu.Unlock()
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

func (c *call) join(peer *msg.InputPeer, callID int64) (err error) {
	if c.activeCallID == 0 {
		err = ErrNoActiveCall
		return
	}

	c.activeCallID = callID
	update := msg.CallUpdateCallPreview{
		CallID: callID,
		Peer:   peer,
	}
	updateData, err := update.Marshal()
	if err != nil {
		return
	}

	c.callUpdate(msg.CallUpdate_CallPreview, updateData)
	return
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

				streamState, innerErr := c.getMediaSettings()
				if innerErr != nil {
					return
				}

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

func (c *call) getParticipantByUserID(callID int64, userID int64) (participant *msg.CallParticipant, err error) {
	connId, info, valid := c.getConnId(callID, userID)
	if !valid {
		err = ErrInvalidCallId
		return
	}

	info.mu.RLock()
	participant = info.participants[connId]
	info.mu.RUnlock()
	return
}

func (c *call) getParticipantByConnId(connId int32) (participant *msg.CallParticipant, err error) {
	if c.activeCallID == 0 {
		err = ErrNoActiveCall
		return
	}

	info := c.getCallInfo(c.activeCallID)
	if info == nil {
		err = ErrInvalidCallId
		return
	}

	info.mu.RLock()
	participant = info.participants[connId]
	info.mu.RUnlock()
	return
}

func (c *call) getParticipantList(callID int64, excludeCurrent bool) (participants []*msg.CallParticipant, err error) {
	info := c.getCallInfo(callID)
	if info == nil {
		err = ErrInvalidCallId
		return
	}

	c.mu.RLock()
	for _, participant := range info.participants {
		if excludeCurrent == false || participant.PhoneParticipant.Peer.UserID == c.userID {
			if conn, ok := c.peerConnections[participant.PhoneParticipant.ConnectionId]; ok && conn.StreamID != 0 {
				participant.Started = true
			}
			participants = append(participants, participant)
		}
	}
	c.mu.RUnlock()

	return
}

func (c *call) muteParticipant(userID int64, muted bool) (err error) {
	if c.activeCallID == 0 {
		err = ErrNoActiveCall
		return
	}

	connId, info, valid := c.getConnId(c.activeCallID, userID)
	if !valid {
		err = ErrInvalidCallId
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

func (c *call) getMediaSettings() (ms *msg.CallMediaSettings, err error) {
	if c.activeCallID == 0 {
		err = ErrNoActiveCall
		return
	}

	info := c.getCallInfo(c.activeCallID)
	if info == nil {
		err = ErrInvalidCallId
		return
	}

	c.mu.RLock()
	ms = info.mediaSettings
	c.mu.RUnlock()
	return
}

func (c *call) initCallParticipants(callID int64, participants []*msg.InputUser) {
	participants = append([]*msg.InputUser{{
		UserID:     c.userID,
		AccessHash: 0,
	}}, participants...)
	callParticipants := make(map[int32]*msg.CallParticipant)
	callParticipantMap := make(map[int64]int32)
	for i, participant := range participants {
		idx := int32(i)
		callParticipants[idx] = &msg.CallParticipant{
			PhoneParticipant: &msg.PhoneParticipant{
				ConnectionId: idx,
				Peer:         participant,
				Initiator:    idx == 0,
				Admin:        idx == 0,
			},
			DeviceType: msg.CallDeviceType_CallDeviceUnknown,
			MediaSettings: &msg.CallMediaSettings{
				Audio:       true,
				ScreenShare: false,
				Video:       true,
			},
			Started: false,
			Muted:   false,
		}
		callParticipantMap[participant.UserID] = idx
	}

	c.mu.Lock()
	c.callInfo[callID] = &Info{
		acceptedParticipantIds: nil,
		acceptedParticipants:   nil,
		allConnected:           false,
		dialed:                 false,
		mediaSettings: &msg.CallMediaSettings{
			Audio:       false,
			ScreenShare: false,
			Video:       false,
		},
		participantMap:        callParticipantMap,
		participants:          callParticipants,
		requestParticipantIds: nil,
		requests:              nil,
		iceServer:             nil,
		mu:                    &sync.RWMutex{},
	}
	c.mu.Unlock()
}

func (c *call) initParticipants(callID int64, participants []*msg.PhoneParticipant, bootstrap bool) {
	fn := func(callParticipants map[int32]*msg.CallParticipant, callParticipantMap map[int64]int32) (map[int32]*msg.CallParticipant, map[int64]int32) {
		for _, participant := range participants {
			callParticipants[participant.ConnectionId] = &msg.CallParticipant{
				PhoneParticipant: &msg.PhoneParticipant{
					ConnectionId: participant.ConnectionId,
					Peer:         participant.Peer,
					Initiator:    participant.Initiator,
					Admin:        participant.Admin,
				},
				DeviceType: msg.CallDeviceType_CallDeviceUnknown,
				MediaSettings: &msg.CallMediaSettings{
					Audio:       true,
					ScreenShare: false,
					Video:       true,
				},
				Started: false,
				Muted:   false,
			}
			callParticipantMap[participant.Peer.UserID] = participant.ConnectionId
		}
		return callParticipants, callParticipantMap
	}

	if info, ok := c.callInfo[callID]; !ok {
		if bootstrap {
			callParticipants, callParticipantMap := fn(make(map[int32]*msg.CallParticipant), make(map[int64]int32))
			c.mu.Lock()
			c.callInfo[callID] = &Info{
				acceptedParticipantIds: nil,
				acceptedParticipants:   nil,
				allConnected:           false,
				dialed:                 false,
				mediaSettings: &msg.CallMediaSettings{
					Audio:       false,
					ScreenShare: false,
					Video:       false,
				},
				participantMap:        callParticipantMap,
				participants:          callParticipants,
				requestParticipantIds: nil,
				requests:              nil,
				iceServer:             nil,
				mu:                    &sync.RWMutex{},
			}
			c.mu.Unlock()
		}
	} else {
		info.mu.RLock()
		callParticipants, callParticipantMap := fn(info.participants, info.participantMap)
		info.mu.RUnlock()

		info.mu.Lock()
		info.participantMap = callParticipantMap
		info.participants = callParticipants
		info.mu.Unlock()
	}
}

func (c *call) initCallRequest(in *UpdatePhoneCall, sdpData *msg.PhoneActionRequested) {
	if info, ok := c.callInfo[in.CallID]; ok {
		requestParticipantIds := info.requestParticipantIds
		hasReq := false
		for idx := range requestParticipantIds {
			if requestParticipantIds[idx] == in.UserID {
				hasReq = true
				break
			}
		}

		if hasReq {
			info.mu.Lock()
			info.requests = append(info.requests, in)
			info.requestParticipantIds = append(info.requestParticipantIds, in.UserID)
			info.mu.Unlock()
			logs.Info("[webrtc] request from", zap.Int64("UserID", in.UserID))
		}
		return
	}

	logs.Info("[webrtc] request from", zap.Int64("UserID", in.UserID))
	callParticipants := make(map[int32]*msg.CallParticipant)
	callParticipantMap := make(map[int64]int32)
	for _, participant := range sdpData.Participants {
		deviceType := msg.CallDeviceType_CallDeviceUnknown
		if in.UserID == participant.Peer.UserID {
			deviceType = sdpData.DeviceType
		}
		callParticipants[participant.ConnectionId] = &msg.CallParticipant{
			PhoneParticipant: &msg.PhoneParticipant{
				ConnectionId: participant.ConnectionId,
				Peer:         participant.Peer,
				Initiator:    participant.Initiator,
				Admin:        participant.Admin,
			},
			DeviceType: deviceType,
			MediaSettings: &msg.CallMediaSettings{
				Audio:       true,
				ScreenShare: false,
				Video:       true,
			},
			Started: false,
			Muted:   false,
		}
		callParticipantMap[participant.Peer.UserID] = participant.ConnectionId
	}

	c.mu.Lock()
	c.callInfo[in.CallID] = &Info{
		acceptedParticipantIds: nil,
		acceptedParticipants:   nil,
		allConnected:           false,
		dialed:                 false,
		mediaSettings: &msg.CallMediaSettings{
			Audio:       false,
			ScreenShare: false,
			Video:       false,
		},
		participantMap:        callParticipantMap,
		participants:          callParticipants,
		requestParticipantIds: []int64{in.UserID},
		requests:              []*UpdatePhoneCall{in},
		iceServer:             nil,
		mu:                    &sync.RWMutex{},
	}
	c.mu.Unlock()
	return
}

func (c *call) initConnections(peer *msg.InputPeer, callID int64, initiator bool, request *UpdatePhoneCall) (res *msg.PhoneCall, err error) {
	currentUserConnId, callInfo, valid := c.getConnId(callID, c.userID)
	if !valid {
		err = ErrInvalidCallId
		return
	}

	wg := &sync.WaitGroup{}
	mu := &sync.RWMutex{}

	var callResults []*msg.PhoneParticipantSDP
	var acceptResults []*msg.PhoneCall

	sdp := &msg.PhoneActionSDPOffer{}
	requestConnId := int32(-1024)

	initAnswerConnection := func(connId int32) (res *msg.PhoneCall, innerErr error) {
		sdpOffer, innerErr := c.initConnection(true, connId, sdp)
		if innerErr != nil {
			return
		}

		p := callInfo.participants[connId]
		phoneParticipant := &msg.PhoneParticipantSDP{
			ConnectionId: p.PhoneParticipant.ConnectionId,
			Peer:         p.PhoneParticipant.Peer,
			SDP:          sdpOffer.SDP,
			Type:         sdpOffer.Type,
		}

		res, innerErr = c.apiAccept(peer, callID, []*msg.PhoneParticipantSDP{phoneParticipant})
		return
	}

	if request != nil {
		sdpData := request.Data.(*msg.PhoneActionRequested)
		sdp.SDP = sdpData.SDP
		sdp.Type = sdpData.Type
		requestConnId, _, valid := c.getConnId(callID, request.UserID)
		if valid && callInfo.dialed {
			return initAnswerConnection(requestConnId)
		}
	}

	shouldCall := !callInfo.dialed
	if shouldCall {
		c.setCallInfoDialed(callID)
	}

	for _, participant := range callInfo.participants {
		// Initialize connections only for greater connId,
		// full mesh initialization will take place here
		if requestConnId == participant.PhoneParticipant.ConnectionId {
			wg.Add(1)
			go func() {
				phoneCall, innerRes := initAnswerConnection(requestConnId)
				if innerRes == nil {
					mu.Lock()
					acceptResults = append(acceptResults, phoneCall)
					mu.Unlock()
				}
				wg.Done()
			}()
		} else if shouldCall && currentUserConnId < participant.PhoneParticipant.ConnectionId {
			wg.Add(1)
			go func(pConnId int32) {
				sdpRes, innerErr := c.initConnection(false, pConnId, nil)
				if innerErr == nil {
					mu.Lock()
					if participant, ok := callInfo.participants[pConnId]; ok {
						callResults = append(callResults, &msg.PhoneParticipantSDP{
							ConnectionId: participant.PhoneParticipant.ConnectionId,
							Peer:         participant.PhoneParticipant.Peer,
							SDP:          sdpRes.SDP,
							Type:         sdpRes.Type,
						})
					}
					mu.Unlock()
				}
				wg.Done()
			}(participant.PhoneParticipant.ConnectionId)
		}
	}

	wg.Wait()

	for _, participantSDP := range callResults {
		fmt.Println(participantSDP)
		// retry here
		if pc, ok := c.peerConnections[participantSDP.ConnectionId]; ok {
			pc.connectTicker = time.NewTicker(time.Duration(RetryInterval) * time.Second)
			go func(participant *msg.PhoneParticipantSDP) {
				select {
				case <-pc.connectTicker.C:
					if pc, ok := c.peerConnections[participant.ConnectionId]; ok {
						pc.mu.Lock()
						pc.Try++
						pc.mu.Unlock()
						_, innerErr := c.callUserSingle(peer, participant, c.activeCallID)
						if innerErr == nil {
							logs.Warn("callUserSingle", zap.Error(innerErr))
						}
						if pc.Try >= RetryLimit {
							if pc.connectTicker != nil {
								pc.connectTicker.Stop()
							}
							if initiator {
								c.checkCallTimeout(participant.ConnectionId)
							}
						}
					}
				}
			}(participantSDP)
		}
	}
	_, _ = c.callUser(peer, initiator, callResults, c.activeCallID)
	return
}

func (c *call) initConnection(remote bool, connId int32, sdp *msg.PhoneActionSDPOffer) (sdpAnswer *msg.PhoneActionSDPAnswer, err error) {
	logs.Debug("[webrtc] init connection", zap.Int32("connId", connId))
	// Client should check local stream
	// otherwise panic

	// Use MediaSteam to mix video and audio track
	// You can use main MediaStream if no shared screen media is present

	iceServer := c.iceServer
	if pc, ok := c.peerConnections[connId]; ok {
		iceServer = pc.IceServers
	}

	// Client should initiate RTCPeerConnection with given server config
	callInitReq := &msg.PhoneInit{
		IceServers: iceServer,
	}
	callInitData, err := callInitReq.Marshal()
	if err != nil {
		return
	}

	rtcConnId, err := c.callback.InitConnection(connId, callInitData)
	if err != nil {
		return
	}

	// Client should listen to icecandidate and send it to SDK

	// Client should listen to iceconnectionstatechange and send it to SDK

	// Client should listen to icecandidateerror and send it to SDK

	conn := &Connection{
		CallConnection: msg.CallConnection{
			Accepted:            remote,
			RTCPeerConnectionID: rtcConnId,
			IceConnectionState:  "",
			IceQueue:            nil,
			IceServers:          nil,
			Init:                false,
			Reconnecting:        false,
			ReconnectingTry:     0,
			ScreenShareStreamID: 0,
			StreamID:            0,
			IntervalID:          0,
			Try:                 0,
		},
		mu:              &sync.RWMutex{},
		connectTicker:   nil,
		reconnectTimout: nil,
	}

	conn.RTCPeerConnectionID = rtcConnId
	if pc, ok := c.peerConnections[connId]; !ok {
		c.peerConnections[connId] = conn
	} else {
		conn = pc
		conn.RTCPeerConnectionID = rtcConnId
	}

	// Client should listen to track and send it to SDK
	conn.Init = true
	conn.Reconnecting = false
	conn.ReconnectingTry = 0
	// clear reconnect timeout

	if remote {
		if sdp != nil {
			// Client should setRemoteDescription(sdp)
			// Client should create answer
			// Client should setLocalDescription and pass it to SDK
			offerSDP := &msg.PhoneActionSDPOffer{
				SDP:  sdp.SDP,
				Type: sdp.Type,
			}

			var offerSDPData []byte
			offerSDPData, err = offerSDP.Marshal()
			if err != nil {
				return
			}

			var clientAnswerSDP []byte
			clientAnswerSDP, err = c.callback.GetAnswerSDP(connId, offerSDPData)
			if err != nil {
				return
			}

			sdpAnswer = &msg.PhoneActionSDPAnswer{}
			err = sdpAnswer.Unmarshal(clientAnswerSDP)
			if err != nil {
				return
			}
		} else {
			err = ErrNoSDP
			return
		}
	} else {
		// Client should create offer
		// Client should setLocalDescription and pass the offer to SDK
		var clientOfferSDP []byte
		clientOfferSDP, err = c.callback.GetOfferSDP(connId)
		if err != nil {
			return
		}

		sdpAnswer = &msg.PhoneActionSDPAnswer{}
		err = sdpAnswer.Unmarshal(clientOfferSDP)
		if err != nil {
			return
		}
	}
	return
}

func (c *call) callUser(peer *msg.InputPeer, initiator bool, phoneParticipants []*msg.PhoneParticipantSDP, callID int64) (res *msg.PhoneCall, err error) {
	randomID := domain.RandomInt64(0)
	res, err = c.apiRequest(peer, randomID, initiator, phoneParticipants, callID, false)
	if err == nil && callID == 0 {
		c.activeCallID = res.ID
	}
	return
}

func (c *call) callUserSingle(peer *msg.InputPeer, phoneParticipant *msg.PhoneParticipantSDP, callID int64) (res *msg.PhoneCall, err error) {
	randomID := domain.RandomInt64(0)
	res, err = c.apiRequest(peer, randomID, false, []*msg.PhoneParticipantSDP{phoneParticipant}, callID, true)
	return
}

func (c *call) sendIceCandidate(callID int64, connId int32, candidate *msg.CallRTCIceCandidate) (err error) {
	if candidate == nil {
		return nil
	}

	conn, hasConn := c.peerConnections[connId]
	if !hasConn {
		err = ErrInvalidConnId
		return
	}

	if !conn.Accepted {
		conn.IceQueue = append(conn.IceQueue, candidate)
		return
	}

	if c.peer == nil {
		err = ErrInvalidPeerInput
		return
	}

	inputUser := c.getInputUserByConnId(callID, connId)
	if inputUser == nil {
		err = ErrInvalidConnId
		return
	}

	action := &msg.PhoneActionIceExchange{
		Candidate:        candidate.Candidate,
		SdpMLineIndex:    candidate.SdpMLineIndex,
		SdpMid:           candidate.SdpMid,
		UsernameFragment: candidate.UsernameFragment,
	}

	actionData, err := action.Marshal()
	if err != nil {
		return
	}

	_, err = c.apiSendUpdate(c.peer, callID, []*msg.InputUser{inputUser}, msg.PhoneCallAction_PhoneCallIceExchange, actionData, false)
	return
}

func (c *call) flushIceCandidates(callID int64, connId int32) {
	c.mu.RLock()
	conn, ok := c.peerConnections[connId]
	c.mu.RUnlock()

	if !ok {
		return
	}

	for _, candidate := range conn.IceQueue {
		go func(ic *msg.CallRTCIceCandidate) {
			_ = c.sendIceCandidate(callID, connId, ic)
		}(candidate)
	}
}

func (c *call) modifyMediaStream(video bool) (err error) {
	if c.activeCallID == 0 {
		err = ErrNoActiveCall
		return
	}

	err = c.callback.InitStream(true, video)
	if err != nil {
		return
	}

	_ = c.upgradeConnection(video)
	c.propagateMediaSettings(MediaSettingsIn{
		Video: &video,
	})
	return
}

func (c *call) upgradeConnection(video bool) (err error) {
	if c.activeCallID == 0 {
		return
	}

	err = c.callback.InitStream(true, video)
	if err != nil {
		return
	}

	var connIds []int32
	c.mu.RLock()
	for _, pc := range c.peerConnections {
		if pc.IceConnectionState == "connected" {
			connIds = append(connIds, pc.ConnId)
		}
	}
	c.mu.RUnlock()

	if len(connIds) == 0 {
		return
	}

	wg := sync.WaitGroup{}
	for _, connId := range connIds {
		wg.Add(1)
		go func(cid int32) {
			var clientOfferSDP []byte
			clientOfferSDP, err = c.callback.GetOfferSDP(cid)
			if err != nil {
				return
			}

			sdpOffer := &msg.PhoneActionSDPOffer{}
			err = sdpOffer.Unmarshal(clientOfferSDP)
			if err != nil {
				return
			}

			c.sendSdpOffer(cid, sdpOffer)
			wg.Done()
		}(connId)
	}

	wg.Wait()
	return
}

func (c *call) propagateMediaSettings(in MediaSettingsIn) {
	if c.activeCallID == 0 {
		return
	}

	if c.peer == nil {
		return
	}

	c.mu.RLock()
	info, ok := c.callInfo[c.activeCallID]
	c.mu.RUnlock()
	if !ok {
		return
	}

	if in.Audio != nil {
		info.mediaSettings.Audio = *in.Audio
	}

	if in.Video != nil {
		info.mediaSettings.Video = *in.Video
	}

	if in.ScreenShare != nil {
		info.mediaSettings.ScreenShare = *in.ScreenShare
	}

	action := &msg.PhoneActionMediaSettingsUpdated{
		Video:       info.mediaSettings.Video,
		Audio:       info.mediaSettings.Audio,
		ScreenShare: info.mediaSettings.ScreenShare,
	}
	actionData, err := action.Marshal()
	if err == nil {
		return
	}

	update := msg.CallUpdateLocalMediaSettingsUpdated{
		MediaSettings: &msg.CallMediaSettings{
			Video:       info.mediaSettings.Video,
			Audio:       info.mediaSettings.Audio,
			ScreenShare: info.mediaSettings.ScreenShare,
		},
	}
	updateData, uErr := update.Marshal()
	if uErr == nil {
		c.callUpdate(msg.CallUpdate_LocalMediaSettingsUpdated, updateData)
	}

	inputUsers := c.getInputUsers(c.activeCallID)
	_, _ = c.apiSendUpdate(c.peer, c.activeCallID, inputUsers, msg.PhoneCallAction_PhoneCallMediaSettingsChanged, actionData, false)
	return
}

func (c *call) mediaSettingsInit(in *msg.CallMediaSettings) {
	if c.activeCallID == 0 {
		return
	}

	connId, info, valid := c.getConnId(c.activeCallID, c.userID)
	if !valid {
		return
	}

	info.mu.Lock()
	info.participants[connId].MediaSettings.Audio = in.Audio
	info.participants[connId].MediaSettings.Video = in.Video
	info.participants[connId].MediaSettings.ScreenShare = in.ScreenShare
	info.mu.Unlock()

	update := msg.CallUpdateMediaSettingsUpdated{
		ConnectionID: connId,
		MediaSettings: &msg.CallMediaSettings{
			Video:       in.Video,
			Audio:       in.Audio,
			ScreenShare: in.ScreenShare,
		},
	}
	updateData, uErr := update.Marshal()
	if uErr == nil {
		c.callUpdate(msg.CallUpdate_MediaSettingsUpdated, updateData)
	}
}

func (c *call) checkAllConnected() {
	if c.activeCallID == 0 {
		return
	}
	if info, ok := c.callInfo[c.activeCallID]; ok {
		if info.allConnected {
			return
		}

		for _, pc := range c.peerConnections {
			if pc.IceConnectionState != "connected" {
				return
			}
		}

		c.callInfo[c.activeCallID].allConnected = true
		time.AfterFunc(time.Duration(255)*time.Millisecond, func() {
			update := msg.CallUpdateAllConnected{}
			updateData, uErr := update.Marshal()
			if uErr == nil {
				c.callUpdate(msg.CallUpdate_AllConnected, updateData)
			}
		})
	}
}

func (c *call) checkDisconnection(connId int32, state string, isIceError bool) (err error) {
	if c.activeCallID == 0 {
		return
	}

	conn, hasConn := c.peerConnections[connId]
	if !hasConn {
		return
	}

	if !conn.Reconnecting &&
		((isIceError && c.peerConnections[connId].Init && (state == "disconnected" || state == "failed" || state == "closed")) ||
			state == "disconnected") {
		err = c.callback.CloseConnection(connId, false)
		if err != nil {
			return
		}

		conn.mu.Lock()
		conn.IceQueue = nil
		conn.Reconnecting = true
		conn.ReconnectingTry++
		if conn.ReconnectingTry <= ReconnectTry {
			conn.reconnectTimout = time.AfterFunc(time.Duration(ReconnectTimeout)*time.Second, func() {
				if _, ok := c.peerConnections[connId]; ok {
					c.peerConnections[connId].Reconnecting = false
				}
			})
		}
		conn.mu.Unlock()

		update := msg.CallUpdateConnectionStatusChanged{
			ConnectionID: connId,
			State:        "reconnecting",
		}
		updateData, uErr := update.Marshal()
		if uErr == nil {
			c.callUpdate(msg.CallUpdate_ConnectionStatusChanged, updateData)
		}

		var initRes *msg.PhoneInit
		initRes, err = c.apiInit(c.peer, c.activeCallID)
		if err != nil {
			return
		}

		_, hasConn = c.peerConnections[connId]
		if !hasConn {
			return
		}

		conn.mu.Lock()
		conn.IceServers = initRes.IceServers
		conn.mu.Unlock()
		currentConnId, _, valid := c.getConnId(c.activeCallID, c.userID)
		if !valid {
			err = ErrInvalidCallId
			return
		}

		if currentConnId < connId {
			_ = c.callSendRestart(connId, true)
		} else {
			_, _ = c.initConnection(true, connId, nil)
		}
	}

	return
}

func (c *call) callSendRestart(connId int32, sender bool) (err error) {
	if c.activeCallID == 0 {
		err = ErrNoActiveCall
		return
	}

	_, hasConn := c.peerConnections[connId]
	if !hasConn {
		err = ErrInvalidConnId
		return
	}

	if c.peer == nil {
		err = ErrInvalidPeerInput
		return
	}

	inputUser := c.getInputUserByConnId(c.activeCallID, connId)
	if inputUser == nil {
		err = ErrInvalidConnId
		return
	}

	action := &msg.PhoneActionRestarted{
		Sender: sender,
	}

	actionData, err := action.Marshal()
	if err != nil {
		return
	}

	_, err = c.apiSendUpdate(c.peer, c.activeCallID, []*msg.InputUser{inputUser}, msg.PhoneCallAction_PhoneCallRestarted, actionData, true)
	return
}

func (c *call) getCallInfo(callID int64) *Info {
	c.mu.RLock()
	info, ok := c.callInfo[callID]
	c.mu.RUnlock()
	if ok {
		return info
	} else {
		return nil
	}
}

func (c *call) getConnId(callID, userID int64) (int32, *Info, bool) {
	info := c.getCallInfo(callID)
	if info == nil {
		return 0, nil, false
	}

	info.mu.RLock()
	connId := int32(info.participantMap[userID])
	info.mu.RUnlock()
	return connId, info, true
}

func (c *call) getUserIDbyCallID(callID int64, connID int32) *int64 {
	info := c.getCallInfo(callID)
	if info == nil {
		return nil
	}

	info.mu.RUnlock()
	d, ok := info.participants[connID]
	info.mu.RUnlock()
	if ok {
		return &d.PhoneParticipant.Peer.UserID
	} else {
		return nil
	}
}

func (c *call) getInputUserByConnId(callID int64, connID int32) (inputUser *msg.InputUser) {
	info := c.getCallInfo(callID)
	if info == nil {
		return nil
	}

	info.mu.RLock()
	d, ok := info.participants[connID]
	info.mu.RUnlock()
	if ok {
		return d.PhoneParticipant.Peer
	} else {
		return nil
	}
}

func (c *call) getInputUserByUserIDs(callID int64, userIDs []int64) (inputUser []*msg.InputUser) {
	info := c.getCallInfo(callID)
	if info == nil {
		return nil
	}

	inputUser = make([]*msg.InputUser, 0, len(userIDs))
	info.mu.RLock()
	for _, userID := range userIDs {
		if connId, ok := info.participantMap[userID]; ok {
			if participant, ok2 := info.participants[connId]; ok2 {
				inputUser = append(inputUser, participant.PhoneParticipant.Peer)
			}
		}
	}
	info.mu.RUnlock()
	return
}

func (c *call) getInputUsers(callID int64) (inputUser []*msg.InputUser) {
	info := c.getCallInfo(callID)
	if info == nil {
		return nil
	}

	inputUser = make([]*msg.InputUser, 0, len(info.participants))
	info.mu.RLock()
	for _, participant := range info.participants {
		inputUser = append(inputUser, participant.PhoneParticipant.Peer)
	}
	info.mu.RUnlock()
	return
}

func (c *call) swapTempInfo(callID int64) {
	info := c.getCallInfo(TempCallID)
	if info == nil {
		return
	}

	c.mu.Lock()
	c.callInfo[callID] = info
	delete(c.callInfo, TempCallID)
	c.mu.Unlock()
}

func (c *call) setCallInfoDialed(callID int64) {
	c.mu.Lock()
	if _, ok := c.callInfo[callID]; ok {
		c.callInfo[callID].dialed = true
	}
	c.mu.Unlock()
}

func (c *call) callBusy(in *UpdatePhoneCall) {
	inputPeer := c.getInputUserFromUpdate(in)

	_, _ = c.apiReject(inputPeer, in.CallID, msg.DiscardReason_DiscardReasonBusy, 0)
	return
}

func (c *call) sendCallAck(in *UpdatePhoneCall) {
	inputPeer := c.getInputUserFromUpdate(in)

	inputUser := &msg.InputUser{
		UserID:     in.UserID,
		AccessHash: in.AccessHash,
	}

	action := &msg.PhoneActionAck{}
	actionData, err := action.Marshal()
	if err != nil {
		return
	}

	_, _ = c.apiSendUpdate(inputPeer, in.CallID, []*msg.InputUser{inputUser}, msg.PhoneCallAction_PhoneCallAck, actionData, false)
	return
}

func (c *call) sendSdpAnswer(connId int32, sdp *msg.PhoneActionSDPAnswer) {
	if c.activeCallID == 0 {
		return
	}

	if c.peer == nil {
		return
	}

	inputUser := c.getInputUserByConnId(c.activeCallID, connId)
	actionData, err := sdp.Marshal()
	if err != nil {
		return
	}

	_, _ = c.apiSendUpdate(c.peer, c.activeCallID, []*msg.InputUser{inputUser}, msg.PhoneCallAction_PhoneCallSDPAnswer, actionData, false)
	return
}

func (c *call) sendSdpOffer(connId int32, sdp *msg.PhoneActionSDPOffer) {
	if c.activeCallID == 0 {
		return
	}

	if c.peer == nil {
		return
	}

	inputUser := c.getInputUserByConnId(c.activeCallID, connId)
	actionData, err := sdp.Marshal()
	if err != nil {
		return
	}

	_, _ = c.apiSendUpdate(c.peer, c.activeCallID, []*msg.InputUser{inputUser}, msg.PhoneCallAction_PhoneCallSDPOffer, actionData, false)
	return
}

func (c *call) getInputUserFromUpdate(in *UpdatePhoneCall) *msg.InputPeer {
	inputPeer := &msg.InputPeer{
		ID:   in.PeerID,
		Type: msg.PeerType(in.PeerType),
	}
	if in.PeerType == int32(msg.PeerType_PeerGroup) {
		inputPeer.AccessHash = 0
	} else {
		inputPeer.AccessHash = in.AccessHash
	}

	return inputPeer
}

func (c *call) clearRetryInterval(connId int32, onlyClearInterval bool) {
	c.mu.RLock()
	pc, ok := c.peerConnections[connId]
	c.mu.RUnlock()

	if !ok {
		return
	}

	if pc.connectTicker != nil {
		pc.connectTicker.Stop()
	}

	if onlyClearInterval == false && c.activeCallID == 0 {
		c.mu.Lock()
		if info, ok := c.callInfo[c.activeCallID]; ok {
			info.acceptedParticipants = append(info.acceptedParticipants, connId)
		}
		c.mu.Unlock()
	}
}

func (c *call) removeParticipant(userID int64, callID *int64) bool {
	activeCallID := c.activeCallID
	if callID != nil {
		activeCallID = *callID
	}

	if activeCallID == 0 {
		return false
	}

	connId, info, valid := c.getConnId(activeCallID, userID)
	if !valid {
		return false
	}

	c.mu.RLock()
	_, hasConn := c.peerConnections[connId]
	c.mu.RUnlock()
	if hasConn {
		_ = c.callback.CloseConnection(connId, false)
	}

	info.mu.Lock()
	for idx, id := range info.acceptedParticipantIds {
		if id == userID {
			info.acceptedParticipantIds = append(info.acceptedParticipantIds[:idx], info.acceptedParticipantIds[idx+1:]...)
		}
	}

	for idx, id := range info.requestParticipantIds {
		if id == userID {
			info.requestParticipantIds = append(info.requestParticipantIds[:idx], info.requestParticipantIds[idx+1:]...)
		}
	}

	for idx, request := range info.requests {
		if request.UserID == userID {
			info.requests = append(info.requests[:idx], info.requests[idx+1:]...)
		}
	}

	delete(info.participants, connId)
	delete(info.participantMap, userID)
	allRemoved := len(info.participants) <= 1
	info.mu.Unlock()

	return allRemoved
}

func (c *call) updateAdmin(userID int64, admin bool) {
	if c.activeCallID == 0 {
		return
	}

	connId, info, valid := c.getConnId(c.activeCallID, userID)
	if !valid {
		return
	}

	info.mu.Lock()
	info.participants[connId].PhoneParticipant.Admin = admin
	info.mu.Unlock()
}

func (c *call) shouldAccept(in *UpdatePhoneCall) bool {
	if c.activeCallID == in.CallID {
		return false
	}

	if c.peer.GetType() == msg.PeerType_PeerUser {
		return false
	}

	if _, ok := c.callInfo[c.activeCallID]; !ok {
		return false
	}

	c.mu.RLock()
	info, ok := c.callInfo[c.activeCallID]
	c.mu.RUnlock()

	if !ok {
		return false
	}

	if ok {
		info.mu.RLock()
		defer info.mu.RUnlock()
		for idx := range info.acceptedParticipantIds {
			if info.acceptedParticipantIds[idx] == in.UserID {
				return false
			}
		}
	}

	info.mu.Lock()
	info.acceptedParticipantIds = append(info.acceptedParticipantIds, in.UserID)
	info.mu.Unlock()
	return true
}

func (c *call) callRequested(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionRequested)
	if c.activeCallID != 0 && c.activeCallID != in.CallID {
		c.callBusy(in)
		return
	}

	// Send ack update so callee ringing indicator activates
	c.sendCallAck(in)
	if _, ok := c.callInfo[in.CallID]; !ok {
		c.initCallRequest(in, data)

		c.mu.Lock()
		c.peer = c.getInputUserFromUpdate(in)
		c.mu.Unlock()

		update := msg.CallUpdateCallRequested{
			Peer: &msg.InputPeer{
				ID:         in.PeerID,
				Type:       msg.PeerType(in.PeerType),
				AccessHash: in.AccessHash,
			},
			CallID: 0,
		}
		updateData, uErr := update.Marshal()
		if uErr == nil {
			c.callUpdate(msg.CallUpdate_CallRequested, updateData)
		}
	} else if c.shouldAccept(in) {
		c.initCallRequest(in, data)
		video := false
		if streamState, err := c.getMediaSettings(); err == nil {
			video = streamState.Video
		}
		_ = c.accept(c.activeCallID, video)
	}
}

func (c *call) callAccepted(in *UpdatePhoneCall) {
	if c.activeCallID == 0 {
		return
	}

	connId, info, valid := c.getConnId(in.CallID, in.UserID)
	if !valid {
		return
	}

	c.mu.RLock()
	pc, ok := c.peerConnections[connId]
	c.mu.RUnlock()
	if !ok {
		return
	}

	data := in.Data.(*msg.PhoneActionAccepted)
	info.mu.Lock()
	info.participants[connId].DeviceType = data.DeviceType
	info.mu.Unlock()

	sdp := &msg.PhoneActionSDPAnswer{
		SDP:  data.SDP,
		Type: data.Type,
	}
	sdpData, err := sdp.Marshal()
	if err != nil {
		return
	}

	err = c.callback.SetAnswerSDP(connId, sdpData)
	if err != nil {
		return
	}

	pc.mu.Lock()
	pc.Accepted = true
	pc.mu.Unlock()
	c.flushIceCandidates(in.CallID, connId)

	streamState, err := c.getMediaSettings()
	if err != nil {
		return
	}

	c.propagateMediaSettings(MediaSettingsIn{
		Audio:       &streamState.Audio,
		ScreenShare: &streamState.ScreenShare,
		Video:       &streamState.Video,
	})
	c.clearRetryInterval(connId, false)
	logs.Info("[webrtc] accept signal", zap.Int32("connId", connId))

	update := msg.CallUpdateCallAccepted{
		ConnectionID: connId,
	}
	updateData, uErr := update.Marshal()
	if uErr == nil {
		c.callUpdate(msg.CallUpdate_CallAccepted, updateData)
	}
}

func (c *call) callDiscarded(in *UpdatePhoneCall) {
	connId, _, valid := c.getConnId(in.CallID, in.UserID)
	if !valid {
		return
	}

	c.clearRetryInterval(connId, false)

	data := in.Data.(*msg.PhoneActionDiscarded)
	if in.PeerType == int32(msg.PeerType_PeerUser) || data.Terminate {
		update := msg.CallUpdateCallRejected{
			CallID: in.CallID,
			Reason: data.Reason,
		}
		updateData, uErr := update.Marshal()
		if uErr == nil {
			c.callUpdate(msg.CallUpdate_CallRejected, updateData)
		}
	} else {
		if c.removeParticipant(in.UserID, &in.CallID) {
			update := msg.CallUpdateCallRejected{
				CallID: in.CallID,
				Reason: data.Reason,
			}
			updateData, uErr := update.Marshal()
			if uErr == nil {
				c.callUpdate(msg.CallUpdate_CallRejected, updateData)
			}
		} else {
			c.checkAllConnected()
			update := msg.CallUpdateParticipantLeft{
				UserID: in.UserID,
			}
			updateData, uErr := update.Marshal()
			if uErr == nil {
				c.callUpdate(msg.CallUpdate_ParticipantLeft, updateData)
			}
		}
	}
}

func (c *call) iceExchange(in *UpdatePhoneCall) {
	connId, _, valid := c.getConnId(in.CallID, in.UserID)
	if !valid {
		return
	}

	c.mu.RLock()
	_, hasConn := c.peerConnections[connId]
	c.mu.RUnlock()

	if !hasConn {
		return
	}

	data := in.Data.(*msg.PhoneActionIceExchange)

	iceCandidate := &msg.CallRTCIceCandidate{
		Candidate:        data.Candidate,
		SdpMLineIndex:    data.SdpMLineIndex,
		SdpMid:           data.SdpMid,
		UsernameFragment: data.UsernameFragment,
	}
	iceCandidateData, err := iceCandidate.Marshal()
	if err != nil {
		return
	}

	_ = c.callback.AddIceCandidate(connId, iceCandidateData)
}

func (c *call) mediaSettingsUpdated(in *UpdatePhoneCall) {
	connId, info, valid := c.getConnId(in.CallID, in.UserID)
	if !valid {
		return
	}

	data := in.Data.(*msg.PhoneActionMediaSettingsUpdated)

	info.mu.Lock()
	info.participants[connId].MediaSettings.Audio = data.Audio
	info.participants[connId].MediaSettings.Video = data.Video
	info.participants[connId].MediaSettings.ScreenShare = data.ScreenShare
	info.mu.Unlock()

	// TOO Call -> msg.CallUpdate_MediaSettingsUpdated
}

func (c *call) sdpOfferUpdated(in *UpdatePhoneCall) {
	if c.activeCallID != in.CallID {
		return
	}

	connId, _, valid := c.getConnId(in.CallID, in.UserID)
	if !valid {
		return
	}

	c.mu.RLock()
	conn, hasConn := c.peerConnections[connId]
	c.mu.RUnlock()

	if !hasConn {
		return
	}

	data := in.Data.(*msg.PhoneActionSDPOffer)

	offerSDP := &msg.PhoneActionSDPOffer{
		SDP:  data.SDP,
		Type: data.Type,
	}
	offerSDPAction, err := offerSDP.Marshal()
	if err != nil {
		return
	}

	clientAnswerSDP, err := c.callback.GetAnswerSDP(connId, offerSDPAction)
	if err != nil {
		return
	}

	conn.mu.Lock()
	conn.Accepted = true
	conn.mu.Unlock()

	c.flushIceCandidates(in.CallID, connId)

	sdpAnswer := &msg.PhoneActionSDPAnswer{}
	err = sdpAnswer.Unmarshal(clientAnswerSDP)
	if err != nil {
		return
	}

	c.sendSdpAnswer(connId, sdpAnswer)
}

func (c *call) sdpAnswerUpdated(in *UpdatePhoneCall) {
	if c.activeCallID != in.CallID {
		return
	}

	connId, _, valid := c.getConnId(in.CallID, in.UserID)
	if !valid {
		return
	}

	c.mu.RLock()
	_, hasConn := c.peerConnections[connId]
	c.mu.RUnlock()

	if !hasConn {
		return
	}

	data := in.Data.(*msg.PhoneActionSDPAnswer)

	answerSDP := &msg.PhoneActionSDPAnswer{
		SDP:  data.SDP,
		Type: data.Type,
	}
	answerSDPData, err := answerSDP.Marshal()
	if err != nil {
		return
	}

	_ = c.callback.SetAnswerSDP(connId, answerSDPData)
}

func (c *call) callAcknowledged(in *UpdatePhoneCall) {
	connId, _, valid := c.getConnId(in.CallID, in.UserID)
	if !valid {
		return
	}

	c.clearRetryInterval(connId, true)
	update := msg.CallUpdateCallAck{
		ConnectionID: connId,
	}
	updateData, uErr := update.Marshal()
	if uErr == nil {
		c.callUpdate(msg.CallUpdate_CallAck, updateData)
	}
}

func (c *call) participantAdded(in *UpdatePhoneCall) {
	if c.activeCallID != in.CallID {
		return
	}

	data := in.Data.(*msg.PhoneActionParticipantAdded)
	c.initParticipants(c.activeCallID, data.Participants, false)
	isNew := true
	var userIDs []int64
	for _, participant := range data.Participants {
		if participant.Peer.UserID == c.userID {
			isNew = false
		}
		userIDs = append(userIDs, participant.Peer.UserID)
	}

	if isNew {
		update := msg.CallUpdateParticipantJoined{
			UserIDs: userIDs,
		}
		updateData, uErr := update.Marshal()
		if uErr == nil {
			c.callUpdate(msg.CallUpdate_ParticipantJoined, updateData)
		}
	}
}

func (c *call) participantRemoved(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionParticipantRemoved)

	for _, userId := range data.UserIDs {
		if userId == c.userID {
			update := msg.CallUpdateCallCancelled{
				CallID: in.CallID,
			}
			updateData, uErr := update.Marshal()
			if uErr == nil {
				c.callUpdate(msg.CallUpdate_CallCancelled, updateData)
			}
			break
		}
	}

	if c.activeCallID != in.CallID {
		for _, userId := range data.UserIDs {
			c.removeParticipant(userId, &in.CallID)
		}
		return
	}

	isCurrentRemoved := false
	for _, userId := range data.UserIDs {
		c.removeParticipant(userId, nil)
		if userId == c.userID {
			isCurrentRemoved = true
		}
	}

	update := msg.CallUpdateParticipantRemoved{
		UserIDs: data.UserIDs,
		Timeout: data.Timeout,
	}
	updateData, uErr := update.Marshal()
	if uErr == nil {
		c.callUpdate(msg.CallUpdate_ParticipantRemoved, updateData)
	}
	if isCurrentRemoved {
		update := msg.CallUpdateCallRejected{
			CallID: in.CallID,
			Reason: msg.DiscardReason_DiscardReasonHangup,
		}
		updateData, uErr := update.Marshal()
		if uErr == nil {
			c.callUpdate(msg.CallUpdate_CallRejected, updateData)
		}
	}

	c.checkAllConnected()
}

func (c *call) adminUpdated(in *UpdatePhoneCall) {
	if c.activeCallID != in.CallID {
		return
	}

	data := in.Data.(*msg.PhoneActionAdminUpdated)
	c.updateAdmin(data.UserID, data.Admin)
	update := msg.CallUpdateParticipantAdminUpdated{
		UserID: data.UserID,
		Admin:  data.Admin,
	}
	updateData, uErr := update.Marshal()
	if uErr == nil {
		c.callUpdate(msg.CallUpdate_ParticipantAdminUpdated, updateData)
	}
}

func (c *call) joinRequested(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionJoinRequested)
	for _, userId := range data.UserIDs {
		if userId == c.userID {
			update := msg.CallUpdateCallJoinRequested{
				CallID:   in.CallID,
				CalleeID: userId,
				Peer: &msg.InputPeer{
					ID:         in.PeerID,
					Type:       msg.PeerType(in.PeerType),
					AccessHash: 0,
				},
			}
			updateData, uErr := update.Marshal()
			if uErr == nil {
				c.callUpdate(msg.CallUpdate_CallJoinRequested, updateData)
			}
			return
		}
	}
}

func (c *call) screenShareUpdated(in *UpdatePhoneCall) {
	//data := in.Data.(*msg.PhoneActionScreenShare)
	//fmt.Println(data)
}

func (c *call) callPicked(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionPicked)
	if data.AuthID != c.authID {
		update := msg.CallUpdateCallCancelled{
			CallID: in.CallID,
		}
		updateData, uErr := update.Marshal()
		if uErr == nil {
			c.callUpdate(msg.CallUpdate_CallCancelled, updateData)
		}
	}
}

func (c *call) callRestarted(in *UpdatePhoneCall) {
	connId, _, valid := c.getConnId(in.CallID, in.UserID)
	if !valid {
		return
	}
	data := in.Data.(*msg.PhoneActionRestarted)
	if data.Sender {
		_ = c.checkDisconnection(connId, "disconnected", false)
		_ = c.callSendRestart(connId, false)
	} else {
		sdp, err := c.initConnection(false, connId, nil)
		if err != nil {
			return
		}

		sdpOffer := &msg.PhoneActionSDPOffer{
			SDP:  sdp.SDP,
			Type: sdp.Type,
		}
		c.sendSdpOffer(connId, sdpOffer)
	}
}

func (c *call) checkCallTimeout(connId int32) {
	if c.activeCallID == 0 {
		return
	}

	if c.peer == nil {
		return
	}

	info := c.getCallInfo(c.activeCallID)
	if info == nil {
		return
	}

	if len(info.acceptedParticipants) == 0 {
		_ = c.reject(c.activeCallID, 0, msg.DiscardReason_DiscardReasonMissed, nil)
		update := msg.CallUpdateCallTimeout{}
		updateData, uErr := update.Marshal()
		if uErr == nil {
			c.callUpdate(msg.CallUpdate_CallTimeout, updateData)
		}
		logs.Info("[webrtc] call timeout", zap.Int32("ConnId", connId))
	} else if c.peer.GetType() == msg.PeerType_PeerGroup {
		var notAnsweringUserIDs []int64
		info.mu.RLock()
		for _, conn := range info.participants {
			matched := false
			for _, connId := range info.acceptedParticipants {
				if conn.PhoneParticipant.ConnectionId == connId {
					matched = true
					break
				}
			}
			if !matched {
				notAnsweringUserIDs = append(notAnsweringUserIDs, conn.PhoneParticipant.Peer.UserID)
			}
		}
		info.mu.RUnlock()
		if len(notAnsweringUserIDs) > 0 {
			_ = c.groupRemoveParticipant(c.activeCallID, notAnsweringUserIDs, true)
		}
	}
}

func (c *call) callUpdate(action msg.CallUpdate, b []byte) {
	c.callback.OnUpdate(int32(action), b)
}
