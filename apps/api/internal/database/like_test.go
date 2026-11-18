package database

import (
	"reflect"
	"testing"
)

func TestContainsPatternEscapesWildcards(t *testing.T) {
	cases := map[string]string{
		"Go":      "%go%",
		"100%":    `%100\%%`,
		"snake_c": `%snake\_c%`,
		`a\b`:     `%a\\b%`,
	}
	for in, want := range cases {
		if got := ContainsPattern(in); got != want {
			t.Errorf("ContainsPattern(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSearchTerms(t *testing.T) {
	got := SearchTerms("  Go  go RUST  zig c d e ", 4)
	want := []string{"go", "rust", "zig", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SearchTerms = %v, want %v", got, want)
	}
	if SearchTerms("   ", 5) != nil {
		t.Fatal("blank query must yield no terms")
	}
}
