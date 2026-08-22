package provider

import (
	"testing"
	"time"
)

func TestSFTimestamp(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		sec  int64
		want time.Time
	}{
		"absent": {
			sec:  0,
			want: time.Time{},
		},
		"a real instant": {
			sec:  1787354572,
			want: time.Unix(1787354572, 0),
		},
		"before the epoch": {
			sec:  -1,
			want: time.Unix(-1, 0),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := sfTimestamp(tc.sec)

			if !got.Equal(tc.want) {
				t.Errorf("expected %v but got %v", tc.want, got)
			}
		})
	}
}

func TestIsSFGhostTrain(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		line        string
		destination string
		want        bool
	}{
		"a trip with passengers": {
			line:        "Red-N",
			destination: "Richmond",
			want:        false,
		},
		"no line": {
			line:        "--",
			destination: "Richmond",
			want:        true,
		},
		"a trip without passengers": {
			line:        "Red-N",
			destination: "NO PASSENGERS",
			want:        true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var mvj sfMonitoredVehicleJourney
			mvj.LineRef = tc.line
			mvj.MonitoredCall.DestinationDisplay = tc.destination

			if got := isSFGhostTrain(mvj); got != tc.want {
				t.Errorf("expected %v but got %v", tc.want, got)
			}
		})
	}
}

func TestAlertText(t *testing.T) {
	t.Parallel()

	english := sfTranslatedText{Translations: []sfTranslation{
		{Language: "es", Text: "Cambio de andén"},
		{Language: "en", Text: "Platform change"},
	}}
	blank := sfTranslatedText{Translations: []sfTranslation{{Language: "en", Text: ""}}}
	detail := sfTranslatedText{Translations: []sfTranslation{{Language: "en", Text: "Train 519 departs from platform 5"}}}

	tests := map[string]struct {
		header      sfTranslatedText
		description sfTranslatedText
		want        string
	}{
		"both are joined when both are filled": {
			header:      english,
			description: detail,
			want:        "Platform change: Train 519 departs from platform 5",
		},
		"bart pairs a banner header with the real text": {
			header:      sfTranslatedText{Translations: []sfTranslation{{Language: "en", Text: "BART.gov Alert"}}},
			description: sfTranslatedText{Translations: []sfTranslation{{Language: "en", Text: "Expect delays of 30-40 minutes. "}}},
			want:        "BART.gov Alert: Expect delays of 30-40 minutes.",
		},
		"header stands alone when the description is empty": {
			header:      english,
			description: blank,
			want:        "Platform change",
		},
		"description fills in when the header is blank": {
			header:      blank,
			description: detail,
			want:        "Train 519 departs from platform 5",
		},
		"null blocks unmarshal to no translations": {
			header:      sfTranslatedText{},
			description: sfTranslatedText{},
			want:        "",
		},
		"english is not always first": {
			header:      english,
			description: sfTranslatedText{},
			want:        "Platform change",
		},
		"no english translation": {
			header:      sfTranslatedText{Translations: []sfTranslation{{Language: "vi", Text: "Thay đổi sân ga"}}},
			description: sfTranslatedText{},
			want:        "",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := alertText(tc.header, tc.description)

			if got != tc.want {
				t.Errorf("expected %q but got %q", tc.want, got)
			}
		})
	}
}
