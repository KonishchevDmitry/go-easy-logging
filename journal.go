package logging

import (
	"github.com/coreos/go-systemd/v22/journal"
	"go.uber.org/zap/zapcore"
)

type journalCore struct {
	zapcore.LevelEnabler
	encoder zapcore.Encoder
	fields  map[string]string
}

var _ zapcore.Core = &journalCore{}

func newJournalCore(name string, level zapcore.LevelEnabler, encoder zapcore.Encoder) zapcore.Core {
	return &journalCore{
		LevelEnabler: level,
		encoder:      encoder,
		fields: map[string]string{
			"SYSLOG_FACILITY":   "3",
			"SYSLOG_IDENTIFIER": name,
		},
	}
}

func (c *journalCore) With(fields []zapcore.Field) zapcore.Core {
	return c
}

func (c *journalCore) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		checkedEntry = checkedEntry.AddCore(entry, c)
	}
	return checkedEntry
}

func (c *journalCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	buf, err := c.encoder.EncodeEntry(entry, fields)
	if err != nil {
		return err
	}
	defer buf.Free()
	return journal.Send(buf.String(), journalPriority(entry.Level), c.fields)
}

func (c *journalCore) Sync() error {
	return nil
}

func journalPriority(level zapcore.Level) journal.Priority {
	switch level {

	case zapcore.DebugLevel:
		return journal.PriDebug
	case zapcore.InfoLevel:
		return journal.PriInfo
	case zapcore.WarnLevel:
		return journal.PriWarning
	case zapcore.ErrorLevel:
		return journal.PriErr
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		return journal.PriCrit
	default:
		return journal.PriErr
	}
}
