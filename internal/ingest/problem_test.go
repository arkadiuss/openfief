package ingest_test

import (
	"testing"

	"github.com/arkadiuss/openfief/internal/ingest"
)

func TestCountUnsupported(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		problems []ingest.Problem
		want     int
	}{
		{name: "none", want: 0},
		{name: "rows left out on purpose don't count", problems: []ingest.Problem{ingest.Skipped("row 1", "summary row")}, want: 0},
		{
			name: "rows OpenFief can't book count",
			problems: []ingest.Problem{
				ingest.Unsupported("row 2", "option trade"), ingest.Skipped("row 3", "before go-live"), ingest.Unsupported("row 4", "interest"),
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ingest.CountUnsupported(tt.problems); got != tt.want {
				t.Errorf("CountUnsupported = %d, want %d", got, tt.want)
			}
		})
	}
}
