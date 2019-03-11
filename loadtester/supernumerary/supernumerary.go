package supernumerary

import (
	"time"

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
	sn := &Supernumerary{
		Actors:      make(map[int64]shared.Acter),
		FromPhoneNo: fromPhoneNo,
		ToPhoneNo:   toPhoneNo,
		chTikerStop: make(chan bool),
	}

	for i := fromPhoneNo; i < toPhoneNo; i++ {
		phone := shared.GetPhone(i)
		act, err := actor.NewActor(phone)
		if err != nil {
			defer sn.Dispose()
			return nil, err
		}
		sn.Actors[i] = act
	}
	return sn, nil
}

// Dispose stops actors and cleans up
func (s *Supernumerary) Dispose() {

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
			}

		case <-s.chTikerStop:
			return
		}
	}
}
