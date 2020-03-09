package supernumerary

import (
	ronak "git.ronaksoftware.com/ronak/toolbox"
	log "git.ronaksoftware.com/ronak/toolbox/logger"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"

	msg "git.ronaksoftware.com/river/msg/chat"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-supernumerary/pkg/scenario"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-supernumerary/pkg/shared"
)

var (
	_Log   log.Logger
	_Redis *ronak.RedisCache
)

func init() {
	_Log = log.NewConsoleLogger()
}

func SetLogger(l log.Logger) {
	_Log = l
}

func SetRedis(r *ronak.RedisCache) {
	_Redis = r
}

// Supernumerary bunch of active actors
type Supernumerary struct {
	Actors      map[int64]shared.Actor
	FromPhoneNo int64
	ToPhoneNo   int64

	chTickerStop chan bool
	ticker       *time.Ticker
}

// NewSupernumerary creates new instance
func NewSupernumerary(fromPhoneNo, toPhoneNo int64) (*Supernumerary, error) {
	// create cache directory
	if _, err := os.Stat("_cache/"); os.IsNotExist(err) {
		err = os.Mkdir("_cache/", os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	s := &Supernumerary{
		Actors:       make(map[int64]shared.Actor),
		FromPhoneNo:  fromPhoneNo,
		ToPhoneNo:    toPhoneNo,
		chTickerStop: make(chan bool),
	}

	ml := sync.Mutex{}
	for i := fromPhoneNo; i < toPhoneNo; i++ {
		go func(i int64) {
			// time.Sleep(time.Duration(ronak.RandomInt(int(shared.DefaultMaxInterval/time.Second))) * time.Second)
			phone := shared.GetPhone(i)
			act, err := NewActor(phone)
			if err != nil {
				_Log.Info("Initialized Actor Failed",
					zap.String("Phone", phone),
					zap.Error(err),
				)
				return
			}
			_Log.Info("Initialized Actor", zap.String("Phone", phone))
			ml.Lock()
			s.Actors[i] = act
			ml.Unlock()

			// metric
			shared.Metrics.Gauge(shared.GaugeActors).Add(1)
		}(i)
		time.Sleep(time.Millisecond)
	}
	return s, nil
}

// Stop calls dispose()
func (s *Supernumerary) Stop() {
	s.dispose()
}

// dispose stops actors and cleans up
func (s *Supernumerary) dispose() {
	if s.ticker != nil {
		_Log.Debug("dispose() send ticker stop signal")
		s.chTickerStop <- true
		s.ticker.Stop()
		_Log.Debug("dispose() ticker stopped")
		s.ticker = nil
	}

	for i := s.FromPhoneNo; i < s.ToPhoneNo; i++ {
		if act, ok := s.Actors[i]; ok {
			_Log.Info("dispose() stopping actor", zap.String("Phone", act.GetPhone()))
			act.Stop()
			s.Actors[i] = nil
		}
	}
	shared.Metrics.Gauge(shared.GaugeActors).Set(0)
	s.Actors = nil
	s.FromPhoneNo = 0
	s.ToPhoneNo = 0
	s = nil // -_^

	_Log.Debug("dispose() done")
}

// CreateAuthKey init step required
func (s *Supernumerary) CreateAuthKey() {
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(len(s.Actors))
	// _Log.Info("Creating DH Keys ...")
	scenario.LoadServerKeys()
	sleepTime := shared.DefaultMaxInterval / time.Duration(len(s.Actors))
	for _, act := range s.Actors {
		go func(act shared.Actor) {
			defer waitGroup.Done()
			if act == nil {
				_Log.Warn("Actor is Nil")
				return
			}
			sen := scenario.NewCreateAuthKey(false)
			_Log.Info("CreateAuthKey() CreatingAuthKey", zap.String("Phone", act.GetPhone()))
			success := scenario.Play(act, sen)
			if success {
				err := act.Save()
				_Log.Debug("CreateAuthKey() save actor", zap.Error(err))
			}
		}(act)
		time.Sleep(sleepTime)
	}
	waitGroup.Wait()
}

// CreateAuthKey init step required
func (s *Supernumerary) CreateAuthKeyTest() {
	waitGroup := sync.WaitGroup{}
	sleepTime := shared.DefaultMaxInterval / time.Duration(len(s.Actors))
	for _, act := range s.Actors {
		waitGroup.Add(1)
		go func(act shared.Actor) {
			defer waitGroup.Done()
			if act == nil {
				_Log.Warn("Actor is Nil")
				return
			}
			sen := scenario.NewCreateAuthKeyTest(false)
			_Log.Info("CreateAuthKeyTest() CreatingAuthKey", zap.String("Phone", act.GetPhone()))
			success := scenario.Play(act, sen)
			if success {
				err := act.Save()
				_Log.Debug("CreateAuthKeyTest() save actor", zap.Error(err))
			}
		}(act)
		time.Sleep(sleepTime)
	}
	waitGroup.Wait()
}

// Register init step required
func (s *Supernumerary) Register() {
	waitGroup := sync.WaitGroup{}
	sleepTime := shared.DefaultMaxInterval / time.Duration(len(s.Actors))
	for _, act := range s.Actors {
		waitGroup.Add(1)
		go func(act shared.Actor) {
			defer waitGroup.Done()
			sen := scenario.NewRegister(false)
			_Log.Info("Register() Registering", zap.String("Phone", act.GetPhone()))
			success := scenario.Play(act, sen)
			if success {
				err := act.Save()
				_Log.Debug("Register() save actor", zap.Error(err))
			}
		}(act)
		time.Sleep(sleepTime)
	}
	waitGroup.Wait()
}

// Login init step required
func (s *Supernumerary) Login() {
	waitGroup := sync.WaitGroup{}
	sleepTime := shared.DefaultMaxInterval / time.Duration(len(s.Actors))
	for _, act := range s.Actors {
		waitGroup.Add(1)
		go func(act shared.Actor) {
			defer waitGroup.Done()
			sen := scenario.NewLogin(false)
			_Log.Info("Login() Logging in", zap.String("Phone", act.GetPhone()))
			success := scenario.Play(act, sen)
			if success {
				err := act.Save()
				_Log.Debug("Login() save actor", zap.Error(err))
			}
		}(act)
		time.Sleep(sleepTime)
	}
	waitGroup.Wait()
}

// Reset Authorizations
func (s *Supernumerary) ResetAuthorizations() {
	for _, act := range s.Actors {
		sen := scenario.NewResetAuthorizations(false)
		_Log.Info("ResetAuthorization()", zap.String("Phone", act.GetPhone()))
		_ = scenario.Play(act, sen)
	}
}

// Create Group
func (s *Supernumerary) CreateGroup(startPhone, endPhone, groupSize int64) {
	for _, act := range s.Actors {
		sen := scenario.NewCreateGroup(false)
		_Log.Info("CreateGroup()", zap.String("Phone", act.GetPhone()))

		success := scenario.Play(act, sen)
		if success {
			err := act.Save()
			_Log.Debug("Login() save actor", zap.Error(err))
		}
	}
}

// SetTickerApplier try to invoke certain action for all actors repeatedly
func (s *Supernumerary) SetTickerApplier(duration time.Duration, action TickerAction) {
	if s.ticker != nil {
		s.chTickerStop <- true
		s.ticker.Stop()
	}
	s.ticker = time.NewTicker(duration)
	go s.tickerApplier(action, duration)
	_Log.Debug("SetTickerApplier() Done")
}

// tickerApplier
func (s *Supernumerary) tickerApplier(action TickerAction, duration time.Duration) {
	for {
		select {
		case t := <-s.ticker.C:
			_Log.Info("tickerApplier() ticker signal", zap.Time("Time", t))
			switch action {
			case TickerActionNone:
				// NOP
			case TickerActionSendMessage:
				// try to send random message
				waitGroup := sync.WaitGroup{}
				waitGroup.Add(len(s.Actors))
				for _, act := range s.Actors {
					go func(act shared.Actor) {
						defer waitGroup.Done()
						time.Sleep(time.Duration(ronak.RandomInt64(duration.Nanoseconds())) * time.Nanosecond)

						sen := scenario.NewSendMessage(false)
						// import random contact for actor
						act.SetPeers([]*shared.PeerInfo{s.getRandomPeerUser(act)})

						_Log.Debug("Actor",
							zap.String("Phone", act.GetPhone()),
							zap.Int64("AuthID", act.GetAuthID()),
							zap.Int64("UserID", act.GetUserID()),
						)

						// async
						sen.Play(act)
					}(act)
				}
				waitGroup.Wait()
			case TickerActionSendFile:
				// try to send random file
				for _, act := range s.Actors {
					sen := scenario.NewSendFile(false)
					// import random contact for actor
					act.SetPeers([]*shared.PeerInfo{s.getRandomPeerUser(act)})
					// async
					sen.Play(act)
				}
			case TickerActionSendGroupMessage:
				// try to send random message
				waitGroup := sync.WaitGroup{}
				waitGroup.Add(len(s.Actors))
				for _, act := range s.Actors {
					go func(act shared.Actor) {
						defer waitGroup.Done()
						time.Sleep(time.Duration(ronak.RandomInt64(duration.Nanoseconds())) * time.Nanosecond)

						sen := scenario.NewSendMessage(false)
						// import random contact for actor
						act.SetPeers([]*shared.PeerInfo{s.getRandomPeerUser(act)})

						_Log.Debug("Actor",
							zap.String("Phone", act.GetPhone()),
							zap.Int64("AuthID", act.GetAuthID()),
							zap.Int64("UserID", act.GetUserID()),
						)

						// async
						sen.Play(act)
					}(act)
				}
				waitGroup.Wait()
			}

		case <-s.chTickerStop:
			_Log.Warn("tickerApplier() stop signal")
			return
		}
	}
}

// getRandomPeerUser returns random peer
func (s *Supernumerary) getRandomPeerUser(act shared.Actor) *shared.PeerInfo {
	fromUserID := act.GetUserID()
	for {
		phone := s.getRandomPhoneNo()
		if a, ok := s.Actors[phone]; ok {
			if fromUserID != a.GetUserID() {
				return scenario.GetPeerInfo(fromUserID, a.GetUserID(), msg.PeerUser)
			}
		}
	}
}

// getRandomPhoneNo select a random phoneNo between fromPhoneNo - toPhoneNo
func (s *Supernumerary) getRandomPhoneNo() int64 {
	return s.FromPhoneNo + ronak.RandomInt64(s.ToPhoneNo-s.FromPhoneNo) + 1 // +1 bcz Int63n(n) returns [0,n)
}
