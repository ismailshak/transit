package gtfs

import "testing"

func TestValueOrFallback(t *testing.T) {
	t.Parallel()

	t.Run("string", func(t *testing.T) {
		t.Parallel()

		tests := map[string]struct {
			value    string
			fallback string
			exists   bool
			want     string
		}{
			"column exists, value used": {
				value:    "downtown",
				fallback: "",
				exists:   true,
				want:     "downtown",
			},
			"column exists but value is empty": {
				value:    "",
				fallback: "default",
				exists:   true,
				want:     "",
			},
			"column missing, fallback used": {
				value:    "",
				fallback: "default",
				exists:   false,
				want:     "default",
			},
		}

		for name, tc := range tests {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				got := valueOrFallback(tc.value, tc.fallback, tc.exists)
				if got != tc.want {
					t.Errorf("expected %q but got %q", tc.want, got)
				}
			})
		}
	})

	t.Run("float64", func(t *testing.T) {
		t.Parallel()

		tests := map[string]struct {
			value    float64
			fallback float64
			exists   bool
			want     float64
		}{
			"column exists, value used": {
				value:    38.8977,
				fallback: 0,
				exists:   true,
				want:     38.8977,
			},
			"column missing, fallback used": {
				value:    0,
				fallback: -1,
				exists:   false,
				want:     -1,
			},
		}

		for name, tc := range tests {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				got := valueOrFallback(tc.value, tc.fallback, tc.exists)
				if got != tc.want {
					t.Errorf("expected %v but got %v", tc.want, got)
				}
			})
		}
	})
}
