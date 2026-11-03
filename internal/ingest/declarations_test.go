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

// TestUndeclaredDeliveries checks D-179's drop in the domain: bond terms
// or index levels a connector may not deliver become one skipped problem
// per kind and an attention detail naming the kinds, never a count; what
// it may deliver, or a delivery of none, gives neither.
func TestUndeclaredDeliveries(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		decl         ingest.Declarations
		terms        int
		levels       int
		wantRefs     []string
		detailNames  string
		wantNoDetail bool
	}{
		{name: "neither declared", terms: 2, levels: 3, wantRefs: []string{"bond_terms", "index_levels"}, detailNames: "bond series terms and market index levels"},
		{name: "only levels undeclared", decl: ingest.Declarations{BondTerms: true}, terms: 2, levels: 3, wantRefs: []string{"index_levels"}, detailNames: "market index levels"},
		{name: "both declared", decl: ingest.Declarations{BondTerms: true, IndexLevels: true}, terms: 2, levels: 3, wantNoDetail: true},
		{name: "nothing delivered", wantNoDetail: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			drop := ingest.UndeclaredDeliveries(tt.decl, tt.terms, tt.levels)
			problems, detail := drop.Problems, drop.Detail
			var refs []string
			for _, p := range problems {
				if p.Kind != ingest.ProblemSkipped {
					t.Errorf("problem %v is not skipped", p)
				}
				refs = append(refs, p.Ref)
			}
			if strings.Join(refs, ",") != strings.Join(tt.wantRefs, ",") {
				t.Errorf("problem refs = %v, want %v", refs, tt.wantRefs)
			}
			if tt.wantNoDetail != (detail == "") || !strings.Contains(detail, tt.detailNames) || strings.ContainsAny(detail, "0123456789") {
				t.Errorf("detail = %q; want one naming %q and no count", detail, tt.detailNames)
			}
		})
	}
}
