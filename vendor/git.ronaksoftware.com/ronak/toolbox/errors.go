package ronak

import "errors"

/*
   Creation Time: 2018 - May - 06
   Created by:  (ehsan)
   Maintainers:
       1.  (ehsan)
   Auditor: Ehsan N. Moosa
   Copyright Ronak Software Group 2018
*/

var (
	ErrServerBusy    = errors.New("server busy")
	ErrTimeout       = errors.New("time out")
	ErrIncorrectSize = errors.New("incorrect size")
)
