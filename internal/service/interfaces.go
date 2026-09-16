// Package service holds the business rules between the HTTP layer and storage.
package service

import (
	"context"
	"time"

	"github.com/Nutto555/hospital-middleware/internal/auth"
	"github.com/Nutto555/hospital-middleware/internal/domain"
)

// HospitalRepository resolves the hospitals the middleware serves.
type HospitalRepository interface {
	// FindByCodeOrName returns domain.ErrNotFound when nothing matches.
	FindByCodeOrName(ctx context.Context, s string) (domain.Hospital, error)
}

// StaffRepository stores staff accounts.
type StaffRepository interface {
	// Create returns domain.ErrConflict when the username exists in that hospital.
	Create(ctx context.Context, s domain.Staff) (domain.Staff, error)
	// FindByUsername returns domain.ErrNotFound when the hospital has no such user.
	FindByUsername(ctx context.Context, hospitalCode, username string) (domain.Staff, error)
}

// PatientRepository stores the local copy of patient records.
type PatientRepository interface {
	Search(ctx context.Context, hospitalCode string, c domain.SearchCriteria) ([]domain.Patient, error)
	Upsert(ctx context.Context, p domain.Patient) (domain.Patient, error)
}

// TokenIssuer signs a token for a logged-in staff member.
type TokenIssuer interface {
	Issue(p auth.Principal) (token string, expiresAt time.Time, err error)
}
