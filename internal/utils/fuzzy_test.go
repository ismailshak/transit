package utils_test

import (
	"testing"

	"github.com/ismailshak/transit/internal/utils"
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
	testData := []string{
		"name",
		"long name",
		"random word",
		"wEiRd CaSiNg",
		"ALL CAPS",
		"__random*()=++characters^&name",
	}

	testCases := []struct {
		input    string
		expected []string
	}{
		{"name", []string{"name", "long name", "__random*()=++characters^&name"}},
		{"word", []string{"random word"}},
		{"weird CASING", []string{"wEiRd CaSiNg"}},
		{"all caps", []string{"ALL CAPS"}},
	}

	for _, test := range testCases {
		matches := utils.FuzzyFind(test.input, testData)
		assert.Len(t, test.expected, matches.Len())

		result := convertMatchesToSlice(matches)

		assert.ElementsMatch(t, result, test.expected)
	}
}
