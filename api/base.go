package api

const (
	DMV_BASE_URL = "https://api.wmata.com"
)

type Timing struct {
	Min             string
	LocationName    string
	Destination     string
	DestinationName string
	Line            string
}

type Api interface {
	ListTimings(station []string) ([]Timing, error)
}

// Build and return a client for the DMV Metro
func DmvClient(api_key *string) *DmvApi {
	return &DmvApi{
		api_key:  api_key,
		base_url: DMV_BASE_URL,
	}
}
