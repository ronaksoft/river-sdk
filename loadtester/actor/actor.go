package actor

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/controller"
	"git.ronaksoftware.com/ronak/riversdk/loadtester/executer"
	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"
	"git.ronaksoftware.com/ronak/riversdk/log"
	"git.ronaksoftware.com/ronak/riversdk/msg"
)

// Actor indicator as user
type Actor struct {
	Phone     string   `json:"phone"`
	PhoneList []string `json:"phone_list"`

	// Auth info will filled after CreateAuthKey
	AuthID  int64  `json:"auth_id"`
	AuthKey []byte `json:"auth_key"`

	// User will get filled after login/register
	UserID       int64  `json:"user_id"`
	UserName     string `json:"user_name"`
	UserFullName string `json:"user_fullname"`

	// Peers will filled after Contact import
	Peers []*shared.PeerInfo `json:"peers"`

	// Reporter data
	CreatedOn     time.Time          `json:"-"`
	Status        *shared.Status     `json:"-"`
	OnStopHandler func(phone string) `json:"-"`

	netCtrl shared.Neter
	exec    *executer.Executer
}

// NewActor create new actor instance
func NewActor(phone string) (shared.Acter, error) {
	var act *Actor
	acter, ok := shared.GetCachedActorByPhone(phone)
	if !ok {
		act = new(Actor)
		err := act.Load(phone)
		if err != nil {
			act = &Actor{
				Phone:     phone,
				PhoneList: make([]string, 0),
				UserID:    0,
				AuthID:    0,
				AuthKey:   make([]byte, 0),
				Peers:     make([]*shared.PeerInfo, 0),
			}
		}
	} else {
		act = acter.(*Actor)
	}

	act.netCtrl = controller.NewCtrlNetwork(act, act.onMessage, act.onUpdate, act.onError)
	act.exec = executer.NewExecuter(act.netCtrl)
	err := act.netCtrl.Start()
	if err != nil {
		return act, err
	}
	act.CreatedOn = time.Now()
	act.Status = new(shared.Status)
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
}

// GetAuthInfo Acter interface func
func (act *Actor) GetAuthInfo() (int64, []byte) { return act.AuthID, act.AuthKey }

// GetAuthID Acter interface func
func (act *Actor) GetAuthID() (authID int64) {
	return act.AuthID
}

// GetAuthKey Acter interface func
func (act *Actor) GetAuthKey() (authKey []byte) {
	return act.AuthKey
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
	atomic.AddInt64(&act.Status.RequestCount, 1)
	act.exec.Exec(message, onSuccess, onTimeOut, shared.DefaultSendTimeout)
}

// Save save actor after register/ login / contact import
func (act *Actor) Save() error {
	buff, err := json.Marshal(act)
	if err != nil {
		return err
	}
	return ioutil.WriteFile("_cache/"+act.Phone, buff, os.ModePerm)
}

// Load saved actor
func (act *Actor) Load(phone string) error {
	jsonBytes, err := ioutil.ReadFile("_cache/" + phone)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonBytes, act)
}

// Stop dispose actor
func (act *Actor) Stop() {
	if act.netCtrl != nil {
		act.Status.NetworkDisconnects = act.netCtrl.DisconnectCount()
		act.netCtrl.Stop()
		// act.netCtrl = nil
	}
	act.Status.LifeTime = time.Since(act.CreatedOn)
	if act.OnStopHandler != nil {
		act.OnStopHandler(act.Phone)
	}
}

// SetTimeout fill reporter data
func (act *Actor) SetTimeout(constructor int64, elapsed time.Duration) {
	act.Status.AverageTimeoutInterval += elapsed
	atomic.AddInt64(&act.Status.TimedoutRequests, 1)
}

// SetSuccess fill reporter data
func (act *Actor) SetSuccess(constructor int64, elapsed time.Duration) {
	act.Status.AverageSuccessInterval += elapsed
	atomic.AddInt64(&act.Status.SucceedRequests, 1)
}

// SetSucceed fill reporter data
func (act *Actor) SetActorSucceed(isSucceed bool) {
	act.Status.ActorSucceed = isSucceed
}

// GetStatus return actor statistics
func (act *Actor) GetStatus() *shared.Status {
	return act.Status
}

//SetStopHandler set on stop callback/delegate
func (act *Actor) SetStopHandler(fn func(phone string)) {
	act.OnStopHandler = fn
}

// ReceivedErrorResponse increase status ErrorRespons
func (act *Actor) ReceivedErrorResponse() {
	atomic.AddInt64(&act.Status.ErrorRespons, 1)
}

// onMessage check requestCallbacks and call callbacks
func (act *Actor) onMessage(messages []*msg.MessageEnvelope) {
	for _, m := range messages {
		// log.LOG_Debug("onMessage() Received ", zap.String("Constructor", msg.ConstructorNames[m.Constructor]), zap.Uint64("ReqID", m.RequestID))
		req := act.exec.GetRequest(m.RequestID)
		if req != nil {
			select {
			case req.ResponseWaitChannel <- m:
				log.LOG_Debug("onMessage() callback singnal sent")
			default:
				log.LOG_Warn("onMessage() callback is skipped probably timedout before",
					zap.Uint64("RequestID", m.RequestID),
					zap.String("Constructor", msg.ConstructorNames[m.Constructor]),
					zap.Duration("Elapsed", time.Since(req.CreatedOn)),
				)
			}
			act.exec.RemoveRequest(m.RequestID)
		} else if m.Constructor != msg.C_AuthRecalled {
			if m.Constructor == msg.C_Error {
				x := new(msg.Error)
				err := x.Unmarshal(m.Message)
				if err == nil {
					log.LOG_Error("onMessage() callback does not exists received Error",
						zap.Uint64("RequestID", m.RequestID),
						zap.String("Code", x.Code),
						zap.String("Item", x.Items),
					)
				} else {
					log.LOG_Error("onMessage() callback does not exists received Error and filed to unmarshal Error",
						zap.Uint64("RequestID", m.RequestID),
					)
				}
			} else {
				log.LOG_Debug("onMessage() callback does not exists", zap.Uint64("RequestID", m.RequestID), zap.String("Constructor", msg.ConstructorNames[m.Constructor]))
			}
		}
	}
}

func (act *Actor) onUpdate(updates []*msg.UpdateContainer) {

	for _, cnt := range updates {
		log.LOG_Debug("onUpdate()",
			zap.Int32("Length", cnt.Length),
			zap.Int64("MinID", cnt.MinUpdateID),
			zap.Int64("MaxID", cnt.MaxUpdateID),
		)
		// for _, u := range cnt.Updates {
		// 	// TODO : Implement actors update reactions
		// 	if u.Constructor == msg.C_UpdateNewMessage {
		// 	}
		// }
	}
}

func (act *Actor) onError(err *msg.Error) {
	// TODO : Add reporter error log
	log.LOG_Error("onError()", zap.String("Error", err.String()))
}
