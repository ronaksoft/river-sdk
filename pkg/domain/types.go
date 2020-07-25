package domain

import (
	"github.com/gogo/protobuf/proto"
	"golang.org/x/sync/singleflight"
	"reflect"
)

// MInt64B simple type to get distinct IDs
type MInt64B map[int64]bool

func (m MInt64B) Add(values ...int64) {
	for _, v := range values {
		m[v] = true
	}
}

func (m MInt64B) Remove(values ...int64) {
	for _, v := range values {
		delete(m, v)
	}
}

// ToArray get values
func (m MInt64B) ToArray() []int64 {
	a := make([]int64, 0, len(m))
	for k := range m {
		a = append(a, k)
	}
	return a
}

// MInt32B simple type to get distinct IDs
type MInt32B map[int32]bool

func (m MInt32B) Add(values ...int32) {
	for _, v := range values {
		m[v] = true
	}
}

func (m MInt32B) Remove(values ...int32) {
	for _, v := range values {
		delete(m, v)
	}
}

// ToArray get values
func (m MInt32B) ToArray() []int32 {
	a := make([]int32, 0, len(m))
	for k := range m {
		a = append(a, k)
	}
	return a
}

// Salt
type Slt struct {
	Value     int64 `json:"value"`
	Timestamp int64 `json:"timestamp"`
}

var SingleFlight singleflight.Group

type Proto interface {
	proto.Sizer
	proto.Marshaler
	MarshalToSizedBuffer(data []byte) (int, error)
}

type (
	M           map[string]interface{}
	MS          map[string]string
	MI          map[string]int64
	UniqueArray map[interface{}]struct{}
)

func (m UniqueArray) ToArray() interface{} {
	keys := reflect.ValueOf(m).MapKeys()
	if len(keys) == 0 {
		return reflect.ValueOf(nil)
	}
	x := reflect.MakeSlice(reflect.SliceOf(keys[0].Elem().Type()), 0, len(keys))
	for _, k := range keys {
		x = reflect.Append(x, k.Elem())
	}
	return x.Interface()
}
