package topic

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestQuoteTrimsLongTitles(t *testing.T) {
	if got := quote("Olá"); got != "“Olá”" {
		t.Fatalf("quote short = %q", got)
	}
	long := strings.Repeat("á", 200)
	got := quote(long)
	if n := utf8.RuneCountInString(got); n != maxQuotedTitle+2 {
		t.Fatalf("quoted rune count = %d, want %d", n, maxQuotedTitle+2)
	}
	if !strings.HasSuffix(got, "…”") {
		t.Fatalf("long title should end with ellipsis: %q", got)
	}
}

func TestTopicLink(t *testing.T) {
	if got := topicLink("t1", ""); got != "/t/t1" {
		t.Fatalf("topicLink = %q", got)
	}
	if got := topicLink("t1", "r9"); got != "/t/t1#reply-r9" {
		t.Fatalf("topicLink with reply = %q", got)
	}
}
