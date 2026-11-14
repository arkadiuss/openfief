package app_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views/app"
)

// heldFixture is the fixture with one deposit recorded in acc-1, so there is
// a date the portfolio is valued through: 2026-03-02.
func heldFixture(t *testing.T) *fixture {
	t.Helper()
	f := newFixture(t)
	deposit, err := ledger.NewDeposit(ledger.TransactionHeader{
		Key: ledger.TransactionKey{Source: "test", ExternalID: "dep-1"}, Account: "acc-1", Date: date(t, "2026-03-01"),
	}, money(t, "1000", "PLN"))
	if err != nil {
		t.Fatalf("NewDeposit: %v", err)
	}
	f.ledger.txs = map[ledger.AccountID][]ledger.Transaction{"acc-1": {deposit}}
	return f
}

func TestHeldValues(t *testing.T) {
	t.Parallel()
	f := heldFixture(t)
	got, err := f.views(t).HeldValues(reader(t))
	if err != nil {
		t.Fatalf("HeldValues: %v", err)
	}
	want := app.HeldValues{
		On: f.on, Total: money(t, "2050.00", "PLN"), Unvalued: 1,
		ByInstrument: map[string]kernel.Money{
			instrument(t, isinPL, "EDO0334").String(): money(t, "1050.00", "PLN"),
			instrument(t, isinIE, "ETF1").String():    money(t, "120.00", "PLN"),
		},
	}
	if diff := cmp.Diff(want, got, cmp.Comparer(func(a, b kernel.Money) bool { return a.Equal(b) }), cmp.Comparer(func(a, b kernel.Date) bool { return a.Equal(b) })); diff != "" {
		t.Errorf("HeldValues (-want +got):\n%s", diff)
	}
}

func TestHeldValuesRefusals(t *testing.T) {
	t.Parallel()
	t.Run("nothing recorded", func(t *testing.T) {
		t.Parallel()
		if _, err := newFixture(t).views(t).HeldValues(reader(t)); !errors.Is(err, app.ErrNothingHeld) {
			t.Errorf("HeldValues error = %v, want ErrNothingHeld", err)
		}
	})
	t.Run("a caller that may view nothing sees an empty portfolio", func(t *testing.T) {
		t.Parallel()
		if _, err := heldFixture(t).views(t).HeldValues(t.Context()); !errors.Is(err, app.ErrNothingHeld) {
			t.Errorf("HeldValues with no caller: error = %v, want ErrNothingHeld", err)
		}
	})
	t.Run("a failing store", func(t *testing.T) {
		t.Parallel()
		f := heldFixture(t)
		f.valuations.fail = errStore
		if _, err := f.views(t).HeldValues(reader(t)); !errors.Is(err, errStore) {
			t.Errorf("HeldValues error = %v, want %v", err, errStore)
		}
	})
}
