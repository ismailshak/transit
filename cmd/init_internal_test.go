package cmd

import (
	"slices"
	"testing"

	"github.com/ismailshak/transit/internal/data"
	"github.com/ismailshak/transit/internal/ui"
)

func TestToChoices(t *testing.T) {
	tests := map[string]struct {
		input    []data.Location
		expected []ui.Choice
	}{
		"converts a location correctly": {
			input: []data.Location{
				{Slug: "slug", Name: "Long Slug Name"},
			},
			expected: []ui.Choice{
				{Key: "slug", Title: "slug", Description: "Long Slug Name", FilterValue: "Long Slug Name"},
			},
		},
		"empty list": {
			input:    []data.Location{},
			expected: []ui.Choice{},
		},
	}

	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			got := toChoices(testCase.input)

			if !slices.Equal(got, testCase.expected) {
				t.Errorf("expected %v but got %v", testCase.expected, got)
			}
		})
	}
}
