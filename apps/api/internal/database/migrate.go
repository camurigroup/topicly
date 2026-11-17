package database

import (
	"database/sql"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"
)

// Migrate applies all pending .sql migrations from fsys/subdir, in filename
// order, each inside its own transaction. It returns the number applied.
func Migrate(db *sql.DB, dialect string, fsys fs.FS, subdir string) (int, error) {
	if err := ensureMigrationsTable(db); err != nil {
		return 0, err
	}
	applied, err := appliedVersions(db)
	if err != nil {
		return 0, err
	}

	entries, err := fs.ReadDir(fsys, subdir)
	if err != nil {
		return 0, fmt.Errorf("read migrations dir %q: %w", subdir, err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	count := 0
	for _, name := range names {
		version := strings.TrimSuffix(name, ".sql")
		if applied[version] {
			continue
		}
		content, err := fs.ReadFile(fsys, subdir+"/"+name)
		if err != nil {
			return count, fmt.Errorf("read migration %s: %w", name, err)
		}
		if err := applyOne(db, dialect, version, string(content)); err != nil {
			return count, fmt.Errorf("apply migration %s: %w", name, err)
		}
		count++
	}
	return count, nil
}

func ensureMigrationsTable(db *sql.DB) error {
	const ddl = `CREATE TABLE IF NOT EXISTS schema_migrations (
        version VARCHAR(255) NOT NULL PRIMARY KEY,
        applied_at TIMESTAMP NOT NULL
    )`
	if _, err := db.Exec(ddl); err != nil {
		return fmt.Errorf("ensure schema_migrations table: %w", err)
	}
	return nil
}

func appliedVersions(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		applied[v] = true
	}
	return applied, rows.Err()
}

func applyOne(db *sql.DB, dialect, version, content string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if strings.TrimSpace(content) != "" {
		if _, err := tx.Exec(content); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	insert := "INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)"
	if dialect == "postgres" {
		insert = "INSERT INTO schema_migrations (version, applied_at) VALUES ($1, $2)"
	}
	if _, err := tx.Exec(insert, version, time.Now().UTC()); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
