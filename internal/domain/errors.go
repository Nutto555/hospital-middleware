package domain

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrConflict           = errors.New("already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnknownHospital    = errors.New("unknown hospital")
	ErrPasswordTooLong    = errors.New("password must be at most 72 bytes")
	ErrHISUnavailable     = errors.New("hospital system unavailable")
)
