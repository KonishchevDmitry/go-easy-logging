package logging

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/coreos/go-systemd/v22/journal"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Daemon           bool
	SyslogIdentifier string

	Level     zapcore.Level
	ShowLevel bool
	ShowTime  bool

	OnError func()
}

func Configure(config Config) (*zap.SugaredLogger, error) {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.ConsoleSeparator = " "

	var toJournal bool
	if config.Daemon {
		isLinux := runtime.GOOS == "linux"

		if isLinux && journal.Enabled() {
			toJournal = true
			encoderConfig.LineEnding = ""
		} else {
			if isLinux {
				_, _ = fmt.Fprintln(os.Stderr, "systemd journal is not available - falling back to stderr.")
			}
			config.ShowTime = true
			config.ShowLevel = true
		}
	}

	if config.ShowTime {
		encoderConfig.EncodeTime = func(timestamp time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(timestamp.Format("2006.01.02 15:04:05.000"))
		}
	} else {
		encoderConfig.TimeKey = ""
	}

	if config.ShowLevel {
		encoderConfig.EncodeLevel = func(level zapcore.Level, encoder zapcore.PrimitiveArrayEncoder) {
			encoder.AppendString(level.CapitalString()[:1] + ":")
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
