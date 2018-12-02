package synchronizer

const (
	MessageActionNope              int32 = 0x00
	MessageActionContactRegistered int32 = 0x01
	MessageActionGroupCreated      int32 = 0x02
	MessageActionGroupAddUser      int32 = 0x03
	MessageActionGroupDeleteUser   int32 = 0x05
	MessageActionGroupTitleChanged int32 = 0x06
	MessageActionClearHistory      int32 = 0x07
)
