package domain

import (
	"time"
)

/*
   Creation Time: 2021 - May - 24
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

type ExecuteContext struct {
	RequestID    int64
	TeamID       int64
	TeamAccess   uint64
	Constructor  int64
	CommandBytes []byte
	Callback     Callback
	Timeout      time.Duration
	Flags        RequestDelegateFlag
}
