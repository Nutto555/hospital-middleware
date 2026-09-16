package domain

import "time"

// Staff is a hospital staff member with login credentials.
type Staff struct {
	ID           string
	HospitalCode string
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}
