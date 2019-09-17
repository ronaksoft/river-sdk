package fileCtrl

import (
	"time"
)

/*
   Creation Time: 2019 - Aug - 18
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

const (
	downloadChunkSize    = 256 * 1 << 10 // 256 KB
	maxDownloadInFlights = 10
	retryMaxAttempts     = 50
	retryWaitTime        = 100 * time.Millisecond
)
