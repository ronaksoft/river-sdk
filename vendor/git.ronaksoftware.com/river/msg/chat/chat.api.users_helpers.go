// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: chat.api.users.proto

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

const C_UsersGet int64 = 1039301579

type poolUsersGet struct {
	pool sync.Pool
}

func (p *poolUsersGet) Get() *UsersGet {
	x, ok := p.pool.Get().(*UsersGet)
	if !ok {
		return &UsersGet{}
	}
	x.Users = x.Users[:0]
	return x
}

func (p *poolUsersGet) Put(x *UsersGet) {
	p.pool.Put(x)
}

var PoolUsersGet = poolUsersGet{}

func ResultUsersGet(out *MessageEnvelope, res *UsersGet) {
	out.Constructor = C_UsersGet
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UsersGetFull int64 = 3343342086

type poolUsersGetFull struct {
	pool sync.Pool
}

func (p *poolUsersGetFull) Get() *UsersGetFull {
	x, ok := p.pool.Get().(*UsersGetFull)
	if !ok {
		return &UsersGetFull{}
	}
	x.Users = x.Users[:0]
	return x
}

func (p *poolUsersGetFull) Put(x *UsersGetFull) {
	p.pool.Put(x)
}

var PoolUsersGetFull = poolUsersGetFull{}

func ResultUsersGetFull(out *MessageEnvelope, res *UsersGetFull) {
	out.Constructor = C_UsersGetFull
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

const C_UsersMany int64 = 801733941

type poolUsersMany struct {
	pool sync.Pool
}

func (p *poolUsersMany) Get() *UsersMany {
	x, ok := p.pool.Get().(*UsersMany)
	if !ok {
		return &UsersMany{}
	}
	x.Users = x.Users[:0]
	return x
}

func (p *poolUsersMany) Put(x *UsersMany) {
	p.pool.Put(x)
}

var PoolUsersMany = poolUsersMany{}

func ResultUsersMany(out *MessageEnvelope, res *UsersMany) {
	out.Constructor = C_UsersMany
	pbytes.Put(out.Message)
	out.Message = pbytes.GetLen(res.Size())
	res.MarshalTo(out.Message)
}

func init() {
	ConstructorNames[1039301579] = "UsersGet"
	ConstructorNames[3343342086] = "UsersGetFull"
	ConstructorNames[801733941] = "UsersMany"
}
