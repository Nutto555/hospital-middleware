package postgres_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Nutto555/hospital-middleware/internal/repository/postgres"
)

// testPool returns a migrated, empty database or skips when TEST_DATABASE_URL is unset.
func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	if err := postgres.Migrate(url); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	pool, err := postgres.Connect(context.Background(), url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	if _, err := pool.Exec(context.Background(), `TRUNCATE patients, staff`); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return pool
}
