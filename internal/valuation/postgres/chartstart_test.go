package postgres_test

import (
	"testing"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// TestChartStartRoundTrip stores the charts' start the owner chose
// (D-167): none reads as the zero date, the portfolio's and each account's
// are kept apart, saving the zero date clears one, and forgetting a
// deleted account forgets its start but not the others' (D-120).
func TestChartStartRoundTrip(t *testing.T) {
	t.Parallel()
	store, _ := newStore(t)
	ctx := t.Context()
	got, err := store.ChartStart(ctx, "")
	check(t, "the portfolio's start before any", kernel.Date{}, got, err)

	saves := []struct {
		account ledger.AccountID
		day     string
	}{{"", "2026-03-02"}, {acc1, "2026-03-05"}, {acc2, "2026-04-01"}, {"", "2026-03-03"}}
	for _, s := range saves {
		if saveErr := store.SaveChartStart(ctx, s.account, date(t, s.day)); saveErr != nil {
			t.Fatalf("SaveChartStart(%q): %v", s.account, saveErr)
		}
	}
	got, err = store.ChartStart(ctx, "")
	check(t, "the portfolio's start, saved twice", date(t, "2026-03-03"), got, err)
	got, err = store.ChartStart(ctx, acc1)
	check(t, "acc-1's start", date(t, "2026-03-05"), got, err)

	if clearErr := store.SaveChartStart(ctx, "", kernel.Date{}); clearErr != nil {
		t.Fatalf("SaveChartStart to clear: %v", clearErr)
	}
	got, err = store.ChartStart(ctx, "")
	check(t, "the portfolio's start, cleared", kernel.Date{}, got, err)
	got, err = store.ChartStart(ctx, acc1)
	check(t, "acc-1's start after clearing the portfolio's", date(t, "2026-03-05"), got, err)

	if forgetErr := store.ForgetAccount(ctx, acc1); forgetErr != nil {
		t.Fatalf("ForgetAccount: %v", forgetErr)
	}
	got, err = store.ChartStart(ctx, acc1)
	check(t, "a forgotten account's start", kernel.Date{}, got, err)
	got, err = store.ChartStart(ctx, acc2)
	check(t, "another account's start", date(t, "2026-04-01"), got, err)
}
