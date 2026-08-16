package data_test

import (
	"testing"

	"github.com/ismailshak/transit/internal/data"
	"github.com/sahilm/fuzzy"
	"github.com/stretchr/testify/assert"
)

func convertMatchesToSlice(matches fuzzy.Matches) []string {
	var found []string
	for _, m := range matches {
		found = append(found, m.Str)
	}

	return found
}

func TestFuzzyFind(t *testing.T) {
	t.Parallel()

	tt := []string{
		"name",
		"long name",
		"random word",
		"wEiRd CaSiNg",
		"ALL CAPS",
		"__random*()=++characters^&name",
	}

	testCases := map[string]struct {
		input    string
		expected []string
	}{
		"an exact and embedded matches": {"name", []string{"name", "long name", "__random*()=++characters^&name"}},
		"an embedded match":             {"word", []string{"random word"}},
		"a mixed case match":            {"weird CASING", []string{"wEiRd CaSiNg"}},
		"an all caps match":             {"all caps", []string{"ALL CAPS"}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			matches := data.FuzzyFind(tc.input, tt)
			assert.Len(t, matches, len(tc.expected))

			result := convertMatchesToSlice(matches)

			assert.ElementsMatch(t, result, tc.expected)
		})
	}
}
