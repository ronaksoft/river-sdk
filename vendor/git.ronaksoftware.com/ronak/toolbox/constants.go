package ronak

import "time"

/*
   Creation Time: 2019 - Feb - 28
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

const (
	defaultMaxBatchSize       = 1000
	defaultFlushPeriod        = 25 * time.Millisecond
	defaultMaxParallelWorkers = 10
)
