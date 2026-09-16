package rest_test

import (
	"context"

	"github.com/Nutto555/hospital-middleware/internal/domain"
	"github.com/Nutto555/hospital-middleware/internal/service"
)

type fakeStaffService struct {
	createFn func(username, password, hospital string) (domain.Staff, error)
	loginFn  func(username, password, hospital string) (service.Session, error)
}

func (f fakeStaffService) Create(_ context.Context, username, password, hospital string) (domain.Staff, error) {
	return f.createFn(username, password, hospital)
}

func (f fakeStaffService) Login(_ context.Context, username, password, hospital string) (service.Session, error) {
	return f.loginFn(username, password, hospital)
}

type fakePatientService struct {
	searchFn func(hospitalCode string, c domain.SearchCriteria) ([]domain.Patient, error)
}

func (f fakePatientService) Search(_ context.Context, hospitalCode string, c domain.SearchCriteria) ([]domain.Patient, error) {
	return f.searchFn(hospitalCode, c)
}
