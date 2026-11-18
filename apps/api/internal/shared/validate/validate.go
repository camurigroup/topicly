// Package validate holds small, dependency-free validators for input at the
// API boundary. The backend is the source of truth for all validation.
package validate

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	emailRe    = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
)

// Username validates a 3–32 char handle of letters, digits and underscores.
func Username(s string) error {
	if l := utf8.RuneCountInString(s); l < 3 || l > 32 {
		return errors.New("username deve ter entre 3 e 32 caracteres")
	}
	if !usernameRe.MatchString(s) {
		return errors.New("username aceita apenas letras, números e underscore")
	}
	return nil
}

// Email validates basic email shape and length.
func Email(s string) error {
	s = strings.TrimSpace(s)
	if len(s) == 0 || len(s) > 255 || !emailRe.MatchString(s) {
		return errors.New("e-mail inválido")
	}
	return nil
}

// Password enforces a minimum length policy.
func Password(s string) error {
	if utf8.RuneCountInString(s) < 8 {
		return errors.New("a senha deve ter ao menos 8 caracteres")
	}
	if len(s) > 200 {
		return errors.New("a senha é muito longa")
	}
	return nil
}

// DisplayName validates the public display name length.
func DisplayName(s string) error {
	s = strings.TrimSpace(s)
	if l := utf8.RuneCountInString(s); l < 2 || l > 80 {
		return errors.New("o nome de exibição deve ter entre 2 e 80 caracteres")
	}
	return nil
}
