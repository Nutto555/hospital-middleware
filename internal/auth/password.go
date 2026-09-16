// Package auth hashes passwords and issues and verifies the bearer tokens staff log in with.
package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword returns the bcrypt hash of password at the default cost.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// CheckPassword reports whether password produced hash.
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
