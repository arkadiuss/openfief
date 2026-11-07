package app_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/views"
	"github.com/arkadiuss/openfief/internal/views/app"
)

// balanceViews returns views over synthetic balance assertions in acc-1 and
// acc-2, and batches of an XTB instance in each, asking an authorizer that
// lets nobody act on acc-2 (D-163).
func balanceViews(t *testing.T) (*app.Views, []views.BalanceCheck) {
	t.Helper()
	f := newFixture(t)
	cashOld := ledger.CashAssertion("acc-1", date(t, "2026-03-01"), money(t, "100.00", "PLN"))
	cashNew := ledger.CashAssertion("acc-1", date(t, "2026-03-02"), money(t, "120.00", "PLN"))
	etf := ledger.HoldingAssertion("acc-1", date(t, "2026-03-02"), instrument(t, isinIE, "ETF1"), kernel.NewQuantity(dec(t, "3")))
	hidden := ledger.CashAssertion("acc-2", date(t, "2026-03-02"), money(t, "7.00", "EUR"))
	f.ledger.checks = map[ledger.AccountID][]ledgerapp.AssertionCheck{
		"acc-1": {
			{Assertion: cashOld, Status: ledgerapp.AssertionFlagged, Derived: dec(t, "90.00")},
			{Assertion: cashNew, Status: ledgerapp.AssertionFlagged, Derived: dec(t, "119.995")},
			{Assertion: etf, Status: ledgerapp.AssertionMatched, Derived: dec(t, "3")},
		},
		"acc-2": {{Assertion: hidden, Status: ledgerapp.AssertionFlagged, Derived: dec(t, "1.00")}},
	}
	at := time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC)
	history := &fakeHistory{batches: []ingest.BatchRecord{
		{ID: "b-2", Instance: "xtb-one", Institution: "xtb", Finished: at, Accounts: []ledger.AccountID{"acc-1"}, Problems: 2, Unsupported: 2},
		{ID: "b-1", Instance: "xtb-two", Institution: "xtb", Finished: at.Add(-time.Hour), Accounts: []ledger.AccountID{"acc-2"}, Problems: 1, Unsupported: 1},
	}}
	v, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "PLN"), cur(t, "PLN"),
		app.WithAuthorizer(&denyingAuthorizer{denied: "acc-2"}), app.WithSyncHistory(history))
	if err != nil {
		t.Fatalf("app.New: %v", err)
	}
	return v, []views.BalanceCheck{
		views.CheckBalance(&cashOld, dec(t, "90.00")),
		views.CheckBalance(&cashNew, dec(t, "119.995")),
		views.CheckBalance(&etf, dec(t, "3")),
	}
}

// TestBalanceReportOwnerOnly checks that only the owner gets the balance
// check, never a holder of read access (D-137), and that it covers only the
// accounts the authorizer lets the owner view (D-163).
func TestBalanceReportOwnerOnly(t *testing.T) {
	t.Parallel()
	for _, h := range []identity.Holder{0, identity.HolderScriptToken, identity.HolderViewPlugin, identity.HolderDemoSession} {
		t.Run(h.String(), func(t *testing.T) {
			t.Parallel()
			v, _ := balanceViews(t)
			ctx := identityapp.WithCaller(t.Context(), identity.Caller{Holder: h, ID: "test"})
			if _, err := v.BalanceReport(ctx, true); !errors.Is(err, app.ErrOwnerOnly) {
				t.Errorf("BalanceReport for %s: error %v, want ErrOwnerOnly", h, err)
			}
		})
	}
}

// TestBalanceReport checks the owner's report: the latest assertion of each
// ledger account by default, every one with all, both ordered, and the
// unsupported rows of the instances tied to the accounts the owner may view.
func TestBalanceReport(t *testing.T) {
	t.Parallel()
	owner := identity.Caller{Holder: identity.HolderOwner, ID: "cli"}
	unsupported := []views.UnsupportedRows{{
		Instance: "xtb-one", Institution: "xtb", Batch: "b-2", Finished: time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC),
		Accounts: []ledger.AccountID{"acc-1"}, Rows: 2,
	}}
	tests := []struct {
		name       string
		all        bool
		pick       []int
		wantFailed int
	}{
		{name: "latest", pick: []int{1, 2}, wantFailed: 0},
		{name: "all", all: true, pick: []int{0, 1, 2}, wantFailed: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v, checks := balanceViews(t)
			got, err := v.BalanceReport(identityapp.WithCaller(t.Context(), owner), tt.all)
			if err != nil {
				t.Fatalf("BalanceReport: %v", err)
			}
			want := views.BalanceReport{Unsupported: unsupported}
			for _, i := range tt.pick {
				want.Checks = append(want.Checks, checks[i])
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("BalanceReport (-want +got):\n%s", diff)
			}
			if n := got.Failed(); n != tt.wantFailed {
				t.Errorf("Failed = %d, want %d", n, tt.wantFailed)
			}
		})
	}
}
