package request

import (
	"git.ronaksoft.com/river/sdk/internal/logs"
)

/*
   Creation Time: 2021 - May - 25
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

var (
	logger *logs.Logger
)

func init() {
	logger = logs.With("Request")
}

// LocalHandler SDK commands that handle user request from client cache
type LocalHandler func(da Callback)
