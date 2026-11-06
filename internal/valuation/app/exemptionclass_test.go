package app_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
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
