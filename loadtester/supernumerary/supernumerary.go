package supernumerary

import (
	"math/rand"
	"os"
	"time"

	"git.ronaksoftware.com/ronak/riversdk/msg"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/scenario"

	"git.ronaksoftware.com/ronak/riversdk/loadtester/actor"
	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"
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
		s.Actors[i] = act
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
		s.chTikerStop <- true
		s.ticker.Stop()
		s.ticker = nil
	}

	for i := s.FromPhoneNo; i < s.ToPhoneNo; i++ {
		if act, ok := s.Actors[i]; ok {
			act.Stop()
			s.Actors[i] = nil
		}
	}
	s.Actors = nil
	s.FromPhoneNo = 0
	s.ToPhoneNo = 0
	s = nil // -_^
}

// CreateAuthKey init step required
func (s *Supernumerary) CreateAuthKey() {
	for _, act := range s.Actors {
		sen := scenario.NewCreateAuthKey(false)
		success := scenario.Play(act, sen)
		if success {
			act.Save()
		}
	}
}

// Register init step required
func (s *Supernumerary) Register() {
	for _, act := range s.Actors {
		sen := scenario.NewRegister(false)
		success := scenario.Play(act, sen)
		if success {
			act.Save()
		}
	}
}

// Login init step required
func (s *Supernumerary) Login() {
	for _, act := range s.Actors {
		sen := scenario.NewLogin(false)
		success := scenario.Play(act, sen)
		if success {
			act.Save()
		}
	}
}

// SetTickerApplier try to invoke certain action for all actors repeatedly
func (s *Supernumerary) SetTickerApplier(duration time.Duration, action TickerAction) {
	if s.ticker != nil {
		s.chTikerStop <- true
		s.ticker.Stop()
	}
	s.ticker = time.NewTicker(duration)
	go s.tickerApplier(action)
}

// tickerApplier
func (s *Supernumerary) tickerApplier(action TickerAction) {
	for {
		select {
		case <-s.ticker.C:
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
				}

			case TickerActionSendFile:
				// try to send random file
				for _, act := range s.Actors {
					sen := scenario.NewSendFile(false)
					// import random contact for actor
					act.SetPeers([]*shared.PeerInfo{s.fnGetRandomPeer(act)})
					// async
					sen.Play(act)
				}
			}

		case <-s.chTikerStop:
			return
		}
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
