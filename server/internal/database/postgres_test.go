package database_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/otoritech/chatat/internal/database"
)

func getDBURL() string {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return "postgres://chatat:chatat_dev@localhost:5433/chatat?sslmode=disable"
	}
	return url
}

func TestNewPool_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pool, err := database.NewPool(ctx, getDBURL())
	require.NoError(t, err)
	defer pool.Close()

	assert.NotNil(t, pool)

	var result int
	err = pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	require.NoError(t, err)
	assert.Equal(t, 1, result)
}

func TestNewPool_InvalidURL(t *testing.T) {
	ctx := context.Background()
	_, err := database.NewPool(ctx, "invalid://url")
	assert.Error(t, err)
}

func TestNewPool_ConcurrentQueries(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pool, err := database.NewPool(ctx, getDBURL())
	require.NoError(t, err)
	defer pool.Close()

	errs := make(chan error, 10)
	for i := 0; i < 10; i++ {
		go func() {
			var result int
			errs <- pool.QueryRow(ctx, "SELECT 1").Scan(&result)
		}()
	}

	for i := 0; i < 10; i++ {
		assert.NoError(t, <-errs)
	}
}
