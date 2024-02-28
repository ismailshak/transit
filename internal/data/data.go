// Package data implements functions that interact with all pieces of
// data used by transit CLI.
//
// This can be data stored in the SQLite database on a user's machine, or
// data downloaded from a server like GTFS
package data

// Static data, that doesn't change often, that we store in the database
type StaticData struct {
	Agencies []*Agency
	Stops    []*Stop
}
