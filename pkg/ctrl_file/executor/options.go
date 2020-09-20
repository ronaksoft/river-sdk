package executor

/*
   Creation Time: 2020 - Sep - 20
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

type Option func(e *Executor)

func WithConcurrency(c int32) Option {
	return func(e *Executor) {
		e.rt = make(chan struct{}, c)
	}
}
