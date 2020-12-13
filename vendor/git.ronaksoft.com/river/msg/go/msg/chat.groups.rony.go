package msg

import (
	edge "github.com/ronaksoft/rony/edge"
	registry "github.com/ronaksoft/rony/registry"
	proto "google.golang.org/protobuf/proto"
	sync "sync"
)

const C_GroupsCreate int64 = 1271969037

type poolGroupsCreate struct {
	pool sync.Pool
}

func (p *poolGroupsCreate) Get() *GroupsCreate {
	x, ok := p.pool.Get().(*GroupsCreate)
	if !ok {
		return &GroupsCreate{}
	}
	return x
}

func (p *poolGroupsCreate) Put(x *GroupsCreate) {
	x.Users = x.Users[:0]
	x.Title = ""
	x.RandomID = 0
	p.pool.Put(x)
}

var PoolGroupsCreate = poolGroupsCreate{}

const C_GroupsAddUser int64 = 394654713

type poolGroupsAddUser struct {
	pool sync.Pool
}

func (p *poolGroupsAddUser) Get() *GroupsAddUser {
	x, ok := p.pool.Get().(*GroupsAddUser)
	if !ok {
		return &GroupsAddUser{}
	}
	return x
}

func (p *poolGroupsAddUser) Put(x *GroupsAddUser) {
	x.GroupID = 0
	if x.User != nil {
		PoolInputUser.Put(x.User)
		x.User = nil
	}
	x.ForwardLimit = 0
	p.pool.Put(x)
}

var PoolGroupsAddUser = poolGroupsAddUser{}

const C_GroupsEditTitle int64 = 2582813461

type poolGroupsEditTitle struct {
	pool sync.Pool
}

func (p *poolGroupsEditTitle) Get() *GroupsEditTitle {
	x, ok := p.pool.Get().(*GroupsEditTitle)
	if !ok {
		return &GroupsEditTitle{}
	}
	return x
}

func (p *poolGroupsEditTitle) Put(x *GroupsEditTitle) {
	x.GroupID = 0
	x.Title = ""
	p.pool.Put(x)
}

var PoolGroupsEditTitle = poolGroupsEditTitle{}

const C_GroupsDeleteUser int64 = 3172322223

type poolGroupsDeleteUser struct {
	pool sync.Pool
}

func (p *poolGroupsDeleteUser) Get() *GroupsDeleteUser {
	x, ok := p.pool.Get().(*GroupsDeleteUser)
	if !ok {
		return &GroupsDeleteUser{}
	}
	return x
}

func (p *poolGroupsDeleteUser) Put(x *GroupsDeleteUser) {
	x.GroupID = 0
	if x.User != nil {
		PoolInputUser.Put(x.User)
		x.User = nil
	}
	p.pool.Put(x)
}

var PoolGroupsDeleteUser = poolGroupsDeleteUser{}

const C_GroupsGetFull int64 = 2986704909

type poolGroupsGetFull struct {
	pool sync.Pool
}

func (p *poolGroupsGetFull) Get() *GroupsGetFull {
	x, ok := p.pool.Get().(*GroupsGetFull)
	if !ok {
		return &GroupsGetFull{}
	}
	return x
}

func (p *poolGroupsGetFull) Put(x *GroupsGetFull) {
	x.GroupID = 0
	p.pool.Put(x)
}

var PoolGroupsGetFull = poolGroupsGetFull{}

const C_GroupsToggleAdmins int64 = 1581076909

type poolGroupsToggleAdmins struct {
	pool sync.Pool
}

func (p *poolGroupsToggleAdmins) Get() *GroupsToggleAdmins {
	x, ok := p.pool.Get().(*GroupsToggleAdmins)
	if !ok {
		return &GroupsToggleAdmins{}
	}
	return x
}

func (p *poolGroupsToggleAdmins) Put(x *GroupsToggleAdmins) {
	x.GroupID = 0
	x.AdminEnabled = false
	p.pool.Put(x)
}

var PoolGroupsToggleAdmins = poolGroupsToggleAdmins{}

const C_GroupsUpdateAdmin int64 = 1345991011

type poolGroupsUpdateAdmin struct {
	pool sync.Pool
}

func (p *poolGroupsUpdateAdmin) Get() *GroupsUpdateAdmin {
	x, ok := p.pool.Get().(*GroupsUpdateAdmin)
	if !ok {
		return &GroupsUpdateAdmin{}
	}
	return x
}

func (p *poolGroupsUpdateAdmin) Put(x *GroupsUpdateAdmin) {
	x.GroupID = 0
	if x.User != nil {
		PoolInputUser.Put(x.User)
		x.User = nil
	}
	x.Admin = false
	p.pool.Put(x)
}

var PoolGroupsUpdateAdmin = poolGroupsUpdateAdmin{}

const C_GroupsUploadPhoto int64 = 2624284907

type poolGroupsUploadPhoto struct {
	pool sync.Pool
}

func (p *poolGroupsUploadPhoto) Get() *GroupsUploadPhoto {
	x, ok := p.pool.Get().(*GroupsUploadPhoto)
	if !ok {
		return &GroupsUploadPhoto{}
	}
	return x
}

func (p *poolGroupsUploadPhoto) Put(x *GroupsUploadPhoto) {
	x.GroupID = 0
	if x.File != nil {
		PoolInputFile.Put(x.File)
		x.File = nil
	}
	x.ReturnObject = false
	p.pool.Put(x)
}

var PoolGroupsUploadPhoto = poolGroupsUploadPhoto{}

const C_GroupsRemovePhoto int64 = 176771682

type poolGroupsRemovePhoto struct {
	pool sync.Pool
}

func (p *poolGroupsRemovePhoto) Get() *GroupsRemovePhoto {
	x, ok := p.pool.Get().(*GroupsRemovePhoto)
	if !ok {
		return &GroupsRemovePhoto{}
	}
	return x
}

func (p *poolGroupsRemovePhoto) Put(x *GroupsRemovePhoto) {
	x.GroupID = 0
	x.PhotoID = 0
	p.pool.Put(x)
}

var PoolGroupsRemovePhoto = poolGroupsRemovePhoto{}

const C_GroupsUpdatePhoto int64 = 3431184397

type poolGroupsUpdatePhoto struct {
	pool sync.Pool
}

func (p *poolGroupsUpdatePhoto) Get() *GroupsUpdatePhoto {
	x, ok := p.pool.Get().(*GroupsUpdatePhoto)
	if !ok {
		return &GroupsUpdatePhoto{}
	}
	return x
}

func (p *poolGroupsUpdatePhoto) Put(x *GroupsUpdatePhoto) {
	x.PhotoID = 0
	x.GroupID = 0
	p.pool.Put(x)
}

var PoolGroupsUpdatePhoto = poolGroupsUpdatePhoto{}

const C_GroupsGetReadHistoryStats int64 = 719309439

type poolGroupsGetReadHistoryStats struct {
	pool sync.Pool
}

func (p *poolGroupsGetReadHistoryStats) Get() *GroupsGetReadHistoryStats {
	x, ok := p.pool.Get().(*GroupsGetReadHistoryStats)
	if !ok {
		return &GroupsGetReadHistoryStats{}
	}
	return x
}

func (p *poolGroupsGetReadHistoryStats) Put(x *GroupsGetReadHistoryStats) {
	x.GroupID = 0
	p.pool.Put(x)
}

var PoolGroupsGetReadHistoryStats = poolGroupsGetReadHistoryStats{}

const C_GroupsHistoryStats int64 = 1080267574

type poolGroupsHistoryStats struct {
	pool sync.Pool
}

func (p *poolGroupsHistoryStats) Get() *GroupsHistoryStats {
	x, ok := p.pool.Get().(*GroupsHistoryStats)
	if !ok {
		return &GroupsHistoryStats{}
	}
	return x
}

func (p *poolGroupsHistoryStats) Put(x *GroupsHistoryStats) {
	x.Stats = x.Stats[:0]
	x.Users = x.Users[:0]
	x.Empty = false
	p.pool.Put(x)
}

var PoolGroupsHistoryStats = poolGroupsHistoryStats{}

const C_ReadHistoryStat int64 = 3486960061

type poolReadHistoryStat struct {
	pool sync.Pool
}

func (p *poolReadHistoryStat) Get() *ReadHistoryStat {
	x, ok := p.pool.Get().(*ReadHistoryStat)
	if !ok {
		return &ReadHistoryStat{}
	}
	return x
}

func (p *poolReadHistoryStat) Put(x *ReadHistoryStat) {
	x.UserID = 0
	x.MessageID = 0
	p.pool.Put(x)
}

var PoolReadHistoryStat = poolReadHistoryStat{}

func init() {
	registry.RegisterConstructor(1271969037, "GroupsCreate")
	registry.RegisterConstructor(394654713, "GroupsAddUser")
	registry.RegisterConstructor(2582813461, "GroupsEditTitle")
	registry.RegisterConstructor(3172322223, "GroupsDeleteUser")
	registry.RegisterConstructor(2986704909, "GroupsGetFull")
	registry.RegisterConstructor(1581076909, "GroupsToggleAdmins")
	registry.RegisterConstructor(1345991011, "GroupsUpdateAdmin")
	registry.RegisterConstructor(2624284907, "GroupsUploadPhoto")
	registry.RegisterConstructor(176771682, "GroupsRemovePhoto")
	registry.RegisterConstructor(3431184397, "GroupsUpdatePhoto")
	registry.RegisterConstructor(719309439, "GroupsGetReadHistoryStats")
	registry.RegisterConstructor(1080267574, "GroupsHistoryStats")
	registry.RegisterConstructor(3486960061, "ReadHistoryStat")
}

func (x *GroupsCreate) DeepCopy(z *GroupsCreate) {
	for idx := range x.Users {
		if x.Users[idx] != nil {
			xx := PoolInputUser.Get()
			x.Users[idx].DeepCopy(xx)
			z.Users = append(z.Users, xx)
		}
	}
	z.Title = x.Title
	z.RandomID = x.RandomID
}

func (x *GroupsAddUser) DeepCopy(z *GroupsAddUser) {
	z.GroupID = x.GroupID
	if x.User != nil {
		z.User = PoolInputUser.Get()
		x.User.DeepCopy(z.User)
	}
	z.ForwardLimit = x.ForwardLimit
}

func (x *GroupsEditTitle) DeepCopy(z *GroupsEditTitle) {
	z.GroupID = x.GroupID
	z.Title = x.Title
}

func (x *GroupsDeleteUser) DeepCopy(z *GroupsDeleteUser) {
	z.GroupID = x.GroupID
	if x.User != nil {
		z.User = PoolInputUser.Get()
		x.User.DeepCopy(z.User)
	}
}

func (x *GroupsGetFull) DeepCopy(z *GroupsGetFull) {
	z.GroupID = x.GroupID
}

func (x *GroupsToggleAdmins) DeepCopy(z *GroupsToggleAdmins) {
	z.GroupID = x.GroupID
	z.AdminEnabled = x.AdminEnabled
}

func (x *GroupsUpdateAdmin) DeepCopy(z *GroupsUpdateAdmin) {
	z.GroupID = x.GroupID
	if x.User != nil {
		z.User = PoolInputUser.Get()
		x.User.DeepCopy(z.User)
	}
	z.Admin = x.Admin
}

func (x *GroupsUploadPhoto) DeepCopy(z *GroupsUploadPhoto) {
	z.GroupID = x.GroupID
	if x.File != nil {
		z.File = PoolInputFile.Get()
		x.File.DeepCopy(z.File)
	}
	z.ReturnObject = x.ReturnObject
}

func (x *GroupsRemovePhoto) DeepCopy(z *GroupsRemovePhoto) {
	z.GroupID = x.GroupID
	z.PhotoID = x.PhotoID
}

func (x *GroupsUpdatePhoto) DeepCopy(z *GroupsUpdatePhoto) {
	z.PhotoID = x.PhotoID
	z.GroupID = x.GroupID
}

func (x *GroupsGetReadHistoryStats) DeepCopy(z *GroupsGetReadHistoryStats) {
	z.GroupID = x.GroupID
}

func (x *GroupsHistoryStats) DeepCopy(z *GroupsHistoryStats) {
	for idx := range x.Stats {
		if x.Stats[idx] != nil {
			xx := PoolReadHistoryStat.Get()
			x.Stats[idx].DeepCopy(xx)
			z.Stats = append(z.Stats, xx)
		}
	}
	for idx := range x.Users {
		if x.Users[idx] != nil {
			xx := PoolUser.Get()
			x.Users[idx].DeepCopy(xx)
			z.Users = append(z.Users, xx)
		}
	}
	z.Empty = x.Empty
}

func (x *ReadHistoryStat) DeepCopy(z *ReadHistoryStat) {
	z.UserID = x.UserID
	z.MessageID = x.MessageID
}

func (x *GroupsCreate) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_GroupsCreate, x)
}

func (x *GroupsAddUser) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_GroupsAddUser, x)
}

func (x *GroupsEditTitle) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_GroupsEditTitle, x)
}

func (x *GroupsDeleteUser) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_GroupsDeleteUser, x)
}

func (x *GroupsGetFull) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_GroupsGetFull, x)
}

func (x *GroupsToggleAdmins) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_GroupsToggleAdmins, x)
}

func (x *GroupsUpdateAdmin) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_GroupsUpdateAdmin, x)
}

func (x *GroupsUploadPhoto) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_GroupsUploadPhoto, x)
}

func (x *GroupsRemovePhoto) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_GroupsRemovePhoto, x)
}

func (x *GroupsUpdatePhoto) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_GroupsUpdatePhoto, x)
}

func (x *GroupsGetReadHistoryStats) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_GroupsGetReadHistoryStats, x)
}

func (x *GroupsHistoryStats) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_GroupsHistoryStats, x)
}

func (x *ReadHistoryStat) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_ReadHistoryStat, x)
}

func (x *GroupsCreate) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *GroupsAddUser) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *GroupsEditTitle) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *GroupsDeleteUser) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *GroupsGetFull) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *GroupsToggleAdmins) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *GroupsUpdateAdmin) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *GroupsUploadPhoto) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *GroupsRemovePhoto) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *GroupsUpdatePhoto) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *GroupsGetReadHistoryStats) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *GroupsHistoryStats) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *ReadHistoryStat) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *GroupsCreate) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *GroupsAddUser) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *GroupsEditTitle) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *GroupsDeleteUser) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *GroupsGetFull) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *GroupsToggleAdmins) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *GroupsUpdateAdmin) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *GroupsUploadPhoto) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *GroupsRemovePhoto) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *GroupsUpdatePhoto) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *GroupsGetReadHistoryStats) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *GroupsHistoryStats) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *ReadHistoryStat) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}
