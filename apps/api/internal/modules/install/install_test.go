package install

import (
	"errors"
	"testing"

	"github.com/forum-framework/api/internal/shared/apperr"
)

func TestValidateInput(t *testing.T) {
	ok := CompleteInput{ForumName: "  Minha Comunidade ", Username: "dono", Email: "d@x.io", Password: "senha-forte-123"}
	clean, err := validateInput(ok)
	if err != nil || clean.ForumName != "Minha Comunidade" {
		t.Fatalf("valid input rejected: %+v %v", clean, err)
	}
	for _, bad := range []CompleteInput{
		{ForumName: "x", Username: "dono", Email: "d@x.io", Password: "senha-forte-123"},
		{ForumName: "Forum", Username: "", Email: "d@x.io", Password: "senha-forte-123"},
		{ForumName: "Forum", Username: "dono", Email: "d@x.io", Password: "curta"},
		{ForumName: "Forum", Username: "dono", Email: "d@x.io", Password: "senha-forte-123", Description: string(make([]rune, 301))},
	} {
		if _, err := validateInput(bad); !errors.Is(err, apperr.ErrValidation) {
			t.Errorf("expected validation error for %+v, got %v", bad, err)
		}
	}
}

func TestTokenOK(t *testing.T) {
	if !tokenOK("", "anything") {
		t.Fatal("no configured token means no check")
	}
	if !tokenOK("s3cret", "s3cret") || tokenOK("s3cret", "") || tokenOK("s3cret", "S3cret") {
		t.Fatal("token comparison wrong")
	}
}
