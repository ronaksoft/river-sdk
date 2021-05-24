package logs

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"runtime/debug"
)

/*
   Creation Time: 2021 - May - 24
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

type Logger struct {
	prefix string
	z      *zap.Logger
}

func (l *Logger) SetLogLevel(lvl int) {
	_LogLevel.SetLevel(zapcore.Level(lvl))
}

func (l *Logger) With(name string, fields ...Field) *Logger {
	return &Logger{
		prefix: name,
		z:      l.z.With(fields...),
	}
}

func (l *Logger) WarnOnErr(guideTxt string, err error, fields ...zap.Field) {
	if err != nil {
		fields = append(fields, zap.Error(err))
		l.z.Warn(guideTxt, fields...)
	}
}

func (l *Logger) ErrorOnErr(guideTxt string, err error, fields ...zap.Field) {
	if err != nil {
		fields = append(fields, zap.Error(err))
		l.z.Error(guideTxt, fields...)
	}
}

func (l *Logger) Debug(msg string, fields ...zap.Field) {
	msg = fmt.Sprintf("[%s]: %s", l.prefix, msg)
	l.z.Debug(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...zap.Field) {
	msg = fmt.Sprintf("[%s]: %s", l.prefix, msg)
	l.z.Warn(msg, fields...)
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	msg = fmt.Sprintf("[%s]: %s", l.prefix, msg)
	l.z.Info(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	msg = fmt.Sprintf("[%s]: %s", l.prefix, msg)
	l.z.Error(msg, fields...)
}

func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	msg = fmt.Sprintf("[%s]: %s", l.prefix, msg)
	l.z.Fatal(msg, fields...)

}

func (l *Logger) RecoverPanic(funcName string, extraInfo interface{}, compensationFunc func()) {
	if r := recover(); r != nil {
		l.Error("Panic Recovered",
			zap.String("Func", funcName),
			zap.Any("Info", extraInfo),
			zap.Any("Recover", r),
			zap.ByteString("StackTrace", debug.Stack()),
		)
		if compensationFunc != nil {
			go compensationFunc()
		}
	}
}
