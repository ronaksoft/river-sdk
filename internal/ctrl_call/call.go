/*
   Creation Time: 2021 - April - 04
   Created by:  (hamidrezakk)
   Maintainers:
      1.  HamidrezaKK (hamidrezakks@gmail.com)
   Auditor: HamidrezaKK
   Copyright Ronak Software Group 2021
*/

package callCtrl

import (
	"fmt"
	"git.ronaksoft.com/river/msg/go/msg"
	"git.ronaksoft.com/river/sdk/internal/domain"
	"git.ronaksoft.com/river/sdk/internal/logs"
	"go.uber.org/zap"
	"sync"
	"time"
)

const (
	RetryInterval    = 10000
	RetryLimit       = 6
	ReconnectTry     = 3
	ReconnectTimeout = 15000

	TempCallID = int64(-27001)
)

type CallController interface {
	ParseUpdate(update *msg.UpdateEnvelope)
}

func NewCallController() CallController {
	callAPI := NewCallAPI()
	return &callController{
		peerConnections: nil,
		peer:            nil,
		activeCallID:    0,
		callInfo:        make(map[int64]*CallInfo, 0),
		iceServer:       nil,
		userID:          0,
		callAPI:         callAPI,
	}
}

type callController struct {
	peerConnections map[int32]*msg.CallConnection
	peer            *msg.InputPeer
	activeCallID    int64
	callInfo        map[int64]*CallInfo
	iceServer       []*msg.CallRTCIceServer
	userID          int64

	callAPI CallAPI
}

func (c *callController) ParseUpdate(update *msg.UpdateEnvelope) {
	go func() {
		if update.Constructor != msg.C_UpdatePhoneCall {
			return
		}
		x := &msg.UpdatePhoneCall{}
		err := x.Unmarshal(update.Update)
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
}

func (c *callController) ToggleVide(enable bool) {

}

func (c *callController) ToggleAudio(enable bool) {

}

func (c *callController) TryReconnect(connId int32) {

}

func (c *callController) CallStart(peer *msg.InputPeer, participants []*msg.InputUser, callID int64) {
	c.peer = peer
	initRes, err := c.callAPI.Init(peer, callID)
	if err != nil {
		logs.Warn("Init", zap.Error(err))
		return
	}
	c.iceServer = c.transformIceServers(initRes.IceServers)
	if callID != 0 {
	} else {
		c.activeCallID = 0
		c.initCallParticipants(TempCallID, participants)

	}
}

func (c *callController) getStreamState() *msg.CallMediaSettings {
	return &msg.CallMediaSettings{
		Audio:       true,
		ScreenShare: false,
		Video:       true,
	}
}

func (c *callController) initCallParticipants(callID int64, participants []*msg.InputUser) {
	participants = append([]*msg.InputUser{{
		UserID:     c.userID,
		AccessHash: 0,
	}}, participants...)
	callParticipants := make(map[int32]*CallParticipant)
	callParticipantMap := make(map[int64]int32)
	for i, participant := range participants {
		idx := int32(i)
		callParticipants[idx] = &CallParticipant{
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
	c.callInfo[callID] = &CallInfo{
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

func (c *callController) initConnections(peer *msg.InputPeer, callID int64, initiator bool, request *UpdatePhoneCall) (res *msg.PhoneCall, err error) {
	currUserConnId, callInfo := c.getConnId(callID, c.userID)
	if callInfo == nil {
		err = ErrInvalidCallId
		return
	}

	currentUserConnId := *currUserConnId
	callWaitGroup := &sync.WaitGroup{}
	callResults := []*msg.PhoneParticipantSDP{}

	acceptWaitGroup := &sync.WaitGroup{}
	acceptResults := []*msg.PhoneCall{}
	mu := &sync.RWMutex{}

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

		res, innerErr = c.callAPI.Accept(peer, callID, []*msg.PhoneParticipantSDP{phoneParticipant})
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
			go func() {
				acceptWaitGroup.Add(1)
				phoneCall, innerRes := initAnswerConnection(requestConnId)
				if innerRes == nil {
					mu.Lock()
					acceptResults = append(acceptResults, phoneCall)
					mu.Unlock()
				}
				acceptWaitGroup.Done()
			}()
		} else if shouldCall && currentUserConnId < participant.ConnectionId {
			go func(pConnId int32) {
				callWaitGroup.Add(1)
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
				callWaitGroup.Done()
			}(participant.ConnectionId)
		}
	}

	callWaitGroup.Wait()
	acceptWaitGroup.Wait()

	for _, participantSDP := range callResults {
		fmt.Println(participantSDP)
		// retry here
	}
	_, _ = c.callUser(peer, initiator, callResults, c.activeCallID)
	return
}

func (c *callController) initConnection(remote bool, connId int32, sdp *msg.PhoneActionSDPOffer) (sdpAnswer *msg.PhoneActionSDPAnswer, err error) {
	logs.Debug("[webrtc] init connection", zap.Int32("connId", connId))
	// Client should check local stream
	// otherwise panic

	// Use MediaSteam to mix video and audio track
	// You can use main MediaStream if no shared screen media is present

	iceServer := c.iceServer
	if pc, ok := c.peerConnections[connId]; ok {
		iceServer = pc.IceServers
	}

	println(iceServer)

	// Client should initiate RTCPeerConnection with given server config
	// TODO call delegate

	rtcConnId := int64(0)

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
		ReconnectingTry:     false,
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
	// clear reconnect timeout

	if remote {
		if sdp != nil {
			// TODO Client should setRemoteDescription(sdp)
			// TODO Client should create answer
			// TODO Client should setLocalDescription and pass it to SDK
		} else {
			err = ErrNoSDP
			return
		}
	} else {
		// TODO Client should create offer
		// TODO Client should setLocalDescription and pass the offer to SDK
	}
	return
}

func (c *callController) callUser(peer *msg.InputPeer, initiator bool, phoneParticipants []*msg.PhoneParticipantSDP, callID int64) (res *msg.PhoneCall, err error) {
	randomID := domain.RandomInt64(0)
	res, err = c.callAPI.Request(peer, randomID, initiator, phoneParticipants, callID, false)
	return
}

func (c *callController) sendIceCandidate(callID int64, candidate *msg.CallRTCIceCandidate, connId int32) (err error) {
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

	_, err = c.callAPI.SendUpdate(c.peer, callID, []*msg.InputUser{inputUser}, msg.PhoneCallAction_PhoneCallIceExchange, actionData, false)
	return
}

func (c *callController) checkAllConnected() {
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

func (c *callController) checkDisconnection(connId int32, state string, isIceError bool) {
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
		c.peerConnections[connId].IceQueue = nil
		c.peerConnections[connId].Reconnecting = true
		//c.peerConnections[connId].ReconnectingTry = true
		if conn.ReconnectingTry <= ReconnectTry {
			time.AfterFunc(time.Duration(ReconnectTimeout)*time.Millisecond, func() {
				if _, ok := c.peerConnections[connId]; ok {
					c.peerConnections[connId].Reconnecting = false
				}
			})
		}
		// TODO call -> msg.CallUpdate_ConnectionStatusChanged with state "reconnecting"
		initRes, err := c.callAPI.Init(c.peer, c.activeCallID)
		if err != nil {
			return
		}
		_, hasConn = c.peerConnections[connId]
		if !hasConn {
			return
		}
		c.peerConnections[connId].IceServers = c.transformIceServers(initRes.IceServers)
		currConnId, _ := c.getConnId(c.activeCallID, c.userID)
		if currConnId == nil {
			return
		}
		currentConnId := *currConnId
		if currentConnId < connId {
			_ = c.callSendRestart(connId, true)
		} else {
			c.initConnection(true, connId, nil)
		}
	}
}

func (c *callController) callSendRestart(connId int32, sender bool) (err error) {
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

	_, err = c.callAPI.SendUpdate(c.peer, c.activeCallID, []*msg.InputUser{inputUser}, msg.PhoneCallAction_PhoneCallRestarted, actionData, true)
	return
}

func (c *callController) getCallInfo(callID int64) *CallInfo {
	if c, ok := c.callInfo[callID]; ok {
		return c
	} else {
		return nil
	}
}

func (c *callController) getConnId(callID, userID int64) (*int32, *CallInfo) {
	info := c.getCallInfo(callID)
	if info == nil {
		return nil, nil
	}

	connId := int32(info.participantMap[userID])
	return &connId, info
}

func (c *callController) getUserIDbyCallID(callID int64, connID int32) *int64 {
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

func (c *callController) getInputUserByConnId(callID int64, connID int32) (inputUser *msg.InputUser) {
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

func (c *callController) getInputUserByUserIDs(callID int64, userIDs []int64) (inputUser []*msg.InputUser) {
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

func (c *callController) swapTempInfo(callID int64) {
	info := c.getCallInfo(TempCallID)
	if info == nil {
		return
	}
	c.callInfo[callID] = info
	delete(c.callInfo, TempCallID)
}

func (c *callController) setCallInfoDialed(callID int64) {
	if _, ok := c.callInfo[callID]; ok {
		c.callInfo[callID].dialed = true
	}
}

func (c *callController) transformIceServers(in []*msg.IceServer) (out []*msg.CallRTCIceServer) {
	out = make([]*msg.CallRTCIceServer, len(in))
	for idx, item := range in {
		out[idx] = &msg.CallRTCIceServer{
			Credential:     item.Credential,
			CredentialType: "",
			Urls:           item.Urls,
			Username:       item.Username,
		}
	}
	return
}

func (c *callController) callRequested(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionRequested)
	fmt.Println(data)
}

func (c *callController) callAccepted(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionAccepted)
	fmt.Println(data)
}

func (c *callController) callDiscarded(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionDiscarded)
	fmt.Println(data)
}

func (c *callController) iceExchange(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionIceExchange)
	fmt.Println(data)
}

func (c *callController) mediaSettingsUpdate(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionMediaSettingsUpdated)
	fmt.Println(data)
}

func (c *callController) sdpOfferUpdated(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionSDPOffer)
	fmt.Println(data)
}

func (c *callController) sdpAnswerUpdated(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionSDPAnswer)
	fmt.Println(data)
}

func (c *callController) callAcknowledged(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionAck)
	fmt.Println(data)
}

func (c *callController) participantAdded(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionParticipantAdded)
	fmt.Println(data)
}

func (c *callController) participantRemoved(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionParticipantRemoved)
	fmt.Println(data)
}

func (c *callController) adminUpdated(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionAdminUpdated)
	fmt.Println(data)
}

func (c *callController) joinRequested(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionJoinRequested)
	fmt.Println(data)
}

func (c *callController) screenShareUpdated(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionScreenShare)
	fmt.Println(data)
}

func (c *callController) callPicked(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionPicked)
	fmt.Println(data)
}

func (c *callController) callRestarted(in *UpdatePhoneCall) {
	data := in.Data.(*msg.PhoneActionRestarted)
	fmt.Println(data)
}
