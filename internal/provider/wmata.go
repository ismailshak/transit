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
	"github.com/ismailshak/transit/internal/logger"
	"github.com/ismailshak/transit/internal/store"
	"github.com/ismailshak/transit/internal/transit"
)

const (
	wmataDateTimeLayout = "2006-01-02T15:04:05"
)

// WMATClient is the API to interact with WMATA
type WMATClient struct {
	apiKey  string
	baseURL string
	http    *http.Client
	log     *logger.Logger
	store   *store.Store
}

// WMATA's predictions API response
type wmataPredictionsResponse struct {
	Trains []Prediction
}

type wmataIncident struct {
	Description   string
	IncidentType  string
	LinesAffected string
	DateUpdated   string
}

// WMATA's incidents API response
type wmataIncidentsResponse struct {
	Incidents []wmataIncident
}

func (w *WMATClient) BuildRequest(ctx context.Context, method string, route ...string) (*http.Request, error) {
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

func (w *WMATClient) FetchStaticData(ctx context.Context) (*transit.StaticData, error) {
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

	defer func() {
		f.Close() //nolint:errcheck // only here for the copy below, we close it ourselves after that
		if err := os.RemoveAll(zipPath); err != nil {
			w.log.Warn(fmt.Sprintf("Leftover gtfs archive at '%s': %s", zipPath, err))
		}
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
		if err := os.RemoveAll(feed); err != nil {
			w.log.Warn(fmt.Sprintf("Leftover gtfs data at '%s': %s", feed, err))
		}
	}()

	err = gtfs.UnzipStaticGTFS(zipPath, feed)
	if err != nil {
		return nil, err
	}

	return gtfs.ParseGTFS(feed, transit.DMVSlug, transit.TrainStation, "MET")
}

func (w *WMATClient) FetchPredictions(ctx context.Context, input []PredictionInput) ([]Prediction, error) {
	codes := make([]string, 0, len(input))
	for _, i := range input {
		codes = append(codes, i.StopID)
	}

	req, err := w.BuildRequest(ctx, http.MethodGet, "StationPrediction.svc/json/GetPrediction", strings.Join(codes, ","))
	if err != nil {
		return nil, err
	}

	resp, err := w.http.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Avoid the expensive conversion unless we have to
	if w.log.Verbose() {
		w.log.Debug(string(body))
	}

	if resp.StatusCode != 200 {
		return nil, &HTTPError{StatusCode: resp.StatusCode, URL: req.URL.String()}
	}

	var predictions wmataPredictionsResponse
	err = json.Unmarshal(body, &predictions)

	if err != nil {
		return nil, fmt.Errorf("parse predictions response: %w", err)
	}

	if len(predictions.Trains) == 0 {
		return nil, ErrNoDepartures
	}

	return predictions.Trains, nil
}

func (w *WMATClient) FetchIncidents(ctx context.Context) ([]Incident, error) {
	req, err := w.BuildRequest(ctx, http.MethodGet, "Incidents.svc/json/Incidents")
	if err != nil {
		return nil, err
	}

	resp, err := w.http.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Avoid the expensive conversion unless we have to
	if w.log.Verbose() {
		w.log.Debug(string(body))
	}

	if resp.StatusCode != 200 {
		return nil, &HTTPError{StatusCode: resp.StatusCode, URL: req.URL.String()}
	}

	var incidentsRes wmataIncidentsResponse
	err = json.Unmarshal(body, &incidentsRes)

	if err != nil {
		return nil, fmt.Errorf("parse incidents response: %w", err)
	}

	var incidents []Incident
	for _, res := range incidentsRes.Incidents {
		date, _ := time.Parse(wmataDateTimeLayout, res.DateUpdated)
		inc := Incident{
			Description: res.Description,
			DateUpdated: date,
			Affected:    parseLinesAffected(res.LinesAffected),
			Type:        res.IncidentType,
		}

		incidents = append(incidents, inc)
	}

	return incidents, nil
}

func (w *WMATClient) GetPredictionInput(ctx context.Context, arg string) ([]PredictionInput, error) {
	stops, err := w.store.StopsByLocation(ctx, transit.DMVSlug, true)
	if err != nil {
		return nil, err
	}

	matches := store.FuzzyFindFrom(arg, store.SearchableStops(stops))

	if matches.Len() == 0 {
		w.log.Warn(fmt.Sprintf("Skipping '%s': could not find a matching station\n", arg))
		return nil, nil
	}

	if matches.Len() > 5 {
		w.log.Warn(fmt.Sprintf("Skipping '%s': too many matches found\n", arg))
		return nil, nil
	}

	input := make([]PredictionInput, 0, matches.Len())

	for _, m := range matches {
		id := stops[m.Index].StopID
		formattedID := formatWMATAStopID(id)
		for _, id := range formattedID {
			input = append(input, PredictionInput{id, stops[m.Index].AgencyID})
		}
	}

	return input, nil
}

func (w *WMATClient) GetLineColor(stop string) (string, string) {
	white, black := "#FFFFFF", "#000000"
	switch stop {
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

func (w *WMATClient) IsGhostTrain(line, destination string) bool {
	return line == "--" || destination == "No Passenger" || line == "No"
}

// Parses the affected format in the incidents response. Semi-colon separated with a space
func parseLinesAffected(lines string) []string {
	splitSlice := strings.Split(strings.ReplaceAll(lines, " ", ""), ";")

	var filteredSlice []string
	for _, s := range splitSlice {
		if s != "" {
			filteredSlice = append(filteredSlice, s)
		}
	}

	return filteredSlice
}

// All WMATA train IDs have the format `STN_X_X` where each X is a unique ID
// (train stations can have multiple)
func formatWMATAStopID(id string) []string {
	return strings.Split(id, "_")[1:]
}
