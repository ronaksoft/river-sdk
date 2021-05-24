package logs

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"time"
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
	z *zap.Logger
}

func (l *Logger) SetLogLevel(lvl int) {
	_LogLevel.SetLevel(zapcore.Level(lvl))
}

func (l *Logger) SetRemoteLog(url string) {
	remoteWriter := RemoteWrite{
		HttpClient: http.Client{
			Timeout: time.Millisecond * 250,
		},
		Url: url,
	}
	l.z = l.z.WithOptions(
		zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewTee(
				core,
				zapcore.NewCore(
					zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
						TimeKey:        "ts",
						LevelKey:       "",
						NameKey:        "logger",
						CallerKey:      "caller",
						MessageKey:     "msg",
						StacktraceKey:  "stacktrace",
						LineEnding:     zapcore.DefaultLineEnding,
						EncodeLevel:    zapcore.CapitalColorLevelEncoder,
						EncodeTime:     zapcore.ISO8601TimeEncoder,
						EncodeDuration: zapcore.StringDurationEncoder,
						EncodeCaller:   zapcore.ShortCallerEncoder,
					}),
					remoteWriter,
					_LogLevel,
				),
			)
		}),
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)
}

func (l *Logger) SetSentry(userID, authID int64, dsn string) {
	if dsn == "" {
		return
	}
	sentry, err := NewSentryCore(
		zapcore.ErrorLevel, dsn, userID,
		map[string]string{
			"AuthID": fmt.Sprintf("%d", authID),
		},
	)
	if err != nil {
		return
	}
	l.z = l.z.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewTee(core, sentry)
	}))

}

func (l *Logger) With(fields ...Field) *Logger {
	return &Logger{
		z: l.z.With(fields...),
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
	l.z.Debug(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.z.Warn(msg, fields...)
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.z.Info(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.z.Error(msg, fields...)
}

func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.z.Fatal(msg, fields...)

}
