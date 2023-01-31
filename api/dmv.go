package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// API to interact with DMV Metro
type DmvApi struct {
	api_key  *string
	base_url string
}

// API response for fetching train arrival information
type TimingResponse struct {
	Trains []Timing
}

// Fetch latest train arrival information
func (dmv DmvApi) ListTimings(stations []string) ([]Timing, error) {
	route := "StationPrediction.svc/json/GetPrediction"
	codes := strings.Join(stations, ",")
	url := fmt.Sprintf("%s/%s/%s", dmv.base_url, route, codes)

	client := http.Client{}
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Add("api_key", *dmv.api_key)
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Non 200 received %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)

	var timings TimingResponse
	err = json.Unmarshal(body, &timings)

	return timings.Trains, nil
}
