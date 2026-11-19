package site

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/forum-framework/api/internal/shared/apperr"
)

func TestDefaultLayoutIsValid(t *testing.T) {
	raw, _ := json.Marshal(DefaultLayout)
	ws, err := ParseLayout(string(raw))
	if err != nil || len(ws) != len(DefaultLayout) {
		t.Fatalf("default layout must validate: %v", err)
	}
}

func TestParseLayoutNormalizes(t *testing.T) {
	ws, err := ParseLayout(`[{"id":"a","type":"topics","enabled":true,"column":"main","title":"  Equipe ","variant":"","limit":99,"source":"roles","roles":["admin","admin","owner"]}]`)
	if err != nil {
		t.Fatal(err)
	}
	w := ws[0]
	if w.Title != "Equipe" || w.Limit != MaxWidgetLimit || w.Variant != "list" || strings.Join(w.Roles, ",") != "admin,owner" {
		t.Fatalf("not normalized: %+v", w)
	}
}

func TestParseLayoutRejects(t *testing.T) {
	bad := []string{
		`not json`,
		`[{"id":"a","type":"banana","column":"main"}]`,
		`[{"id":"a","type":"stats","column":"middle"}]`,
		`[{"id":"a","type":"stats","column":"main"},{"id":"a","type":"intro","column":"side"}]`,
		`[{"id":"Bad Id","type":"stats","column":"main"}]`,
		`[{"id":"a","type":"topics","column":"main","source":"secret"}]`,
		`[{"id":"a","type":"topics","column":"main","source":"roles","roles":[]}]`,
		`[{"id":"a","type":"topics","column":"main","source":"category"}]`,
		`[{"id":"a","type":"stats","column":"main","variant":"pie"}]`,
		`[{"id":"a","type":"hero","column":"full"},{"id":"b","type":"hero","column":"full"}]`,
	}
	for _, b := range bad {
		if _, err := ParseLayout(b); !errors.Is(err, apperr.ErrValidation) {
			t.Errorf("expected validation error for %s, got %v", b, err)
		}
	}
}

func TestNormalizeLayoutSetting(t *testing.T) {
	v, err := Normalize("home.layout", `[{"id":"s","type":"stats","enabled":true,"column":"side"}]`)
	if err != nil || !strings.Contains(v, `"variant":"grid"`) {
		t.Fatalf("layout setting must be validated and canonical: %q %v", v, err)
	}
}
