package data_test

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/ismailshak/transit/internal/data"
	"github.com/stretchr/testify/assert"
)

func TestUnzipGTFS(t *testing.T) {
	t.Parallel()

	cwd, _ := os.Getwd() // Resolves to this file's directory
	pathToZip := filepath.Join(cwd, "testdata", "sample-feed.zip")
	dest := t.TempDir()

	err := data.UnzipStaticGTFS(pathToZip, dest)
	if err != nil {
		t.Fatalf("UnzipStaticGTFS() returned an error: %s", err)
	}

	filepath.WalkDir(dest, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			t.Errorf("Error reading file in unzipped destination: %s", err)
			return err
		}

		if d.IsDir() && d.Name() == "001" {
			return nil
		}

		fileName := d.Name()
		fixturePath := filepath.Join(cwd, "testdata", "sample-feed", fileName)
		fixtureContent, err := os.ReadFile(fixturePath)
		if err != nil {
			t.Errorf("Error reading fixture file content: %s", err)
			return nil
		}

		unzippedContent, err := os.ReadFile(filepath.Join(dest, fileName))
		if err != nil {
			t.Errorf("Error reading unzipped file content: %s", err)
			return nil
		}

		if !bytes.Equal(removeCarriageReturn(fixtureContent), removeCarriageReturn(unzippedContent)) {
			t.Errorf("Fixture and unzipped don't have the same content")
			t.Logf("Fixture content: %s\n", string(fixtureContent))
			t.Logf("Unzipped content: %s\n", string(unzippedContent))
		}

		return nil
	})

}

func TestParseGTFS(t *testing.T) {
	t.Parallel()

	cwd, _ := os.Getwd() // Resolves to this file's directory
	pathToFeed := filepath.Join(cwd, "testdata", "sample-feed")

	gtfs, err := data.ParseGTFS(pathToFeed, "train")
	if err != nil {
		t.Fatalf("ParseGTFS() returned an error: %s", err)
	}

	if len(gtfs.Stops) != 9 {
		t.Errorf("Expected 9 stops. Got %d", len(gtfs.Stops))
	}

	expectedStops := []struct {
		StopId    string
		Name      string
		Latitude  string
		Longitude string
		Type      data.StopType
		ParentID  string
	}{
		{"FUR_CREEK_RES", "Furnace Creek Resort (Demo)", "36.425288", "-117.133162", "train", ""},
		{"BEATTY_AIRPORT", "Nye County Airport (Demo)", "36.868446", "-116.784582", "train", ""},
		{"BULLFROG", "Bullfrog (Demo)", "36.88108", "-116.81797", "train", ""},
		{"STAGECOACH", "Stagecoach Hotel & Casino (Demo)", "36.915682", "-116.751677", "train", ""},
		{"NADAV", "North Ave / D Ave N (Demo)", "36.914893", "-116.76821", "train", ""},
		{"NANAA", "North Ave / N A Ave (Demo)", "36.914944", "-116.761472", "train", ""},
		{"DADAN", "Doing Ave / D Ave N (Demo)", "36.909489", "-116.768242", "train", ""},
		{"EMSI", "E Main St / S Irving St (Demo)", "36.905697", "-116.76218", "train", ""},
		{"AMV", "Amargosa Valley (Demo)", "36.641496", "-116.40094", "train", ""},
	}

	for i, stop := range gtfs.Stops {
		expected := expectedStops[i]

		assert.Equal(t, expected.StopId, stop.StopID)
		assert.Equal(t, expected.Name, stop.Name)
		assert.Equal(t, expected.Latitude, stop.Latitude)
		assert.Equal(t, expected.Longitude, stop.Longitude)
	}
}

// Filters out `\r` to make testing on Windows easier
func removeCarriageReturn(s []byte) []byte {
	filtered := make([]byte, 0, len(s))

	for _, b := range s {
		if b != 13 {
			filtered = append(filtered, b)
		}
	}

	return filtered
}
