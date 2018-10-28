package domain

type RiverConnector struct {
	AuthID    int64
	AuthKey   [256]byte
	UserID    int64
	Username  string
	Phone     string
	FirstName string
	LastName  string
}

type RiverConfiger interface {
	Save()
	Load() error
	ChangeAuthID(authID int64)
	ChangeAuthKey(authKey [256]byte)
	ChangeUserID(userID int64)
	ChangeUsername(username string)
	ChangePhone(phone string)
	ChangeFirstName(firstName string)
	ChangeLastName(lastName string)
	PickupAuthID() int64
	PickupAuthKey() [256]byte
	PickupUserID() int64
	PickupUsername() string
	PickupPhone() string
	PickupFirstName() string
	PickupLastName() string
}
