package search

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/forum-framework/api/internal/shared/apperr"
)

func TestSearchRejectsInvalidParams(t *testing.T) {
	svc := NewService(nil, nil) // every case fails before touching the database
	cases := map[string]Params{
		"blank without filters": {Query: "   "},
		"too long":              {Query: strings.Repeat("a", 101)},
		"unknown type":          {Query: "go", Type: "posts"},
		"unknown sort":          {Query: "go", Sort: "random"},
		"bad from date":         {Query: "go", From: "18/11/2026"},
		"bad to date":           {Query: "go", To: "2026-13-01"},
		"inverted range":        {Query: "go", From: "2026-11-20", To: "2026-11-01"},
	}
	for name, p := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := svc.Search(context.Background(), p)
			if !errors.Is(err, apperr.ErrValidation) {
				t.Fatalf("err = %v, want validation error", err)
			}
		})
	}
}
