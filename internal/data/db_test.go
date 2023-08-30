package data_test

import (
	"testing"

	"github.com/ismailshak/transit/internal/data"
	"github.com/ismailshak/transit/internal/testutils"
)

func TestMigrationCompletes(t *testing.T) {
	db := testutils.BlankDB(t)
	err := data.SyncMigrations(db)

	if err != nil {
		t.Errorf("Failed to sync migrations. %s", err)
	}
}
