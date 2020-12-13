package msg

import (
	edge "github.com/ronaksoft/rony/edge"
	registry "github.com/ronaksoft/rony/registry"
	proto "google.golang.org/protobuf/proto"
	sync "sync"
)

const C_PasswordAlgorithmVer6A int64 = 341860043

type poolPasswordAlgorithmVer6A struct {
	pool sync.Pool
}

func (p *poolPasswordAlgorithmVer6A) Get() *PasswordAlgorithmVer6A {
	x, ok := p.pool.Get().(*PasswordAlgorithmVer6A)
	if !ok {
		return &PasswordAlgorithmVer6A{}
	}
	return x
}

func (p *poolPasswordAlgorithmVer6A) Put(x *PasswordAlgorithmVer6A) {
	x.Salt1 = x.Salt1[:0]
	x.Salt2 = x.Salt2[:0]
	x.G = 0
	x.P = x.P[:0]
	p.pool.Put(x)
}

var PoolPasswordAlgorithmVer6A = poolPasswordAlgorithmVer6A{}

func init() {
	registry.RegisterConstructor(341860043, "PasswordAlgorithmVer6A")
}

func (x *PasswordAlgorithmVer6A) DeepCopy(z *PasswordAlgorithmVer6A) {
	z.Salt1 = append(z.Salt1[:0], x.Salt1...)
	z.Salt2 = append(z.Salt2[:0], x.Salt2...)
	z.G = x.G
	z.P = append(z.P[:0], x.P...)
}

func (x *PasswordAlgorithmVer6A) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_PasswordAlgorithmVer6A, x)
}

func (x *PasswordAlgorithmVer6A) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *PasswordAlgorithmVer6A) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}
