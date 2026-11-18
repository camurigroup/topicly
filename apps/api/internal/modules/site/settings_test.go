package site

import (
	"errors"
	"strings"
	"testing"

	"github.com/forum-framework/api/internal/shared/apperr"
)

func TestNormalize(t *testing.T) {
	ok := []struct{ key, in, want string }{
		{"home.banner.title", "  Olá  ", "Olá"},
		{"home.show.stats", "on", "true"},
		{"home.show.stats", "FALSE", "false"},
		{"home.banner.ctaLink", "/discover", "/discover"},
		{"home.banner.image", "https://cdn.exemplo.com/a.png", "https://cdn.exemplo.com/a.png"},
		{"home.banner.image", "", ""},
		{"home.intro.text", "linha 1\nlinha 2", "linha 1\nlinha 2"},
	}
	for _, c := range ok {
		got, err := Normalize(c.key, c.in)
		if err != nil || got != c.want {
			t.Errorf("Normalize(%q, %q) = %q, %v; want %q", c.key, c.in, got, err, c.want)
		}
	}

	bad := []struct{ key, in string }{
		{"home.unknown", "x"},
		{"home.banner.title", strings.Repeat("a", 121)},
		{"home.banner.title", "duas\nlinhas"},
		{"home.show.stats", "talvez"},
		{"home.banner.ctaLink", "https://evil.example"},
		{"home.banner.ctaLink", "//evil.example"},
		{"home.banner.image", "javascript:alert(1)"},
	}
	for _, c := range bad {
		if _, err := Normalize(c.key, c.in); !errors.Is(err, apperr.ErrValidation) {
			t.Errorf("Normalize(%q, %q) err = %v, want validation", c.key, c.in, err)
		}
	}
}

func TestDefinitionsDefaultsAreValid(t *testing.T) {
	seen := map[string]bool{}
	for _, d := range Definitions {
		if seen[d.Key] {
			t.Fatalf("duplicate key %s", d.Key)
		}
		seen[d.Key] = true
		if got, err := Normalize(d.Key, d.Default); err != nil || got != d.Default {
			t.Errorf("default of %s is not canonical: %q, %v", d.Key, got, err)
		}
	}
}
