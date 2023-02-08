package helpers

// Returns a 256-color mode for a given metro line. A background
// and foreground tuple are returned (in that order)
func GetColorFromLine(line string) (string, string) {
	white, black := "#FFFFFF", "#000000"
	switch line {
	case "SV", "Silver":
		return "7", black
	case "RD", "Red":
		return "124", white
	case "BL", "Blue":
		return "21", white
	case "YL", "Yellow":
		return "11", black
	case "OR", "Orange":
		return "208", black
	case "GR", "Green":
		return "40", white
	default:
		return white, black
	}
}

// If a train isn't for passengers
func IsGhostTrain(line, destination string) bool {
	return line == "--" || destination == "No Passenger" || line == "No"
}
