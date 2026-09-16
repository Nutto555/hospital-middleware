package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Nutto555/hospital-middleware/internal/auth"
	"github.com/Nutto555/hospital-middleware/internal/domain"
	"github.com/Nutto555/hospital-middleware/internal/service"
)

func TestStaffCreate(t *testing.T) {
	ctx := context.Background()
	repo := &fakeStaff{}
	svc := service.NewStaff(fakeHospitals{}, repo, fakeTokens{})

	created, err := svc.Create(ctx, "nurse1", "correct horse battery", "Hospital A")
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == "" || created.HospitalCode != "hospital-a" || created.Username != "nurse1" {
		t.Fatalf("got %+v", created)
	}
	if created.PasswordHash == "correct horse battery" || !auth.CheckPassword(created.PasswordHash, "correct horse battery") {
		t.Fatal("password must be stored as a bcrypt hash of the input")
	}

	_, err = svc.Create(ctx, "nurse1", "another password", "hospital-a")
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("duplicate: err = %v, want ErrConflict", err)
	}
	if _, err := svc.Create(ctx, "nurse1", "another password", "hospital-b"); err != nil {
		t.Fatalf("same username in another hospital: %v", err)
	}
	_, err = svc.Create(ctx, "nurse2", "another password", "hospital-z")
	if !errors.Is(err, domain.ErrUnknownHospital) {
		t.Fatalf("unknown hospital: err = %v, want ErrUnknownHospital", err)
	}
	_, err = svc.Create(ctx, "nurse3", strings.Repeat("a", 73), "hospital-a")
	if !errors.Is(err, domain.ErrPasswordTooLong) {
		t.Fatalf("73 bytes: err = %v, want ErrPasswordTooLong", err)
	}
	if _, err := svc.Create(ctx, "nurse4", strings.Repeat("a", 72), "hospital-a"); err != nil {
		t.Fatalf("72 bytes must be accepted: %v", err)
	}
}

func TestStaffLogin(t *testing.T) {
	ctx := context.Background()
	repo := &fakeStaff{}
	svc := service.NewStaff(fakeHospitals{}, repo, fakeTokens{})
	if _, err := svc.Create(ctx, "nurse1", "correct horse battery", "hospital-a"); err != nil {
		t.Fatal(err)
	}

	sess, err := svc.Login(ctx, "nurse1", "correct horse battery", "hospital-a")
	if err != nil {
		t.Fatal(err)
	}
	if sess.Token != "token:staff-1@hospital-a" || sess.ExpiresAt.IsZero() {
		t.Fatalf("got %+v", sess)
	}

	cases := map[string][3]string{
		"wrong password":   {"nurse1", "wrong", "hospital-a"},
		"unknown user":     {"nobody", "correct horse battery", "hospital-a"},
		"wrong hospital":   {"nurse1", "correct horse battery", "hospital-b"},
		"unknown hospital": {"nurse1", "correct horse battery", "hospital-z"},
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := svc.Login(ctx, in[0], in[1], in[2])
			if !errors.Is(err, domain.ErrInvalidCredentials) {
				t.Fatalf("err = %v, want ErrInvalidCredentials", err)
			}
		})
	}
}

func TestStaffCreatePassesThroughRepositoryErrors(t *testing.T) {
	boom := errors.New("connection lost")
	svc := service.NewStaff(fakeHospitals{}, &fakeStaff{createErr: boom}, fakeTokens{})
	_, err := svc.Create(context.Background(), "nurse1", "correct horse battery", "hospital-a")
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want the repository error", err)
	}
}
