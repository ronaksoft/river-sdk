package testenv

import (
    "github.com/ronaksoft/river-sdk/internal/logs"
)

/*
   Creation Time: 2021 - May - 24
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

var logger *logs.Logger

func init() {
    logger = logs.With("TEST")
}

func Log() *logs.Logger {
    return logger
}
