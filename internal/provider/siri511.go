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
	store   staticLookup
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
		ResponseTimestamp      string `json:"ResponseTimestamp"`
		StopMonitoringDelivery struct {
			MonitoredStopVisit []struct {
				MonitoredVehicleJourney sfMonitoredVehicleJourney `json:"MonitoredVehicleJourney"`
			} `json:"MonitoredStopVisit"`
		} `json:"StopMonitoringDelivery"`
	} `json:"ServiceDelivery"`
}

type sfTranslation struct {
	Language string `json:"Language"`
	Text     string `json:"Text"`
}

type sfTranslatedText struct {
	Translations []sfTranslation `json:"Translations"`
}

type sfServiceAlertsResponse struct {
	Header struct {
		GtfsRealtimeVersion string `json:"GtfsRealtimeVersion"`
		Incrementality      int    `json:"incrementality"`
		Timestamp           int64  `json:"Timestamp"`
	} `json:"Header"`
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
				RouteID  string `json:"RouteId"`
				Trip     struct {
					TripID string `json:"TripId"`
				} `json:"Trip"`
			} `json:"InformedEntities"`
			Cause           int              `json:"cause"`
			Effect          int              `json:"effect"`
			HeaderText      sfTranslatedText `json:"HeaderText"`
			DescriptionText sfTranslatedText `json:"DescriptionText"`
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

func (sf *SFClient) Seed(ctx context.Context) (*transit.Static, error) {
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

func (sf *SFClient) fetchDepartures(ctx context.Context, ref transit.StopRef) ([]transit.Departure, time.Time, error) {
	req, err := sf.BuildRequest(ctx, http.MethodGet, "transit", "StopMonitoring")
	if err != nil {
		return nil, time.Time{}, err
	}

	q := req.URL.Query()
	q.Add("api_key", sf.apiKey)
	q.Add("agency", ref.AgencyID)
	q.Add("stopcode", ref.StopID)
	q.Add("format", "json")
	req.URL.RawQuery = q.Encode()

	resp, err := sf.http.Do(req)
	if err != nil {
		return nil, time.Time{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, time.Time{}, newSFHTTPError(req, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, time.Time{}, err
	}

	// Remove BOM from response
	body = bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))

	var stopMonitoring sfStopMonitoringResponse

	err = json.Unmarshal(body, &stopMonitoring)
	if err != nil {
		return nil, time.Time{}, err
	}

	// A missing or malformed timestamp should only affect the age and not the departure list.
	asOf, _ := time.Parse(time.RFC3339, stopMonitoring.ServiceDelivery.ResponseTimestamp)

	visits := stopMonitoring.ServiceDelivery.StopMonitoringDelivery.MonitoredStopVisit
	departures := make([]transit.Departure, 0, len(visits))

	for _, msv := range visits {
		mvj := msv.MonitoredVehicleJourney
		if isSFGhostTrain(mvj) {
			continue
		}

		var arrivalString string
		// Caltain API does not return ExpectedArrivalTime, it's set to null
		if mvj.MonitoredCall.ExpectedArrivalTime != "" {
			arrivalString = mvj.MonitoredCall.ExpectedArrivalTime
		} else {
			arrivalString = mvj.MonitoredCall.AimedArrivalTime
		}

		arrivalTime, err := time.Parse(time.RFC3339, arrivalString)
		if err != nil {
			return nil, time.Time{}, err
		}

		bg, fg := sfLineColor(mvj.LineRef)

		d := transit.Departure{
			Source:    source511,
			StopID:    ref.StopID,
			StopName:  mvj.MonitoredCall.StopPointName,
			AgencyID:  ref.AgencyID,
			Mode:      "",
			Line:      mvj.LineRef,
			LineColor: bg,
			LineText:  fg,
			Headsign:  mvj.MonitoredCall.DestinationDisplay,
			Direction: mvj.DirectionRef,
			Arrives:   arrivalTime,
		}

		departures = append(departures, d)
	}

	return departures, asOf, nil
}

func (sf *SFClient) Departures(ctx context.Context, refs []transit.StopRef) (transit.DepartureSet, error) {
	var departures []transit.Departure
	var errs []error
	var asOf time.Time

	for _, r := range refs {
		d, t, err := sf.fetchDepartures(ctx, r)
		if err != nil {
			errs = append(errs, fmt.Errorf("departures at %s: %w", r.Name, err))
			continue
		}

		departures = append(departures, d...)
		asOf = older(asOf, t)
	}

	// Every stop lost its request, so there's no set to hand back and nothing to date it with.
	if len(refs) > 0 && len(errs) == len(refs) {
		return transit.DepartureSet{}, errors.Join(errs...)
	}

	return transit.DepartureSet{
		Departures: departures,
		Sources: []transit.SourceStatus{{
			Source: source511,
			AsOf:   asOf,
			Err:    errors.Join(errs...),
		}},
	}, nil
}

func (sf *SFClient) Alerts(ctx context.Context) (transit.AlertSet, error) {
	agencies, err := sf.store.Agencies(ctx, transit.SFSlug)
	if err != nil {
		return transit.AlertSet{}, err
	}

	var alerts []transit.Alert
	var errs []error
	var asOf time.Time

	// TODO surely there's a way to fetch multiple agencies at once
	for _, agency := range agencies {
		a, t, err := sf.fetchAgencyAlerts(ctx, agency)
		if err != nil {
			errs = append(errs, fmt.Errorf("alerts for %s: %w", agency.Name, err))
			continue
		}

		alerts = append(alerts, a...)
		asOf = older(asOf, t)
	}

	return transit.AlertSet{
		Alerts: alerts,
		Sources: []transit.SourceStatus{{
			Source: source511,
			AsOf:   asOf,
			Err:    errors.Join(errs...),
		}},
	}, nil
}

func (sf *SFClient) fetchAgencyAlerts(ctx context.Context, agency transit.Agency) ([]transit.Alert, time.Time, error) {
	req, err := sf.BuildRequest(ctx, http.MethodGet, "transit", "servicealerts")
	if err != nil {
		return nil, time.Time{}, err
	}

	q := req.URL.Query()
	q.Add("api_key", sf.apiKey)
	q.Add("agency", agency.AgencyID)
	q.Add("format", "json")
	req.URL.RawQuery = q.Encode()

	resp, err := sf.http.Do(req)
	if err != nil {
		return nil, time.Time{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, time.Time{}, newSFHTTPError(req, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, time.Time{}, err
	}

	// Remove BOM from response
	body = bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))

	var serviceAlerts sfServiceAlertsResponse
	err = json.Unmarshal(body, &serviceAlerts)
	if err != nil {
		return nil, time.Time{}, err
	}

	asOf := sfTimestamp(serviceAlerts.Header.Timestamp)

	var alerts []transit.Alert

	for _, entity := range serviceAlerts.Entities {
		var start, end time.Time
		if len(entity.Alert.ActivePeriods) > 0 {
			start = sfTimestamp(entity.Alert.ActivePeriods[0].Start)
			end = sfTimestamp(entity.Alert.ActivePeriods[0].End)
		}

		var affected []transit.AlertRef
		// TODO: StopID could be duplicated because they would have different RouteIDs * sigh *
		// TODO: Trip.TripID needs the trips.txt join from P5b-1 before it can become a route ref
		// Agency is on every informed entity. The alert already knows which agency is affected.
		for _, e := range entity.Alert.InformedEntities {
			if e.StopID != "" {
				affected = append(affected, transit.AlertRef{
					Kind: transit.RefStop,
					ID:   e.StopID,
				})
			}

			if e.RouteID != "" {
				bg, fg := sfLineColor(e.RouteID)
				affected = append(affected, transit.AlertRef{
					Kind:      transit.RefRoute,
					ID:        e.RouteID,
					Color:     bg,
					TextColor: fg,
				})
			}
		}

		alert := transit.Alert{
			Source:      source511,
			Description: alertText(entity.Alert.HeaderText, entity.Alert.DescriptionText),
			Effect:      gtfs.ResolveGTFSAlertEffect(entity.Alert.Effect),
			AgencyID:    agency.AgencyID,
			Affected:    affected,
			Starts:      start, // TODO: Figure out what scenario triggers multiple active periods
			Ends:        end,   // TODO: ^ here too
			Updated:     asOf,
		}

		alerts = append(alerts, alert)
	}

	return alerts, asOf, nil

}

func older(a, b time.Time) time.Time {
	if b.IsZero() {
		return a
	}

	if a.IsZero() || b.Before(a) {
		return b
	}

	return a
}

// 511 omits a timestamp rather than sending zero, and time.Unix(0, 0) is 1970.
func sfTimestamp(sec int64) time.Time {
	if sec == 0 {
		return time.Time{}
	}

	return time.Unix(sec, 0)
}

// Caltrain sends its text in HeaderText and leaves DescriptionText empty.
// BART sends a banner like "BART.gov Alert" as the header and the real text as the description.
func alertText(header, description sfTranslatedText) string {
	headerText := strings.TrimSpace(englishText(header))
	descriptionText := strings.TrimSpace(englishText(description))

	if headerText == "" {
		return descriptionText
	}

	if descriptionText == "" {
		return headerText
	}

	return headerText + ": " + descriptionText
}

// 511 ships four languages and doesn't promise an order.
func englishText(t sfTranslatedText) string {
	for _, tr := range t.Translations {
		if tr.Language == "en" {
			return tr.Text
		}
	}

	return ""
}

// StopRefs returns the ref 511 understands. Its stop codes are the seeded IDs verbatim.
func (sf *SFClient) StopRefs(s transit.Stop) []transit.StopRef {
	return []transit.StopRef{{
		StopID:   s.StopID,
		Name:     s.Name,
		AgencyID: s.AgencyID,
		Source:   source511,
	}}
}

func sfLineColor(line string) (string, string) {
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

func isSFGhostTrain(mvj sfMonitoredVehicleJourney) bool {
	return mvj.LineRef == "--" || mvj.MonitoredCall.DestinationDisplay == "NO PASSENGERS"
}
