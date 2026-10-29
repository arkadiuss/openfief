package portfolio_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/portfolio"
)

// Synthetic target names: instruments as "ISIN/symbol" with the XX test
// prefix, an index by a made-up name.
const (
	etfA   = "XX0000000AA1/ETFA"
	etfB   = "XX0000000BB2/ETFB"
	index1 = "FAKE20"
)

// mustDec parses a decimal or fails the test.
func mustDec(t *testing.T, s string) kernel.Decimal {
	t.Helper()
	d, err := kernel.ParseDecimal(s)
	if err != nil {
		t.Fatalf("ParseDecimal(%q): %v", s, err)
	}
	return d
}

// mustCurrency parses a currency code or fails the test.
func mustCurrency(t *testing.T, code string) kernel.Currency {
	t.Helper()
	c, err := kernel.ParseCurrency(code)
	if err != nil {
		t.Fatalf("ParseCurrency(%q): %v", code, err)
	}
	return c
}

// targetSpec is a target as text, so tables need no *testing.T.
type targetSpec struct {
	kind     portfolio.TargetKind
	name     string
	currency string
	weight   string
}

// build turns a spec into a Target.
func (s targetSpec) build(t *testing.T) portfolio.Target {
	t.Helper()
	tg := portfolio.Target{Kind: s.kind, Name: s.name, Weight: mustDec(t, s.weight)}
	if s.currency != "" {
		tg.Currency = mustCurrency(t, s.currency)
	}
	return tg
}

// buildTargets turns specs into Targets.
func buildTargets(t *testing.T, specs []targetSpec) []portfolio.Target {
	t.Helper()
	out := make([]portfolio.Target, 0, len(specs))
	for _, s := range specs {
		out = append(out, s.build(t))
	}
	return out
}

func TestTargetKind(t *testing.T) {
	t.Parallel()
	for _, k := range []portfolio.TargetKind{portfolio.TargetInstrument, portfolio.TargetIndex} {
		got, err := portfolio.ParseTargetKind(k.String())
		if err != nil || got != k {
			t.Errorf("ParseTargetKind(%q) = %v, %v; want %v", k.String(), got, err, k)
		}
	}
	if got := portfolio.TargetKind(0).String(); got != "TargetKind(0)" {
		t.Errorf("TargetKind(0).String() = %q", got)
	}
	if _, err := portfolio.ParseTargetKind("bond"); !cmp.Equal(err, portfolio.ErrInvalidModel, cmpopts.EquateErrors()) {
		t.Errorf("ParseTargetKind(bond) = %v, want %v", err, portfolio.ErrInvalidModel)
	}
}

func TestModelValidate(t *testing.T) {
	t.Parallel()
	inst := func(name, weight string) targetSpec {
		return targetSpec{kind: portfolio.TargetInstrument, name: name, weight: weight}
	}
	// most is MaxTargets targets of 2%; tooMany is one more, still summing
	// to 100%: 50 of 1.96% and one of 2%.
	most := make([]targetSpec, 0, portfolio.MaxTargets)
	tooMany := make([]targetSpec, 0, portfolio.MaxTargets+1)
	tooMany = append(tooMany, inst("XX/B", "2"))
	for i := range portfolio.MaxTargets {
		most = append(most, inst("XX/"+strings.Repeat("A", i+1), "2"))
		tooMany = append(tooMany, inst("XX/"+strings.Repeat("A", i+1), "1.96"))
	}
	tests := []struct {
		name    string
		enabled bool
		targets []targetSpec
		wantErr error
	}{
		{name: "off and empty, the default"},
		{name: "on and empty", enabled: true, wantErr: portfolio.ErrInvalidModel},
		{name: "60/40 on", enabled: true, targets: []targetSpec{inst(etfA, "60"), inst(etfB, "40")}},
		{name: "an index with its currency", enabled: true, targets: []targetSpec{
			inst(etfA, "33.33"), inst(etfB, "33.33"), {kind: portfolio.TargetIndex, name: index1, currency: "PLN", weight: "33.34"},
		}},
		{name: "off with targets is still checked", targets: []targetSpec{inst(etfA, "60")}, wantErr: portfolio.ErrInvalidModel},
		{name: "weights sum to 99.99", enabled: true, targets: []targetSpec{inst(etfA, "60"), inst(etfB, "39.99")}, wantErr: portfolio.ErrInvalidModel},
		{name: "weights sum to 100.01", enabled: true, targets: []targetSpec{inst(etfA, "60"), inst(etfB, "40.01")}, wantErr: portfolio.ErrInvalidModel},
		{name: "a third decimal place", enabled: true, targets: []targetSpec{inst(etfA, "59.995"), inst(etfB, "40.005")}, wantErr: portfolio.ErrInvalidModel},
		{name: "a zero weight", enabled: true, targets: []targetSpec{inst(etfA, "100"), inst(etfB, "0")}, wantErr: portfolio.ErrInvalidModel},
		{name: "a negative weight", enabled: true, targets: []targetSpec{inst(etfA, "110"), inst(etfB, "-10")}, wantErr: portfolio.ErrInvalidModel},
		{name: "one target at 100", enabled: true, targets: []targetSpec{inst(etfA, "100")}},
		{name: "the same target twice", enabled: true, targets: []targetSpec{inst(etfA, "50"), inst(etfA, "50")}, wantErr: portfolio.ErrInvalidModel},
		{name: "an instrument and an index of one name", enabled: true, targets: []targetSpec{
			inst(index1, "50"), {kind: portfolio.TargetIndex, name: index1, currency: "PLN", weight: "50"},
		}},
		{name: "an index without a currency", enabled: true, targets: []targetSpec{{kind: portfolio.TargetIndex, name: index1, weight: "100"}}, wantErr: portfolio.ErrInvalidModel},
		{name: "an instrument with a currency", enabled: true, targets: []targetSpec{{kind: portfolio.TargetInstrument, name: etfA, currency: "USD", weight: "100"}}, wantErr: portfolio.ErrInvalidModel},
		{name: "no kind", enabled: true, targets: []targetSpec{{name: etfA, weight: "100"}}, wantErr: portfolio.ErrInvalidModel},
		{name: "no name", enabled: true, targets: []targetSpec{inst("", "100")}, wantErr: portfolio.ErrInvalidModel},
		{name: "a name too long", enabled: true, targets: []targetSpec{inst(strings.Repeat("x", 65), "100")}, wantErr: portfolio.ErrInvalidModel},
		{name: "the most targets", enabled: true, targets: most},
		{name: "too many targets", enabled: true, targets: tooMany, wantErr: portfolio.ErrInvalidModel},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := portfolio.Model{Enabled: tt.enabled, Targets: buildTargets(t, tt.targets)}
			if diff := cmp.Diff(tt.wantErr, m.Validate(), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Validate() error (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTargetKey(t *testing.T) {
	t.Parallel()
	tg := portfolio.Target{Kind: portfolio.TargetIndex, Name: index1}
	if got := tg.Key(); got != "index:FAKE20" {
		t.Errorf("Key() = %q, want index:FAKE20", got)
	}
}
