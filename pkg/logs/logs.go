package logs

import (
	"fmt"
	"git.ronaksoftware.com/ronak/riversdk/msg/ext"
	"git.ronaksoftware.com/ronak/riversdk/pkg/domain"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

var (
	_UpdateLog *zap.Logger
	_Log       *zap.Logger
	_LogLevel  zap.AtomicLevel
)

func init() {
	_LogLevel = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	_Log = zap.New(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
				CallerKey:      "caller",
				LevelKey:       "level",
				MessageKey:     "msg",
				NameKey:        "name",
				StacktraceKey:  "stack",
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

}

func SetLogLevel(l int) {
	_LogLevel.SetLevel(zapcore.Level(l))
}

func SetLogFilePath(logDir string) error {
	// support IOS file path
	if strings.HasPrefix(logDir, "file://") {
		logDir = logDir[7:]
	}

	t := time.Now()
	logFileName := fmt.Sprintf("LOG-%d-%02d-%02d.log", t.Year(), t.Month(), t.Day())
	logFile, err := os.OpenFile(path.Join(logDir, logFileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_Log = _Log.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
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
					EncodeLevel:    zapcore.CapitalColorLevelEncoder,
					EncodeTime:     zapcore.ISO8601TimeEncoder,
					EncodeDuration: zapcore.StringDurationEncoder,
					EncodeCaller:   zapcore.ShortCallerEncoder,
				}),
				zapcore.Lock(logFile),
				_LogLevel,
			),
		)
	}))

	updateLogFileName := fmt.Sprintf("UPDT-%04d-%02d-%02d.log", t.Year(), t.Month(), t.Day())
	updateLogFile, err := os.OpenFile(path.Join(logDir, updateLogFileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_UpdateLog = zap.New(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
				TimeKey:        "ts",
				LineEnding:     zapcore.DefaultLineEnding,
				EncodeLevel:    zapcore.CapitalLevelEncoder,
				EncodeTime:     zapcore.EpochTimeEncoder,
				EncodeDuration: zapcore.StringDurationEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
			}),
			zapcore.Lock(updateLogFile),
			_LogLevel,
		),
	)
	return nil
}

func SetRemoteLog(url string) {
	remoteWriter := RemoteWrite{
		HttpClient: http.Client{
			Timeout: time.Millisecond * 250,
		},
		Url: url,
	}
	_Log = _Log.WithOptions(
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

func SetSentry(userID, authID int64) {
	sentry, err := NewSentryCore(zapcore.ErrorLevel, map[string]string{
		"AuthID": fmt.Sprintf("%d", authID),
		"UserID": fmt.Sprintf("%d", userID),
		"SDK":    domain.SDKVersion,
	})
	if err != nil {
		return
	}
	_Log = _Log.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewTee(core, sentry)
	}))

}
func Debug(msg string, fields ...zap.Field) {
	_Log.Debug(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	_Log.Warn(msg, fields...)
}

func WarnOnErr(guideTxt string, err error, fields ...zap.Field) {
	if err != nil {
		fields = append(fields, zap.Error(err))
		Warn(guideTxt, fields...)
	}
}

func ErrorOnErr(guideTxt string, err error, fields ...zap.Field) {
	if err != nil {
		fields = append(fields, zap.Error(err))
		Error(guideTxt, fields...)
	}
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

func UpdateLog(updateID int64, constructor int64) {
	if _UpdateLog != nil {
		constructorName, _ := msg.ConstructorNames[constructor]
		_UpdateLog.Info("Update",
			zap.Int64("ID", updateID),
			zap.String("Constructor", constructorName),
			zap.Int64("ConstructorID", constructor),
		)
	}
}
