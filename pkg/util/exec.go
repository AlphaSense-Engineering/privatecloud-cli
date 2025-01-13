package util

import (
	"bytes"
	"io"
	"os/exec"

	"github.com/charmbracelet/log"
)

// Exec is the function that executes a command.
func Exec(l *log.Logger, outBuf *bytes.Buffer, bin string, args ...string) error {
	// logMsgRunningCommand is the message that is logged when running a command.
	const logMsgRunningCommand = "running command: %s"

	cmd := exec.Command(bin, args...)

	var writer io.Writer

	if outBuf != nil {
		writer = io.MultiWriter(&LogDebugWriter{Logger: l}, outBuf)
	} else {
		writer = &LogDebugWriter{Logger: l}
	}

	cmd.Stdout = writer

	cmd.Stderr = &LogErrorWriter{Logger: l}

	l.Debugf(logMsgRunningCommand, cmd.String())

	return cmd.Run()
}
