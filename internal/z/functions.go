package z

import (
	"bytes"
	"encoding/binary"
	"github.com/ronaksoft/rony/pools"
	"reflect"
	"strings"
)

/*
   Creation Time: 2019 - Oct - 13
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

func DeleteItemFromArray(slice interface{}, index int) {
	v := reflect.ValueOf(slice)
	if v.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
	}
	vLength := v.Len()
	if v.Kind() != reflect.Slice {
		panic("slice is not valid")
	}
	if index >= vLength || index < 0 {
		panic("invalid index")
	}
	switch vLength {
	case 1:
		v.SetLen(0)
	default:
		v.Index(index).Set(v.Index(v.Len() - 1))
		v.SetLen(vLength - 1)
	}
}

func TruncateString(str string, l int) string {
	n := 0
	for i := range str {
		n++
		if n > l {
			return str[:i]
		}
	}
	return str
}

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

func ByteToUInt32(bts []byte) uint32 {
	buff := bytes.NewReader(bts)
	var nowVar uint32
	binary.Read(buff, binary.BigEndian, &nowVar)
	return nowVar
}

func UInt32ToByte(i uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, i)
	return b
}
