package rest_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Nutto555/hospital-middleware/internal/domain"
	"github.com/Nutto555/hospital-middleware/internal/rest"
	"github.com/Nutto555/hospital-middleware/internal/service"
)

func post(r http.Handler, path, body string, headers ...string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for i := 0; i+1 < len(headers); i += 2 {
		if headers[i+1] != "" {
			req.Header.Set(headers[i], headers[i+1])
		}
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestCreateStaff(t *testing.T) {
	svc := fakeStaffService{createFn: func(username, password, hospital string) (domain.Staff, error) {
		switch {
		case hospital == "hospital-z":
			return domain.Staff{}, domain.ErrUnknownHospital
		case username == "taken":
			return domain.Staff{}, domain.ErrConflict
		case len(password) > 72:
			return domain.Staff{}, domain.ErrPasswordTooLong
		case username == "boom":
			return domain.Staff{}, errors.New("connection lost")
		}
		return domain.Staff{ID: "id-1", HospitalCode: "hospital-a", Username: username, PasswordHash: "secret-hash"}, nil
	}}
	r := rest.NewRouter(rest.Deps{Staff: svc})

	cases := []struct {
		name     string
		body     string
		status   int
		contains string
	}{
		{"created", `{"username":"nurse.1","password":"correct horse","hospital":" hospital-a "}`, 201, `{"id":"id-1","username":"nurse.1","hospital":"hospital-a"}`},
		{"malformed json", `{"username":`, 400, "invalid request body"},
		{"username too short", `{"username":"ab","password":"correct horse","hospital":"hospital-a"}`, 400, "username"},
		{"username bad chars", `{"username":"nurse one","password":"correct horse","hospital":"hospital-a"}`, 400, "username"},
		{"password too short", `{"username":"nurse1","password":"short","hospital":"hospital-a"}`, 400, "password"},
		{"password too long", `{"username":"nurse1","password":"` + strings.Repeat("a", 73) + `","hospital":"hospital-a"}`, 400, "72 bytes"},
		{"hospital missing", `{"username":"nurse1","password":"correct horse"}`, 400, "hospital"},
		{"hospital unknown", `{"username":"nurse1","password":"correct horse","hospital":"hospital-z"}`, 400, "unknown hospital"},
		{"duplicate", `{"username":"taken","password":"correct horse","hospital":"hospital-a"}`, 409, "already exists"},
		{"unexpected error", `{"username":"boom","password":"correct horse","hospital":"hospital-a"}`, 500, "internal error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := post(r, "/staff/create", tc.body)
			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d; body %s", rec.Code, tc.status, rec.Body)
			}
			if !strings.Contains(rec.Body.String(), tc.contains) {
				t.Fatalf("body %s does not contain %q", rec.Body, tc.contains)
			}
			if strings.Contains(rec.Body.String(), "secret-hash") {
				t.Fatal("password hash leaked into the response")
			}
		})
	}
}

func TestCreateStaffTrimsHospital(t *testing.T) {
	var gotHospital string
	svc := fakeStaffService{createFn: func(username, password, hospital string) (domain.Staff, error) {
		gotHospital = hospital
		return domain.Staff{ID: "id-1", HospitalCode: "hospital-a", Username: username}, nil
	}}
	r := rest.NewRouter(rest.Deps{Staff: svc})
	rec := post(r, "/staff/create", `{"username":"nurse.1","password":"correct horse","hospital":" hospital-a "}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body %s", rec.Code, rec.Body)
	}
	if gotHospital != "hospital-a" {
		t.Fatalf("hospital passed to the service = %q, want it trimmed", gotHospital)
	}
}

func TestLogin(t *testing.T) {
	exp := time.Date(2026, 9, 17, 13, 0, 0, 0, time.UTC)
	svc := fakeStaffService{loginFn: func(username, password, hospital string) (service.Session, error) {
		if username == "nurse1" && password == "correct horse" && hospital == "hospital-a" {
			return service.Session{Token: "tok", ExpiresAt: exp}, nil
		}
		return service.Session{}, domain.ErrInvalidCredentials
	}}
	r := rest.NewRouter(rest.Deps{Staff: svc})

	cases := []struct {
		name     string
		body     string
		status   int
		contains string
	}{
		{"ok", `{"username":"nurse1","password":"correct horse","hospital":"hospital-a"}`, 200, `{"token":"tok","expires_at":"2026-09-17T13:00:00Z"}`},
		{"wrong password", `{"username":"nurse1","password":"nope","hospital":"hospital-a"}`, 401, "invalid credentials"},
		{"wrong hospital", `{"username":"nurse1","password":"correct horse","hospital":"hospital-b"}`, 401, "invalid credentials"},
		{"missing password", `{"username":"nurse1","hospital":"hospital-a"}`, 400, "password"},
		{"malformed json", `nope`, 400, "invalid request body"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := post(r, "/staff/login", tc.body)
			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d; body %s", rec.Code, tc.status, rec.Body)
			}
			if !strings.Contains(rec.Body.String(), tc.contains) {
				t.Fatalf("body %s does not contain %q", rec.Body, tc.contains)
			}
		})
	}
}
