package kernel_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
)

func TestParseDate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr error
	}{
		{name: "ISO date", in: "2026-03-31", want: "2026-03-31"},
		{name: "leap day", in: "2024-02-29", want: "2024-02-29"},
		{name: "no leap day", in: "2026-02-29", wantErr: kernel.ErrInvalidDate},
		{name: "with time", in: "2026-03-31T10:00:00Z", wantErr: kernel.ErrInvalidDate},
		{name: "day first", in: "31.03.2026", wantErr: kernel.ErrInvalidDate},
		{name: "empty", in: "", wantErr: kernel.ErrInvalidDate},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := kernel.ParseDate(tt.in)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("ParseDate(%q) error (-want +got):\n%s", tt.in, diff)
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(tt.want, got.String()); diff != "" {
				t.Errorf("ParseDate(%q) (-want +got):\n%s", tt.in, diff)
			}
		})
	}
}

func TestNewDate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		year    int
		month   time.Month
		day     int
		want    string
		wantErr error
	}{
		{name: "valid", year: 2026, month: time.January, day: 15, want: "2026-01-15"},
		{name: "day 31 in a 30-day month", year: 2026, month: time.April, day: 31, wantErr: kernel.ErrInvalidDate},
		{name: "month 13", year: 2026, month: 13, day: 1, wantErr: kernel.ErrInvalidDate},
		{name: "day 0", year: 2026, month: time.May, day: 0, wantErr: kernel.ErrInvalidDate},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := kernel.NewDate(tt.year, tt.month, tt.day)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("NewDate error (-want +got):\n%s", diff)
			}
			if err != nil {
				return
			}
			gotFields := []any{got.String(), got.Year(), got.Month(), got.Day()}
			wantFields := []any{tt.want, tt.year, tt.month, tt.day}
			if diff := cmp.Diff(wantFields, gotFields); diff != "" {
				t.Errorf("NewDate (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDateArithmetic(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		from     string
		days     int
		wantDate string
	}{
		{name: "next day", from: "2026-03-31", days: 1, wantDate: "2026-04-01"},
		{name: "across a leap day", from: "2024-02-28", days: 2, wantDate: "2024-03-01"},
		{name: "backwards across a year", from: "2026-01-01", days: -1, wantDate: "2025-12-31"},
		{name: "across a DST change in Europe", from: "2026-03-28", days: 2, wantDate: "2026-03-30"},
		{name: "a year", from: "2026-01-01", days: 365, wantDate: "2027-01-01"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			from := date(t, tt.from)
			got := from.AddDays(tt.days)
			if diff := cmp.Diff(date(t, tt.wantDate), got); diff != "" {
				t.Errorf("AddDays(%d) (-want +got):\n%s", tt.days, diff)
			}
			if diff := cmp.Diff(tt.days, from.DaysUntil(got)); diff != "" {
				t.Errorf("DaysUntil (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDateCompare(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		a, b string
		want int
	}{
		{name: "earlier", a: "2025-12-31", b: "2026-01-01", want: -1},
		{name: "same", a: "2026-01-01", b: "2026-01-01", want: 0},
		{name: "later", a: "2026-02-01", b: "2026-01-31", want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a, b := date(t, tt.a), date(t, tt.b)
			got := []any{a.Compare(b), a.Before(b), a.After(b), a.Equal(b), a == b}
			want := []any{tt.want, tt.want < 0, tt.want > 0, tt.want == 0, tt.want == 0}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Compare, Before, After, Equal, == (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDateZeroValue(t *testing.T) {
	t.Parallel()
	got := []bool{kernel.Date{}.IsZero(), date(t, "2026-01-01").IsZero()}
	if diff := cmp.Diff([]bool{true, false}, got); diff != "" {
		t.Errorf("IsZero (-want +got):\n%s", diff)
	}
}
