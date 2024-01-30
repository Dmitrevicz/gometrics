// Package logger is based on go.uber.org/zap.
package logger

import (
	"go.uber.org/zap"
)

// Log is a logger instance.
// Log variable must only be changed by Initialize function.
// No-op Logger is set by default, so must be Initialized.
var Log *zap.Logger = zap.NewNop()

// Initialize configures logger with provided level.
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	// cfg := zap.NewProductionConfig()
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = zl

	return nil
}

// Sync ignores err check (to avoid lint warning).
// Otherwise use logger.Log.Sync().
func Sync() {
	_ = Log.Sync()
}
