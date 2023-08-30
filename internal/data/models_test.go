package data_test

import (
	"testing"

	"github.com/ismailshak/transit/internal/data"
	"github.com/ismailshak/transit/internal/testutils"
	"github.com/stretchr/testify/assert"
)

var fixtureLocation data.LocationSlug = "moon"

var fixture = []*data.Stop{
	{StopID: "A", Name: "AAA", Location: fixtureLocation, Latitude: "12.1818181", Longitude: "-332.99933", Type: "train", ParentID: ""},
	{StopID: "B", Name: "BBB", Location: fixtureLocation, Latitude: "12.1813458", Longitude: "-332.99993", Type: "train", ParentID: "A"},
	{StopID: "C", Name: "CCC", Location: fixtureLocation, Latitude: "12.1814451", Longitude: "-332.99773", Type: "train", ParentID: "B"},
	{StopID: "D", Name: "DDD", Location: fixtureLocation, Latitude: "12.1812341", Longitude: "-332.98833", Type: "train", ParentID: "C"},
}

func TestGetStopsByLocationExcludesParent(t *testing.T) {
	t.Parallel()

	db := testutils.MigratedDB(t)

	for _, f := range fixture {
		_, err := db.Exec(
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

	stops, err := data.GetStopsByLocation(db, fixtureLocation, true)
	if err != nil {
		t.Fatalf("GetStopsByLocation() returned an error: %s", err)
	}

	if len(stops) != 1 {
		t.Fatalf("Expected 1 stop without parent. Got %d", len(stops))
	}

	assert.Equal(t, fixture[0].StopID, stops[0].StopID)
	assert.Equal(t, fixture[0].Location, stops[0].Location)
	assert.Equal(t, fixture[0].Name, stops[0].Name)
	assert.Equal(t, fixture[0].Latitude, stops[0].Latitude)
	assert.Equal(t, fixture[0].Longitude, stops[0].Longitude)
	assert.Equal(t, fixture[0].Type, stops[0].Type)
	assert.Equal(t, fixture[0].ParentID, stops[0].ParentID)
	assert.NotEqual(t, stops[0].CreatedAt, "")
	assert.NotNil(t, stops[0].CreatedAt)
	assert.NotEqual(t, stops[0].UpdatedAt, "")
	assert.NotNil(t, stops[0].UpdatedAt)
}

func TestGetStopsByLocationIncludesParent(t *testing.T) {
	t.Parallel()

	db := testutils.MigratedDB(t)

	for _, f := range fixture {
		_, err := db.Exec(
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

	stops, err := data.GetStopsByLocation(db, fixtureLocation, false)
	if err != nil {
		t.Fatalf("GetStopsByLocation() returned an error: %s", err)
	}

	if len(stops) != 4 {
		t.Fatalf("Expected 4 stop without parent. Got %d", len(stops))
	}

	for i, stop := range stops {
		expected := fixture[i]
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
	if err := data.InsertStops(db, fixture); err != nil {
		t.Fatalf("InsertStops() returned an error: %s", err)
	}

	rows, err := db.Query("SELECT rowid, * FROM stops")
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

	if len(fixture) != len(stopRows) {
		t.Errorf("Expected length %d. Got %d", len(fixture), len(stopRows))
	}

	for i, stop := range stopRows {
		expected := fixture[i]

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
