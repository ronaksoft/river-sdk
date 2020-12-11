package msg

import (
	edge "github.com/ronaksoft/rony/edge"
	registry "github.com/ronaksoft/rony/registry"
	proto "google.golang.org/protobuf/proto"
	sync "sync"
)

const C_GifGetSaved int64 = 351289811

type poolGifGetSaved struct {
	pool sync.Pool
}

func (p *poolGifGetSaved) Get() *GifGetSaved {
	x, ok := p.pool.Get().(*GifGetSaved)
	if !ok {
		return &GifGetSaved{}
	}
	return x
}

func (p *poolGifGetSaved) Put(x *GifGetSaved) {
	x.Hash = 0
	p.pool.Put(x)
}

var PoolGifGetSaved = poolGifGetSaved{}

const C_GifSave int64 = 1433539893

type poolGifSave struct {
	pool sync.Pool
}

func (p *poolGifSave) Get() *GifSave {
	x, ok := p.pool.Get().(*GifSave)
	if !ok {
		return &GifSave{}
	}
	return x
}

func (p *poolGifSave) Put(x *GifSave) {
	if x.Doc != nil {
		PoolInputDocument.Put(x.Doc)
		x.Doc = nil
	}
	x.Attributes = x.Attributes[:0]
	p.pool.Put(x)
}

var PoolGifSave = poolGifSave{}

const C_GifDelete int64 = 743009261

type poolGifDelete struct {
	pool sync.Pool
}

func (p *poolGifDelete) Get() *GifDelete {
	x, ok := p.pool.Get().(*GifDelete)
	if !ok {
		return &GifDelete{}
	}
	return x
}

func (p *poolGifDelete) Put(x *GifDelete) {
	if x.Doc != nil {
		PoolInputDocument.Put(x.Doc)
		x.Doc = nil
	}
	p.pool.Put(x)
}

var PoolGifDelete = poolGifDelete{}

const C_GifSearch int64 = 2729168077

type poolGifSearch struct {
	pool sync.Pool
}

func (p *poolGifSearch) Get() *GifSearch {
	x, ok := p.pool.Get().(*GifSearch)
	if !ok {
		return &GifSearch{}
	}
	return x
}

func (p *poolGifSearch) Put(x *GifSearch) {
	x.Query = ""
	x.Hash = 0
	p.pool.Put(x)
}

var PoolGifSearch = poolGifSearch{}

const C_FoundGifs int64 = 3258313539

type poolFoundGifs struct {
	pool sync.Pool
}

func (p *poolFoundGifs) Get() *FoundGifs {
	x, ok := p.pool.Get().(*FoundGifs)
	if !ok {
		return &FoundGifs{}
	}
	return x
}

func (p *poolFoundGifs) Put(x *FoundGifs) {
	x.NextOffset = 0
	x.Gifs = x.Gifs[:0]
	p.pool.Put(x)
}

var PoolFoundGifs = poolFoundGifs{}

const C_FoundGif int64 = 3984495849

type poolFoundGif struct {
	pool sync.Pool
}

func (p *poolFoundGif) Get() *FoundGif {
	x, ok := p.pool.Get().(*FoundGif)
	if !ok {
		return &FoundGif{}
	}
	return x
}

func (p *poolFoundGif) Put(x *FoundGif) {
	x.Url = ""
	if x.Doc != nil {
		PoolDocument.Put(x.Doc)
		x.Doc = nil
	}
	if x.Thumb != nil {
		PoolDocument.Put(x.Thumb)
		x.Thumb = nil
	}
	p.pool.Put(x)
}

var PoolFoundGif = poolFoundGif{}

const C_SavedGifs int64 = 1791431697

type poolSavedGifs struct {
	pool sync.Pool
}

func (p *poolSavedGifs) Get() *SavedGifs {
	x, ok := p.pool.Get().(*SavedGifs)
	if !ok {
		return &SavedGifs{}
	}
	return x
}

func (p *poolSavedGifs) Put(x *SavedGifs) {
	x.Hash = 0
	x.Docs = x.Docs[:0]
	x.NotModified = false
	p.pool.Put(x)
}

var PoolSavedGifs = poolSavedGifs{}

func init() {
	registry.RegisterConstructor(351289811, "msg.GifGetSaved")
	registry.RegisterConstructor(1433539893, "msg.GifSave")
	registry.RegisterConstructor(743009261, "msg.GifDelete")
	registry.RegisterConstructor(2729168077, "msg.GifSearch")
	registry.RegisterConstructor(3258313539, "msg.FoundGifs")
	registry.RegisterConstructor(3984495849, "msg.FoundGif")
	registry.RegisterConstructor(1791431697, "msg.SavedGifs")
}

func (x *GifGetSaved) DeepCopy(z *GifGetSaved) {
	z.Hash = x.Hash
}

func (x *GifSave) DeepCopy(z *GifSave) {
	if x.Doc != nil {
		z.Doc = PoolInputDocument.Get()
		x.Doc.DeepCopy(z.Doc)
	}
	for idx := range x.Attributes {
		if x.Attributes[idx] != nil {
			xx := PoolDocumentAttribute.Get()
			x.Attributes[idx].DeepCopy(xx)
			z.Attributes = append(z.Attributes, xx)
		}
	}
}

func (x *GifDelete) DeepCopy(z *GifDelete) {
	if x.Doc != nil {
		z.Doc = PoolInputDocument.Get()
		x.Doc.DeepCopy(z.Doc)
	}
}

func (x *GifSearch) DeepCopy(z *GifSearch) {
	z.Query = x.Query
	z.Hash = x.Hash
}

func (x *FoundGifs) DeepCopy(z *FoundGifs) {
	z.NextOffset = x.NextOffset
	for idx := range x.Gifs {
		if x.Gifs[idx] != nil {
			xx := PoolFoundGif.Get()
			x.Gifs[idx].DeepCopy(xx)
			z.Gifs = append(z.Gifs, xx)
		}
	}
}

func (x *FoundGif) DeepCopy(z *FoundGif) {
	z.Url = x.Url
	if x.Doc != nil {
		z.Doc = PoolDocument.Get()
		x.Doc.DeepCopy(z.Doc)
	}
	if x.Thumb != nil {
		z.Thumb = PoolDocument.Get()
		x.Thumb.DeepCopy(z.Thumb)
	}
}

func (x *SavedGifs) DeepCopy(z *SavedGifs) {
	z.Hash = x.Hash
	for idx := range x.Docs {
		if x.Docs[idx] != nil {
			xx := PoolMediaDocument.Get()
			x.Docs[idx].DeepCopy(xx)
			z.Docs = append(z.Docs, xx)
		}
	}
	z.NotModified = x.NotModified
}

func (x *GifGetSaved) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_GifGetSaved, x)
}

func (x *GifSave) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_GifSave, x)
}

func (x *GifDelete) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_GifDelete, x)
}

func (x *GifSearch) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_GifSearch, x)
}

func (x *FoundGifs) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_FoundGifs, x)
}

func (x *FoundGif) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_FoundGif, x)
}

func (x *SavedGifs) PushToContext(ctx *edge.RequestCtx) {
	ctx.PushMessage(C_SavedGifs, x)
}

func (x *GifGetSaved) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *GifSave) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *GifDelete) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *GifSearch) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *FoundGifs) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *FoundGif) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *SavedGifs) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *GifGetSaved) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *GifSave) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *GifDelete) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *GifSearch) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *FoundGifs) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *FoundGif) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}

func (x *SavedGifs) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{}.Unmarshal(b, x)
}
