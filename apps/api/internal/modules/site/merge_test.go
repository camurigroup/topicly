package site

import (
	"testing"
	"time"
)

func TestMergeActivityNewestFirstAndCapped(t *testing.T) {
	base := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	at := func(id string, minutesAgo int) Activity {
		return Activity{ID: id, CreatedAt: base.Add(-time.Duration(minutesAgo) * time.Minute)}
	}
	topics := []Activity{at("t1", 1), at("t2", 10), at("t3", 30)}
	replies := []Activity{at("r1", 0), at("r2", 5), at("r3", 20)}

	got := mergeActivity(topics, replies, 4)
	want := []string{"r1", "t1", "r2", "t2"}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i, id := range want {
		if got[i].ID != id {
			t.Fatalf("position %d = %s, want %s", i, got[i].ID, id)
		}
	}
	if n := len(mergeActivity(nil, replies, 10)); n != 3 {
		t.Fatalf("merge with empty side = %d items, want 3", n)
	}
}
