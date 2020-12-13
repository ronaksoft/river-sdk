package msg

import (
	edge "github.com/ronaksoft/rony/edge"
	registry "github.com/ronaksoft/rony/registry"
	proto "google.golang.org/protobuf/proto"
	sync "sync"
)

const C_PhoneInitCall int64 = 2975617068

type poolPhoneInitCall struct {
	pool sync.Pool
}

func (p *poolPhoneInitCall) Get() *PhoneInitCall {
	x, ok := p.pool.Get().(*PhoneInitCall)
	if !ok {
		return &PhoneInitCall{}
	}
	return x
}

func (p *poolPhoneInitCall) Put(x *PhoneInitCall) {
	if x.Peer != nil {
		PoolInputPeer.Put(x.Peer)
		x.Peer = nil
	}
	p.pool.Put(x)
}

var PoolPhoneInitCall = poolPhoneInitCall{}

const C_PhoneRequestCall int64 = 907942641

type poolPhoneRequestCall struct {
	pool sync.Pool
}

func (p *poolPhoneRequestCall) Get() *PhoneRequestCall {
	x, ok := p.pool.Get().(*PhoneRequestCall)
	if !ok {
		return &PhoneRequestCall{}
	}
	return x
}

func (p *poolPhoneRequestCall) Put(x *PhoneRequestCall) {
	x.RandomID = 0
	if x.Peer != nil {
		PoolInputPeer.Put(x.Peer)
		x.Peer = nil
	}
	x.Initiator = false
	x.Participants = x.Participants[:0]
	x.CallID = 0
	p.pool.Put(x)
}

var PoolPhoneRequestCall = poolPhoneRequestCall{}

const C_PhoneAcceptCall int64 = 4133092858

type poolPhoneAcceptCall struct {
	pool sync.Pool
}

func (p *poolPhoneAcceptCall) Get() *PhoneAcceptCall {
	x, ok := p.pool.Get().(*PhoneAcceptCall)
	if !ok {
		return &PhoneAcceptCall{}
	}
	return x
}

func (p *poolPhoneAcceptCall) Put(x *PhoneAcceptCall) {
	if x.Peer != nil {
		PoolInputPeer.Put(x.Peer)
		x.Peer = nil
	}
	x.CallID = 0
	x.Participants = x.Participants[:0]
	p.pool.Put(x)
}

var PoolPhoneAcceptCall = poolPhoneAcceptCall{}

const C_PhoneDiscardCall int64 = 2712700137

type poolPhoneDiscardCall struct {
	pool sync.Pool
}

func (p *poolPhoneDiscardCall) Get() *PhoneDiscardCall {
	x, ok := p.pool.Get().(*PhoneDiscardCall)
	if !ok {
		return &PhoneDiscardCall{}
	}
	return x
}

func (p *poolPhoneDiscardCall) Put(x *PhoneDiscardCall) {
	if x.Peer != nil {
		PoolInputPeer.Put(x.Peer)
		x.Peer = nil
	}
	x.CallID = 0
	x.Participants = x.Participants[:0]
	x.Duration = 0
	x.Reason = 0
	p.pool.Put(x)
}

var PoolPhoneDiscardCall = poolPhoneDiscardCall{}

const C_PhoneUpdateCall int64 = 1976202226

type poolPhoneUpdateCall struct {
	pool sync.Pool
}

func (p *poolPhoneUpdateCall) Get() *PhoneUpdateCall {
	x, ok := p.pool.Get().(*PhoneUpdateCall)
	if !ok {
		return &PhoneUpdateCall{}
	}
	return x
}

func (p *poolPhoneUpdateCall) Put(x *PhoneUpdateCall) {
	if x.Peer != nil {
		PoolInputPeer.Put(x.Peer)
		x.Peer = nil
	}
	x.CallID = 0
	x.Participants = x.Participants[:0]
	x.Action = 0
	x.ActionData = x.ActionData[:0]
	p.pool.Put(x)
}

var PoolPhoneUpdateCall = poolPhoneUpdateCall{}

const C_PhoneRateCall int64 = 2215486159

type poolPhoneRateCall struct {
	pool sync.Pool
}

func (p *poolPhoneRateCall) Get() *PhoneRateCall {
	x, ok := p.pool.Get().(*PhoneRateCall)
	if !ok {
		return &PhoneRateCall{}
	}
	return x
}

func (p *poolPhoneRateCall) Put(x *PhoneRateCall) {
	if x.Peer != nil {
		PoolInputPeer.Put(x.Peer)
		x.Peer = nil
	}
	x.CallID = 0
	x.Rate = 0
	x.Comment = ""
	p.pool.Put(x)
}

var PoolPhoneRateCall = poolPhoneRateCall{}

const C_PhoneCall int64 = 3296664529

type poolPhoneCall struct {
	pool sync.Pool
}

func (p *poolPhoneCall) Get() *PhoneCall {
	x, ok := p.pool.Get().(*PhoneCall)
	if !ok {
		return &PhoneCall{}
	}
	return x
}

func (p *poolPhoneCall) Put(x *PhoneCall) {
	x.ID = 0
	x.Date = 0
	p.pool.Put(x)
}

var PoolPhoneCall = poolPhoneCall{}

const C_PhoneInit int64 = 3464876187

type poolPhoneInit struct {
	pool sync.Pool
}

func (p *poolPhoneInit) Get() *PhoneInit {
	x, ok := p.pool.Get().(*PhoneInit)
	if !ok {
		return &PhoneInit{}
	}
	return x
}

func (p *poolPhoneInit) Put(x *PhoneInit) {
	x.IceServers = x.IceServers[:0]
	p.pool.Put(x)
}

var PoolPhoneInit = poolPhoneInit{}

const C_IceServer int64 = 4291892363

type poolIceServer struct {
	pool sync.Pool
}

func (p *poolIceServer) Get() *IceServer {
	x, ok := p.pool.Get().(*IceServer)
	if !ok {
		return &IceServer{}
	}
	return x
}

func (p *poolIceServer) Put(x *IceServer) {
	x.Urls = x.Urls[:0]
	x.Username = ""
	x.Credential = ""
	p.pool.Put(x)
}

var PoolIceServer = poolIceServer{}

const C_PhoneParticipant int64 = 226273622

type poolPhoneParticipant struct {
	pool sync.Pool
}

func (p *poolPhoneParticipant) Get() *PhoneParticipant {
	x, ok := p.pool.Get().(*PhoneParticipant)
	if !ok {
		return &PhoneParticipant{}
	}
	return x
}

func (p *poolPhoneParticipant) Put(x *PhoneParticipant) {
	x.ConnectionId = 0
	if x.Peer != nil {
		PoolInputUser.Put(x.Peer)
		x.Peer = nil
	}
	x.Initiator = false
	x.Admin = false
	p.pool.Put(x)
}

var PoolPhoneParticipant = poolPhoneParticipant{}

const C_PhoneParticipantSDP int64 = 545454774

type poolPhoneParticipantSDP struct {
	pool sync.Pool
}

func (p *poolPhoneParticipantSDP) Get() *PhoneParticipantSDP {
	x, ok := p.pool.Get().(*PhoneParticipantSDP)
	if !ok {
		return &PhoneParticipantSDP{}
	}
	return x
}

func (p *poolPhoneParticipantSDP) Put(x *PhoneParticipantSDP) {
	x.ConnectionId = 0
	if x.Peer != nil {
		PoolInputUser.Put(x.Peer)
		x.Peer = nil
	}
	x.SDP = ""
	x.Type = ""
	p.pool.Put(x)
}

var PoolPhoneParticipantSDP = poolPhoneParticipantSDP{}

const C_PhoneActionCallEmpty int64 = 1073285997

type poolPhoneActionCallEmpty struct {
	pool sync.Pool
}

func (p *poolPhoneActionCallEmpty) Get() *PhoneActionCallEmpty {
	x, ok := p.pool.Get().(*PhoneActionCallEmpty)
	if !ok {
		return &PhoneActionCallEmpty{}
	}
	return x
}

func (p *poolPhoneActionCallEmpty) Put(x *PhoneActionCallEmpty) {
	x.Empty = false
	p.pool.Put(x)
}

var PoolPhoneActionCallEmpty = poolPhoneActionCallEmpty{}

const C_PhoneActionAccepted int64 = 2493210645

type poolPhoneActionAccepted struct {
	pool sync.Pool
}

func (p *poolPhoneActionAccepted) Get() *PhoneActionAccepted {
	x, ok := p.pool.Get().(*PhoneActionAccepted)
	if !ok {
		return &PhoneActionAccepted{}
	}
	return x
}

func (p *poolPhoneActionAccepted) Put(x *PhoneActionAccepted) {
	x.SDP = ""
	x.Type = ""
	p.pool.Put(x)
}

var PoolPhoneActionAccepted = poolPhoneActionAccepted{}

const C_PhoneActionRequested int64 = 1678316869

type poolPhoneActionRequested struct {
	pool sync.Pool
}

func (p *poolPhoneActionRequested) Get() *PhoneActionRequested {
	x, ok := p.pool.Get().(*PhoneActionRequested)
	if !ok {
		return &PhoneActionRequested{}
	}
	return x
}

func (p *poolPhoneActionRequested) Put(x *PhoneActionRequested) {
	x.SDP = ""
	x.Type = ""
	x.Participants = x.Participants[:0]
	p.pool.Put(x)
}

var PoolPhoneActionRequested = poolPhoneActionRequested{}

const C_PhoneActionCallWaiting int64 = 3634710697

type poolPhoneActionCallWaiting struct {
	pool sync.Pool
}

func (p *poolPhoneActionCallWaiting) Get() *PhoneActionCallWaiting {
	x, ok := p.pool.Get().(*PhoneActionCallWaiting)
	if !ok {
		return &PhoneActionCallWaiting{}
	}
	return x
}

func (p *poolPhoneActionCallWaiting) Put(x *PhoneActionCallWaiting) {
	x.Empty = false
	p.pool.Put(x)
}

var PoolPhoneActionCallWaiting = poolPhoneActionCallWaiting{}

const C_PhoneActionDiscarded int64 = 4285966731

type poolPhoneActionDiscarded struct {
	pool sync.Pool
}

func (p *poolPhoneActionDiscarded) Get() *PhoneActionDiscarded {
	x, ok := p.pool.Get().(*PhoneActionDiscarded)
	if !ok {
		return &PhoneActionDiscarded{}
	}
	return x
}

func (p *poolPhoneActionDiscarded) Put(x *PhoneActionDiscarded) {
	x.Duration = 0
	x.Video = false
	x.Reason = 0
	p.pool.Put(x)
}

var PoolPhoneActionDiscarded = poolPhoneActionDiscarded{}

const C_PhoneActionIceExchange int64 = 1618781621

type poolPhoneActionIceExchange struct {
	pool sync.Pool
}

func (p *poolPhoneActionIceExchange) Get() *PhoneActionIceExchange {
	x, ok := p.pool.Get().(*PhoneActionIceExchange)
	if !ok {
		return &PhoneActionIceExchange{}
	}
	return x
}

func (p *poolPhoneActionIceExchange) Put(x *PhoneActionIceExchange) {
	x.Candidate = ""
	x.SdpMLineIndex = 0
	x.SdpMid = ""
	x.UsernameFragment = ""
	p.pool.Put(x)
}

var PoolPhoneActionIceExchange = poolPhoneActionIceExchange{}

const C_PhoneMediaSettingsUpdated int64 = 163140236

type poolPhoneMediaSettingsUpdated struct {
	pool sync.Pool
}

func (p *poolPhoneMediaSettingsUpdated) Get() *PhoneMediaSettingsUpdated {
	x, ok := p.pool.Get().(*PhoneMediaSettingsUpdated)
	if !ok {
		return &PhoneMediaSettingsUpdated{}
	}
	return x
}

func (p *poolPhoneMediaSettingsUpdated) Put(x *PhoneMediaSettingsUpdated) {
	x.Video = false
	x.Audio = false
	p.pool.Put(x)
}

var PoolPhoneMediaSettingsUpdated = poolPhoneMediaSettingsUpdated{}

const C_PhoneReactionSet int64 = 3821475130

type poolPhoneReactionSet struct {
	pool sync.Pool
}

func (p *poolPhoneReactionSet) Get() *PhoneReactionSet {
	x, ok := p.pool.Get().(*PhoneReactionSet)
	if !ok {
		return &PhoneReactionSet{}
	}
	return x
}

func (p *poolPhoneReactionSet) Put(x *PhoneReactionSet) {
	x.Reaction = ""
	p.pool.Put(x)
}

var PoolPhoneReactionSet = poolPhoneReactionSet{}

const C_PhoneSDPOffer int64 = 2063600460

type poolPhoneSDPOffer struct {
	pool sync.Pool
}

func (p *poolPhoneSDPOffer) Get() *PhoneSDPOffer {
	x, ok := p.pool.Get().(*PhoneSDPOffer)
	if !ok {
		return &PhoneSDPOffer{}
	}
	return x
}

func (p *poolPhoneSDPOffer) Put(x *PhoneSDPOffer) {
	x.SDP = ""
	x.Type = ""
	p.pool.Put(x)
}

var PoolPhoneSDPOffer = poolPhoneSDPOffer{}

const C_PhoneSDPAnswer int64 = 1686408377

type poolPhoneSDPAnswer struct {
	pool sync.Pool
}

func (p *poolPhoneSDPAnswer) Get() *PhoneSDPAnswer {
	x, ok := p.pool.Get().(*PhoneSDPAnswer)
	if !ok {
		return &PhoneSDPAnswer{}
	}
	return x
}

func (p *poolPhoneSDPAnswer) Put(x *PhoneSDPAnswer) {
	x.SDP = ""
	x.Type = ""
	p.pool.Put(x)
}

var PoolPhoneSDPAnswer = poolPhoneSDPAnswer{}

func init() {
	registry.RegisterConstructor(2975617068, "PhoneInitCall")
	registry.RegisterConstructor(907942641, "PhoneRequestCall")
	registry.RegisterConstructor(4133092858, "PhoneAcceptCall")
	registry.RegisterConstructor(2712700137, "PhoneDiscardCall")
	registry.RegisterConstructor(1976202226, "PhoneUpdateCall")
	registry.RegisterConstructor(2215486159, "PhoneRateCall")
	registry.RegisterConstructor(3296664529, "PhoneCall")
	registry.RegisterConstructor(3464876187, "PhoneInit")
	registry.RegisterConstructor(4291892363, "IceServer")
	registry.RegisterConstructor(226273622, "PhoneParticipant")
	registry.RegisterConstructor(545454774, "PhoneParticipantSDP")
	registry.RegisterConstructor(1073285997, "PhoneActionCallEmpty")
	registry.RegisterConstructor(2493210645, "PhoneActionAccepted")
	registry.RegisterConstructor(1678316869, "PhoneActionRequested")
	registry.RegisterConstructor(3634710697, "PhoneActionCallWaiting")
	registry.RegisterConstructor(4285966731, "PhoneActionDiscarded")
	registry.RegisterConstructor(1618781621, "PhoneActionIceExchange")
	registry.RegisterConstructor(163140236, "PhoneMediaSettingsUpdated")
	registry.RegisterConstructor(3821475130, "PhoneReactionSet")
	registry.RegisterConstructor(2063600460, "PhoneSDPOffer")
	registry.RegisterConstructor(1686408377, "PhoneSDPAnswer")
}

func (x *PhoneInitCall) DeepCopy(z *PhoneInitCall) {
	if x.Peer != nil {
		z.Peer = PoolInputPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
}

func (x *PhoneRequestCall) DeepCopy(z *PhoneRequestCall) {
	z.RandomID = x.RandomID
	if x.Peer != nil {
		z.Peer = PoolInputPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.Initiator = x.Initiator
	for idx := range x.Participants {
		if x.Participants[idx] != nil {
			xx := PoolPhoneParticipantSDP.Get()
			x.Participants[idx].DeepCopy(xx)
			z.Participants = append(z.Participants, xx)
		}
	}
	z.CallID = x.CallID
}

func (x *PhoneAcceptCall) DeepCopy(z *PhoneAcceptCall) {
	if x.Peer != nil {
		z.Peer = PoolInputPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.CallID = x.CallID
	for idx := range x.Participants {
		if x.Participants[idx] != nil {
			xx := PoolPhoneParticipantSDP.Get()
			x.Participants[idx].DeepCopy(xx)
			z.Participants = append(z.Participants, xx)
		}
	}
}

func (x *PhoneDiscardCall) DeepCopy(z *PhoneDiscardCall) {
	if x.Peer != nil {
		z.Peer = PoolInputPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.CallID = x.CallID
	for idx := range x.Participants {
		if x.Participants[idx] != nil {
			xx := PoolInputUser.Get()
			x.Participants[idx].DeepCopy(xx)
			z.Participants = append(z.Participants, xx)
		}
	}
	z.Duration = x.Duration
	z.Reason = x.Reason
}

func (x *PhoneUpdateCall) DeepCopy(z *PhoneUpdateCall) {
	if x.Peer != nil {
		z.Peer = PoolInputPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.CallID = x.CallID
	for idx := range x.Participants {
		if x.Participants[idx] != nil {
			xx := PoolInputUser.Get()
			x.Participants[idx].DeepCopy(xx)
			z.Participants = append(z.Participants, xx)
		}
	}
	z.Action = x.Action
	z.ActionData = append(z.ActionData[:0], x.ActionData...)
}

func (x *PhoneRateCall) DeepCopy(z *PhoneRateCall) {
	if x.Peer != nil {
		z.Peer = PoolInputPeer.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.CallID = x.CallID
	z.Rate = x.Rate
	z.Comment = x.Comment
}

func (x *PhoneCall) DeepCopy(z *PhoneCall) {
	z.ID = x.ID
	z.Date = x.Date
}

func (x *PhoneInit) DeepCopy(z *PhoneInit) {
	for idx := range x.IceServers {
		if x.IceServers[idx] != nil {
			xx := PoolIceServer.Get()
			x.IceServers[idx].DeepCopy(xx)
			z.IceServers = append(z.IceServers, xx)
		}
	}
}

func (x *IceServer) DeepCopy(z *IceServer) {
	z.Urls = append(z.Urls[:0], x.Urls...)
	z.Username = x.Username
	z.Credential = x.Credential
}

func (x *PhoneParticipant) DeepCopy(z *PhoneParticipant) {
	z.ConnectionId = x.ConnectionId
	if x.Peer != nil {
		z.Peer = PoolInputUser.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.Initiator = x.Initiator
	z.Admin = x.Admin
}

func (x *PhoneParticipantSDP) DeepCopy(z *PhoneParticipantSDP) {
	z.ConnectionId = x.ConnectionId
	if x.Peer != nil {
		z.Peer = PoolInputUser.Get()
		x.Peer.DeepCopy(z.Peer)
	}
	z.SDP = x.SDP
	z.Type = x.Type
}

func (x *PhoneActionCallEmpty) DeepCopy(z *PhoneActionCallEmpty) {
	z.Empty = x.Empty
}

func (x *PhoneActionAccepted) DeepCopy(z *PhoneActionAccepted) {
	z.SDP = x.SDP
	z.Type = x.Type
}

func (x *PhoneActionRequested) DeepCopy(z *PhoneActionRequested) {
	z.SDP = x.SDP
	z.Type = x.Type
	for idx := range x.Participants {
		if x.Participants[idx] != nil {
			xx := PoolPhoneParticipant.Get()
			x.Participants[idx].DeepCopy(xx)
			z.Participants = append(z.Participants, xx)
		}
	}
}

func (x *PhoneActionCallWaiting) DeepCopy(z *PhoneActionCallWaiting) {
	z.Empty = x.Empty
}

func (x *PhoneActionDiscarded) DeepCopy(z *PhoneActionDiscarded) {
	z.Duration = x.Duration
	z.Video = x.Video
	z.Reason = x.Reason
}

func (x *PhoneActionIceExchange) DeepCopy(z *PhoneActionIceExchange) {
	z.Candidate = x.Candidate
	z.SdpMLineIndex = x.SdpMLineIndex
	z.SdpMid = x.SdpMid
	z.UsernameFragment = x.UsernameFragment
}

func (x *PhoneMediaSettingsUpdated) DeepCopy(z *PhoneMediaSettingsUpdated) {
	z.Video = x.Video
	z.Audio = x.Audio
}

func (x *PhoneReactionSet) DeepCopy(z *PhoneReactionSet) {
	z.Reaction = x.Reaction
}

func (x *PhoneSDPOffer) DeepCopy(z *PhoneSDPOffer) {
	z.SDP = x.SDP
	z.Type = x.Type
}

func (x *PhoneSDPAnswer) DeepCopy(z *PhoneSDPAnswer) {
	z.SDP = x.SDP
	z.Type = x.Type
}

func (x *PhoneInitCall) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_PhoneInitCall, x)
}

func (x *PhoneRequestCall) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_PhoneRequestCall, x)
}

func (x *PhoneAcceptCall) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_PhoneAcceptCall, x)
}

func (x *PhoneDiscardCall) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_PhoneDiscardCall, x)
}

func (x *PhoneUpdateCall) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_PhoneUpdateCall, x)
}

func (x *PhoneRateCall) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_PhoneRateCall, x)
}

func (x *PhoneCall) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_PhoneCall, x)
}

func (x *PhoneInit) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_PhoneInit, x)
}

func (x *IceServer) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_IceServer, x)
}

func (x *PhoneParticipant) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_PhoneParticipant, x)
}

func (x *PhoneParticipantSDP) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_PhoneParticipantSDP, x)
}

func (x *PhoneActionCallEmpty) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_PhoneActionCallEmpty, x)
}

func (x *PhoneActionAccepted) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_PhoneActionAccepted, x)
}

func (x *PhoneActionRequested) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_PhoneActionRequested, x)
}

func (x *PhoneActionCallWaiting) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_PhoneActionCallWaiting, x)
}

func (x *PhoneActionDiscarded) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_PhoneActionDiscarded, x)
}

func (x *PhoneActionIceExchange) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_PhoneActionIceExchange, x)
}

func (x *PhoneMediaSettingsUpdated) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_PhoneMediaSettingsUpdated, x)
}

func (x *PhoneReactionSet) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_PhoneReactionSet, x)
}

func (x *PhoneSDPOffer) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_PhoneSDPOffer, x)
}

func (x *PhoneSDPAnswer) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_PhoneSDPAnswer, x)
}

func (x *PhoneInitCall) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *PhoneRequestCall) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *PhoneAcceptCall) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *PhoneDiscardCall) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *PhoneUpdateCall) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *PhoneRateCall) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *PhoneCall) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *PhoneInit) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *IceServer) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *PhoneParticipant) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *PhoneParticipantSDP) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *PhoneActionCallEmpty) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *PhoneActionAccepted) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *PhoneActionRequested) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *PhoneActionCallWaiting) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *PhoneActionDiscarded) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *PhoneActionIceExchange) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *PhoneMediaSettingsUpdated) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *PhoneReactionSet) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *PhoneSDPOffer) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *PhoneSDPAnswer) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *PhoneInitCall) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *PhoneRequestCall) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *PhoneAcceptCall) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *PhoneDiscardCall) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *PhoneUpdateCall) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *PhoneRateCall) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *PhoneCall) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *PhoneInit) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *IceServer) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *PhoneParticipant) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *PhoneParticipantSDP) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *PhoneActionCallEmpty) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *PhoneActionAccepted) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *PhoneActionRequested) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *PhoneActionCallWaiting) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *PhoneActionDiscarded) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *PhoneActionIceExchange) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *PhoneMediaSettingsUpdated) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *PhoneReactionSet) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *PhoneSDPOffer) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *PhoneSDPAnswer) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}
