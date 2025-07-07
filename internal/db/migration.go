package db

import (
	"errors"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// ApplyMigrations runs all unapplied migrations. Returns true if migrations were applied, false if no change is done.
func ApplyMigrations(dbUrl string, migrationsSource string) (bool, error) {
	m, err := migrate.New(
		migrationsSource,
		dbUrl)

	if err != nil {
		return false, err
	}

	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return false, nil
		}
		return false, err
	}
	return true, err
}
