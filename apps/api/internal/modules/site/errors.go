package site

import "github.com/forum-framework/api/internal/shared/apperr"

func validationErr(msg string) error { return &validationError{msg: msg} }

type validationError struct{ msg string }

func (e *validationError) Error() string { return e.msg }
func (e *validationError) Unwrap() error { return apperr.ErrValidation }
