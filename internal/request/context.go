package request

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

type Context struct {
	Callback     Callback
	Timeout      time.Duration
	Flags        DelegateFlag
}
