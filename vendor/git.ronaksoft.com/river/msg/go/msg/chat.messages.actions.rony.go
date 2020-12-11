package msg

import (
	edge "github.com/ronaksoft/rony/edge"
	registry "github.com/ronaksoft/rony/registry"
	proto "google.golang.org/protobuf/proto"
	sync "sync"
)

const C_MessageActionGroupAddUser int64 = 3350337765

type poolMessageActionGroupAddUser struct {
	pool sync.Pool
}

func (p *poolMessageActionGroupAddUser) Get() *MessageActionGroupAddUser {
	x, ok := p.pool.Get().(*MessageActionGroupAddUser)
	if !ok {
		return &MessageActionGroupAddUser{}
	}
	return x
}

func (p *poolMessageActionGroupAddUser) Put(x *MessageActionGroupAddUser) {
	x.UserIDs = x.UserIDs[:0]
	p.pool.Put(x)
}

var PoolMessageActionGroupAddUser = poolMessageActionGroupAddUser{}

const C_MessageActionGroupDeleteUser int64 = 2904432195

type poolMessageActionGroupDeleteUser struct {
	pool sync.Pool
}

func (p *poolMessageActionGroupDeleteUser) Get() *MessageActionGroupDeleteUser {
	x, ok := p.pool.Get().(*MessageActionGroupDeleteUser)
	if !ok {
		return &MessageActionGroupDeleteUser{}
	}
	return x
}

func (p *poolMessageActionGroupDeleteUser) Put(x *MessageActionGroupDeleteUser) {
	x.UserIDs = x.UserIDs[:0]
	p.pool.Put(x)
}

var PoolMessageActionGroupDeleteUser = poolMessageActionGroupDeleteUser{}

const C_MessageActionGroupCreated int64 = 907021784

type poolMessageActionGroupCreated struct {
	pool sync.Pool
}

func (p *poolMessageActionGroupCreated) Get() *MessageActionGroupCreated {
	x, ok := p.pool.Get().(*MessageActionGroupCreated)
	if !ok {
		return &MessageActionGroupCreated{}
	}
	return x
}

func (p *poolMessageActionGroupCreated) Put(x *MessageActionGroupCreated) {
	x.GroupTitle = ""
	x.UserIDs = x.UserIDs[:0]
	p.pool.Put(x)
}

var PoolMessageActionGroupCreated = poolMessageActionGroupCreated{}

const C_MessageActionGroupTitleChanged int64 = 3747325827

type poolMessageActionGroupTitleChanged struct {
	pool sync.Pool
}

func (p *poolMessageActionGroupTitleChanged) Get() *MessageActionGroupTitleChanged {
	x, ok := p.pool.Get().(*MessageActionGroupTitleChanged)
	if !ok {
		return &MessageActionGroupTitleChanged{}
	}
	return x
}

func (p *poolMessageActionGroupTitleChanged) Put(x *MessageActionGroupTitleChanged) {
	x.GroupTitle = ""
	p.pool.Put(x)
}

var PoolMessageActionGroupTitleChanged = poolMessageActionGroupTitleChanged{}

const C_MessageActionGroupPhotoChanged int64 = 1145423234

type poolMessageActionGroupPhotoChanged struct {
	pool sync.Pool
}

func (p *poolMessageActionGroupPhotoChanged) Get() *MessageActionGroupPhotoChanged {
	x, ok := p.pool.Get().(*MessageActionGroupPhotoChanged)
	if !ok {
		return &MessageActionGroupPhotoChanged{}
	}
	return x
}

func (p *poolMessageActionGroupPhotoChanged) Put(x *MessageActionGroupPhotoChanged) {
	if x.Photo != nil {
		PoolGroupPhoto.Put(x.Photo)
		x.Photo = nil
	}
	p.pool.Put(x)
}

var PoolMessageActionGroupPhotoChanged = poolMessageActionGroupPhotoChanged{}

const C_MessageActionClearHistory int64 = 4164590160

type poolMessageActionClearHistory struct {
	pool sync.Pool
}

func (p *poolMessageActionClearHistory) Get() *MessageActionClearHistory {
	x, ok := p.pool.Get().(*MessageActionClearHistory)
	if !ok {
		return &MessageActionClearHistory{}
	}
	return x
}

func (p *poolMessageActionClearHistory) Put(x *MessageActionClearHistory) {
	x.MaxID = 0
	x.Delete = false
	p.pool.Put(x)
}

var PoolMessageActionClearHistory = poolMessageActionClearHistory{}

const C_MessageActionContactRegistered int64 = 3229435742

type poolMessageActionContactRegistered struct {
	pool sync.Pool
}

func (p *poolMessageActionContactRegistered) Get() *MessageActionContactRegistered {
	x, ok := p.pool.Get().(*MessageActionContactRegistered)
	if !ok {
		return &MessageActionContactRegistered{}
	}
	return x
}

func (p *poolMessageActionContactRegistered) Put(x *MessageActionContactRegistered) {
	p.pool.Put(x)
}

var PoolMessageActionContactRegistered = poolMessageActionContactRegistered{}

const C_MessageActionScreenShotTaken int64 = 2021478678

type poolMessageActionScreenShotTaken struct {
	pool sync.Pool
}

func (p *poolMessageActionScreenShotTaken) Get() *MessageActionScreenShotTaken {
	x, ok := p.pool.Get().(*MessageActionScreenShotTaken)
	if !ok {
		return &MessageActionScreenShotTaken{}
	}
	return x
}

func (p *poolMessageActionScreenShotTaken) Put(x *MessageActionScreenShotTaken) {
	x.MinID = 0
	x.MaxID = 0
	p.pool.Put(x)
}

var PoolMessageActionScreenShotTaken = poolMessageActionScreenShotTaken{}

const C_MessageActionThreadClosed int64 = 3807512538

type poolMessageActionThreadClosed struct {
	pool sync.Pool
}

func (p *poolMessageActionThreadClosed) Get() *MessageActionThreadClosed {
	x, ok := p.pool.Get().(*MessageActionThreadClosed)
	if !ok {
		return &MessageActionThreadClosed{}
	}
	return x
}

func (p *poolMessageActionThreadClosed) Put(x *MessageActionThreadClosed) {
	x.ThreadID = 0
	p.pool.Put(x)
}

var PoolMessageActionThreadClosed = poolMessageActionThreadClosed{}

func init() {
	registry.RegisterConstructor(3350337765, "msg.MessageActionGroupAddUser")
	registry.RegisterConstructor(2904432195, "msg.MessageActionGroupDeleteUser")
	registry.RegisterConstructor(907021784, "msg.MessageActionGroupCreated")
	registry.RegisterConstructor(3747325827, "msg.MessageActionGroupTitleChanged")
	registry.RegisterConstructor(1145423234, "msg.MessageActionGroupPhotoChanged")
	registry.RegisterConstructor(4164590160, "msg.MessageActionClearHistory")
	registry.RegisterConstructor(3229435742, "msg.MessageActionContactRegistered")
	registry.RegisterConstructor(2021478678, "msg.MessageActionScreenShotTaken")
	registry.RegisterConstructor(3807512538, "msg.MessageActionThreadClosed")
}

func (x *MessageActionGroupAddUser) DeepCopy(z *MessageActionGroupAddUser) {
	z.UserIDs = append(z.UserIDs[:0], x.UserIDs...)
}

func (x *MessageActionGroupDeleteUser) DeepCopy(z *MessageActionGroupDeleteUser) {
	z.UserIDs = append(z.UserIDs[:0], x.UserIDs...)
}

func (x *MessageActionGroupCreated) DeepCopy(z *MessageActionGroupCreated) {
	z.GroupTitle = x.GroupTitle
	z.UserIDs = append(z.UserIDs[:0], x.UserIDs...)
}

func (x *MessageActionGroupTitleChanged) DeepCopy(z *MessageActionGroupTitleChanged) {
	z.GroupTitle = x.GroupTitle
}

func (x *MessageActionGroupPhotoChanged) DeepCopy(z *MessageActionGroupPhotoChanged) {
	if x.Photo != nil {
		z.Photo = PoolGroupPhoto.Get()
		x.Photo.DeepCopy(z.Photo)
	}
}

func (x *MessageActionClearHistory) DeepCopy(z *MessageActionClearHistory) {
	z.MaxID = x.MaxID
	z.Delete = x.Delete
}

func (x *MessageActionContactRegistered) DeepCopy(z *MessageActionContactRegistered) {
}

func (x *MessageActionScreenShotTaken) DeepCopy(z *MessageActionScreenShotTaken) {
	z.MinID = x.MinID
	z.MaxID = x.MaxID
}

func (x *MessageActionThreadClosed) DeepCopy(z *MessageActionThreadClosed) {
	z.ThreadID = x.ThreadID
}

func (x *MessageActionGroupAddUser) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessageActionGroupAddUser, x)
}

func (x *MessageActionGroupDeleteUser) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessageActionGroupDeleteUser, x)
}

func (x *MessageActionGroupCreated) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessageActionGroupCreated, x)
}

func (x *MessageActionGroupTitleChanged) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessageActionGroupTitleChanged, x)
}

func (x *MessageActionGroupPhotoChanged) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessageActionGroupPhotoChanged, x)
}

func (x *MessageActionClearHistory) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessageActionClearHistory, x)
}

func (x *MessageActionContactRegistered) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessageActionContactRegistered, x)
}

func (x *MessageActionScreenShotTaken) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessageActionScreenShotTaken, x)
}

func (x *MessageActionThreadClosed) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_MessageActionThreadClosed, x)
}

func (x *MessageActionGroupAddUser) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessageActionGroupDeleteUser) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessageActionGroupCreated) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessageActionGroupTitleChanged) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessageActionGroupPhotoChanged) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessageActionClearHistory) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessageActionContactRegistered) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessageActionScreenShotTaken) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessageActionThreadClosed) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *MessageActionGroupAddUser) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessageActionGroupDeleteUser) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessageActionGroupCreated) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessageActionGroupTitleChanged) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessageActionGroupPhotoChanged) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessageActionClearHistory) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessageActionContactRegistered) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessageActionScreenShotTaken) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *MessageActionThreadClosed) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}
