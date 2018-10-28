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
	SetAuthID(authID int64)
	SetAuthKey(authKey [256]byte)
	SetUserID(userID int64)
	SetUsername(username string)
	SetPhone(phone string)
	SetFirstName(firstName string)
	SetLastName(lastName string)
	GetAuthID() int64
	GetAuthKey() [256]byte
	GetUserID() int64
	GetUsername() string
	GetPhone() string
	GetFirstName() string
	GetLastName() string
}
