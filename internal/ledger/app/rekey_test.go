package app_test

import (
	"context"
	"slices"
	"testing"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
)

// TestRekeyInstrument re-keys every account's holding of a provisional
// instrument to the registry's ISIN (D-214): each transaction posting to
// it, in any account, gets a new version that supersedes it under the
// same key, the import hooks see the earliest date re-keyed, and running
// it again changes nothing and runs no hook.
func TestRekeyInstrument(t *testing.T) {
	t.Parallel()
	store := newMemStore()
	var seen []app.ImportSummary
	l := app.New(store, func(_ context.Context, _ app.Store, sum *app.ImportSummary) error {
		seen = append(seen, *sum)
		return nil
	})
	saveBondAccounts(t, l)
	from, to := instrument(t, "PLEDO0536000", "EDO0536"), instrument(t, "PL9999000018", "EDO0536")
	other := instrument(t, isinA, "SYNA")
	txs := []ledger.Transaction{
		must(t)(ledger.NewBuy(header(t, acc1, "lot:EDO0536:2026-05-04:buy", "2026-05-04"), from, qty(t, "10"), money(t, "1000", "PLN"), money(t, "0", "PLN"))),
		must(t)(ledger.NewInterest(header(t, acc1, "lot:EDO0536:2026-05-04:coupon:2026-10-04", "2026-10-04"), from, money(t, "30", "PLN"), money(t, "0", "PLN"))),
		must(t)(ledger.NewBuy(header(t, acc2, "lot:EDO0536:2026-06-01:buy", "2026-06-01"), from, qty(t, "2"), money(t, "200", "PLN"), money(t, "0", "PLN"))),
		must(t)(ledger.NewBuy(header(t, acc2, "b-other", "2026-03-03"), other, qty(t, "1"), money(t, "10", "PLN"), money(t, "0", "PLN"))),
	}
	if _, err := l.Import(t.Context(), txs, nil); err != nil {
		t.Fatalf("Import: %v", err)
	}
	seen = nil
	if err := l.RekeyInstrument(t.Context(), from, to); err != nil {
		t.Fatalf("RekeyInstrument: %v", err)
	}
	checkNotHeld(t, l, from)
	checkRekeyedLog(t, l, txs[:2], from, to)
	versions, err := store.Versions(t.Context(), txs[0].Key)
	if err != nil || len(versions) != 2 {
		t.Errorf("versions of %s = %d, %v; want the original and the re-keyed one", txs[0].Key, len(versions), err)
	}
	wantEarliest := map[ledger.AccountID]kernel.Date{acc1: date(t, "2026-05-04"), acc2: date(t, "2026-06-01")}
	if len(seen) != 1 || seen[0].NewVersions != 3 || !sameDates(seen[0].EarliestAffected, wantEarliest) {
		t.Errorf("hooks saw %+v, want one re-key of three transactions from %v", seen, wantEarliest)
	}
	seen = nil
	if err := l.RekeyInstrument(t.Context(), from, to); err != nil || len(seen) != 0 {
		t.Errorf("RekeyInstrument again = %v with %d hook runs, want nothing done", err, len(seen))
	}
	checkRekeyedAgain(t, l, store, txs[0].Key, to)
}

// saveBondAccounts saves acc-1 and acc-2 in l.
func saveBondAccounts(t *testing.T, l *app.Ledger) {
	t.Helper()
	for _, id := range []ledger.AccountID{acc1, acc2} {
		if err := l.SaveAccount(t.Context(), ledger.Account{ID: id, Institution: "treasury-bonds", Name: string(id)}); err != nil {
			t.Fatalf("SaveAccount: %v", err)
		}
	}
}

// checkRekeyedAgain fails t unless key still has its two versions after a
// re-key run again, and a re-key from no instrument fails.
func checkRekeyedAgain(t *testing.T, l *app.Ledger, store *memStore, key ledger.TransactionKey, to ledger.Instrument) {
	t.Helper()
	if again, err := store.Versions(t.Context(), key); err != nil || len(again) != 2 {
		t.Errorf("versions after re-keying again = %d, %v; want still 2", len(again), err)
	}
	if err := l.RekeyInstrument(t.Context(), ledger.Instrument{}, to); err == nil {
		t.Errorf("RekeyInstrument from no instrument = nil, want an error")
	}
}

// checkNotHeld fails t when acc-1 or acc-2 holds in.
func checkNotHeld(t *testing.T, l *app.Ledger, in ledger.Instrument) {
	t.Helper()
	for _, id := range []ledger.AccountID{acc1, acc2} {
		holdings, err := l.HoldingsOn(t.Context(), id, date(t, "2026-11-07"))
		if err != nil {
			t.Fatalf("HoldingsOn: %v", err)
		}
		if slices.ContainsFunc(holdings, func(h ledger.Holding) bool { return h.Instrument == in }) {
			t.Errorf("%s still holds %s after the re-key", id, in)
		}
	}
}

// checkRekeyedLog fails t unless acc-1's current transactions are txs,
// re-keyed (ledger.Rekeyed).
func checkRekeyedLog(t *testing.T, l *app.Ledger, txs []ledger.Transaction, from, to ledger.Instrument) {
	t.Helper()
	current, err := l.CurrentTransactions(t.Context(), acc1)
	if err != nil {
		t.Fatalf("CurrentTransactions: %v", err)
	}
	want := make([]string, 0, len(txs))
	for i := range txs {
		r, _ := ledger.Rekeyed(&txs[i], from, to)
		want = append(want, ledger.Fingerprint(&r))
	}
	got := make([]string, 0, len(current))
	for i := range current {
		got = append(got, ledger.Fingerprint(&current[i]))
	}
	if !slices.Equal(want, got) {
		t.Errorf("acc-1's transactions after the re-key = %v, want the re-keyed ones", current)
	}
}

// sameDates reports whether a and b hold the same date for each account.
func sameDates(a, b map[ledger.AccountID]kernel.Date) bool {
	if len(a) != len(b) {
		return false
	}
	for id, d := range a {
		if !d.Equal(b[id]) {
			return false
		}
	}
	return true
}

// TestNamesRekeyInstrument gives the registry's ISIN the names its
// provisional instrument showed (D-214), once.
func TestNamesRekeyInstrument(t *testing.T) {
	t.Parallel()
	from, to := instrument(t, "PLEDO0536000", "EDO0536"), instrument(t, "PL9999000018", "EDO0536")
	store := &memNames{records: []ledger.InstrumentName{
		{ISIN: from.ISIN, Name: "EDO0536", Source: "bonds-main"}, {ISIN: from.ISIN, Name: "My EDO", ByOwner: true},
	}}
	n := &app.Names{Store: store}
	for range 2 {
		if err := n.RekeyInstrument(t.Context(), from, to); err != nil {
			t.Fatalf("RekeyInstrument: %v", err)
		}
	}
	checkName(t, n, to.ISIN, "My EDO")
	if len(store.records) != 4 {
		t.Errorf("stored %d names, want the two carried over once", len(store.records))
	}
}
