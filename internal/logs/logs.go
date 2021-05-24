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
		logFile, err := os.OpenFile(path.Join(logDir, logFileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
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
