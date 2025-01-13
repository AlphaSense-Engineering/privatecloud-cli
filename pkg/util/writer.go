package util

import (
	"io"

	"github.com/charmbracelet/log"
)

// LogDebugWriter is a writer that logs messages at the debug level.
type LogDebugWriter struct {
	// Logger is the logger.
	Logger *log.Logger
}

var _ io.Writer = &LogDebugWriter{}

// Write is the implementation of the io.Writer interface.
func (w *LogDebugWriter) Write(p []byte) (n int, err error) {
	w.Logger.Debug(string(p[:len(p)-1]))

	return len(p), nil
}

// LogErrorWriter is a writer that logs messages at the error level.
type LogErrorWriter struct {
	// Logger is the logger.
	Logger *log.Logger
}

var _ io.Writer = &LogErrorWriter{}

// Write is the implementation of the io.Writer interface.
func (w *LogErrorWriter) Write(p []byte) (n int, err error) {
	w.Logger.Error(string(p[:len(p)-1]))

	return len(p), nil
}
