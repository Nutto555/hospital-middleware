package postgres

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5" // registers the pgx5:// scheme
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/Nutto555/hospital-middleware/migrations"
)

// Migrate applies every pending migration. Running it on an up-to-date database is a no-op.
func Migrate(databaseURL string) error {
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("load migrations: %w", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", src, pgx5URL(databaseURL))
	if err != nil {
		return fmt.Errorf("open migrator: %w", err)
	}
	defer m.Close()
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}

// pgx5URL rewrites a postgres:// URL to the scheme golang-migrate's pgx/v5 driver registers.
func pgx5URL(u string) string {
	for _, prefix := range []string{"postgres://", "postgresql://"} {
		if rest, ok := strings.CutPrefix(u, prefix); ok {
			return "pgx5://" + rest
		}
	}
	return u
}
