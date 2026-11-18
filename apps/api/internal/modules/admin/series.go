// Package admin provides the administrative dashboard: real metrics from the
// database and filterable listings of users and content.
package admin

import (
	"time"

	"github.com/forum-framework/api/internal/shared/apperr"
)

// Series periods.
const (
	PeriodDaily   = "daily"   // last 30 days
	PeriodWeekly  = "weekly"  // last 12 ISO weeks (Monday start)
	PeriodMonthly = "monthly" // last 12 months
)

const (
	dailyPoints   = 30
	weeklyPoints  = 12
	monthlyPoints = 12
	dateLayout    = "2006-01-02"
)

// Point is one bucket of a time series.
type Point struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
}

type seriesError string

func (e seriesError) Error() string { return string(e) }
func (e seriesError) Unwrap() error { return apperr.ErrValidation }

func truncDay(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func weekStart(t time.Time) time.Time {
	d := truncDay(t)
	offset := (int(d.Weekday()) + 6) % 7 // Monday = 0
	return d.AddDate(0, 0, -offset)
}

func monthStart(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}

// WindowStart returns the first instant (UTC) covered by a period's series.
func WindowStart(period string, now time.Time) (time.Time, error) {
	switch period {
	case PeriodDaily:
		return truncDay(now).AddDate(0, 0, -(dailyPoints - 1)), nil
	case PeriodWeekly:
		return weekStart(now).AddDate(0, 0, -7*(weeklyPoints-1)), nil
	case PeriodMonthly:
		return monthStart(now).AddDate(0, -(monthlyPoints - 1), 0), nil
	}
	return time.Time{}, seriesError("período inválido (use daily, weekly ou monthly)")
}

// BuildSeries folds per-day counts ("YYYY-MM-DD" → n) into zero-filled
// buckets for the period, oldest first. Days outside the window are ignored.
func BuildSeries(period string, daily map[string]int64, now time.Time) ([]Point, error) {
	start, err := WindowStart(period, now)
	if err != nil {
		return nil, err
	}
	var (
		points []Point
		index  = map[string]int{}
		bucket func(time.Time) string
	)
	switch period {
	case PeriodDaily:
		bucket = func(t time.Time) string { return t.Format(dateLayout) }
		for i := 0; i < dailyPoints; i++ {
			points = append(points, Point{Label: start.AddDate(0, 0, i).Format(dateLayout)})
		}
	case PeriodWeekly:
		bucket = func(t time.Time) string { return weekStart(t).Format(dateLayout) }
		for i := 0; i < weeklyPoints; i++ {
			points = append(points, Point{Label: start.AddDate(0, 0, 7*i).Format(dateLayout)})
		}
	case PeriodMonthly:
		bucket = func(t time.Time) string { return t.Format("2006-01") }
		for i := 0; i < monthlyPoints; i++ {
			points = append(points, Point{Label: start.AddDate(0, i, 0).Format("2006-01")})
		}
	}
	for i, p := range points {
		index[p.Label] = i
	}
	for d, n := range daily {
		t, err := time.Parse(dateLayout, d)
		if err != nil || t.Before(start) {
			continue
		}
		if i, ok := index[bucket(t)]; ok {
			points[i].Value += n
		}
	}
	return points, nil
}
