package logging

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Daemon           bool
	SyslogIdentifier string

	Level     zapcore.Level
	ShowLevel bool

	OnError func()
}

func Configure(config Config) (*zap.SugaredLogger, error) {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = ""
	encoderConfig.ConsoleSeparator = " "

	toJournal := config.Daemon
	if toJournal {
		encoderConfig.LineEnding = ""
	}

	if config.ShowLevel {
		encoderConfig.EncodeLevel = func(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(level.CapitalString()[:1] + ":")
		}
	} else {
		encoderConfig.LevelKey = ""
	}

	encoder := newEncoder(zapcore.NewConsoleEncoder(encoderConfig), config.OnError)

	var core zapcore.Core
	if toJournal {
		if config.SyslogIdentifier == "" {
			return nil, errors.New("syslog identifier is not set")
		}
		core = newJournalCore(config.SyslogIdentifier, config.Level, encoder)
	} else {
		core = newStdoutCore(config.Level, encoder)
	}

	return zap.New(core).Sugar(), nil
}

type contextKey struct{}

func L(ctx context.Context) *zap.SugaredLogger {
	return ctx.Value(contextKey{}).(*zap.SugaredLogger)
}

func WithLogger(ctx context.Context, logger *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, contextKey{}, logger)
}
