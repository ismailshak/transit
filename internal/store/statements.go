package store

/*
	MIGRATIONS TABLE
*/

const createMigrationsTableSQL = `CREATE TABLE IF NOT EXISTS migrations (
	name TEXT NOT NULL,
	migrated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`

const countMigrationsSQL = "SELECT COUNT(*) FROM migrations"

const selectMigrationsSQL = "SELECT rowid, name, DATETIME(migrated_at, 'localtime') FROM migrations"

const insertMigrationSQL = "INSERT INTO migrations (name) VALUES (?)"

/*
	AGENCIES TABLE
*/

const createAgenciesTableSQL = `CREATE TABLE agencies (
	agency_id TEXT NOT NULL,
	name TEXT NOT NULL,
	location REFERENCES locations(slug),
	timezone TEXT NOT NULL,
	language TEXT NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`

const insertAgencySQL = "INSERT INTO agencies (agency_id, name, location, timezone, language) VALUES (?, ?, ?, ?, ?)"

const selectAgenciesByLocationSQL = "SELECT rowid, * FROM agencies WHERE location = ?"

/*
	LOCATIONS TABLE
*/

// createLocationsTableSQL creates the locations table. An index will be created for `slug` due to 'UNIQUE' constraint.
const createLocationsTableSQL = `CREATE TABLE locations (
	slug TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,
	supports_gtfs BOOLEAN NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`

const selectLocationSQL = "SELECT rowid, * FROM locations WHERE slug = ?"

const selectAllLocationsSQL = "SELECT rowid, * FROM locations"

const insertLocationSQL = "INSERT INTO locations (slug, name, supports_gtfs) VALUES (?, ?, ?)"

/*
	STOPS TABLE
*/

const createStopsTableSQL = `CREATE TABLE stops (
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

const createStopLocationIndexSQL = "CREATE INDEX stop_location_index ON stops(location)"

const countStopsByLocationSQL = "SELECT COUNT(*) FROM stops WHERE location = ?"

const selectStopsByLocationSQL = "SELECT rowid, * FROM stops WHERE location = ?"

const selectParentStopsByLocationSQL = `SELECT rowid, * FROM stops WHERE location = ? AND parent_id = ""`

const insertStopSQL = "INSERT INTO stops (stop_id, name, location, agency_id, latitude, longitude, type, parent_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
