package logs

import (
	"github.com/getsentry/sentry-go"
	"go.uber.org/zap/zapcore"
	"time"
)

/*
   Creation Time: 2019 - Apr - 24
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

func NewSentryCore(level zapcore.Level, tags map[string]string) (zapcore.Core, error) {
	err := sentry.Init(sentry.ClientOptions{
		Dsn: "***REMOVED***",
		Release: "v0.5.0",
	})
	if err != nil {
		return zapcore.NewNopCore(), err
	}

	return &sentryCore{
		hub:          sentry.CurrentHub(),
		tags:         tags,
		LevelEnabler: level,
		fields:       make(map[string]interface{}),
	}, nil
}

func (c *sentryCore) With(fs []zapcore.Field) zapcore.Core {
	return c.with(fs)
}

func (c *sentryCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.LevelEnabler.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

func (c *sentryCore) Write(ent zapcore.Entry, fs []zapcore.Field) error {
	clone := c.with(fs)

	event := sentry.NewEvent()
	event.Message = ent.Message
	event.Timestamp = ent.Time.Unix()
	event.Level = sentryLevel(ent.Level)
	event.Platform = ""
	event.Extra = clone.fields
	c.hub.CaptureEvent(event)


	// We may be crashing the program, so should flush any buffered events.
	if ent.Level > zapcore.ErrorLevel {
		c.hub.Flush(time.Second)
	}
	return nil
}

func (c *sentryCore) Sync() error {
	c.hub.Flush(time.Second * 3)
	return nil
}

func (c *sentryCore) with(fs []zapcore.Field) *sentryCore {
	// Copy our map.
	m := make(map[string]interface{}, len(c.fields))
	for k, v := range c.fields {
		m[k] = v
	}

	// Add fields to an in-memory encoder.
	enc := zapcore.NewMapObjectEncoder()
	for _, f := range fs {
		f.AddTo(enc)
	}

	// Merge the two maps.
	for k, v := range enc.Fields {
		m[k] = v
	}

	return &sentryCore{
		hub:          c.hub,
		tags:         c.tags,
		fields:       m,
		LevelEnabler: c.LevelEnabler,
	}
}

type ClientGetter interface {
	GetHub() *sentry.Hub
}

func (c *sentryCore) GetHub() *sentry.Hub {
	return c.hub
}

type sentryCore struct {
	hub *sentry.Hub
	zapcore.LevelEnabler

	tags   map[string]string
	fields map[string]interface{}
}

func sentryLevel(lvl zapcore.Level) sentry.Level {
	switch lvl {
	case zapcore.DebugLevel:
		return sentry.LevelDebug
	case zapcore.InfoLevel:
		return sentry.LevelInfo
	case zapcore.WarnLevel:
		return sentry.LevelWarning
	case zapcore.ErrorLevel:
		return sentry.LevelError
	case zapcore.DPanicLevel:
		return sentry.LevelFatal
	case zapcore.PanicLevel:
		return sentry.LevelFatal
	case zapcore.FatalLevel:
		return sentry.LevelFatal
	default:
		// Unrecognized levels are fatal.
		return sentry.LevelFatal
	}
}
