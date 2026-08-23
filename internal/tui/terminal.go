// Package ui owns everything that writes to the user's terminal (specifically, stdout).
//
// Prompts, selection lists, spinners and the terminal interface live and here.
package tui

import (
	"io"

	"golang.org/x/term"
)

const (
	cursorHome     = "\x1b[H"      // Moves the cursor to the top-left corner.
	eraseToEnd     = "\x1b[J"      // Erases from the cursor to the end of the terminal.
	openAltBuffer  = "\x1b[?1049h" // Opens terminal alt buffer
	closeAltBuffer = "\x1b[?1049l" // Closes terminal alt buffer
)

// Terminal writes rendered output and owns the underlying output it
// writes to. The underlying terminal is detected from the writer it holds.
type Terminal struct {
	w                     io.Writer
	fd                    int
	tty                   bool
	alternateBufferActive bool
}

// fdWriter is an io.Writer with a file descriptor (e.g. *os.File).
type fdWriter interface {
	io.Writer
	Fd() uintptr
}

// NewTerminal wraps w. When w isn't a terminal, Terminal skips all escape sequences.
func NewTerminal(w io.Writer) *Terminal {
	s := &Terminal{
		w:  w,
		fd: -1,
	}

	f, ok := s.w.(fdWriter)
	if ok {
		s.fd = int(f.Fd())
		s.tty = term.IsTerminal(s.fd)
	}

	return s
}

// TTY checks if the output is a terminal.
func (s *Terminal) TTY() bool {
	return s.tty
}

// Width returns the size of the terminal. It will return 0 if it's not a terminal.
func (s *Terminal) Width() int {
	if !s.tty {
		return 0
	}

	w, _, err := term.GetSize(s.fd)
	if err != nil {
		return 0
	}

	return w
}

// Enter will open an alternate buffer if the output is a terminal and return early
// otherwise.
func (s *Terminal) Enter() error {
	if s.alternateBufferActive || !s.tty {
		return nil
	}

	if _, err := io.WriteString(s.w, openAltBuffer); err != nil {
		return err
	}

	s.alternateBufferActive = true
	return nil
}

// Exit will close an open alternate buffer if the output is a terminal and return early
// otherwise.
func (s *Terminal) Exit() error {
	if !s.tty {
		return nil
	}

	// This is safe to call when an alternate buffer is not actually active. Not guarding
	// to allow unconditional exiting in case of failure.
	if _, err := io.WriteString(s.w, closeAltBuffer); err != nil {
		return err
	}

	s.alternateBufferActive = false

	return nil
}

// Draw replaces the terminal's contents with frame in a single write. It will append
// if the output is not a terminal.
func (s *Terminal) Draw(frame string) {
	if !s.tty {
		_, _ = io.WriteString(s.w, frame)
		return
	}

	// Overwrite existing output with frame then erase to the end of the window
	_, _ = io.WriteString(s.w, cursorHome+frame+eraseToEnd)
}

// Print appends text to the terminal.
func (s *Terminal) Print(text string) {
	_, _ = io.WriteString(s.w, text)
}

// Write implements [io.Writer].
func (s *Terminal) Write(p []byte) (int, error) {
	return s.w.Write(p)
}
