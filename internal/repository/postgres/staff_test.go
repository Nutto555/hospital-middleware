package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Nutto555/hospital-middleware/internal/domain"
	"github.com/Nutto555/hospital-middleware/internal/repository/postgres"
)

func TestStaffCreateAndFind(t *testing.T) {
	repo := postgres.NewStaffRepo(testPool(t))
	ctx := context.Background()

	created, err := repo.Create(ctx, domain.Staff{HospitalCode: "hospital-a", Username: "nurse1", PasswordHash: "hash"})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == "" || created.CreatedAt.IsZero() {
		t.Fatalf("id and created_at must be filled: %+v", created)
	}

	if _, err := repo.Create(ctx, domain.Staff{HospitalCode: "hospital-a", Username: "nurse1", PasswordHash: "x"}); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("duplicate in same hospital: err = %v, want ErrConflict", err)
	}
	if _, err := repo.Create(ctx, domain.Staff{HospitalCode: "hospital-b", Username: "nurse1", PasswordHash: "x"}); err != nil {
		t.Fatalf("same username in another hospital must be allowed: %v", err)
	}

	got, err := repo.FindByUsername(ctx, "hospital-a", "nurse1")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != created.ID || got.PasswordHash != "hash" || got.HospitalCode != "hospital-a" {
		t.Fatalf("got %+v", got)
	}
	if _, err := repo.FindByUsername(ctx, "hospital-a", "nobody"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("unknown user: err = %v, want ErrNotFound", err)
	}
	if _, err := repo.FindByUsername(ctx, "hospital-b", "nurse1"); err != nil {
		t.Fatalf("hospital-b nurse1 exists: %v", err)
	}

	if _, err := repo.Create(ctx, domain.Staff{HospitalCode: "hospital-a", Username: "nurse2", PasswordHash: "x"}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.FindByUsername(ctx, "hospital-b", "nurse2"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("nurse2 exists only in hospital-a: err = %v, want ErrNotFound", err)
	}
}
