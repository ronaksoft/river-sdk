package logs

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var (
	_Log      *zap.Logger
	_LogLevel zap.AtomicLevel
)

func init() {
	_LogLevel = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	_Log = zap.New(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
				LineEnding:     zapcore.DefaultLineEnding,
				EncodeLevel:    zapcore.CapitalColorLevelEncoder,
				EncodeTime:     zapcore.ISO8601TimeEncoder,
				EncodeDuration: zapcore.StringDurationEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
			}),
			zapcore.Lock(os.Stdout),
			_LogLevel,
		),
	)
	_Sentry, err := NewSentryCore(zapcore.ErrorLevel, nil)
	if err != nil {
		return
	}
	_Log = _Log.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewTee(core, _Sentry)
	}))
}

func SetLogLevel(l int) {
	_LogLevel.SetLevel(zapcore.Level(l))
}

func SetHook(fn func(logLevel int, msg string)) {
	_Log.WithOptions(zap.Hooks(func(entry zapcore.Entry) error {
		fn(int(entry.Level), entry.Message)
		return nil
	}))
}

func SetLogFilePath(filePath string) error {
	if len(filePath) > 0 {
		logFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		_Log = _Log.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewTee(
				core,
				zapcore.NewCore(
					zapcore.NewJSONEncoder(zapcore.EncoderConfig{
						TimeKey:        "ts",
						LevelKey:       "level",
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
					zapcore.AddSync(logFile),
					_LogLevel,
				),
			)
		}))
	}
	return nil
}

func Message(msg string, fields ...zap.Field) {
	_Log.Debug(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	_Log.Debug(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	_Log.Warn(msg, fields...)

}

func Info(msg string, fields ...zap.Field) {
	_Log.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	_Log.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	_Log.Fatal(msg, fields...)

}
