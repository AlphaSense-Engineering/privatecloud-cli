package util

import (
	"io"

	"github.com/charmbracelet/log"
)

// LogInfoWriter is a writer that logs messages at the info level.
type LogInfoWriter struct {
	// Logger is the logger.
	Logger *log.Logger
}

var _ io.Writer = &LogInfoWriter{}

// Write is the implementation of the io.Writer interface.
func (w *LogInfoWriter) Write(p []byte) (n int, err error) {
	w.Logger.Log(log.InfoLevel, string(p[:len(p)-1]))

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
	w.Logger.Log(log.ErrorLevel, string(p[:len(p)-1]))

	return len(p), nil
}
