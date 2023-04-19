package logging

import (
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

type encoder struct {
	zapcore.Encoder
	onError func()
}

func newEncoder(impl zapcore.Encoder, onError func()) zapcore.Encoder {
	return encoder{
		Encoder: impl,
		onError: onError,
	}
}

func (e encoder) Clone() zapcore.Encoder {
	return newEncoder(e.Encoder.Clone(), e.onError)
}

func (e encoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	if e.onError != nil && entry.Level >= zapcore.ErrorLevel {
		e.onError()
	}
	return e.Encoder.EncodeEntry(entry, fields)
}
