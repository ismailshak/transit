package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ismailshak/transit/internal/data"
	"github.com/ismailshak/transit/internal/logger"
	"github.com/ismailshak/transit/internal/utils"
)

// API to interact with San Francisco's 511 API
type SFApi struct {
	apiKey  *string
	baseUrl string
}

type SF_StopPlace struct {
	ID       string `json:"@id"`
	Name     string `json:"Name"`
	Centroid struct {
		Location struct {
			Latitude  string
			Longitude string
		}
	}
	TrasnsportMode string `json:"TransportMode"`
}

type SF_StopPlacesResponse struct {
	Siri struct {
		ServiceDelivery struct {
			ResponseTimestamp  string `json:"ResponseTimestamp"`
			DataObjectDelivery struct {
				ResponseTimestamp string `json:"ResponseTimestamp"`
				DataObjects       struct {
					SiteFrame struct {
						StopPlaces struct {
							StopPlace []SF_StopPlace
						} `json:"stopPlaces"`
					}
				} `json:"dataObjects"`
			}
		}
	}
}

type SF_MonitoredVehicleJourney struct {
	LineRef         string
	DestinationRef  string
	DestinationName string
	MonitoredCall   struct {
		AimedArrivalTime    string
		DestinationDisplay  string
		ExpectedArrivalTime string
		StopPointName       string
	}
}

type SF_StopMonitoringResponse struct {
	ServiceDelivery struct {
		StopMonitoringDelivery struct {
			MonitoredStopVisit []struct {
				MonitoredVehicleJourney SF_MonitoredVehicleJourney
			}
		}
	}
}

func (sf *SFApi) BuildRequest(method string, route ...string) (*http.Request, error) {
	parts := make([]string, 0, len(route)+1)
	parts = append(parts, sf.baseUrl)
	parts = append(parts, route...)
	url := strings.Join(parts, "/")

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (sf *SFApi) fetchStopsForAgency(agency *data.Agency) ([]*data.Stop, error) {
	req, err := sf.BuildRequest(http.MethodGet, "transit", "stopplaces")
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("api_key", *sf.apiKey)
	q.Add("operator_id", agency.AgencyID)
	q.Add("format", "json")
	req.URL.RawQuery = q.Encode()

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("received %d", resp.StatusCode)
	}

	// Remove BOM from response
	body = bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))

	var stopPlaces SF_StopPlacesResponse
	err = json.Unmarshal(body, &stopPlaces)

	if err != nil {
		return nil, err
	}

	var stops []*data.Stop

	for _, sp := range stopPlaces.Siri.ServiceDelivery.DataObjectDelivery.DataObjects.SiteFrame.StopPlaces.StopPlace {
		var stopType data.StopType
		if sp.TrasnsportMode == "bus" {
			stopType = data.BusStop
		} else if sp.TrasnsportMode == "rail" { // CT train type
			stopType = data.TrainStation
		} else if sp.TrasnsportMode == "intercityRail" { // BART train type
			stopType = data.TrainStation
		} else {
			continue
		}

		stop := data.Stop{
			AgencyID:  agency.AgencyID,
			Latitude:  sp.Centroid.Location.Latitude,
			Location:  data.SFSlug,
			Longitude: sp.Centroid.Location.Longitude,
			Name:      sp.Name,
			StopID:    sp.ID,
			Type:      stopType,
		}

		stops = append(stops, &stop)
	}

	return stops, nil
}

func (sf *SFApi) FetchStaticData() (*data.StaticData, error) {
	bart := &data.Agency{
		AgencyID: "BA",
		Language: "en",
		Location: data.SFSlug,
		Name:     "Bay Area Rapid Transit",
		Timezone: "America/Los_Angeles",
	}

	bartStops, err := sf.fetchStopsForAgency(bart)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch BART stops: %s", err)
	}

	cal := &data.Agency{
		AgencyID: "CT",
		Language: "en",
		Location: data.SFSlug,
		Name:     "Caltrain",
		Timezone: "America/Los_Angeles",
	}

	calStops, err := sf.fetchStopsForAgency(cal)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Caltrain stops: %s", err)
	}

	var stops []*data.Stop

	stops = append(bartStops, calStops...)

	staticData := data.StaticData{
		Agencies: []*data.Agency{bart, cal},
		Stops:    stops,
	}

	return &staticData, nil
}

// Removes the `-X` suffix from the line name where X is a direction (e.g. -N, -S, -E, -W)
// and abbreviates the line name
func (sf *SFApi) formatLine(line string) string {
	trimmed := strings.Split(line, "-")[0]
	switch trimmed {
	case "Yellow":
		return "YL"
	case "Red":
		return "RD"
	case "Orange":
		return "OR"
	case "Green":
		return "GR"
	case "Blue":
		return "BL"
	default:
		return trimmed
	}
}

func (sf *SFApi) fetchPrediction(in PredictionInput) ([]Prediction, error) {
	req, err := sf.BuildRequest(http.MethodGet, "transit", "StopMonitoring")
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("api_key", *sf.apiKey)
	q.Add("agency", in.AgencyID)
	q.Add("stopcode", in.StopID)
	q.Add("format", "json")
	req.URL.RawQuery = q.Encode()

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("received %d", resp.StatusCode)
	}

	// Remove BOM from response
	body = bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))

	var stopMonitoring SF_StopMonitoringResponse

	err = json.Unmarshal(body, &stopMonitoring)
	if err != nil {
		return nil, err
	}

	monitoredStopVisits := len(stopMonitoring.ServiceDelivery.StopMonitoringDelivery.MonitoredStopVisit)
	if monitoredStopVisits == 0 {
		return nil, fmt.Errorf("no data returned")
	}

	predictions := make([]Prediction, 0, monitoredStopVisits)

	for _, msv := range stopMonitoring.ServiceDelivery.StopMonitoringDelivery.MonitoredStopVisit {
		mvj := msv.MonitoredVehicleJourney
		var arrival_string string
		// Caltain API does not return ExpectedArrivalTime, it's set to null
		if mvj.MonitoredCall.ExpectedArrivalTime != "" {
			arrival_string = mvj.MonitoredCall.ExpectedArrivalTime
		} else {
			arrival_string = mvj.MonitoredCall.AimedArrivalTime
		}

		arrival_time, err := time.Parse(time.RFC3339, arrival_string)
		if err != nil {
			return nil, err
		}

		now := time.Now()
		arrival := strconv.Itoa(int(arrival_time.Sub(now).Minutes()))

		p := Prediction{
			LocationName:    mvj.MonitoredCall.StopPointName,
			Destination:     mvj.MonitoredCall.DestinationDisplay,
			DestinationName: mvj.DestinationName,
			Line:            sf.formatLine(mvj.LineRef),
			Min:             arrival,
		}

		predictions = append(predictions, p)
	}

	return predictions, nil
}

func (sf *SFApi) FetchPredictions(input []PredictionInput) ([]Prediction, error) {
	predictions := make([]Prediction, 0)

	for _, in := range input {
		p, err := sf.fetchPrediction(in)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch predictions for %s: %s", in.StopID, err)
		}

		predictions = append(predictions, p...)
	}

	return predictions, nil
}

func (sf *SFApi) FetchIncidents() ([]Incident, error) {
	i := []Incident{
		{Description: "Trains delayed by 3 hours", DateUpdated: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC), Affected: []string{"Outer"}, Type: "Delay"},
		{Description: "All trains are broken", DateUpdated: time.Date(2009, time.December, 11, 23, 0, 0, 0, time.UTC), Affected: []string{"Central"}, Type: "Alert"},
	}

	return i, nil
}

func (sf *SFApi) GetPredictionInput(arg string) ([]PredictionInput, error) {
	db, err := data.GetDB()
	if err != nil {
		return nil, err
	}

	stops, err := db.GetStopsByLocation(data.SFSlug, true)
	if err != nil {
		return nil, err
	}

	matches := utils.FuzzyFindFrom(arg, data.SearchableStops(stops))

	if matches.Len() == 0 {
		logger.Warn(fmt.Sprintf("Skipping '%s': could not find a matching station\n", arg))
		return nil, nil
	}

	if matches.Len() > 5 {
		logger.Warn(fmt.Sprintf("Skipping '%s': too many matches found\n", arg))
		return nil, nil
	}

	input := make([]PredictionInput, 0, matches.Len())

	for _, m := range matches {
		id := stops[m.Index].StopID
		agency := stops[m.Index].AgencyID
		input = append(input, PredictionInput{id, agency})
	}

	return input, nil
}

func (sf *SFApi) GetLineColor(stop string) (string, string) {
	white, black := "#FFFFFF", "#000000"
	trimmed := strings.Trim(stop, " ")
	switch trimmed {
	case "RD":
		return "#ED1D24", black
	case "OR":
		return "#FAA61A", black
	case "YL":
		return "#FFE600", black
	case "GR":
		return "#50B848", white
	case "BL":
		return "#009AD9", white
	default:
		return white, black
	}
}

func (sf *SFApi) IsGhostTrain(line, destination string) bool {
	return line == "--" || destination == "NO PASSENGERS"
}
