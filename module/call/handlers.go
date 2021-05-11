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

func (c *call) ToggleVide(enable bool) {

}

func (c *call) ToggleAudio(enable bool) {

}

func (c *call) TryReconnect(connId int32) {

}

func (c *call) CallStart(peer *msg.InputPeer, participants []*msg.InputUser, callID int64) {
	c.peer = peer
	initRes, err := c.apiInit(peer, callID)
	if err != nil {
		logs.Warn("Init", zap.Error(err))
		return
	}
	c.iceServer = initRes.IceServers
	if callID != 0 {
	} else {
		c.activeCallID = 0
		c.initCallParticipants(TempCallID, participants)
	}
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

func (c *call) initParticipant(callID int64, participants []*msg.PhoneParticipant, bootstrap bool) {
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
	currUserConnId, callInfo := c.getConnId(callID, c.userID)
	if callInfo == nil {
		err = ErrInvalidCallId
		return
	}

	currentUserConnId := *currUserConnId
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
		reConnId, _ := c.getConnId(callID, request.UserID)
		if reConnId != nil && callInfo.dialed {
			requestConnId = *reConnId
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
			var clientAnswerSQP []byte
			clientAnswerSQP, err = c.callback.GetAnswerSDP(connId)
			if err != nil {
				return
			}

			sdpAnswer = &msg.PhoneActionSDPAnswer{}
			err = sdpAnswer.Unmarshal(clientAnswerSQP)
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

	cid, info := c.getConnId(c.activeCallID, c.userID)
	if cid == nil {
		return
	}

	connId := *cid
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
		err = c.callback.CloseConnection(connId)
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
		currConnId, _ := c.getConnId(c.activeCallID, c.userID)
		if currConnId == nil {
			return
		}

		currentConnId := *currConnId
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
	if c, ok := c.callInfo[callID]; ok {
		return c
	} else {
		return nil
	}
}

func (c *call) getConnId(callID, userID int64) (*int32, *Info) {
	info := c.getCallInfo(callID)
	if info == nil {
		return nil, nil
	}

	connId := int32(info.participantMap[userID])
	return &connId, info
}

func (c *call) getUserIDbyCallID(callID int64, connID int32) *int64 {
	info := c.getCallInfo(callID)
	if info == nil {
		return nil
	}

	if d, ok := info.participants[connID]; ok {
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

	if d, ok := info.participants[connID]; ok {
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
	for _, userID := range userIDs {
		if connId, ok := info.participantMap[userID]; ok {
			if participant, ok2 := info.participants[connId]; ok2 {
				inputUser = append(inputUser, participant.Peer)
			}
		}
	}
	return
}

func (c *call) getInputUsers(callID int64) (inputUser []*msg.InputUser) {
	info := c.getCallInfo(callID)
	if info == nil {
		return nil
	}

	inputUser = make([]*msg.InputUser, 0, len(info.participants))
	for _, participant := range info.participants {
		inputUser = append(inputUser, participant.Peer)
	}
	return
}

func (c *call) swapTempInfo(callID int64) {
	info := c.getCallInfo(TempCallID)
	if info == nil {
		return
	}
	c.callInfo[callID] = info
	delete(c.callInfo, TempCallID)
}

func (c *call) setCallInfoDialed(callID int64) {
	if _, ok := c.callInfo[callID]; ok {
		c.callInfo[callID].dialed = true
	}
}

func (c *call) callBusy(in *UpdatePhoneCall) {
	inputPeer := c.getInputUserFromUpdate(in)

	_, _ = c.apiReject(inputPeer, in.CallID, msg.DiscardReason_DiscardReasonBusy, 0)
	return
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

	c.activeCallID = callID
	c.iceServer = initRes.IceServers

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
	data := in.Data.(*msg.PhoneActionAccepted)
	fmt.Println(data)
}

func (c *call) callDiscarded(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionDiscarded)
	fmt.Println(data)
}

func (c *call) iceExchange(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionIceExchange)
	fmt.Println(data)
}

func (c *call) mediaSettingsUpdated(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionMediaSettingsUpdated)
	cid, info := c.getConnId(in.CallID, in.UserID)
	if cid == nil {
		return
	}

	connId := *cid
	info.mu.Lock()
	info.participants[connId].MediaSettings.Audio = data.Audio
	info.participants[connId].MediaSettings.Video = data.Video
	info.participants[connId].MediaSettings.ScreenShare = data.ScreenShare
	info.mu.Unlock()

	// TOO Call -> msg.CallUpdate_MediaSettingsUpdated
}

func (c *call) sdpOfferUpdated(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionSDPOffer)
	fmt.Println(data)
}

func (c *call) sdpAnswerUpdated(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionSDPAnswer)
	fmt.Println(data)
}

func (c *call) callAcknowledged(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionAck)
	fmt.Println(data)
}

func (c *call) participantAdded(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionParticipantAdded)
	fmt.Println(data)
}

func (c *call) participantRemoved(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionParticipantRemoved)
	fmt.Println(data)
}

func (c *call) adminUpdated(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionAdminUpdated)
	fmt.Println(data)
}

func (c *call) joinRequested(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionJoinRequested)
	fmt.Println(data)
}

func (c *call) screenShareUpdated(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionScreenShare)
	fmt.Println(data)
}

func (c *call) callPicked(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionPicked)
	fmt.Println(data)
}

func (c *call) callRestarted(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionRestarted)
	fmt.Println(data)
}
