package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Nutto555/hospital-middleware/internal/domain"
)

// StaffRepo stores staff accounts.
type StaffRepo struct {
	db *pgxpool.Pool
}

func NewStaffRepo(db *pgxpool.Pool) *StaffRepo {
	return &StaffRepo{db: db}
}

// Create inserts a staff member and returns it with id and created_at filled.
// A username already taken in the same hospital yields domain.ErrConflict.
func (r *StaffRepo) Create(ctx context.Context, s domain.Staff) (domain.Staff, error) {
	err := r.db.QueryRow(ctx,
		`INSERT INTO staff (hospital_code, username, password_hash)
		 VALUES ($1, $2, $3)
		 RETURNING id::text, created_at`,
		s.HospitalCode, s.Username, s.PasswordHash,
	).Scan(&s.ID, &s.CreatedAt)
	if isUniqueViolation(err) {
		return domain.Staff{}, domain.ErrConflict
	}
	if err != nil {
		return domain.Staff{}, fmt.Errorf("create staff: %w", err)
	}
	return s, nil
}

// FindByUsername looks a staff member up inside one hospital.
func (r *StaffRepo) FindByUsername(ctx context.Context, hospitalCode, username string) (domain.Staff, error) {
	var s domain.Staff
	err := r.db.QueryRow(ctx,
		`SELECT id::text, hospital_code, username, password_hash, created_at
		 FROM staff WHERE hospital_code = $1 AND username = $2`,
		hospitalCode, username,
	).Scan(&s.ID, &s.HospitalCode, &s.Username, &s.PasswordHash, &s.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Staff{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Staff{}, fmt.Errorf("find staff: %w", err)
	}
	return s, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
