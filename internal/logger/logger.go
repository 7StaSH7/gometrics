// Package logger provides a global logger instance.
package logger

import "go.uber.org/zap"

// Log is the global logger instance.
var Log *zap.Logger = zap.NewNop()

// Initialize sets up the global logger with the specified level.
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = zl

	return nil
}
