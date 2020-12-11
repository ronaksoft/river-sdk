package msg

import (
	edge "github.com/ronaksoft/rony/edge"
	registry "github.com/ronaksoft/rony/registry"
	proto "google.golang.org/protobuf/proto"
	sync "sync"
)

const C_EchoWithDelay int64 = 606168969

type poolEchoWithDelay struct {
	pool sync.Pool
}

func (p *poolEchoWithDelay) Get() *EchoWithDelay {
	x, ok := p.pool.Get().(*EchoWithDelay)
	if !ok {
		return &EchoWithDelay{}
	}
	return x
}

func (p *poolEchoWithDelay) Put(x *EchoWithDelay) {
	x.DelayInSeconds = 0
	p.pool.Put(x)
}

var PoolEchoWithDelay = poolEchoWithDelay{}

const C_TestRequest int64 = 179707619

type poolTestRequest struct {
	pool sync.Pool
}

func (p *poolTestRequest) Get() *TestRequest {
	x, ok := p.pool.Get().(*TestRequest)
	if !ok {
		return &TestRequest{}
	}
	return x
}

func (p *poolTestRequest) Put(x *TestRequest) {
	x.Payload = x.Payload[:0]
	x.Hash = x.Hash[:0]
	p.pool.Put(x)
}

var PoolTestRequest = poolTestRequest{}

const C_TestResponse int64 = 1744377725

type poolTestResponse struct {
	pool sync.Pool
}

func (p *poolTestResponse) Get() *TestResponse {
	x, ok := p.pool.Get().(*TestResponse)
	if !ok {
		return &TestResponse{}
	}
	return x
}

func (p *poolTestResponse) Put(x *TestResponse) {
	x.Hash = x.Hash[:0]
	p.pool.Put(x)
}

var PoolTestResponse = poolTestResponse{}

const C_TestRequestWithString int64 = 4230997651

type poolTestRequestWithString struct {
	pool sync.Pool
}

func (p *poolTestRequestWithString) Get() *TestRequestWithString {
	x, ok := p.pool.Get().(*TestRequestWithString)
	if !ok {
		return &TestRequestWithString{}
	}
	return x
}

func (p *poolTestRequestWithString) Put(x *TestRequestWithString) {
	x.Payload = ""
	x.Hash = ""
	p.pool.Put(x)
}

var PoolTestRequestWithString = poolTestRequestWithString{}

const C_TestResponseWithString int64 = 2503087606

type poolTestResponseWithString struct {
	pool sync.Pool
}

func (p *poolTestResponseWithString) Get() *TestResponseWithString {
	x, ok := p.pool.Get().(*TestResponseWithString)
	if !ok {
		return &TestResponseWithString{}
	}
	return x
}

func (p *poolTestResponseWithString) Put(x *TestResponseWithString) {
	x.Hash = x.Hash[:0]
	p.pool.Put(x)
}

var PoolTestResponseWithString = poolTestResponseWithString{}

func init() {
	registry.RegisterConstructor(606168969, "msg.EchoWithDelay")
	registry.RegisterConstructor(179707619, "msg.TestRequest")
	registry.RegisterConstructor(1744377725, "msg.TestResponse")
	registry.RegisterConstructor(4230997651, "msg.TestRequestWithString")
	registry.RegisterConstructor(2503087606, "msg.TestResponseWithString")
}

func (x *EchoWithDelay) DeepCopy(z *EchoWithDelay) {
	z.DelayInSeconds = x.DelayInSeconds
}

func (x *TestRequest) DeepCopy(z *TestRequest) {
	z.Payload = append(z.Payload[:0], x.Payload...)
	z.Hash = append(z.Hash[:0], x.Hash...)
}

func (x *TestResponse) DeepCopy(z *TestResponse) {
	z.Hash = append(z.Hash[:0], x.Hash...)
}

func (x *TestRequestWithString) DeepCopy(z *TestRequestWithString) {
	z.Payload = x.Payload
	z.Hash = x.Hash
}

func (x *TestResponseWithString) DeepCopy(z *TestResponseWithString) {
	z.Hash = append(z.Hash[:0], x.Hash...)
}

func (x *EchoWithDelay) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_EchoWithDelay, x)
}

func (x *TestRequest) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_TestRequest, x)
}

func (x *TestResponse) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_TestResponse, x)
}

func (x *TestRequestWithString) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_TestRequestWithString, x)
}

func (x *TestResponseWithString) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_TestResponseWithString, x)
}

func (x *EchoWithDelay) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *TestRequest) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *TestResponse) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *TestRequestWithString) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *TestResponseWithString) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *EchoWithDelay) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *TestRequest) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *TestResponse) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *TestRequestWithString) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *TestResponseWithString) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}
