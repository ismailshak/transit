package testutils

import (
	"context"
	"testing"
	"time"

	"github.com/ismailshak/transit/internal/data"
	"github.com/ismailshak/transit/internal/provider"
	"github.com/ismailshak/transit/internal/transit"
)

const testLocation transit.LocationSlug = "test-location"

type TestAPI struct {
	apiKey  *string
	baseURL string
}

func NewTestAPI(t *testing.T) provider.API {
	t.Helper()

	testBaseURL := "http://localhost:3210"
	apiKey := "abcd"

	return &TestAPI{
		apiKey:  &apiKey,
		baseURL: testBaseURL,
	}
}

var allStops = []*transit.Stop{
	{StopID: "A", Name: "A Stop", Location: testLocation, Latitude: "34.12301", Longitude: "-11.12301", ParentID: "", Type: transit.TrainStation},
	{StopID: "B", Name: "B Stop", Location: testLocation, Latitude: "34.12302", Longitude: "-11.12302", ParentID: "", Type: transit.TrainStation},
	{StopID: "C", Name: "C Stop", Location: testLocation, Latitude: "34.12303", Longitude: "-11.12303", ParentID: "", Type: transit.TrainStation},
	{StopID: "D", Name: "D Stop", Location: testLocation, Latitude: "34.12304", Longitude: "-11.12304", ParentID: "", Type: transit.TrainStation},
	{StopID: "E", Name: "E Stop", Location: testLocation, Latitude: "34.12305", Longitude: "-11.12305", ParentID: "A", Type: transit.TrainStation},
	{StopID: "F", Name: "F Stop", Location: testLocation, Latitude: "34.12306", Longitude: "-11.12306", ParentID: "B", Type: transit.TrainStation},
	{StopID: "G", Name: "G Stop", Location: testLocation, Latitude: "34.12307", Longitude: "-11.12307", ParentID: "C", Type: transit.TrainStation},
}

func (t *TestAPI) FetchStaticData(_ context.Context) (*transit.StaticData, error) {
	d := &transit.StaticData{
		Stops: allStops,
	}

	return d, nil
}

func (t *TestAPI) FetchPredictions(_ context.Context, input []provider.PredictionInput) ([]provider.Prediction, error) {
	p := []provider.Prediction{
		{Min: "1", LocationName: "Stn 1", Destination: "Dest A", DestinationName: "Destination A", Line: "Central"},
		{Min: "3", LocationName: "Stn 1", Destination: "NO PASSENGERS", DestinationName: "Destination A", Line: "Central"},
		{Min: "ARR", LocationName: "Stn 1", Destination: "Dest B", DestinationName: "Destination B", Line: "Outer"},
		{Min: "22", LocationName: "Stn 1", Destination: "Dest B", DestinationName: "Destination B", Line: "Outer"},
		{Min: "2", LocationName: "Stn 1", Destination: "Dest C", DestinationName: "Destination C", Line: "Inner"},
		{Min: "--", LocationName: "Stn 1", Destination: "Dest C", DestinationName: "Destination C", Line: "Inner"},
	}

	return p, nil
}

func (t *TestAPI) FetchIncidents(_ context.Context) ([]provider.Incident, error) {
	i := []provider.Incident{
		{Description: "Trains delayed by 3 hours", DateUpdated: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC), Affected: []string{"Outer"}, Type: "Delay"},
		{Description: "All trains are broken", DateUpdated: time.Date(2009, time.December, 11, 23, 0, 0, 0, time.UTC), Affected: []string{"Central"}, Type: "Alert"},
	}

	return i, nil
}

func (t *TestAPI) GetPredictionInput(_ context.Context, arg string) ([]provider.PredictionInput, error) {
	matches := data.FuzzyFindFrom(arg, data.SearchableStops(allStops))

	ids := make([]provider.PredictionInput, 0, matches.Len())

	for _, m := range matches {
		id := allStops[m.Index].StopID
		agency := allStops[m.Index].AgencyID
		ids = append(ids, provider.PredictionInput{StopID: id, AgencyID: agency})
	}

	return ids, nil
}

func (t *TestAPI) GetLineColor(stop string) (string, string) {
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

func (t *TestAPI) IsGhostTrain(line, destination string) bool {
	return line == "--" || destination == "NO PASSENGERS"
}
