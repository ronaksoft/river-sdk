package logs

import (
	"github.com/getsentry/raven-go"
	"testing"
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
	err := raven.SetDSN("***REMOVED***")
	if err != nil {
		t.Error(err)
	}
	raven.CaptureMessageAndWait("This is a message from Ehsan N. Moosa",
		map[string]string{
			"Label1": "v1",
			"Label2": "v2",
		},
	)
}
