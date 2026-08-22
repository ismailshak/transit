package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ismailshak/transit/internal/config"
	"github.com/ismailshak/transit/internal/gtfs"
	"github.com/ismailshak/transit/internal/transit"
)

const (
	wmataDateTimeLayout = "2006-01-02T15:04:05"
	sourceWMATARail     = "wmata-rail"
	agencyWMATA         = "MET"
)

// WMATAClient is the API to interact with WMATA.
type WMATAClient struct {
	apiKey   string
	baseURL  string
	location *time.Location
	http     *http.Client
	now      func() time.Time
}

type wmataTrain struct {
	Car             string `json:"Car"`
	Destination     string `json:"Destination"`
	DestinationCode string `json:"DestinationCode"`
	DestinationName string `json:"DestinationName"`
	Group           string `json:"Group"`
	Line            string `json:"Line"`
	LocationCode    string `json:"LocationCode"`
	LocationName    string `json:"LocationName"`
	Min             string `json:"Min"`
}

type wmataPredictionsResponse struct {
	Trains []wmataTrain `json:"Trains"`
}

type wmataIncident struct {
	IncidentID   string `json:"IncidentID"`
	Description  string `json:"Description"`
	IncidentType string `json:"IncidentType"`
	// WMATA plans to return this as an array.
	// https://developer.wmata.com/api-details#api=54763641281d83086473f232&operation=54763641281d830c946a3d77
	LinesAffected string `json:"LinesAffected"`
	DateUpdated   string `json:"DateUpdated"`
}

type wmataIncidentsResponse struct {
	Incidents []wmataIncident `json:"Incidents"`
}

func (w *WMATAClient) BuildRequest(ctx context.Context, method string, route ...string) (*http.Request, error) {
	parts := make([]string, 0, len(route)+1)
	parts = append(parts, w.baseURL)
	parts = append(parts, route...)
	url := strings.Join(parts, "/")

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("api_key", w.apiKey)

	return req, nil
}

func (w *WMATAClient) Seed(ctx context.Context) (*transit.Static, error) {
	req, err := w.BuildRequest(ctx, http.MethodGet, "gtfs/rail-gtfs-static.zip")
	if err != nil {
		return nil, err
	}

	resp, err := w.http.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, &HTTPError{StatusCode: resp.StatusCode, URL: req.URL.String()}
	}

	configDir, err := config.GetConfigDir()
	if err != nil {
		return nil, err
	}

	zipPath := filepath.Join(configDir, "wmata_rail_gtfs_static.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		return nil, err
	}

	// TODO: name the file we couldn't clean up, once there's a debug stream to name it on
	defer func() {
		f.Close() //nolint:errcheck // only here for the copy below, we close it ourselves after that
		_ = os.RemoveAll(zipPath)
	}()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return nil, fmt.Errorf("download gtfs archive: %w", err)
	}

	// Close it before we read it back, a bad write only shows up here
	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("write gtfs archive %s: %w", zipPath, err)
	}

	dirName := "gtfs_static_" + strconv.FormatInt(time.Now().Unix(), 10)
	feed := filepath.Join(configDir, dirName)
	if err = os.MkdirAll(feed, 0o755); err != nil {
		return nil, err
	}

	defer func() {
		_ = os.RemoveAll(feed)
	}()

	err = gtfs.UnzipStaticGTFS(zipPath, feed)
	if err != nil {
		return nil, err
	}

	return gtfs.ParseGTFS(feed, transit.DMVSlug, transit.TrainStation, agencyWMATA)
}

func (w *WMATAClient) Departures(ctx context.Context, refs []transit.StopRef) (transit.DepartureSet, error) {
	departures, asOf, err := w.fetchDepartures(ctx, refs)
	if err != nil {
		return transit.DepartureSet{}, err
	}

	return transit.DepartureSet{
		Departures: departures,
		Sources:    []transit.SourceStatus{{Source: sourceWMATARail, AsOf: asOf}},
	}, nil
}

func (w *WMATAClient) fetchDepartures(ctx context.Context, refs []transit.StopRef) ([]transit.Departure, time.Time, error) {
	codes := make([]string, 0, len(refs))
	for _, i := range refs {
		codes = append(codes, i.StopID)
	}

	req, err := w.BuildRequest(ctx, http.MethodGet, "StationPrediction.svc/json/GetPrediction", strings.Join(codes, ","))
	if err != nil {
		return nil, time.Time{}, err
	}

	resp, err := w.http.Do(req)

	if err != nil {
		return nil, time.Time{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, time.Time{}, &HTTPError{StatusCode: resp.StatusCode, URL: req.URL.String()}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, time.Time{}, err
	}

	asOf := w.now()

	var predictionsResponse wmataPredictionsResponse
	err = json.Unmarshal(body, &predictionsResponse)

	if err != nil {
		return nil, time.Time{}, fmt.Errorf("parse predictions response: %w", err)
	}

	departures := make([]transit.Departure, 0, len(predictionsResponse.Trains))
	for _, t := range predictionsResponse.Trains {
		if isWMATAGhostTrain(t) {
			continue
		}

		arrives, ok := railArrival(t.Min, asOf)
		if !ok {
			continue
		}

		bg, fg := wmataLineColor(t.Line)

		departures = append(departures, transit.Departure{
			Source:    sourceWMATARail,
			StopID:    t.LocationCode,
			StopName:  t.LocationName,
			AgencyID:  agencyWMATA,
			Mode:      "",
			Line:      t.Line,
			LineColor: bg,
			LineText:  fg,
			Headsign:  t.Destination,
			Direction: t.Group,
			Arrives:   arrives,
		})
	}

	return departures, asOf, nil
}

func (w *WMATAClient) Alerts(ctx context.Context) (transit.AlertSet, error) {
	alerts, err := w.fetchAlerts(ctx)

	source := transit.SourceStatus{Source: sourceWMATARail, Err: err}

	if err == nil {
		source.AsOf = w.now()
	}

	return transit.AlertSet{
		Alerts:  alerts,
		Sources: []transit.SourceStatus{source},
	}, nil
}

func (w *WMATAClient) fetchAlerts(ctx context.Context) ([]transit.Alert, error) {
	req, err := w.BuildRequest(ctx, http.MethodGet, "Incidents.svc/json/Incidents")
	if err != nil {
		return nil, err
	}

	resp, err := w.http.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, &HTTPError{StatusCode: resp.StatusCode, URL: req.URL.String()}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var incidentsRes wmataIncidentsResponse
	err = json.Unmarshal(body, &incidentsRes)

	if err != nil {
		return nil, fmt.Errorf("parse incidents response: %w", err)
	}

	alerts := make([]transit.Alert, 0, len(incidentsRes.Incidents))
	for _, inc := range incidentsRes.Incidents {
		// A missing or malformed timestamp should only affect the age and not the alert.
		date, _ := time.ParseInLocation(wmataDateTimeLayout, inc.DateUpdated, w.location)
		alert := transit.Alert{
			Source:      sourceWMATARail,
			AgencyID:    agencyWMATA,
			Affected:    parseAffected(inc.LinesAffected),
			Description: inc.Description,
			Effect:      inc.IncidentType,
			Updated:     date,
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil

}

// StopRefs splits a station into the platform codes WMATA understands.
func (w *WMATAClient) StopRefs(s transit.Stop) []transit.StopRef {
	ids := formatWMATAStopID(s.StopID)
	refs := make([]transit.StopRef, len(ids))
	for i, id := range ids {
		refs[i] = transit.StopRef{
			StopID:   id,
			Name:     s.Name,
			AgencyID: s.AgencyID,
			Source:   sourceWMATARail,
		}
	}

	return refs
}

// All WMATA train IDs have the format `STN_X_X` where each X is a unique ID.
// A station can have more than one ID.
func formatWMATAStopID(id string) []string {
	return strings.Split(id, "_")[1:]
}

func wmataLineColor(line string) (string, string) {
	white, black := "#FFFFFF", "#000000"
	switch line {
	case "SV", "Silver":
		return "#919D9D", black
	case "RD", "Red":
		return "#BF0D3E", white
	case "BL", "Blue":
		return "#009CDE", white
	case "YL", "Yellow":
		return "#FFD100", black
	case "OR", "Orange":
		return "#ED8B00", black
	case "GR", "Green":
		return "#00B140", white
	default:
		return white, black
	}
}

func isWMATAGhostTrain(t wmataTrain) bool {
	return t.Line == "--" || t.Line == "No" || t.DestinationName == "No Passenger"
}

// Parses the affected format in the incidents response. Semi-colon separated with a space.
func parseAffected(lines string) []transit.AlertRef {
	parts := strings.Split(strings.ReplaceAll(lines, " ", ""), ";")

	var refs []transit.AlertRef
	for _, p := range parts {
		if p == "" {
			continue
		}

		bg, fg := wmataLineColor(p)
		refs = append(refs, transit.AlertRef{
			Kind:      transit.RefRoute,
			ID:        p,
			Color:     bg,
			TextColor: fg,
		})
	}

	return refs
}

// railArrival converts the rendered countdown WMATA returns into an absolute
// instant in time.
func railArrival(min string, asOf time.Time) (time.Time, bool) {
	switch min {
	case "BRD":
		return asOf, true
	case "ARR":
		return asOf.Add(30 * time.Second), true // 30s is a guess. WMATA doesn't specify what it translates to.
	case "---", "":
		return time.Time{}, false
	}

	n, err := strconv.Atoi(min)
	if err != nil {
		return time.Time{}, false
	}

	return asOf.Add(time.Duration(n) * time.Minute), true
}
