package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestIssueAndVerify(t *testing.T) {
	j := NewJWT("secret", time.Hour)
	want := Principal{StaffID: "staff-1", HospitalCode: "hospital-a"}
	token, exp, err := j.Issue(want)
	if err != nil {
		t.Fatal(err)
	}
	if d := time.Until(exp); d < 59*time.Minute || d > 61*time.Minute {
		t.Errorf("expires in %v, want about an hour", d)
	}
	got, err := j.Verify(token)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestVerifyRejects(t *testing.T) {
	j := NewJWT("secret", time.Hour)
	good, _, err := j.Issue(Principal{StaffID: "staff-1", HospitalCode: "hospital-a"})
	if err != nil {
		t.Fatal(err)
	}
	otherSecret, _, err := NewJWT("other", time.Hour).Issue(Principal{StaffID: "staff-1", HospitalCode: "hospital-a"})
	if err != nil {
		t.Fatal(err)
	}
	stale := NewJWT("secret", time.Hour)
	stale.now = func() time.Time { return time.Now().Add(-2 * time.Hour) }
	expired, _, err := stale.Issue(Principal{StaffID: "staff-1", HospitalCode: "hospital-a"})
	if err != nil {
		t.Fatal(err)
	}
	noHospital, _, err := j.Issue(Principal{StaffID: "staff-1"})
	if err != nil {
		t.Fatal(err)
	}
	valid := claims{
		Hospital: "hospital-a",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "staff-1",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	hs512, err := jwt.NewWithClaims(jwt.SigningMethodHS512, valid).SignedString([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	algNone, err := jwt.NewWithClaims(jwt.SigningMethodNone, valid).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatal(err)
	}
	noExp := valid
	noExp.ExpiresAt = nil
	forever, err := jwt.NewWithClaims(jwt.SigningMethodHS256, noExp).SignedString([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}

	for name, token := range map[string]string{
		"empty":             "",
		"garbage":           "not.a.token",
		"other secret":      otherSecret,
		"expired":           expired,
		"missing hospital":  noHospital,
		"tampered":          good + "x",
		"hs512 same secret": hs512,
		"alg none":          algNone,
		"no exp":            forever,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := j.Verify(token); !errors.Is(err, ErrInvalidToken) {
				t.Fatalf("err = %v, want ErrInvalidToken", err)
			}
		})
	}
}

func TestVerifyUsesInjectedClock(t *testing.T) {
	j := NewJWT("secret", time.Hour)
	token, _, err := j.Issue(Principal{StaffID: "staff-1", HospitalCode: "hospital-a"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := j.Verify(token); err != nil {
		t.Fatalf("live token rejected on the real clock: %v", err)
	}
	future := NewJWT("secret", time.Hour)
	future.now = func() time.Time { return time.Now().Add(2 * time.Hour) }
	if _, err := future.Verify(token); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("err = %v, want ErrInvalidToken", err)
	}
}
