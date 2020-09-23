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
	defaultChunkSize       = 512 * 1 << 10 // 256KB
	maxChunkSize           = 512 * 1 << 10 // 512KB
	maxFileSizeAllowedSize = 750 * 1 << 20 // 750 MB
	maxParts               = 3000
	retryMaxAttempts       = 25
	retryWaitTime          = 100 * time.Millisecond
)

var chunkSizesKB = []int32{128, 256, 512}
