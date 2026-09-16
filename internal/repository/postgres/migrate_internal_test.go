package postgres

import "testing"

func TestPgx5URL(t *testing.T) {
	cases := map[string]string{
		"postgres://u:p@h:5432/db?sslmode=disable": "pgx5://u:p@h:5432/db?sslmode=disable",
		"postgresql://u:p@h:5432/db":               "pgx5://u:p@h:5432/db",
		"pgx5://u:p@h:5432/db":                     "pgx5://u:p@h:5432/db",
		"":                                         "",
	}
	for in, want := range cases {
		if got := pgx5URL(in); got != want {
			t.Errorf("pgx5URL(%q) = %q, want %q", in, got, want)
		}
	}
}
