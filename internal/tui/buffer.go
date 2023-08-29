package tui

import (
	"github.com/ismailshak/transit/internal/logger"
)

// An alternate buffer drawn on top of active terminal screen
type Buffer struct {
	alternateBufferEnabled bool
	alternateBufferActive  bool
}

func NewBuffer() *Buffer {
	return &Buffer{}
}

func (b *Buffer) StartAlternateBuffer() {
	if !b.alternateBufferActive {
		logger.Print("\x1b[?1049h")
		b.alternateBufferActive = true
	}
}

func (b *Buffer) StopAlternateBuffer() {
	if b.alternateBufferActive {
		logger.Print("\x1b[?1049l")
		b.alternateBufferActive = false
	}
}

func (b *Buffer) RefreshScreen() {
	// Move cursor to 0,0
	logger.Print("\x1b[0;0H")
	// Clear from cursor to bottom of screen
	logger.Print("\x1b[J")
}
