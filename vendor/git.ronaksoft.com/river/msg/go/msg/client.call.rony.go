// Code generated by Rony's protoc plugin; DO NOT EDIT.

package msg

import (
	edge "github.com/ronaksoft/rony/edge"
	registry "github.com/ronaksoft/rony/registry"
	proto "google.golang.org/protobuf/proto"
	sync "sync"
)

const C_CallMediaSettings int64 = 1147111688

type poolCallMediaSettings struct {
	pool sync.Pool
}

func (p *poolCallMediaSettings) Get() *CallMediaSettings {
	x, ok := p.pool.Get().(*CallMediaSettings)
	if !ok {
		x = &CallMediaSettings{}
	}
	return x
}

func (p *poolCallMediaSettings) Put(x *CallMediaSettings) {
	if x == nil {
		return
	}
	x.Audio = false
	x.ScreenShare = false
	x.Video = false
	p.pool.Put(x)
}

var PoolCallMediaSettings = poolCallMediaSettings{}

func (x *CallMediaSettings) DeepCopy(z *CallMediaSettings) {
	z.Audio = x.Audio
	z.ScreenShare = x.ScreenShare
	z.Video = x.Video
}

func (x *CallMediaSettings) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CallMediaSettings) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *CallMediaSettings) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CallMediaSettings, x)
}

const C_CallParticipant int64 = 2652007354

type poolCallParticipant struct {
	pool sync.Pool
}

func (p *poolCallParticipant) Get() *CallParticipant {
	x, ok := p.pool.Get().(*CallParticipant)
	if !ok {
		x = &CallParticipant{}
	}
	return x
}

func (p *poolCallParticipant) Put(x *CallParticipant) {
	if x == nil {
		return
	}
	x.ConnectionID = 0
	PoolInputUser.Put(x.Peer)
	x.Peer = nil
	x.Initiator = false
	x.Admin = false
	x.DeviceType = 0
	PoolCallMediaSettings.Put(x.MediaSettings)
	x.MediaSettings = nil
	x.Started = false
	p.pool.Put(x)
}

var PoolCallParticipant = poolCallParticipant{}

func (x *CallParticipant) DeepCopy(z *CallParticipant) {
	z.ConnectionID = x.ConnectionID
	if x.Peer != nil {
		if z.Peer == nil {
			z.Peer = PoolInputUser.Get()
		}
		x.Peer.DeepCopy(z.Peer)
	} else {
		z.Peer = nil
	}
	z.Initiator = x.Initiator
	z.Admin = x.Admin
	z.DeviceType = x.DeviceType
	if x.MediaSettings != nil {
		if z.MediaSettings == nil {
			z.MediaSettings = PoolCallMediaSettings.Get()
		}
		x.MediaSettings.DeepCopy(z.MediaSettings)
	} else {
		z.MediaSettings = nil
	}
	z.Started = x.Started
}

func (x *CallParticipant) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CallParticipant) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *CallParticipant) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CallParticipant, x)
}

const C_CallRTCIceCandidate int64 = 2748774954

type poolCallRTCIceCandidate struct {
	pool sync.Pool
}

func (p *poolCallRTCIceCandidate) Get() *CallRTCIceCandidate {
	x, ok := p.pool.Get().(*CallRTCIceCandidate)
	if !ok {
		x = &CallRTCIceCandidate{}
	}
	return x
}

func (p *poolCallRTCIceCandidate) Put(x *CallRTCIceCandidate) {
	if x == nil {
		return
	}
	x.Candidate = ""
	x.SdpMLineIndex = 0
	x.SdpMid = ""
	x.UsernameFragment = ""
	p.pool.Put(x)
}

var PoolCallRTCIceCandidate = poolCallRTCIceCandidate{}

func (x *CallRTCIceCandidate) DeepCopy(z *CallRTCIceCandidate) {
	z.Candidate = x.Candidate
	z.SdpMLineIndex = x.SdpMLineIndex
	z.SdpMid = x.SdpMid
	z.UsernameFragment = x.UsernameFragment
}

func (x *CallRTCIceCandidate) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CallRTCIceCandidate) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *CallRTCIceCandidate) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CallRTCIceCandidate, x)
}

const C_CallRTCIceServer int64 = 1457208388

type poolCallRTCIceServer struct {
	pool sync.Pool
}

func (p *poolCallRTCIceServer) Get() *CallRTCIceServer {
	x, ok := p.pool.Get().(*CallRTCIceServer)
	if !ok {
		x = &CallRTCIceServer{}
	}
	return x
}

func (p *poolCallRTCIceServer) Put(x *CallRTCIceServer) {
	if x == nil {
		return
	}
	x.Credential = ""
	x.CredentialType = ""
	x.Urls = x.Urls[:0]
	x.Username = ""
	p.pool.Put(x)
}

var PoolCallRTCIceServer = poolCallRTCIceServer{}

func (x *CallRTCIceServer) DeepCopy(z *CallRTCIceServer) {
	z.Credential = x.Credential
	z.CredentialType = x.CredentialType
	z.Urls = append(z.Urls[:0], x.Urls...)
	z.Username = x.Username
}

func (x *CallRTCIceServer) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CallRTCIceServer) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *CallRTCIceServer) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CallRTCIceServer, x)
}

const C_CallConnection int64 = 3450901888

type poolCallConnection struct {
	pool sync.Pool
}

func (p *poolCallConnection) Get() *CallConnection {
	x, ok := p.pool.Get().(*CallConnection)
	if !ok {
		x = &CallConnection{}
	}
	return x
}

func (p *poolCallConnection) Put(x *CallConnection) {
	if x == nil {
		return
	}
	x.Accepted = false
	x.RTCPeerConnectionID = 0
	x.iceConnectionState = ""
	for _, z := range x.IceQueue {
		PoolCallRTCIceCandidate.Put(z)
	}
	x.IceQueue = x.IceQueue[:0]
	for _, z := range x.IceServers {
		PoolCallRTCIceServer.Put(z)
	}
	x.IceServers = x.IceServers[:0]
	x.Init = false
	x.Reconnecting = false
	x.ReconnectingTry = false
	x.ScreenShareStreamID = 0
	x.StreamID = 0
	x.IntervalID = 0
	x.Try = 0
	p.pool.Put(x)
}

var PoolCallConnection = poolCallConnection{}

func (x *CallConnection) DeepCopy(z *CallConnection) {
	z.Accepted = x.Accepted
	z.RTCPeerConnectionID = x.RTCPeerConnectionID
	z.iceConnectionState = x.iceConnectionState
	for idx := range x.IceQueue {
		if x.IceQueue[idx] != nil {
			xx := PoolCallRTCIceCandidate.Get()
			x.IceQueue[idx].DeepCopy(xx)
			z.IceQueue = append(z.IceQueue, xx)
		}
	}
	for idx := range x.IceServers {
		if x.IceServers[idx] != nil {
			xx := PoolCallRTCIceServer.Get()
			x.IceServers[idx].DeepCopy(xx)
			z.IceServers = append(z.IceServers, xx)
		}
	}
	z.Init = x.Init
	z.Reconnecting = x.Reconnecting
	z.ReconnectingTry = x.ReconnectingTry
	z.ScreenShareStreamID = x.ScreenShareStreamID
	z.StreamID = x.StreamID
	z.IntervalID = x.IntervalID
	z.Try = x.Try
}

func (x *CallConnection) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CallConnection) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *CallConnection) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CallConnection, x)
}

const C_CallUpdateCallRequested int64 = 2556114354

type poolCallUpdateCallRequested struct {
	pool sync.Pool
}

func (p *poolCallUpdateCallRequested) Get() *CallUpdateCallRequested {
	x, ok := p.pool.Get().(*CallUpdateCallRequested)
	if !ok {
		x = &CallUpdateCallRequested{}
	}
	return x
}

func (p *poolCallUpdateCallRequested) Put(x *CallUpdateCallRequested) {
	if x == nil {
		return
	}
	x.ID = 0
	x.Type = 0
	x.CallID = 0
	p.pool.Put(x)
}

var PoolCallUpdateCallRequested = poolCallUpdateCallRequested{}

func (x *CallUpdateCallRequested) DeepCopy(z *CallUpdateCallRequested) {
	z.ID = x.ID
	z.Type = x.Type
	z.CallID = x.CallID
}

func (x *CallUpdateCallRequested) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CallUpdateCallRequested) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *CallUpdateCallRequested) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CallUpdateCallRequested, x)
}

const C_CallUpdateCallAccepted int64 = 2134109006

type poolCallUpdateCallAccepted struct {
	pool sync.Pool
}

func (p *poolCallUpdateCallAccepted) Get() *CallUpdateCallAccepted {
	x, ok := p.pool.Get().(*CallUpdateCallAccepted)
	if !ok {
		x = &CallUpdateCallAccepted{}
	}
	return x
}

func (p *poolCallUpdateCallAccepted) Put(x *CallUpdateCallAccepted) {
	if x == nil {
		return
	}
	x.ConnectionID = 0
	p.pool.Put(x)
}

var PoolCallUpdateCallAccepted = poolCallUpdateCallAccepted{}

func (x *CallUpdateCallAccepted) DeepCopy(z *CallUpdateCallAccepted) {
	z.ConnectionID = x.ConnectionID
}

func (x *CallUpdateCallAccepted) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CallUpdateCallAccepted) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *CallUpdateCallAccepted) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CallUpdateCallAccepted, x)
}

const C_CallUpdateStreamUpdated int64 = 3496218809

type poolCallUpdateStreamUpdated struct {
	pool sync.Pool
}

func (p *poolCallUpdateStreamUpdated) Get() *CallUpdateStreamUpdated {
	x, ok := p.pool.Get().(*CallUpdateStreamUpdated)
	if !ok {
		x = &CallUpdateStreamUpdated{}
	}
	return x
}

func (p *poolCallUpdateStreamUpdated) Put(x *CallUpdateStreamUpdated) {
	if x == nil {
		return
	}
	x.ConnectionID = 0
	x.StreamID = ""
	p.pool.Put(x)
}

var PoolCallUpdateStreamUpdated = poolCallUpdateStreamUpdated{}

func (x *CallUpdateStreamUpdated) DeepCopy(z *CallUpdateStreamUpdated) {
	z.ConnectionID = x.ConnectionID
	z.StreamID = x.StreamID
}

func (x *CallUpdateStreamUpdated) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CallUpdateStreamUpdated) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *CallUpdateStreamUpdated) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CallUpdateStreamUpdated, x)
}

const C_CallUpdateCallRejected int64 = 2339651845

type poolCallUpdateCallRejected struct {
	pool sync.Pool
}

func (p *poolCallUpdateCallRejected) Get() *CallUpdateCallRejected {
	x, ok := p.pool.Get().(*CallUpdateCallRejected)
	if !ok {
		x = &CallUpdateCallRejected{}
	}
	return x
}

func (p *poolCallUpdateCallRejected) Put(x *CallUpdateCallRejected) {
	if x == nil {
		return
	}
	x.CallID = 0
	x.Reason = 0
	p.pool.Put(x)
}

var PoolCallUpdateCallRejected = poolCallUpdateCallRejected{}

func (x *CallUpdateCallRejected) DeepCopy(z *CallUpdateCallRejected) {
	z.CallID = x.CallID
	z.Reason = x.Reason
}

func (x *CallUpdateCallRejected) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CallUpdateCallRejected) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *CallUpdateCallRejected) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CallUpdateCallRejected, x)
}

const C_CallUpdateMediaSettingsUpdated int64 = 3922101985

type poolCallUpdateMediaSettingsUpdated struct {
	pool sync.Pool
}

func (p *poolCallUpdateMediaSettingsUpdated) Get() *CallUpdateMediaSettingsUpdated {
	x, ok := p.pool.Get().(*CallUpdateMediaSettingsUpdated)
	if !ok {
		x = &CallUpdateMediaSettingsUpdated{}
	}
	return x
}

func (p *poolCallUpdateMediaSettingsUpdated) Put(x *CallUpdateMediaSettingsUpdated) {
	if x == nil {
		return
	}
	x.ConnectionID = 0
	PoolCallMediaSettings.Put(x.MediaSettings)
	x.MediaSettings = nil
	p.pool.Put(x)
}

var PoolCallUpdateMediaSettingsUpdated = poolCallUpdateMediaSettingsUpdated{}

func (x *CallUpdateMediaSettingsUpdated) DeepCopy(z *CallUpdateMediaSettingsUpdated) {
	z.ConnectionID = x.ConnectionID
	if x.MediaSettings != nil {
		if z.MediaSettings == nil {
			z.MediaSettings = PoolCallMediaSettings.Get()
		}
		x.MediaSettings.DeepCopy(z.MediaSettings)
	} else {
		z.MediaSettings = nil
	}
}

func (x *CallUpdateMediaSettingsUpdated) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CallUpdateMediaSettingsUpdated) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *CallUpdateMediaSettingsUpdated) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CallUpdateMediaSettingsUpdated, x)
}

const C_CallUpdateLocalStreamUpdated int64 = 1043624904

type poolCallUpdateLocalStreamUpdated struct {
	pool sync.Pool
}

func (p *poolCallUpdateLocalStreamUpdated) Get() *CallUpdateLocalStreamUpdated {
	x, ok := p.pool.Get().(*CallUpdateLocalStreamUpdated)
	if !ok {
		x = &CallUpdateLocalStreamUpdated{}
	}
	return x
}

func (p *poolCallUpdateLocalStreamUpdated) Put(x *CallUpdateLocalStreamUpdated) {
	if x == nil {
		return
	}
	x.StreamID = ""
	p.pool.Put(x)
}

var PoolCallUpdateLocalStreamUpdated = poolCallUpdateLocalStreamUpdated{}

func (x *CallUpdateLocalStreamUpdated) DeepCopy(z *CallUpdateLocalStreamUpdated) {
	z.StreamID = x.StreamID
}

func (x *CallUpdateLocalStreamUpdated) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CallUpdateLocalStreamUpdated) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *CallUpdateLocalStreamUpdated) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CallUpdateLocalStreamUpdated, x)
}

const C_CallUpdateCallTimeout int64 = 420503198

type poolCallUpdateCallTimeout struct {
	pool sync.Pool
}

func (p *poolCallUpdateCallTimeout) Get() *CallUpdateCallTimeout {
	x, ok := p.pool.Get().(*CallUpdateCallTimeout)
	if !ok {
		x = &CallUpdateCallTimeout{}
	}
	return x
}

func (p *poolCallUpdateCallTimeout) Put(x *CallUpdateCallTimeout) {
	if x == nil {
		return
	}
	p.pool.Put(x)
}

var PoolCallUpdateCallTimeout = poolCallUpdateCallTimeout{}

func (x *CallUpdateCallTimeout) DeepCopy(z *CallUpdateCallTimeout) {
}

func (x *CallUpdateCallTimeout) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CallUpdateCallTimeout) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *CallUpdateCallTimeout) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CallUpdateCallTimeout, x)
}

const C_CallUpdateCallAck int64 = 1424725011

type poolCallUpdateCallAck struct {
	pool sync.Pool
}

func (p *poolCallUpdateCallAck) Get() *CallUpdateCallAck {
	x, ok := p.pool.Get().(*CallUpdateCallAck)
	if !ok {
		x = &CallUpdateCallAck{}
	}
	return x
}

func (p *poolCallUpdateCallAck) Put(x *CallUpdateCallAck) {
	if x == nil {
		return
	}
	x.ConnectionID = 0
	p.pool.Put(x)
}

var PoolCallUpdateCallAck = poolCallUpdateCallAck{}

func (x *CallUpdateCallAck) DeepCopy(z *CallUpdateCallAck) {
	z.ConnectionID = x.ConnectionID
}

func (x *CallUpdateCallAck) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CallUpdateCallAck) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *CallUpdateCallAck) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CallUpdateCallAck, x)
}

const C_CallUpdateParticipantJoined int64 = 1005455524

type poolCallUpdateParticipantJoined struct {
	pool sync.Pool
}

func (p *poolCallUpdateParticipantJoined) Get() *CallUpdateParticipantJoined {
	x, ok := p.pool.Get().(*CallUpdateParticipantJoined)
	if !ok {
		x = &CallUpdateParticipantJoined{}
	}
	return x
}

func (p *poolCallUpdateParticipantJoined) Put(x *CallUpdateParticipantJoined) {
	if x == nil {
		return
	}
	x.UserIDs = x.UserIDs[:0]
	p.pool.Put(x)
}

var PoolCallUpdateParticipantJoined = poolCallUpdateParticipantJoined{}

func (x *CallUpdateParticipantJoined) DeepCopy(z *CallUpdateParticipantJoined) {
	z.UserIDs = append(z.UserIDs[:0], x.UserIDs...)
}

func (x *CallUpdateParticipantJoined) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CallUpdateParticipantJoined) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *CallUpdateParticipantJoined) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CallUpdateParticipantJoined, x)
}

const C_CallUpdateParticipantLeft int64 = 2062471712

type poolCallUpdateParticipantLeft struct {
	pool sync.Pool
}

func (p *poolCallUpdateParticipantLeft) Get() *CallUpdateParticipantLeft {
	x, ok := p.pool.Get().(*CallUpdateParticipantLeft)
	if !ok {
		x = &CallUpdateParticipantLeft{}
	}
	return x
}

func (p *poolCallUpdateParticipantLeft) Put(x *CallUpdateParticipantLeft) {
	if x == nil {
		return
	}
	x.UserID = 0
	p.pool.Put(x)
}

var PoolCallUpdateParticipantLeft = poolCallUpdateParticipantLeft{}

func (x *CallUpdateParticipantLeft) DeepCopy(z *CallUpdateParticipantLeft) {
	z.UserID = x.UserID
}

func (x *CallUpdateParticipantLeft) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CallUpdateParticipantLeft) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *CallUpdateParticipantLeft) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CallUpdateParticipantLeft, x)
}

const C_CallUpdateParticipantRemoved int64 = 4138786615

type poolCallUpdateParticipantRemoved struct {
	pool sync.Pool
}

func (p *poolCallUpdateParticipantRemoved) Get() *CallUpdateParticipantRemoved {
	x, ok := p.pool.Get().(*CallUpdateParticipantRemoved)
	if !ok {
		x = &CallUpdateParticipantRemoved{}
	}
	return x
}

func (p *poolCallUpdateParticipantRemoved) Put(x *CallUpdateParticipantRemoved) {
	if x == nil {
		return
	}
	x.UserIDs = x.UserIDs[:0]
	x.Timeout = false
	p.pool.Put(x)
}

var PoolCallUpdateParticipantRemoved = poolCallUpdateParticipantRemoved{}

func (x *CallUpdateParticipantRemoved) DeepCopy(z *CallUpdateParticipantRemoved) {
	z.UserIDs = append(z.UserIDs[:0], x.UserIDs...)
	z.Timeout = x.Timeout
}

func (x *CallUpdateParticipantRemoved) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CallUpdateParticipantRemoved) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *CallUpdateParticipantRemoved) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CallUpdateParticipantRemoved, x)
}

const C_CallUpdateCallPreview int64 = 567542844

type poolCallUpdateCallPreview struct {
	pool sync.Pool
}

func (p *poolCallUpdateCallPreview) Get() *CallUpdateCallPreview {
	x, ok := p.pool.Get().(*CallUpdateCallPreview)
	if !ok {
		x = &CallUpdateCallPreview{}
	}
	return x
}

func (p *poolCallUpdateCallPreview) Put(x *CallUpdateCallPreview) {
	if x == nil {
		return
	}
	x.CallID = 0
	PoolInputPeer.Put(x.Peer)
	x.Peer = nil
	p.pool.Put(x)
}

var PoolCallUpdateCallPreview = poolCallUpdateCallPreview{}

func (x *CallUpdateCallPreview) DeepCopy(z *CallUpdateCallPreview) {
	z.CallID = x.CallID
	if x.Peer != nil {
		if z.Peer == nil {
			z.Peer = PoolInputPeer.Get()
		}
		x.Peer.DeepCopy(z.Peer)
	} else {
		z.Peer = nil
	}
}

func (x *CallUpdateCallPreview) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CallUpdateCallPreview) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *CallUpdateCallPreview) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CallUpdateCallPreview, x)
}

const C_CallUpdateCallCancelled int64 = 4194096602

type poolCallUpdateCallCancelled struct {
	pool sync.Pool
}

func (p *poolCallUpdateCallCancelled) Get() *CallUpdateCallCancelled {
	x, ok := p.pool.Get().(*CallUpdateCallCancelled)
	if !ok {
		x = &CallUpdateCallCancelled{}
	}
	return x
}

func (p *poolCallUpdateCallCancelled) Put(x *CallUpdateCallCancelled) {
	if x == nil {
		return
	}
	x.CallID = 0
	p.pool.Put(x)
}

var PoolCallUpdateCallCancelled = poolCallUpdateCallCancelled{}

func (x *CallUpdateCallCancelled) DeepCopy(z *CallUpdateCallCancelled) {
	z.CallID = x.CallID
}

func (x *CallUpdateCallCancelled) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CallUpdateCallCancelled) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *CallUpdateCallCancelled) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CallUpdateCallCancelled, x)
}

const C_CallUpdateCallJoinRequested int64 = 945899454

type poolCallUpdateCallJoinRequested struct {
	pool sync.Pool
}

func (p *poolCallUpdateCallJoinRequested) Get() *CallUpdateCallJoinRequested {
	x, ok := p.pool.Get().(*CallUpdateCallJoinRequested)
	if !ok {
		x = &CallUpdateCallJoinRequested{}
	}
	return x
}

func (p *poolCallUpdateCallJoinRequested) Put(x *CallUpdateCallJoinRequested) {
	if x == nil {
		return
	}
	x.CallID = 0
	x.CalleeID = 0
	PoolInputPeer.Put(x.Peer)
	x.Peer = nil
	p.pool.Put(x)
}

var PoolCallUpdateCallJoinRequested = poolCallUpdateCallJoinRequested{}

func (x *CallUpdateCallJoinRequested) DeepCopy(z *CallUpdateCallJoinRequested) {
	z.CallID = x.CallID
	z.CalleeID = x.CalleeID
	if x.Peer != nil {
		if z.Peer == nil {
			z.Peer = PoolInputPeer.Get()
		}
		x.Peer.DeepCopy(z.Peer)
	} else {
		z.Peer = nil
	}
}

func (x *CallUpdateCallJoinRequested) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CallUpdateCallJoinRequested) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *CallUpdateCallJoinRequested) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CallUpdateCallJoinRequested, x)
}

const C_CallUpdateParticipantAdminUpdated int64 = 2487316396

type poolCallUpdateParticipantAdminUpdated struct {
	pool sync.Pool
}

func (p *poolCallUpdateParticipantAdminUpdated) Get() *CallUpdateParticipantAdminUpdated {
	x, ok := p.pool.Get().(*CallUpdateParticipantAdminUpdated)
	if !ok {
		x = &CallUpdateParticipantAdminUpdated{}
	}
	return x
}

func (p *poolCallUpdateParticipantAdminUpdated) Put(x *CallUpdateParticipantAdminUpdated) {
	if x == nil {
		return
	}
	x.UserID = 0
	x.Admin = false
	p.pool.Put(x)
}

var PoolCallUpdateParticipantAdminUpdated = poolCallUpdateParticipantAdminUpdated{}

func (x *CallUpdateParticipantAdminUpdated) DeepCopy(z *CallUpdateParticipantAdminUpdated) {
	z.UserID = x.UserID
	z.Admin = x.Admin
}

func (x *CallUpdateParticipantAdminUpdated) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CallUpdateParticipantAdminUpdated) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *CallUpdateParticipantAdminUpdated) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CallUpdateParticipantAdminUpdated, x)
}

const C_CallUpdateShareScreenStreamUpdated int64 = 2763784404

type poolCallUpdateShareScreenStreamUpdated struct {
	pool sync.Pool
}

func (p *poolCallUpdateShareScreenStreamUpdated) Get() *CallUpdateShareScreenStreamUpdated {
	x, ok := p.pool.Get().(*CallUpdateShareScreenStreamUpdated)
	if !ok {
		x = &CallUpdateShareScreenStreamUpdated{}
	}
	return x
}

func (p *poolCallUpdateShareScreenStreamUpdated) Put(x *CallUpdateShareScreenStreamUpdated) {
	if x == nil {
		return
	}
	x.CallID = 0
	x.ConnectionID = 0
	x.StreamID = ""
	p.pool.Put(x)
}

var PoolCallUpdateShareScreenStreamUpdated = poolCallUpdateShareScreenStreamUpdated{}

func (x *CallUpdateShareScreenStreamUpdated) DeepCopy(z *CallUpdateShareScreenStreamUpdated) {
	z.CallID = x.CallID
	z.ConnectionID = x.ConnectionID
	z.StreamID = x.StreamID
}

func (x *CallUpdateShareScreenStreamUpdated) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CallUpdateShareScreenStreamUpdated) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *CallUpdateShareScreenStreamUpdated) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CallUpdateShareScreenStreamUpdated, x)
}

const C_CallUpdateAllConnected int64 = 1993183151

type poolCallUpdateAllConnected struct {
	pool sync.Pool
}

func (p *poolCallUpdateAllConnected) Get() *CallUpdateAllConnected {
	x, ok := p.pool.Get().(*CallUpdateAllConnected)
	if !ok {
		x = &CallUpdateAllConnected{}
	}
	return x
}

func (p *poolCallUpdateAllConnected) Put(x *CallUpdateAllConnected) {
	if x == nil {
		return
	}
	p.pool.Put(x)
}

var PoolCallUpdateAllConnected = poolCallUpdateAllConnected{}

func (x *CallUpdateAllConnected) DeepCopy(z *CallUpdateAllConnected) {
}

func (x *CallUpdateAllConnected) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CallUpdateAllConnected) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *CallUpdateAllConnected) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CallUpdateAllConnected, x)
}

const C_CallUpdateConnectionStatusChanged int64 = 4028141073

type poolCallUpdateConnectionStatusChanged struct {
	pool sync.Pool
}

func (p *poolCallUpdateConnectionStatusChanged) Get() *CallUpdateConnectionStatusChanged {
	x, ok := p.pool.Get().(*CallUpdateConnectionStatusChanged)
	if !ok {
		x = &CallUpdateConnectionStatusChanged{}
	}
	return x
}

func (p *poolCallUpdateConnectionStatusChanged) Put(x *CallUpdateConnectionStatusChanged) {
	if x == nil {
		return
	}
	x.ConnectionID = 0
	x.State = ""
	p.pool.Put(x)
}

var PoolCallUpdateConnectionStatusChanged = poolCallUpdateConnectionStatusChanged{}

func (x *CallUpdateConnectionStatusChanged) DeepCopy(z *CallUpdateConnectionStatusChanged) {
	z.ConnectionID = x.ConnectionID
	z.State = x.State
}

func (x *CallUpdateConnectionStatusChanged) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CallUpdateConnectionStatusChanged) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *CallUpdateConnectionStatusChanged) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CallUpdateConnectionStatusChanged, x)
}

const C_CallUpdateParticipantMuted int64 = 2194386679

type poolCallUpdateParticipantMuted struct {
	pool sync.Pool
}

func (p *poolCallUpdateParticipantMuted) Get() *CallUpdateParticipantMuted {
	x, ok := p.pool.Get().(*CallUpdateParticipantMuted)
	if !ok {
		x = &CallUpdateParticipantMuted{}
	}
	return x
}

func (p *poolCallUpdateParticipantMuted) Put(x *CallUpdateParticipantMuted) {
	if x == nil {
		return
	}
	x.ConnectionID = 0
	x.Muted = false
	x.UserID = 0
	p.pool.Put(x)
}

var PoolCallUpdateParticipantMuted = poolCallUpdateParticipantMuted{}

func (x *CallUpdateParticipantMuted) DeepCopy(z *CallUpdateParticipantMuted) {
	z.ConnectionID = x.ConnectionID
	z.Muted = x.Muted
	z.UserID = x.UserID
}

func (x *CallUpdateParticipantMuted) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CallUpdateParticipantMuted) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *CallUpdateParticipantMuted) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CallUpdateParticipantMuted, x)
}

const C_CallUpdateCallDestroyed int64 = 3684039715

type poolCallUpdateCallDestroyed struct {
	pool sync.Pool
}

func (p *poolCallUpdateCallDestroyed) Get() *CallUpdateCallDestroyed {
	x, ok := p.pool.Get().(*CallUpdateCallDestroyed)
	if !ok {
		x = &CallUpdateCallDestroyed{}
	}
	return x
}

func (p *poolCallUpdateCallDestroyed) Put(x *CallUpdateCallDestroyed) {
	if x == nil {
		return
	}
	x.CallID = 0
	p.pool.Put(x)
}

var PoolCallUpdateCallDestroyed = poolCallUpdateCallDestroyed{}

func (x *CallUpdateCallDestroyed) DeepCopy(z *CallUpdateCallDestroyed) {
	z.CallID = x.CallID
}

func (x *CallUpdateCallDestroyed) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CallUpdateCallDestroyed) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *CallUpdateCallDestroyed) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CallUpdateCallDestroyed, x)
}

const C_CallUpdateLocalMediaSettingsUpdated int64 = 587913546

type poolCallUpdateLocalMediaSettingsUpdated struct {
	pool sync.Pool
}

func (p *poolCallUpdateLocalMediaSettingsUpdated) Get() *CallUpdateLocalMediaSettingsUpdated {
	x, ok := p.pool.Get().(*CallUpdateLocalMediaSettingsUpdated)
	if !ok {
		x = &CallUpdateLocalMediaSettingsUpdated{}
	}
	return x
}

func (p *poolCallUpdateLocalMediaSettingsUpdated) Put(x *CallUpdateLocalMediaSettingsUpdated) {
	if x == nil {
		return
	}
	p.pool.Put(x)
}

var PoolCallUpdateLocalMediaSettingsUpdated = poolCallUpdateLocalMediaSettingsUpdated{}

func (x *CallUpdateLocalMediaSettingsUpdated) DeepCopy(z *CallUpdateLocalMediaSettingsUpdated) {
}

func (x *CallUpdateLocalMediaSettingsUpdated) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *CallUpdateLocalMediaSettingsUpdated) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *CallUpdateLocalMediaSettingsUpdated) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_CallUpdateLocalMediaSettingsUpdated, x)
}

func init() {
	registry.RegisterConstructor(1147111688, "CallMediaSettings")
	registry.RegisterConstructor(2652007354, "CallParticipant")
	registry.RegisterConstructor(2748774954, "CallRTCIceCandidate")
	registry.RegisterConstructor(1457208388, "CallRTCIceServer")
	registry.RegisterConstructor(3450901888, "CallConnection")
	registry.RegisterConstructor(2556114354, "CallUpdateCallRequested")
	registry.RegisterConstructor(2134109006, "CallUpdateCallAccepted")
	registry.RegisterConstructor(3496218809, "CallUpdateStreamUpdated")
	registry.RegisterConstructor(2339651845, "CallUpdateCallRejected")
	registry.RegisterConstructor(3922101985, "CallUpdateMediaSettingsUpdated")
	registry.RegisterConstructor(1043624904, "CallUpdateLocalStreamUpdated")
	registry.RegisterConstructor(420503198, "CallUpdateCallTimeout")
	registry.RegisterConstructor(1424725011, "CallUpdateCallAck")
	registry.RegisterConstructor(1005455524, "CallUpdateParticipantJoined")
	registry.RegisterConstructor(2062471712, "CallUpdateParticipantLeft")
	registry.RegisterConstructor(4138786615, "CallUpdateParticipantRemoved")
	registry.RegisterConstructor(567542844, "CallUpdateCallPreview")
	registry.RegisterConstructor(4194096602, "CallUpdateCallCancelled")
	registry.RegisterConstructor(945899454, "CallUpdateCallJoinRequested")
	registry.RegisterConstructor(2487316396, "CallUpdateParticipantAdminUpdated")
	registry.RegisterConstructor(2763784404, "CallUpdateShareScreenStreamUpdated")
	registry.RegisterConstructor(1993183151, "CallUpdateAllConnected")
	registry.RegisterConstructor(4028141073, "CallUpdateConnectionStatusChanged")
	registry.RegisterConstructor(2194386679, "CallUpdateParticipantMuted")
	registry.RegisterConstructor(3684039715, "CallUpdateCallDestroyed")
	registry.RegisterConstructor(587913546, "CallUpdateLocalMediaSettingsUpdated")
}
