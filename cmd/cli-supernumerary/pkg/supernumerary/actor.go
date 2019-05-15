package supernumerary

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-supernumerary/pkg/controller"
	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-supernumerary/pkg/executer"
	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-supernumerary/pkg/shared"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
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
	exec    *executer.Executor

	mxUpdate      sync.Mutex
	updateApplier map[int64]shared.UpdateApplier
}

// NewActor create new actor instance
func NewActor(phone string) (shared.Actor, error) {
	var act *Actor
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

	act.updateApplier = make(map[int64]shared.UpdateApplier)
	act.netCtrl = controller.NewCtrlNetwork(act, act.onMessage, act.onUpdate, act.onError)
	act.exec = executer.NewExecutor(act.netCtrl)
	err = act.netCtrl.Start()
	if err != nil {
		return act, err
	}
	act.CreatedOn = time.Now()
	act.Status = new(shared.Status)
	return act, nil
}

// GetPhone Actor interface func
func (act *Actor) GetPhone() string { return act.Phone }

// SetPhone Actor interface func
func (act *Actor) SetPhone(phone string) { act.Phone = phone }

// GetPhoneList Actor interface func
func (act *Actor) GetPhoneList() []string { return act.PhoneList }

// SetPhoneList Actor interface func
func (act *Actor) SetPhoneList(phoneList []string) { act.PhoneList = phoneList }

// SetAuthInfo set authID and authKey after CreateAuthKey completed
func (act *Actor) SetAuthInfo(authID int64, authKey []byte) {
	act.AuthID = authID
	act.AuthKey = make([]byte, len(authKey))
	copy(act.AuthKey, authKey)
}

// GetAuthInfo Actor interface func
func (act *Actor) GetAuthInfo() (int64, []byte) { return act.AuthID, act.AuthKey }

// GetAuthID Actor interface func
func (act *Actor) GetAuthID() (authID int64) {
	return act.AuthID
}

// GetAuthKey Actor interface func
func (act *Actor) GetAuthKey() (authKey []byte) {
	return act.AuthKey
}

// GetUserID Actor interface func
func (act *Actor) GetUserID() int64 { return act.UserID }

// GetUserInfo Actor interface func
func (act *Actor) GetUserInfo() (userID int64, username, userFullName string) {
	return act.UserID, act.UserName, act.UserFullName
}

// SetUserInfo Actor interface func
func (act *Actor) SetUserInfo(userID int64, username, userFullName string) {
	act.UserID = userID
	act.UserName = username
	act.UserFullName = userFullName
}

// GetPeers Actor interface func
func (act *Actor) GetPeers() []*shared.PeerInfo { return act.Peers }

// SetPeers Actor interface func
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
	_, err = _Redis.Set(act.Phone, buff)
	return err
}

// Load saved actor
func (act *Actor) Load(phone string) error {
	jsonBytes, err := _Redis.GetBytes(phone)
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
	shared.Metrics.CounterVec(shared.CntTimeout).WithLabelValues(msg.ConstructorNames[constructor]).Add(1)
	shared.Metrics.Histogram(shared.HistTimeoutLatency).Observe(float64(elapsed / time.Millisecond))

	act.Status.AverageTimeoutInterval += elapsed
	atomic.AddInt64(&act.Status.TimeoutRequests, 1)
}

// SetSuccess fill reporter data
func (act *Actor) SetSuccess(constructor int64, elapsed time.Duration) {
	// metric
	shared.Metrics.CounterVec(shared.CntSuccess).WithLabelValues(msg.ConstructorNames[constructor]).Add(1)
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
		shared.Metrics.Counter(shared.CntFailedScenario).Add(1)
	}
}

// GetStatus return actor statistics
func (act *Actor) GetStatus() *shared.Status {
	return act.Status
}

// SetStopHandler set on stop callback/delegate
func (act *Actor) SetStopHandler(fn func(phone string)) {
	act.OnStopHandler = fn
}

// ReceivedErrorResponse increase status ErrorResponses
func (act *Actor) ReceivedErrorResponse() {
	// metrics
	shared.Metrics.Counter(shared.CntError).Add(1)

	atomic.AddInt64(&act.Status.ErrorResponses, 1)
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
				_Log.Debug("onMessage() callback signal sent",
					zap.Uint64("RequestID", m.RequestID),
					zap.String("Constructor", msg.ConstructorNames[m.Constructor]),
					zap.Duration("Elapsed", time.Since(req.CreatedOn)),
				)
			default:
				elapsed := time.Since(req.CreatedOn)
				// metric
				shared.Metrics.CounterVec(shared.CntDiffered).WithLabelValues(msg.ConstructorNames[m.Constructor]).Add(1)
				shared.Metrics.Histogram(shared.HistDifferedLatency).Observe(float64(elapsed / time.Millisecond))

				_Log.Error("onMessage() callback is skipped probably timeout before",
					zap.Uint64("RequestID", m.RequestID),
					zap.String("Constructor", msg.ConstructorNames[m.Constructor]),
					zap.Duration("Elapsed", elapsed),
				)
			}
			act.exec.RemoveRequest(m.RequestID)
			return
		}
		if m.Constructor != msg.C_AuthRecalled {
			if m.Constructor == msg.C_Error {
				x := new(msg.Error)
				err := x.Unmarshal(m.Message)
				if err == nil {
					_Log.Error("onMessage() callback does not exists received Error",
						zap.Uint64("RequestID", m.RequestID),
						zap.String("Code", x.Code),
						zap.String("Item", x.Items),
					)
				} else {
					_Log.Error("onMessage() callback does not exists received Error and filed to unmarshal Error",
						zap.Uint64("RequestID", m.RequestID),
					)
				}
			} else {
				_Log.Debug("onMessage() callback does not exists", zap.Uint64("RequestID", m.RequestID), zap.String("Constructor", msg.ConstructorNames[m.Constructor]))
			}
		}
	}
}

func (act *Actor) onUpdate(updates []*msg.UpdateContainer) {
	_Log.Debug("onUpdate() Debounced UpdateContainer",
		zap.Int("Length", len(updates)),
	)
	for _, cnt := range updates {
		_Log.Debug("onUpdate() Processing UpdateContainer",
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
			_Log.Debug("onUpdate() Received ", zap.String("Constructor", msg.ConstructorNames[u.Constructor]))
		}
	}
}

func (act *Actor) onError(err *msg.Error) {
	// metric
	shared.Metrics.Counter(shared.CntError).Add(1)

	// TODO : Add reporter error log
	_Log.Error("onError()", zap.String("Error", err.String()))
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
