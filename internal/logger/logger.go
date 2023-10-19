package logger

import (
	"go.uber.org/zap"
)

// Log is a logger instance.
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
