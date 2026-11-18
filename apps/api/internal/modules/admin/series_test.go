package admin

import (
	"testing"
	"time"
)

func day(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }

func TestBuildSeriesDaily(t *testing.T) {
	now := day(2026, 11, 18)
	counts := map[string]int64{"2026-11-18": 3, "2026-11-16": 2, "2026-10-01": 99}
	pts, err := BuildSeries(PeriodDaily, counts, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(pts) != 30 {
		t.Fatalf("daily series must have 30 points, got %d", len(pts))
	}
	last := pts[len(pts)-1]
	if last.Label != "2026-11-18" || last.Value != 3 {
		t.Fatalf("last point = %+v", last)
	}
	if pts[len(pts)-3].Value != 2 || pts[len(pts)-2].Value != 0 {
		t.Fatalf("gaps must be zero-filled: %+v", pts[len(pts)-3:])
	}
}

func TestBuildSeriesWeeklySumsDays(t *testing.T) {
	now := day(2026, 11, 18) // Wednesday; ISO week starts Monday 2026-11-16
	counts := map[string]int64{"2026-11-16": 1, "2026-11-18": 2, "2026-11-15": 5}
	pts, err := BuildSeries(PeriodWeekly, counts, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(pts) != 12 {
		t.Fatalf("weekly series must have 12 points, got %d", len(pts))
	}
	if pts[11].Label != "2026-11-16" || pts[11].Value != 3 {
		t.Fatalf("current week = %+v", pts[11])
	}
	if pts[10].Value != 5 {
		t.Fatalf("previous week = %+v", pts[10])
	}
}

func TestBuildSeriesMonthly(t *testing.T) {
	now := day(2026, 11, 18)
	counts := map[string]int64{"2026-11-01": 4, "2026-11-30": 1, "2026-01-15": 7, "2025-11-30": 100}
	pts, err := BuildSeries(PeriodMonthly, counts, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(pts) != 12 || pts[0].Label != "2025-12" || pts[11].Label != "2026-11" {
		t.Fatalf("monthly window wrong: first=%+v last=%+v", pts[0], pts[len(pts)-1])
	}
	if pts[11].Value != 5 || pts[1].Value != 7 {
		t.Fatalf("monthly sums wrong: %+v", pts)
	}
}

func TestBuildSeriesRejectsUnknownPeriod(t *testing.T) {
	if _, err := BuildSeries("hourly", nil, time.Now()); err == nil {
		t.Fatal("expected error")
	}
}

func TestWindowStart(t *testing.T) {
	now := day(2026, 11, 18)
	cases := map[string]time.Time{
		PeriodDaily:   day(2026, 10, 20),
		PeriodWeekly:  day(2026, 8, 31),
		PeriodMonthly: day(2025, 12, 1),
	}
	for p, want := range cases {
		got, err := WindowStart(p, now)
		if err != nil || !got.Equal(want) {
			t.Errorf("WindowStart(%s) = %v, %v; want %v", p, got, err, want)
		}
	}
}
