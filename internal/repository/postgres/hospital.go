package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Nutto555/hospital-middleware/internal/domain"
)

// HospitalRepo reads the hospitals reference table.
type HospitalRepo struct {
	db *pgxpool.Pool
}

func NewHospitalRepo(db *pgxpool.Pool) *HospitalRepo {
	return &HospitalRepo{db: db}
}

// FindByCodeOrName matches a hospital by code or display name, ignoring case and surrounding spaces.
func (r *HospitalRepo) FindByCodeOrName(ctx context.Context, s string) (domain.Hospital, error) {
	var h domain.Hospital
	err := r.db.QueryRow(ctx,
		`SELECT code, name FROM hospitals WHERE lower(code) = lower($1) OR lower(name) = lower($1)`,
		strings.TrimSpace(s),
	).Scan(&h.Code, &h.Name)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Hospital{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Hospital{}, fmt.Errorf("find hospital: %w", err)
	}
	return h, nil
}
