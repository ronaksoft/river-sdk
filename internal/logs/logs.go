package logs

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
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
	_FileLog  *Logger
)

func init() {
	_LogLevel = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	_Log = &Logger{
		z: zap.New(
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
		),
	}
}

func SetFilePath(logDir string) error {
	// support IOS file path
	if strings.HasPrefix(logDir, "file://") {
		logDir = logDir[7:]
	}
	_LogDir = logDir
	if logDir != "" {
		fmt.Println(logDir)
		defer func() {
			// Let's clean all the old log files
			go CleanUP()
		}()
		t := time.Now()
		logFileName := fmt.Sprintf("LOG-%d-%02d-%02d.log", t.Year(), t.Month(), t.Day())
		logFile, err := os.OpenFile(path.Join(logDir, logFileName), os.O_APPEND|os.O_CREATE, 0600)
		if err != nil {
			return err
		}
		_FileLog = &Logger{
			z: zap.New(
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
			),
		}
	}
	return nil
}

func SetRemoteLog(url string) {
	remoteWriter := RemoteWrite{
		HttpClient: http.Client{
			Timeout: time.Millisecond * 250,
		},
		Url: url,
	}
	_Log.z = _Log.z.WithOptions(
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

func SetSentry(userID, authID int64, dsn string) {
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
	_Log.z = _Log.z.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewTee(core, sentry)
	}))

}

func With(name string) *Logger {
	return _Log.With(name)
}

func Directory() string {
	return _LogDir
}

func PanicF(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}
