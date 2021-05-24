package logs

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path"
	"strings"
	"time"
)

type (
	AtomicLevel = zap.AtomicLevel
	Field       = zap.Field
)

var (
	_LogDir   string
	_LogLevel AtomicLevel
	_Log      *Logger
)

func init() {
	_LogLevel = zap.NewAtomicLevelAt(zapcore.WarnLevel)
}

func New(logDir string) (*Logger, error) {
	defer func() {
		// Let's clean all the old log files
		go CleanUP()
	}()
	// support IOS file path
	if strings.HasPrefix(logDir, "file://") {
		logDir = logDir[7:]
	}
	_LogDir = logDir

	l := &Logger{}
	l.z = zap.New(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
				CallerKey:      "caller",
				LevelKey:       "level",
				MessageKey:     "msg",
				NameKey:        "name",
				StacktraceKey:  "stack",
				LineEnding:     zapcore.DefaultLineEnding,
				EncodeLevel:    zapcore.CapitalColorLevelEncoder,
				EncodeTime:     TimeEncoder,
				EncodeDuration: zapcore.StringDurationEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
			}),
			zapcore.Lock(os.Stdout),
			_LogLevel,
		),
	)

	if logDir != "" {
		t := time.Now()
		logFileName := fmt.Sprintf("LOG-%d-%02d-%02d.log", t.Year(), t.Month(), t.Day())
		_ = os.MkdirAll(logDir, 0600)
		logFile, err := os.OpenFile(path.Join(logDir, logFileName), os.O_APPEND|os.O_CREATE, 0600)
		if err != nil {
			return nil, err
		}
		l.z = l.z.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewTee(
				core,
				zapcore.NewCore(
					zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
						TimeKey:        "ts",
						LevelKey:       "level",
						NameKey:        "logger",
						CallerKey:      "caller",
						MessageKey:     "msg",
						StacktraceKey:  "stacktrace",
						LineEnding:     zapcore.DefaultLineEnding,
						EncodeLevel:    zapcore.CapitalLevelEncoder,
						EncodeTime:     TimeEncoder,
						EncodeDuration: zapcore.StringDurationEncoder,
						EncodeCaller:   zapcore.ShortCallerEncoder,
					}),
					zapcore.Lock(logFile),
					_LogLevel,
				),
			)
		}))
	}
	_Log = l
	return l, nil
}

func Directory() string {
	return _LogDir
}

func SetLogLevel(l int) {
	_LogLevel.SetLevel(zapcore.Level(l))
}

func Debug(msg string, fields ...zap.Field) {
	if _Log == nil {
		return
	}
	_Log.Debug(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	if _Log == nil {
		return
	}
	_Log.Warn(msg, fields...)
}

func WarnOnErr(guideTxt string, err error, fields ...zap.Field) {
	if _Log == nil {
		return
	}
	if err != nil {
		fields = append(fields, zap.Error(err))
		_Log.Warn(guideTxt, fields...)
	}
}

func Info(msg string, fields ...zap.Field) {
	if _Log == nil {
		return
	}
	_Log.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	if _Log == nil {
		return
	}
	_Log.Error(msg, fields...)
}

func ErrorOnErr(guideTxt string, err error, fields ...zap.Field) {
	if _Log == nil {
		return
	}
	if err != nil {
		fields = append(fields, zap.Error(err))
		_Log.Error(guideTxt, fields...)
	}
}

func Fatal(msg string, fields ...zap.Field) {
	if _Log == nil {
		return
	}
	_Log.Fatal(msg, fields...)
}

func PanicF(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}
