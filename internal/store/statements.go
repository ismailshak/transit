package store

/*
	MIGRATIONS TABLE
*/

const CreateMigrationsTableSQL = `CREATE TABLE IF NOT EXISTS migrations (
	name TEXT NOT NULL,
	migrated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`

const CountMigrationsSQL = "SELECT COUNT(*) FROM migrations"

const SelectMigrationsSQL = "SELECT rowid, name, DATETIME(migrated_at, 'localtime') FROM migrations"

const InsertMigrationSQL = "INSERT INTO migrations (name) VALUES (?)"

/*
	AGENCIES TABLE
*/

const CreateAgenciesTableSQL = `CREATE TABLE agencies (
	agency_id TEXT NOT NULL,
	name TEXT NOT NULL,
	location REFERENCES locations(slug),
	timezone TEXT NOT NULL,
	language TEXT NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`

const InsertAgencySQL = "INSERT INTO agencies (agency_id, name, location, timezone, language) VALUES (?, ?, ?, ?, ?)"

const SelectAgenciesByLocationSQL = "SELECT rowid, * FROM agencies WHERE location = ?"

/*
	LOCATIONS TABLE
*/

// CreateLocationsTableSQL creates the locations table. An index will be created for `slug` due to 'UNIQUE' constraint
const CreateLocationsTableSQL = `CREATE TABLE locations (
	slug TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,
	supports_gtfs BOOLEAN NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`

const SelectLocationSQL = "SELECT rowid, * FROM locations WHERE slug = ?"

const SelectAllLocationsSQL = "SELECT rowid, * FROM locations"

const InsertLocationSQL = "INSERT INTO locations (slug, name, supports_gtfs) VALUES (?, ?, ?)"

/*
	STOPS TABLE
*/

const CreateStopsTableSQL = `CREATE TABLE stops (
	stop_id TEXT NOT NULL,
	name TEXT NOT NULL,
	location REFERENCES locations(slug),
	agency_id REFERENCES agencies(agency_id),
	latitude TEXT,
	longitude TEXT,
	type TEXT NOT NULL,
	parent_id TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`

const CreateStopLocationIndexSQL = "CREATE INDEX stop_location_index ON stops(location)"

const CountStopsByLocationSQL = "SELECT COUNT(*) FROM stops WHERE location = ?"

const SelectStopsByLocationSQL = "SELECT rowid, * FROM stops WHERE location = ?"

const SelectParentStopsByLocationSQL = `SELECT rowid, * FROM stops WHERE location = ? AND parent_id = ""`

const InsertStopSQL = "INSERT INTO stops (stop_id, name, location, agency_id, latitude, longitude, type, parent_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
