package msg

import (
	edge "github.com/ronaksoft/rony/edge"
	registry "github.com/ronaksoft/rony/registry"
	proto "google.golang.org/protobuf/proto"
	sync "sync"
)

const C_TeamGet int64 = 1172720786

type poolTeamGet struct {
	pool sync.Pool
}

func (p *poolTeamGet) Get() *TeamGet {
	x, ok := p.pool.Get().(*TeamGet)
	if !ok {
		return &TeamGet{}
	}
	return x
}

func (p *poolTeamGet) Put(x *TeamGet) {
	x.ID = 0
	p.pool.Put(x)
}

var PoolTeamGet = poolTeamGet{}

const C_TeamAddMember int64 = 3889056091

type poolTeamAddMember struct {
	pool sync.Pool
}

func (p *poolTeamAddMember) Get() *TeamAddMember {
	x, ok := p.pool.Get().(*TeamAddMember)
	if !ok {
		return &TeamAddMember{}
	}
	return x
}

func (p *poolTeamAddMember) Put(x *TeamAddMember) {
	x.TeamID = 0
	x.UserID = 0
	x.Manager = false
	p.pool.Put(x)
}

var PoolTeamAddMember = poolTeamAddMember{}

const C_TeamRemoveMember int64 = 4200364613

type poolTeamRemoveMember struct {
	pool sync.Pool
}

func (p *poolTeamRemoveMember) Get() *TeamRemoveMember {
	x, ok := p.pool.Get().(*TeamRemoveMember)
	if !ok {
		return &TeamRemoveMember{}
	}
	return x
}

func (p *poolTeamRemoveMember) Put(x *TeamRemoveMember) {
	x.TeamID = 0
	x.UserID = 0
	p.pool.Put(x)
}

var PoolTeamRemoveMember = poolTeamRemoveMember{}

const C_TeamPromote int64 = 382328820

type poolTeamPromote struct {
	pool sync.Pool
}

func (p *poolTeamPromote) Get() *TeamPromote {
	x, ok := p.pool.Get().(*TeamPromote)
	if !ok {
		return &TeamPromote{}
	}
	return x
}

func (p *poolTeamPromote) Put(x *TeamPromote) {
	x.TeamID = 0
	x.UserID = 0
	p.pool.Put(x)
}

var PoolTeamPromote = poolTeamPromote{}

const C_TeamDemote int64 = 2331393294

type poolTeamDemote struct {
	pool sync.Pool
}

func (p *poolTeamDemote) Get() *TeamDemote {
	x, ok := p.pool.Get().(*TeamDemote)
	if !ok {
		return &TeamDemote{}
	}
	return x
}

func (p *poolTeamDemote) Put(x *TeamDemote) {
	x.TeamID = 0
	x.UserID = 0
	p.pool.Put(x)
}

var PoolTeamDemote = poolTeamDemote{}

const C_TeamLeave int64 = 1413785879

type poolTeamLeave struct {
	pool sync.Pool
}

func (p *poolTeamLeave) Get() *TeamLeave {
	x, ok := p.pool.Get().(*TeamLeave)
	if !ok {
		return &TeamLeave{}
	}
	return x
}

func (p *poolTeamLeave) Put(x *TeamLeave) {
	x.TeamID = 0
	p.pool.Put(x)
}

var PoolTeamLeave = poolTeamLeave{}

const C_TeamJoin int64 = 1725794017

type poolTeamJoin struct {
	pool sync.Pool
}

func (p *poolTeamJoin) Get() *TeamJoin {
	x, ok := p.pool.Get().(*TeamJoin)
	if !ok {
		return &TeamJoin{}
	}
	return x
}

func (p *poolTeamJoin) Put(x *TeamJoin) {
	x.TeamID = 0
	x.Token = x.Token[:0]
	p.pool.Put(x)
}

var PoolTeamJoin = poolTeamJoin{}

const C_TeamListMembers int64 = 3107323194

type poolTeamListMembers struct {
	pool sync.Pool
}

func (p *poolTeamListMembers) Get() *TeamListMembers {
	x, ok := p.pool.Get().(*TeamListMembers)
	if !ok {
		return &TeamListMembers{}
	}
	return x
}

func (p *poolTeamListMembers) Put(x *TeamListMembers) {
	x.TeamID = 0
	p.pool.Put(x)
}

var PoolTeamListMembers = poolTeamListMembers{}

const C_TeamEdit int64 = 3481894956

type poolTeamEdit struct {
	pool sync.Pool
}

func (p *poolTeamEdit) Get() *TeamEdit {
	x, ok := p.pool.Get().(*TeamEdit)
	if !ok {
		return &TeamEdit{}
	}
	return x
}

func (p *poolTeamEdit) Put(x *TeamEdit) {
	x.TeamID = 0
	x.Name = ""
	p.pool.Put(x)
}

var PoolTeamEdit = poolTeamEdit{}

const C_TeamUploadPhoto int64 = 1595699082

type poolTeamUploadPhoto struct {
	pool sync.Pool
}

func (p *poolTeamUploadPhoto) Get() *TeamUploadPhoto {
	x, ok := p.pool.Get().(*TeamUploadPhoto)
	if !ok {
		return &TeamUploadPhoto{}
	}
	return x
}

func (p *poolTeamUploadPhoto) Put(x *TeamUploadPhoto) {
	x.TeamID = 0
	if x.File != nil {
		PoolInputFile.Put(x.File)
		x.File = nil
	}
	p.pool.Put(x)
}

var PoolTeamUploadPhoto = poolTeamUploadPhoto{}

const C_TeamRemovePhoto int64 = 3388888323

type poolTeamRemovePhoto struct {
	pool sync.Pool
}

func (p *poolTeamRemovePhoto) Get() *TeamRemovePhoto {
	x, ok := p.pool.Get().(*TeamRemovePhoto)
	if !ok {
		return &TeamRemovePhoto{}
	}
	return x
}

func (p *poolTeamRemovePhoto) Put(x *TeamRemovePhoto) {
	x.TeamID = 0
	p.pool.Put(x)
}

var PoolTeamRemovePhoto = poolTeamRemovePhoto{}

const C_TeamMembers int64 = 2208941294

type poolTeamMembers struct {
	pool sync.Pool
}

func (p *poolTeamMembers) Get() *TeamMembers {
	x, ok := p.pool.Get().(*TeamMembers)
	if !ok {
		return &TeamMembers{}
	}
	return x
}

func (p *poolTeamMembers) Put(x *TeamMembers) {
	x.Members = x.Members[:0]
	x.Users = x.Users[:0]
	p.pool.Put(x)
}

var PoolTeamMembers = poolTeamMembers{}

const C_TeamMember int64 = 1965775170

type poolTeamMember struct {
	pool sync.Pool
}

func (p *poolTeamMember) Get() *TeamMember {
	x, ok := p.pool.Get().(*TeamMember)
	if !ok {
		return &TeamMember{}
	}
	return x
}

func (p *poolTeamMember) Put(x *TeamMember) {
	x.UserID = 0
	x.Admin = false
	if x.User != nil {
		PoolUser.Put(x.User)
		x.User = nil
	}
	p.pool.Put(x)
}

var PoolTeamMember = poolTeamMember{}

const C_TeamsMany int64 = 2225718663

type poolTeamsMany struct {
	pool sync.Pool
}

func (p *poolTeamsMany) Get() *TeamsMany {
	x, ok := p.pool.Get().(*TeamsMany)
	if !ok {
		return &TeamsMany{}
	}
	return x
}

func (p *poolTeamsMany) Put(x *TeamsMany) {
	x.Teams = x.Teams[:0]
	x.Users = x.Users[:0]
	x.Empty = false
	p.pool.Put(x)
}

var PoolTeamsMany = poolTeamsMany{}

func init() {
	registry.RegisterConstructor(1172720786, "TeamGet")
	registry.RegisterConstructor(3889056091, "TeamAddMember")
	registry.RegisterConstructor(4200364613, "TeamRemoveMember")
	registry.RegisterConstructor(382328820, "TeamPromote")
	registry.RegisterConstructor(2331393294, "TeamDemote")
	registry.RegisterConstructor(1413785879, "TeamLeave")
	registry.RegisterConstructor(1725794017, "TeamJoin")
	registry.RegisterConstructor(3107323194, "TeamListMembers")
	registry.RegisterConstructor(3481894956, "TeamEdit")
	registry.RegisterConstructor(1595699082, "TeamUploadPhoto")
	registry.RegisterConstructor(3388888323, "TeamRemovePhoto")
	registry.RegisterConstructor(2208941294, "TeamMembers")
	registry.RegisterConstructor(1965775170, "TeamMember")
	registry.RegisterConstructor(2225718663, "TeamsMany")
}

func (x *TeamGet) DeepCopy(z *TeamGet) {
	z.ID = x.ID
}

func (x *TeamAddMember) DeepCopy(z *TeamAddMember) {
	z.TeamID = x.TeamID
	z.UserID = x.UserID
	z.Manager = x.Manager
}

func (x *TeamRemoveMember) DeepCopy(z *TeamRemoveMember) {
	z.TeamID = x.TeamID
	z.UserID = x.UserID
}

func (x *TeamPromote) DeepCopy(z *TeamPromote) {
	z.TeamID = x.TeamID
	z.UserID = x.UserID
}

func (x *TeamDemote) DeepCopy(z *TeamDemote) {
	z.TeamID = x.TeamID
	z.UserID = x.UserID
}

func (x *TeamLeave) DeepCopy(z *TeamLeave) {
	z.TeamID = x.TeamID
}

func (x *TeamJoin) DeepCopy(z *TeamJoin) {
	z.TeamID = x.TeamID
	z.Token = append(z.Token[:0], x.Token...)
}

func (x *TeamListMembers) DeepCopy(z *TeamListMembers) {
	z.TeamID = x.TeamID
}

func (x *TeamEdit) DeepCopy(z *TeamEdit) {
	z.TeamID = x.TeamID
	z.Name = x.Name
}

func (x *TeamUploadPhoto) DeepCopy(z *TeamUploadPhoto) {
	z.TeamID = x.TeamID
	if x.File != nil {
		z.File = PoolInputFile.Get()
		x.File.DeepCopy(z.File)
	}
}

func (x *TeamRemovePhoto) DeepCopy(z *TeamRemovePhoto) {
	z.TeamID = x.TeamID
}

func (x *TeamMembers) DeepCopy(z *TeamMembers) {
	for idx := range x.Members {
		if x.Members[idx] != nil {
			xx := PoolTeamMember.Get()
			x.Members[idx].DeepCopy(xx)
			z.Members = append(z.Members, xx)
		}
	}
	for idx := range x.Users {
		if x.Users[idx] != nil {
			xx := PoolUser.Get()
			x.Users[idx].DeepCopy(xx)
			z.Users = append(z.Users, xx)
		}
	}
}

func (x *TeamMember) DeepCopy(z *TeamMember) {
	z.UserID = x.UserID
	z.Admin = x.Admin
	if x.User != nil {
		z.User = PoolUser.Get()
		x.User.DeepCopy(z.User)
	}
}

func (x *TeamsMany) DeepCopy(z *TeamsMany) {
	for idx := range x.Teams {
		if x.Teams[idx] != nil {
			xx := PoolTeam.Get()
			x.Teams[idx].DeepCopy(xx)
			z.Teams = append(z.Teams, xx)
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

func (x *TeamGet) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_TeamGet, x)
}

func (x *TeamAddMember) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_TeamAddMember, x)
}

func (x *TeamRemoveMember) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_TeamRemoveMember, x)
}

func (x *TeamPromote) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_TeamPromote, x)
}

func (x *TeamDemote) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_TeamDemote, x)
}

func (x *TeamLeave) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_TeamLeave, x)
}

func (x *TeamJoin) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_TeamJoin, x)
}

func (x *TeamListMembers) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_TeamListMembers, x)
}

func (x *TeamEdit) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_TeamEdit, x)
}

func (x *TeamUploadPhoto) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_TeamUploadPhoto, x)
}

func (x *TeamRemovePhoto) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_TeamRemovePhoto, x)
}

func (x *TeamMembers) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_TeamMembers, x)
}

func (x *TeamMember) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_TeamMember, x)
}

func (x *TeamsMany) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_TeamsMany, x)
}

func (x *TeamGet) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *TeamAddMember) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *TeamRemoveMember) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *TeamPromote) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *TeamDemote) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *TeamLeave) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *TeamJoin) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *TeamListMembers) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *TeamEdit) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *TeamUploadPhoto) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *TeamRemovePhoto) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *TeamMembers) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *TeamMember) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *TeamsMany) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *TeamGet) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *TeamAddMember) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *TeamRemoveMember) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *TeamPromote) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *TeamDemote) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *TeamLeave) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *TeamJoin) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *TeamListMembers) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *TeamEdit) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *TeamUploadPhoto) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *TeamRemovePhoto) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *TeamMembers) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *TeamMember) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *TeamsMany) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}
