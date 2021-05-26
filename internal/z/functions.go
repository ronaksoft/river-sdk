package z

import (
	"bytes"
	"encoding/binary"
	"github.com/ronaksoft/rony/pools"
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

func AppendStrInt64(sb *strings.Builder, x int64) {
	b := pools.TinyBytes.GetLen(8)
	binary.BigEndian.PutUint64(b, uint64(x))
	sb.Write(b)
	pools.TinyBytes.Put(b)
}

func AppendStrUInt64(sb *strings.Builder, x uint64) {
	b := pools.TinyBytes.GetLen(8)
	binary.BigEndian.PutUint64(b, x)
	sb.Write(b)
	pools.TinyBytes.Put(b)
}

func AppendStrInt32(sb *strings.Builder, x int32) {
	b := pools.TinyBytes.GetLen(4)
	binary.BigEndian.PutUint32(b, uint32(x))
	sb.Write(b)
	pools.TinyBytes.Put(b)
}

func AppendStrUInt32(sb *strings.Builder, x uint32) {
	b := pools.TinyBytes.GetLen(4)
	binary.BigEndian.PutUint32(b, x)
	sb.Write(b)
	pools.TinyBytes.Put(b)

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
