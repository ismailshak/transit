package data

/*
	MIGRATIONS TABLE
*/

const CREATE_MIGRATIONS_TABLE = `CREATE TABLE IF NOT EXISTS migrations (
	name TEXT NOT NULL,
	migrated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`

const COUNT_MIGRATIONS = "SELECT COUNT(*) FROM migrations"

const SELECT_MIGRATIONS = "SELECT rowid, name, DATETIME(migrated_at, 'localtime') FROM migrations"

const INSERT_MIGRATION = "INSERT INTO migrations (name) VALUES (?)"

/*
	AGENCIES TABLE
*/

const CREATE_AGENCIES_TABLE = `CREATE TABLE agencies (
	agency_id TEXT NOT NULL,
	name TEXT NOT NULL,
	location REFERENCES locations(slug),
	timezone TEXT NOT NULL,
	language TEXT NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`

const INSERT_AGENCY = "INSERT INTO agencies (agency_id, name, location, timezone, language) VALUES (?, ?, ?, ?, ?)"

const SELECT_AGENCIES_BY_LOCATION = "SELECT rowid, * FROM agencies WHERE location = ?"

/*
	LOCATIONS TABLE
*/

// An index will be created for `slug` due to 'UNIQUE' constraint
const CREATE_LOCATIONS_TABLE = `CREATE TABLE locations (
	slug TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,
	supports_gtfs BOOLEAN NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`

const SELECT_LOCATION = "SELECT rowid, * FROM locations WHERE slug = ?"

const SELECT_ALL_LOCATIONS = "SELECT rowid, * FROM locations"

const INSERT_LOCATION = "INSERT INTO locations (slug, name, supports_gtfs) VALUES (?, ?, ?)"

/*
	STOPS TABLE
*/

const CREATE_STOPS_TABLE = `CREATE TABLE stops (
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

const CREATE_STOP_LOCATION_INDEX = "CREATE INDEX stop_location_index ON stops(location)"

const COUNT_STOPS_BY_LOCATION = "SELECT COUNT(*) FROM stops WHERE location = ?"

const SELECT_STOPS_BY_LOCATION = "SELECT rowid, * FROM stops WHERE location = ?"

const SELECT_PARENT_STOPS_BY_LOCATION = `SELECT rowid, * FROM stops WHERE location = ? AND parent_id = ""`

const INSERT_STOP = "INSERT INTO stops (stop_id, name, location, agency_id, latitude, longitude, type, parent_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
