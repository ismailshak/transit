package tui

import (
	"github.com/ismailshak/transit/internal/logger"
)

// An alternate buffer drawn on top of the active terminal screen.
// This buffer will not erase the user's current terminal, but render itself
// on top of the existing buffer.
type TerminalBuffer struct {
	alternateBufferActive bool
}

// Returns an alternate buffer instance
func NewBuffer() *TerminalBuffer {
	return &TerminalBuffer{}
}

// Renders a buffer on top of the current one
func (b *TerminalBuffer) StartAlternateBuffer() {
	if !b.alternateBufferActive {
		logger.Print("\x1b[?1049h")
		b.alternateBufferActive = true
	}
}

// Closes the alternate buffer and returns to the original buffer
func (b *TerminalBuffer) StopAlternateBuffer() {
	if b.alternateBufferActive {
		logger.Print("\x1b[?1049l")
		b.alternateBufferActive = false
	}
}

// Erase the entire content of the buffer and leaves
// cursor in the in the first row and column of the buffer
func (b *TerminalBuffer) RefreshScreen() {
	// Move cursor to 0,0
	logger.Print("\x1b[0;0H")
	// Clear from cursor to bottom of screen
	logger.Print("\x1b[J")
}
