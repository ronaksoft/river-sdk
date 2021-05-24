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
	parent *Logger
	z      *zap.Logger
}

func (l *Logger) SetLogLevel(lvl int) {
	_LogLevel.SetLevel(zapcore.Level(lvl))
}

func (l *Logger) With(name string) *Logger {
	return &Logger{
		prefix: name,
		parent: l,
	}
}

func (l *Logger) write(lvl zapcore.Level, msg string, fields ...zap.Field) {
	if l.prefix != "" {
		msg = fmt.Sprintf("[%s]: %s", l.prefix, msg)
	}
	if l.parent != nil {
		l.parent.write(lvl, msg, fields...)
		return
	}
	if ce := l.z.Check(lvl, msg); ce != nil {
		ce.Write(fields...)
	}
	if _FileLog != nil {
		if ce := _FileLog.z.Check(lvl, msg); ce != nil {
			ce.Write(fields...)
		}
	}
}

func (l *Logger) WarnOnErr(guideTxt string, err error, fields ...zap.Field) {
	if err != nil {
		fields = append(fields, zap.Error(err))
		l.Warn(guideTxt, fields...)
	}
}

func (l *Logger) ErrorOnErr(guideTxt string, err error, fields ...zap.Field) {
	if err != nil {
		fields = append(fields, zap.Error(err))
		l.Error(guideTxt, fields...)
	}
}

func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.write(zap.DebugLevel, msg, fields...)

}

func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.write(zap.WarnLevel, msg, fields...)
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.write(zap.InfoLevel, msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.write(zap.ErrorLevel, msg, fields...)
}

func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.write(zap.FatalLevel, msg, fields...)

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
