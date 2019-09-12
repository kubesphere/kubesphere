package term

import (
	"fmt"
	"github.com/docker/docker/pkg/term"
	"io"
)

// TerminalSize returns the current width and height of the user's terminal. If it isn't a terminal,
// nil is returned. On error, zero values are returned for width and height.
// Usually w must be the stdout of the process. Stderr won't work.
func TerminalSize(w io.Writer) (int, int, error) {
	outFd, isTerminal := term.GetFdInfo(w)
	if !isTerminal {
		return 0, 0, fmt.Errorf("given writer is no terminal")
	}
	winsize, err := term.GetWinsize(outFd)
	if err != nil {
		return 0, 0, err
	}
	return int(winsize.Width), int(winsize.Height), nil
}
