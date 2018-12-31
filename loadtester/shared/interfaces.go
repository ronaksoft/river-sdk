package shared

import (
	"git.ronaksoftware.com/ronak/riversdk/msg"
)

// Acter actor interface
type Acter interface {
	GetPhone() (phone string)
	SetPhone(phone string)

	GetPhoneList() (phoneList []string)
	SetPhoneList(phoneList []string)

	SetAuthInfo(authID int64, authKey []byte)
	GetAuthInfo() (authID int64, authKey []byte)
	GetAuthID() (authID int64)

	GetUserID() (userID int64)
	GetUserInfo() (userID int64, username, userFullName string)
	SetUserInfo(userID int64, username, userFullName string)

	GetPeers() (peers []*PeerInfo)
	SetPeers(peers []*PeerInfo)

	ExecuteRequest(message *msg.MessageEnvelope, onSuccess SuccessCallback, onTimeOut TimeoutCallback)

	Save() error

	Stop()
}

// Screenwriter scenario interface
type Screenwriter interface {
	Play(act Acter)
	Wait()
}

// Neter network interface
type Neter interface {
	Send(msgEnvelope *msg.MessageEnvelope) error
	SetAuthInfo(authID int64, authKey []byte)
	Start() error
	Stop()
}
