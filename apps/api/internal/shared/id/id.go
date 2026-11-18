// Package id generates unique identifiers (UUID v4 strings) for records.
package id

import "github.com/google/uuid"

// New returns a new random UUID string.
func New() string { return uuid.NewString() }
