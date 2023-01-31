package helpers

import "github.com/sahilm/fuzzy"

func FuzzyFind(name string, data []string) fuzzy.Matches {
	return fuzzy.Find(name, data)
}
