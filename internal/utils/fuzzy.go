package utils

import "github.com/sahilm/fuzzy"

// Fuzzy find a `name` in a slice of strings
func FuzzyFind(name string, data []string) fuzzy.Matches {
	return fuzzy.Find(name, data)
}

// Fuzzy find `name` in a slice of any data source that implements `Len()` and `String()`
func FuzzyFindFrom(name string, data fuzzy.Source) fuzzy.Matches {
	return fuzzy.FindFrom(name, data)
}
