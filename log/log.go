package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	LOG       *zap.Logger
	LOG_LEVEL zap.AtomicLevel
)

func init() {

	LOG_LEVEL = zap.NewAtomicLevelAt(zap.DebugLevel)
	logConfig := zap.NewProductionConfig()
	logConfig.Encoding = "console"
	logConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logConfig.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	logConfig.Level = LOG_LEVEL
	if l, err := logConfig.Build(); err != nil {
		os.Exit(1)
	} else {
		LOG = l
	}

}

func SetLogLevel(l int) {
	LOG_LEVEL.SetLevel(zapcore.Level(l))
}
