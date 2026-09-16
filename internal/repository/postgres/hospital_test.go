package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Nutto555/hospital-middleware/internal/domain"
	"github.com/Nutto555/hospital-middleware/internal/repository/postgres"
)

func TestHospitalFindByCodeOrName(t *testing.T) {
	repo := postgres.NewHospitalRepo(testPool(t))
	ctx := context.Background()

	for _, in := range []string{"hospital-a", "HOSPITAL-A", "Hospital A", "hospital a", "  hospital-a "} {
		h, err := repo.FindByCodeOrName(ctx, in)
		if err != nil {
			t.Errorf("%q: %v", in, err)
			continue
		}
		if h.Code != "hospital-a" || h.Name != "Hospital A" {
			t.Errorf("%q: got %+v", in, h)
		}
	}

	_, err := repo.FindByCodeOrName(ctx, "hospital-z")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("unknown hospital: err = %v, want ErrNotFound", err)
	}
}
