package category

import (
	"errors"
	"testing"

	"github.com/forum-framework/api/internal/shared/apperr"
)

func sp(s string) *string { return &s }
func ip(i int) *int       { return &i }
func bp(b bool) *bool     { return &b }

func base() Category {
	return Category{Name: "Geral", Description: sp("desc"), Icon: sp("💬"), CoverURL: sp("/files/brand/c.png"), Position: 3, IsActive: true}
}

func TestApplyUpdateKeepsOmittedFields(t *testing.T) {
	c := base()
	if err := applyUpdate(&c, UpdateInput{IsActive: bp(false)}); err != nil {
		t.Fatal(err)
	}
	if c.IsActive || c.Name != "Geral" || *c.Description != "desc" || *c.Icon != "💬" ||
		*c.CoverURL != "/files/brand/c.png" || c.Position != 3 {
		t.Fatalf("toggle must keep other fields: %+v", c)
	}
}

func TestApplyUpdateEmptyStringClears(t *testing.T) {
	c := base()
	if err := applyUpdate(&c, UpdateInput{CoverURL: sp(""), Description: sp("  "), Position: ip(0)}); err != nil {
		t.Fatal(err)
	}
	if c.CoverURL != nil || c.Description != nil || c.Position != 0 || c.Icon == nil {
		t.Fatalf("empty strings must clear only those fields: %+v", c)
	}
}

func TestApplyUpdateValidatesName(t *testing.T) {
	c := base()
	if err := applyUpdate(&c, UpdateInput{Name: sp("x")}); !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("short name must be invalid, got %v", err)
	}
}
