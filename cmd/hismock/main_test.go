package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMockServesFixtures(t *testing.T) {
	patients, err := load(fixtures)
	if err != nil {
		t.Fatal(err)
	}
	if len(patients) != 3 {
		t.Fatalf("fixtures = %d, want 3", len(patients))
	}
	srv := httptest.NewServer(handler(patients))
	defer srv.Close()

	cases := []struct {
		name   string
		id     string
		status int
		hn     string
	}{
		{"national id", "1234567890121", 200, "000123"},
		{"second national id", "3101234567893", 200, "000124"},
		{"passport", "AB1234567", 200, "000125"},
		{"unknown", "0000000000000", 404, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Get(srv.URL + "/patient/search/" + tc.id)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != tc.status {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tc.status)
			}
			if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
				t.Fatalf("content-type = %s", ct)
			}
			var body map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if tc.status == 404 {
				if body["error"] != "patient not found" {
					t.Fatalf("body %v", body)
				}
				return
			}
			if body["patient_hn"] != tc.hn {
				t.Fatalf("patient_hn = %v, want %s", body["patient_hn"], tc.hn)
			}
			if _, isString := body["date_of_birth"].(string); !isString {
				t.Fatalf("date_of_birth must be a YYYY-MM-DD string, got %v", body["date_of_birth"])
			}
		})
	}
}

func TestMockRejectsOtherRoutes(t *testing.T) {
	srv := httptest.NewServer(handler(nil))
	defer srv.Close()
	resp, err := http.Post(srv.URL+"/patient/search/1", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", resp.StatusCode)
	}
}
