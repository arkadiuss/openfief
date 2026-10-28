package portfolio_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/portfolio"
)

func TestMaxDrawdown(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		points []string
		// want is "peak trough peakValue troughValue depth"; empty for the
		// zero Drawdown.
		want    []string
		wantErr error
	}{
		{
			name:   "the deeper of two falls",
			points: []string{"2024-01-01 100", "2024-01-02 120", "2024-01-03 90", "2024-01-04 130", "2024-01-05 65", "2024-01-06 70"},
			want:   []string{"2024-01-04", "2024-01-05", "130", "65", "0.5"},
		},
		{
			name:   "a deep fall from an early peak beats a later, larger drop in points",
			points: []string{"2024-01-01 10", "2024-01-02 5", "2024-01-03 1000", "2024-01-04 600"},
			want:   []string{"2024-01-01", "2024-01-02", "10", "5", "0.5"},
		},
		{
			name:   "of equal falls, the earliest",
			points: []string{"2024-01-01 100", "2024-01-02 80", "2024-01-03 100", "2024-01-04 80"},
			want:   []string{"2024-01-01", "2024-01-02", "100", "80", "0.2"},
		},
		{
			name:   "the trough is the lowest point after the peak",
			points: []string{"2024-01-01 100", "2024-01-02 90", "2024-01-03 70", "2024-01-04 95"},
			want:   []string{"2024-01-01", "2024-01-03", "100", "70", "0.3"},
		},
		{
			name:   "never falls",
			points: []string{"2024-01-01 1", "2024-01-02 2", "2024-01-03 3"},
			want:   []string{"2024-01-01", "2024-01-01", "1", "1", "0"},
		},
		{
			name:   "one point",
			points: []string{"2024-01-01 7"},
			want:   []string{"2024-01-01", "2024-01-01", "7", "7", "0"},
		},
		{
			name:   "the depth keeps 12 places",
			points: []string{"2024-01-01 3", "2024-01-02 1"},
			want:   []string{"2024-01-01", "2024-01-02", "3", "1", "0.666666666667"},
		},
		{name: "no points"},
		{name: "a value that is not positive", points: []string{"2024-01-01 3", "2024-01-02 0"}, wantErr: portfolio.ErrInvalidHistory},
		{name: "out of date order", points: []string{"2024-01-02 3", "2024-01-01 2"}, wantErr: portfolio.ErrInvalidHistory},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := portfolio.MaxDrawdown(pointsOf(t, tt.points))
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("MaxDrawdown() error (-want +got):\n%s", diff)
			}
			var want portfolio.Drawdown
			if tt.want != nil {
				want = portfolio.Drawdown{
					Peak: mustDate(t, tt.want[0]), Trough: mustDate(t, tt.want[1]),
					PeakValue: mustDec(t, tt.want[2]), TroughValue: mustDec(t, tt.want[3]), Depth: mustDec(t, tt.want[4]),
				}
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("MaxDrawdown() (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSince(t *testing.T) {
	t.Parallel()
	series := []string{"2024-01-01 100", "2024-01-02 150", "2024-01-03 120", "2024-01-04 90"}
	tests := []struct {
		name string
		from string
		want []string
	}{
		{name: "from a day inside", from: "2024-01-02", want: []string{"2024-01-02 100", "2024-01-03 80", "2024-01-04 60"}},
		{name: "from before the first day", from: "2023-06-01", want: series},
		{name: "from after the last day", from: "2024-02-01"},
		{name: "from the last day", from: "2024-01-04", want: []string{"2024-01-04 100"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := portfolio.Since(pointsOf(t, series), mustDate(t, tt.from))
			if err != nil {
				t.Fatalf("Since: %v", err)
			}
			if diff := cmp.Diff(pointsOf(t, tt.want), got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("Since() (-want +got):\n%s", diff)
			}
		})
	}
	thirds, err := portfolio.Since(pointsOf(t, []string{"2024-01-01 3", "2024-01-02 1"}), mustDate(t, "2024-01-01"))
	if err != nil {
		t.Fatalf("Since: %v", err)
	}
	if want := mustDec(t, "33.333333333333"); !thirds[1].Value.Equal(want) {
		t.Errorf("Since() rounds to %s, want %s", thirds[1].Value, want)
	}
}
