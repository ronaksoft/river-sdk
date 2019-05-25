package log

import (
	"github.com/getsentry/raven-go"
	"go.uber.org/zap/zapcore"
)

/*
   Creation Time: 2019 - Apr - 24
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

const (
	sentryDSN = "***REMOVED***"
)

func NewSentryCore(level zapcore.Level, tags map[string]string) (zapcore.Core, error) {
	client, err := raven.New(sentryDSN)
	if err != nil {
		return zapcore.NewNopCore(), err
	}
	return &sentryCore{
		client:       client,
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

	packet := &raven.Packet{
		Message:   ent.Message,
		Timestamp: raven.Timestamp(ent.Time),
		Level:     ravenSeverity(ent.Level),
		Platform:  "Golang",
		Extra:     clone.fields,
	}

	_, _ = c.client.Capture(packet, c.tags)

	// We may be crashing the program, so should flush any buffered events.
	if ent.Level > zapcore.ErrorLevel {
		c.client.Wait()
	}
	return nil
}

func (c *sentryCore) Sync() error {
	c.client.Wait()
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
		client:       c.client,
		tags:         c.tags,
		fields:       m,
		LevelEnabler: c.LevelEnabler,
	}
}

type ClientGetter interface {
	GetClient() *raven.Client
}

func (c *sentryCore) GetClient() *raven.Client {
	return c.client
}

type sentryCore struct {
	client *raven.Client
	zapcore.LevelEnabler

	tags   map[string]string
	fields map[string]interface{}
}

func ravenSeverity(lvl zapcore.Level) raven.Severity {
	switch lvl {
	case zapcore.DebugLevel:
		return raven.INFO
	case zapcore.InfoLevel:
		return raven.INFO
	case zapcore.WarnLevel:
		return raven.WARNING
	case zapcore.ErrorLevel:
		return raven.ERROR
	case zapcore.DPanicLevel:
		return raven.FATAL
	case zapcore.PanicLevel:
		return raven.FATAL
	case zapcore.FatalLevel:
		return raven.FATAL
	default:
		// Unrecognized levels are fatal.
		return raven.FATAL
	}
}
