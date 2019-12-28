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
	defaultChunkSize       = 256 * 1 << 10 // 256 KB
	maxChunkSize           = 512 * 1 << 10 // 512KB
	maxFileSizeAllowedSize = 750 * 1 << 20 // 750 MB
	maxDownloadInFlights   = 10
	maxUploadInFlights     = 10
	maxParts               = 3000
	retryMaxAttempts       = 200
	retryWaitTime          = 100 * time.Millisecond
)

var chunkSizesKB = []int32{1, 2, 4, 8, 16, 32, 64, 128, 256, 512}
