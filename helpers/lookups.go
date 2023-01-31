package helpers

// Map of real/full name to station codes
var dmvStationCodes = map[string][]string{
	"Farragut North":             {"A02"},
	"Tenleytown-AU":              {"A07"},
	"Twinbrook":                  {"A13"},
	"Wheaton":                    {"B10"},
	"Metro Center":               {"C01", "A01"},
	"Foggy Bottom-GWU":           {"C04"},
	"Rosslyn":                    {"C05"},
	"Pentagon":                   {"C07"},
	"Braddock Road":              {"C12"},
	"King St-Old Town":           {"C13"},
	"Smithsonian":                {"D02"},
	"L'Enfant Plaza":             {"D03", "F03"},
	"Eastern Market":             {"D06"},
	"Stadium-Armory":             {"D08"},
	"West Hyattsville":           {"E07"},
	"Gallery Pl-Chinatown":       {"F01", "B01"},
	"Southern Avenue":            {"F08"},
	"Court House":                {"K01"},
	"Greensboro":                 {"N03"},
	"Innovation Center":          {"N09"},
	"Rhode Island Ave-Brentwood": {"B04"},
	"Takoma":                     {"B07"},
	"Forest Glen":                {"B09"},
	"NoMa-Gallaudet U":           {"B35"},
	"McPherson Square":           {"C02"},
	"Farragut West":              {"C03"},
	"Ronald Reagan Washington National Airport": {"C10"},
	"Minnesota Ave": {"D09"},
	"U Street/African-Amer Civil War Memorial/Cardozo": {"E03"},
	"Waterfront":                            {"F04"},
	"Ballston-MU":                           {"K04"},
	"West Falls Church":                     {"K06"},
	"Spring Hill":                           {"N04"},
	"Herndon":                               {"N08"},
	"Arlington Cemetery":                    {"C06"},
	"Federal Triangle":                      {"D01"},
	"Federal Center SW":                     {"D04"},
	"Capitol South":                         {"D05"},
	"Potomac Ave":                           {"D07"},
	"Mt Vernon Sq 7th St-Convention Center": {"E01"},
	"Georgia Ave-Petworth":                  {"E05"},
	"Hyattsville Crossing":                  {"E08"},
	"Suitland":                              {"F10"},
	"Addison Road-Seat Pleasant":            {"G03"},
	"East Falls Church":                     {"K05"},
	"Tysons":                                {"N02"},
	"Dupont Circle":                         {"A03"},
	"Van Ness-UDC":                          {"A06"},
	"Medical Center":                        {"A10"},
	"Grosvenor-Strathmore":                  {"A11"},
	"Rockville":                             {"A14"},
	"Pentagon City":                         {"C08"},
	"Crystal City":                          {"C09"},
	"Deanwood":                              {"D10"},
	"Shaw-Howard U":                         {"E02"},
	"College Park-U of Md":                  {"E09"},
	"Congress Heights":                      {"F07"},
	"Benning Road":                          {"G01"},
	"Van Dorn Street":                       {"J02"},
	"Clarendon":                             {"K02"},
	"Virginia Square-GMU":                   {"K03"},
	"Cleveland Park":                        {"A05"},
	"Friendship Heights":                    {"A08"},
	"Union Station":                         {"B03"},
	"Silver Spring":                         {"B08"},
	"Greenbelt":                             {"E10"},
	"Navy Yard-Ballpark":                    {"F05"},
	"Morgan Boulevard":                      {"G04"},
	"McLean":                                {"N01"},
	"Reston Town Center":                    {"N07"},
	"Washington Dulles International Airport": {"N10"},
	"Loudon Gateway":                      {"N11"},
	"North Bethesda":                      {"A12"},
	"Columbia Heights":                    {"E04"},
	"Fort Totten":                         {"E06", "B06"},
	"Archives-Navy Memorial-Penn Quarter": {"F02"},
	"Naylor Road":                         {"F09"},
	"Capitol Heights":                     {"G02"},
	"Woodley Park-Zoo/Adams Morgan":       {"A04"},
	"Bethesda":                            {"A09"},
	"Judiciary Square":                    {"B02"},
	"Cheverly":                            {"D11"},
	"Anacostia":                           {"F06"},
	"Wiehle-Reston East":                  {"N06"},
	"Shady Grove":                         {"A15"},
	"Eisenhower Avenue":                   {"C14"},
	"Huntington":                          {"C15"},
	"Franconia-Springfield":               {"J03"},
	"Landover":                            {"D12"},
	"Branch Ave":                          {"F11"},
	"Vienna/Fairfax-GMU":                  {"K08"},
	"Dunn Loring-Merrifield":              {"K07"},
	"Downtown Largo":                      {"G05"},
	"Ashburn":                             {"N12"},
	"Brookland-CUA":                       {"B05"},
}

// Name has to match the keys of the station code look ups
func GetStationCodeFromName(name string) ([]string, bool) {
	codes, exists := dmvStationCodes[name]
	return codes, exists
}

func GetDmvStationNames() []string {
	return getLookupKeys(&dmvStationCodes)
}

func getLookupKeys(lookup *map[string][]string) []string {
	var keys []string
	for k := range *lookup {
		keys = append(keys, k)
	}

	return keys
}
