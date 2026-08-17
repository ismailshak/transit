package store

import (
	"github.com/ismailshak/transit/internal/transit"
	"github.com/sahilm/fuzzy"
)

// searchableStops is a wrapper type that implements the fuzzy matching interface to enable search
type searchableStops []*transit.Stop

func (s searchableStops) Len() int            { return len(s) }
func (s searchableStops) String(i int) string { return s[i].Name }

// fuzzyFindFrom fuzzy finds a `name` in a slice of any data source that implements `Len()` and `String()`
func fuzzyFindFrom(name string, data fuzzy.Source) fuzzy.Matches {
	return fuzzy.FindFrom(name, data)
}
