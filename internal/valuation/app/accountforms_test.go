package app_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/valuation/app"
)

// TestRecordAccountTerms checks an OKI account's form (D-147) and an IKE's
// earlier payments (D-149): each is recorded without a revaluation, a form
// only one the pack offers for the wrapper, earlier payments only for a
// wrapper that taxes what is withdrawn above what was paid in, in PLN and
// not below zero.
func TestRecordAccountTerms(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		record  func(r *app.InputRecorder) error
		want    []string
		wantErr error
	}{
		{name: "an OKI brokerage account", record: func(r *app.InputRecorder) error {
			return r.RecordAccountForm(t.Context(), "acc-1", "oki", pl.FormBrokerage)
		}, want: []string{"form acc-1 brokerage-account"}},
		{name: "a form OKI doesn't have", record: func(r *app.InputRecorder) error {
			return r.RecordAccountForm(t.Context(), "acc-1", "oki", "savings-book")
		}, wantErr: app.ErrInvalidInput},
		{name: "a form for a taxable account", record: func(r *app.InputRecorder) error {
			return r.RecordAccountForm(t.Context(), "acc-1", "taxable", pl.FormBank)
		}, wantErr: app.ErrInvalidInput},
		{name: "an IKE's earlier payments", record: func(r *app.InputRecorder) error {
			return r.RecordEarlierPayments(t.Context(), "acc-1", "ike", money(t, "40000", "PLN"))
		}, want: []string{"earlier payments acc-1 40000 PLN"}},
		{name: "earlier payments of an IKZE", record: func(r *app.InputRecorder) error {
			return r.RecordEarlierPayments(t.Context(), "acc-1", "ikze", money(t, "40000", "PLN"))
		}, wantErr: app.ErrInvalidInput},
		{name: "earlier payments below zero", record: func(r *app.InputRecorder) error {
			return r.RecordEarlierPayments(t.Context(), "acc-1", "ike", money(t, "-1", "PLN"))
		}, wantErr: app.ErrInvalidInput},
		{name: "earlier payments in euro", record: func(r *app.InputRecorder) error {
			return r.RecordEarlierPayments(t.Context(), "acc-1", "ike", money(t, "100", "EUR"))
		}, wantErr: app.ErrInvalidInput},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rec := &fakeRecorder{}
			l := twoAccounts(t)
			err := tt.record(app.NewInputRecorder(app.NewPlanner(l, l), rec, pl.NewPack()))
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
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("recorded (-want +got):\n%s", diff)
			}
		})
	}
}
