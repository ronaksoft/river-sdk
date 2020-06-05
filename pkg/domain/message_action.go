package domain

import (
	"git.ronaksoftware.com/river/msg/msg"
)

const (
	// MessageActionNope glass messages type
	MessageActionNope int32 = 0x00
	// MessageActionContactRegistered glass messages type
	MessageActionContactRegistered int32 = 0x01
	// MessageActionGroupCreated glass messages type
	MessageActionGroupCreated int32 = 0x02
	// MessageActionGroupAddUser glass messages type
	MessageActionGroupAddUser int32 = 0x03
	// MessageActionGroupDeleteUser glass messages type
	MessageActionGroupDeleteUser int32 = 0x05
	// MessageActionGroupTitleChanged glass messages type
	MessageActionGroupTitleChanged int32 = 0x06
	// MessageActionClearHistory glass messages type
	MessageActionClearHistory int32 = 0x07
)

// ExtractActionUserIDs get user ids from  MessageActions
func ExtractActionUserIDs(act int32, data []byte) []int64 {
	res := make([]int64, 0)
	switch act {
	case MessageActionNope:
	case MessageActionContactRegistered:
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
		// TODO:: implement it
	case MessageActionClearHistory:
		// TODO:: implement it
	}

	return res
}
