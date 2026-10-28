package app_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/portfolio"
)

// Synthetic IDs and ISINs with valid check digits; nothing real is implied.
const (
	isinA = "PLTEST000010"

	acc1 ledger.AccountID   = "acc-1"
	acc2 ledger.AccountID   = "acc-2"
	core portfolio.SleeveID = "core"
)

func dec(t *testing.T, s string) kernel.Decimal {
	t.Helper()
	d, err := kernel.ParseDecimal(s)
	if err != nil {
		t.Fatalf("ParseDecimal(%q): %v", s, err)
	}
	return d
}

func money(t *testing.T, amount, code string) kernel.Money {
	t.Helper()
	c, err := kernel.ParseCurrency(code)
	if err != nil {
		t.Fatalf("ParseCurrency(%q): %v", code, err)
	}
	return kernel.NewMoney(dec(t, amount), c)
}

func qty(t *testing.T, s string) kernel.Quantity {
	t.Helper()
	return kernel.NewQuantity(dec(t, s))
}

func date(t *testing.T, s string) kernel.Date {
	t.Helper()
	d, err := kernel.ParseDate(s)
	if err != nil {
		t.Fatalf("ParseDate(%q): %v", s, err)
	}
	return d
}

func instrument(t *testing.T, isin, symbol string) ledger.Instrument {
	t.Helper()
	i, err := ledger.ParseISIN(isin)
	if err != nil {
		t.Fatalf("ParseISIN(%q): %v", isin, err)
	}
	return ledger.Instrument{ISIN: i, Symbol: symbol}
}

// must unwraps a transaction constructor's result: must(t)(ledger.NewBuy(...)).
func must(t *testing.T) func(ledger.Transaction, error) ledger.Transaction {
	t.Helper()
	return func(tx ledger.Transaction, err error) ledger.Transaction {
		t.Helper()
		if err != nil {
			t.Fatalf("building transaction: %v", err)
		}
		return tx
	}
}

// header returns a header in account from source "test".
func header(t *testing.T, account ledger.AccountID, id, day string) ledger.TransactionHeader {
	t.Helper()
	return ledger.TransactionHeader{
		Key:     ledger.TransactionKey{Source: "test", ExternalID: id},
		Account: account,
		Date:    date(t, day),
	}
}

// seeded returns the use cases over an empty memStore with the built-in
// sleeve kinds, the sleeve core and the accounts acc-1 and acc-2 in it.
func seeded(t *testing.T) (*app.Ledger, *memStore) {
	t.Helper()
	store := newMemStore()
	l := app.New(store)
	for _, k := range portfolio.DefaultKinds() {
		if err := l.SaveKind(t.Context(), k); err != nil {
			t.Fatalf("SaveKind(%s): %v", k.Key, err)
		}
	}
	if err := l.SaveSleeve(t.Context(), coreSleeve()); err != nil {
		t.Fatalf("SaveSleeve: %v", err)
	}
	for _, id := range []ledger.AccountID{acc1, acc2} {
		if err := l.SaveAccount(t.Context(), ledger.Account{ID: id, Institution: "test-broker", Name: string(id), Sleeve: core}); err != nil {
			t.Fatalf("SaveAccount(%s): %v", id, err)
		}
	}
	return l, store
}

// coreSleeve returns the sleeve core that seeded stores.
func coreSleeve() portfolio.Sleeve {
	return portfolio.Sleeve{ID: core, Name: "Core", Kind: portfolio.KindCore, Role: portfolio.RoleDefensive}
}

// checkResult fails t on err, or unless got is want; empty and nil slices
// are equal.
func checkResult[T any](t *testing.T, name string, want, got T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("%s (-want +got):\n%s", name, diff)
	}
}
