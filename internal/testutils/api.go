package testutils

import (
	"testing"
	"time"

	"github.com/ismailshak/transit/internal/data"
	"github.com/ismailshak/transit/internal/utils"
	"github.com/ismailshak/transit/pkg/api"
)

const TEST_LOCATION data.LocationSlug = "test-location"

type TestApi struct {
	apiKey  *string
	baseUrl string
}

func NewTestApi(t *testing.T) api.Api {
	t.Helper()

	testBaseUrl := "http://localhost:3210"
	apiKey := "abcd"

	return &TestApi{
		apiKey:  &apiKey,
		baseUrl: testBaseUrl,
	}
}

var ALL_STOPS = []*data.Stop{
	{StopID: "A", Name: "A Stop", Location: TEST_LOCATION, Latitude: "34.12301", Longitude: "-11.12301", ParentID: "", Type: data.TrainStation},
	{StopID: "B", Name: "B Stop", Location: TEST_LOCATION, Latitude: "34.12302", Longitude: "-11.12302", ParentID: "", Type: data.TrainStation},
	{StopID: "C", Name: "C Stop", Location: TEST_LOCATION, Latitude: "34.12303", Longitude: "-11.12303", ParentID: "", Type: data.TrainStation},
	{StopID: "D", Name: "D Stop", Location: TEST_LOCATION, Latitude: "34.12304", Longitude: "-11.12304", ParentID: "", Type: data.TrainStation},
	{StopID: "E", Name: "E Stop", Location: TEST_LOCATION, Latitude: "34.12305", Longitude: "-11.12305", ParentID: "A", Type: data.TrainStation},
	{StopID: "F", Name: "F Stop", Location: TEST_LOCATION, Latitude: "34.12306", Longitude: "-11.12306", ParentID: "B", Type: data.TrainStation},
	{StopID: "G", Name: "G Stop", Location: TEST_LOCATION, Latitude: "34.12307", Longitude: "-11.12307", ParentID: "C", Type: data.TrainStation},
}

func (t *TestApi) FetchStaticData() (*data.StaticData, error) {
	d := &data.StaticData{
		Stops: ALL_STOPS,
	}

	return d, nil
}

func (t *TestApi) FetchPredictions(ids []string) ([]api.Prediction, error) {
	p := []api.Prediction{
		{Min: "1", LocationName: "Stn 1", Destination: "Dest A", DestinationName: "Destination A", Line: "Central"},
		{Min: "3", LocationName: "Stn 1", Destination: "NO PASSENGERS", DestinationName: "Destination A", Line: "Central"},
		{Min: "ARR", LocationName: "Stn 1", Destination: "Dest B", DestinationName: "Destination B", Line: "Outer"},
		{Min: "22", LocationName: "Stn 1", Destination: "Dest B", DestinationName: "Destination B", Line: "Outer"},
		{Min: "2", LocationName: "Stn 1", Destination: "Dest C", DestinationName: "Destination C", Line: "Inner"},
		{Min: "--", LocationName: "Stn 1", Destination: "Dest C", DestinationName: "Destination C", Line: "Inner"},
	}

	return p, nil
}

func (t *TestApi) FetchIncidents() ([]api.Incident, error) {
	i := []api.Incident{
		{Description: "Trains delayed by 3 hours", DateUpdated: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC), Affected: []string{"Outer"}, Type: "Delay"},
		{Description: "All trains are broken", DateUpdated: time.Date(2009, time.December, 11, 23, 0, 0, 0, time.UTC), Affected: []string{"Central"}, Type: "Alert"},
	}

	return i, nil
}

func (t *TestApi) GetIDFromArg(arg string) ([]string, error) {
	matches := utils.FuzzyFindFrom(arg, data.SearchableStops(ALL_STOPS))

	ids := make([]string, 0, matches.Len())

	for _, m := range matches {
		id := ALL_STOPS[m.Index].StopID
		ids = append(ids, id)
	}

	return ids, nil
}

func (t *TestApi) GetLineColor(stop string) (string, string) {
	white, black := "#FFFFFF", "#000000"
	switch stop {
	case "Central":
		return "#FF0000", black
	case "Outer":
		return "#00FF00", black
	case "Inner":
		return "#0000FF", "#FFFFFF"
	default:
		return white, black
	}
}

func (t *TestApi) IsGhostTrain(line, destination string) bool {
	return line == "--" || destination == "NO PASSENGERS"
}
