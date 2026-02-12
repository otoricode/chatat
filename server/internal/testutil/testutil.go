package testutil

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/otoritech/chatat/internal/database"
)

// SetupTestDB creates a connection pool to the test database and runs migrations.
// The caller is responsible for closing the pool.
func SetupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dbURL := GetDBURL()
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = FindMigrationsPath()
	}

	err := database.RunMigrations(dbURL, migrationsPath)
	require.NoError(t, err, "run migrations")

	ctx := context.Background()
	pool, err := database.NewPool(ctx, dbURL)
	require.NoError(t, err, "create pool")

	return pool
}

// SetupTestPool creates a pool for use in TestMain (no *testing.T needed).
func SetupTestPool() (*pgxpool.Pool, error) {
	dbURL := GetDBURL()
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = FindMigrationsPath()
	}

	if err := database.RunMigrations(dbURL, migrationsPath); err != nil {
		return nil, err
	}

	ctx := context.Background()
	return database.NewPool(ctx, dbURL)
}

// GetDBURL returns the database URL from env or default.
func GetDBURL() string {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://chatat:chatat_dev@localhost:5433/chatat?sslmode=disable"
	}
	return dbURL
}

// CleanTables truncates all application tables for test isolation.
func CleanTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	_, err := pool.Exec(ctx, `TRUNCATE message_status, topic_message_status, document_entities, document_tags, document_history, blocks, document_signers, document_collaborators, topic_messages, topic_members, topics, messages, chat_members, chats, entities, documents, users CASCADE`)
	require.NoError(t, err, "clean tables")
}

// FindMigrationsPath returns the path to the migrations directory.
func FindMigrationsPath() string {
	candidates := []string{
		"migrations",
		"../migrations",
		"../../migrations",
		"../../../migrations",
		"server/migrations",
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return "migrations"
}
