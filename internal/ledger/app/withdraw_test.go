package app_test

import (
	"errors"
	"testing"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
)

// TestWithdrawAssertions checks D-153's withdrawal: a holding assertion of
// the instrument a ticker's ISIN was corrected from is removed with its
// check, leaving the others; withdrawing it again changes nothing; a cash
// assertion is refused and nothing is removed.
func TestWithdrawAssertions(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	l, _ := seeded(t)
	on := date(t, "2026-06-30")
	old := ledger.HoldingAssertion(acc1, on, instrument(t, isinA, "TSTA"), qty(t, "7"))
	cash := ledger.CashAssertion(acc1, on, money(t, "0", "PLN"))
	if _, err := l.Import(ctx, nil, []ledger.BalanceAssertion{old, cash}); err != nil {
		t.Fatalf("Import: %v", err)
	}
	if err := l.WithdrawAssertions(ctx, []ledger.BalanceAssertion{cash}); !errors.Is(err, app.ErrNotHoldingAssertion) {
		t.Errorf("withdrawing a cash assertion = %v, want ErrNotHoldingAssertion", err)
	}
	for range 2 {
		if err := l.WithdrawAssertions(ctx, []ledger.BalanceAssertion{old}); err != nil {
			t.Fatalf("WithdrawAssertions: %v", err)
		}
	}
	checks, err := l.AssertionChecks(ctx, acc1)
	if err != nil || len(checks) != 1 || checks[0].Assertion.Book.Kind != ledger.BookCash {
		t.Errorf("checks after withdrawing = %+v, %v; want the cash assertion only", checks, err)
	}
}
