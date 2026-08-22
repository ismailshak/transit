package store_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/ismailshak/transit/internal/store"
	"github.com/ismailshak/transit/internal/transit"
	"github.com/stretchr/testify/assert"
)

var testLocation transit.LocationSlug = "moon"

var stopsFixture = []transit.Stop{
	{StopID: "A", Name: "AAA", Location: testLocation, AgencyID: "MET", Latitude: "12.1818181", Longitude: "-332.99933", Type: "train", ParentID: ""},
	{StopID: "B", Name: "BBB", Location: testLocation, AgencyID: "MET", Latitude: "12.1813458", Longitude: "-332.99993", Type: "train", ParentID: "A"},
	{StopID: "C", Name: "CCC", Location: testLocation, AgencyID: "MET", Latitude: "12.1814451", Longitude: "-332.99773", Type: "train", ParentID: "B"},
	{StopID: "D", Name: "DDD", Location: testLocation, AgencyID: "MET", Latitude: "12.1812341", Longitude: "-332.98833", Type: "train", ParentID: "C"},
}

var matchFixture = []transit.Stop{
	{StopID: "STN_A07", Name: "Van Ness-UDC", Location: testLocation, AgencyID: "MET", Type: "train", ParentID: ""},
	{StopID: "STN_J02", Name: "Van Dorn Street", Location: testLocation, AgencyID: "MET", Type: "train", ParentID: ""},
	{StopID: "STN_A01", Name: "Metro Center", Location: testLocation, AgencyID: "MET", Type: "train", ParentID: ""},
	{StopID: "STN_C03", Name: "Farragut West", Location: testLocation, AgencyID: "MET", Type: "train", ParentID: ""},
	{StopID: "PF_A01_1", Name: "Metro Center Upper Platform", Location: testLocation, AgencyID: "MET", Type: "train", ParentID: "STN_A01"},
	{StopID: "STN_X01", Name: "Van Dorn Depot", Location: "mars", AgencyID: "MRS", Type: "train", ParentID: ""},
}

func TestMigrationCompletes(t *testing.T) {
	db := blankDB(t)
	err := db.SyncMigrations(t.Context())

	if err != nil {
		t.Errorf("Failed to sync migrations. %s", err)
	}
}

func TestGetValidLocation(t *testing.T) {
	t.Parallel()

	db := migratedDB(t)

	// The dmv row arrives with the migrations. Nothing else can insert a location.
	locationRow, err := db.Location(t.Context(), transit.DMVSlug)

	if err != nil {
		t.Fatalf("Failed to get location from db: %s", err)
	}

	assert.Equal(t, transit.DMVSlug, locationRow.Slug)
	assert.Equal(t, "District Of Columbia, Maryland and Virginia (US)", locationRow.Name)
	assert.True(t, locationRow.SupportsGTFS)
	assert.NotEqual(t, locationRow.CreatedAt, "")
	assert.NotNil(t, locationRow.CreatedAt)
	assert.NotEqual(t, locationRow.UpdatedAt, "")
	assert.NotNil(t, locationRow.UpdatedAt)
}

func TestGetInvalidLocation(t *testing.T) {
	t.Parallel()

	db := migratedDB(t)

	locationRow, err := db.Location(t.Context(), "invalid")

	if err != nil {
		t.Fatalf("Failed to get location from db: %s", err)
	}

	assert.Nil(t, locationRow)
}

func TestGetLocationSeparatesMissingFromFailed(t *testing.T) {
	t.Parallel()

	t.Run("a slug with no row", func(t *testing.T) {
		t.Parallel()

		db := migratedDB(t)

		location, err := db.Location(t.Context(), "nowhere")
		if err != nil {
			t.Fatalf("expected no error but got %v", err)
		}

		if location != nil {
			t.Errorf("expected no location but got %+v", location)
		}
	})

	t.Run("a query that fails", func(t *testing.T) {
		t.Parallel()

		db := migratedDB(t)
		if err := db.Close(); err != nil {
			t.Fatalf("expected no error but got %v", err)
		}

		location, err := db.Location(t.Context(), transit.DMVSlug)
		if err == nil {
			t.Fatal("expected an error but got nil, a closed database reads as an empty one")
		}

		if location != nil {
			t.Errorf("expected no location but got %+v", location)
		}
	})
}

func TestCancelledContextReachesTheDriver(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		call func(ctx context.Context, db *store.Store) error
	}{
		"many rows": {
			call: func(ctx context.Context, db *store.Store) error {
				_, err := db.AllLocations(ctx)
				return err
			},
		},
		"one row": {
			call: func(ctx context.Context, db *store.Store) error {
				_, err := db.Location(ctx, transit.DMVSlug)
				return err
			},
		},
		"transaction": {
			call: func(ctx context.Context, db *store.Store) error {
				return db.InsertStops(ctx, stopsFixture)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			db := migratedDB(t)

			ctx, cancel := context.WithCancel(t.Context())
			cancel()

			if err := tc.call(ctx, db); !errors.Is(err, context.Canceled) {
				t.Errorf("expected context.Canceled but got %v, the ctx never reached the driver", err)
			}
		})
	}
}

func TestGetStopsByLocationExcludesParent(t *testing.T) {
	t.Parallel()

	db := migratedDB(t)

	if err := db.InsertStops(t.Context(), stopsFixture); err != nil {
		t.Fatalf("Failed to insert stop fixture data: %s", err)
	}

	stops, err := db.StopsByLocation(t.Context(), testLocation, true)
	if err != nil {
		t.Fatalf("GetStopsByLocation() returned an error: %s", err)
	}

	if len(stops) != 1 {
		t.Fatalf("Expected 1 stop without parent. Got %d", len(stops))
	}

	assert.Equal(t, stopsFixture[0].StopID, stops[0].StopID)
	assert.Equal(t, stopsFixture[0].Name, stops[0].Name)
	assert.Equal(t, stopsFixture[0].Location, stops[0].Location)
	assert.Equal(t, stopsFixture[0].AgencyID, stops[0].AgencyID)
	assert.Equal(t, stopsFixture[0].Latitude, stops[0].Latitude)
	assert.Equal(t, stopsFixture[0].Longitude, stops[0].Longitude)
	assert.Equal(t, stopsFixture[0].Type, stops[0].Type)
	assert.Equal(t, stopsFixture[0].ParentID, stops[0].ParentID)
	assert.NotEqual(t, stops[0].CreatedAt, "")
	assert.NotNil(t, stops[0].CreatedAt)
	assert.NotEqual(t, stops[0].UpdatedAt, "")
	assert.NotNil(t, stops[0].UpdatedAt)
}

func TestGetStopsByLocationIncludesParent(t *testing.T) {
	t.Parallel()

	db := migratedDB(t)

	if err := db.InsertStops(t.Context(), stopsFixture); err != nil {
		t.Fatalf("Failed to insert fixture data: %s", err)
	}

	stops, err := db.StopsByLocation(t.Context(), testLocation, false)
	if err != nil {
		t.Fatalf("GetStopsByLocation() returned an error: %s", err)
	}

	if len(stops) != 4 {
		t.Fatalf("Expected 4 stop without parent. Got %d", len(stops))
	}

	for i, stop := range stops {
		expected := stopsFixture[i]
		assert.Equal(t, expected.StopID, stop.StopID)
		assert.Equal(t, expected.Name, stop.Name)
		assert.Equal(t, expected.Location, stop.Location)
		assert.Equal(t, expected.AgencyID, stop.AgencyID)
		assert.Equal(t, expected.Latitude, stop.Latitude)
		assert.Equal(t, expected.Longitude, stop.Longitude)
		assert.Equal(t, expected.Type, stop.Type)
		assert.Equal(t, expected.ParentID, stop.ParentID)
		assert.NotEqual(t, stop.CreatedAt, "")
		assert.NotNil(t, stop.CreatedAt)
		assert.NotEqual(t, stop.UpdatedAt, "")
		assert.NotNil(t, stop.UpdatedAt)
	}
}

func TestMatchStops(t *testing.T) {
	t.Parallel()

	db := migratedDB(t)
	if err := db.InsertStops(t.Context(), matchFixture); err != nil {
		t.Fatalf("Failed to insert stop fixture data: %s", err)
	}

	tests := map[string]struct {
		query    string
		expected []string
	}{
		"a station and not the platform under it": {"Metro Center", []string{"Metro Center"}},
		"a word from the middle of a name":        {"west", []string{"Farragut West"}},
		"a query in another case":                 {"VAN DORN", []string{"Van Dorn Street"}},
		"characters in order but not together":    {"frgt", []string{"Farragut West"}},
		"the closest match first":                 {"van d", []string{"Van Dorn Street", "Van Ness-UDC"}},
		"a match but in another location":         {"depot", nil},
		"nothing close enough":                    {"asdfghjkl", nil},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			matched, err := db.MatchStops(t.Context(), testLocation, tc.query)
			if err != nil {
				t.Fatalf("expected no error but got %v", err)
			}

			names := make([]string, 0, len(matched))
			for _, stop := range matched {
				names = append(names, stop.Name)
			}

			if !slices.Equal(tc.expected, names) {
				t.Errorf("expected %v but got %v", tc.expected, names)
			}
		})
	}
}

func TestInsertManyStops(t *testing.T) {
	t.Parallel()

	db := migratedDB(t)
	if err := db.InsertStops(t.Context(), stopsFixture); err != nil {
		t.Fatalf("InsertStops() returned an error: %s", err)
	}

	stopRows, err := db.StopsByLocation(t.Context(), testLocation, false)
	if err != nil {
		t.Fatalf("StopsByLocation() returned an error: %s", err)
	}

	if len(stopsFixture) != len(stopRows) {
		t.Errorf("Expected length %d. Got %d", len(stopsFixture), len(stopRows))
	}

	for i, stop := range stopRows {
		expected := stopsFixture[i]

		assert.Equal(t, i+1, stop.ID)
		assert.Equal(t, expected.StopID, stop.StopID)
		assert.Equal(t, expected.Name, stop.Name)
		assert.Equal(t, expected.Location, stop.Location)
		assert.Equal(t, expected.AgencyID, stop.AgencyID)
		assert.Equal(t, expected.Latitude, stop.Latitude)
		assert.Equal(t, expected.Longitude, stop.Longitude)
		assert.Equal(t, expected.Type, stop.Type)
		assert.Equal(t, expected.ParentID, stop.ParentID)
		assert.NotEqual(t, stop.CreatedAt, "")
		assert.NotNil(t, stop.CreatedAt)
		assert.NotEqual(t, stop.UpdatedAt, "")
		assert.NotNil(t, stop.UpdatedAt)
	}
}
