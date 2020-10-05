package tools

import (
	"bytes"
	"encoding/binary"
	"git.ronaksoft.com/river/sdk/internal/pools"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

/*
   Creation Time: 2020 - May - 06
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

func AppendStrInt64(sb *strings.Builder, x int64) {
	b := pools.TinyBytes.GetLen(8)
	binary.BigEndian.PutUint64(b, uint64(x))
	sb.Write(b)
	pools.TinyBytes.Put(b)
	return
}

func AppendStrUInt64(sb *strings.Builder, x uint64) {
	b := pools.TinyBytes.GetLen(8)
	binary.BigEndian.PutUint64(b, x)
	sb.Write(b)
	pools.TinyBytes.Put(b)
	return
}

func AppendStrInt32(sb *strings.Builder, x int32) {
	b := pools.TinyBytes.GetLen(4)
	binary.BigEndian.PutUint32(b, uint32(x))
	sb.Write(b)
	pools.TinyBytes.Put(b)
	return
}

func AppendStrUInt32(sb *strings.Builder, x uint32) {
	b := pools.TinyBytes.GetLen(4)
	binary.BigEndian.PutUint32(b, x)
	sb.Write(b)
	pools.TinyBytes.Put(b)
	return
}

func StrToInt64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

func StrToInt32(s string) int32 {
	v, _ := strconv.ParseInt(s, 10, 32)
	return int32(v)
}

func StrToInt(s string) int {
	v, _ := strconv.ParseInt(s, 10, 32)
	return int(v)
}

func StrToFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
func StrToUInt64(s string) uint64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return uint64(v)
}

func StrToUInt32(s string) uint32 {
	v, _ := strconv.ParseInt(s, 10, 32)
	return uint32(v)
}

func Int64ToStr(x int64) string {
	return strconv.FormatInt(x, 10)
}

func UInt64ToStr(x uint64) string {
	return strconv.FormatUint(x, 10)
}

func Int32ToStr(x int32) string {
	return strconv.FormatInt(int64(x), 10)
}

// ByteToStr converts byte slice to a string without memory allocation.
// Note it may break if string and/or slice header will change
// in the future go versions.
func ByteToStr(bts []byte) string {
	b := *(*reflect.SliceHeader)(unsafe.Pointer(&bts))
	s := &reflect.StringHeader{Data: b.Data, Len: b.Len}
	return *(*string)(unsafe.Pointer(s))
}

func ByteToUInt32(bts []byte) uint32 {
	buff := bytes.NewReader(bts)
	var nowVar uint32
	binary.Read(buff, binary.BigEndian, &nowVar)
	return nowVar
}

// StrToByte converts string to a byte slice without memory allocation.
// Note it may break if string and/or slice header will change
// in the future go versions.
func StrToByte(str string) []byte {
	s := *(*reflect.StringHeader)(unsafe.Pointer(&str))
	b := &reflect.SliceHeader{Data: s.Data, Len: s.Len, Cap: s.Len}
	return *(*[]byte)(unsafe.Pointer(b))
}

func UInt32ToByte(i uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, i)
	return b
}
