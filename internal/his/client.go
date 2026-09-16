// Package his describes what a hospital information system can answer and how the
// middleware reaches one per hospital.
package his

import (
	"context"
	"errors"
)

// Patient is the record a hospital information system returns, with the field names Hospital A uses.
type Patient struct {
	FirstNameTH  *string `json:"first_name_th"`
	MiddleNameTH *string `json:"middle_name_th"`
	LastNameTH   *string `json:"last_name_th"`
	FirstNameEN  *string `json:"first_name_en"`
	MiddleNameEN *string `json:"middle_name_en"`
	LastNameEN   *string `json:"last_name_en"`
	DateOfBirth  *Date   `json:"date_of_birth"`
	PatientHN    string  `json:"patient_hn"`
	NationalID   *string `json:"national_id"`
	PassportID   *string `json:"passport_id"`
	PhoneNumber  *string `json:"phone_number"`
	Email        *string `json:"email"`
	Gender       *string `json:"gender"`
}

// ErrNotFound means the hospital has no patient with that id.
var ErrNotFound = errors.New("his: patient not found")

// Client looks a patient up by national id or passport id.
type Client interface {
	SearchPatient(ctx context.Context, id string) (Patient, error)
}

// Registry maps a hospital code to its client. Hospitals without an entry have no HIS integration.
type Registry map[string]Client

// For returns the client for hospitalCode, if any.
func (r Registry) For(hospitalCode string) (Client, bool) {
	c, ok := r[hospitalCode]
	return c, ok
}
