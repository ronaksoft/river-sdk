package fileCtrl

import "time"

/*
   Creation Time: 2019 - Aug - 18
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

const (
	retryMaxAttempts              = 10
	retryWaitTime                 = 100 * time.Millisecond
)