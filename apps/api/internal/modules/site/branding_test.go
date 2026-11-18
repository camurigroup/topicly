package site

import (
	"errors"
	"testing"

	"github.com/forum-framework/api/internal/shared/apperr"
)

func TestNormalizeBranding(t *testing.T) {
	ok := []struct{ key, in, want string }{
		{"brand.name", "  Minha Comunidade ", "Minha Comunidade"},
		{"theme.primary", "#7C3AED", "#7c3aed"},
		{"theme.primary", "#fff", "#ffffff"},
		{"theme.primary", "", ""},
		{"social.github", "https://github.com/exemplo", "https://github.com/exemplo"},
		{"brand.logo", "/files/logo.png", "/files/logo.png"},
		{"legal.terms", "Linha 1\nLinha 2", "Linha 1\nLinha 2"},
		{"footer.links", "Sobre|/about\nContato|https://exemplo.com", "Sobre|/about\nContato|https://exemplo.com"},
	}
	for _, c := range ok {
		got, err := Normalize(c.key, c.in)
		if err != nil || got != c.want {
			t.Errorf("Normalize(%q, %q) = %q, %v; want %q", c.key, c.in, got, err, c.want)
		}
	}

	bad := []struct{ key, in string }{
		{"theme.primary", "red"},
		{"theme.primary", "#12345"},
		{"theme.primary", "#gggggg"},
		{"social.github", "javascript:alert(1)"},
		{"footer.links", "Sem separador"},
		{"footer.links", "Mau|javascript:alert(1)"},
	}
	for _, c := range bad {
		if _, err := Normalize(c.key, c.in); !errors.Is(err, apperr.ErrValidation) {
			t.Errorf("Normalize(%q, %q) err = %v, want validation", c.key, c.in, err)
		}
	}
}

func TestDefinitionsHaveGroupAndLabel(t *testing.T) {
	for _, d := range Definitions {
		if d.Group == "" || d.Label == "" {
			t.Errorf("definition %q needs group and label", d.Key)
		}
	}
}
