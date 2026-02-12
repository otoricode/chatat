package database

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog/log"
)

// RunMigrations applies all pending database migrations from the given path.
// It returns nil if there are no pending migrations.
func RunMigrations(databaseURL, migrationsPath string) error {
	m, err := migrate.New("file://"+migrationsPath, databaseURL)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			log.Error().Err(srcErr).Msg("failed to close migration source")
		}
		if dbErr != nil {
			log.Error().Err(dbErr).Msg("failed to close migration database")
		}
	}()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("run migrations: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("get migration version: %w", err)
	}

	log.Info().
		Uint("version", version).
		Bool("dirty", dirty).
		Msg("database migrations applied")

	return nil
}
