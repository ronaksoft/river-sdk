package minirepo

import (
	"encoding/binary"
	"github.com/boltdb/bolt"
	"github.com/ronaksoft/rony/store"
)

/*
   Creation Time: 2021 - May - 07
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

var (
	bucketGenerals = []byte("GEN")
)

type repoGenerals struct {
	*repository
}

func newGeneral(r *repository) *repoGenerals {
	rd := &repoGenerals{
		repository: r,
	}
	return rd
}

func (d *repoGenerals) Save(key, val []byte) error {
	alloc := store.NewAllocator()
	defer alloc.ReleaseAll()

	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketGenerals)
		err := b.Put(key, val)
		if err != nil {
			return err
		}
		return nil
	})
}

func (d *repoGenerals) SaveInt64(key []byte, val int64) error {
	alloc := store.NewAllocator()
	defer alloc.ReleaseAll()

	return d.db.Update(func(tx *bolt.Tx) error {
		var valB [8]byte
		binary.BigEndian.PutUint64(valB[:], uint64(val))
		b := tx.Bucket(bucketGenerals)
		err := b.Put(key, valB[:])
		if err != nil {
			return err
		}
		return nil
	})
}

func (d *repoGenerals) SaveUInt64(key []byte, val uint64) error {
	alloc := store.NewAllocator()
	defer alloc.ReleaseAll()

	return d.db.Update(func(tx *bolt.Tx) error {
		var valB [8]byte
		binary.BigEndian.PutUint64(valB[:], val)
		b := tx.Bucket(bucketGenerals)
		err := b.Put(key, valB[:])
		if err != nil {
			return err
		}
		return nil
	})
}

func (d *repoGenerals) SaveInt32(key []byte, val int32) error {
	alloc := store.NewAllocator()
	defer alloc.ReleaseAll()

	return d.db.Update(func(tx *bolt.Tx) error {
		var valB [4]byte
		binary.BigEndian.PutUint32(valB[:], uint32(val))
		b := tx.Bucket(bucketGenerals)
		err := b.Put(key, valB[:])
		if err != nil {
			return err
		}
		return nil
	})
}

func (d *repoGenerals) SaveUInt32(key []byte, val uint32) error {
	alloc := store.NewAllocator()
	defer alloc.ReleaseAll()

	return d.db.Update(func(tx *bolt.Tx) error {
		var valB [4]byte
		binary.BigEndian.PutUint32(valB[:], val)
		b := tx.Bucket(bucketGenerals)
		err := b.Put(key, valB[:])
		if err != nil {
			return err
		}
		return nil
	})
}

func (d *repoGenerals) Get(key []byte) []byte {
	var v []byte
	_ = d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketGenerals)
		v = append(v, b.Get(key)...)
		return nil
	})
	return v
}

func (d *repoGenerals) GetInt64(key []byte) int64 {
	v := d.Get(key)
	return int64(binary.BigEndian.Uint64(v))
}

func (d *repoGenerals) GetUInt64(key []byte) uint64 {
	v := d.Get(key)
	return binary.BigEndian.Uint64(v)
}

func (d *repoGenerals) GetInt32(key []byte) int32 {
	v := d.Get(key)
	return int32(binary.BigEndian.Uint32(v))
}

func (d *repoGenerals) GetUInt32(key []byte) uint32 {
	v := d.Get(key)
	return binary.BigEndian.Uint32(v)
}
