package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ismailshak/transit/internal/gtfs"
	"github.com/ismailshak/transit/internal/logger"
	"github.com/ismailshak/transit/internal/store"
	"github.com/ismailshak/transit/internal/transit"
)

// SFClient is the API to interact with San Francisco's 511 API
type SFClient struct {
	apiKey  string
	baseURL string
	http    *http.Client
	log     *logger.Logger
	now     func() time.Time
	store   *store.Store
}

type sfStopPlace struct {
	ID       string `json:"@id"`
	Name     string `json:"Name"`
	Centroid struct {
		Location struct {
			Latitude  string
			Longitude string
		}
	}
	TransportMode string `json:"TransportMode"`
}

type sfStopPlacesResponse struct {
	Siri struct {
		ServiceDelivery struct {
			ResponseTimestamp  string `json:"ResponseTimestamp"`
			DataObjectDelivery struct {
				ResponseTimestamp string `json:"ResponseTimestamp"`
				DataObjects       struct {
					SiteFrame struct {
						StopPlaces struct {
							StopPlace []sfStopPlace
						} `json:"stopPlaces"`
					}
				} `json:"dataObjects"`
			}
		}
	}
}

type sfMonitoredVehicleJourney struct {
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

type sfStopMonitoringResponse struct {
	ServiceDelivery struct {
		StopMonitoringDelivery struct {
			MonitoredStopVisit []struct {
				MonitoredVehicleJourney sfMonitoredVehicleJourney
			}
		}
	}
}

type sfServiceAlertsResponse struct {
	Entities []struct {
		ID    string `json:"Id"`
		Alert struct {
			ActivePeriods []struct {
				Start int64
				End   int64
			}
			InformedEntities []struct {
				AgencyID string
				StopID   string
			}
			Cause           int `json:"cause"`
			Effect          int `json:"effect"`
			DescriptionText struct {
				Translations []struct {
					Language string
					Text     string
				}
			}
		}
	}
}

func newSFHTTPError(req *http.Request, statusCode int) *HTTPError {
	u := *req.URL
	q := u.Query()
	q.Del("api_key")
	u.RawQuery = q.Encode()
	return &HTTPError{StatusCode: statusCode, URL: u.String()}

}

func (sf *SFClient) BuildRequest(ctx context.Context, method string, route ...string) (*http.Request, error) {
	parts := make([]string, 0, len(route)+1)
	parts = append(parts, sf.baseURL)
	parts = append(parts, route...)
	url := strings.Join(parts, "/")

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (sf *SFClient) fetchStopsForAgency(ctx context.Context, agency *transit.Agency) ([]*transit.Stop, error) {
	req, err := sf.BuildRequest(ctx, http.MethodGet, "transit", "stopplaces")
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("api_key", sf.apiKey)
	q.Add("operator_id", agency.AgencyID)
	q.Add("format", "json")
	req.URL.RawQuery = q.Encode()

	resp, err := sf.http.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, newSFHTTPError(req, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	// Remove BOM from response
	body = bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))

	var stopPlaces sfStopPlacesResponse
	err = json.Unmarshal(body, &stopPlaces)

	if err != nil {
		return nil, err
	}

	var stops []*transit.Stop

	for _, sp := range stopPlaces.Siri.ServiceDelivery.DataObjectDelivery.DataObjects.SiteFrame.StopPlaces.StopPlace {
		var stopType transit.StopType

		switch sp.TransportMode {
		case "bus":
			stopType = transit.BusStop
		case "rail": // CT train type
			stopType = transit.TrainStation
		case "intercityRail": // BART train type
			stopType = transit.TrainStation
		default:
			continue
		}

		stop := transit.Stop{
			AgencyID:  agency.AgencyID,
			Latitude:  sp.Centroid.Location.Latitude,
			Location:  transit.SFSlug,
			Longitude: sp.Centroid.Location.Longitude,
			Name:      sp.Name,
			StopID:    sp.ID,
			Type:      stopType,
		}

		stops = append(stops, &stop)
	}

	return stops, nil
}

func (sf *SFClient) FetchStaticData(ctx context.Context) (*transit.StaticData, error) {
	bart := &transit.Agency{
		AgencyID: "BA",
		Language: "en",
		Location: transit.SFSlug,
		Name:     "Bay Area Rapid Transit",
		Timezone: "America/Los_Angeles",
	}

	bartStops, err := sf.fetchStopsForAgency(ctx, bart)
	if err != nil {
		return nil, fmt.Errorf("fetch BART stops: %w", err)
	}

	cal := &transit.Agency{
		AgencyID: "CT",
		Language: "en",
		Location: transit.SFSlug,
		Name:     "Caltrain",
		Timezone: "America/Los_Angeles",
	}

	calStops, err := sf.fetchStopsForAgency(ctx, cal)
	if err != nil {
		return nil, fmt.Errorf("fetch Caltrain stops: %w", err)
	}

	var stops []*transit.Stop

	stops = append(bartStops, calStops...)

	staticData := transit.StaticData{
		Agencies: []*transit.Agency{bart, cal},
		Stops:    stops,
	}

	return &staticData, nil
}

// Removes the `-X` suffix from the line name where X is a direction (e.g. -N, -S, -E, -W)
// and abbreviates the line name
func (sf *SFClient) formatLine(line string) string {
	trimmed, _, _ := strings.Cut(line, "-")
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

func (sf *SFClient) fetchPrediction(ctx context.Context, in PredictionInput) ([]Prediction, error) {
	req, err := sf.BuildRequest(ctx, http.MethodGet, "transit", "StopMonitoring")
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("api_key", sf.apiKey)
	q.Add("agency", in.AgencyID)
	q.Add("stopcode", in.StopID)
	q.Add("format", "json")
	req.URL.RawQuery = q.Encode()

	resp, err := sf.http.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, newSFHTTPError(req, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Remove BOM from response
	body = bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))

	// Avoid the expensive conversion unless we have to
	if sf.log.Verbose() {
		sf.log.Debug(string(body))
	}

	var stopMonitoring sfStopMonitoringResponse

	err = json.Unmarshal(body, &stopMonitoring)
	if err != nil {
		return nil, err
	}

	monitoredStopVisits := len(stopMonitoring.ServiceDelivery.StopMonitoringDelivery.MonitoredStopVisit)
	if monitoredStopVisits == 0 {
		return nil, ErrNoDepartures
	}

	predictions := make([]Prediction, 0, monitoredStopVisits)

	for _, msv := range stopMonitoring.ServiceDelivery.StopMonitoringDelivery.MonitoredStopVisit {
		mvj := msv.MonitoredVehicleJourney
		var arrivalString string
		// Caltain API does not return ExpectedArrivalTime, it's set to null
		if mvj.MonitoredCall.ExpectedArrivalTime != "" {
			arrivalString = mvj.MonitoredCall.ExpectedArrivalTime
		} else {
			arrivalString = mvj.MonitoredCall.AimedArrivalTime
		}

		arrivalTime, err := time.Parse(time.RFC3339, arrivalString)
		if err != nil {
			return nil, err
		}

		now := sf.now()
		arrival := strconv.Itoa(int(arrivalTime.Sub(now).Minutes()))

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

func (sf *SFClient) FetchPredictions(ctx context.Context, input []PredictionInput) ([]Prediction, error) {
	predictions := make([]Prediction, 0)

	for _, in := range input {
		p, err := sf.fetchPrediction(ctx, in)
		if errors.Is(err, ErrNoDepartures) {
			continue
		}

		if err != nil {
			return nil, fmt.Errorf("fetch predictions for %s: %w", in.StopID, err)
		}

		predictions = append(predictions, p...)
	}

	if len(predictions) == 0 {
		return nil, ErrNoDepartures
	}

	return predictions, nil
}

func (sf *SFClient) FetchIncidents(ctx context.Context) ([]Incident, error) {
	agencies, err := sf.store.LocationAgencies(ctx, transit.SFSlug)
	if err != nil {
		return nil, err
	}

	shouldDisplayAgency := len(agencies) > 1

	incidents := make([]Incident, 0, 16) // TODO: figure out a good capacity (16 is arbitrary)

	// TODO surely there's a way to fetch multiple agencies at once
	for _, agency := range agencies {
		var agencyName string
		if shouldDisplayAgency {
			agencyName = agency.Name
		}

		inc, err := sf.fetchAgencyIncidents(ctx, agency, agencyName)
		if err != nil {
			return nil, fmt.Errorf("alerts for %s: %w", agency.AgencyID, err)
		}

		incidents = append(incidents, inc...)
	}

	return incidents, nil
}

func (sf *SFClient) fetchAgencyIncidents(ctx context.Context, agency transit.Agency, agencyName string) ([]Incident, error) {
	req, err := sf.BuildRequest(ctx, http.MethodGet, "transit", "servicealerts")
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("api_key", sf.apiKey)
	q.Add("agency", agency.AgencyID)
	q.Add("format", "json")
	req.URL.RawQuery = q.Encode()

	resp, err := sf.http.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, newSFHTTPError(req, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Remove BOM from response
	body = bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))

	var serviceAlerts sfServiceAlertsResponse
	err = json.Unmarshal(body, &serviceAlerts)
	if err != nil {
		return nil, err
	}

	// Avoid the expensive conversion unless we have to
	if sf.log.Verbose() {
		sf.log.Debug(string(body))
	}

	var incidents []Incident

	for _, entity := range serviceAlerts.Entities {
		var start, end time.Time
		if len(entity.Alert.ActivePeriods) > 0 {
			start = time.Unix(entity.Alert.ActivePeriods[0].Start, 0)
			end = time.Unix(entity.Alert.ActivePeriods[0].End, 0)
		}

		var affected []string
		// TODO: StopID could be duplicated because they would have different RouteIDs * sigh *
		for _, e := range entity.Alert.InformedEntities {
			if e.StopID != "" {
				affected = append(affected, e.StopID) // TODO: stop name instead? hopefully not a performance hit (but maybe doesn't matter due to incidents being rare)
			}
		}

		inc := Incident{
			ActivePeriodStart: start,                                             // TODO: Figure out what scenario triggers multiple active periods
			ActivePeriodEnd:   end,                                               // TODO: ^ here too
			Affected:          affected,                                          // TODO: parse informed entities
			Agency:            agencyName,                                        // Derive from header?
			Description:       entity.Alert.DescriptionText.Translations[0].Text, // TODO: assumes english is always first
			DateUpdated:       time.Time{},                                       // TODO: figure out where this is
			Type:              gtfs.ResolveGTFSAlertEffect(entity.Alert.Effect),
		}

		incidents = append(incidents, inc)
	}

	return incidents, nil

}

func (sf *SFClient) GetPredictionInput(ctx context.Context, arg string) ([]PredictionInput, error) {
	stops, err := sf.store.MatchStops(ctx, transit.SFSlug, arg)
	if err != nil {
		return nil, err
	}

	if len(stops) == 0 {
		sf.log.Warn(fmt.Sprintf("Skipping '%s': could not find a matching station\n", arg))
		return nil, nil
	}

	if len(stops) > 5 {
		sf.log.Warn(fmt.Sprintf("Skipping '%s': too many matches found\n", arg))
		return nil, nil
	}

	input := make([]PredictionInput, 0, len(stops))

	for _, s := range stops {
		input = append(input, PredictionInput{StopID: s.StopID, AgencyID: s.AgencyID})
	}

	return input, nil
}

func (sf *SFClient) GetLineColor(stop string) (string, string) {
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

func (sf *SFClient) IsGhostTrain(line, destination string) bool {
	return line == "--" || destination == "NO PASSENGERS"
}
