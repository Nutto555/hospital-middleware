package hospitala_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Nutto555/hospital-middleware/internal/his"
	"github.com/Nutto555/hospital-middleware/internal/his/hospitala"
)

const somchaiJSON = `{
	"first_name_th": "สมชาย", "middle_name_th": null, "last_name_th": "ใจดี",
	"first_name_en": "Somchai", "middle_name_en": null, "last_name_en": "Jaidee",
	"date_of_birth": "1985-03-14", "patient_hn": "000123",
	"national_id": "1234567890121", "passport_id": null,
	"phone_number": "0812345678", "email": "somchai@example.com", "gender": "M"
}`

func server(t *testing.T, handler http.HandlerFunc) *hospitala.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return hospitala.New(srv.URL+"/", srv.Client())
}

func TestSearchPatientFound(t *testing.T) {
	var gotPath string
	c := server(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(somchaiJSON))
	})
	p, err := c.SearchPatient(context.Background(), "1234567890121")
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/patient/search/1234567890121" {
		t.Fatalf("path = %s", gotPath)
	}
	if p.PatientHN != "000123" || *p.FirstNameTH != "สมชาย" || *p.LastNameEN != "Jaidee" || *p.Gender != "M" ||
		p.MiddleNameEN != nil || p.PassportID != nil || p.DateOfBirth.Format("2006-01-02") != "1985-03-14" {
		t.Fatalf("got %+v", p)
	}
}

func TestSearchPatientStatuses(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
		want   error
	}{
		{"not found", 404, `{"error":"patient not found"}`, his.ErrNotFound},
		{"server error", 500, `oops`, nil},
		{"bad json", 200, `{"patient_hn": 12}`, nil},
		{"no hn", 200, `{"first_name_en":"x"}`, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := server(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			})
			_, err := c.SearchPatient(context.Background(), "x")
			if err == nil {
				t.Fatal("want an error")
			}
			if tc.want != nil && !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
			if tc.want == nil && errors.Is(err, his.ErrNotFound) {
				t.Fatalf("err = %v must not be ErrNotFound", err)
			}
		})
	}
}

func TestSearchPatientTimeout(t *testing.T) {
	c := server(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
	})
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := c.SearchPatient(ctx, "x"); err == nil || errors.Is(err, his.ErrNotFound) {
		t.Fatalf("err = %v, want a transport error", err)
	}
}
