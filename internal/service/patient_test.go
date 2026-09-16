package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Nutto555/hospital-middleware/internal/domain"
	"github.com/Nutto555/hospital-middleware/internal/his"
	"github.com/Nutto555/hospital-middleware/internal/service"
)

func hisSomchai() his.Patient {
	return his.Patient{PatientHN: "000123", NationalID: str("1234567890121"),
		FirstNameTH: str("สมชาย"), LastNameTH: str("ใจดี"), FirstNameEN: str("Somchai"), LastNameEN: str("Jaidee"), Gender: str("M")}
}

func TestPatientSearch(t *testing.T) {
	ctx := context.Background()
	local := domain.Patient{ID: "p1", HospitalCode: "hospital-a", PatientHN: "000001", NationalID: str("1111111111111"), FirstNameEN: str("Local")}

	cases := []struct {
		name         string
		rows         []domain.Patient
		hisPatients  map[string]his.Patient
		hisErr       error
		noHIS        bool
		hospital     string
		criteria     domain.SearchCriteria
		wantHNs      []string
		wantHISCalls []string
		wantUpserts  int
		wantErr      error
	}{
		{name: "local hit skips the his", rows: []domain.Patient{local}, hospital: "hospital-a",
			criteria: domain.SearchCriteria{NationalID: str("1111111111111")}, wantHNs: []string{"000001"}},
		{name: "no id means no his call", hospital: "hospital-a",
			criteria: domain.SearchCriteria{FirstName: str("Somchai")}, wantHNs: nil},
		{name: "hospital without his stays local", noHIS: true, hospital: "hospital-b",
			criteria: domain.SearchCriteria{NationalID: str("1234567890121")}, wantHNs: nil},
		{name: "his hit is stored and returned", hisPatients: map[string]his.Patient{"1234567890121": hisSomchai()}, hospital: "hospital-a",
			criteria: domain.SearchCriteria{NationalID: str("1234567890121")}, wantHNs: []string{"000123"}, wantHISCalls: []string{"1234567890121"}, wantUpserts: 1},
		{name: "passport is used when there is no national id", hisPatients: map[string]his.Patient{"AB1234567": hisSomchai()}, hospital: "hospital-a",
			criteria: domain.SearchCriteria{PassportID: str("AB1234567")}, wantHNs: nil, wantHISCalls: []string{"AB1234567"}, wantUpserts: 1},
		{name: "national id wins over passport", hisPatients: map[string]his.Patient{"1234567890121": hisSomchai()}, hospital: "hospital-a",
			criteria: domain.SearchCriteria{NationalID: str("1234567890121"), PassportID: str("ZZ")}, wantHNs: nil, wantHISCalls: []string{"1234567890121"}, wantUpserts: 1},
		{name: "other criteria still apply after the his hit", hisPatients: map[string]his.Patient{"1234567890121": hisSomchai()}, hospital: "hospital-a",
			criteria: domain.SearchCriteria{NationalID: str("1234567890121"), FirstName: str("Nobody")}, wantHNs: nil, wantHISCalls: []string{"1234567890121"}, wantUpserts: 1},
		{name: "his not found is an empty result", hospital: "hospital-a",
			criteria: domain.SearchCriteria{NationalID: str("9999999999999")}, wantHNs: nil, wantHISCalls: []string{"9999999999999"}},
		{name: "his failure is reported", hisErr: errors.New("timeout"), hospital: "hospital-a",
			criteria: domain.SearchCriteria{NationalID: str("1234567890121")}, wantErr: domain.ErrHISUnavailable, wantHISCalls: []string{"1234567890121"}},
		{name: "another hospital cannot see the row", rows: []domain.Patient{local}, noHIS: true, hospital: "hospital-b",
			criteria: domain.SearchCriteria{NationalID: str("1111111111111")}, wantHNs: nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakePatients{rows: tc.rows}
			hisClient := &fakeHIS{patients: tc.hisPatients, err: tc.hisErr}
			registry := his.Registry{"hospital-a": hisClient}
			if tc.noHIS {
				registry = his.Registry{}
			}
			svc := service.NewPatient(repo, registry)

			got, err := svc.Search(ctx, tc.hospital, tc.criteria)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("err = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var hns []string
			for _, p := range got {
				hns = append(hns, p.PatientHN)
			}
			if !equal(hns, tc.wantHNs) {
				t.Errorf("hns = %v, want %v", hns, tc.wantHNs)
			}
			if !equal(hisClient.calls, tc.wantHISCalls) {
				t.Errorf("his calls = %v, want %v", hisClient.calls, tc.wantHISCalls)
			}
			if len(repo.upserted) != tc.wantUpserts {
				t.Errorf("upserts = %d, want %d", len(repo.upserted), tc.wantUpserts)
			}
			if repo.lastHospital != tc.hospital {
				t.Errorf("repository searched %q, want %q", repo.lastHospital, tc.hospital)
			}
			for _, u := range repo.upserted {
				if u.HospitalCode != tc.hospital {
					t.Errorf("stored under %q, want %q", u.HospitalCode, tc.hospital)
				}
			}
		})
	}
}

func TestPatientSearchStoresEveryField(t *testing.T) {
	repo := &fakePatients{}
	src := hisSomchai()
	src.DateOfBirth = &his.Date{}
	if err := src.DateOfBirth.UnmarshalJSON([]byte(`"1985-03-14"`)); err != nil {
		t.Fatal(err)
	}
	src.PhoneNumber, src.Email = str("0812345678"), str("somchai@example.com")
	svc := service.NewPatient(repo, his.Registry{"hospital-a": &fakeHIS{patients: map[string]his.Patient{"1234567890121": src}}})
	if _, err := svc.Search(context.Background(), "hospital-a", domain.SearchCriteria{NationalID: str("1234567890121")}); err != nil {
		t.Fatal(err)
	}
	got := repo.upserted[0]
	if got.HospitalCode != "hospital-a" || got.PatientHN != "000123" || *got.NationalID != "1234567890121" ||
		*got.FirstNameTH != "สมชาย" || *got.LastNameTH != "ใจดี" || *got.FirstNameEN != "Somchai" || *got.LastNameEN != "Jaidee" ||
		got.DateOfBirth.Format("2006-01-02") != "1985-03-14" || *got.PhoneNumber != "0812345678" ||
		*got.Email != "somchai@example.com" || *got.Gender != "M" || got.MiddleNameEN != nil || got.PassportID != nil {
		t.Fatalf("got %+v", got)
	}
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
