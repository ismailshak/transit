package helpers

// If a train isn't for passengers
func IsGhostTrain(line, destination string) bool {
	return line == "--" || destination == "No Passenger" || line == "No"
}
