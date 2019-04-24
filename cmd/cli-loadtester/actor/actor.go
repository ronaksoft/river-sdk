package actor

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-loadtester/controller"
	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-loadtester/executer"
	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-loadtester/logs"
	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-loadtester/shared"
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

	mxUpdate      sync.Mutex
	updateApplier map[int64]shared.UpdateApplier
}

// NewActor create new actor instance
func NewActor(phone string) (shared.Acter, error) {
	var act *Actor
	acter, ok := shared.GetCachedActorByPhone(phone)
	if !ok {
		logs.Warn("NewActor() Actor not found in Cache", zap.String("Phone", phone))
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
	act.updateApplier = make(map[int64]shared.UpdateApplier)
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

	// save to cached actors
	_, ok := shared.GetCachedActorByPhone(act.Phone)
	if !ok {
		shared.CacheActor(act)
	}

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
	// metric
	shared.Metrics.CounterVec(shared.CntTimedout).WithLabelValues(msg.ConstructorNames[constructor]).Add(1)
	shared.Metrics.Histogram(shared.HistTimeoutLatency).Observe(float64(elapsed / time.Millisecond))

	act.Status.AverageTimeoutInterval += elapsed
	atomic.AddInt64(&act.Status.TimedoutRequests, 1)
}

// SetSuccess fill reporter data
func (act *Actor) SetSuccess(constructor int64, elapsed time.Duration) {
	// metric
	shared.Metrics.CounterVec(shared.CntSucceess).WithLabelValues(msg.ConstructorNames[constructor]).Add(1)
	shared.Metrics.Histogram(shared.HistSuccessLatency).Observe(float64(elapsed / time.Millisecond))

	act.Status.AverageSuccessInterval += elapsed
	atomic.AddInt64(&act.Status.SucceedRequests, 1)
}

// SetSucceed fill reporter data
func (act *Actor) SetActorSucceed(isSucceed bool) {
	act.Status.ActorSucceed = isSucceed
	if isSucceed {
		shared.Metrics.Counter(shared.CntSucceedScenario).Add(1)
	} else {
		shared.Metrics.Counter(shared.CntFaildScenario).Add(1)
	}
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
	// metrics
	shared.Metrics.Counter(shared.CntError).Add(1)

	atomic.AddInt64(&act.Status.ErrorRespons, 1)
}

// onMessage check requestCallbacks and call callbacks
func (act *Actor) onMessage(messages []*msg.MessageEnvelope) {
	for _, m := range messages {
		// metric
		shared.Metrics.CounterVec(shared.CntResponse).WithLabelValues(msg.ConstructorNames[m.Constructor]).Add(1)

		req := act.exec.GetRequest(m.RequestID)
		if req != nil {
			select {
			case req.ResponseWaitChannel <- m:
				logs.Debug("onMessage() callback singnal sent",
					zap.Uint64("RequestID", m.RequestID),
					zap.String("Constructor", msg.ConstructorNames[m.Constructor]),
					zap.Duration("Elapsed", time.Since(req.CreatedOn)),
				)
			default:
				elapsed := time.Since(req.CreatedOn)
				if elapsed < shared.DefaultTimeout {
					// if elapsed time is less than timeout retry to pass request until 1 sec
					go func(request *executer.Request, message *msg.MessageEnvelope) {
						select {
						case request.ResponseWaitChannel <- message:
							return
						case <-time.After(time.Second):
						}
						// this is not differed response
						// // metric
						// shared.Metrics.CounterVec(shared.CntDiffered).WithLabelValues(msg.ConstructorNames[message.Constructor]).Add(1)
					}(req, m)

				} else {

					// metric
					shared.Metrics.CounterVec(shared.CntDiffered).WithLabelValues(msg.ConstructorNames[m.Constructor]).Add(1)
					shared.Metrics.Histogram(shared.HistDifferedLatency).Observe(float64(elapsed / time.Millisecond))

					logs.Error("onMessage() callback is skipped probably timedout before",
						zap.Uint64("RequestID", m.RequestID),
						zap.String("Constructor", msg.ConstructorNames[m.Constructor]),
						zap.Duration("Elapsed", elapsed),
					)
				}
			}
			act.exec.RemoveRequest(m.RequestID)
		} else if m.Constructor != msg.C_AuthRecalled {
			if m.Constructor == msg.C_Error {
				x := new(msg.Error)
				err := x.Unmarshal(m.Message)
				if err == nil {
					logs.Error("onMessage() callback does not exists received Error",
						zap.Uint64("RequestID", m.RequestID),
						zap.String("Code", x.Code),
						zap.String("Item", x.Items),
					)
				} else {
					logs.Error("onMessage() callback does not exists received Error and filed to unmarshal Error",
						zap.Uint64("RequestID", m.RequestID),
					)
				}
			} else {
				logs.Debug("onMessage() callback does not exists", zap.Uint64("RequestID", m.RequestID), zap.String("Constructor", msg.ConstructorNames[m.Constructor]))
			}
		}
	}
}

func (act *Actor) onUpdate(updates []*msg.UpdateContainer) {
	logs.Debug("onUpdate() Debounced UpdateContainer",
		zap.Int("Length", len(updates)),
	)
	for _, cnt := range updates {
		logs.Debug("onUpdate() Processing UpdateContainer",
			zap.Int32("Length", cnt.Length),
			zap.Int64("MinID", cnt.MinUpdateID),
			zap.Int64("MaxID", cnt.MaxUpdateID),
		)
		for _, u := range cnt.Updates {
			// metric
			shared.Metrics.CounterVec(shared.CntResponse).WithLabelValues(msg.ConstructorNames[u.Constructor]).Add(1)

			if fn, ok := act.updateApplier[u.Constructor]; ok {
				fn(act, u)
			}
			logs.Debug("onUpdate() Received ", zap.String("Constructor", msg.ConstructorNames[u.Constructor]))
		}
	}
}

func (act *Actor) onError(err *msg.Error) {
	// metric
	shared.Metrics.Counter(shared.CntError).Add(1)

	// TODO : Add reporter error log
	logs.Error("onError()", zap.String("Error", err.String()))
}

// SetUpdateApplier set update appliers
func (act *Actor) SetUpdateApplier(constructor int64, fn shared.UpdateApplier) {
	act.mxUpdate.Lock()
	act.updateApplier[constructor] = fn
	act.mxUpdate.Unlock()
}

// ExecFileRequest execute request against file server
func (act *Actor) ExecFileRequest(msgEnvelope *msg.MessageEnvelope) (*msg.MessageEnvelope, error) {

	// metric
	shared.Metrics.CounterVec(shared.CntRequest).WithLabelValues(msg.ConstructorNames[msgEnvelope.Constructor]).Add(1)
	shared.Metrics.Counter(shared.CntFile).Add(1)

	sw := time.Now()
	env, err := controller.ExecuteFileRequest(msgEnvelope, act)

	if err != nil {
		shared.Metrics.Counter(shared.CntFileError).Add(1)
		return env, err
	}
	elapsed := time.Since(sw)

	// metric
	shared.Metrics.CounterVec(shared.CntResponse).WithLabelValues(msg.ConstructorNames[env.Constructor]).Add(1)
	shared.Metrics.Histogram(shared.HistFileLatency).Observe(float64(elapsed / time.Millisecond))

	return env, err
}
