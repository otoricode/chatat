package repository_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/otoritech/chatat/internal/testutil"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	pool, err := testutil.SetupTestPool()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to setup test pool: %v\n", err)
		os.Exit(1)
	}
	testPool = pool

	code := m.Run()

	testPool.Close()
	os.Exit(code)
}
