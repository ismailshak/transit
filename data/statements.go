package data

/*
	MIGRATIONS TABLE
*/

const CREATE_MIGRATIONS_TABLE = `CREATE TABLE IF NOT EXISTS migrations (
	name TEXT NOT NULL,
	migrated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`

const MIGRATIONS_COUNT = "SELECT COUNT(*) FROM migrations"

const SELECT_MIGRATIONS = "SELECT rowid, name, DATETIME(migrated_at, 'localtime') FROM migrations"

const INSERT_MIGRATION = "INSERT INTO migrations (name) VALUES (?)"

/*
	LOCATIONS TABLE
*/

// An index will be created for slug due to 'UNIQUE'
const CREATE_LOCATIONS_TABLE = `CREATE TABLE locations (
	slug TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,
	supports_gtfs BOOLEAN NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`

const INSERT_LOCATION = "INSERT INTO locations (slug, name, supports_gtfs) VALUES (?, ?, ?)"

/*
	STOPS TABLE
*/

const CREATE_STOPS_TABLE = `CREATE TABLE stops (
	id TEXT NOT NULL,
	name TEXT NOT NULL,
	location REFERENCES locations(slug)
)`

const CREATE_STOP_LOCATION_INDEX = "CREATE INDEX stop_location_index ON stops(location)"
