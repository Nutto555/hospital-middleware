package domain

import "time"

// Patient is the local copy of a hospital's patient record. Optional fields are nil when the
// hospital did not supply them.
type Patient struct {
	ID           string
	HospitalCode string
	PatientHN    string
	NationalID   *string
	PassportID   *string
	FirstNameTH  *string
	MiddleNameTH *string
	LastNameTH   *string
	FirstNameEN  *string
	MiddleNameEN *string
	LastNameEN   *string
	DateOfBirth  *time.Time
	PhoneNumber  *string
	Email        *string
	Gender       *string
}

// SearchCriteria are the optional patient filters. A nil field does not filter.
type SearchCriteria struct {
	NationalID  *string
	PassportID  *string
	FirstName   *string
	MiddleName  *string
	LastName    *string
	DateOfBirth *time.Time
	PhoneNumber *string
	Email       *string
}

// HasID reports whether the criteria identify a patient by national id or passport id,
// which is the only way a hospital information system can be asked for one.
func (c SearchCriteria) HasID() bool {
	return c.NationalID != nil || c.PassportID != nil
}
