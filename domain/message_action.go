package domain

import (
	"git.ronaksoftware.com/ronak/riversdk/msg"
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

func ExtractActionUserIDs(act int32, data []byte) []int64 {

	res := make([]int64, 0)
	switch act {
	case MessageActionNope:
		// Do Nothing
	case MessageActionContactRegistered:
		//x :=new(msg.MessageActionContactRegistered)

	case MessageActionGroupCreated:
		x := new(msg.MessageActionGroupCreated)
		err := x.Unmarshal(data)
		if err == nil {
			res = append(res, x.UserIDs...)
		}
	case MessageActionGroupAddUser:
		x := new(msg.MessageActionGroupAddUser)
		err := x.Unmarshal(data)
		if err == nil {
			res = append(res, x.UserIDs...)
		}
	case MessageActionGroupDeleteUser:
		x := new(msg.MessageActionGroupDeleteUser)
		err := x.Unmarshal(data)
		if err == nil {
			res = append(res, x.UserIDs...)
		}
	case MessageActionGroupTitleChanged:
		// x := new(msg.MessageActionGroupTitleChanged)
	case MessageActionClearHistory:
		// x := new(msg.MessageActionClearHistory)
	}

	return res
}
