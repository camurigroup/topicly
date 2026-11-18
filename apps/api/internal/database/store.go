package database

import "database/sql"

// Store bundles the connection pool with its dialect so repositories can write
// a single set of queries and rebind placeholders per database.
type Store struct {
	DB      *sql.DB
	Dialect string
}

// NewStore builds a Store.
func NewStore(db *sql.DB, dialect string) *Store {
	return &Store{DB: db, Dialect: dialect}
}

// Rebind adapts "?" placeholders to the store's dialect.
func (s *Store) Rebind(query string) string { return Rebind(s.Dialect, query) }
