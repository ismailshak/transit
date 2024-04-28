package data

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ismailshak/transit/internal/utils"
)

// Resolves the GTFS Service Alert "Effect" field to a human-readable string.
// An invalid effect will return an empty string
func ResolveGTFSAlertEffect(effect int) string {
	switch effect {
	case 1:
		return "No Service"
	case 2:
		return "Reduced Service"
	case 3:
		return "Significant Delays"
	case 4:
		return "Detour"
	case 5:
		return "Additional Service"
	case 6:
		return "Modified Service"
	case 7:
		return "Other Effect"
	case 8:
		return "Unknown Effect"
	case 9:
		return "No Effect"
	case 10:
		return "Accessibility Issue"
	default:
		return "Notice"
	}
}

// Unzips file (which holds the content of a GTFS Static feed)
// located at `path` into a destination provided by `dest`.
// Destination is assumed to already exist. Zip file will not be deleted
// after it's unzipped.
//
// The GTFS Static reference can be found here: https://gtfs.org/schedule/reference/
func UnzipStaticGTFS(path string, dest string) error {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("failed to open zip reader: %s", err)
	}

	defer reader.Close()

	err = utils.Unzip(reader, dest)
	if err != nil {
		return fmt.Errorf("failed to unzip gtfs static: %s", err)
	}

	return nil
}

// Parses an unzipped directory that contains the GTFS Static feed
func ParseGTFS(path string, location LocationSlug, st StopType, agency string) (*StaticData, error) {
	agencyFile := filepath.Join(path, "agency.txt")
	stopsFile := filepath.Join(path, "stops.txt")

	agencies, err := parseGTFSAgency(agencyFile, location)
	if err != nil {
		return nil, fmt.Errorf("failed to parse agency.txt: %s", err)
	}

	stops, err := parseGTFSStops(stopsFile, location, st, agency)
	if err != nil {
		return nil, fmt.Errorf("failed to parse stops.txt: %s", err)
	}

	static := &StaticData{
		Agencies: agencies,
		Stops:    stops,
	}

	return static, nil
}

func parseGTFSAgency(path string, location LocationSlug) ([]*Agency, error) {
	agencies := make([]*Agency, 0)
	err := parseGTFSEntity(path, func(record []string, headerMap map[string]int) {
		lang, hasLang := headerMap["agency_lang"]
		agency := &Agency{
			Location: location,
			AgencyID: record[headerMap["agency_id"]],
			Name:     record[headerMap["agency_name"]],
			Timezone: record[headerMap["agency_timezone"]],
			Language: valueOrFallback(record[lang], "", hasLang),
		}

		agencies = append(agencies, agency)
	})

	if err != nil {
		return nil, err
	}

	return agencies, nil
}

func parseGTFSStops(path string, location LocationSlug, st StopType, agency string) ([]*Stop, error) {
	stops := make([]*Stop, 0, 64) // Random safe-bet high number to avoid excessive reallocations
	err := parseGTFSEntity(path, func(record []string, headerMap map[string]int) {
		lat, hasLat := headerMap["stop_lat"]
		lon, hasLon := headerMap["stop_lon"]
		stop := &Stop{
			Location:  location,
			Type:      st,
			AgencyID:  agency,
			StopID:    record[headerMap["stop_id"]],
			Name:      record[headerMap["stop_name"]],
			Latitude:  valueOrFallback(record[lat], "", hasLat),
			Longitude: valueOrFallback(record[lon], "", hasLon),
			ParentID:  record[headerMap["parent_station"]],
		}

		stops = append(stops, stop)
	})

	if err != nil {
		return nil, err
	}

	return stops, nil
}

// Callback func that takes the current parsed row and the header-to-index map as arguments
type ParseEntityFunc func(record []string, headerMap map[string]int)

// Generic GTFS file parser that takes a callback that can handle it's own data via a closure
func parseGTFSEntity(path string, fn ParseEntityFunc) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	defer f.Close()

	r := csv.NewReader(f)
	r.LazyQuotes = true // Fields are often quoted, without this it breaks

	// Read the header column separately
	header, err := r.Read()
	if err != nil {
		return err
	}

	// Create a map from column name to it's index
	headerMap := make(map[string]int, len(header))
	for i, name := range header {
		headerMap[name] = i
	}

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		fn(record, headerMap)
	}

	return nil
}

// Convenience function to help fallback to a value if the GTFS
// column was not found in file headers.
//
// If a column is not in the headersMap, it will return 0. This is
// undesirable because a real column will exist at index 0 of the record.
func valueOrFallback[T any](value, fallback T, exists bool) T {
	if exists {
		return value
	}

	return fallback
}
