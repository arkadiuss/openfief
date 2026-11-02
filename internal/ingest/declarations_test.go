package ingest_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/arkadiuss/openfief/internal/ingest"
)

func TestUploadMode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		seeded   bool
		override ingest.Mode
		want     ingest.Mode
	}{
		{name: "an account not yet seeded is seeded by the upload", want: ingest.ModeOpening},
		{name: "a seeded account takes transactions", seeded: true, want: ingest.ModeTransactions},
		{name: "the owner overrides to transactions", override: ingest.ModeTransactions, want: ingest.ModeTransactions},
		{name: "the owner overrides to opening", seeded: true, override: ingest.ModeOpening, want: ingest.ModeOpening},
		{name: "an unknown mode is no override", seeded: true, override: ingest.Mode(9), want: ingest.ModeTransactions},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ingest.UploadMode(tt.seeded, tt.override); got != tt.want {
				t.Errorf("UploadMode(%v, %s) = %s, want %s", tt.seeded, tt.override, got, tt.want)
			}
		})
	}
}

func TestCheckInstitution(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		declared    []ingest.InstitutionID
		institution ingest.InstitutionID
		wantErr     error
	}{
		{name: "declared", declared: []ingest.InstitutionID{"ibkr"}, institution: "ibkr"},
		{name: "one of several", declared: []ingest.InstitutionID{"ibkr", "xtb"}, institution: "xtb"},
		{name: "another institution", declared: []ingest.InstitutionID{"aaa"}, institution: "eodhd", wantErr: ingest.ErrUndeclaredInstitution},
		{name: "none declared", institution: "nbp", wantErr: ingest.ErrUndeclaredInstitution},
		{name: "empty institution", declared: []ingest.InstitutionID{"ibkr"}, wantErr: ingest.ErrUndeclaredInstitution},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := ingest.Declarations{Institutions: tt.declared}
			if err := d.CheckInstitution(tt.institution); !errors.Is(err, tt.wantErr) {
				t.Errorf("CheckInstitution(%q) = %v, want %v", tt.institution, err, tt.wantErr)
			}
		})
	}
}

func TestUndeclaredCrossCheckPrices(t *testing.T) {
	t.Parallel()
	p := ingest.UndeclaredCrossCheckPrices(3)
	if p.Kind != ingest.ProblemSkipped || p.Ref != "cross_check_prices" || !strings.HasPrefix(p.Reason, "3 cross-check prices not stored") {
		t.Errorf("UndeclaredCrossCheckPrices(3) = %v", p)
	}
}
