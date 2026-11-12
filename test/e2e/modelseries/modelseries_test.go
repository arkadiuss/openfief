package modelseries_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/test/e2e/modelseries"
)

// day parses a YYYY-MM-DD date or fails the test.
func day(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse(time.DateOnly, s)
	if err != nil {
		t.Fatalf("Parse(%q): %v", s, err)
	}
	return d
}

func TestFind(t *testing.T) {
	t.Parallel()
	a, ok := modelseries.Find("MODA.INDX")
	if !ok || a.Index != "MODA" {
		t.Errorf("Find(MODA.INDX) = %+v, %t, want MODA", a, ok)
	}
	if _, ok := modelseries.Find("FAKE20.INDX"); ok {
		t.Error("Find(FAKE20.INDX) found a series the model portfolio doesn't use")
	}
}

func TestCloses(t *testing.T) {
	t.Parallel()
	a, _ := modelseries.Find("MODA.INDX")
	b, _ := modelseries.Find("MODB.INDX")
	tests := []struct {
		name     string
		series   modelseries.Series
		from, to string
		want     []modelseries.Close
	}{
		{
			name: "a step's day holds the new level, and the weekend has no close", series: a, from: "2025-12-08", to: "2025-12-12",
			want: []modelseries.Close{
				{Date: "2025-12-08", Level: "100"},
				{Date: "2025-12-09", Level: "100"},
				{Date: "2025-12-10", Level: "120"},
				{Date: "2025-12-11", Level: "120"},
				{Date: "2025-12-12", Level: "120"},
			},
		},
		{
			name: "a series starts on its first step", series: b, from: "2025-11-01", to: "2025-12-05",
			want: []modelseries.Close{{Date: "2025-12-03", Level: "50"}, {Date: "2025-12-04", Level: "50"}, {Date: "2025-12-05", Level: "50"}},
		},
		{
			name: "it ends on the Friday that ends the IBKR week", series: b, from: "2026-03-06", to: "2026-04-30",
			want: []modelseries.Close{
				{Date: "2026-03-06", Level: "27"},
				{Date: "2026-03-09", Level: "45"},
				{Date: "2026-03-10", Level: "45"},
				{Date: "2026-03-11", Level: "45"},
				{Date: "2026-03-12", Level: "45"},
				{Date: "2026-03-13", Level: "45"},
			},
		},
		{name: "a range before the series", series: a, from: "2025-01-01", to: "2025-11-30"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.series.Closes(day(t, tt.from), day(t, tt.to))
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Closes(%s, %s) (-want +got):\n%s", tt.from, tt.to, diff)
			}
		})
	}
}

func TestClosesUnbounded(t *testing.T) {
	t.Parallel()
	a, _ := modelseries.Find("MODA.INDX")
	all := a.Closes(time.Time{}, time.Time{})
	if len(all) != 75 || all[0].Date != "2025-12-01" || all[len(all)-1].Date != modelseries.End {
		t.Errorf("MODA's closes = %d from %s to %s, want 75 from 2025-12-01 to %s", len(all), all[0].Date, all[len(all)-1].Date, modelseries.End)
	}
	for _, c := range all {
		if w := day(t, c.Date).Weekday(); w == time.Saturday || w == time.Sunday {
			t.Errorf("a close on %s, a %s", c.Date, w)
		}
	}
}
