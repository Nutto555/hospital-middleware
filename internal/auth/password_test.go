package auth

import (
	"strings"
	"testing"
)

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("correct horse battery")
	if err != nil {
		t.Fatal(err)
	}
	if hash == "correct horse battery" || !strings.HasPrefix(hash, "$2") {
		t.Fatalf("not a bcrypt hash: %s", hash)
	}
	if !CheckPassword(hash, "correct horse battery") {
		t.Error("right password rejected")
	}
	if CheckPassword(hash, "wrong") {
		t.Error("wrong password accepted")
	}
	if CheckPassword("not-a-hash", "correct horse battery") {
		t.Error("garbage hash accepted")
	}
}
