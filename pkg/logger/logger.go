// Package logger provides structured logging capabilities for getblobz.
// It supports both text and JSON output formats with configurable log levels.
package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.SugaredLogger for structured logging.
type Logger struct {
	*zap.SugaredLogger
}

// Config contains logger configuration options.
type Config struct {
	// Level specifies the minimum log level (debug, info, warn, error).
	Level string
	// Format specifies the output format (text, json).
	Format string
}

// New creates a new Logger instance with the given configuration.
// Returns an error if the configuration is invalid.
func New(cfg Config) (*Logger, error) {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		level,
	)

	zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return &Logger{zapLogger.Sugar()}, nil
}

// Close flushes any buffered log entries.
func (l *Logger) Close() error {
	return l.Sync()
}
