package tui

import "fmt"

// TerminalBuffer is an alternate buffer drawn on top of the active terminal screen.
// This buffer will not erase the user's current terminal, but render itself
// on top of the existing buffer.
type TerminalBuffer struct {
	alternateBufferActive bool
}

// NewBuffer returns an alternate buffer instance
func NewBuffer() *TerminalBuffer {
	return &TerminalBuffer{}
}

// StartAlternateBuffer renders a buffer on top of the current one
func (b *TerminalBuffer) StartAlternateBuffer() {
	if !b.alternateBufferActive {
		fmt.Println("\x1b[?1049h")
		b.alternateBufferActive = true
	}
}

// StopAlternateBuffer closes the alternate buffer and returns to the original buffer
func (b *TerminalBuffer) StopAlternateBuffer() {
	if b.alternateBufferActive {
		fmt.Println("\x1b[?1049l")
		b.alternateBufferActive = false
	}
}

// RefreshScreen erases the entire content of the buffer and leaves
// cursor in the in the first row and column of the buffer
func (b *TerminalBuffer) RefreshScreen() {
	// Move cursor to 0,0
	fmt.Println("\x1b[0;0H")
	// Clear from cursor to bottom of screen
	fmt.Println("\x1b[J")
}
