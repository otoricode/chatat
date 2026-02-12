package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/otoritech/chatat/internal/database"
	"github.com/otoritech/chatat/internal/testutil"
)

func TestRunMigrations_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	dbURL := getDBURL()
	migrationsPath := testutil.FindMigrationsPath()

	err := database.RunMigrations(dbURL, migrationsPath)
	assert.NoError(t, err)
}

func TestRunMigrations_InvalidPath(t *testing.T) {
	dbURL := getDBURL()
	err := database.RunMigrations(dbURL, "/nonexistent/path")
	assert.Error(t, err)
}
