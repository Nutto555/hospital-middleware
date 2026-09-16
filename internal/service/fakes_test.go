package service_test

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Nutto555/hospital-middleware/internal/auth"
	"github.com/Nutto555/hospital-middleware/internal/domain"
	"github.com/Nutto555/hospital-middleware/internal/his"
)

type fakeHospitals struct{}

func (fakeHospitals) FindByCodeOrName(_ context.Context, s string) (domain.Hospital, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "hospital-a", "hospital a":
		return domain.Hospital{Code: "hospital-a", Name: "Hospital A"}, nil
	case "hospital-b", "hospital b":
		return domain.Hospital{Code: "hospital-b", Name: "Hospital B"}, nil
	}
	return domain.Hospital{}, domain.ErrNotFound
}

type fakeStaff struct {
	rows      []domain.Staff
	createErr error
}

func (f *fakeStaff) Create(_ context.Context, s domain.Staff) (domain.Staff, error) {
	if f.createErr != nil {
		return domain.Staff{}, f.createErr
	}
	for _, r := range f.rows {
		if r.HospitalCode == s.HospitalCode && r.Username == s.Username {
			return domain.Staff{}, domain.ErrConflict
		}
	}
	s.ID = fmt.Sprintf("staff-%d", len(f.rows)+1)
	s.CreatedAt = time.Now()
	f.rows = append(f.rows, s)
	return s, nil
}

func (f *fakeStaff) FindByUsername(_ context.Context, hospitalCode, username string) (domain.Staff, error) {
	for _, r := range f.rows {
		if r.HospitalCode == hospitalCode && r.Username == username {
			return r, nil
		}
	}
	return domain.Staff{}, domain.ErrNotFound
}

type fakeTokens struct{}

func (fakeTokens) Issue(p auth.Principal) (string, time.Time, error) {
	return "token:" + p.StaffID + "@" + p.HospitalCode, time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC), nil
}

type fakePatients struct {
	rows         []domain.Patient
	searchCalls  int
	lastHospital string
	upserted     []domain.Patient
}

func (f *fakePatients) Search(_ context.Context, hospitalCode string, c domain.SearchCriteria) ([]domain.Patient, error) {
	f.searchCalls++
	f.lastHospital = hospitalCode
	var out []domain.Patient
	for _, p := range f.rows {
		if p.HospitalCode == hospitalCode && matches(p, c) {
			out = append(out, p)
		}
	}
	return out, nil
}

func (f *fakePatients) Upsert(_ context.Context, p domain.Patient) (domain.Patient, error) {
	p.ID = "patient:" + p.PatientHN
	f.rows = append(f.rows, p)
	f.upserted = append(f.upserted, p)
	return p, nil
}

// matches mirrors the subset of the SQL rules the service tests rely on.
func matches(p domain.Patient, c domain.SearchCriteria) bool {
	if c.NationalID != nil && (p.NationalID == nil || *p.NationalID != *c.NationalID) {
		return false
	}
	if c.PassportID != nil && (p.PassportID == nil || *p.PassportID != *c.PassportID) {
		return false
	}
	if c.FirstName != nil && !(equalFold(p.FirstNameEN, c.FirstName) || equalFold(p.FirstNameTH, c.FirstName)) {
		return false
	}
	return true
}

func equalFold(a, b *string) bool {
	return a != nil && b != nil && strings.EqualFold(*a, *b)
}

type fakeHIS struct {
	patients map[string]his.Patient
	err      error
	calls    []string
}

func (f *fakeHIS) SearchPatient(_ context.Context, id string) (his.Patient, error) {
	f.calls = append(f.calls, id)
	if f.err != nil {
		return his.Patient{}, f.err
	}
	if p, ok := f.patients[id]; ok {
		return p, nil
	}
	return his.Patient{}, his.ErrNotFound
}

func str(s string) *string { return &s }
