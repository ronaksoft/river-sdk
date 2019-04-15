package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

/*
   Creation Time: 2019 - Mar - 02
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

type Level = zapcore.Level
type Field = zapcore.Field
type CheckedEntry = zapcore.CheckedEntry

type Logger interface {
	Log(level Level, msg string, fields ...Field)
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	Check(Level, string) *CheckedEntry
	Sync() error
	SetLevel(level Level)
}

func log(l *zap.Logger, level Level, msg string, fields ...Field) {
	l.Check(level, msg).Write(fields...)
}

type zapLogger struct {
	*zap.Logger
	zap.AtomicLevel
}

func (l *zapLogger) Log(level Level, msg string, fields ...Field) {
	log(l.Logger, level, msg, fields...)
}

func (l *zapLogger) Debug(msg string, fields ...Field) {
	log(l.Logger, Debug, msg, fields...)
}

func (l *zapLogger) Info(msg string, fields ...Field) {
	log(l.Logger, Info, msg, fields...)
}

func (l *zapLogger) Warn(msg string, fields ...Field) {
	log(l.Logger, Warn, msg, fields...)
}

func (l *zapLogger) Error(msg string, fields ...Field) {
	log(l.Logger, Error, msg, fields...)
}

func (l *zapLogger) Fatal(msg string, fields ...Field) {
	log(l.Logger, Fatal, msg, fields...)
}

func (l *zapLogger) Check(level Level, msg string) *CheckedEntry {
	return l.Logger.Check(level, msg)
}

func (l *zapLogger) SetLevel(level Level) {
	l.AtomicLevel.SetLevel(level)
}

func NewConsoleLogger() *zapLogger {
	l := new(zapLogger)
	l.AtomicLevel = zap.NewAtomicLevel()
	consoleWriteSyncer := zapcore.Lock(os.Stdout)
	consoleEncoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "zapLogger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})
	l.Logger = zap.New(
		zapcore.NewCore(consoleEncoder, consoleWriteSyncer, l.AtomicLevel),
		zap.AddCaller(),
		zap.AddCallerSkip(2),
	)
	return l
}

func NewFileLogger(filename string) *zapLogger {
	l := new(zapLogger)

	l.AtomicLevel = zap.NewAtomicLevel()
	fileLog, _ := os.Create(filename)

	syncer := zapcore.Lock(fileLog)
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.EpochTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})
	l.Logger = zap.New(
		zapcore.NewCore(encoder, syncer, l.AtomicLevel),
		zap.AddCaller(),
		zap.AddCallerSkip(2),
	)
	return l
}

func NewZapLogger(core zapcore.Core) *zapLogger {
	l := new(zapLogger)
	l.Logger = zap.New(
		core,
		zap.AddCaller(),
		zap.AddCallerSkip(2),
	)
	return l
}

func NewNop() *zapLogger {
	l := new(zapLogger)
	l.AtomicLevel = zap.NewAtomicLevel()
	l.Logger = zap.NewNop()
	return l
}
