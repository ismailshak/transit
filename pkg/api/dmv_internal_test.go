package api

import (
	"slices"
	"testing"
)

func TestParseLinesAffected(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		lines string
		want  []string
	}{
		"single line affected": {
			lines: "RD;",
			want:  []string{"RD"},
		},
		"multiple lines affected": {
			lines: "RD; YL; GR; OR; SV; BL;",
			want:  []string{"RD", "YL", "GR", "OR", "SV", "BL"},
		},
		"no lines": {
			lines: "",
			want:  []string{},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := parseLinesAffected(tc.lines)

			if !slices.Equal(got, tc.want) {
				t.Errorf("expected %v but got %v", tc.want, got)
			}
		})
	}
}

func TestFormatDmvStopId(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		id   string
		want []string
	}{
		"no id": {
			id:   "",
			want: []string{},
		},
		"id with single platform": {
			id:   "STN_K01",
			want: []string{"K01"},
		},
		"id with multiple platforms": {
			id:   "STN_A01_C01",
			want: []string{"A01", "C01"},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := formatDmvStopId(tc.id)
			if !slices.Equal(got, tc.want) {
				t.Errorf("expected %v but got %v", tc.want, got)
			}
		})
	}
}
