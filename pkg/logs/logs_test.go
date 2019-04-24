package logs

import (
	"testing"
	"time"
)

/*
   Creation Time: 2019 - Apr - 24
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

func TestSentry(t *testing.T) {
	_Log.Debug("This should not be sentry")
	_Log.Error("This goes to sentry")
	time.Sleep(5 * time.Second)
}
