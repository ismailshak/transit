package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/ismailshak/transit/internal/gtfs"
	"github.com/ismailshak/transit/internal/store"
	"github.com/ismailshak/transit/internal/transit"
)

const (
	source511 = "bayarea511"
)

// SFClient is the API to interact with San Francisco's 511 API.
type SFClient struct {
	apiKey  string
	baseURL string
	http    *http.Client
	store   *store.Store
}

type sfStopPlace struct {
	ID       string `json:"@id"`
	Name     string `json:"Name"`
	Centroid struct {
		Location struct {
			Latitude  string `json:"Latitude"`
			Longitude string `json:"Longitude"`
		} `json:"Location"`
	} `json:"Centroid"`
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
							StopPlace []sfStopPlace `json:"StopPlace"`
						} `json:"stopPlaces"`
					} `json:"SiteFrame"`
				} `json:"dataObjects"`
			} `json:"DataObjectDelivery"`
		} `json:"ServiceDelivery"`
	} `json:"Siri"`
}

type sfMonitoredVehicleJourney struct {
	DirectionRef    string `json:"DirectionRef"`
	LineRef         string `json:"LineRef"`
	DestinationRef  string `json:"DestinationRef"`
	DestinationName string `json:"DestinationName"`
	MonitoredCall   struct {
		AimedArrivalTime    string `json:"AimedArrivalTime"`
		DestinationDisplay  string `json:"DestinationDisplay"`
		ExpectedArrivalTime string `json:"ExpectedArrivalTime"`
		StopPointName       string `json:"StopPointName"`
	} `json:"MonitoredCall"`
}

type sfStopMonitoringResponse struct {
	ServiceDelivery struct {
		StopMonitoringDelivery struct {
			MonitoredStopVisit []struct {
				MonitoredVehicleJourney sfMonitoredVehicleJourney `json:"MonitoredVehicleJourney"`
			} `json:"MonitoredStopVisit"`
		} `json:"StopMonitoringDelivery"`
	} `json:"ServiceDelivery"`
}

type sfServiceAlertsResponse struct {
	Entities []struct {
		ID    string `json:"Id"`
		Alert struct {
			ActivePeriods []struct {
				Start int64 `json:"Start"`
				End   int64 `json:"End"`
			} `json:"ActivePeriods"`
			InformedEntities []struct {
				AgencyID string `json:"AgencyId"`
				StopID   string `json:"StopId"`
			} `json:"InformedEntities"`
			Cause           int `json:"cause"`
			Effect          int `json:"effect"`
			DescriptionText struct {
				Translations []struct {
					Language string `json:"Language"`
					Text     string `json:"Text"`
				} `json:"Translations"`
			} `json:"DescriptionText"`
		} `json:"Alert"`
	} `json:"Entities"`
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

func (sf *SFClient) fetchStopsForAgency(ctx context.Context, agency transit.Agency) ([]transit.Stop, error) {
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

	var stops []transit.Stop

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

		stops = append(stops, stop)
	}

	return stops, nil
}

func (sf *SFClient) FetchStaticData(ctx context.Context) (*transit.Static, error) {
	bart := transit.Agency{
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

	cal := transit.Agency{
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

	staticData := transit.Static{
		Agencies: []transit.Agency{bart, cal},
		Stops:    slices.Concat(bartStops, calStops),
	}

	return &staticData, nil
}

func (sf *SFClient) fetchPrediction(ctx context.Context, in PredictionInput) ([]transit.Departure, error) {
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

	var stopMonitoring sfStopMonitoringResponse

	err = json.Unmarshal(body, &stopMonitoring)
	if err != nil {
		return nil, err
	}

	monitoredStopVisits := len(stopMonitoring.ServiceDelivery.StopMonitoringDelivery.MonitoredStopVisit)
	if monitoredStopVisits == 0 {
		return nil, ErrNoDepartures
	}

	predictions := make([]transit.Departure, 0, monitoredStopVisits)

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

		bg, fg := sf.GetLineColor(mvj.LineRef)

		d := transit.Departure{
			Source:    source511,
			StopID:    in.StopID,
			StopName:  mvj.MonitoredCall.StopPointName,
			AgencyID:  in.AgencyID,
			Mode:      "",
			Line:      mvj.LineRef,
			LineColor: bg,
			LineText:  fg,
			Headsign:  mvj.MonitoredCall.DestinationDisplay,
			Direction: mvj.DirectionRef,
			Arrives:   arrivalTime,
		}

		predictions = append(predictions, d)
	}

	return predictions, nil
}

func (sf *SFClient) FetchPredictions(ctx context.Context, input []PredictionInput) ([]transit.Departure, error) {
	predictions := make([]transit.Departure, 0)

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

	// TODO: report these to the caller so it can warn on different situations
	if len(stops) == 0 {
		return nil, nil
	}

	if len(stops) > 5 {
		return nil, nil
	}

	input := make([]PredictionInput, 0, len(stops))

	for _, s := range stops {
		input = append(input, PredictionInput{StopID: s.StopID, AgencyID: s.AgencyID})
	}

	return input, nil
}

func (sf *SFClient) GetLineColor(line string) (string, string) {
	white, black := "#FFFFFF", "#000000"
	trimmed, _, _ := strings.Cut(line, "-")
	switch trimmed {
	case "Red":
		return "#ED1D24", black
	case "Orange":
		return "#FAA61A", black
	case "Yellow":
		return "#FFE600", black
	case "Green":
		return "#50B848", white
	case "Blue":
		return "#009AD9", white
	default:
		return white, black
	}
}

func (sf *SFClient) IsGhostTrain(line, destination string) bool {
	return line == "--" || destination == "NO PASSENGERS"
}
