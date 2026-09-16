package postgres_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Nutto555/hospital-middleware/internal/domain"
	"github.com/Nutto555/hospital-middleware/internal/repository/postgres"
)

func str(s string) *string { return &s }

func date(s string) *time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return &t
}

func somchai() domain.Patient {
	return domain.Patient{
		HospitalCode: "hospital-a", PatientHN: "000123",
		NationalID:  str("1234567890121"),
		FirstNameTH: str("สมชาย"), LastNameTH: str("ใจดี"),
		FirstNameEN: str("Somchai"), LastNameEN: str("Jaidee"),
		DateOfBirth: date("1985-03-14"), PhoneNumber: str("0812345678"),
		Email: str("somchai@example.com"), Gender: str("M"),
	}
}

func seedPatients(t *testing.T, repo *postgres.PatientRepo) {
	t.Helper()
	ctx := context.Background()
	for _, p := range []domain.Patient{
		somchai(),
		{HospitalCode: "hospital-a", PatientHN: "000125", PassportID: str("AB1234567"),
			FirstNameEN: str("John"), MiddleNameEN: str("Robert"), LastNameEN: str("Smith"),
			DateOfBirth: date("1978-07-21"), Gender: str("M")},
		{HospitalCode: "hospital-b", PatientHN: "B-0001", NationalID: str("1234567890121"),
			FirstNameEN: str("Somchai"), LastNameEN: str("Jaidee"), Gender: str("M")},
	} {
		if _, err := repo.Upsert(ctx, p); err != nil {
			t.Fatal(err)
		}
	}
}

func hns(ps []domain.Patient) []string {
	out := make([]string, 0, len(ps))
	for _, p := range ps {
		out = append(out, p.PatientHN)
	}
	return out
}

func TestPatientSearch(t *testing.T) {
	repo := postgres.NewPatientRepo(testPool(t))
	seedPatients(t, repo)
	ctx := context.Background()

	cases := []struct {
		name     string
		hospital string
		criteria domain.SearchCriteria
		want     []string
	}{
		{"no criteria lists the hospital in hn order", "hospital-a", domain.SearchCriteria{}, []string{"000123", "000125"}},
		{"national id", "hospital-a", domain.SearchCriteria{NationalID: str("1234567890121")}, []string{"000123"}},
		{"passport id", "hospital-a", domain.SearchCriteria{PassportID: str("AB1234567")}, []string{"000125"}},
		{"english first name", "hospital-a", domain.SearchCriteria{FirstName: str("somchai")}, []string{"000123"}},
		{"thai first name", "hospital-a", domain.SearchCriteria{FirstName: str("สมชาย")}, []string{"000123"}},
		{"last name case insensitive", "hospital-a", domain.SearchCriteria{LastName: str("SMITH")}, []string{"000125"}},
		{"date of birth", "hospital-a", domain.SearchCriteria{DateOfBirth: date("1985-03-14")}, []string{"000123"}},
		{"phone", "hospital-a", domain.SearchCriteria{PhoneNumber: str("0812345678")}, []string{"000123"}},
		{"email case insensitive", "hospital-a", domain.SearchCriteria{Email: str("Somchai@Example.com")}, []string{"000123"}},
		{"criteria are anded", "hospital-a", domain.SearchCriteria{FirstName: str("Somchai"), LastName: str("Smith")}, nil},
		{"english middle name", "hospital-a", domain.SearchCriteria{MiddleName: str("robert")}, []string{"000125"}},
		{"middle name absent does not match", "hospital-a", domain.SearchCriteria{MiddleName: str("x")}, nil},
		{"other hospital sees only its own", "hospital-b", domain.SearchCriteria{NationalID: str("1234567890121")}, []string{"B-0001"}},
		{"unknown hospital sees nothing", "hospital-z", domain.SearchCriteria{}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := repo.Search(ctx, tc.hospital, tc.criteria)
			if err != nil {
				t.Fatal(err)
			}
			if fmt.Sprint(hns(got)) != fmt.Sprint(tc.want) {
				t.Fatalf("got %v, want %v", hns(got), tc.want)
			}
		})
	}
}

func TestPatientSearchReturnsFullRecord(t *testing.T) {
	repo := postgres.NewPatientRepo(testPool(t))
	seedPatients(t, repo)
	got, err := repo.Search(context.Background(), "hospital-a", domain.SearchCriteria{NationalID: str("1234567890121")})
	if err != nil || len(got) != 1 {
		t.Fatalf("got %v, %v", got, err)
	}
	p := got[0]
	if p.ID == "" || p.HospitalCode != "hospital-a" || *p.NationalID != "1234567890121" ||
		*p.FirstNameTH != "สมชาย" || *p.LastNameTH != "ใจดี" ||
		*p.FirstNameEN != "Somchai" || *p.LastNameEN != "Jaidee" ||
		*p.PhoneNumber != "0812345678" || *p.Email != "somchai@example.com" ||
		p.DateOfBirth.Format("2006-01-02") != "1985-03-14" || *p.Gender != "M" ||
		p.PassportID != nil || p.MiddleNameEN != nil {
		t.Fatalf("record not round-tripped: %+v", p)
	}
}

func TestPatientSearchCap(t *testing.T) {
	repo := postgres.NewPatientRepo(testPool(t))
	ctx := context.Background()
	for i := 0; i < 101; i++ {
		if _, err := repo.Upsert(ctx, domain.Patient{HospitalCode: "hospital-a", PatientHN: fmt.Sprintf("%06d", i)}); err != nil {
			t.Fatal(err)
		}
	}
	got, err := repo.Search(ctx, "hospital-a", domain.SearchCriteria{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 100 {
		t.Fatalf("len = %d, want the cap of 100", len(got))
	}
}

func TestPatientUpsertUpdatesInPlace(t *testing.T) {
	repo := postgres.NewPatientRepo(testPool(t))
	ctx := context.Background()
	first, err := repo.Upsert(ctx, somchai())
	if err != nil {
		t.Fatal(err)
	}
	changed := somchai()
	changed.PhoneNumber = str("0899999999")
	second, err := repo.Upsert(ctx, changed)
	if err != nil {
		t.Fatal(err)
	}
	if second.ID != first.ID {
		t.Fatalf("upsert must keep the row: %s != %s", second.ID, first.ID)
	}
	got, err := repo.Search(ctx, "hospital-a", domain.SearchCriteria{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || *got[0].PhoneNumber != "0899999999" {
		t.Fatalf("got %+v", got)
	}
}
