package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken covers every way a token can fail: missing, malformed, expired, wrong signature.
var ErrInvalidToken = errors.New("invalid or missing token")

// Principal is who a valid token speaks for.
type Principal struct {
	StaffID      string
	HospitalCode string
}

// JWT issues and verifies HS256 tokens.
type JWT struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewJWT(secret string, ttl time.Duration) *JWT {
	return &JWT{secret: []byte(secret), ttl: ttl, now: time.Now}
}

type claims struct {
	Hospital string `json:"hospital"`
	jwt.RegisteredClaims
}

// Issue signs a token for p that expires after the configured ttl.
func (j *JWT) Issue(p Principal) (string, time.Time, error) {
	now := j.now()
	expiresAt := now.Add(j.ttl)
	c := claims{
		Hospital: p.HospitalCode,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   p.StaffID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(j.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	return token, expiresAt, nil
}

// Verify parses and validates token and returns the principal it carries.
func (j *JWT) Verify(token string) (Principal, error) {
	var c claims
	_, err := jwt.ParseWithClaims(token, &c,
		func(*jwt.Token) (any, error) { return j.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(j.now),
	)
	if err != nil || c.Subject == "" || c.Hospital == "" {
		return Principal{}, ErrInvalidToken
	}
	return Principal{StaffID: c.Subject, HospitalCode: c.Hospital}, nil
}
