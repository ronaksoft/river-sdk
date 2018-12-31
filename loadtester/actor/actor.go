package actor

import (
	"git.ronaksoftware.com/ronak/riversdk/loadtester/controller"
	"git.ronaksoftware.com/ronak/riversdk/loadtester/executer"
	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"
	"git.ronaksoftware.com/ronak/riversdk/log"
	"git.ronaksoftware.com/ronak/riversdk/msg"
	"go.uber.org/zap"
)

// Actor indicator as user
type Actor struct {
	Phone     string
	PhoneList []string

	// Auth info will filled after CreateAuthKey
	AuthID  int64
	AuthKey []byte //[256]byte

	// User will get filled after login/register
	UserID       int64
	UserName     string
	UserFullName string

	// Peers will filled after Contact import
	Peers []*shared.PeerInfo

	netCtrl shared.Neter
	exec    *executer.Executer
}

// NewActor create new actor instance
func NewActor(phone string) (shared.Acter, error) {

	act := &Actor{
		Phone:     phone,
		PhoneList: make([]string, 0),
		UserID:    0,
		AuthID:    0,
		AuthKey:   make([]byte, 0),
		Peers:     make([]*shared.PeerInfo, 0),
	}
	act.netCtrl = controller.NewCtrlNetwork(act.AuthID, act.AuthKey, act.onMessage, act.onUpdate, act.onError)
	err := act.netCtrl.Start()
	if err != nil {
		return act, err
	}
	act.exec = executer.NewExecuter(act.netCtrl)

	return act, nil
}

// GetPhone Acter interface func
func (act *Actor) GetPhone() string { return act.Phone }

// SetPhone Acter interface func
func (act *Actor) SetPhone(phone string) { act.Phone = phone }

// GetPhoneList Acter interface func
func (act *Actor) GetPhoneList() []string { return act.PhoneList }

// SetPhoneList Acter interface func
func (act *Actor) SetPhoneList(phoneList []string) { act.PhoneList = phoneList }

// SetAuthInfo set authID and authKey after CreateAuthKey completed
func (act *Actor) SetAuthInfo(authID int64, authKey []byte) {
	act.AuthID = authID
	act.AuthKey = make([]byte, len(authKey))
	copy(act.AuthKey, authKey)
	act.netCtrl.SetAuthInfo(act.AuthID, act.AuthKey)
}

// GetAuthInfo Acter interface func
func (act *Actor) GetAuthInfo() (int64, []byte) { return act.AuthID, act.AuthKey }

// GetAuthID Acter interface func
func (act *Actor) GetAuthID() (authID int64) {
	return act.AuthID
}

// GetUserID Acter interface func
func (act *Actor) GetUserID() int64 { return act.UserID }

// GetUserInfo Acter interface func
func (act *Actor) GetUserInfo() (userID int64, username, userFullName string) {
	return act.UserID, act.UserName, act.UserFullName
}

// SetUserInfo Acter interface func
func (act *Actor) SetUserInfo(userID int64, username, userFullName string) {
	act.UserID = userID
	act.UserName = username
	act.UserFullName = userFullName
}

// GetPeers Acter interface func
func (act *Actor) GetPeers() []*shared.PeerInfo { return act.Peers }

// SetPeers Acter interface func
func (act *Actor) SetPeers(peers []*shared.PeerInfo) { act.Peers = peers }

// ExecuteRequest send request to server
func (act *Actor) ExecuteRequest(message *msg.MessageEnvelope, onSuccess shared.SuccessCallback, onTimeOut shared.TimeoutCallback) {
	if message == nil {
		return
	}
	act.exec.Exec(message, onSuccess, onTimeOut, shared.DefaultSendTimeout)
}

// onMessage check requestCallbacks and call callbacks
func (act *Actor) onMessage(messages []*msg.MessageEnvelope) {
	for _, m := range messages {
		req := act.exec.GetRequest(m.RequestID)
		if req != nil {
			select {
			case req.ResponseWaitChannel <- m:
				log.LOG_Debug("Actor::onMessage() callback singnal sent")
			default:
				log.LOG_Debug("Actor::onMessage() callback is skipped probably timedout before")
			}
			act.exec.RemoveRequest(m.RequestID)
		} else {
			log.LOG_Warn("Actor::onMessage() callback does not exists",
				zap.Uint64("RequestID", m.RequestID),
			)
		}
	}
}

func (act *Actor) onUpdate(updates []*msg.UpdateContainer) {
	for _, cnt := range updates {
		for _, u := range cnt.Updates {
			// TODO : Implement actors update reactions
			if u.Constructor == msg.C_UpdateNewMessage {
			}
		}
	}
}

func (act *Actor) onError(err *msg.Error) {
	// TODO : Add reporter error log
	log.LOG_Warn("Actor::onError() ", zap.String("Error", err.String()))
}
