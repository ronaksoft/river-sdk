// Code generated by Rony's protoc plugin; DO NOT EDIT.
// ProtoC ver. v4.25.2
// Rony ver. v0.12.22
// Source: chat.wallpaper.proto

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

const C_WallPaperGet int64 = 183906980

type poolWallPaperGet struct {
	pool sync.Pool
}

func (p *poolWallPaperGet) Get() *WallPaperGet {
	x, ok := p.pool.Get().(*WallPaperGet)
	if !ok {
		x = &WallPaperGet{}
	}

	return x
}

func (p *poolWallPaperGet) Put(x *WallPaperGet) {
	if x == nil {
		return
	}

	x.Crc32Hash = 0

	p.pool.Put(x)
}

var PoolWallPaperGet = poolWallPaperGet{}

func (x *WallPaperGet) DeepCopy(z *WallPaperGet) {
	z.Crc32Hash = x.Crc32Hash
}

func (x *WallPaperGet) Clone() *WallPaperGet {
	z := &WallPaperGet{}
	x.DeepCopy(z)
	return z
}

func (x *WallPaperGet) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{Merge: true}.Unmarshal(b, x)
}

func (x *WallPaperGet) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *WallPaperGet) UnmarshalJSON(b []byte) error {
	return protojson.Unmarshal(b, x)
}

func (x *WallPaperGet) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(x)
}

func (x *WallPaperGet) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_WallPaperGet, x)
}

const C_WallPaperSave int64 = 3559907599

type poolWallPaperSave struct {
	pool sync.Pool
}

func (p *poolWallPaperSave) Get() *WallPaperSave {
	x, ok := p.pool.Get().(*WallPaperSave)
	if !ok {
		x = &WallPaperSave{}
	}

	x.WallPaper = PoolInputWallPaper.Get()

	x.Settings = PoolWallPaperSettings.Get()

	return x
}

func (p *poolWallPaperSave) Put(x *WallPaperSave) {
	if x == nil {
		return
	}

	PoolInputWallPaper.Put(x.WallPaper)
	PoolWallPaperSettings.Put(x.Settings)

	p.pool.Put(x)
}

var PoolWallPaperSave = poolWallPaperSave{}

func (x *WallPaperSave) DeepCopy(z *WallPaperSave) {
	if x.WallPaper != nil {
		if z.WallPaper == nil {
			z.WallPaper = PoolInputWallPaper.Get()
		}
		x.WallPaper.DeepCopy(z.WallPaper)
	} else {
		PoolInputWallPaper.Put(z.WallPaper)
		z.WallPaper = nil
	}
	if x.Settings != nil {
		if z.Settings == nil {
			z.Settings = PoolWallPaperSettings.Get()
		}
		x.Settings.DeepCopy(z.Settings)
	} else {
		PoolWallPaperSettings.Put(z.Settings)
		z.Settings = nil
	}
}

func (x *WallPaperSave) Clone() *WallPaperSave {
	z := &WallPaperSave{}
	x.DeepCopy(z)
	return z
}

func (x *WallPaperSave) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{Merge: true}.Unmarshal(b, x)
}

func (x *WallPaperSave) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *WallPaperSave) UnmarshalJSON(b []byte) error {
	return protojson.Unmarshal(b, x)
}

func (x *WallPaperSave) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(x)
}

func (x *WallPaperSave) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_WallPaperSave, x)
}

const C_WallPaperDelete int64 = 3006268108

type poolWallPaperDelete struct {
	pool sync.Pool
}

func (p *poolWallPaperDelete) Get() *WallPaperDelete {
	x, ok := p.pool.Get().(*WallPaperDelete)
	if !ok {
		x = &WallPaperDelete{}
	}

	x.WallPaper = PoolInputWallPaper.Get()

	return x
}

func (p *poolWallPaperDelete) Put(x *WallPaperDelete) {
	if x == nil {
		return
	}

	PoolInputWallPaper.Put(x.WallPaper)

	p.pool.Put(x)
}

var PoolWallPaperDelete = poolWallPaperDelete{}

func (x *WallPaperDelete) DeepCopy(z *WallPaperDelete) {
	if x.WallPaper != nil {
		if z.WallPaper == nil {
			z.WallPaper = PoolInputWallPaper.Get()
		}
		x.WallPaper.DeepCopy(z.WallPaper)
	} else {
		PoolInputWallPaper.Put(z.WallPaper)
		z.WallPaper = nil
	}
}

func (x *WallPaperDelete) Clone() *WallPaperDelete {
	z := &WallPaperDelete{}
	x.DeepCopy(z)
	return z
}

func (x *WallPaperDelete) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{Merge: true}.Unmarshal(b, x)
}

func (x *WallPaperDelete) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *WallPaperDelete) UnmarshalJSON(b []byte) error {
	return protojson.Unmarshal(b, x)
}

func (x *WallPaperDelete) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(x)
}

func (x *WallPaperDelete) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_WallPaperDelete, x)
}

const C_WallPaperUpload int64 = 2661259348

type poolWallPaperUpload struct {
	pool sync.Pool
}

func (p *poolWallPaperUpload) Get() *WallPaperUpload {
	x, ok := p.pool.Get().(*WallPaperUpload)
	if !ok {
		x = &WallPaperUpload{}
	}

	x.UploadedFile = PoolInputFile.Get()

	x.File = PoolInputDocument.Get()

	x.Settings = PoolWallPaperSettings.Get()

	return x
}

func (p *poolWallPaperUpload) Put(x *WallPaperUpload) {
	if x == nil {
		return
	}

	PoolInputFile.Put(x.UploadedFile)
	PoolInputDocument.Put(x.File)
	x.MimeType = ""
	PoolWallPaperSettings.Put(x.Settings)

	p.pool.Put(x)
}

var PoolWallPaperUpload = poolWallPaperUpload{}

func (x *WallPaperUpload) DeepCopy(z *WallPaperUpload) {
	if x.UploadedFile != nil {
		if z.UploadedFile == nil {
			z.UploadedFile = PoolInputFile.Get()
		}
		x.UploadedFile.DeepCopy(z.UploadedFile)
	} else {
		PoolInputFile.Put(z.UploadedFile)
		z.UploadedFile = nil
	}
	if x.File != nil {
		if z.File == nil {
			z.File = PoolInputDocument.Get()
		}
		x.File.DeepCopy(z.File)
	} else {
		PoolInputDocument.Put(z.File)
		z.File = nil
	}
	z.MimeType = x.MimeType
	if x.Settings != nil {
		if z.Settings == nil {
			z.Settings = PoolWallPaperSettings.Get()
		}
		x.Settings.DeepCopy(z.Settings)
	} else {
		PoolWallPaperSettings.Put(z.Settings)
		z.Settings = nil
	}
}

func (x *WallPaperUpload) Clone() *WallPaperUpload {
	z := &WallPaperUpload{}
	x.DeepCopy(z)
	return z
}

func (x *WallPaperUpload) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{Merge: true}.Unmarshal(b, x)
}

func (x *WallPaperUpload) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *WallPaperUpload) UnmarshalJSON(b []byte) error {
	return protojson.Unmarshal(b, x)
}

func (x *WallPaperUpload) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(x)
}

func (x *WallPaperUpload) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_WallPaperUpload, x)
}

const C_WallPaperReset int64 = 2714244308

type poolWallPaperReset struct {
	pool sync.Pool
}

func (p *poolWallPaperReset) Get() *WallPaperReset {
	x, ok := p.pool.Get().(*WallPaperReset)
	if !ok {
		x = &WallPaperReset{}
	}

	return x
}

func (p *poolWallPaperReset) Put(x *WallPaperReset) {
	if x == nil {
		return
	}

	p.pool.Put(x)
}

var PoolWallPaperReset = poolWallPaperReset{}

func (x *WallPaperReset) DeepCopy(z *WallPaperReset) {
}

func (x *WallPaperReset) Clone() *WallPaperReset {
	z := &WallPaperReset{}
	x.DeepCopy(z)
	return z
}

func (x *WallPaperReset) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{Merge: true}.Unmarshal(b, x)
}

func (x *WallPaperReset) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *WallPaperReset) UnmarshalJSON(b []byte) error {
	return protojson.Unmarshal(b, x)
}

func (x *WallPaperReset) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(x)
}

func (x *WallPaperReset) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_WallPaperReset, x)
}

const C_InputWallPaper int64 = 4000784410

type poolInputWallPaper struct {
	pool sync.Pool
}

func (p *poolInputWallPaper) Get() *InputWallPaper {
	x, ok := p.pool.Get().(*InputWallPaper)
	if !ok {
		x = &InputWallPaper{}
	}

	return x
}

func (p *poolInputWallPaper) Put(x *InputWallPaper) {
	if x == nil {
		return
	}

	x.ID = 0
	x.AccessHash = 0

	p.pool.Put(x)
}

var PoolInputWallPaper = poolInputWallPaper{}

func (x *InputWallPaper) DeepCopy(z *InputWallPaper) {
	z.ID = x.ID
	z.AccessHash = x.AccessHash
}

func (x *InputWallPaper) Clone() *InputWallPaper {
	z := &InputWallPaper{}
	x.DeepCopy(z)
	return z
}

func (x *InputWallPaper) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{Merge: true}.Unmarshal(b, x)
}

func (x *InputWallPaper) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *InputWallPaper) UnmarshalJSON(b []byte) error {
	return protojson.Unmarshal(b, x)
}

func (x *InputWallPaper) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(x)
}

func (x *InputWallPaper) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_InputWallPaper, x)
}

const C_WallPaperSettings int64 = 1098244882

type poolWallPaperSettings struct {
	pool sync.Pool
}

func (p *poolWallPaperSettings) Get() *WallPaperSettings {
	x, ok := p.pool.Get().(*WallPaperSettings)
	if !ok {
		x = &WallPaperSettings{}
	}

	return x
}

func (p *poolWallPaperSettings) Put(x *WallPaperSettings) {
	if x == nil {
		return
	}

	x.Blur = false
	x.Motion = false
	x.BackgroundColour = 0
	x.BackgroundSecondColour = 0
	x.Opacity = 0
	x.Rotation = 0

	p.pool.Put(x)
}

var PoolWallPaperSettings = poolWallPaperSettings{}

func (x *WallPaperSettings) DeepCopy(z *WallPaperSettings) {
	z.Blur = x.Blur
	z.Motion = x.Motion
	z.BackgroundColour = x.BackgroundColour
	z.BackgroundSecondColour = x.BackgroundSecondColour
	z.Opacity = x.Opacity
	z.Rotation = x.Rotation
}

func (x *WallPaperSettings) Clone() *WallPaperSettings {
	z := &WallPaperSettings{}
	x.DeepCopy(z)
	return z
}

func (x *WallPaperSettings) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{Merge: true}.Unmarshal(b, x)
}

func (x *WallPaperSettings) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *WallPaperSettings) UnmarshalJSON(b []byte) error {
	return protojson.Unmarshal(b, x)
}

func (x *WallPaperSettings) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(x)
}

func (x *WallPaperSettings) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_WallPaperSettings, x)
}

const C_WallPaper int64 = 2527250827

type poolWallPaper struct {
	pool sync.Pool
}

func (p *poolWallPaper) Get() *WallPaper {
	x, ok := p.pool.Get().(*WallPaper)
	if !ok {
		x = &WallPaper{}
	}

	x.Document = PoolDocument.Get()

	x.Settings = PoolWallPaperSettings.Get()

	return x
}

func (p *poolWallPaper) Put(x *WallPaper) {
	if x == nil {
		return
	}

	x.ID = 0
	x.AccessHash = 0
	x.Creator = false
	x.Default = false
	x.Pattern = false
	x.Dark = false
	PoolDocument.Put(x.Document)
	PoolWallPaperSettings.Put(x.Settings)

	p.pool.Put(x)
}

var PoolWallPaper = poolWallPaper{}

func (x *WallPaper) DeepCopy(z *WallPaper) {
	z.ID = x.ID
	z.AccessHash = x.AccessHash
	z.Creator = x.Creator
	z.Default = x.Default
	z.Pattern = x.Pattern
	z.Dark = x.Dark
	if x.Document != nil {
		if z.Document == nil {
			z.Document = PoolDocument.Get()
		}
		x.Document.DeepCopy(z.Document)
	} else {
		PoolDocument.Put(z.Document)
		z.Document = nil
	}
	if x.Settings != nil {
		if z.Settings == nil {
			z.Settings = PoolWallPaperSettings.Get()
		}
		x.Settings.DeepCopy(z.Settings)
	} else {
		PoolWallPaperSettings.Put(z.Settings)
		z.Settings = nil
	}
}

func (x *WallPaper) Clone() *WallPaper {
	z := &WallPaper{}
	x.DeepCopy(z)
	return z
}

func (x *WallPaper) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{Merge: true}.Unmarshal(b, x)
}

func (x *WallPaper) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *WallPaper) UnmarshalJSON(b []byte) error {
	return protojson.Unmarshal(b, x)
}

func (x *WallPaper) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(x)
}

func (x *WallPaper) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_WallPaper, x)
}

const C_WallPapersMany int64 = 3121104857

type poolWallPapersMany struct {
	pool sync.Pool
}

func (p *poolWallPapersMany) Get() *WallPapersMany {
	x, ok := p.pool.Get().(*WallPapersMany)
	if !ok {
		x = &WallPapersMany{}
	}

	return x
}

func (p *poolWallPapersMany) Put(x *WallPapersMany) {
	if x == nil {
		return
	}

	for _, z := range x.WallPapers {
		PoolWallPaper.Put(z)
	}
	x.WallPapers = x.WallPapers[:0]
	x.Count = 0
	x.Crc32Hash = 0
	x.Empty = false

	p.pool.Put(x)
}

var PoolWallPapersMany = poolWallPapersMany{}

func (x *WallPapersMany) DeepCopy(z *WallPapersMany) {
	for idx := range x.WallPapers {
		if x.WallPapers[idx] == nil {
			continue
		}
		xx := PoolWallPaper.Get()
		x.WallPapers[idx].DeepCopy(xx)
		z.WallPapers = append(z.WallPapers, xx)
	}
	z.Count = x.Count
	z.Crc32Hash = x.Crc32Hash
	z.Empty = x.Empty
}

func (x *WallPapersMany) Clone() *WallPapersMany {
	z := &WallPapersMany{}
	x.DeepCopy(z)
	return z
}

func (x *WallPapersMany) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{Merge: true}.Unmarshal(b, x)
}

func (x *WallPapersMany) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *WallPapersMany) UnmarshalJSON(b []byte) error {
	return protojson.Unmarshal(b, x)
}

func (x *WallPapersMany) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(x)
}

func (x *WallPapersMany) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_WallPapersMany, x)
}

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

var _ = bytes.MinRead
