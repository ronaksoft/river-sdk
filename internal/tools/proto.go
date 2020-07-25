package tools

import "github.com/gogo/protobuf/proto"

/*
   Creation Time: 2020 - Jul - 03
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

type ProtoBuffer interface {
	proto.Message
	proto.Sizer
	proto.Marshaler
	proto.Unmarshaler
	MarshalToSizedBuffer([]byte) (int, error)
}
