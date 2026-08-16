package testutils

import (
	"context"
	"testing"
	"time"

	"github.com/ismailshak/transit/internal/data"
	"github.com/ismailshak/transit/pkg/api"
)

const testLocation data.LocationSlug = "test-location"

type TestAPI struct {
	apiKey  *string
	baseURL string
}

func NewTestAPI(t *testing.T) api.API {
	t.Helper()

	testBaseURL := "http://localhost:3210"
	apiKey := "abcd"

	return &TestAPI{
		apiKey:  &apiKey,
		baseURL: testBaseURL,
	}
}

var allStops = []*data.Stop{
	{StopID: "A", Name: "A Stop", Location: testLocation, Latitude: "34.12301", Longitude: "-11.12301", ParentID: "", Type: data.TrainStation},
	{StopID: "B", Name: "B Stop", Location: testLocation, Latitude: "34.12302", Longitude: "-11.12302", ParentID: "", Type: data.TrainStation},
	{StopID: "C", Name: "C Stop", Location: testLocation, Latitude: "34.12303", Longitude: "-11.12303", ParentID: "", Type: data.TrainStation},
	{StopID: "D", Name: "D Stop", Location: testLocation, Latitude: "34.12304", Longitude: "-11.12304", ParentID: "", Type: data.TrainStation},
	{StopID: "E", Name: "E Stop", Location: testLocation, Latitude: "34.12305", Longitude: "-11.12305", ParentID: "A", Type: data.TrainStation},
	{StopID: "F", Name: "F Stop", Location: testLocation, Latitude: "34.12306", Longitude: "-11.12306", ParentID: "B", Type: data.TrainStation},
	{StopID: "G", Name: "G Stop", Location: testLocation, Latitude: "34.12307", Longitude: "-11.12307", ParentID: "C", Type: data.TrainStation},
}

func (t *TestAPI) FetchStaticData(_ context.Context) (*data.StaticData, error) {
	d := &data.StaticData{
		Stops: allStops,
	}

	return d, nil
}

func (t *TestAPI) FetchPredictions(_ context.Context, input []api.PredictionInput) ([]api.Prediction, error) {
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

func (t *TestAPI) FetchIncidents(_ context.Context) ([]api.Incident, error) {
	i := []api.Incident{
		{Description: "Trains delayed by 3 hours", DateUpdated: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC), Affected: []string{"Outer"}, Type: "Delay"},
		{Description: "All trains are broken", DateUpdated: time.Date(2009, time.December, 11, 23, 0, 0, 0, time.UTC), Affected: []string{"Central"}, Type: "Alert"},
	}

	return i, nil
}

func (t *TestAPI) GetPredictionInput(_ context.Context, arg string) ([]api.PredictionInput, error) {
	matches := data.FuzzyFindFrom(arg, data.SearchableStops(allStops))

	ids := make([]api.PredictionInput, 0, matches.Len())

	for _, m := range matches {
		id := allStops[m.Index].StopID
		agency := allStops[m.Index].AgencyID
		ids = append(ids, api.PredictionInput{StopID: id, AgencyID: agency})
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
