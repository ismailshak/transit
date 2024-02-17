package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ismailshak/transit/internal/config"
	"github.com/ismailshak/transit/internal/data"
	"github.com/ismailshak/transit/internal/logger"
	"github.com/ismailshak/transit/internal/utils"
)

const (
	DATETIME_LAYOUT = "2006-01-02T15:04:05"
)

// API to interact with WMATA
type DmvApi struct {
	apiKey  *string
	baseUrl string
}

// WMATA's predictions API response
type WMATA_PredictionsResponse struct {
	Trains []Prediction
}

type WMATA_Incident struct {
	Description   string
	IncidentType  string
	LinesAffected string
	DateUpdated   string
}

// WMATA's incidents API response
type WMATA_IncidentsResponse struct {
	Incidents []WMATA_Incident
}

func (dmv *DmvApi) BuildRequest(method string, route ...string) (*http.Request, error) {
	parts := make([]string, 0, len(route)+1)
	parts = append(parts, dmv.baseUrl)
	parts = append(parts, route...)
	url := strings.Join(parts, "/")

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("api_key", *dmv.apiKey)

	return req, nil
}

func (dmv *DmvApi) FetchStaticData() (*data.StaticData, error) {
	req, err := dmv.BuildRequest(http.MethodGet, "gtfs/rail-gtfs-static.zip")
	if err != nil {
		return nil, err
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch: received %d", resp.StatusCode)
	}

	configDir, err := config.GetConfigDir()
	if err != nil {
		return nil, err
	}

	zipPath := filepath.Join(configDir, "dmv_gtfs_static.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		return nil, err
	}

	defer func() {
		f.Close()
		os.RemoveAll(zipPath)
	}()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return nil, err
	}

	dirName := "gtfs_static_" + strconv.FormatInt(time.Now().Unix(), 10)
	feed := filepath.Join(configDir, dirName)
	if err = utils.CreateDir(feed); err != nil {
		return nil, err
	}

	defer os.RemoveAll(feed)

	err = data.UnzipStaticGTFS(zipPath, feed)
	if err != nil {
		return nil, err
	}

	return data.ParseGTFS(feed, data.DMVSlug, data.TrainStation, "MET")
}

func (dmv *DmvApi) FetchPredictions(stations []string) ([]Prediction, error) {
	codes := strings.Join(stations, ",")
	client := http.Client{}
	req, _ := dmv.BuildRequest(http.MethodGet, "StationPrediction.svc/json/GetPrediction", codes)
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	logger.Debug(string(body))

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch: received %d", resp.StatusCode)
	}

	var predictions WMATA_PredictionsResponse
	err = json.Unmarshal(body, &predictions)

	return predictions.Trains, err
}

func (dmv *DmvApi) FetchIncidents() ([]Incident, error) {
	client := http.Client{}
	req, _ := dmv.BuildRequest(http.MethodGet, "Incidents.svc/json/Incidents")
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	logger.Debug(string(body))

	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("failed to fetch: received %d", resp.StatusCode))
	}

	var incidentsRes WMATA_IncidentsResponse
	err = json.Unmarshal(body, &incidentsRes)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to parse response: %s", err))
	}

	var incidents []Incident
	for _, res := range incidentsRes.Incidents {
		date, _ := time.Parse(DATETIME_LAYOUT, res.DateUpdated)
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

func (dmv *DmvApi) GetIDFromArg(arg string) ([]string, error) {
	db, err := data.GetDB()
	if err != nil {
		return nil, err
	}

	stops, err := db.GetStopsByLocation(data.DMVSlug, true)
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

	ids := make([]string, 0, matches.Len())

	for _, m := range matches {
		id := stops[m.Index].StopID
		formattedId := formatDmvStopId(id)
		ids = append(ids, formattedId...)
	}

	return ids, nil
}

func (dmv *DmvApi) GetLineColor(stop string) (string, string) {
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

func (dmv *DmvApi) IsGhostTrain(line, destination string) bool {
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

// All DMV train IDs have the format `STN_X_X` where each X is a unique ID
// (train stations can have multiple)
func formatDmvStopId(id string) []string {
	return strings.Split(id, "_")[1:]
}
