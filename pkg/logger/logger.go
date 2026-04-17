package logger

import (
	"context"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	once   sync.Once
	global *zap.Logger
)

// Logger defines the interface for structured logging.
type Logger interface {
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
}

// Config holds the configuration for the logger initialization.
type Config struct {
	Level      string `json:"level"`
	Production bool   `json:"production"`
}

// Init initializes the global logger singleton.
func Init(cfg Config) error {
	var err error
	once.Do(func() {
		var zapCfg zap.Config
		if cfg.Production {
			zapCfg = zap.NewProductionConfig()
		} else {
			zapCfg = zap.NewDevelopmentConfig()
		}

		var level zapcore.Level
		if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
			level = zap.InfoLevel
		}
		zapCfg.Level = zap.NewAtomicLevelAt(level)

		global, err = zapCfg.Build(zap.AddCallerSkip(1))
	})
	return err
}

// New returns the global logger instance as a Logger interface.
func New() Logger {
	if global == nil {
		l, _ := zap.NewDevelopment()
		return l
	}
	return global
}

// L returns the underlying zap.Logger.
func L() *zap.Logger {
	if global == nil {
		l, _ := zap.NewDevelopment()
		return l
	}
	return global
}

// ContextLogger returns a logger with context fields.
func ContextLogger(ctx context.Context) Logger {
	return New()
}

// Unified logging functions for global access
func Info(msg string, fields ...zap.Field)  { L().Info(msg, fields...) }
func Warn(msg string, fields ...zap.Field)  { L().Warn(msg, fields...) }
func Error(msg string, fields ...zap.Field) { L().Error(msg, fields...) }
func Debug(msg string, fields ...zap.Field) { L().Debug(msg, fields...) }
func Fatal(msg string, fields ...zap.Field) { L().Fatal(msg, fields...) }

// Field aliases
var (
	String   = zap.String
	Int      = zap.Int
	Int64    = zap.Int64
	Float64  = zap.Float64
	Bool     = zap.Bool
	Duration = zap.Duration
	Any      = zap.Any
	ErrorF   = zap.Error
)

func Sync() error {
	if global != nil {
		return global.Sync()
	}
	return nil
}
