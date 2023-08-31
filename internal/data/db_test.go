package data_test

import (
	"testing"

	"github.com/ismailshak/transit/internal/data"
	"github.com/ismailshak/transit/internal/testutils"
	"github.com/stretchr/testify/assert"
)

var testLocation data.LocationSlug = "moon"

var locationFixture = &data.Location{Slug: "x", Name: "XYZ", SupportsGTFS: true}

var stopsFixture = []*data.Stop{
	{StopID: "A", Name: "AAA", Location: testLocation, Latitude: "12.1818181", Longitude: "-332.99933", Type: "train", ParentID: ""},
	{StopID: "B", Name: "BBB", Location: testLocation, Latitude: "12.1813458", Longitude: "-332.99993", Type: "train", ParentID: "A"},
	{StopID: "C", Name: "CCC", Location: testLocation, Latitude: "12.1814451", Longitude: "-332.99773", Type: "train", ParentID: "B"},
	{StopID: "D", Name: "DDD", Location: testLocation, Latitude: "12.1812341", Longitude: "-332.98833", Type: "train", ParentID: "C"},
}

func TestMigrationCompletes(t *testing.T) {
	db := testutils.BlankDB(t)
	err := db.SyncMigrations()

	if err != nil {
		t.Errorf("Failed to sync migrations. %s", err)
	}
}

func TestGetValidLocation(t *testing.T) {
	t.Parallel()

	db := testutils.MigratedDB(t)

	_, err := db.DB.Exec(
		"INSERT INTO locations (slug, name, supports_gtfs) VALUES (?, ?, ?)",
		locationFixture.Slug,
		locationFixture.Name,
		locationFixture.SupportsGTFS,
	)

	if err != nil {
		t.Fatalf("Failed to insert location fixture data: %s", err)
	}

	locationRow, err := db.GetLocation(locationFixture.Slug)

	if err != nil {
		t.Fatalf("Failed to get location from db: %s", err)
	}

	assert.Equal(t, locationFixture.Slug, locationRow.Slug)
	assert.Equal(t, locationFixture.Name, locationRow.Name)
	assert.Equal(t, locationFixture.SupportsGTFS, locationRow.SupportsGTFS)
	assert.NotEqual(t, locationRow.CreatedAt, "")
	assert.NotNil(t, locationRow.CreatedAt)
	assert.NotEqual(t, locationRow.UpdatedAt, "")
	assert.NotNil(t, locationRow.UpdatedAt)
}

func TestGetInvalidLocation(t *testing.T) {
	t.Parallel()

	db := testutils.MigratedDB(t)

	_, err := db.DB.Exec(
		"INSERT INTO locations (slug, name, supports_gtfs) VALUES (?, ?, ?)",
		locationFixture.Slug,
		locationFixture.Name,
		locationFixture.SupportsGTFS,
	)

	if err != nil {
		t.Fatalf("Failed to insert location fixture data: %s", err)
	}

	locationRow, err := db.GetLocation("invalid")

	if err != nil {
		t.Fatalf("Failed to get location from db: %s", err)
	}

	assert.Nil(t, locationRow)
}

func TestGetStopsByLocationExcludesParent(t *testing.T) {
	t.Parallel()

	db := testutils.MigratedDB(t)

	for _, f := range stopsFixture {
		_, err := db.DB.Exec(
			"INSERT INTO stops (stop_id, name, location, latitude, longitude, type, parent_id) VALUES (?, ?, ?, ?, ?, ?, ?)",
			f.StopID,
			f.Name,
			f.Location,
			f.Latitude,
			f.Longitude,
			f.Type,
			f.ParentID,
		)

		if err != nil {
			t.Fatalf("Failed to insert stop fixture data: %s", err)
		}
	}

	stops, err := db.GetStopsByLocation(testLocation, true)
	if err != nil {
		t.Fatalf("GetStopsByLocation() returned an error: %s", err)
	}

	if len(stops) != 1 {
		t.Fatalf("Expected 1 stop without parent. Got %d", len(stops))
	}

	assert.Equal(t, stopsFixture[0].StopID, stops[0].StopID)
	assert.Equal(t, stopsFixture[0].Location, stops[0].Location)
	assert.Equal(t, stopsFixture[0].Name, stops[0].Name)
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

	db := testutils.MigratedDB(t)

	for _, f := range stopsFixture {
		_, err := db.DB.Exec(
			"INSERT INTO stops (stop_id, name, location, latitude, longitude, type, parent_id) VALUES (?, ?, ?, ?, ?, ?, ?)",
			f.StopID,
			f.Name,
			f.Location,
			f.Latitude,
			f.Longitude,
			f.Type,
			f.ParentID,
		)

		if err != nil {
			t.Fatalf("Failed to insert fixture data: %s", err)
		}
	}

	stops, err := db.GetStopsByLocation(testLocation, false)
	if err != nil {
		t.Fatalf("GetStopsByLocation() returned an error: %s", err)
	}

	if len(stops) != 4 {
		t.Fatalf("Expected 4 stop without parent. Got %d", len(stops))
	}

	for i, stop := range stops {
		expected := stopsFixture[i]
		assert.Equal(t, expected.StopID, stop.StopID)
		assert.Equal(t, expected.Location, stop.Location)
		assert.Equal(t, expected.Name, stop.Name)
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

func TestInsertManyStops(t *testing.T) {
	t.Parallel()

	db := testutils.MigratedDB(t)
	if err := db.InsertStops(stopsFixture); err != nil {
		t.Fatalf("InsertStops() returned an error: %s", err)
	}

	rows, err := db.DB.Query("SELECT rowid, * FROM stops")
	if err != nil {
		t.Fatalf("SELECT returned an error: %s", err)
	}

	defer rows.Close()

	stopRows := make([]*data.Stop, 0, 4)

	for rows.Next() {
		var row data.Stop
		err = rows.Scan(
			&row.ID,
			&row.StopID,
			&row.Name,
			&row.Location,
			&row.Latitude,
			&row.Longitude,
			&row.Type,
			&row.ParentID,
			&row.CreatedAt,
			&row.UpdatedAt,
		)

		if err != nil {
			t.Errorf("Failed to scan stop row. %s", err)
		}

		stopRows = append(stopRows, &row)
	}

	if len(stopsFixture) != len(stopRows) {
		t.Errorf("Expected length %d. Got %d", len(stopsFixture), len(stopRows))
	}

	for i, stop := range stopRows {
		expected := stopsFixture[i]

		assert.Equal(t, i+1, stop.ID)
		assert.Equal(t, expected.StopID, stop.StopID)
		assert.Equal(t, expected.Location, stop.Location)
		assert.Equal(t, expected.Name, stop.Name)
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
