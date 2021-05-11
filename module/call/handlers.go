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
			c.mediaSettingsUpdate(update)
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

func (c *call) getStreamState() *msg.CallMediaSettings {
	return &msg.CallMediaSettings{
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
	}
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

	conn := &msg.CallConnection{
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
	}

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

		c.peerConnections[connId].IceQueue = nil
		c.peerConnections[connId].Reconnecting = true
		c.peerConnections[connId].ReconnectingTry++
		if conn.ReconnectingTry <= ReconnectTry {
			time.AfterFunc(time.Duration(ReconnectTimeout)*time.Millisecond, func() {
				if _, ok := c.peerConnections[connId]; ok {
					c.peerConnections[connId].Reconnecting = false
				}
			})
		}

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

		c.peerConnections[connId].IceServers = initRes.IceServers
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

func (c *call) callRequested(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionRequested)
	fmt.Println(data)
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

func (c *call) mediaSettingsUpdate(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionMediaSettingsUpdated)
	fmt.Println(data)
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
