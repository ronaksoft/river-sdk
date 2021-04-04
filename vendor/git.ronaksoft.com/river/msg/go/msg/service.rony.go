// Code generated by Rony's protoc plugin; DO NOT EDIT.

package msg

import (
	edge "github.com/ronaksoft/rony/edge"
	registry "github.com/ronaksoft/rony/registry"
	proto "google.golang.org/protobuf/proto"
	sync "sync"
)

const C_ServiceSendMessage int64 = 824547051

type poolServiceSendMessage struct {
	pool sync.Pool
}

func (p *poolServiceSendMessage) Get() *ServiceSendMessage {
	x, ok := p.pool.Get().(*ServiceSendMessage)
	if !ok {
		x = &ServiceSendMessage{}
	}
	return x
}

func (p *poolServiceSendMessage) Put(x *ServiceSendMessage) {
	if x == nil {
		return
	}
	x.OnBehalf = 0
	x.RandomID = 0
	PoolInputPeer.Put(x.Peer)
	x.Peer = nil
	x.Body = ""
	x.ReplyTo = 0
	x.ClearDraft = false
	for _, z := range x.Entities {
		PoolMessageEntity.Put(z)
	}
	x.Entities = x.Entities[:0]
	p.pool.Put(x)
}

var PoolServiceSendMessage = poolServiceSendMessage{}

func (x *ServiceSendMessage) DeepCopy(z *ServiceSendMessage) {
	z.OnBehalf = x.OnBehalf
	z.RandomID = x.RandomID
	if x.Peer != nil {
		if z.Peer == nil {
			z.Peer = PoolInputPeer.Get()
		}
		x.Peer.DeepCopy(z.Peer)
	} else {
		z.Peer = nil
	}
	z.Body = x.Body
	z.ReplyTo = x.ReplyTo
	z.ClearDraft = x.ClearDraft
	for idx := range x.Entities {
		if x.Entities[idx] != nil {
			xx := PoolMessageEntity.Get()
			x.Entities[idx].DeepCopy(xx)
			z.Entities = append(z.Entities, xx)
		}
	}
}

func (x *ServiceSendMessage) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *ServiceSendMessage) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *ServiceSendMessage) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_ServiceSendMessage, x)
}

func init() {
	registry.RegisterConstructor(824547051, "ServiceSendMessage")
}
