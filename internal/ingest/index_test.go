package ingest_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
)

func TestIndexReadingValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		index   ingest.Index
		date    string
		source  string
		wantErr error
	}{
		{name: "CPI for a month", index: ingest.IndexCPI, date: "2026-09-01", source: "gus-cpi"},
		{name: "reference rate from a day", index: ingest.IndexReferenceRate, date: "2026-07-03", source: "nbp-rates"},
		{name: "CPI not on the first", index: ingest.IndexCPI, date: "2026-09-15", source: "gus-cpi", wantErr: ingest.ErrInvalidIndexReading},
		{name: "no index", date: "2026-09-01", source: "gus-cpi", wantErr: ingest.ErrInvalidIndexReading},
		{name: "unknown index", index: 9, date: "2026-09-01", source: "gus-cpi", wantErr: ingest.ErrInvalidIndexReading},
		{name: "no date", index: ingest.IndexReferenceRate, source: "nbp-rates", wantErr: ingest.ErrInvalidIndexReading},
		{name: "no source", index: ingest.IndexCPI, date: "2026-09-01", wantErr: ingest.ErrInvalidIndexReading},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := ingest.IndexReading{Index: tt.index, Value: kernel.NewDecimal(49, -3), Source: tt.source}
			if tt.date != "" {
				d, err := kernel.ParseDate(tt.date)
				if err != nil {
					t.Fatalf("ParseDate: %v", err)
				}
				r.Date = d
			}
			if diff := cmp.Diff(tt.wantErr, r.Validate(), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Validate (-want +got):\n%s", diff)
			}
		})
	}
}

func TestIndexString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		got, want string
	}{
		{ingest.IndexCPI.String(), "cpi"},
		{ingest.IndexReferenceRate.String(), "reference-rate"},
		{ingest.Index(0).String(), "Index(0)"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("String() = %q, want %q", tt.got, tt.want)
		}
	}
}
