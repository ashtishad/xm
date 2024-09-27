package common

import (
	"log/slog"
	"os"
	"path/filepath"
)

// NewSlogger sets up a logger level based on app env, source information, and simplified file paths.
//
// Usage:
// logger := common.NewLogger()
//
//	slog.SetDefault(logger)
//	slog.Info("Application started", "version", "1.0.0")
func NewSlogger() *slog.Logger {
	opts := getHandlerOpts()

	handler := slog.NewTextHandler(os.Stdout, opts)
	return slog.New(handler)
}

// getHandlerOpts strips the full directory path from the source's filename and sets LevelDebug
//
// Returns:
//   - *slog.HandlerOptions: Pointer to a slog.HandlerOptions struct containing the logging configurations.
func getHandlerOpts() *slog.HandlerOptions {
	replace := func(groups []string, a slog.Attr) slog.Attr {
		// Remove the directory from the source's filename.
		if a.Key == slog.SourceKey {
			sourceVal, ok := a.Value.Any().(*slog.Source)
			if !ok {
				return a
			}

			sourceVal.File = filepath.Base(sourceVal.File)
		}

		return a
	}

	return &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: replace,
	}
}
