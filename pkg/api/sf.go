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

func (sf *SFApi) FetchStaticData() (*data.StaticData, error) {
	req, err := sf.BuildRequest(http.MethodGet, "transit", "stopplaces")
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("api_key", *sf.apiKey)
	q.Add("operator_id", "BA")
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

	agency := &data.Agency{
		AgencyID: "BA",
		Language: "en",
		Location: data.SFSlug,
		Name:     "Bay Area Rapid Transit",
		Timezone: "America/Los_Angeles",
	}

	var stops []*data.Stop

	for _, sp := range stopPlaces.Siri.ServiceDelivery.DataObjectDelivery.DataObjects.SiteFrame.StopPlaces.StopPlace {
		var stopType data.StopType
		if sp.TrasnsportMode == "bus" {
			stopType = data.BusStop
		} else {
			stopType = data.TrainStation
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

	staticData := data.StaticData{
		Agencies: []*data.Agency{agency},
		Stops:    stops,
	}

	return &staticData, nil
}

// Removes the `-X` from the line name where X is a direction (e.g. -N, -S, -E, -W)
// and adds padding so that it's always 6 characters long
func (sf *SFApi) formatLine(line string) string {
	trimmed := strings.Split(line, "-")[0]
	return fmt.Sprintf("%-6s", trimmed)
}

func (sf *SFApi) FetchPredictions(ids []string) ([]Prediction, error) {
	req, err := sf.BuildRequest(http.MethodGet, "transit", "StopMonitoring")
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("api_key", *sf.apiKey)
	q.Add("agency", "BA")
	q.Add("stopcode", ids[0])
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
		arrival_time, err := time.Parse(time.RFC3339, mvj.MonitoredCall.ExpectedArrivalTime)
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

func (sf *SFApi) FetchIncidents() ([]Incident, error) {
	i := []Incident{
		{Description: "Trains delayed by 3 hours", DateUpdated: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC), Affected: []string{"Outer"}, Type: "Delay"},
		{Description: "All trains are broken", DateUpdated: time.Date(2009, time.December, 11, 23, 0, 0, 0, time.UTC), Affected: []string{"Central"}, Type: "Alert"},
	}

	return i, nil
}

func (sf *SFApi) GetIDFromArg(arg string) ([]string, error) {
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

	ids := make([]string, 0, matches.Len())

	for _, m := range matches {
		id := stops[m.Index].StopID
		ids = append(ids, id)
	}

	return ids, nil
}

func (sf *SFApi) GetLineColor(stop string) (string, string) {
	white, black := "#FFFFFF", "#000000"
	trimmed := strings.Trim(stop, " ")
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

func (sf *SFApi) IsGhostTrain(line, destination string) bool {
	return line == "--" || destination == "NO PASSENGERS"
}
