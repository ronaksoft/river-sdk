package supernumerary

import (
	"math/rand"
	"os"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-loadtester/logs"
	"go.uber.org/zap"

	"git.ronaksoftware.com/ronak/riversdk/msg"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-loadtester/scenario"

	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-loadtester/actor"
	"git.ronaksoftware.com/ronak/riversdk/cmd/cli-loadtester/shared"
)

// Supernumerary bunch of active actors
type Supernumerary struct {
	Actors      map[int64]shared.Acter
	FromPhoneNo int64
	ToPhoneNo   int64

	chTikerStop chan bool
	ticker      *time.Ticker
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
		Actors:      make(map[int64]shared.Acter),
		FromPhoneNo: fromPhoneNo,
		ToPhoneNo:   toPhoneNo,
		chTikerStop: make(chan bool),
	}

	for i := fromPhoneNo; i < toPhoneNo; i++ {
		phone := shared.GetPhone(i)
		act, err := actor.NewActor(phone)
		if err != nil {
			defer s.dispose()
			return nil, err
		}
		logs.Info("Initialized Actor", zap.String("Phone", phone))
		s.Actors[i] = act

		// metric
		shared.Metrics.Gauge(shared.GaugeActors).Add(1)
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
		logs.Debug("dispose() send ticker stop signal")
		s.chTikerStop <- true
		s.ticker.Stop()
		logs.Debug("dispose() ticker stopped")
		s.ticker = nil
	}

	for i := s.FromPhoneNo; i < s.ToPhoneNo; i++ {
		if act, ok := s.Actors[i]; ok {
			logs.Info("dispose() stopping actor", zap.String("Phone", act.GetPhone()))
			act.Stop()
			s.Actors[i] = nil
		}
	}
	shared.Metrics.Gauge(shared.GaugeActors).Set(0)
	s.Actors = nil
	s.FromPhoneNo = 0
	s.ToPhoneNo = 0
	s = nil // -_^

	logs.Debug("dispose() done")
}

// CreateAuthKey init step required
func (s *Supernumerary) CreateAuthKey() {
	for _, act := range s.Actors {
		sen := scenario.NewCreateAuthKey(false)
		logs.Info("CreateAuthKey() CreatingAuthKey", zap.String("Phone", act.GetPhone()))
		success := scenario.Play(act, sen)
		if success {
			err := act.Save()
			logs.Debug("CreateAuthKey() save actor", zap.Error(err))
		}
	}
}

// Register init step required
func (s *Supernumerary) Register() {
	for _, act := range s.Actors {
		sen := scenario.NewRegister(false)
		logs.Info("Register() Registering", zap.String("Phone", act.GetPhone()))
		success := scenario.Play(act, sen)
		if success {
			err := act.Save()
			logs.Debug("Register() save actor", zap.Error(err))
		}
	}
}

// Login init step required
func (s *Supernumerary) Login() {
	for _, act := range s.Actors {
		sen := scenario.NewLogin(false)
		logs.Info("Login() Loging in", zap.String("Phone", act.GetPhone()))
		success := scenario.Play(act, sen)
		if success {
			err := act.Save()
			logs.Debug("Login() save actor", zap.Error(err))
		}
	}
}

// SetTickerApplier try to invoke certain action for all actors repeatedly
func (s *Supernumerary) SetTickerApplier(duration time.Duration, action TickerAction) {
	if s.ticker != nil {
		logs.Debug("SetTickerApplier() send ticker stop signal")
		s.chTikerStop <- true
		s.ticker.Stop()
		logs.Debug("SetTickerApplier() ticker stopped")
	}
	s.ticker = time.NewTicker(duration)
	go s.tickerApplier(action)
	logs.Debug("SetTickerApplier() Done")
}

// tickerApplier
func (s *Supernumerary) tickerApplier(action TickerAction) {
	for {
		select {
		case t := <-s.ticker.C:
			logs.Info("tickerApplier() ticker signal", zap.Time("Time", t))
			switch action {
			case TickerActionNone:
				// NOP
			case TickerActionSendMessage:
				// try to send random message
				for _, act := range s.Actors {
					sen := scenario.NewSendMessage(false)
					// import random contact for actor
					act.SetPeers([]*shared.PeerInfo{s.fnGetRandomPeer(act)})
					// async
					sen.Play(act)

					// check stop signal while executing
					if s.fnCheckStopSignal() {
						return
					}
				}

			case TickerActionSendFile:
				// try to send random file
				for _, act := range s.Actors {
					sen := scenario.NewSendFile(false)
					// import random contact for actor
					act.SetPeers([]*shared.PeerInfo{s.fnGetRandomPeer(act)})
					// async
					sen.Play(act)

					// check stop signal while executing
					if s.fnCheckStopSignal() {
						return
					}
				}
			}

		case <-s.chTikerStop:
			logs.Warn("tickerApplier() stop signal")
			return
		}
	}
}

func (s *Supernumerary) fnCheckStopSignal() bool {
	select {
	case <-s.chTikerStop:
		return true
	default:
		return false
	}
}

// fnGetRandomPeer returns random peer
func (s *Supernumerary) fnGetRandomPeer(act shared.Acter) *shared.PeerInfo {
	fromUserID := act.GetUserID()
	for {
		phone := s.fnGetRandomPhoneNo()
		if a, ok := s.Actors[phone]; ok {
			if fromUserID != a.GetUserID() {
				return scenario.GetPeerInfo(fromUserID, a.GetUserID(), msg.PeerUser)
			}
		}
	}
}

// fnGetRandomPhoneNo select a random phoneNo between fromPhoneNo - toPhoneNo
func (s *Supernumerary) fnGetRandomPhoneNo() int64 {
	src := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(src)
	return s.FromPhoneNo + rnd.Int63n(s.ToPhoneNo-s.FromPhoneNo) + 1 // +1 bcz Int63n(n) returns [0,n)
}
