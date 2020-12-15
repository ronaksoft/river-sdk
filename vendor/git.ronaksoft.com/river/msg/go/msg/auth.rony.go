package msg

import (
	edge "github.com/ronaksoft/rony/edge"
	registry "github.com/ronaksoft/rony/registry"
	proto "google.golang.org/protobuf/proto"
	sync "sync"
)

const C_InitConnect int64 = 4150793517

type poolInitConnect struct {
	pool sync.Pool
}

func (p *poolInitConnect) Get() *InitConnect {
	x, ok := p.pool.Get().(*InitConnect)
	if !ok {
		return &InitConnect{}
	}
	return x
}

func (p *poolInitConnect) Put(x *InitConnect) {
	x.ClientNonce = 0
	p.pool.Put(x)
}

var PoolInitConnect = poolInitConnect{}

const C_InitCompleteAuth int64 = 1583178320

type poolInitCompleteAuth struct {
	pool sync.Pool
}

func (p *poolInitCompleteAuth) Get() *InitCompleteAuth {
	x, ok := p.pool.Get().(*InitCompleteAuth)
	if !ok {
		return &InitCompleteAuth{}
	}
	return x
}

func (p *poolInitCompleteAuth) Put(x *InitCompleteAuth) {
	x.ClientNonce = 0
	x.ServerNonce = 0
	x.ClientDHPubKey = x.ClientDHPubKey[:0]
	x.P = 0
	x.Q = 0
	x.EncryptedPayload = x.EncryptedPayload[:0]
	p.pool.Put(x)
}

var PoolInitCompleteAuth = poolInitCompleteAuth{}

const C_InitConnectTest int64 = 3188015450

type poolInitConnectTest struct {
	pool sync.Pool
}

func (p *poolInitConnectTest) Get() *InitConnectTest {
	x, ok := p.pool.Get().(*InitConnectTest)
	if !ok {
		return &InitConnectTest{}
	}
	return x
}

func (p *poolInitConnectTest) Put(x *InitConnectTest) {
	p.pool.Put(x)
}

var PoolInitConnectTest = poolInitConnectTest{}

const C_InitBindUser int64 = 1933549113

type poolInitBindUser struct {
	pool sync.Pool
}

func (p *poolInitBindUser) Get() *InitBindUser {
	x, ok := p.pool.Get().(*InitBindUser)
	if !ok {
		return &InitBindUser{}
	}
	return x
}

func (p *poolInitBindUser) Put(x *InitBindUser) {
	x.AuthKey = ""
	x.Username = ""
	x.Phone = ""
	x.FirstName = ""
	x.LastName = ""
	p.pool.Put(x)
}

var PoolInitBindUser = poolInitBindUser{}

const C_AuthRegister int64 = 2228369460

type poolAuthRegister struct {
	pool sync.Pool
}

func (p *poolAuthRegister) Get() *AuthRegister {
	x, ok := p.pool.Get().(*AuthRegister)
	if !ok {
		return &AuthRegister{}
	}
	return x
}

func (p *poolAuthRegister) Put(x *AuthRegister) {
	x.Phone = ""
	x.FirstName = ""
	x.LastName = ""
	x.PhoneCode = ""
	x.PhoneCodeHash = ""
	x.LangCode = ""
	p.pool.Put(x)
}

var PoolAuthRegister = poolAuthRegister{}

const C_AuthBotRegister int64 = 1579606687

type poolAuthBotRegister struct {
	pool sync.Pool
}

func (p *poolAuthBotRegister) Get() *AuthBotRegister {
	x, ok := p.pool.Get().(*AuthBotRegister)
	if !ok {
		return &AuthBotRegister{}
	}
	return x
}

func (p *poolAuthBotRegister) Put(x *AuthBotRegister) {
	x.Name = ""
	x.Username = ""
	x.OwnerID = 0
	p.pool.Put(x)
}

var PoolAuthBotRegister = poolAuthBotRegister{}

const C_AuthLogin int64 = 2587620888

type poolAuthLogin struct {
	pool sync.Pool
}

func (p *poolAuthLogin) Get() *AuthLogin {
	x, ok := p.pool.Get().(*AuthLogin)
	if !ok {
		return &AuthLogin{}
	}
	return x
}

func (p *poolAuthLogin) Put(x *AuthLogin) {
	x.Phone = ""
	x.PhoneCodeHash = ""
	x.PhoneCode = ""
	p.pool.Put(x)
}

var PoolAuthLogin = poolAuthLogin{}

const C_AuthCheckPassword int64 = 3346962908

type poolAuthCheckPassword struct {
	pool sync.Pool
}

func (p *poolAuthCheckPassword) Get() *AuthCheckPassword {
	x, ok := p.pool.Get().(*AuthCheckPassword)
	if !ok {
		return &AuthCheckPassword{}
	}
	return x
}

func (p *poolAuthCheckPassword) Put(x *AuthCheckPassword) {
	if x.Password != nil {
		PoolInputPassword.Put(x.Password)
		x.Password = nil
	}
	p.pool.Put(x)
}

var PoolAuthCheckPassword = poolAuthCheckPassword{}

const C_AuthRecoverPassword int64 = 2711231991

type poolAuthRecoverPassword struct {
	pool sync.Pool
}

func (p *poolAuthRecoverPassword) Get() *AuthRecoverPassword {
	x, ok := p.pool.Get().(*AuthRecoverPassword)
	if !ok {
		return &AuthRecoverPassword{}
	}
	return x
}

func (p *poolAuthRecoverPassword) Put(x *AuthRecoverPassword) {
	x.Code = ""
	p.pool.Put(x)
}

var PoolAuthRecoverPassword = poolAuthRecoverPassword{}

const C_AuthLogout int64 = 992431648

type poolAuthLogout struct {
	pool sync.Pool
}

func (p *poolAuthLogout) Get() *AuthLogout {
	x, ok := p.pool.Get().(*AuthLogout)
	if !ok {
		return &AuthLogout{}
	}
	return x
}

func (p *poolAuthLogout) Put(x *AuthLogout) {
	x.AuthIDs = x.AuthIDs[:0]
	p.pool.Put(x)
}

var PoolAuthLogout = poolAuthLogout{}

const C_AuthLoginByToken int64 = 2851553023

type poolAuthLoginByToken struct {
	pool sync.Pool
}

func (p *poolAuthLoginByToken) Get() *AuthLoginByToken {
	x, ok := p.pool.Get().(*AuthLoginByToken)
	if !ok {
		return &AuthLoginByToken{}
	}
	return x
}

func (p *poolAuthLoginByToken) Put(x *AuthLoginByToken) {
	x.Token = ""
	x.Provider = ""
	x.Firstname = ""
	x.Lastname = ""
	p.pool.Put(x)
}

var PoolAuthLoginByToken = poolAuthLoginByToken{}

const C_AuthCheckPhone int64 = 4134648516

type poolAuthCheckPhone struct {
	pool sync.Pool
}

func (p *poolAuthCheckPhone) Get() *AuthCheckPhone {
	x, ok := p.pool.Get().(*AuthCheckPhone)
	if !ok {
		return &AuthCheckPhone{}
	}
	return x
}

func (p *poolAuthCheckPhone) Put(x *AuthCheckPhone) {
	x.Phone = ""
	p.pool.Put(x)
}

var PoolAuthCheckPhone = poolAuthCheckPhone{}

const C_AuthSendCode int64 = 3984043365

type poolAuthSendCode struct {
	pool sync.Pool
}

func (p *poolAuthSendCode) Get() *AuthSendCode {
	x, ok := p.pool.Get().(*AuthSendCode)
	if !ok {
		return &AuthSendCode{}
	}
	return x
}

func (p *poolAuthSendCode) Put(x *AuthSendCode) {
	x.Phone = ""
	x.AppHash = ""
	p.pool.Put(x)
}

var PoolAuthSendCode = poolAuthSendCode{}

const C_AuthResendCode int64 = 2682713491

type poolAuthResendCode struct {
	pool sync.Pool
}

func (p *poolAuthResendCode) Get() *AuthResendCode {
	x, ok := p.pool.Get().(*AuthResendCode)
	if !ok {
		return &AuthResendCode{}
	}
	return x
}

func (p *poolAuthResendCode) Put(x *AuthResendCode) {
	x.Phone = ""
	x.PhoneCodeHash = ""
	x.AppHash = ""
	p.pool.Put(x)
}

var PoolAuthResendCode = poolAuthResendCode{}

const C_AuthRecall int64 = 1172029049

type poolAuthRecall struct {
	pool sync.Pool
}

func (p *poolAuthRecall) Get() *AuthRecall {
	x, ok := p.pool.Get().(*AuthRecall)
	if !ok {
		return &AuthRecall{}
	}
	return x
}

func (p *poolAuthRecall) Put(x *AuthRecall) {
	x.ClientID = 0
	x.Version = 0
	x.AppVersion = ""
	x.Platform = ""
	x.Vendor = ""
	x.OSVersion = ""
	p.pool.Put(x)
}

var PoolAuthRecall = poolAuthRecall{}

const C_AuthDestroyKey int64 = 3673422656

type poolAuthDestroyKey struct {
	pool sync.Pool
}

func (p *poolAuthDestroyKey) Get() *AuthDestroyKey {
	x, ok := p.pool.Get().(*AuthDestroyKey)
	if !ok {
		return &AuthDestroyKey{}
	}
	return x
}

func (p *poolAuthDestroyKey) Put(x *AuthDestroyKey) {
	p.pool.Put(x)
}

var PoolAuthDestroyKey = poolAuthDestroyKey{}

const C_InitTestAuth int64 = 2762878006

type poolInitTestAuth struct {
	pool sync.Pool
}

func (p *poolInitTestAuth) Get() *InitTestAuth {
	x, ok := p.pool.Get().(*InitTestAuth)
	if !ok {
		return &InitTestAuth{}
	}
	return x
}

func (p *poolInitTestAuth) Put(x *InitTestAuth) {
	x.AuthID = 0
	x.AuthKey = x.AuthKey[:0]
	p.pool.Put(x)
}

var PoolInitTestAuth = poolInitTestAuth{}

const C_InitResponse int64 = 4130340247

type poolInitResponse struct {
	pool sync.Pool
}

func (p *poolInitResponse) Get() *InitResponse {
	x, ok := p.pool.Get().(*InitResponse)
	if !ok {
		return &InitResponse{}
	}
	return x
}

func (p *poolInitResponse) Put(x *InitResponse) {
	x.ClientNonce = 0
	x.ServerNonce = 0
	x.RSAPubKeyFingerPrint = 0
	x.DHGroupFingerPrint = 0
	x.PQ = 0
	x.ServerTimestamp = 0
	p.pool.Put(x)
}

var PoolInitResponse = poolInitResponse{}

const C_InitCompleteAuthInternal int64 = 2360982492

type poolInitCompleteAuthInternal struct {
	pool sync.Pool
}

func (p *poolInitCompleteAuthInternal) Get() *InitCompleteAuthInternal {
	x, ok := p.pool.Get().(*InitCompleteAuthInternal)
	if !ok {
		return &InitCompleteAuthInternal{}
	}
	return x
}

func (p *poolInitCompleteAuthInternal) Put(x *InitCompleteAuthInternal) {
	x.SecretNonce = x.SecretNonce[:0]
	p.pool.Put(x)
}

var PoolInitCompleteAuthInternal = poolInitCompleteAuthInternal{}

const C_InitAuthCompleted int64 = 627708982

type poolInitAuthCompleted struct {
	pool sync.Pool
}

func (p *poolInitAuthCompleted) Get() *InitAuthCompleted {
	x, ok := p.pool.Get().(*InitAuthCompleted)
	if !ok {
		return &InitAuthCompleted{}
	}
	return x
}

func (p *poolInitAuthCompleted) Put(x *InitAuthCompleted) {
	x.ClientNonce = 0
	x.ServerNonce = 0
	x.Status = 0
	x.SecretHash = 0
	x.ServerDHPubKey = x.ServerDHPubKey[:0]
	p.pool.Put(x)
}

var PoolInitAuthCompleted = poolInitAuthCompleted{}

const C_InitUserBound int64 = 128391141

type poolInitUserBound struct {
	pool sync.Pool
}

func (p *poolInitUserBound) Get() *InitUserBound {
	x, ok := p.pool.Get().(*InitUserBound)
	if !ok {
		return &InitUserBound{}
	}
	return x
}

func (p *poolInitUserBound) Put(x *InitUserBound) {
	x.AuthID = 0
	p.pool.Put(x)
}

var PoolInitUserBound = poolInitUserBound{}

const C_AuthPasswordRecovery int64 = 3813475914

type poolAuthPasswordRecovery struct {
	pool sync.Pool
}

func (p *poolAuthPasswordRecovery) Get() *AuthPasswordRecovery {
	x, ok := p.pool.Get().(*AuthPasswordRecovery)
	if !ok {
		return &AuthPasswordRecovery{}
	}
	return x
}

func (p *poolAuthPasswordRecovery) Put(x *AuthPasswordRecovery) {
	x.EmailPattern = ""
	p.pool.Put(x)
}

var PoolAuthPasswordRecovery = poolAuthPasswordRecovery{}

const C_AuthRecalled int64 = 3249025459

type poolAuthRecalled struct {
	pool sync.Pool
}

func (p *poolAuthRecalled) Get() *AuthRecalled {
	x, ok := p.pool.Get().(*AuthRecalled)
	if !ok {
		return &AuthRecalled{}
	}
	return x
}

func (p *poolAuthRecalled) Put(x *AuthRecalled) {
	x.ClientID = 0
	x.Timestamp = 0
	x.UpdateID = 0
	x.Available = false
	x.Force = false
	x.CurrentVersion = ""
	p.pool.Put(x)
}

var PoolAuthRecalled = poolAuthRecalled{}

const C_AuthAuthorization int64 = 1140037965

type poolAuthAuthorization struct {
	pool sync.Pool
}

func (p *poolAuthAuthorization) Get() *AuthAuthorization {
	x, ok := p.pool.Get().(*AuthAuthorization)
	if !ok {
		return &AuthAuthorization{}
	}
	return x
}

func (p *poolAuthAuthorization) Put(x *AuthAuthorization) {
	x.Expired = 0
	if x.User != nil {
		PoolUser.Put(x.User)
		x.User = nil
	}
	x.ActiveSessions = 0
	p.pool.Put(x)
}

var PoolAuthAuthorization = poolAuthAuthorization{}

const C_AuthBotAuthorization int64 = 3304560814

type poolAuthBotAuthorization struct {
	pool sync.Pool
}

func (p *poolAuthBotAuthorization) Get() *AuthBotAuthorization {
	x, ok := p.pool.Get().(*AuthBotAuthorization)
	if !ok {
		return &AuthBotAuthorization{}
	}
	return x
}

func (p *poolAuthBotAuthorization) Put(x *AuthBotAuthorization) {
	x.AuthID = 0
	x.AuthKey = x.AuthKey[:0]
	if x.Bot != nil {
		PoolBot.Put(x.Bot)
		x.Bot = nil
	}
	p.pool.Put(x)
}

var PoolAuthBotAuthorization = poolAuthBotAuthorization{}

const C_AuthCheckedPhone int64 = 2236203131

type poolAuthCheckedPhone struct {
	pool sync.Pool
}

func (p *poolAuthCheckedPhone) Get() *AuthCheckedPhone {
	x, ok := p.pool.Get().(*AuthCheckedPhone)
	if !ok {
		return &AuthCheckedPhone{}
	}
	return x
}

func (p *poolAuthCheckedPhone) Put(x *AuthCheckedPhone) {
	x.Invited = false
	x.Registered = false
	p.pool.Put(x)
}

var PoolAuthCheckedPhone = poolAuthCheckedPhone{}

const C_AuthSentCode int64 = 2375498471

type poolAuthSentCode struct {
	pool sync.Pool
}

func (p *poolAuthSentCode) Get() *AuthSentCode {
	x, ok := p.pool.Get().(*AuthSentCode)
	if !ok {
		return &AuthSentCode{}
	}
	return x
}

func (p *poolAuthSentCode) Put(x *AuthSentCode) {
	x.Phone = ""
	x.PhoneCodeHash = ""
	x.SendToPhone = false
	p.pool.Put(x)
}

var PoolAuthSentCode = poolAuthSentCode{}

func init() {
	registry.RegisterConstructor(4150793517, "InitConnect")
	registry.RegisterConstructor(1583178320, "InitCompleteAuth")
	registry.RegisterConstructor(3188015450, "InitConnectTest")
	registry.RegisterConstructor(1933549113, "InitBindUser")
	registry.RegisterConstructor(2228369460, "AuthRegister")
	registry.RegisterConstructor(1579606687, "AuthBotRegister")
	registry.RegisterConstructor(2587620888, "AuthLogin")
	registry.RegisterConstructor(3346962908, "AuthCheckPassword")
	registry.RegisterConstructor(2711231991, "AuthRecoverPassword")
	registry.RegisterConstructor(992431648, "AuthLogout")
	registry.RegisterConstructor(2851553023, "AuthLoginByToken")
	registry.RegisterConstructor(4134648516, "AuthCheckPhone")
	registry.RegisterConstructor(3984043365, "AuthSendCode")
	registry.RegisterConstructor(2682713491, "AuthResendCode")
	registry.RegisterConstructor(1172029049, "AuthRecall")
	registry.RegisterConstructor(3673422656, "AuthDestroyKey")
	registry.RegisterConstructor(2762878006, "InitTestAuth")
	registry.RegisterConstructor(4130340247, "InitResponse")
	registry.RegisterConstructor(2360982492, "InitCompleteAuthInternal")
	registry.RegisterConstructor(627708982, "InitAuthCompleted")
	registry.RegisterConstructor(128391141, "InitUserBound")
	registry.RegisterConstructor(3813475914, "AuthPasswordRecovery")
	registry.RegisterConstructor(3249025459, "AuthRecalled")
	registry.RegisterConstructor(1140037965, "AuthAuthorization")
	registry.RegisterConstructor(3304560814, "AuthBotAuthorization")
	registry.RegisterConstructor(2236203131, "AuthCheckedPhone")
	registry.RegisterConstructor(2375498471, "AuthSentCode")
}

func (x *InitConnect) DeepCopy(z *InitConnect) {
	z.ClientNonce = x.ClientNonce
}

func (x *InitCompleteAuth) DeepCopy(z *InitCompleteAuth) {
	z.ClientNonce = x.ClientNonce
	z.ServerNonce = x.ServerNonce
	z.ClientDHPubKey = append(z.ClientDHPubKey[:0], x.ClientDHPubKey...)
	z.P = x.P
	z.Q = x.Q
	z.EncryptedPayload = append(z.EncryptedPayload[:0], x.EncryptedPayload...)
}

func (x *InitConnectTest) DeepCopy(z *InitConnectTest) {
}

func (x *InitBindUser) DeepCopy(z *InitBindUser) {
	z.AuthKey = x.AuthKey
	z.Username = x.Username
	z.Phone = x.Phone
	z.FirstName = x.FirstName
	z.LastName = x.LastName
}

func (x *AuthRegister) DeepCopy(z *AuthRegister) {
	z.Phone = x.Phone
	z.FirstName = x.FirstName
	z.LastName = x.LastName
	z.PhoneCode = x.PhoneCode
	z.PhoneCodeHash = x.PhoneCodeHash
	z.LangCode = x.LangCode
}

func (x *AuthBotRegister) DeepCopy(z *AuthBotRegister) {
	z.Name = x.Name
	z.Username = x.Username
	z.OwnerID = x.OwnerID
}

func (x *AuthLogin) DeepCopy(z *AuthLogin) {
	z.Phone = x.Phone
	z.PhoneCodeHash = x.PhoneCodeHash
	z.PhoneCode = x.PhoneCode
}

func (x *AuthCheckPassword) DeepCopy(z *AuthCheckPassword) {
	if x.Password != nil {
		z.Password = PoolInputPassword.Get()
		x.Password.DeepCopy(z.Password)
	}
}

func (x *AuthRecoverPassword) DeepCopy(z *AuthRecoverPassword) {
	z.Code = x.Code
}

func (x *AuthLogout) DeepCopy(z *AuthLogout) {
	z.AuthIDs = append(z.AuthIDs[:0], x.AuthIDs...)
}

func (x *AuthLoginByToken) DeepCopy(z *AuthLoginByToken) {
	z.Token = x.Token
	z.Provider = x.Provider
	z.Firstname = x.Firstname
	z.Lastname = x.Lastname
}

func (x *AuthCheckPhone) DeepCopy(z *AuthCheckPhone) {
	z.Phone = x.Phone
}

func (x *AuthSendCode) DeepCopy(z *AuthSendCode) {
	z.Phone = x.Phone
	z.AppHash = x.AppHash
}

func (x *AuthResendCode) DeepCopy(z *AuthResendCode) {
	z.Phone = x.Phone
	z.PhoneCodeHash = x.PhoneCodeHash
	z.AppHash = x.AppHash
}

func (x *AuthRecall) DeepCopy(z *AuthRecall) {
	z.ClientID = x.ClientID
	z.Version = x.Version
	z.AppVersion = x.AppVersion
	z.Platform = x.Platform
	z.Vendor = x.Vendor
	z.OSVersion = x.OSVersion
}

func (x *AuthDestroyKey) DeepCopy(z *AuthDestroyKey) {
}

func (x *InitTestAuth) DeepCopy(z *InitTestAuth) {
	z.AuthID = x.AuthID
	z.AuthKey = append(z.AuthKey[:0], x.AuthKey...)
}

func (x *InitResponse) DeepCopy(z *InitResponse) {
	z.ClientNonce = x.ClientNonce
	z.ServerNonce = x.ServerNonce
	z.RSAPubKeyFingerPrint = x.RSAPubKeyFingerPrint
	z.DHGroupFingerPrint = x.DHGroupFingerPrint
	z.PQ = x.PQ
	z.ServerTimestamp = x.ServerTimestamp
}

func (x *InitCompleteAuthInternal) DeepCopy(z *InitCompleteAuthInternal) {
	z.SecretNonce = append(z.SecretNonce[:0], x.SecretNonce...)
}

func (x *InitAuthCompleted) DeepCopy(z *InitAuthCompleted) {
	z.ClientNonce = x.ClientNonce
	z.ServerNonce = x.ServerNonce
	z.Status = x.Status
	z.SecretHash = x.SecretHash
	z.ServerDHPubKey = append(z.ServerDHPubKey[:0], x.ServerDHPubKey...)
}

func (x *InitUserBound) DeepCopy(z *InitUserBound) {
	z.AuthID = x.AuthID
}

func (x *AuthPasswordRecovery) DeepCopy(z *AuthPasswordRecovery) {
	z.EmailPattern = x.EmailPattern
}

func (x *AuthRecalled) DeepCopy(z *AuthRecalled) {
	z.ClientID = x.ClientID
	z.Timestamp = x.Timestamp
	z.UpdateID = x.UpdateID
	z.Available = x.Available
	z.Force = x.Force
	z.CurrentVersion = x.CurrentVersion
}

func (x *AuthAuthorization) DeepCopy(z *AuthAuthorization) {
	z.Expired = x.Expired
	if x.User != nil {
		z.User = PoolUser.Get()
		x.User.DeepCopy(z.User)
	}
	z.ActiveSessions = x.ActiveSessions
}

func (x *AuthBotAuthorization) DeepCopy(z *AuthBotAuthorization) {
	z.AuthID = x.AuthID
	z.AuthKey = append(z.AuthKey[:0], x.AuthKey...)
	if x.Bot != nil {
		z.Bot = PoolBot.Get()
		x.Bot.DeepCopy(z.Bot)
	}
}

func (x *AuthCheckedPhone) DeepCopy(z *AuthCheckedPhone) {
	z.Invited = x.Invited
	z.Registered = x.Registered
}

func (x *AuthSentCode) DeepCopy(z *AuthSentCode) {
	z.Phone = x.Phone
	z.PhoneCodeHash = x.PhoneCodeHash
	z.SendToPhone = x.SendToPhone
}

func (x *InitConnect) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_InitConnect, x)
}

func (x *InitCompleteAuth) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_InitCompleteAuth, x)
}

func (x *InitConnectTest) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_InitConnectTest, x)
}

func (x *InitBindUser) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_InitBindUser, x)
}

func (x *AuthRegister) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AuthRegister, x)
}

func (x *AuthBotRegister) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AuthBotRegister, x)
}

func (x *AuthLogin) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AuthLogin, x)
}

func (x *AuthCheckPassword) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AuthCheckPassword, x)
}

func (x *AuthRecoverPassword) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AuthRecoverPassword, x)
}

func (x *AuthLogout) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AuthLogout, x)
}

func (x *AuthLoginByToken) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AuthLoginByToken, x)
}

func (x *AuthCheckPhone) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AuthCheckPhone, x)
}

func (x *AuthSendCode) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AuthSendCode, x)
}

func (x *AuthResendCode) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AuthResendCode, x)
}

func (x *AuthRecall) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AuthRecall, x)
}

func (x *AuthDestroyKey) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AuthDestroyKey, x)
}

func (x *InitTestAuth) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_InitTestAuth, x)
}

func (x *InitResponse) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_InitResponse, x)
}

func (x *InitCompleteAuthInternal) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_InitCompleteAuthInternal, x)
}

func (x *InitAuthCompleted) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_InitAuthCompleted, x)
}

func (x *InitUserBound) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_InitUserBound, x)
}

func (x *AuthPasswordRecovery) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AuthPasswordRecovery, x)
}

func (x *AuthRecalled) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AuthRecalled, x)
}

func (x *AuthAuthorization) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AuthAuthorization, x)
}

func (x *AuthBotAuthorization) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AuthBotAuthorization, x)
}

func (x *AuthCheckedPhone) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AuthCheckedPhone, x)
}

func (x *AuthSentCode) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_AuthSentCode, x)
}

func (x *InitConnect) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *InitCompleteAuth) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *InitConnectTest) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *InitBindUser) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AuthRegister) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AuthBotRegister) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AuthLogin) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AuthCheckPassword) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AuthRecoverPassword) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AuthLogout) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AuthLoginByToken) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AuthCheckPhone) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AuthSendCode) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AuthResendCode) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AuthRecall) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AuthDestroyKey) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *InitTestAuth) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *InitResponse) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *InitCompleteAuthInternal) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *InitAuthCompleted) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *InitUserBound) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AuthPasswordRecovery) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AuthRecalled) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AuthAuthorization) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AuthBotAuthorization) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AuthCheckedPhone) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *AuthSentCode) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *InitConnect) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *InitCompleteAuth) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *InitConnectTest) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *InitBindUser) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AuthRegister) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AuthBotRegister) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AuthLogin) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AuthCheckPassword) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AuthRecoverPassword) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AuthLogout) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AuthLoginByToken) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AuthCheckPhone) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AuthSendCode) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AuthResendCode) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AuthRecall) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AuthDestroyKey) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *InitTestAuth) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *InitResponse) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *InitCompleteAuthInternal) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *InitAuthCompleted) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *InitUserBound) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AuthPasswordRecovery) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AuthRecalled) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AuthAuthorization) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AuthBotAuthorization) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AuthCheckedPhone) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *AuthSentCode) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}
