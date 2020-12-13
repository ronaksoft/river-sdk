package msg

import (
	edge "github.com/ronaksoft/rony/edge"
	registry "github.com/ronaksoft/rony/registry"
	proto "google.golang.org/protobuf/proto"
	sync "sync"
)

const C_AdminBroadcastMessage int64 = 3981417409

type poolAdminBroadcastMessage struct {
	pool sync.Pool
}

func (p *poolAdminBroadcastMessage) Get() *AdminBroadcastMessage {
	x, ok := p.pool.Get().(*AdminBroadcastMessage)
	if !ok {
		return &AdminBroadcastMessage{}
	}
	return x
}

func (p *poolAdminBroadcastMessage) Put(x *AdminBroadcastMessage) {
	x.Body = ""
	x.ReceiverIDs = x.ReceiverIDs[:0]
	x.Entities = x.Entities[:0]
	x.MediaType = 0
	x.MediaData = x.MediaData[:0]
	p.pool.Put(x)
}

var PoolAdminBroadcastMessage = poolAdminBroadcastMessage{}

const C_AdminSetWelcomeMessage int64 = 1149591874

type poolAdminSetWelcomeMessage struct {
	pool sync.Pool
}

func (p *poolAdminSetWelcomeMessage) Get() *AdminSetWelcomeMessage {
	x, ok := p.pool.Get().(*AdminSetWelcomeMessage)
	if !ok {
		return &AdminSetWelcomeMessage{}
	}
	return x
}

func (p *poolAdminSetWelcomeMessage) Put(x *AdminSetWelcomeMessage) {
	x.AccessToken = ""
	x.Lang = ""
	x.Template = ""
	p.pool.Put(x)
}

var PoolAdminSetWelcomeMessage = poolAdminSetWelcomeMessage{}

const C_AdminGetWelcomeMessages int64 = 2794709448

type poolAdminGetWelcomeMessages struct {
	pool sync.Pool
}

func (p *poolAdminGetWelcomeMessages) Get() *AdminGetWelcomeMessages {
	x, ok := p.pool.Get().(*AdminGetWelcomeMessages)
	if !ok {
		return &AdminGetWelcomeMessages{}
	}
	return x
}

func (p *poolAdminGetWelcomeMessages) Put(x *AdminGetWelcomeMessages) {
	x.AccessToken = ""
	p.pool.Put(x)
}

var PoolAdminGetWelcomeMessages = poolAdminGetWelcomeMessages{}

const C_AdminDeleteWelcomeMessage int64 = 3940015991

type poolAdminDeleteWelcomeMessage struct {
	pool sync.Pool
}

func (p *poolAdminDeleteWelcomeMessage) Get() *AdminDeleteWelcomeMessage {
	x, ok := p.pool.Get().(*AdminDeleteWelcomeMessage)
	if !ok {
		return &AdminDeleteWelcomeMessage{}
	}
	return x
}

func (p *poolAdminDeleteWelcomeMessage) Put(x *AdminDeleteWelcomeMessage) {
	x.AccessToken = ""
	x.Lang = ""
	p.pool.Put(x)
}

var PoolAdminDeleteWelcomeMessage = poolAdminDeleteWelcomeMessage{}

const C_AdminSetPushProvider int64 = 1758606947

type poolAdminSetPushProvider struct {
	pool sync.Pool
}

func (p *poolAdminSetPushProvider) Get() *AdminSetPushProvider {
	x, ok := p.pool.Get().(*AdminSetPushProvider)
	if !ok {
		return &AdminSetPushProvider{}
	}
	return x
}

func (p *poolAdminSetPushProvider) Put(x *AdminSetPushProvider) {
	x.AccessToken = ""
	if x.Provider != nil {
		PoolPushProvider.Put(x.Provider)
		x.Provider = nil
	}
	p.pool.Put(x)
}

var PoolAdminSetPushProvider = poolAdminSetPushProvider{}

const C_AdminGetPushProviders int64 = 4257963974

type poolAdminGetPushProviders struct {
	pool sync.Pool
}

func (p *poolAdminGetPushProviders) Get() *AdminGetPushProviders {
	x, ok := p.pool.Get().(*AdminGetPushProviders)
	if !ok {
		return &AdminGetPushProviders{}
	}
	return x
}

func (p *poolAdminGetPushProviders) Put(x *AdminGetPushProviders) {
	x.AccessToken = ""
	p.pool.Put(x)
}

var PoolAdminGetPushProviders = poolAdminGetPushProviders{}

const C_AdminDeletePushProvider int64 = 1864898932

type poolAdminDeletePushProvider struct {
	pool sync.Pool
}

func (p *poolAdminDeletePushProvider) Get() *AdminDeletePushProvider {
	x, ok := p.pool.Get().(*AdminDeletePushProvider)
	if !ok {
		return &AdminDeletePushProvider{}
	}
	return x
}

func (p *poolAdminDeletePushProvider) Put(x *AdminDeletePushProvider) {
	x.AccessToken = ""
	x.Name = ""
	x.Type = 0
	p.pool.Put(x)
}

var PoolAdminDeletePushProvider = poolAdminDeletePushProvider{}

const C_AdminSetVersion int64 = 1311023404

type poolAdminSetVersion struct {
	pool sync.Pool
}

func (p *poolAdminSetVersion) Get() *AdminSetVersion {
	x, ok := p.pool.Get().(*AdminSetVersion)
	if !ok {
		return &AdminSetVersion{}
	}
	return x
}

func (p *poolAdminSetVersion) Put(x *AdminSetVersion) {
	x.AccessToken = ""
	if x.Version != nil {
		PoolVersion.Put(x.Version)
		x.Version = nil
	}
	p.pool.Put(x)
}

var PoolAdminSetVersion = poolAdminSetVersion{}

const C_AdminGetVersions int64 = 934752256

type poolAdminGetVersions struct {
	pool sync.Pool
}

func (p *poolAdminGetVersions) Get() *AdminGetVersions {
	x, ok := p.pool.Get().(*AdminGetVersions)
	if !ok {
		return &AdminGetVersions{}
	}
	return x
}

func (p *poolAdminGetVersions) Put(x *AdminGetVersions) {
	x.AccessToken = ""
	p.pool.Put(x)
}

var PoolAdminGetVersions = poolAdminGetVersions{}

const C_AdminSetToken int64 = 2892519162

type poolAdminSetToken struct {
	pool sync.Pool
}

func (p *poolAdminSetToken) Get() *AdminSetToken {
	x, ok := p.pool.Get().(*AdminSetToken)
	if !ok {
		return &AdminSetToken{}
	}
	return x
}

func (p *poolAdminSetToken) Put(x *AdminSetToken) {
	x.AccessToken = ""
	x.Privilege = 0
	p.pool.Put(x)
}

var PoolAdminSetToken = poolAdminSetToken{}

const C_AdminDeleteToken int64 = 3154441897

type poolAdminDeleteToken struct {
	pool sync.Pool
}

func (p *poolAdminDeleteToken) Get() *AdminDeleteToken {
	x, ok := p.pool.Get().(*AdminDeleteToken)
	if !ok {
		return &AdminDeleteToken{}
	}
	return x
}

func (p *poolAdminDeleteToken) Put(x *AdminDeleteToken) {
	x.AccessToken = ""
	x.DeletedToken = ""
	p.pool.Put(x)
}

var PoolAdminDeleteToken = poolAdminDeleteToken{}

const C_AdminReserveUsername int64 = 1947723452

type poolAdminReserveUsername struct {
	pool sync.Pool
}

func (p *poolAdminReserveUsername) Get() *AdminReserveUsername {
	x, ok := p.pool.Get().(*AdminReserveUsername)
	if !ok {
		return &AdminReserveUsername{}
	}
	return x
}

func (p *poolAdminReserveUsername) Put(x *AdminReserveUsername) {
	x.AccessToken = ""
	x.Usernames = x.Usernames[:0]
	x.Delete = false
	p.pool.Put(x)
}

var PoolAdminReserveUsername = poolAdminReserveUsername{}

const C_AdminGetReservedUsernames int64 = 1588181579

type poolAdminGetReservedUsernames struct {
	pool sync.Pool
}

func (p *poolAdminGetReservedUsernames) Get() *AdminGetReservedUsernames {
	x, ok := p.pool.Get().(*AdminGetReservedUsernames)
	if !ok {
		return &AdminGetReservedUsernames{}
	}
	return x
}

func (p *poolAdminGetReservedUsernames) Put(x *AdminGetReservedUsernames) {
	x.AccessToken = ""
	p.pool.Put(x)
}

var PoolAdminGetReservedUsernames = poolAdminGetReservedUsernames{}

const C_AdminTeamCreate int64 = 2797066608

type poolAdminTeamCreate struct {
	pool sync.Pool
}

func (p *poolAdminTeamCreate) Get() *AdminTeamCreate {
	x, ok := p.pool.Get().(*AdminTeamCreate)
	if !ok {
		return &AdminTeamCreate{}
	}
	return x
}

func (p *poolAdminTeamCreate) Put(x *AdminTeamCreate) {
	x.AccessToken = ""
	x.Capacity = 0
	x.ExpireDate = 0
	x.Community = false
	x.Title = ""
	x.CreatorID = 0
	p.pool.Put(x)
}

var PoolAdminTeamCreate = poolAdminTeamCreate{}

const C_AdminToken int64 = 2895609620

type poolAdminToken struct {
	pool sync.Pool
}

func (p *poolAdminToken) Get() *AdminToken {
	x, ok := p.pool.Get().(*AdminToken)
	if !ok {
		return &AdminToken{}
	}
	return x
}

func (p *poolAdminToken) Put(x *AdminToken) {
	x.Privilege = 0
	x.Token = ""
	p.pool.Put(x)
}

var PoolAdminToken = poolAdminToken{}

const C_WelcomeMessagesMany int64 = 414982091

type poolWelcomeMessagesMany struct {
	pool sync.Pool
}

func (p *poolWelcomeMessagesMany) Get() *WelcomeMessagesMany {
	x, ok := p.pool.Get().(*WelcomeMessagesMany)
	if !ok {
		return &WelcomeMessagesMany{}
	}
	return x
}

func (p *poolWelcomeMessagesMany) Put(x *WelcomeMessagesMany) {
	x.Messages = x.Messages[:0]
	x.Count = 0
	p.pool.Put(x)
}

var PoolWelcomeMessagesMany = poolWelcomeMessagesMany{}

const C_VersionsMany int64 = 2123920547

type poolVersionsMany struct {
	pool sync.Pool
}

func (p *poolVersionsMany) Get() *VersionsMany {
	x, ok := p.pool.Get().(*VersionsMany)
	if !ok {
		return &VersionsMany{}
	}
	return x
}

func (p *poolVersionsMany) Put(x *VersionsMany) {
	x.Versions = x.Versions[:0]
	x.Count = 0
	p.pool.Put(x)
}

var PoolVersionsMany = poolVersionsMany{}

const C_PushProvidersMany int64 = 5873573

type poolPushProvidersMany struct {
	pool sync.Pool
}

func (p *poolPushProvidersMany) Get() *PushProvidersMany {
	x, ok := p.pool.Get().(*PushProvidersMany)
	if !ok {
		return &PushProvidersMany{}
	}
	return x
}

func (p *poolPushProvidersMany) Put(x *PushProvidersMany) {
	x.Providers = x.Providers[:0]
	x.Count = 0
	p.pool.Put(x)
}

var PoolPushProvidersMany = poolPushProvidersMany{}

const C_WelcomeMessage int64 = 2506678571

type poolWelcomeMessage struct {
	pool sync.Pool
}

func (p *poolWelcomeMessage) Get() *WelcomeMessage {
	x, ok := p.pool.Get().(*WelcomeMessage)
	if !ok {
		return &WelcomeMessage{}
	}
	return x
}

func (p *poolWelcomeMessage) Put(x *WelcomeMessage) {
	x.Lang = ""
	x.Template = ""
	p.pool.Put(x)
}

var PoolWelcomeMessage = poolWelcomeMessage{}

const C_PushProvider int64 = 1015984470

type poolPushProvider struct {
	pool sync.Pool
}

func (p *poolPushProvider) Get() *PushProvider {
	x, ok := p.pool.Get().(*PushProvider)
	if !ok {
		return &PushProvider{}
	}
	return x
}

func (p *poolPushProvider) Put(x *PushProvider) {
	x.Name = ""
	x.Type = 0
	x.TestMode = false
	x.Credentials = x.Credentials[:0]
	x.KeyID = ""
	x.TeamID = ""
	x.Topic = ""
	p.pool.Put(x)
}

var PoolPushProvider = poolPushProvider{}

const C_Version int64 = 1889659487

type poolVersion struct {
	pool sync.Pool
}

func (p *poolVersion) Get() *Version {
	x, ok := p.pool.Get().(*Version)
	if !ok {
		return &Version{}
	}
	return x
}

func (p *poolVersion) Put(x *Version) {
	x.Vendor = ""
	x.Stage = ""
	x.OS = ""
	x.MinVersion = ""
	x.CurrentVersion = ""
	x.ForcedVersions = x.ForcedVersions[:0]
	p.pool.Put(x)
}

var PoolVersion = poolVersion{}

const C_ReservedUsernames int64 = 1388055751

type poolReservedUsernames struct {
	pool sync.Pool
}

func (p *poolReservedUsernames) Get() *ReservedUsernames {
	x, ok := p.pool.Get().(*ReservedUsernames)
	if !ok {
		return &ReservedUsernames{}
	}
	return x
}

func (p *poolReservedUsernames) Put(x *ReservedUsernames) {
	x.Usernames = x.Usernames[:0]
	x.Count = 0
	p.pool.Put(x)
}

var PoolReservedUsernames = poolReservedUsernames{}

func init() {
	registry.RegisterConstructor(3981417409, "AdminBroadcastMessage")
	registry.RegisterConstructor(1149591874, "AdminSetWelcomeMessage")
	registry.RegisterConstructor(2794709448, "AdminGetWelcomeMessages")
	registry.RegisterConstructor(3940015991, "AdminDeleteWelcomeMessage")
	registry.RegisterConstructor(1758606947, "AdminSetPushProvider")
	registry.RegisterConstructor(4257963974, "AdminGetPushProviders")
	registry.RegisterConstructor(1864898932, "AdminDeletePushProvider")
	registry.RegisterConstructor(1311023404, "AdminSetVersion")
	registry.RegisterConstructor(934752256, "AdminGetVersions")
	registry.RegisterConstructor(2892519162, "AdminSetToken")
	registry.RegisterConstructor(3154441897, "AdminDeleteToken")
	registry.RegisterConstructor(1947723452, "AdminReserveUsername")
	registry.RegisterConstructor(1588181579, "AdminGetReservedUsernames")
	registry.RegisterConstructor(2797066608, "AdminTeamCreate")
	registry.RegisterConstructor(2895609620, "AdminToken")
	registry.RegisterConstructor(414982091, "WelcomeMessagesMany")
	registry.RegisterConstructor(2123920547, "VersionsMany")
	registry.RegisterConstructor(5873573, "PushProvidersMany")
	registry.RegisterConstructor(2506678571, "WelcomeMessage")
	registry.RegisterConstructor(1015984470, "PushProvider")
	registry.RegisterConstructor(1889659487, "Version")
	registry.RegisterConstructor(1388055751, "ReservedUsernames")
}

func (x *AdminBroadcastMessage) DeepCopy(z *AdminBroadcastMessage) {
	z.Body = x.Body
	z.ReceiverIDs = append(z.ReceiverIDs[:0], x.ReceiverIDs...)
	for idx := range x.Entities {
		if x.Entities[idx] != nil {
			xx := PoolMessageEntity.Get()
			x.Entities[idx].DeepCopy(xx)
			z.Entities = append(z.Entities, xx)
		}
	}
	z.MediaType = x.MediaType
	z.MediaData = append(z.MediaData[:0], x.MediaData...)
}

func (x *AdminSetWelcomeMessage) DeepCopy(z *AdminSetWelcomeMessage) {
	z.AccessToken = x.AccessToken
	z.Lang = x.Lang
	z.Template = x.Template
}

func (x *AdminGetWelcomeMessages) DeepCopy(z *AdminGetWelcomeMessages) {
	z.AccessToken = x.AccessToken
}

func (x *AdminDeleteWelcomeMessage) DeepCopy(z *AdminDeleteWelcomeMessage) {
	z.AccessToken = x.AccessToken
	z.Lang = x.Lang
}

func (x *AdminSetPushProvider) DeepCopy(z *AdminSetPushProvider) {
	z.AccessToken = x.AccessToken
	if x.Provider != nil {
		z.Provider = PoolPushProvider.Get()
		x.Provider.DeepCopy(z.Provider)
	}
}

func (x *AdminGetPushProviders) DeepCopy(z *AdminGetPushProviders) {
	z.AccessToken = x.AccessToken
}

func (x *AdminDeletePushProvider) DeepCopy(z *AdminDeletePushProvider) {
	z.AccessToken = x.AccessToken
	z.Name = x.Name
	z.Type = x.Type
}

func (x *AdminSetVersion) DeepCopy(z *AdminSetVersion) {
	z.AccessToken = x.AccessToken
	if x.Version != nil {
		z.Version = PoolVersion.Get()
		x.Version.DeepCopy(z.Version)
	}
}

func (x *AdminGetVersions) DeepCopy(z *AdminGetVersions) {
	z.AccessToken = x.AccessToken
}

func (x *AdminSetToken) DeepCopy(z *AdminSetToken) {
	z.AccessToken = x.AccessToken
	z.Privilege = x.Privilege
}

func (x *AdminDeleteToken) DeepCopy(z *AdminDeleteToken) {
	z.AccessToken = x.AccessToken
	z.DeletedToken = x.DeletedToken
}

func (x *AdminReserveUsername) DeepCopy(z *AdminReserveUsername) {
	z.AccessToken = x.AccessToken
	z.Usernames = append(z.Usernames[:0], x.Usernames...)
	z.Delete = x.Delete
}

func (x *AdminGetReservedUsernames) DeepCopy(z *AdminGetReservedUsernames) {
	z.AccessToken = x.AccessToken
}

func (x *AdminTeamCreate) DeepCopy(z *AdminTeamCreate) {
	z.AccessToken = x.AccessToken
	z.Capacity = x.Capacity
	z.ExpireDate = x.ExpireDate
	z.Community = x.Community
	z.Title = x.Title
	z.CreatorID = x.CreatorID
}

func (x *AdminToken) DeepCopy(z *AdminToken) {
	z.Privilege = x.Privilege
	z.Token = x.Token
}

func (x *WelcomeMessagesMany) DeepCopy(z *WelcomeMessagesMany) {
	for idx := range x.Messages {
		if x.Messages[idx] != nil {
			xx := PoolWelcomeMessage.Get()
			x.Messages[idx].DeepCopy(xx)
			z.Messages = append(z.Messages, xx)
		}
	}
	z.Count = x.Count
}

func (x *VersionsMany) DeepCopy(z *VersionsMany) {
	for idx := range x.Versions {
		if x.Versions[idx] != nil {
			xx := PoolVersion.Get()
			x.Versions[idx].DeepCopy(xx)
			z.Versions = append(z.Versions, xx)
		}
	}
	z.Count = x.Count
}

func (x *PushProvidersMany) DeepCopy(z *PushProvidersMany) {
	for idx := range x.Providers {
		if x.Providers[idx] != nil {
			xx := PoolPushProvider.Get()
			x.Providers[idx].DeepCopy(xx)
			z.Providers = append(z.Providers, xx)
		}
	}
	z.Count = x.Count
}

func (x *WelcomeMessage) DeepCopy(z *WelcomeMessage) {
	z.Lang = x.Lang
	z.Template = x.Template
}

func (x *PushProvider) DeepCopy(z *PushProvider) {
	z.Name = x.Name
	z.Type = x.Type
	z.TestMode = x.TestMode
	z.Credentials = append(z.Credentials[:0], x.Credentials...)
	z.KeyID = x.KeyID
	z.TeamID = x.TeamID
	z.Topic = x.Topic
}

func (x *Version) DeepCopy(z *Version) {
	z.Vendor = x.Vendor
	z.Stage = x.Stage
	z.OS = x.OS
	z.MinVersion = x.MinVersion
	z.CurrentVersion = x.CurrentVersion
	z.ForcedVersions = append(z.ForcedVersions[:0], x.ForcedVersions...)
}

func (x *ReservedUsernames) DeepCopy(z *ReservedUsernames) {
	z.Usernames = append(z.Usernames[:0], x.Usernames...)
	z.Count = x.Count
}

func (x *AdminBroadcastMessage) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AdminBroadcastMessage, x)
}

func (x *AdminSetWelcomeMessage) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AdminSetWelcomeMessage, x)
}

func (x *AdminGetWelcomeMessages) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AdminGetWelcomeMessages, x)
}

func (x *AdminDeleteWelcomeMessage) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AdminDeleteWelcomeMessage, x)
}

func (x *AdminSetPushProvider) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AdminSetPushProvider, x)
}

func (x *AdminGetPushProviders) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AdminGetPushProviders, x)
}

func (x *AdminDeletePushProvider) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AdminDeletePushProvider, x)
}

func (x *AdminSetVersion) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AdminSetVersion, x)
}

func (x *AdminGetVersions) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AdminGetVersions, x)
}

func (x *AdminSetToken) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AdminSetToken, x)
}

func (x *AdminDeleteToken) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AdminDeleteToken, x)
}

func (x *AdminReserveUsername) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AdminReserveUsername, x)
}

func (x *AdminGetReservedUsernames) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AdminGetReservedUsernames, x)
}

func (x *AdminTeamCreate) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AdminTeamCreate, x)
}

func (x *AdminToken) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AdminToken, x)
}

func (x *WelcomeMessagesMany) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_WelcomeMessagesMany, x)
}

func (x *VersionsMany) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_VersionsMany, x)
}

func (x *PushProvidersMany) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_PushProvidersMany, x)
}

func (x *WelcomeMessage) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_WelcomeMessage, x)
}

func (x *PushProvider) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_PushProvider, x)
}

func (x *Version) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_Version, x)
}

func (x *ReservedUsernames) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_ReservedUsernames, x)
}

func (x *AdminBroadcastMessage) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AdminSetWelcomeMessage) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AdminGetWelcomeMessages) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AdminDeleteWelcomeMessage) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AdminSetPushProvider) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AdminGetPushProviders) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AdminDeletePushProvider) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AdminSetVersion) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AdminGetVersions) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AdminSetToken) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AdminDeleteToken) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AdminReserveUsername) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AdminGetReservedUsernames) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AdminTeamCreate) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AdminToken) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *WelcomeMessagesMany) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *VersionsMany) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *PushProvidersMany) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *WelcomeMessage) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *PushProvider) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *Version) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *ReservedUsernames) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AdminBroadcastMessage) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AdminSetWelcomeMessage) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AdminGetWelcomeMessages) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AdminDeleteWelcomeMessage) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AdminSetPushProvider) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AdminGetPushProviders) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AdminDeletePushProvider) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AdminSetVersion) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AdminGetVersions) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AdminSetToken) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AdminDeleteToken) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AdminReserveUsername) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AdminGetReservedUsernames) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AdminTeamCreate) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AdminToken) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *WelcomeMessagesMany) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *VersionsMany) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *PushProvidersMany) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *WelcomeMessage) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *PushProvider) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *Version) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *ReservedUsernames) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}
