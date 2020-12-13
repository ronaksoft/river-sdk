package msg

import (
	edge "github.com/ronaksoft/rony/edge"
	registry "github.com/ronaksoft/rony/registry"
	proto "google.golang.org/protobuf/proto"
	sync "sync"
)

const C_WallPaperGet int64 = 183906980

type poolWallPaperGet struct {
	pool sync.Pool
}

func (p *poolWallPaperGet) Get() *WallPaperGet {
	x, ok := p.pool.Get().(*WallPaperGet)
	if !ok {
		return &WallPaperGet{}
	}
	return x
}

func (p *poolWallPaperGet) Put(x *WallPaperGet) {
	x.Crc32Hash = 0
	p.pool.Put(x)
}

var PoolWallPaperGet = poolWallPaperGet{}

const C_WallPaperSave int64 = 3559907599

type poolWallPaperSave struct {
	pool sync.Pool
}

func (p *poolWallPaperSave) Get() *WallPaperSave {
	x, ok := p.pool.Get().(*WallPaperSave)
	if !ok {
		return &WallPaperSave{}
	}
	return x
}

func (p *poolWallPaperSave) Put(x *WallPaperSave) {
	if x.WallPaper != nil {
		PoolInputWallPaper.Put(x.WallPaper)
		x.WallPaper = nil
	}
	if x.Settings != nil {
		PoolWallPaperSettings.Put(x.Settings)
		x.Settings = nil
	}
	p.pool.Put(x)
}

var PoolWallPaperSave = poolWallPaperSave{}

const C_WallPaperDelete int64 = 3006268108

type poolWallPaperDelete struct {
	pool sync.Pool
}

func (p *poolWallPaperDelete) Get() *WallPaperDelete {
	x, ok := p.pool.Get().(*WallPaperDelete)
	if !ok {
		return &WallPaperDelete{}
	}
	return x
}

func (p *poolWallPaperDelete) Put(x *WallPaperDelete) {
	if x.WallPaper != nil {
		PoolInputWallPaper.Put(x.WallPaper)
		x.WallPaper = nil
	}
	p.pool.Put(x)
}

var PoolWallPaperDelete = poolWallPaperDelete{}

const C_WallPaperUpload int64 = 2661259348

type poolWallPaperUpload struct {
	pool sync.Pool
}

func (p *poolWallPaperUpload) Get() *WallPaperUpload {
	x, ok := p.pool.Get().(*WallPaperUpload)
	if !ok {
		return &WallPaperUpload{}
	}
	return x
}

func (p *poolWallPaperUpload) Put(x *WallPaperUpload) {
	if x.UploadedFile != nil {
		PoolInputFile.Put(x.UploadedFile)
		x.UploadedFile = nil
	}
	if x.File != nil {
		PoolInputDocument.Put(x.File)
		x.File = nil
	}
	x.MimeType = ""
	if x.Settings != nil {
		PoolWallPaperSettings.Put(x.Settings)
		x.Settings = nil
	}
	p.pool.Put(x)
}

var PoolWallPaperUpload = poolWallPaperUpload{}

const C_WallPaperReset int64 = 2714244308

type poolWallPaperReset struct {
	pool sync.Pool
}

func (p *poolWallPaperReset) Get() *WallPaperReset {
	x, ok := p.pool.Get().(*WallPaperReset)
	if !ok {
		return &WallPaperReset{}
	}
	return x
}

func (p *poolWallPaperReset) Put(x *WallPaperReset) {
	p.pool.Put(x)
}

var PoolWallPaperReset = poolWallPaperReset{}

const C_InputWallPaper int64 = 4000784410

type poolInputWallPaper struct {
	pool sync.Pool
}

func (p *poolInputWallPaper) Get() *InputWallPaper {
	x, ok := p.pool.Get().(*InputWallPaper)
	if !ok {
		return &InputWallPaper{}
	}
	return x
}

func (p *poolInputWallPaper) Put(x *InputWallPaper) {
	x.ID = 0
	x.AccessHash = 0
	p.pool.Put(x)
}

var PoolInputWallPaper = poolInputWallPaper{}

const C_WallPaperSettings int64 = 1098244882

type poolWallPaperSettings struct {
	pool sync.Pool
}

func (p *poolWallPaperSettings) Get() *WallPaperSettings {
	x, ok := p.pool.Get().(*WallPaperSettings)
	if !ok {
		return &WallPaperSettings{}
	}
	return x
}

func (p *poolWallPaperSettings) Put(x *WallPaperSettings) {
	x.Blur = false
	x.Motion = false
	x.BackgroundColour = 0
	x.BackgroundSecondColour = 0
	x.Opacity = 0
	x.Rotation = 0
	p.pool.Put(x)
}

var PoolWallPaperSettings = poolWallPaperSettings{}

const C_WallPaper int64 = 2527250827

type poolWallPaper struct {
	pool sync.Pool
}

func (p *poolWallPaper) Get() *WallPaper {
	x, ok := p.pool.Get().(*WallPaper)
	if !ok {
		return &WallPaper{}
	}
	return x
}

func (p *poolWallPaper) Put(x *WallPaper) {
	x.ID = 0
	x.AccessHash = 0
	x.Creator = false
	x.Default = false
	x.Pattern = false
	x.Dark = false
	if x.Document != nil {
		PoolDocument.Put(x.Document)
		x.Document = nil
	}
	if x.Settings != nil {
		PoolWallPaperSettings.Put(x.Settings)
		x.Settings = nil
	}
	p.pool.Put(x)
}

var PoolWallPaper = poolWallPaper{}

const C_WallPapersMany int64 = 3121104857

type poolWallPapersMany struct {
	pool sync.Pool
}

func (p *poolWallPapersMany) Get() *WallPapersMany {
	x, ok := p.pool.Get().(*WallPapersMany)
	if !ok {
		return &WallPapersMany{}
	}
	return x
}

func (p *poolWallPapersMany) Put(x *WallPapersMany) {
	x.WallPapers = x.WallPapers[:0]
	x.Count = 0
	x.Crc32Hash = 0
	x.Empty = false
	p.pool.Put(x)
}

var PoolWallPapersMany = poolWallPapersMany{}

func init() {
	registry.RegisterConstructor(183906980, "WallPaperGet")
	registry.RegisterConstructor(3559907599, "WallPaperSave")
	registry.RegisterConstructor(3006268108, "WallPaperDelete")
	registry.RegisterConstructor(2661259348, "WallPaperUpload")
	registry.RegisterConstructor(2714244308, "WallPaperReset")
	registry.RegisterConstructor(4000784410, "InputWallPaper")
	registry.RegisterConstructor(1098244882, "WallPaperSettings")
	registry.RegisterConstructor(2527250827, "WallPaper")
	registry.RegisterConstructor(3121104857, "WallPapersMany")
}

func (x *WallPaperGet) DeepCopy(z *WallPaperGet) {
	z.Crc32Hash = x.Crc32Hash
}

func (x *WallPaperSave) DeepCopy(z *WallPaperSave) {
	if x.WallPaper != nil {
		z.WallPaper = PoolInputWallPaper.Get()
		x.WallPaper.DeepCopy(z.WallPaper)
	}
	if x.Settings != nil {
		z.Settings = PoolWallPaperSettings.Get()
		x.Settings.DeepCopy(z.Settings)
	}
}

func (x *WallPaperDelete) DeepCopy(z *WallPaperDelete) {
	if x.WallPaper != nil {
		z.WallPaper = PoolInputWallPaper.Get()
		x.WallPaper.DeepCopy(z.WallPaper)
	}
}

func (x *WallPaperUpload) DeepCopy(z *WallPaperUpload) {
	if x.UploadedFile != nil {
		z.UploadedFile = PoolInputFile.Get()
		x.UploadedFile.DeepCopy(z.UploadedFile)
	}
	if x.File != nil {
		z.File = PoolInputDocument.Get()
		x.File.DeepCopy(z.File)
	}
	z.MimeType = x.MimeType
	if x.Settings != nil {
		z.Settings = PoolWallPaperSettings.Get()
		x.Settings.DeepCopy(z.Settings)
	}
}

func (x *WallPaperReset) DeepCopy(z *WallPaperReset) {
}

func (x *InputWallPaper) DeepCopy(z *InputWallPaper) {
	z.ID = x.ID
	z.AccessHash = x.AccessHash
}

func (x *WallPaperSettings) DeepCopy(z *WallPaperSettings) {
	z.Blur = x.Blur
	z.Motion = x.Motion
	z.BackgroundColour = x.BackgroundColour
	z.BackgroundSecondColour = x.BackgroundSecondColour
	z.Opacity = x.Opacity
	z.Rotation = x.Rotation
}

func (x *WallPaper) DeepCopy(z *WallPaper) {
	z.ID = x.ID
	z.AccessHash = x.AccessHash
	z.Creator = x.Creator
	z.Default = x.Default
	z.Pattern = x.Pattern
	z.Dark = x.Dark
	if x.Document != nil {
		z.Document = PoolDocument.Get()
		x.Document.DeepCopy(z.Document)
	}
	if x.Settings != nil {
		z.Settings = PoolWallPaperSettings.Get()
		x.Settings.DeepCopy(z.Settings)
	}
}

func (x *WallPapersMany) DeepCopy(z *WallPapersMany) {
	for idx := range x.WallPapers {
		if x.WallPapers[idx] != nil {
			xx := PoolWallPaper.Get()
			x.WallPapers[idx].DeepCopy(xx)
			z.WallPapers = append(z.WallPapers, xx)
		}
	}
	z.Count = x.Count
	z.Crc32Hash = x.Crc32Hash
	z.Empty = x.Empty
}

func (x *WallPaperGet) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_WallPaperGet, x)
}

func (x *WallPaperSave) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_WallPaperSave, x)
}

func (x *WallPaperDelete) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_WallPaperDelete, x)
}

func (x *WallPaperUpload) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_WallPaperUpload, x)
}

func (x *WallPaperReset) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_WallPaperReset, x)
}

func (x *InputWallPaper) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_InputWallPaper, x)
}

func (x *WallPaperSettings) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_WallPaperSettings, x)
}

func (x *WallPaper) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_WallPaper, x)
}

func (x *WallPapersMany) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_WallPapersMany, x)
}

func (x *WallPaperGet) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *WallPaperSave) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *WallPaperDelete) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *WallPaperUpload) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *WallPaperReset) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *InputWallPaper) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *WallPaperSettings) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *WallPaper) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *WallPapersMany) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *WallPaperGet) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *WallPaperSave) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *WallPaperDelete) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *WallPaperUpload) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *WallPaperReset) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *InputWallPaper) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *WallPaperSettings) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *WallPaper) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *WallPapersMany) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}
