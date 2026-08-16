package store

import (
	"github.com/ismailshak/transit/internal/transit"
	"github.com/sahilm/fuzzy"
)

// SearchableStops is a wrapper type that implements the fuzzy matching interface to enable search
type SearchableStops []*transit.Stop

func (s SearchableStops) Len() int            { return len(s) }
func (s SearchableStops) String(i int) string { return s[i].Name }

// FuzzyFind fuzzy finds a `name` in a slice of strings
func FuzzyFind(name string, data []string) fuzzy.Matches {
	return fuzzy.Find(name, data)
}

// FuzzyFindFrom fuzzy finds a `name` in a slice of any data source that implements `Len()` and `String()`
func FuzzyFindFrom(name string, data fuzzy.Source) fuzzy.Matches {
	return fuzzy.FindFrom(name, data)
}
