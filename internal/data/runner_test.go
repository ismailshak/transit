package data_test

import (
	"testing"

	"github.com/ismailshak/transit/internal/data"
	"github.com/ismailshak/transit/internal/testutils"
)

func TestCreatingMigrationTable(t *testing.T) {
	t.Parallel()

	db := testutils.BlankDB(t)

	err := data.CreateMigrationTable(db.DB)
	if err != nil {
		t.Errorf("Failed to create migrations table: %s", err)
	}
}

func TestSkippingMigrationsTableIfExists(t *testing.T) {
	t.Parallel()

	db := testutils.BlankDB(t)
	testutils.InitMigrationsTable(t, db.DB)

	err := data.CreateMigrationTable(db.DB)
	if err != nil {
		t.Errorf("Failed skipping migrations table: %s", err)
	}

	count, err := data.GetMigrationCount(db.DB)
	if err != nil {
		t.Errorf("Failed to get migration count: %s", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 migration. Got %d", count)
	}

	migrations, err := data.GetCurrentMigrations(db.DB, count)
	if err != nil {
		t.Errorf("Failed to get migrations: %s", err)
	}

	if migrations[0].ID != 1 {
		t.Errorf("Expected migration to have ID 1. Got %d", migrations[0].ID)
	}

	if migrations[0].Name != "1_FakeMigration" {
		t.Errorf("Expected migration name \"1_FakeMigration\". Got %q", migrations[0].Name)
	}

	if migrations[0].MigratedAt == "" {
		t.Errorf("Expected migration to have a valid date. Got %q", migrations[0].MigratedAt)
	}
}
