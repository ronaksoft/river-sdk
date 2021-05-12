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
	// TODO Call -> msg.CallUpdate_CallPreview
}

func (c *call) callAccept(callID int64, video bool) (err error) {
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

	// TODO Call -> msg.CallUpdate_ParticipantRemoved
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
	// TODO Call -> msg.CallUpdate_ParticipantAdminUpdated

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

	// TODO Call -> msg.CallUpdate_ParticipantMuted
}

func (c *call) updatePhoneCall(u *msg.UpdateEnvelope) ([]*msg.UpdateEnvelope, error) {
	go func() {
		if u.Constructor != msg.C_UpdatePhoneCall {
			return
		}
		x := &msg.UpdatePhoneCall{}
		err := x.Unmarshal(u.Update)
		if err != nil {
			return
		}

		now := domain.Now().Unix()
		if !(x.Timestamp == 0 || now-x.Timestamp < 60) {
			return
		}

		data, err := parseCallAction(x.Action, x.ActionData)
		if err != nil {
			logs.Debug("parseCallAction", zap.Error(err))
			return
		}

		update := &UpdatePhoneCall{
			UpdatePhoneCall: x,
			Data:            data,
		}

		switch x.Action {
		case msg.PhoneCallAction_PhoneCallRequested:
			c.callRequested(update)
		case msg.PhoneCallAction_PhoneCallAccepted:
			c.callAccepted(update)
		case msg.PhoneCallAction_PhoneCallDiscarded:
			c.callDiscarded(update)
		case msg.PhoneCallAction_PhoneCallIceExchange:
			c.iceExchange(update)
		case msg.PhoneCallAction_PhoneCallMediaSettingsChanged:
			c.mediaSettingsUpdated(update)
		case msg.PhoneCallAction_PhoneCallSDPOffer:
			c.sdpOfferUpdated(update)
		case msg.PhoneCallAction_PhoneCallSDPAnswer:
			c.sdpAnswerUpdated(update)
		case msg.PhoneCallAction_PhoneCallAck:
			c.callAcknowledged(update)
		case msg.PhoneCallAction_PhoneCallParticipantAdded:
			c.participantAdded(update)
		case msg.PhoneCallAction_PhoneCallParticipantRemoved:
			c.participantRemoved(update)
		case msg.PhoneCallAction_PhoneCallAdminUpdated:
			c.adminUpdated(update)
		case msg.PhoneCallAction_PhoneCallJoinRequested:
			c.joinRequested(update)
		case msg.PhoneCallAction_PhoneCallScreenShare:
			c.screenShareUpdated(update)
		case msg.PhoneCallAction_PhoneCallPicked:
			c.callPicked(update)
		case msg.PhoneCallAction_PhoneCallRestarted:
			c.callRestarted(update)
		}
	}()

	res := []*msg.UpdateEnvelope{u}
	return res, nil
}

func (c *call) getStreamState() MediaSettings {
	return MediaSettings{
		Audio:       true,
		ScreenShare: false,
		Video:       true,
	}
}

func (c *call) initCallParticipants(callID int64, participants []*msg.InputUser) {
	participants = append([]*msg.InputUser{{
		UserID:     c.userID,
		AccessHash: 0,
	}}, participants...)
	callParticipants := make(map[int32]*Participant)
	callParticipantMap := make(map[int64]int32)
	for i, participant := range participants {
		idx := int32(i)
		callParticipants[idx] = &Participant{
			PhoneParticipant: msg.PhoneParticipant{
				ConnectionId: idx,
				Peer:         participant,
				Initiator:    idx == 0,
				Admin:        idx == 0,
			},
			DeviceType: msg.CallDeviceType_CallDeviceUnknown,
			MediaSettings: msg.CallMediaSettings{
				Audio:       true,
				ScreenShare: false,
				Video:       true,
			},
			Started: false,
		}
		callParticipantMap[participant.UserID] = idx
	}

	mediaState := c.getStreamState()
	c.mu.Lock()
	c.callInfo[callID] = &Info{
		acceptedParticipantIds: nil,
		acceptedParticipants:   nil,
		allConnected:           false,
		dialed:                 false,
		mediaSettings:          mediaState,
		participantMap:         callParticipantMap,
		participants:           callParticipants,
		requestParticipantIds:  nil,
		requests:               nil,
		iceServer:              nil,
		mu:                     &sync.RWMutex{},
	}
	c.mu.Unlock()
}

func (c *call) initParticipants(callID int64, participants []*msg.PhoneParticipant, bootstrap bool) {
	fn := func(callParticipants map[int32]*Participant, callParticipantMap map[int64]int32) (map[int32]*Participant, map[int64]int32) {
		for _, participant := range participants {
			callParticipants[participant.ConnectionId] = &Participant{
				PhoneParticipant: msg.PhoneParticipant{
					ConnectionId: participant.ConnectionId,
					Peer:         participant.Peer,
					Initiator:    participant.Initiator,
					Admin:        participant.Admin,
				},
				DeviceType: msg.CallDeviceType_CallDeviceUnknown,
				MediaSettings: msg.CallMediaSettings{
					Audio:       true,
					ScreenShare: false,
					Video:       true,
				},
				Started: false,
			}
			callParticipantMap[participant.Peer.UserID] = participant.ConnectionId
		}
		return callParticipants, callParticipantMap
	}

	if info, ok := c.callInfo[callID]; !ok {
		if bootstrap {
			callParticipants, callParticipantMap := fn(make(map[int32]*Participant), make(map[int64]int32))
			mediaState := c.getStreamState()
			c.mu.Lock()
			c.callInfo[callID] = &Info{
				acceptedParticipantIds: nil,
				acceptedParticipants:   nil,
				allConnected:           false,
				dialed:                 false,
				mediaSettings:          mediaState,
				participantMap:         callParticipantMap,
				participants:           callParticipants,
				requestParticipantIds:  nil,
				requests:               nil,
				iceServer:              nil,
				mu:                     &sync.RWMutex{},
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
	callParticipants := make(map[int32]*Participant)
	callParticipantMap := make(map[int64]int32)
	for _, participant := range sdpData.Participants {
		deviceType := msg.CallDeviceType_CallDeviceUnknown
		if in.UserID == participant.Peer.UserID {
			deviceType = sdpData.DeviceType
		}
		callParticipants[participant.ConnectionId] = &Participant{
			PhoneParticipant: msg.PhoneParticipant{
				ConnectionId: participant.ConnectionId,
				Peer:         participant.Peer,
				Initiator:    participant.Initiator,
				Admin:        participant.Admin,
			},
			DeviceType: deviceType,
			MediaSettings: msg.CallMediaSettings{
				Audio:       true,
				ScreenShare: false,
				Video:       true,
			},
			Started: false,
		}
		callParticipantMap[participant.Peer.UserID] = participant.ConnectionId
	}

	mediaState := c.getStreamState()
	c.mu.Lock()
	c.callInfo[in.CallID] = &Info{
		acceptedParticipantIds: nil,
		acceptedParticipants:   nil,
		allConnected:           false,
		dialed:                 false,
		mediaSettings:          mediaState,
		participantMap:         callParticipantMap,
		participants:           callParticipants,
		requestParticipantIds:  []int64{in.UserID},
		requests:               []*UpdatePhoneCall{in},
		iceServer:              nil,
		mu:                     &sync.RWMutex{},
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
			ConnectionId: p.ConnectionId,
			Peer:         p.Peer,
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
		if requestConnId == participant.ConnectionId {
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
		} else if shouldCall && currentUserConnId < participant.ConnectionId {
			wg.Add(1)
			go func(pConnId int32) {
				sdpRes, innerErr := c.initConnection(false, pConnId, nil)
				if innerErr == nil {
					mu.Lock()
					if participant, ok := callInfo.participants[pConnId]; ok {
						callResults = append(callResults, &msg.PhoneParticipantSDP{
							ConnectionId: participant.ConnectionId,
							Peer:         participant.Peer,
							SDP:          sdpRes.SDP,
							Type:         sdpRes.Type,
						})
					}
					mu.Unlock()
				}
				wg.Done()
			}(participant.ConnectionId)
		}
	}

	wg.Wait()

	for _, participantSDP := range callResults {
		fmt.Println(participantSDP)
		// retry here
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
	// TODO call delegate
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
	// TODO execute local command then call -> c.sendIceCandidate()

	// Client should listen to iceconnectionstatechange and send it to SDK
	// TODO call update handlers -> msg.CallUpdate_ConnectionStatusChanged
	// TODO check all connected
	// TODO checkDisconnection(connId, pc.iceConnectionState

	// Client should listen to icecandidateerror and send it to SDK
	// TODO checkDisconnection(connId, pc.iceConnectionState, true)

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
		connectTimout:   nil,
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
			// TODO Client should setRemoteDescription(sdp)
			// TODO Client should create answer
			// TODO Client should setLocalDescription and pass it to SDK
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
		// TODO Client should create offer
		// TODO Client should setLocalDescription and pass the offer to SDK
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
	return
}

func (c *call) sendIceCandidate(callID int64, candidate *msg.CallRTCIceCandidate, connId int32) (err error) {
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
			_ = c.sendIceCandidate(callID, ic, connId)
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

	// TODO -> msg.CallUpdate_LocalMediaSettingsUpdated

	inputUsers := c.getInputUsers(c.activeCallID)
	_, _ = c.apiSendUpdate(c.peer, c.activeCallID, inputUsers, msg.PhoneCallAction_PhoneCallMediaSettingsChanged, actionData, false)
	return
}

func (c *call) mediaSettingsInit(in MediaSettings) {
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

	// TOO Call -> msg.CallUpdate_MediaSettingsUpdated
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
		// TODO call -> msg.CallUpdate_AllConnected with 255 msec delay
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
		// TODO close connection with connID
		err = c.callback.CloseConnection(connId, false)
		if err != nil {
			return
		}

		conn.mu.Lock()
		conn.IceQueue = nil
		conn.Reconnecting = true
		conn.ReconnectingTry++
		if conn.ReconnectingTry <= ReconnectTry {
			conn.reconnectTimout = time.AfterFunc(time.Duration(ReconnectTimeout)*time.Millisecond, func() {
				if _, ok := c.peerConnections[connId]; ok {
					c.peerConnections[connId].Reconnecting = false
				}
			})
		}
		conn.mu.Unlock()

		// TODO call -> msg.CallUpdate_ConnectionStatusChanged with state "reconnecting"
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
		return &d.Peer.UserID
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
		return d.Peer
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
				inputUser = append(inputUser, participant.Peer)
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
		inputUser = append(inputUser, participant.Peer)
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

	if pc.connectTimout != nil {
		pc.connectTimout.Stop()
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
	info.participants[connId].Admin = admin
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
	if c.activeCallID != in.CallID {
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

		// TODO send -> msg.CallUpdate_CallRequested
	} else if c.shouldAccept(in) {
		streamState := c.getStreamState()
		_ = c.callAccept(c.activeCallID, streamState.Video)
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

	streamState := c.getStreamState()
	c.propagateMediaSettings(MediaSettingsIn{
		Audio:       &streamState.Audio,
		ScreenShare: &streamState.ScreenShare,
		Video:       &streamState.Video,
	})
	c.clearRetryInterval(connId, false)
	logs.Info("[webrtc] accept signal", zap.Int32("connId", connId))

	// TODO Client -> msg.CallUpdate_CallAccepted
}

func (c *call) callDiscarded(in *UpdatePhoneCall) {
	connId, _, valid := c.getConnId(in.CallID, in.UserID)
	if !valid {
		return
	}

	c.clearRetryInterval(connId, false)

	data := in.Data.(*msg.PhoneActionDiscarded)
	if in.PeerType == int32(msg.PeerType_PeerUser) || data.Terminate {
		// TODO Client -> msg.CallUpdate_CallRejected
	} else {
		if c.removeParticipant(in.UserID, &in.CallID) {
			// TODO Client -> msg.CallUpdate_CallRejected
		} else {
			c.checkAllConnected()
			// TODO Client -> msg.CallUpdate_ParticipantLeft
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
	// TODO Call -> msg.CallUpdate_CallAck
}

func (c *call) participantAdded(in *UpdatePhoneCall) {
	if c.activeCallID != in.CallID {
		return
	}

	data := in.Data.(*msg.PhoneActionParticipantAdded)
	c.initParticipants(c.activeCallID, data.Participants, false)
	isNew := true
	for _, participant := range data.Participants {
		if participant.Peer.UserID == c.userID {
			isNew = false
			break
		}
	}

	if isNew {
		// TODO CALL -> msg.CallUpdate_ParticipantJoined
	}
}

func (c *call) participantRemoved(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionParticipantRemoved)

	for _, userId := range data.UserIDs {
		if userId == c.userID {
			// TODO Call -> msg.CallUpdate_CallCancelled
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

	// TODO Call -> msg.CallUpdate_ParticipantRemoved
	if isCurrentRemoved {
		// TODO Call -> msg.CallUpdate_CallRejected
	}

	c.checkAllConnected()
}

func (c *call) adminUpdated(in *UpdatePhoneCall) {
	if c.activeCallID != in.CallID {
		return
	}

	data := in.Data.(*msg.PhoneActionAdminUpdated)
	c.updateAdmin(data.UserID, data.Admin)
	// TODO Call -> msg.CallUpdate_ParticipantAdminUpdated
}

func (c *call) joinRequested(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionJoinRequested)
	for _, userId := range data.UserIDs {
		if userId == c.userID {
			inputPeer := &msg.InputPeer{
				ID:         in.PeerID,
				Type:       msg.PeerType(in.PeerType),
				AccessHash: 0,
			}
			fmt.Println(inputPeer)
			// TODO Call -> msg.CallUpdate_CallJoinRequested
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
		// TODO call -> msg.CallUpdate_CallCancelled
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
