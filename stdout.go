package logging

import (
	"os"
	"sync"

	"go.uber.org/zap/zapcore"
)

type stdoutCore struct {
	zapcore.LevelEnabler
	encoder zapcore.Encoder
	lock    sync.Mutex
}

var _ zapcore.Core = &stdoutCore{}

func newStdoutCore(level zapcore.LevelEnabler, encoder zapcore.Encoder) zapcore.Core {
	return &stdoutCore{
		LevelEnabler: level,
		encoder:      encoder,
	}
}

func (c *stdoutCore) With(fields []zapcore.Field) zapcore.Core {
	return c
}

func (c *stdoutCore) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		checkedEntry = checkedEntry.AddCore(entry, c)
	}
	return checkedEntry
}

func (c *stdoutCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	buf, err := c.encoder.EncodeEntry(entry, fields)
	if err != nil {
		return err
	}
	defer buf.Free()

	file := os.Stdout
	if entry.Level >= zapcore.WarnLevel {
		file = os.Stderr
	}

	c.lock.Lock()
	defer c.lock.Unlock()
	_, _ = file.Write(buf.Bytes())

	return nil
}

func (c *stdoutCore) Sync() error {
	return nil
}
