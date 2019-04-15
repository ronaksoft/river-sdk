package log

import "go.uber.org/zap/zapcore"

/*
   Creation Time: 2019 - Mar - 02
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

var (
	Debug = zapcore.DebugLevel
	Info  = zapcore.InfoLevel
	Warn  = zapcore.WarnLevel
	Error = zapcore.ErrorLevel
	Fatal = zapcore.FatalLevel
)
