// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: chat.api.seal.proto

package msg

import (
	fmt "fmt"
	pbytes "github.com/gobwas/pool/pbytes"
	proto "github.com/gogo/protobuf/proto"
	math "math"
	sync "sync"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

const C_SealSetPubKey int64 = 2075713772

type poolSealSetPubKey struct {
	pool sync.Pool
}

func (p *poolSealSetPubKey) Get() *SealSetPubKey {
	x, ok := p.pool.Get().(*SealSetPubKey)
	if !ok {
		return &SealSetPubKey{}
	}
	return x
}

func (p *poolSealSetPubKey) Put(x *SealSetPubKey) {
	p.pool.Put(x)
}

var PoolSealSetPubKey = poolSealSetPubKey{}

func ResultSealSetPubKey(out *MessageEnvelope, res *SealSetPubKey) {
	out.Constructor = C_SealSetPubKey
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
	ConstructorNames[2075713772] = "SealSetPubKey"
}