package rest_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Nutto555/hospital-middleware/internal/auth"
	"github.com/Nutto555/hospital-middleware/internal/domain"
	"github.com/Nutto555/hospital-middleware/internal/rest"
)

func str(s string) *string { return &s }

func TestSearchPatients(t *testing.T) {
	tokens := auth.NewJWT("secret", time.Hour)
	token, _, err := tokens.Issue(auth.Principal{StaffID: "staff-1", HospitalCode: "hospital-a"})
	if err != nil {
		t.Fatal(err)
	}
	dob := time.Date(1985, 3, 14, 0, 0, 0, 0, time.UTC)
	somchai := domain.Patient{ID: "p1", HospitalCode: "hospital-a", PatientHN: "000123", NationalID: str("1234567890121"),
		FirstNameTH: str("สมชาย"), LastNameTH: str("ใจดี"), FirstNameEN: str("Somchai"), LastNameEN: str("Jaidee"),
		DateOfBirth: &dob, PhoneNumber: str("0812345678"), Email: str("somchai@example.com"), Gender: str("M")}

	var gotHospital string
	var gotCriteria domain.SearchCriteria
	svc := fakePatientService{searchFn: func(hospitalCode string, c domain.SearchCriteria) ([]domain.Patient, error) {
		gotHospital, gotCriteria = hospitalCode, c
		switch {
		case c.NationalID != nil && *c.NationalID == "1234567890121":
			return []domain.Patient{somchai}, nil
		case c.NationalID != nil && *c.NationalID == "down":
			return nil, domain.ErrHISUnavailable
		}
		return nil, nil
	}}
	r := rest.NewRouter(rest.Deps{Patients: svc, Tokens: tokens})
	bearer := "Bearer " + token

	t.Run("found", func(t *testing.T) {
		rec := post(r, "/patient/search", `{"national_id":"1234567890121"}`, "Authorization", bearer)
		if rec.Code != 200 {
			t.Fatalf("status = %d body %s", rec.Code, rec.Body)
		}
		var body struct {
			Patients []map[string]any `json:"patients"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if len(body.Patients) != 1 {
			t.Fatalf("body %s", rec.Body)
		}
		p := body.Patients[0]
		if p["id"] != "p1" || p["hospital"] != "hospital-a" || p["patient_hn"] != "000123" || p["date_of_birth"] != "1985-03-14" ||
			p["first_name_th"] != "สมชาย" || p["gender"] != "M" || p["email"] != "somchai@example.com" {
			t.Fatalf("body %s", rec.Body)
		}
		for _, key := range []string{"middle_name_th", "middle_name_en", "passport_id"} {
			if v, ok := p[key]; !ok || v != nil {
				t.Fatalf("%s must be present and null, body %s", key, rec.Body)
			}
		}
		if gotHospital != "hospital-a" {
			t.Fatalf("hospital passed to the service = %q, want the token's", gotHospital)
		}
	})

	t.Run("empty result is an empty array", func(t *testing.T) {
		rec := post(r, "/patient/search", `{"first_name":"nobody"}`, "Authorization", bearer)
		if rec.Code != 200 || rec.Body.String() != `{"patients":[]}` {
			t.Fatalf("got %d %s", rec.Code, rec.Body)
		}
	})

	t.Run("empty body means no criteria", func(t *testing.T) {
		rec := post(r, "/patient/search", ``, "Authorization", bearer)
		if rec.Code != 200 || gotCriteria != (domain.SearchCriteria{}) {
			t.Fatalf("got %d %s, criteria %+v", rec.Code, rec.Body, gotCriteria)
		}
	})

	t.Run("blank fields are absent and values are trimmed", func(t *testing.T) {
		rec := post(r, "/patient/search", `{"national_id":" 1234567890121 ","first_name":"   ","email":"","date_of_birth":"1985-03-14"}`, "Authorization", bearer)
		if rec.Code != 200 {
			t.Fatalf("got %d %s", rec.Code, rec.Body)
		}
		if *gotCriteria.NationalID != "1234567890121" || gotCriteria.FirstName != nil || gotCriteria.Email != nil ||
			gotCriteria.DateOfBirth == nil || !gotCriteria.DateOfBirth.Equal(dob) {
			t.Fatalf("criteria %+v", gotCriteria)
		}
	})

	cases := []struct {
		name     string
		body     string
		header   string
		status   int
		contains string
	}{
		{"no token", `{}`, "", 401, "invalid or missing token"},
		{"bad token", `{}`, "Bearer nope", 401, "invalid or missing token"},
		{"malformed json", `{`, bearer, 400, "invalid request body"},
		{"bad date", `{"date_of_birth":"14/03/1985"}`, bearer, 400, "date_of_birth"},
		{"bad email", `{"email":"not-an-email"}`, bearer, 400, "email"},
		{"his down", `{"national_id":"down"}`, bearer, 502, "hospital system unavailable"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := post(r, "/patient/search", tc.body, "Authorization", tc.header)
			if rec.Code != tc.status || !strings.Contains(rec.Body.String(), tc.contains) {
				t.Fatalf("got %d %s, want %d containing %q", rec.Code, rec.Body, tc.status, tc.contains)
			}
		})
	}
}
