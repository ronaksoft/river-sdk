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
	DefaultChunkSize       = 512 * 1 << 10 // 512KB
	MaxChunkSize           = 512 * 1 << 10 // 512KB
	MaxFileSizeAllowedSize = 750 * 1 << 20 // 750 MB
	MaxParts               = 3000
	retryMaxAttempts       = 25
	retryWaitTime          = 100 * time.Millisecond
)

var chunkSizesKB = []int32{256, 512}
