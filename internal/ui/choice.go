package ui

import (
	"errors"
)

// ErrCancelled is returned when the user backs out of a prompt (Esc, Ctrl+C)
// rather than choosing something.
var ErrCancelled = errors.New("cancelled")

// ErrNoSelection is returned when the user confirms with nothing matched,
// e.g. hitting enter on a filter with zero results. Distinct from
// ErrCancelled so callers can tell "backed out" from "confirmed nothing".
var ErrNoSelection = errors.New("nothing selected")

type Choice struct {
	// Key is the unique identifier for the item and will be the value returned when the user selects an item
	Key string
	// Title is the primary display value for the item
	Title string
	// Description is the secondary display value for the item that adds more context
	Description string
	// FilterValue is the value that will be used for fuzzy matching when filtering the list
	FilterValue string
}

// choiceItem adapts a Choice to bubbles/list.Item (and list.DefaultItem) so
// the library's requirements stay out of Choice itself. Unexported: nothing
// outside this package needs to know it exists.
type choiceItem struct {
	Choice
}

func (c choiceItem) Key() string         { return c.Choice.Key }
func (c choiceItem) Title() string       { return c.Choice.Title }
func (c choiceItem) Description() string { return c.Choice.Description }
func (c choiceItem) FilterValue() string { return c.Choice.FilterValue }
