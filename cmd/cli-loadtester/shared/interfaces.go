package shared

import (
	"time"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
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
	GetAuthKey() (authKey []byte)

	GetUserID() (userID int64)
	GetUserInfo() (userID int64, username, userFullName string)
	SetUserInfo(userID int64, username, userFullName string)

	GetPeers() (peers []*PeerInfo)
	SetPeers(peers []*PeerInfo)

	ExecuteRequest(message *msg.MessageEnvelope, onSuccess SuccessCallback, onTimeOut TimeoutCallback)

	Save() error

	Stop()

	SetTimeout(constructor int64, elapsed time.Duration)
	SetSuccess(constructor int64, elapsed time.Duration)
	SetActorSucceed(isSucceed bool)
	GetStatus() *Status
	SetStopHandler(func(phone string))
	ReceivedErrorResponse()

	SetUpdateApplier(constructor int64, fn UpdateApplier)
	ExecFileRequest(msgEnvelop *msg.MessageEnvelope) (*msg.MessageEnvelope, error)
}

// Screenwriter scenario interface
type Screenwriter interface {
	Play(act Acter)
	Wait(act Acter) bool
	SetFinal(isFinal bool)
	IsFinal() bool
	AddJobs(i int)
	GetResult() bool
}

// Neter network interface
type Neter interface {
	Send(msgEnvelope *msg.MessageEnvelope) error
	Start() error
	Stop()
	IsConnected() bool
	DisconnectCount() int64
}

// Reporter report interface
type Reporter interface {
	Register(act Acter)
	String() string
	Print()
	Clear()
	IsActive() bool
	SetIsActive(isActive bool)
}
