package admin

import (
	"errors"
	"testing"

	"github.com/forum-framework/api/internal/shared/apperr"
)

func TestOrderBy(t *testing.T) {
	cols := map[string]string{"created": "u.created_at", "name": "u.username"}
	cases := []struct {
		sort, order, want string
	}{
		{"", "", "u.created_at DESC"},
		{"name", "asc", "u.username ASC"},
		{"name", "DESC", "u.username DESC"},
		{"created", "weird", "u.created_at DESC"},
	}
	for _, c := range cases {
		got, err := OrderBy(cols, "created", c.sort, c.order)
		if err != nil || got != c.want {
			t.Errorf("OrderBy(%q,%q) = %q, %v; want %q", c.sort, c.order, got, err, c.want)
		}
	}
	if _, err := OrderBy(cols, "created", "password_hash; DROP TABLE users", ""); !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("unknown sort must be rejected, got %v", err)
	}
}

func TestOneOf(t *testing.T) {
	if err := OneOf("status", "", "active", "banned"); err != nil {
		t.Fatalf("empty is allowed: %v", err)
	}
	if err := OneOf("status", "banned", "active", "banned"); err != nil {
		t.Fatal(err)
	}
	if err := OneOf("status", "deleted", "active", "banned"); !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("unknown value must be rejected, got %v", err)
	}
}
