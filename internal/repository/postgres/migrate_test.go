package postgres_test

import (
	"context"
	"testing"
)

func TestMigrateSeedsHospitalsAndIsIdempotent(t *testing.T) {
	pool := testPool(t) // runs Migrate once
	var n int
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM hospitals`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("hospitals = %d, want 2", n)
	}
	testPool(t) // second Migrate must be a no-op, not an error
}
