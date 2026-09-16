package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Nutto555/hospital-middleware/internal/domain"
	"github.com/Nutto555/hospital-middleware/internal/his"
)

// Patient searches the local copy of a hospital's patients and fills it from the hospital's
// information system when a lookup by id misses.
type Patient struct {
	patients PatientRepository
	his      his.Registry
}

func NewPatient(patients PatientRepository, registry his.Registry) *Patient {
	return &Patient{patients: patients, his: registry}
}

// Search returns the hospital's patients matching c. When nothing matches locally and c names
// a national id or passport id, the hospital's HIS is asked; a hit is stored and the search is
// run again so every criterion is still applied by the same SQL.
func (s *Patient) Search(ctx context.Context, hospitalCode string, c domain.SearchCriteria) ([]domain.Patient, error) {
	found, err := s.patients.Search(ctx, hospitalCode, c)
	if err != nil {
		return nil, err
	}
	if len(found) > 0 || !c.HasID() {
		return found, nil
	}
	client, ok := s.his.For(hospitalCode)
	if !ok {
		return found, nil
	}
	id := c.PassportID
	if c.NationalID != nil {
		id = c.NationalID
	}
	p, err := client.SearchPatient(ctx, *id)
	if errors.Is(err, his.ErrNotFound) {
		return found, nil
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrHISUnavailable, err)
	}
	if _, err := s.patients.Upsert(ctx, fromHIS(hospitalCode, p)); err != nil {
		return nil, fmt.Errorf("store patient: %w", err)
	}
	return s.patients.Search(ctx, hospitalCode, c)
}

func fromHIS(hospitalCode string, p his.Patient) domain.Patient {
	d := domain.Patient{
		HospitalCode: hospitalCode,
		PatientHN:    p.PatientHN,
		NationalID:   p.NationalID,
		PassportID:   p.PassportID,
		FirstNameTH:  p.FirstNameTH,
		MiddleNameTH: p.MiddleNameTH,
		LastNameTH:   p.LastNameTH,
		FirstNameEN:  p.FirstNameEN,
		MiddleNameEN: p.MiddleNameEN,
		LastNameEN:   p.LastNameEN,
		PhoneNumber:  p.PhoneNumber,
		Email:        p.Email,
		Gender:       p.Gender,
	}
	if p.DateOfBirth != nil {
		t := p.DateOfBirth.Time
		d.DateOfBirth = &t
	}
	return d
}
