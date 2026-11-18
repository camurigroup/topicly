// Package apperr defines sentinel errors shared across domain modules.
package apperr

import "errors"

var (
	// ErrNotFound indicates a requested record does not exist.
	ErrNotFound = errors.New("not found")
	// ErrConflict indicates a uniqueness or state conflict.
	ErrConflict = errors.New("conflict")
	// ErrForbidden indicates the caller lacks permission.
	ErrForbidden = errors.New("forbidden")
	// ErrUnauthorized indicates missing or invalid authentication.
	ErrUnauthorized = errors.New("unauthorized")
	// ErrValidation indicates invalid input.
	ErrValidation = errors.New("validation")
)
