package repo

import (
	"bytes"
	"encoding/binary"
	"github.com/gobwas/pool/pbytes"
	"github.com/gogo/protobuf/proto"
	"hash/crc32"
)

/*
   Creation Time: 2019 - Jul - 03
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software GroupSearch 2018
*/

func alreadySaved(id string, message proto.Message) bool {
	msgBytes, _ := proto.Marshal(message)
	checkSumBytes := pbytes.GetLen(4)
	binary.BigEndian.PutUint32(checkSumBytes, crc32.ChecksumIEEE(msgBytes))

	cachedBytes, err := lCache.Get(id)
	if err != nil {
		return false
	}
	if bytes.Equal(cachedBytes, checkSumBytes) {
		return true
	}
	return false
}
