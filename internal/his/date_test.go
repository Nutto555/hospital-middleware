package his

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDateJSON(t *testing.T) {
	var p Patient
	if err := json.Unmarshal([]byte(`{"patient_hn":"1","date_of_birth":"1985-03-14"}`), &p); err != nil {
		t.Fatal(err)
	}
	if p.DateOfBirth == nil || p.DateOfBirth.Format("2006-01-02") != "1985-03-14" {
		t.Fatalf("got %v", p.DateOfBirth)
	}
	out, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if want := `"date_of_birth":"1985-03-14"`; !strings.Contains(string(out), want) {
		t.Fatalf("%s does not contain %s", out, want)
	}

	var q Patient
	if err := json.Unmarshal([]byte(`{"patient_hn":"1","date_of_birth":null}`), &q); err != nil || q.DateOfBirth != nil {
		t.Fatalf("null must stay nil: %v %v", q.DateOfBirth, err)
	}
	if err := json.Unmarshal([]byte(`{"patient_hn":"1","date_of_birth":"14/03/1985"}`), &q); err == nil {
		t.Fatal("wrong layout must fail")
	}
}
