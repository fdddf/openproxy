package services

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
)

// golang-migrate's pure-Go sqlite driver pulls in modernc.org/sqlite, which
// registers the database/sql driver name "sqlite" — the same name the gorm
// driver's github.com/glebarez/go-sqlite registers, so importing both panics at
// init. Rather than drop back to a CGO driver, the SQLite path gets this small
// runner. It writes the same schema_migrations table golang-migrate uses, so
// the `migrate` CLI still understands a database produced here.

type sqliteMigration struct {
	version int64
	name    string
	sql     string
}

func runSQLiteMigrations(db *sql.DB, source fs.FS) error {
	migrations, err := loadSQLiteMigrations(source)
	if err != nil {
		return err
	}

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version BIGINT NOT NULL PRIMARY KEY,
		dirty BOOLEAN NOT NULL
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	var (
		current int64
		dirty   bool
	)
	err = db.QueryRow(`SELECT version, dirty FROM schema_migrations LIMIT 1`).Scan(&current, &dirty)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("read schema version: %w", err)
	}
	if dirty {
		return fmt.Errorf("database is dirty at migration version %d; resolve it manually before starting", current)
	}

	for _, m := range migrations {
		if m.version <= current {
			continue
		}
		if err := applySQLiteMigration(db, m); err != nil {
			return err
		}
	}

	return nil
}

func applySQLiteMigration(db *sql.DB, m sqliteMigration) error {
	// Mark dirty first so a crash mid-migration is visible on the next start.
	if err := setSQLiteVersion(db, m.version, true); err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin migration %d: %w", m.version, err)
	}
	if _, err := tx.Exec(m.sql); err != nil {
		tx.Rollback()
		return fmt.Errorf("apply migration %d (%s): %w", m.version, m.name, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %d: %w", m.version, err)
	}

	return setSQLiteVersion(db, m.version, false)
}

func setSQLiteVersion(db *sql.DB, version int64, dirty bool) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin version update: %w", err)
	}
	if _, err := tx.Exec(`DELETE FROM schema_migrations`); err != nil {
		tx.Rollback()
		return fmt.Errorf("clear schema_migrations: %w", err)
	}
	if _, err := tx.Exec(`INSERT INTO schema_migrations (version, dirty) VALUES (?, ?)`, version, dirty); err != nil {
		tx.Rollback()
		return fmt.Errorf("record schema version %d: %w", version, err)
	}
	return tx.Commit()
}

// loadSQLiteMigrations reads every *.up.sql in source, ordered by the numeric
// version prefix that golang-migrate's naming convention puts on each file.
func loadSQLiteMigrations(source fs.FS) ([]sqliteMigration, error) {
	entries, err := fs.ReadDir(source, ".")
	if err != nil {
		return nil, fmt.Errorf("read migrations: %w", err)
	}

	var migrations []sqliteMigration
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".up.sql") {
			continue
		}

		prefix, rest, found := strings.Cut(name, "_")
		if !found {
			return nil, fmt.Errorf("migration %q is missing a version prefix", name)
		}
		version, err := strconv.ParseInt(prefix, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("migration %q has a non-numeric version prefix: %w", name, err)
		}

		body, err := fs.ReadFile(source, name)
		if err != nil {
			return nil, fmt.Errorf("read migration %q: %w", name, err)
		}

		migrations = append(migrations, sqliteMigration{
			version: version,
			name:    strings.TrimSuffix(rest, ".up.sql"),
			sql:     string(body),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})
	return migrations, nil
}
