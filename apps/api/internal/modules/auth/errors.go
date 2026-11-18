package auth

import "github.com/forum-framework/api/internal/shared/apperr"

// FieldError is a validation or conflict error tied to an input field.
type FieldError struct {
	Field string
	Msg   string
	kind  error
}

func (e *FieldError) Error() string { return e.Msg }

// Unwrap exposes the underlying sentinel (apperr.ErrValidation / ErrConflict).
func (e *FieldError) Unwrap() error { return e.kind }

func wrapValidation(err error) error {
	return &FieldError{Msg: err.Error(), kind: apperr.ErrValidation}
}

func fieldConflict(field, msg string) error {
	return &FieldError{Field: field, Msg: msg, kind: apperr.ErrConflict}
}
