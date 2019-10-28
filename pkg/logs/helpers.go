package logs

import (
	"go.uber.org/zap/zapcore"
	"time"
)

/*
   Creation Time: 2019 - Oct - 28
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/


func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("06-01-02T15:04:05"))
}
