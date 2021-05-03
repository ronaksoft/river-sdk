package domain

import (
	"git.ronaksoft.com/river/msg/go/msg"
)

const (
	MessageActionNope              int32 = 0x00
	MessageActionContactRegistered int32 = 0x01
	MessageActionGroupCreated      int32 = 0x02
	MessageActionGroupAddUser      int32 = 0x03
	MessageActionGroupDeleteUser   int32 = 0x05
	MessageActionGroupTitleChanged int32 = 0x06
	MessageActionClearHistory      int32 = 0x07
)

// ExtractActionUserIDs get user ids from  MessageActions
func ExtractActionUserIDs(act int32, data []byte) []int64 {
	res := make([]int64, 0)
	switch act {
	case MessageActionNope:
	case MessageActionContactRegistered:
	case MessageActionGroupCreated:
		x := &msg.MessageActionGroupCreated{}
		err := x.Unmarshal(data)
		if err == nil {
			res = append(res, x.UserIDs...)
		}
	case MessageActionGroupAddUser:
		x := &msg.MessageActionGroupAddUser{}
		err := x.Unmarshal(data)
		if err == nil {
			res = append(res, x.UserIDs...)
		}
	case MessageActionGroupDeleteUser:
		x := &msg.MessageActionGroupDeleteUser{}
		err := x.Unmarshal(data)
		if err == nil {
			res = append(res, x.UserIDs...)
		}
	case MessageActionGroupTitleChanged:
		// TODO:: implement it
	case MessageActionClearHistory:
		// TODO:: implement it
	}

	return res
}
