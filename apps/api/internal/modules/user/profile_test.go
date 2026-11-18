package user

import (
	"errors"
	"testing"

	"github.com/forum-framework/api/internal/shared/apperr"
)

func strp(s string) *string { return &s }

func TestMergeProfileKeepsOmittedFields(t *testing.T) {
	cur := &User{DisplayName: "Dono", Bio: strp("Fundador")}

	name, bio, err := mergeProfile(cur, UpdateProfileInput{DisplayName: strp("tauz")})
	if err != nil || name != "tauz" || bio == nil || *bio != "Fundador" {
		t.Fatalf("omitted bio must be kept: %q %v %v", name, bio, err)
	}

	name, bio, err = mergeProfile(cur, UpdateProfileInput{Bio: strp("  nova bio ")})
	if err != nil || name != "Dono" || bio == nil || *bio != "nova bio" {
		t.Fatalf("omitted name must be kept, bio trimmed: %q %v %v", name, bio, err)
	}
}

func TestMergeProfileEmptyBioClears(t *testing.T) {
	cur := &User{DisplayName: "Dono", Bio: strp("Fundador")}
	_, bio, err := mergeProfile(cur, UpdateProfileInput{Bio: strp("   ")})
	if err != nil || bio != nil {
		t.Fatalf("empty bio must clear it, got %v %v", bio, err)
	}
}

func TestMergeProfileValidates(t *testing.T) {
	cur := &User{DisplayName: "Dono"}
	if _, _, err := mergeProfile(cur, UpdateProfileInput{DisplayName: strp("")}); !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("empty display name must be invalid, got %v", err)
	}
	long := make([]rune, 501)
	for i := range long {
		long[i] = 'a'
	}
	if _, _, err := mergeProfile(cur, UpdateProfileInput{Bio: strp(string(long))}); !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("bio over 500 must be invalid, got %v", err)
	}
}
