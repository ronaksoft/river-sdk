// Code generated by Rony's protoc plugin; DO NOT EDIT.
// ProtoC ver. v3.15.8
// Rony ver. v0.12.19
// Source: imsg.proto

package msg

import (
	bytes "bytes"
	rony "github.com/ronaksoft/rony"
	pools "github.com/ronaksoft/rony/pools"
	registry "github.com/ronaksoft/rony/registry"
	store "github.com/ronaksoft/rony/store"
	tools "github.com/ronaksoft/rony/tools"
	protojson "google.golang.org/protobuf/encoding/protojson"
	proto "google.golang.org/protobuf/proto"
	sync "sync"
)

var _ = pools.Imported

const C_GetPage int64 = 3721890413

type poolGetPage struct {
	pool sync.Pool
}

func (p *poolGetPage) Get() *GetPage {
	x, ok := p.pool.Get().(*GetPage)
	if !ok {
		x = &GetPage{}
	}

	return x
}

func (p *poolGetPage) Put(x *GetPage) {
	if x == nil {
		return
	}

	x.PageID = 0
	x.ReplicaSet = 0

	p.pool.Put(x)
}

var PoolGetPage = poolGetPage{}

func (x *GetPage) DeepCopy(z *GetPage) {
	z.PageID = x.PageID
	z.ReplicaSet = x.ReplicaSet
}

func (x *GetPage) Clone() *GetPage {
	z := &GetPage{}
	x.DeepCopy(z)
	return z
}

func (x *GetPage) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{Merge: true}.Unmarshal(b, x)
}

func (x *GetPage) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *GetPage) UnmarshalJSON(b []byte) error {
	return protojson.Unmarshal(b, x)
}

func (x *GetPage) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(x)
}

const C_TunnelMessage int64 = 3271476222

type poolTunnelMessage struct {
	pool sync.Pool
}

func (p *poolTunnelMessage) Get() *TunnelMessage {
	x, ok := p.pool.Get().(*TunnelMessage)
	if !ok {
		x = &TunnelMessage{}
	}

	x.Envelope = rony.PoolMessageEnvelope.Get()

	return x
}

func (p *poolTunnelMessage) Put(x *TunnelMessage) {
	if x == nil {
		return
	}

	x.SenderID = x.SenderID[:0]
	x.SenderReplicaSet = 0
	for _, z := range x.Store {
		rony.PoolKeyValue.Put(z)
	}
	x.Store = x.Store[:0]
	rony.PoolMessageEnvelope.Put(x.Envelope)

	p.pool.Put(x)
}

var PoolTunnelMessage = poolTunnelMessage{}

func (x *TunnelMessage) DeepCopy(z *TunnelMessage) {
	z.SenderID = append(z.SenderID[:0], x.SenderID...)
	z.SenderReplicaSet = x.SenderReplicaSet
	for idx := range x.Store {
		if x.Store[idx] == nil {
			continue
		}
		xx := rony.PoolKeyValue.Get()
		x.Store[idx].DeepCopy(xx)
		z.Store = append(z.Store, xx)
	}
	if x.Envelope != nil {
		if z.Envelope == nil {
			z.Envelope = rony.PoolMessageEnvelope.Get()
		}
		x.Envelope.DeepCopy(z.Envelope)
	} else {
		rony.PoolMessageEnvelope.Put(z.Envelope)
		z.Envelope = nil
	}
}

func (x *TunnelMessage) Clone() *TunnelMessage {
	z := &TunnelMessage{}
	x.DeepCopy(z)
	return z
}

func (x *TunnelMessage) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{Merge: true}.Unmarshal(b, x)
}

func (x *TunnelMessage) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *TunnelMessage) UnmarshalJSON(b []byte) error {
	return protojson.Unmarshal(b, x)
}

func (x *TunnelMessage) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(x)
}

const C_EdgeNode int64 = 999040174

type poolEdgeNode struct {
	pool sync.Pool
}

func (p *poolEdgeNode) Get() *EdgeNode {
	x, ok := p.pool.Get().(*EdgeNode)
	if !ok {
		x = &EdgeNode{}
	}

	return x
}

func (p *poolEdgeNode) Put(x *EdgeNode) {
	if x == nil {
		return
	}

	x.ServerID = x.ServerID[:0]
	x.ReplicaSet = 0
	x.Hash = 0
	x.GatewayAddr = x.GatewayAddr[:0]
	x.TunnelAddr = x.TunnelAddr[:0]

	p.pool.Put(x)
}

var PoolEdgeNode = poolEdgeNode{}

func (x *EdgeNode) DeepCopy(z *EdgeNode) {
	z.ServerID = append(z.ServerID[:0], x.ServerID...)
	z.ReplicaSet = x.ReplicaSet
	z.Hash = x.Hash
	z.GatewayAddr = append(z.GatewayAddr[:0], x.GatewayAddr...)
	z.TunnelAddr = append(z.TunnelAddr[:0], x.TunnelAddr...)
}

func (x *EdgeNode) Clone() *EdgeNode {
	z := &EdgeNode{}
	x.DeepCopy(z)
	return z
}

func (x *EdgeNode) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{Merge: true}.Unmarshal(b, x)
}

func (x *EdgeNode) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *EdgeNode) UnmarshalJSON(b []byte) error {
	return protojson.Unmarshal(b, x)
}

func (x *EdgeNode) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(x)
}

const C_Page int64 = 3023575326

type poolPage struct {
	pool sync.Pool
}

func (p *poolPage) Get() *Page {
	x, ok := p.pool.Get().(*Page)
	if !ok {
		x = &Page{}
	}

	return x
}

func (p *poolPage) Put(x *Page) {
	if x == nil {
		return
	}

	x.ID = 0
	x.ReplicaSet = 0

	p.pool.Put(x)
}

var PoolPage = poolPage{}

func (x *Page) DeepCopy(z *Page) {
	z.ID = x.ID
	z.ReplicaSet = x.ReplicaSet
}

func (x *Page) Clone() *Page {
	z := &Page{}
	x.DeepCopy(z)
	return z
}

func (x *Page) Unmarshal(b []byte) error {
	return proto.UnmarshalOptions{Merge: true}.Unmarshal(b, x)
}

func (x *Page) Marshal() ([]byte, error) {
	return proto.Marshal(x)
}

func (x *Page) UnmarshalJSON(b []byte) error {
	return protojson.Unmarshal(b, x)
}

func (x *Page) MarshalJSON() ([]byte, error) {
	return protojson.Marshal(x)
}

func init() {
	registry.RegisterConstructor(3721890413, "GetPage")
	registry.RegisterConstructor(3271476222, "TunnelMessage")
	registry.RegisterConstructor(999040174, "EdgeNode")
	registry.RegisterConstructor(3023575326, "Page")
}

var _ = bytes.MinRead

type PagePrimaryKey interface {
	makePagePrivate()
}

type PagePK struct {
	ID uint32
}

func (PagePK) makePagePrivate() {}

type PageReplicaSetIDPK struct {
	ReplicaSet uint64
	ID         uint32
}

func (PageReplicaSetIDPK) makePagePrivate() {}

type PageLocalRepo struct {
	s rony.Store
}

func NewPageLocalRepo(s rony.Store) *PageLocalRepo {
	return &PageLocalRepo{
		s: s,
	}
}

func (r *PageLocalRepo) CreateWithTxn(txn *rony.StoreTxn, alloc *tools.Allocator, m *Page) (err error) {
	if alloc == nil {
		alloc = tools.NewAllocator()
		defer alloc.ReleaseAll()
	}
	key := alloc.Gen('M', C_Page, 299066170, m.ID)
	if store.ExistsByKey(txn, alloc, key) {
		return store.ErrAlreadyExists
	}

	// save table entry
	val := alloc.Marshal(m)
	err = store.SetByKey(txn, val, key)
	if err != nil {
		return
	}

	// save view entry
	err = store.Set(txn, alloc, val, 'M', C_Page, 1040696757, m.ReplicaSet, m.ID)
	if err != nil {
		return err
	}

	return
}

func (r *PageLocalRepo) Create(m *Page) error {
	alloc := tools.NewAllocator()
	defer alloc.ReleaseAll()
	return r.s.Update(func(txn *rony.StoreTxn) error {
		return r.CreateWithTxn(txn, alloc, m)
	})
}

func (r *PageLocalRepo) UpdateWithTxn(txn *rony.StoreTxn, alloc *tools.Allocator, m *Page) error {
	if alloc == nil {
		alloc = tools.NewAllocator()
		defer alloc.ReleaseAll()
	}

	err := r.DeleteWithTxn(txn, alloc, m.ID)
	if err != nil {
		return err
	}

	return r.CreateWithTxn(txn, alloc, m)
}

func (r *PageLocalRepo) Update(id uint32, m *Page) error {
	alloc := tools.NewAllocator()
	defer alloc.ReleaseAll()

	if m == nil {
		return store.ErrEmptyObject
	}

	err := r.s.Update(func(txn *rony.StoreTxn) (err error) {
		return r.UpdateWithTxn(txn, alloc, m)
	})

	return err
}

func (r *PageLocalRepo) SaveWithTxn(txn *rony.StoreTxn, alloc *tools.Allocator, m *Page) (err error) {
	if store.Exists(txn, alloc, 'M', C_Page, 299066170, m.ID) {
		return r.UpdateWithTxn(txn, alloc, m)
	} else {
		return r.CreateWithTxn(txn, alloc, m)
	}
}

func (r *PageLocalRepo) Save(m *Page) error {
	alloc := tools.NewAllocator()
	defer alloc.ReleaseAll()

	return r.s.Update(func(txn *rony.StoreTxn) error {
		return r.SaveWithTxn(txn, alloc, m)
	})
}

func (r *PageLocalRepo) ReadWithTxn(txn *rony.StoreTxn, alloc *tools.Allocator, id uint32, m *Page) (*Page, error) {
	if alloc == nil {
		alloc = tools.NewAllocator()
		defer alloc.ReleaseAll()
	}

	err := store.Unmarshal(txn, alloc, m, 'M', C_Page, 299066170, id)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *PageLocalRepo) Read(id uint32, m *Page) (*Page, error) {
	alloc := tools.NewAllocator()
	defer alloc.ReleaseAll()

	if m == nil {
		m = &Page{}
	}

	err := r.s.View(func(txn *rony.StoreTxn) (err error) {
		m, err = r.ReadWithTxn(txn, alloc, id, m)
		return err
	})
	return m, err
}

func (r *PageLocalRepo) ReadByReplicaSetIDWithTxn(
	txn *rony.StoreTxn, alloc *tools.Allocator,
	replicaSet uint64, id uint32, m *Page,
) (*Page, error) {
	if alloc == nil {
		alloc = tools.NewAllocator()
		defer alloc.ReleaseAll()
	}

	err := store.Unmarshal(txn, alloc, m, 'M', C_Page, 1040696757, replicaSet, id)
	if err != nil {
		return nil, err
	}
	return m, err
}

func (r *PageLocalRepo) ReadByReplicaSetID(replicaSet uint64, id uint32, m *Page) (*Page, error) {
	alloc := tools.NewAllocator()
	defer alloc.ReleaseAll()

	if m == nil {
		m = &Page{}
	}

	err := r.s.View(func(txn *rony.StoreTxn) (err error) {
		m, err = r.ReadByReplicaSetIDWithTxn(txn, alloc, replicaSet, id, m)
		return err
	})
	return m, err
}

func (r *PageLocalRepo) DeleteWithTxn(txn *rony.StoreTxn, alloc *tools.Allocator, id uint32) error {
	if alloc == nil {
		alloc = tools.NewAllocator()
		defer alloc.ReleaseAll()
	}

	m := &Page{}
	err := store.Unmarshal(txn, alloc, m, 'M', C_Page, 299066170, id)
	if err != nil {
		return err
	}
	err = store.Delete(txn, alloc, 'M', C_Page, 299066170, m.ID)
	if err != nil {
		return err
	}

	err = store.Delete(txn, alloc, 'M', C_Page, 1040696757, m.ReplicaSet, m.ID)
	if err != nil {
		return err
	}

	return nil
}

func (r *PageLocalRepo) Delete(id uint32) error {
	alloc := tools.NewAllocator()
	defer alloc.ReleaseAll()

	return r.s.Update(func(txn *rony.StoreTxn) error {
		return r.DeleteWithTxn(txn, alloc, id)
	})
}

func (r *PageLocalRepo) ListWithTxn(
	txn *rony.StoreTxn, alloc *tools.Allocator, offset PagePrimaryKey, lo *store.ListOption, cond func(m *Page) bool,
) ([]*Page, error) {
	if alloc == nil {
		alloc = tools.NewAllocator()
		defer alloc.ReleaseAll()
	}

	var seekKey []byte
	opt := store.DefaultIteratorOptions
	opt.Reverse = lo.Backward()
	res := make([]*Page, 0, lo.Limit())

	switch offset := offset.(type) {
	case PagePK:
		opt.Prefix = alloc.Gen('M', C_Page, 299066170, offset.ID)
		seekKey = alloc.Gen('M', C_Page, 299066170, offset.ID)

	case PageReplicaSetIDPK:
		opt.Prefix = alloc.Gen('M', C_Page, 1040696757, offset.ReplicaSet)
		seekKey = alloc.Gen('M', C_Page, 1040696757, offset.ReplicaSet, offset.ID)

	default:
		opt.Prefix = alloc.Gen('M', C_Page, 299066170)
		seekKey = opt.Prefix
	}

	err := r.s.View(func(txn *rony.StoreTxn) (err error) {
		iter := txn.NewIterator(opt)
		offset := lo.Skip()
		limit := lo.Limit()
		for iter.Seek(seekKey); iter.ValidForPrefix(opt.Prefix); iter.Next() {
			if offset--; offset >= 0 {
				continue
			}
			if limit--; limit < 0 {
				break
			}
			err = iter.Item().Value(func(val []byte) error {
				m := &Page{}
				err := m.Unmarshal(val)
				if err != nil {
					return err
				}
				if cond == nil || cond(m) {
					res = append(res, m)
				} else {
					limit++
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		iter.Close()
		return
	})

	return res, err
}

func (r *PageLocalRepo) List(
	pk PagePrimaryKey, lo *store.ListOption, cond func(m *Page) bool,
) ([]*Page, error) {
	alloc := tools.NewAllocator()
	defer alloc.ReleaseAll()

	var (
		res []*Page
		err error
	)
	err = r.s.View(func(txn *rony.StoreTxn) error {
		res, err = r.ListWithTxn(txn, alloc, pk, lo, cond)
		return err
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *PageLocalRepo) IterWithTxn(
	txn *rony.StoreTxn, alloc *tools.Allocator, offset PagePrimaryKey, ito *store.IterOption, cb func(m *Page) bool,
) error {
	if alloc == nil {
		alloc = tools.NewAllocator()
		defer alloc.ReleaseAll()
	}

	var seekKey []byte
	opt := store.DefaultIteratorOptions
	opt.Reverse = ito.Backward()

	switch offset := offset.(type) {
	case PagePK:
		opt.Prefix = alloc.Gen('M', C_Page, 299066170, offset.ID)
		seekKey = alloc.Gen('M', C_Page, 299066170, offset.ID)

	case PageReplicaSetIDPK:
		opt.Prefix = alloc.Gen('M', C_Page, 1040696757, offset.ReplicaSet)
		seekKey = alloc.Gen('M', C_Page, 1040696757, offset.ReplicaSet, offset.ID)

	default:
		opt.Prefix = alloc.Gen('M', C_Page, 299066170)
		seekKey = opt.Prefix
	}

	err := r.s.View(func(txn *rony.StoreTxn) (err error) {
		iter := txn.NewIterator(opt)
		if ito.OffsetKey() == nil {
			iter.Seek(seekKey)
		} else {
			iter.Seek(ito.OffsetKey())
		}
		exitLoop := false
		for ; iter.ValidForPrefix(opt.Prefix); iter.Next() {
			err = iter.Item().Value(func(val []byte) error {
				m := &Page{}
				err := m.Unmarshal(val)
				if err != nil {
					return err
				}
				if !cb(m) {
					exitLoop = true
				}
				return nil
			})
			if err != nil || exitLoop {
				break
			}
		}
		iter.Close()

		return
	})

	return err
}

func (r *PageLocalRepo) Iter(
	pk PagePrimaryKey, ito *store.IterOption, cb func(m *Page) bool,
) error {
	alloc := tools.NewAllocator()
	defer alloc.ReleaseAll()

	return r.s.View(func(txn *rony.StoreTxn) error {
		return r.IterWithTxn(txn, alloc, pk, ito, cb)
	})
}