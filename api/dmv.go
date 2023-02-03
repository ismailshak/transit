package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ismailshak/transit/logger"
)

// API to interact with DMV Metro
type DmvApi struct {
	apiKey  *string
	baseUrl string
}

// API response for fetching train arrival information
type TimingResponse struct {
	Trains []Timing
}

// Fetch latest train arrival information
func (dmv DmvApi) ListTimings(stations []string) ([]Timing, error) {
	route := "StationPrediction.svc/json/GetPrediction"
	codes := strings.Join(stations, ",")
	url := fmt.Sprintf("%s/%s/%s", dmv.baseUrl, route, codes)

	client := http.Client{}
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Add("api_key", *dmv.apiKey)
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		logger.Debug(string(body))
		logger.Error(fmt.Sprintf("Failed to fetch. Received %d", resp.StatusCode))
		return nil, errors.New("Failed to fetch")
	}

	var timings TimingResponse
	err = json.Unmarshal(body, &timings)

	return timings.Trains, nil
}
