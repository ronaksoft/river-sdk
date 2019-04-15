package ronak

import (
	"git.ronaksoftware.com/ronak/toolbox/logger"
)

/*
   Creation Time: 2018 - Apr - 07
   Created by:  Ehsan N. Moosa (ehsan)
   Maintainers:
       1.  Ehsan N. Moosa (ehsan)
   Auditor: Ehsan N. Moosa
   Copyright Ronak Software Group 2018
*/

var (
	_Log log.Logger
)

type (
	M map[string]interface{}
	MS map[string]string
	MI map[string]int64
)

func init() {
	_Log = log.NewNop()
}

func SetLogger(l log.Logger) {
	_Log = l
}
