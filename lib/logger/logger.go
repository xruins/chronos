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
