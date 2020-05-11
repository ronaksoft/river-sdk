// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: chat.api.phone.proto

package msg

import (
	fmt "fmt"
	pbytes "github.com/gobwas/pool/pbytes"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	math "math"
	sync "sync"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

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
	p.pool.Put(x)
}

var PoolPhoneAcceptCall = poolPhoneAcceptCall{}

func ResultPhoneAcceptCall(out *MessageEnvelope, res *PhoneAcceptCall) {
	out.Constructor = C_PhoneAcceptCall
	protoSize := res.Size()
	if protoSize > cap(out.Message) {
		pbytes.Put(out.Message)
		out.Message = pbytes.GetLen(protoSize)
	} else {
		out.Message = out.Message[:protoSize]
	}
	res.MarshalToSizedBuffer(out.Message)
}

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
	p.pool.Put(x)
}

var PoolPhoneRequestCall = poolPhoneRequestCall{}

func ResultPhoneRequestCall(out *MessageEnvelope, res *PhoneRequestCall) {
	out.Constructor = C_PhoneRequestCall
	protoSize := res.Size()
	if protoSize > cap(out.Message) {
		pbytes.Put(out.Message)
		out.Message = pbytes.GetLen(protoSize)
	} else {
		out.Message = out.Message[:protoSize]
	}
	res.MarshalToSizedBuffer(out.Message)
}

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
	p.pool.Put(x)
}

var PoolPhoneDiscardCall = poolPhoneDiscardCall{}

func ResultPhoneDiscardCall(out *MessageEnvelope, res *PhoneDiscardCall) {
	out.Constructor = C_PhoneDiscardCall
	protoSize := res.Size()
	if protoSize > cap(out.Message) {
		pbytes.Put(out.Message)
		out.Message = pbytes.GetLen(protoSize)
	} else {
		out.Message = out.Message[:protoSize]
	}
	res.MarshalToSizedBuffer(out.Message)
}

const C_PhoneReceivedCall int64 = 1863246318

type poolPhoneReceivedCall struct {
	pool sync.Pool
}

func (p *poolPhoneReceivedCall) Get() *PhoneReceivedCall {
	x, ok := p.pool.Get().(*PhoneReceivedCall)
	if !ok {
		return &PhoneReceivedCall{}
	}
	return x
}

func (p *poolPhoneReceivedCall) Put(x *PhoneReceivedCall) {
	p.pool.Put(x)
}

var PoolPhoneReceivedCall = poolPhoneReceivedCall{}

func ResultPhoneReceivedCall(out *MessageEnvelope, res *PhoneReceivedCall) {
	out.Constructor = C_PhoneReceivedCall
	protoSize := res.Size()
	if protoSize > cap(out.Message) {
		pbytes.Put(out.Message)
		out.Message = pbytes.GetLen(protoSize)
	} else {
		out.Message = out.Message[:protoSize]
	}
	res.MarshalToSizedBuffer(out.Message)
}

const C_PhoneSetCallRating int64 = 2805134474

type poolPhoneSetCallRating struct {
	pool sync.Pool
}

func (p *poolPhoneSetCallRating) Get() *PhoneSetCallRating {
	x, ok := p.pool.Get().(*PhoneSetCallRating)
	if !ok {
		return &PhoneSetCallRating{}
	}
	x.Comment = ""
	return x
}

func (p *poolPhoneSetCallRating) Put(x *PhoneSetCallRating) {
	p.pool.Put(x)
}

var PoolPhoneSetCallRating = poolPhoneSetCallRating{}

func ResultPhoneSetCallRating(out *MessageEnvelope, res *PhoneSetCallRating) {
	out.Constructor = C_PhoneSetCallRating
	protoSize := res.Size()
	if protoSize > cap(out.Message) {
		pbytes.Put(out.Message)
		out.Message = pbytes.GetLen(protoSize)
	} else {
		out.Message = out.Message[:protoSize]
	}
	res.MarshalToSizedBuffer(out.Message)
}

const C_PhoneCall int64 = 3296664529

type poolPhoneCall struct {
	pool sync.Pool
}

func (p *poolPhoneCall) Get() *PhoneCall {
	x, ok := p.pool.Get().(*PhoneCall)
	if !ok {
		return &PhoneCall{}
	}
	x.StunServers = x.StunServers[:0]
	return x
}

func (p *poolPhoneCall) Put(x *PhoneCall) {
	p.pool.Put(x)
}

var PoolPhoneCall = poolPhoneCall{}

func ResultPhoneCall(out *MessageEnvelope, res *PhoneCall) {
	out.Constructor = C_PhoneCall
	protoSize := res.Size()
	if protoSize > cap(out.Message) {
		pbytes.Put(out.Message)
		out.Message = pbytes.GetLen(protoSize)
	} else {
		out.Message = out.Message[:protoSize]
	}
	res.MarshalToSizedBuffer(out.Message)
}

func init() {
	ConstructorNames[4133092858] = "PhoneAcceptCall"
	ConstructorNames[907942641] = "PhoneRequestCall"
	ConstructorNames[2712700137] = "PhoneDiscardCall"
	ConstructorNames[1863246318] = "PhoneReceivedCall"
	ConstructorNames[2805134474] = "PhoneSetCallRating"
	ConstructorNames[3296664529] = "PhoneCall"
}
