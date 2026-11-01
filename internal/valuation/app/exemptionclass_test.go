package app_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/valuation/app"
)

// fakeAttention records the items of kind only it was asked to raise, with
// their detail, and to clear.
type fakeAttention struct {
	only            string
	raised, cleared []string
}

func (f *fakeAttention) Raise(_ context.Context, kind, subject, detail string) error {
	if kind == f.only {
		f.raised = append(f.raised, kind+" "+subject+": "+detail)
	}
	return nil
}

func (f *fakeAttention) Clear(_ context.Context, kind, subject string) error {
	if kind == f.only {
		f.cleared = append(f.cleared, kind+" "+subject)
	}
	return nil
}

// TestRevalueRaisesUnclassified checks D-132: revaluing an OKI account that
// holds a fund the pack can't classify raises one attention item per
// instrument, none once the owner classified it, and none in a taxable
// account, whose wrapper has no yearly tax; the owner's classification
// clears the item.
func TestRevalueRaisesUnclassified(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		wrapper jurisdiction.Wrapper
		classes map[string]jurisdiction.ExemptionClass
		want    []string
		cleared []string
	}{
		{name: "OKI, unclassified", wrapper: "oki", want: []string{app.UnclassifiedKind + " " + etfISIN + ": the jurisdiction pack can't tell this instrument's exemption class for oki's yearly tax"}},
		{
			name: "OKI, classified by the owner", wrapper: "oki", classes: map[string]jurisdiction.ExemptionClass{etfISIN: pl.Investment},
			cleared: []string{app.UnclassifiedKind + " " + etfISIN},
		},
		{name: "taxable", wrapper: "taxable"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := newFakeStore()
			s.prices = []valuation.Price{usdPrice(t, "2026-03-02", "100"), usdPrice(t, "2026-03-03", "101.5")}
			s.rates = []valuation.FXRate{rate(t, "USD", "2026-02-27", "4.0000"), rate(t, "USD", "2026-03-02", "4.0000")}
			s.wrappers[acc1], s.classes = tt.wrapper, tt.classes
			attention := &fakeAttention{only: app.UnclassifiedKind}
			r := revaluer(t, etfLedger(t), s, "2026-03-03").WithAttention(attention)
			if _, err := r.Revalue(t.Context(), acc1, date(t, "2026-03-02")); err != nil {
				t.Fatalf("Revalue: %v", err)
			}
			if diff := cmp.Diff([][]string{tt.want, tt.cleared}, [][]string{attention.raised, attention.cleared}); diff != "" {
				t.Errorf("raised, cleared (-want +got):\n%s", diff)
			}
		})
	}
}

// TestSetExemptionClass checks the owner's classification (D-132): the owner
// sets one of the pack's classes for a valid ISIN, nothing is revalued,
// and anything else is refused.
func TestSetExemptionClass(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, token, isin string
		class             jurisdiction.ExemptionClass
		want              []string
		wantErr           error
	}{
		{name: "the owner classifies a fund", token: ownerToken, isin: etfISIN, class: pl.Savings, want: []string{"exemption class " + etfISIN + " savings"}},
		{name: "a class the pack lacks", token: ownerToken, isin: etfISIN, class: "bonds", wantErr: app.ErrInvalidInput, want: []string{}},
		{name: "not an ISIN", token: ownerToken, isin: "FUND", class: pl.Other, wantErr: app.ErrInvalidInput, want: []string{}},
		{name: "not the owner", token: "read-session", isin: etfISIN, class: pl.Other, wantErr: app.ErrNotOwner, want: []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rec := &fakeRecorder{}
			err := newSettings(t, newFakeStore(), rec).SetExemptionClass(t.Context(), "https://fief.example.org", tt.token, tt.isin, tt.class)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("error (-want +got):\n%s", diff)
			}
			got := make([]string, 0, len(rec.calls))
			for _, c := range rec.calls {
				got = append(got, c.what)
				if len(c.revals) != 0 {
					t.Errorf("%s revalued %v", c.what, c.revals)
				}
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("recorded (-want +got):\n%s", diff)
			}
		})
	}
}

func TestOfferedClasses(t *testing.T) {
	t.Parallel()
	if diff := cmp.Diff([]jurisdiction.ExemptionClass{"savings", "investment", "other"}, app.OfferedClasses(pl.NewPack())); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
}
