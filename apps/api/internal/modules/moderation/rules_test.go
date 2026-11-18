package moderation

import (
	"errors"
	"testing"
	"time"

	"github.com/forum-framework/api/internal/shared/apperr"
)

func TestExpiry(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	exp, err := Expiry(0, now)
	if err != nil || exp != nil {
		t.Fatalf("0 hours must mean permanent, got %v %v", exp, err)
	}
	exp, err = Expiry(48, now)
	if err != nil || exp == nil || !exp.Equal(now.Add(48*time.Hour)) {
		t.Fatalf("48h expiry wrong: %v %v", exp, err)
	}
	for _, bad := range []int{-1, MaxDurationHours + 1} {
		if _, err := Expiry(bad, now); !errors.Is(err, apperr.ErrValidation) {
			t.Errorf("duration %d should be invalid, got %v", bad, err)
		}
	}
}

func TestSanctionActive(t *testing.T) {
	now := time.Now().UTC()
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)
	cases := []struct {
		name string
		s    Sanction
		want bool
	}{
		{"permanent", Sanction{}, true},
		{"future expiry", Sanction{ExpiresAt: &future}, true},
		{"expired", Sanction{ExpiresAt: &past}, false},
		{"revoked", Sanction{RevokedAt: &past}, false},
	}
	for _, c := range cases {
		if got := c.s.ActiveAt(now); got != c.want {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}

func TestValidateReport(t *testing.T) {
	ok := CreateReportInput{TargetType: TargetTopic, TargetID: "x", Reason: "spam"}
	if err := ok.Validate(); err != nil {
		t.Fatalf("valid report rejected: %v", err)
	}
	bad := []CreateReportInput{
		{TargetType: "category", TargetID: "x", Reason: "spam"},
		{TargetType: TargetUser, TargetID: "", Reason: "spam"},
		{TargetType: TargetReply, TargetID: "x", Reason: "whatever"},
		{TargetType: TargetReply, TargetID: "x", Reason: "other", Details: string(make([]rune, 1001))},
	}
	for i, in := range bad {
		if err := in.Validate(); !errors.Is(err, apperr.ErrValidation) {
			t.Errorf("case %d should be invalid, got %v", i, err)
		}
	}
}

func TestValidStatus(t *testing.T) {
	for _, s := range []string{StatusPending, StatusReviewing, StatusResolved, StatusRejected} {
		if !ValidStatus(s) {
			t.Errorf("%s should be valid", s)
		}
	}
	if ValidStatus("closed") {
		t.Error("closed should be invalid")
	}
}
