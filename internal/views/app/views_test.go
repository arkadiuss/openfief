package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/valuation"
	valuationapp "github.com/arkadiuss/openfief/internal/valuation/app"
	"github.com/arkadiuss/openfief/internal/views"
	"github.com/arkadiuss/openfief/internal/views/app"
)

func TestAllocation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		by   views.Dimension
		want [][2]string
	}{
		// 1050 bond + 120 ETF1 + 50 PLN + 400 (100 USD) + 430 (100 EUR) = 2050 PLN.
		{views.ByAssetClass, [][2]string{{"treasury_bond", "1050.00"}, {"cash", "880.00"}, {"listed", "120.00"}}},
		{views.ByCurrency, [][2]string{{"PLN", "1100.00"}, {"USD", "520.00"}, {"EUR", "430.00"}}},
		{views.ByInstitution, [][2]string{{"broker-a", "1620.00"}, {"broker-b", "430.00"}}},
		{views.BySleeve, [][2]string{{"core", "1620.00"}, {"offense", "430.00"}}},
	}
	for _, tt := range tests {
		t.Run(tt.by.String(), func(t *testing.T) {
			t.Parallel()
			f := newFixture(t)
			got, err := f.views(t).Allocation(context.Background(), f.on, tt.by, views.Gross)
			if err != nil {
				t.Fatalf("Allocation() error = %v", err)
			}
			want := views.Allocation{By: tt.by, Total: money(t, "2050.00", "PLN")}
			for _, g := range tt.want {
				want.Groups = append(want.Groups, views.AllocationGroup{Key: g[0], Reporting: money(t, g[1], "PLN")})
			}
			if diff := cmp.Diff(want, got.Allocation); diff != "" {
				t.Errorf("Allocation() mismatch (-want +got):\n%s", diff)
			}
			wantUnvalued := []views.Unvalued{{Account: "acc-1", Holding: f.ledger.holdings["acc-1"][2], Missing: "price"}}
			if diff := cmp.Diff(wantUnvalued, got.Unvalued); diff != "" {
				t.Errorf("Allocation() unvalued mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestAllocationFailures(t *testing.T) {
	t.Parallel()
	t.Run("cash without an FX rate", func(t *testing.T) {
		t.Parallel()
		f := newFixture(t)
		f.rates = f.rates[:1] // no EUR rate
		if _, err := f.views(t).Allocation(context.Background(), f.on, views.BySleeve, views.Gross); !errors.Is(err, valuation.ErrNoRate) {
			t.Errorf("Allocation() error = %v, want %v", err, valuation.ErrNoRate)
		}
	})
	t.Run("a failing store", func(t *testing.T) {
		t.Parallel()
		f := newFixture(t)
		f.valuations.fail = errStore
		if _, err := f.views(t).Allocation(context.Background(), f.on, views.BySleeve, views.Gross); !errors.Is(err, errStore) {
			t.Errorf("Allocation() error = %v, want %v", err, errStore)
		}
	})
}

func TestReportingCurrency(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	if got := f.views(t).ReportingCurrency(); !got.Equal(cur(t, "PLN")) {
		t.Errorf("ReportingCurrency() = %s, want PLN", got)
	}
	_, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "EUR"))
	unsupported, ok := errors.AsType[*app.UnsupportedReportingCurrencyError](err)
	if !ok || !unsupported.Currency.Equal(cur(t, "EUR")) {
		t.Errorf("app.New(EUR) error = %v, want *UnsupportedReportingCurrencyError for EUR", err)
	}
}

func TestAccountValuations(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	v := f.views(t)
	ctx := context.Background()
	got, err := v.AccountValuations(ctx, "acc-1", date(t, "2026-03-02"), date(t, "2026-03-31"))
	if err != nil {
		t.Fatalf("AccountValuations() error = %v", err)
	}
	want := app.AccountValuations{
		Valuations: f.valuations.vals[2:4],
		Totals: []views.DateTotal{{
			Date: f.on, Reporting: money(t, "1170.00", "PLN"),
			ByCurrency: []kernel.Money{money(t, "1050.00", "PLN"), money(t, "30", "USD")},
			Net:        money(t, "1170.00", "PLN"), Liquidation: money(t, "1170.00", "PLN"), Cash: money(t, "0", "PLN"),
		}},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("AccountValuations() mismatch (-want +got):\n%s", diff)
	}
	if _, err := v.AccountValuations(ctx, "nope", f.on, f.on); !errors.Is(err, app.ErrUnknownAccount) {
		t.Errorf("AccountValuations(unknown) error = %v, want %v", err, app.ErrUnknownAccount)
	}
	if _, err := v.AccountValuations(ctx, "acc-1", f.on, f.on.AddDays(-1)); !errors.Is(err, views.ErrInvalidRange) {
		t.Errorf("AccountValuations(reversed) error = %v, want %v", err, views.ErrInvalidRange)
	}
}

func TestValuationTotals(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	got, err := f.views(t).ValuationTotals(context.Background(), date(t, "2026-03-01"), f.on)
	if err != nil {
		t.Fatalf("ValuationTotals() error = %v", err)
	}
	want := []views.DateTotal{
		{
			Date: date(t, "2026-03-01"), Reporting: money(t, "1165.00", "PLN"), ByCurrency: []kernel.Money{money(t, "1049.00", "PLN"), money(t, "29", "USD")},
			Net: money(t, "1165.00", "PLN"), Liquidation: money(t, "1165.00", "PLN"), Cash: money(t, "0", "PLN"),
		},
		{
			Date: f.on, Reporting: money(t, "1170.00", "PLN"), ByCurrency: []kernel.Money{money(t, "1050.00", "PLN"), money(t, "30", "USD")},
			Net: money(t, "1170.00", "PLN"), Liquidation: money(t, "1170.00", "PLN"), Cash: money(t, "0", "PLN"),
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ValuationTotals() mismatch (-want +got):\n%s", diff)
	}
}

func TestValuationStatus(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	v := f.views(t)
	ctx := context.Background()
	got, err := v.ValuationStatuses(ctx)
	if err != nil {
		t.Fatalf("ValuationStatuses() error = %v", err)
	}
	want := []valuationapp.ValuationStatus{{Account: "acc-1", ValuedThrough: f.on, Pending: true}, {Account: "acc-2"}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ValuationStatuses() mismatch (-want +got):\n%s", diff)
	}
	one, err := v.ValuationStatus(ctx, "acc-1")
	if err != nil || !one.Pending || !one.ValuedThrough.Equal(f.on) {
		t.Errorf("ValuationStatus(acc-1) = %+v, %v", one, err)
	}
	if _, err := v.ValuationStatus(ctx, "nope"); !errors.Is(err, app.ErrUnknownAccount) {
		t.Errorf("ValuationStatus(unknown) error = %v, want %v", err, app.ErrUnknownAccount)
	}
}

// withLedgerRows adds two transactions and a matched and a flagged balance
// assertion to acc-1.
func withLedgerRows(t *testing.T, f *fixture) {
	t.Helper()
	f.ledger.checks = map[ledger.AccountID][]ledgerapp.AssertionCheck{"acc-1": {
		{Assertion: ledger.CashAssertion("acc-1", f.on, money(t, "50", "PLN")), Status: ledgerapp.AssertionMatched, Derived: dec(t, "50")},
		{Assertion: ledger.CashAssertion("acc-1", f.on, money(t, "99", "USD")), Status: ledgerapp.AssertionFlagged, Derived: dec(t, "100")},
	}}
	f.ledger.txs = map[ledger.AccountID][]ledger.Transaction{"acc-1": {
		{Key: ledger.TransactionKey{Source: "test", ExternalID: "1"}, Account: "acc-1", Kind: ledger.KindDeposit, Date: f.on},
		{Key: ledger.TransactionKey{Source: "test", ExternalID: "2"}, Account: "acc-1", Kind: ledger.KindDeposit, Date: f.on},
	}}
}

func TestAssertionChecks(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	withLedgerRows(t, f)
	v := f.views(t)
	flagged, err := v.AssertionChecks(context.Background(), "acc-1", ledgerapp.AssertionFlagged)
	if err != nil || len(flagged) != 1 || !flagged[0].Flagged() {
		t.Errorf("AssertionChecks(flagged) = %v, %v, want the one flagged check", flagged, err)
	}
	all, err := v.AssertionChecks(context.Background(), "acc-1", 0)
	if err != nil || len(all) != 2 {
		t.Errorf("AssertionChecks(all) = %v, %v, want both checks", all, err)
	}
}

func TestTransactions(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	withLedgerRows(t, f)
	v := f.views(t)
	page, err := v.Transactions(context.Background(), "acc-1", nil, 1)
	if err != nil || len(page.Transactions) != 1 || page.Next == nil {
		t.Errorf("Transactions(size 1) = %+v, %v, want one and a cursor", page, err)
	}
	if _, err := v.Transactions(context.Background(), "acc-1", nil, 0); !errors.Is(err, views.ErrInvalidPageSize) {
		t.Errorf("Transactions(size 0) error = %v, want %v", err, views.ErrInvalidPageSize)
	}
}

func TestUnknownAccount(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	v := f.views(t)
	ctx := context.Background()
	_, errAccount := v.Account(ctx, "nope")
	_, errTxs := v.Transactions(ctx, "nope", nil, 1)
	_, errHoldings := v.Holdings(ctx, "nope", f.on)
	_, errLots := v.Lots(ctx, "nope", f.on)
	_, errCash := v.Cash(ctx, "nope", f.on)
	_, errChecks := v.AssertionChecks(ctx, "nope", 0)
	for i, err := range []error{errAccount, errTxs, errHoldings, errLots, errCash, errChecks} {
		if !errors.Is(err, app.ErrUnknownAccount) {
			t.Errorf("query %d error = %v, want %v", i, err, app.ErrUnknownAccount)
		}
	}
}

func TestStoreFailure(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.ledger.fail = errStore
	v := f.views(t)
	_, errSleeves := v.Sleeves(context.Background())
	_, errHoldings := v.Holdings(context.Background(), "acc-1", f.on)
	for i, err := range []error{errSleeves, errHoldings} {
		if !errors.Is(err, errStore) || errors.Is(err, app.ErrUnknownAccount) {
			t.Errorf("query %d error = %v, want only %v", i, err, errStore)
		}
	}
}
