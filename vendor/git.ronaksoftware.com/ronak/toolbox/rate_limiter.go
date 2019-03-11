package ronak

import "time"

/*
   Creation Time: 2018 - May - 06
   Created by:  (ehsan)
   Maintainers:
       1.  (ehsan)
   Auditor: Ehsan N. Moosa
   Copyright Ronak Software Group 2018
*/

type RateLimiter interface {
	Enter() error
	Leave()
}

// SimpleRateLimiter
// It implements RateLimiter interface and has two input options, 'waitingTimeout' sets the amount
// that Hit() function must wait before returning error
type SimpleRateLimiter struct {
	waitingTimeout time.Duration
	ch             chan bool
}

// NewSimpleRateLimiter is the constructor of the SimpleRateLimiter
func NewSimpleRateLimiter(waitingTimeout time.Duration, size int) (*SimpleRateLimiter, error) {
	rateLimiter := new(SimpleRateLimiter)
	if size <= 0 {
		return nil, ErrIncorrectSize
	}
	rateLimiter.ch = make(chan bool, size)
	if waitingTimeout > 0 {
		rateLimiter.waitingTimeout = waitingTimeout
	}
	return rateLimiter, nil
}

func (r *SimpleRateLimiter) Enter() error {
	select {
	case r.ch <- true:
	case <-time.After(r.waitingTimeout):
		return ErrTimeout
	}
	return nil
}

func (r *SimpleRateLimiter) Leave() {
	select {
	case <-r.ch:
	case <-time.After(r.waitingTimeout):
		// This must not happen
		panic("SimpleRateLimiter:: reading channel")
	}
}
