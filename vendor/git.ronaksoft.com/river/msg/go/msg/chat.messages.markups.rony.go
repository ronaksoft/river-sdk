package msg

import (
	edge "github.com/ronaksoft/rony/edge"
	registry "github.com/ronaksoft/rony/registry"
	proto "google.golang.org/protobuf/proto"
	sync "sync"
)

const C_ReplyKeyboardMarkup int64 = 1353207252

type poolReplyKeyboardMarkup struct {
	pool sync.Pool
}

func (p *poolReplyKeyboardMarkup) Get() *ReplyKeyboardMarkup {
	x, ok := p.pool.Get().(*ReplyKeyboardMarkup)
	if !ok {
		return &ReplyKeyboardMarkup{}
	}
	return x
}

func (p *poolReplyKeyboardMarkup) Put(x *ReplyKeyboardMarkup) {
	x.SingleUse = false
	x.Selective = false
	x.Resize = false
	x.Rows = x.Rows[:0]
	p.pool.Put(x)
}

var PoolReplyKeyboardMarkup = poolReplyKeyboardMarkup{}

const C_ReplyInlineMarkup int64 = 3617178965

type poolReplyInlineMarkup struct {
	pool sync.Pool
}

func (p *poolReplyInlineMarkup) Get() *ReplyInlineMarkup {
	x, ok := p.pool.Get().(*ReplyInlineMarkup)
	if !ok {
		return &ReplyInlineMarkup{}
	}
	return x
}

func (p *poolReplyInlineMarkup) Put(x *ReplyInlineMarkup) {
	x.Rows = x.Rows[:0]
	p.pool.Put(x)
}

var PoolReplyInlineMarkup = poolReplyInlineMarkup{}

const C_ReplyKeyboardHide int64 = 4235074858

type poolReplyKeyboardHide struct {
	pool sync.Pool
}

func (p *poolReplyKeyboardHide) Get() *ReplyKeyboardHide {
	x, ok := p.pool.Get().(*ReplyKeyboardHide)
	if !ok {
		return &ReplyKeyboardHide{}
	}
	return x
}

func (p *poolReplyKeyboardHide) Put(x *ReplyKeyboardHide) {
	x.Selective = false
	p.pool.Put(x)
}

var PoolReplyKeyboardHide = poolReplyKeyboardHide{}

const C_ReplyKeyboardForceReply int64 = 4261878523

type poolReplyKeyboardForceReply struct {
	pool sync.Pool
}

func (p *poolReplyKeyboardForceReply) Get() *ReplyKeyboardForceReply {
	x, ok := p.pool.Get().(*ReplyKeyboardForceReply)
	if !ok {
		return &ReplyKeyboardForceReply{}
	}
	return x
}

func (p *poolReplyKeyboardForceReply) Put(x *ReplyKeyboardForceReply) {
	x.SingleUse = false
	x.Selective = false
	p.pool.Put(x)
}

var PoolReplyKeyboardForceReply = poolReplyKeyboardForceReply{}

const C_KeyboardButtonRow int64 = 3268825054

type poolKeyboardButtonRow struct {
	pool sync.Pool
}

func (p *poolKeyboardButtonRow) Get() *KeyboardButtonRow {
	x, ok := p.pool.Get().(*KeyboardButtonRow)
	if !ok {
		return &KeyboardButtonRow{}
	}
	return x
}

func (p *poolKeyboardButtonRow) Put(x *KeyboardButtonRow) {
	x.Buttons = x.Buttons[:0]
	p.pool.Put(x)
}

var PoolKeyboardButtonRow = poolKeyboardButtonRow{}

const C_KeyboardButtonEnvelope int64 = 692302489

type poolKeyboardButtonEnvelope struct {
	pool sync.Pool
}

func (p *poolKeyboardButtonEnvelope) Get() *KeyboardButtonEnvelope {
	x, ok := p.pool.Get().(*KeyboardButtonEnvelope)
	if !ok {
		return &KeyboardButtonEnvelope{}
	}
	return x
}

func (p *poolKeyboardButtonEnvelope) Put(x *KeyboardButtonEnvelope) {
	x.Constructor = 0
	x.Data = x.Data[:0]
	p.pool.Put(x)
}

var PoolKeyboardButtonEnvelope = poolKeyboardButtonEnvelope{}

const C_Button int64 = 3346660205

type poolButton struct {
	pool sync.Pool
}

func (p *poolButton) Get() *Button {
	x, ok := p.pool.Get().(*Button)
	if !ok {
		return &Button{}
	}
	return x
}

func (p *poolButton) Put(x *Button) {
	x.Text = ""
	p.pool.Put(x)
}

var PoolButton = poolButton{}

const C_ButtonUrl int64 = 1386588692

type poolButtonUrl struct {
	pool sync.Pool
}

func (p *poolButtonUrl) Get() *ButtonUrl {
	x, ok := p.pool.Get().(*ButtonUrl)
	if !ok {
		return &ButtonUrl{}
	}
	return x
}

func (p *poolButtonUrl) Put(x *ButtonUrl) {
	x.Text = ""
	x.Url = ""
	p.pool.Put(x)
}

var PoolButtonUrl = poolButtonUrl{}

const C_ButtonCallback int64 = 3657475579

type poolButtonCallback struct {
	pool sync.Pool
}

func (p *poolButtonCallback) Get() *ButtonCallback {
	x, ok := p.pool.Get().(*ButtonCallback)
	if !ok {
		return &ButtonCallback{}
	}
	return x
}

func (p *poolButtonCallback) Put(x *ButtonCallback) {
	x.Text = ""
	x.Data = x.Data[:0]
	p.pool.Put(x)
}

var PoolButtonCallback = poolButtonCallback{}

const C_ButtonRequestPhone int64 = 1977121245

type poolButtonRequestPhone struct {
	pool sync.Pool
}

func (p *poolButtonRequestPhone) Get() *ButtonRequestPhone {
	x, ok := p.pool.Get().(*ButtonRequestPhone)
	if !ok {
		return &ButtonRequestPhone{}
	}
	return x
}

func (p *poolButtonRequestPhone) Put(x *ButtonRequestPhone) {
	x.Text = ""
	p.pool.Put(x)
}

var PoolButtonRequestPhone = poolButtonRequestPhone{}

const C_ButtonRequestGeoLocation int64 = 4134316262

type poolButtonRequestGeoLocation struct {
	pool sync.Pool
}

func (p *poolButtonRequestGeoLocation) Get() *ButtonRequestGeoLocation {
	x, ok := p.pool.Get().(*ButtonRequestGeoLocation)
	if !ok {
		return &ButtonRequestGeoLocation{}
	}
	return x
}

func (p *poolButtonRequestGeoLocation) Put(x *ButtonRequestGeoLocation) {
	x.Text = ""
	p.pool.Put(x)
}

var PoolButtonRequestGeoLocation = poolButtonRequestGeoLocation{}

const C_ButtonSwitchInline int64 = 3041502501

type poolButtonSwitchInline struct {
	pool sync.Pool
}

func (p *poolButtonSwitchInline) Get() *ButtonSwitchInline {
	x, ok := p.pool.Get().(*ButtonSwitchInline)
	if !ok {
		return &ButtonSwitchInline{}
	}
	return x
}

func (p *poolButtonSwitchInline) Put(x *ButtonSwitchInline) {
	x.Text = ""
	x.Query = ""
	x.SamePeer = false
	p.pool.Put(x)
}

var PoolButtonSwitchInline = poolButtonSwitchInline{}

const C_ButtonBuy int64 = 1766878669

type poolButtonBuy struct {
	pool sync.Pool
}

func (p *poolButtonBuy) Get() *ButtonBuy {
	x, ok := p.pool.Get().(*ButtonBuy)
	if !ok {
		return &ButtonBuy{}
	}
	return x
}

func (p *poolButtonBuy) Put(x *ButtonBuy) {
	x.Text = ""
	p.pool.Put(x)
}

var PoolButtonBuy = poolButtonBuy{}

func init() {
	registry.RegisterConstructor(1353207252, "msg.ReplyKeyboardMarkup")
	registry.RegisterConstructor(3617178965, "msg.ReplyInlineMarkup")
	registry.RegisterConstructor(4235074858, "msg.ReplyKeyboardHide")
	registry.RegisterConstructor(4261878523, "msg.ReplyKeyboardForceReply")
	registry.RegisterConstructor(3268825054, "msg.KeyboardButtonRow")
	registry.RegisterConstructor(692302489, "msg.KeyboardButtonEnvelope")
	registry.RegisterConstructor(3346660205, "msg.Button")
	registry.RegisterConstructor(1386588692, "msg.ButtonUrl")
	registry.RegisterConstructor(3657475579, "msg.ButtonCallback")
	registry.RegisterConstructor(1977121245, "msg.ButtonRequestPhone")
	registry.RegisterConstructor(4134316262, "msg.ButtonRequestGeoLocation")
	registry.RegisterConstructor(3041502501, "msg.ButtonSwitchInline")
	registry.RegisterConstructor(1766878669, "msg.ButtonBuy")
}

func (x *ReplyKeyboardMarkup) DeepCopy(z *ReplyKeyboardMarkup) {
	z.SingleUse = x.SingleUse
	z.Selective = x.Selective
	z.Resize = x.Resize
	for idx := range x.Rows {
		if x.Rows[idx] != nil {
			xx := PoolKeyboardButtonRow.Get()
			x.Rows[idx].DeepCopy(xx)
			z.Rows = append(z.Rows, xx)
		}
	}
}

func (x *ReplyInlineMarkup) DeepCopy(z *ReplyInlineMarkup) {
	for idx := range x.Rows {
		if x.Rows[idx] != nil {
			xx := PoolKeyboardButtonRow.Get()
			x.Rows[idx].DeepCopy(xx)
			z.Rows = append(z.Rows, xx)
		}
	}
}

func (x *ReplyKeyboardHide) DeepCopy(z *ReplyKeyboardHide) {
	z.Selective = x.Selective
}

func (x *ReplyKeyboardForceReply) DeepCopy(z *ReplyKeyboardForceReply) {
	z.SingleUse = x.SingleUse
	z.Selective = x.Selective
}

func (x *KeyboardButtonRow) DeepCopy(z *KeyboardButtonRow) {
	for idx := range x.Buttons {
		if x.Buttons[idx] != nil {
			xx := PoolKeyboardButtonEnvelope.Get()
			x.Buttons[idx].DeepCopy(xx)
			z.Buttons = append(z.Buttons, xx)
		}
	}
}

func (x *KeyboardButtonEnvelope) DeepCopy(z *KeyboardButtonEnvelope) {
	z.Constructor = x.Constructor
	z.Data = append(z.Data[:0], x.Data...)
}

func (x *Button) DeepCopy(z *Button) {
	z.Text = x.Text
}

func (x *ButtonUrl) DeepCopy(z *ButtonUrl) {
	z.Text = x.Text
	z.Url = x.Url
}

func (x *ButtonCallback) DeepCopy(z *ButtonCallback) {
	z.Text = x.Text
	z.Data = append(z.Data[:0], x.Data...)
}

func (x *ButtonRequestPhone) DeepCopy(z *ButtonRequestPhone) {
	z.Text = x.Text
}

func (x *ButtonRequestGeoLocation) DeepCopy(z *ButtonRequestGeoLocation) {
	z.Text = x.Text
}

func (x *ButtonSwitchInline) DeepCopy(z *ButtonSwitchInline) {
	z.Text = x.Text
	z.Query = x.Query
	z.SamePeer = x.SamePeer
}

func (x *ButtonBuy) DeepCopy(z *ButtonBuy) {
	z.Text = x.Text
}

func (x *ReplyKeyboardMarkup) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_ReplyKeyboardMarkup, x)
}

func (x *ReplyInlineMarkup) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_ReplyInlineMarkup, x)
}

func (x *ReplyKeyboardHide) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_ReplyKeyboardHide, x)
}

func (x *ReplyKeyboardForceReply) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_ReplyKeyboardForceReply, x)
}

func (x *KeyboardButtonRow) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_KeyboardButtonRow, x)
}

func (x *KeyboardButtonEnvelope) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_KeyboardButtonEnvelope, x)
}

func (x *Button) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_Button, x)
}

func (x *ButtonUrl) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_ButtonUrl, x)
}

func (x *ButtonCallback) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_ButtonCallback, x)
}

func (x *ButtonRequestPhone) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_ButtonRequestPhone, x)
}

func (x *ButtonRequestGeoLocation) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_ButtonRequestGeoLocation, x)
}

func (x *ButtonSwitchInline) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_ButtonSwitchInline, x)
}

func (x *ButtonBuy) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_ButtonBuy, x)
}

func (x *ReplyKeyboardMarkup) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *ReplyInlineMarkup) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *ReplyKeyboardHide) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *ReplyKeyboardForceReply) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *KeyboardButtonRow) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *KeyboardButtonEnvelope) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *Button) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *ButtonUrl) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *ButtonCallback) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *ButtonRequestPhone) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *ButtonRequestGeoLocation) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *ButtonSwitchInline) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *ButtonBuy) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *ReplyKeyboardMarkup) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *ReplyInlineMarkup) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *ReplyKeyboardHide) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *ReplyKeyboardForceReply) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *KeyboardButtonRow) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *KeyboardButtonEnvelope) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *Button) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *ButtonUrl) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *ButtonCallback) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *ButtonRequestPhone) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *ButtonRequestGeoLocation) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *ButtonSwitchInline) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *ButtonBuy) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}
