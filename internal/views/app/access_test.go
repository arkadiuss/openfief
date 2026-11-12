package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views"
	"github.com/arkadiuss/openfief/internal/views/app"
)

var errAuthz = errors.New("authorizer unavailable")

// denyingAuthorizer answers as identityapp.Policy, except that nobody may
// act on the account denied, as a SpiceDB that granted the caller other
// accounts only would (D-163); with fail set it can't tell. It records the
// callers it was asked about.
type denyingAuthorizer struct {
	denied ledger.AccountID
	fail   error
	asked  map[identity.Caller]bool
}

func (d *denyingAuthorizer) MayAct(_ context.Context, c identity.Caller, a identity.Action, account string) (bool, error) {
	if d.fail != nil {
		return false, d.fail
	}
	if d.asked == nil {
		d.asked = map[identity.Caller]bool{}
	}
	d.asked[c] = true
	return identity.MayAct(c, a) && account != string(d.denied), nil
}

// withAuthorizer returns f's views asking a, with f's bindings of
// ibkr-main to acc-1 and acc-2.
func withAuthorizer(t *testing.T, f *fixture, a identityapp.Authorizer) *app.Views {
	t.Helper()
	v, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "PLN"), cur(t, "PLN"), app.WithPortfolio(f.ledger),
		app.WithAuthorizer(a), app.WithConnections(twoBindings{}))
	if err != nil {
		t.Fatalf("app.New: %v", err)
	}
	return v
}

// twoBindings binds ibkr-main to acc-1 and acc-2.
type twoBindings struct{}

func (twoBindings) Bindings(context.Context) ([]ingest.Binding, error) {
	return []ingest.Binding{{Instance: "ibkr-main", Account: "acc-1"}, {Instance: "ibkr-main", Account: "acc-2"}}, nil
}

// withoutAcc2 returns the fixture build makes without acc-2, as the views
// of a caller who may not view acc-2 should see it.
func withoutAcc2(t *testing.T, build func(t *testing.T) *fixture) *fixture {
	t.Helper()
	f := build(t)
	var kept []ledger.Account
	for _, a := range f.ledger.accounts {
		if a.ID != "acc-2" {
			kept = append(kept, a)
		}
	}
	f.ledger.accounts = kept
	return f
}

// TestViewsComputeOverViewableAccounts checks that the lists, totals, the
// allocation and performance of a caller who may not view acc-2 are those
// of a portfolio without acc-2 (D-163).
func TestViewsComputeOverViewableAccounts(t *testing.T) {
	t.Parallel()
	ctx := reader(t)
	authz := &denyingAuthorizer{denied: "acc-2"}
	on := date(t, "2026-03-02")
	tests := []struct {
		name    string
		fixture func(t *testing.T) *fixture
		get     func(v *app.Views) (any, error)
	}{
		{"accounts", newFixture, func(v *app.Views) (any, error) { return v.Accounts(ctx) }},
		{"valuation totals", newFixture, func(v *app.Views) (any, error) { return v.ValuationTotals(ctx, on, on) }},
		{"valuation statuses", newFixture, func(v *app.Views) (any, error) { return v.ValuationStatuses(ctx) }},
		{"allocation", newFixture, func(v *app.Views) (any, error) { return v.Allocation(ctx, on, views.BySleeve, views.Gross) }},
		{"performance", newPerformanceFixture, func(v *app.Views) (any, error) { return v.Performance(ctx, kernel.Date{}, kernel.Date{}) }},
	}
	for _, tt := range tests {
		denied := withAuthorizer(t, tt.fixture(t), authz)
		want := withAuthorizer(t, withoutAcc2(t, tt.fixture), &denyingAuthorizer{})
		got, err := tt.get(denied)
		if err != nil {
			t.Fatalf("%s: %v", tt.name, err)
		}
		expected, err := tt.get(want)
		if err != nil {
			t.Fatalf("%s without acc-2: %v", tt.name, err)
		}
		if diff := cmp.Diff(expected, got, cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("%s with acc-2 hidden differs from without it (-want +got):\n%s", tt.name, diff)
		}
	}
	hidden := withAuthorizer(t, newFixture(t), authz)
	if accounts, err := hidden.Accounts(ctx); err != nil || len(accounts) != 1 {
		t.Errorf("Accounts = %v, %v; want acc-1 alone", accounts, err)
	}
	bindings, err := hidden.ConnectorBindings(ctx)
	if diff := cmp.Diff([]ingest.Binding{{Instance: "ibkr-main", Account: "acc-1"}}, bindings); err != nil || diff != "" {
		t.Errorf("ConnectorBindings = %v (-want +got):\n%s", err, diff)
	}
	if want := map[identity.Caller]bool{{Holder: identity.HolderScriptToken, ID: "test"}: true}; !cmp.Equal(want, authz.asked) {
		t.Errorf("the authorizer was asked about %v, want only the request's caller", authz.asked)
	}
}

// TestViewsRefuseAnAccountTheCallerMayNotView checks that every query
// about one account answers for an account the caller may not view as for
// one not stored (D-163).
func TestViewsRefuseAnAccountTheCallerMayNotView(t *testing.T) {
	t.Parallel()
	ctx := reader(t)
	v := withAuthorizer(t, newFixture(t), &denyingAuthorizer{denied: "acc-2"})
	on := date(t, "2026-03-02")
	queries := map[string]func(id ledger.AccountID) error{
		"Account":           func(id ledger.AccountID) error { return errOf(v.Account(ctx, id)) },
		"Transactions":      func(id ledger.AccountID) error { return errOf(v.Transactions(ctx, id, nil, views.DefaultPageSize)) },
		"Holdings":          func(id ledger.AccountID) error { return errOf(v.Holdings(ctx, id, on)) },
		"Lots":              func(id ledger.AccountID) error { return errOf(v.Lots(ctx, id, on)) },
		"Cash":              func(id ledger.AccountID) error { return errOf(v.Cash(ctx, id, on)) },
		"AssertionChecks":   func(id ledger.AccountID) error { return errOf(v.AssertionChecks(ctx, id, 0)) },
		"ValuationStatus":   func(id ledger.AccountID) error { return errOf(v.ValuationStatus(ctx, id)) },
		"AccountValuations": func(id ledger.AccountID) error { return errOf(v.AccountValuations(ctx, id, on, on)) },
		"AccountPerformance": func(id ledger.AccountID) error {
			return errOf(v.AccountPerformance(ctx, id, kernel.Date{}, kernel.Date{}))
		},
	}
	for name, query := range queries {
		if err := query("acc-2"); !errors.Is(err, app.ErrUnknownAccount) {
			t.Errorf("%s(acc-2) = %v, want %v", name, err, app.ErrUnknownAccount)
		}
		if err := query("acc-missing"); !errors.Is(err, app.ErrUnknownAccount) {
			t.Errorf("%s(acc-missing) = %v, want %v", name, err, app.ErrUnknownAccount)
		}
		if err := query("acc-1"); err != nil {
			t.Errorf("%s(acc-1) = %v, want it viewable", name, err)
		}
	}
}

// errOf returns a query's error, dropping its result.
func errOf[T any](_ T, err error) error { return err }

func TestViewsWhenTheAuthorizerFails(t *testing.T) {
	t.Parallel()
	ctx := reader(t)
	v := withAuthorizer(t, newFixture(t), &denyingAuthorizer{fail: errAuthz})
	if _, err := v.Accounts(ctx); !errors.Is(err, errAuthz) {
		t.Errorf("Accounts = %v, want %v", err, errAuthz)
	}
	if _, err := v.Account(ctx, "acc-1"); !errors.Is(err, errAuthz) || errors.Is(err, app.ErrUnknownAccount) {
		t.Errorf("Account = %v, want %v and not an unknown account", err, errAuthz)
	}
	if _, err := v.ConnectorBindings(ctx); !errors.Is(err, errAuthz) {
		t.Errorf("ConnectorBindings = %v, want %v", err, errAuthz)
	}
	if _, err := v.Transfers(ctx); !errors.Is(err, errAuthz) {
		t.Errorf("Transfers = %v, want %v", err, errAuthz)
	}
}

// TestViewsWithoutACaller checks that a request without a caller views
// nothing.
func TestViewsWithoutACaller(t *testing.T) {
	t.Parallel()
	v := newFixture(t).views(t)
	if accounts, err := v.Accounts(t.Context()); err != nil || len(accounts) != 0 {
		t.Errorf("Accounts without a caller = %v, %v; want none", accounts, err)
	}
	if _, err := v.Account(t.Context(), "acc-1"); !errors.Is(err, app.ErrUnknownAccount) {
		t.Errorf("Account without a caller = %v, want %v", err, app.ErrUnknownAccount)
	}
}

// TestTransfersOverViewableAccounts checks that a caller who may not view
// acc-2 sees no transfer with a leg there, and that a confirmed pair whose
// leg there still matches is not reported broken (D-163).
func TestTransfersOverViewableAccounts(t *testing.T) {
	t.Parallel()
	confirmed := func(w, d string) ledger.TransferLink {
		return ledger.TransferLink{Pair: ledger.TransferPair{Withdrawal: transferKey(w), Deposit: transferKey(d)}, Decision: ledger.TransferConfirmed, Version: 1}
	}
	// w1 -> d1 matches; w2 -> d2 matches; w1 -> gone has a deleted leg, so
	// it is broken.
	f := newTransferFixture(t, confirmed("w1", "d1"), confirmed("w2", "d2"), confirmed("w1", "gone"))
	ctx := reader(t)
	all, err := withAuthorizer(t, f, &denyingAuthorizer{}).Transfers(ctx)
	if err != nil {
		t.Fatalf("Transfers: %v", err)
	}
	hidden, err := withAuthorizer(t, f, &denyingAuthorizer{denied: "acc-2"}).Transfers(ctx)
	if err != nil {
		t.Fatalf("Transfers with acc-2 hidden: %v", err)
	}
	want := map[string][][2]string{
		"all confirmed":    {{"w1", "d1"}, {"w2", "d2"}},
		"all broken":       {{"w1", "gone"}},
		"hidden confirmed": {},
		"hidden suggested": {},
		"hidden broken":    {{"w1", "gone"}},
	}
	got := map[string][][2]string{
		"all confirmed":    legKeys(all.Confirmed),
		"all broken":       legKeys(all.Broken),
		"hidden confirmed": legKeys(hidden.Confirmed),
		"hidden suggested": legKeys(hidden.Suggested),
		"hidden broken":    legKeys(hidden.Broken),
	}
	if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("transfers (-want +got):\n%s", diff)
	}
}
