// Code generated by Rony's protoc plugin; DO NOT EDIT.
// ProtoC ver. v3.15.8
// Rony ver. v0.12.22
// Source: dev.proto

package msg

import (
	bytes "bytes"
	edge "github.com/ronaksoft/rony/edge"
	pools "github.com/ronaksoft/rony/pools"
	registry "github.com/ronaksoft/rony/registry"
	protojson "google.golang.org/protobuf/encoding/protojson"
	proto "google.golang.org/protobuf/proto"
	sync "sync"
)

var _ = pools.Imported

const C_EchoWithDelay int64 = 2861516000

type poolEchoWithDelay struct {
	pool sync.Pool
}

func (p *poolEchoWithDelay) Get() *EchoWithDelay {
	x, ok := p.pool.Get().(*EchoWithDelay)
	if !ok {
		x = &EchoWithDelay{}
	}

	return x
}

func (p *poolEchoWithDelay) Put(x *EchoWithDelay) {
	if x == nil {
		return
	}

	x.DelayInSeconds = 0

	p.pool.Put(x)
}

var PoolEchoWithDelay = poolEchoWithDelay{}

func (x *EchoWithDelay) DeepCopy(z *EchoWithDelay) {
	z.DelayInSeconds = x.DelayInSeconds
}

func (x *EchoWithDelay) Clone() *EchoWithDelay {
	z := &EchoWithDelay{}
	x.DeepCopy(z)
	return z
}

func (x *EchoWithDelay) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{Merge: true}.Unmarshal(b, x)
}

func (x *EchoWithDelay) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *EchoWithDelay) UnmarshalJSON(b []byte) error {
	return protojson.Unmarshal(b, x)
}

func (x *EchoWithDelay) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(x)
}

func (x *EchoWithDelay) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_EchoWithDelay, x)
}

const C_TestRequest int64 = 475847033

type poolTestRequest struct {
	pool sync.Pool
}

func (p *poolTestRequest) Get() *TestRequest {
	x, ok := p.pool.Get().(*TestRequest)
	if !ok {
		x = &TestRequest{}
	}

	return x
}

func (p *poolTestRequest) Put(x *TestRequest) {
	if x == nil {
		return
	}

	x.Payload = x.Payload[:0]
	x.Hash = x.Hash[:0]

	p.pool.Put(x)
}

var PoolTestRequest = poolTestRequest{}

func (x *TestRequest) DeepCopy(z *TestRequest) {
	z.Payload = append(z.Payload[:0], x.Payload...)
	z.Hash = append(z.Hash[:0], x.Hash...)
}

func (x *TestRequest) Clone() *TestRequest {
	z := &TestRequest{}
	x.DeepCopy(z)
	return z
}

func (x *TestRequest) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{Merge: true}.Unmarshal(b, x)
}

func (x *TestRequest) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *TestRequest) UnmarshalJSON(b []byte) error {
	return protojson.Unmarshal(b, x)
}

func (x *TestRequest) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(x)
}

func (x *TestRequest) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_TestRequest, x)
}

const C_TestResponse int64 = 1999996896

type poolTestResponse struct {
	pool sync.Pool
}

func (p *poolTestResponse) Get() *TestResponse {
	x, ok := p.pool.Get().(*TestResponse)
	if !ok {
		x = &TestResponse{}
	}

	return x
}

func (p *poolTestResponse) Put(x *TestResponse) {
	if x == nil {
		return
	}

	x.Hash = x.Hash[:0]

	p.pool.Put(x)
}

var PoolTestResponse = poolTestResponse{}

func (x *TestResponse) DeepCopy(z *TestResponse) {
	z.Hash = append(z.Hash[:0], x.Hash...)
}

func (x *TestResponse) Clone() *TestResponse {
	z := &TestResponse{}
	x.DeepCopy(z)
	return z
}

func (x *TestResponse) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{Merge: true}.Unmarshal(b, x)
}

func (x *TestResponse) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *TestResponse) UnmarshalJSON(b []byte) error {
	return protojson.Unmarshal(b, x)
}

func (x *TestResponse) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(x)
}

func (x *TestResponse) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_TestResponse, x)
}

const C_TestRequestWithString int64 = 3760062575

type poolTestRequestWithString struct {
	pool sync.Pool
}

func (p *poolTestRequestWithString) Get() *TestRequestWithString {
	x, ok := p.pool.Get().(*TestRequestWithString)
	if !ok {
		x = &TestRequestWithString{}
	}

	return x
}

func (p *poolTestRequestWithString) Put(x *TestRequestWithString) {
	if x == nil {
		return
	}

	x.Payload = ""
	x.Hash = ""

	p.pool.Put(x)
}

var PoolTestRequestWithString = poolTestRequestWithString{}

func (x *TestRequestWithString) DeepCopy(z *TestRequestWithString) {
	z.Payload = x.Payload
	z.Hash = x.Hash
}

func (x *TestRequestWithString) Clone() *TestRequestWithString {
	z := &TestRequestWithString{}
	x.DeepCopy(z)
	return z
}

func (x *TestRequestWithString) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{Merge: true}.Unmarshal(b, x)
}

func (x *TestRequestWithString) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *TestRequestWithString) UnmarshalJSON(b []byte) error {
	return protojson.Unmarshal(b, x)
}

func (x *TestRequestWithString) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(x)
}

func (x *TestRequestWithString) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_TestRequestWithString, x)
}

const C_TestResponseWithString int64 = 556112423

type poolTestResponseWithString struct {
	pool sync.Pool
}

func (p *poolTestResponseWithString) Get() *TestResponseWithString {
	x, ok := p.pool.Get().(*TestResponseWithString)
	if !ok {
		x = &TestResponseWithString{}
	}

	return x
}

func (p *poolTestResponseWithString) Put(x *TestResponseWithString) {
	if x == nil {
		return
	}

	x.Hash = x.Hash[:0]

	p.pool.Put(x)
}

var PoolTestResponseWithString = poolTestResponseWithString{}

func (x *TestResponseWithString) DeepCopy(z *TestResponseWithString) {
	z.Hash = append(z.Hash[:0], x.Hash...)
}

func (x *TestResponseWithString) Clone() *TestResponseWithString {
	z := &TestResponseWithString{}
	x.DeepCopy(z)
	return z
}

func (x *TestResponseWithString) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{Merge: true}.Unmarshal(b, x)
}

func (x *TestResponseWithString) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *TestResponseWithString) UnmarshalJSON(b []byte) error {
	return protojson.Unmarshal(b, x)
}

func (x *TestResponseWithString) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(x)
}

func (x *TestResponseWithString) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_TestResponseWithString, x)
}

func init() {
	registry.RegisterConstructor(2861516000, "EchoWithDelay")
	registry.RegisterConstructor(475847033, "TestRequest")
	registry.RegisterConstructor(1999996896, "TestResponse")
	registry.RegisterConstructor(3760062575, "TestRequestWithString")
	registry.RegisterConstructor(556112423, "TestResponseWithString")
}

var _ = bytes.MinRead
