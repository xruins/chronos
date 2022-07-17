package logger

import (
	"errors"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the interface for logging messages.
type Logger interface {
	// Printf writes a formated message to the log.
	Infof(format string, v ...interface{})

	// Print writes a message to the log.
	Info(v ...interface{})

	// wrn writes a warning message to the log and aborts.
	Warn(v ...interface{})

	// Warnf writes a warning message to the log.
	Warnf(format string, v ...interface{})

	// Error writes an error message to the log and aborts.
	Error(v ...interface{})

	// Errorf writes an error message to the log.
	Errorf(format string, v ...interface{})

	// Debug writes a debug message to the log and aborts.
	Debug(v ...interface{})

	// Debugf writes a debug message to the log.
	Debugf(format string, v ...interface{})

	// Fatal writes a message to the log and aborts.
	Fatal(v ...interface{})

	// Fatalf writes a formated message to the log and aborts.
	Fatalf(format string, v ...interface{})
}

func zapLogLevel(level string) (zap.AtomicLevel, error) {
	var lvl zapcore.Level
	switch level {
	case "debug":
		lvl = zap.DebugLevel
	case "info":
		lvl = zap.InfoLevel
	case "warn":
		lvl = zap.WarnLevel
	case "error":
		lvl = zap.ErrorLevel
	case "":
		lvl = zap.InfoLevel
	default:
		return zap.AtomicLevel{}, errors.New("malformed level")
	}

	return zap.NewAtomicLevelAt(lvl), nil
}

func NewZapLogger(level string) (*zap.SugaredLogger, error) {
	logLevel, err := zapLogLevel(string(level))
	if err != nil {
		return nil, fmt.Errorf("failed to get loglevel: %s", err)
	}

	config := zap.Config{
		Level:    logLevel,
		Encoding: "console",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:    "Time",
			LevelKey:   "Level",
			NameKey:    "Name",
			CallerKey:  "Caller",
			MessageKey: "Msg",
			//StacktraceKey:  "St",
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to generate logger: %w", err)
	}
	return logger.Sugar(), nil
}

// NopLogger is the logger for testing.
// it won't output anything on all log levels.
type NopLogger struct{}

// Printf does nothing
func (n *NopLogger) Infof(_ string, _ ...interface{}) {
	return
}

// Print does nothing
func (n *NopLogger) Info(v ...interface{}) {
	return
}

// Warn does nothing
func (n *NopLogger) Warn(v ...interface{}) {
	return
}

// Warnf does nothing
func (n *NopLogger) Warnf(_ string, _ ...interface{}) {
	return
}

// Error does nothing
func (n *NopLogger) Error(v ...interface{}) {
	return
}

// Errorf does nothing
func (n *NopLogger) Errorf(_ string, _ ...interface{}) {
	return
}

// Debug does nothing
func (n *NopLogger) Debug(_ ...interface{}) {
	return
}

// Debugf does nothing
func (n *NopLogger) Debugf(_ string, _ ...interface{}) {
	return
}

// Fatal does nothing
func (n *NopLogger) Fatal(_ ...interface{}) {
	return
}

// Fatalf does nothing
func (n *NopLogger) Fatalf(_ string, _ ...interface{}) {
	return
}
