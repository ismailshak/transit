package data

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ismailshak/transit/helpers"
)

// Unzips file (which holds the content of a GTFS Static feed)
// located at `path` into a destination provided by `dest`.
//
// Destination is assumed to already exist. Zip file will not be deleted
// after it's unzipped
func UnzipStaticGTFS(path string, dest string) error {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return err
	}

	defer reader.Close()

	err = helpers.Unzip(reader, dest)
	if err != nil {
		return fmt.Errorf("failed to unzip gtfs static: %s", err)
	}

	return nil
}

// Parses an unzipped directory that contains the GTFS Static feed
func ParseGTFS(path string, st StopType) (*Data, error) {
	stopsFile := filepath.Join(path, "stops.txt")

	stops, err := parseStops(stopsFile, st)
	if err != nil {
		return nil, err
	}

	gtfs := &Data{
		Stops: stops,
	}

	return gtfs, nil
}

func parseStops(path string, st StopType) ([]*Stop, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	r := csv.NewReader(f)

	// Read the header column separately
	header, err := r.Read()
	if err != nil {
		return nil, err
	}

	// Create a map from column name to it's index
	fields := make(map[string]int, len(header))
	for i, name := range header {
		fields[name] = i
	}

	stops := make([]*Stop, 0, 64) // Random safe-bet high number to avoid excessive reallocations

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		stop := &Stop{
			Location:  "dmv",
			Type:      st,
			StopID:    record[fields["stop_id"]],
			Name:      record[fields["stop_name"]],
			Latitude:  record[fields["stop_lat"]],
			Longitude: record[fields["stop_lon"]],
			ParentID:  record[fields["parent_station"]],
		}

		stops = append(stops, stop)
	}

	return stops, nil
}
