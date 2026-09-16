package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Nutto555/hospital-middleware/internal/auth"
	"github.com/Nutto555/hospital-middleware/internal/domain"
)

// maxPasswordBytes is the bcrypt input limit.
const maxPasswordBytes = 72

// dummyHash is compared against when the hospital or the user is unknown, so a failed login
// takes about as long as a wrong password and reveals nothing by timing.
var dummyHash = mustHash("hospital-middleware-dummy-password")

func mustHash(password string) string {
	h, err := auth.HashPassword(password)
	if err != nil {
		panic(err)
	}
	return h
}

// Staff creates accounts and logs staff in.
type Staff struct {
	hospitals HospitalRepository
	staff     StaffRepository
	tokens    TokenIssuer
}

func NewStaff(hospitals HospitalRepository, staff StaffRepository, tokens TokenIssuer) *Staff {
	return &Staff{hospitals: hospitals, staff: staff, tokens: tokens}
}

// Session is the result of a successful login.
type Session struct {
	Token     string
	ExpiresAt time.Time
}

// Create registers a staff member in the hospital named by code or display name.
func (s *Staff) Create(ctx context.Context, username, password, hospital string) (domain.Staff, error) {
	if len(password) > maxPasswordBytes {
		return domain.Staff{}, domain.ErrPasswordTooLong
	}
	h, err := s.hospitals.FindByCodeOrName(ctx, hospital)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.Staff{}, domain.ErrUnknownHospital
	}
	if err != nil {
		return domain.Staff{}, fmt.Errorf("find hospital: %w", err)
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return domain.Staff{}, err
	}
	return s.staff.Create(ctx, domain.Staff{HospitalCode: h.Code, Username: username, PasswordHash: hash})
}

// Login checks the credentials against the hospital's staff and returns a signed session.
// Every failure is domain.ErrInvalidCredentials.
func (s *Staff) Login(ctx context.Context, username, password, hospital string) (Session, error) {
	h, err := s.hospitals.FindByCodeOrName(ctx, hospital)
	if errors.Is(err, domain.ErrNotFound) {
		auth.CheckPassword(dummyHash, password)
		return Session{}, domain.ErrInvalidCredentials
	}
	if err != nil {
		return Session{}, fmt.Errorf("find hospital: %w", err)
	}
	st, err := s.staff.FindByUsername(ctx, h.Code, username)
	if errors.Is(err, domain.ErrNotFound) {
		auth.CheckPassword(dummyHash, password)
		return Session{}, domain.ErrInvalidCredentials
	}
	if err != nil {
		return Session{}, fmt.Errorf("find staff: %w", err)
	}
	if !auth.CheckPassword(st.PasswordHash, password) {
		return Session{}, domain.ErrInvalidCredentials
	}
	token, expiresAt, err := s.tokens.Issue(auth.Principal{StaffID: st.ID, HospitalCode: st.HospitalCode})
	if err != nil {
		return Session{}, fmt.Errorf("issue token: %w", err)
	}
	return Session{Token: token, ExpiresAt: expiresAt}, nil
}
